package foreignvalue

import (
	"math"
	"unsafe"
)

// Yoga.JSON's readJSON and writeJSON at a type the compiler knows the whole
// codec of. The compiler emits one decoder and one encoder per type, which
// read the text straight into the program's Go values and write them straight
// out, without building the Foreign tree in between; these are their leaves.
//
// A decoder only has a fast path to offer: when the text does not decode,
// for whatever reason, Decode reports false and the program reads the text
// again the way Yoga.JSON does, which reports exactly Yoga.JSON's errors.

// A Decoder reads one JSON text front to back.
type Decoder struct {
	parser jsonParser
}

// decodeFailure ends a decode whose text is JSON but not of the type.
type decodeFailure struct{}

// Decode is the value decode reads from the whole of payload, and whether
// payload was one JSON text of the type.
func Decode[T any](payload string, decode func(*Decoder) T) (value T, ok bool) {
	d := &Decoder{parser: jsonParser{input: payload}}
	defer func() {
		if failure := recover(); failure != nil {
			switch failure.(type) {
			case decodeFailure, jsonSyntaxError:
				var zero T
				value, ok = zero, false
			default:
				panic(failure)
			}
		}
	}()
	value = decode(d)
	d.parser.skipWhitespace()
	return value, d.parser.position == len(payload)
}

// Fail ends the decode: the text is not of the type.
func (d *Decoder) Fail() {
	panic(decodeFailure{})
}

// peek is the next byte after whitespace, or zero at the end.
func (d *Decoder) peek() byte {
	d.parser.skipWhitespace()
	return d.parser.peek()
}

// Object enters an object, reporting whether it has a member.
func (d *Decoder) Object() bool {
	return d.open('{', '}')
}

// Array enters an array, reporting whether it has an element.
func (d *Decoder) Array() bool {
	return d.open('[', ']')
}

func (d *Decoder) open(opening, closing byte) bool {
	if d.peek() != opening {
		d.Fail()
	}
	d.parser.position++
	if d.peek() == closing {
		d.parser.position++
		return false
	}
	return true
}

// Key reads a member's key and the colon after it.
func (d *Decoder) Key() string {
	p := &d.parser
	if d.peek() != '"' {
		p.fail()
	}
	key := p.string()
	if d.peek() != ':' {
		p.fail()
	}
	p.position++
	return key
}

// MoreMembers moves past the comma before the next member, or reports the
// object's end.
func (d *Decoder) MoreMembers() bool {
	return d.more('}')
}

// MoreElements moves past the comma before the next element, or reports the
// array's end.
func (d *Decoder) MoreElements() bool {
	return d.more(']')
}

func (d *Decoder) more(closing byte) bool {
	switch d.peek() {
	case ',':
		d.parser.position++
		return true
	case closing:
		d.parser.position++
		return false
	}
	d.parser.fail()
	return false
}

// Null reads a null if one comes next.
func (d *Decoder) Null() bool {
	if d.peek() != 'n' {
		return false
	}
	d.parser.keyword("null")
	return true
}

// String is Foreign.readString.
func (d *Decoder) String() string {
	if d.peek() != '"' {
		d.Fail()
	}
	return d.parser.string()
}

// Number is Foreign.readNumber.
func (d *Decoder) Number() float64 {
	if c := d.peek(); c != '-' && (c < '0' || c > '9') {
		d.Fail()
	}
	return d.parser.numberValue()
}

// Int is Foreign.readInt: a Number that Data.Int.fromNumber accepts.
func (d *Decoder) Int() int {
	number := d.Number()
	if number != math.Trunc(number) || number < math.MinInt32 || number > math.MaxInt32 {
		d.Fail()
	}
	return int(number)
}

// Boolean is Foreign.readBoolean.
func (d *Decoder) Boolean() bool {
	switch d.peek() {
	case 't':
		d.parser.keyword("true")
		return true
	case 'f':
		d.parser.keyword("false")
		return false
	}
	d.Fail()
	return false
}

// Skip reads a value nothing asks for, which must still be JSON.
func (d *Decoder) Skip() {
	d.parser.value()
}

// An Encoder writes one JSON text as JSON.stringify does.
type Encoder struct {
	out []byte
}

// Encode is the text encode writes for value.
func Encode[T any](encode func(*Encoder, T), value T) string {
	e := &Encoder{}
	encode(e, value)
	return unsafe.String(unsafe.SliceData(e.out), len(e.out))
}

// Raw writes text that is already JSON, such as a quoted key and its colon.
func (e *Encoder) Raw(text string) {
	e.out = append(e.out, text...)
}

// Member writes a member's quoted key and colon, after a comma unless it is
// the object's first.
func (e *Encoder) Member(key string) {
	if e.out[len(e.out)-1] != '{' {
		e.out = append(e.out, ',')
	}
	e.out = append(e.out, key...)
}

// Key is Member for a key that is not quoted yet.
func (e *Encoder) Key(key string) {
	if e.out[len(e.out)-1] != '{' {
		e.out = append(e.out, ',')
	}
	e.out = appendJSONString(e.out, key)
	e.out = append(e.out, ':')
}

// String writes a quoted string.
func (e *Encoder) String(s string) {
	e.out = appendJSONString(e.out, s)
}

// Number writes a Number as Number::toString does, NaN and the infinities as
// null.
func (e *Encoder) Number(n float64) {
	w := jsonWriter{out: e.out}
	w.number(n)
	e.out = w.out
}

// Int writes an Int.
func (e *Encoder) Int(i int) {
	e.out = appendInt32(e.out, i)
}

// Boolean writes true or false.
func (e *Encoder) Boolean(b bool) {
	if b {
		e.out = append(e.out, "true"...)
	} else {
		e.out = append(e.out, "false"...)
	}
}

func appendInt32(out []byte, i int) []byte {
	if 0 <= i && i < 10 {
		return append(out, byte('0'+i))
	}
	var digits [11]byte
	at := len(digits)
	n := int64(i)
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		at--
		digits[at] = byte('0' + n%10)
		n /= 10
	}
	if negative {
		at--
		digits[at] = '-'
	}
	return append(out, digits[at:]...)
}
