package foreignvalue

import (
	"math"
	"strconv"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/prim"
)

// ParseJSON implements Yoga.JSON._parseJSON (`JSON.parse`): objects keep
// their keys in insertion order (a repeated key keeps its first position and
// its last value), numbers are float64, and malformed input throws a
// SyntaxError "Invalid JSON at offset N", N counting UTF-16 code units up to
// the first character that cannot continue a JSON text.
//
// It reads the text in one pass of recursive descent. A string without
// escapes is a substring of the payload, so it shares the payload's bytes.
func ParseJSON(payload string) any {
	value, failedAt, ok := parseJSON(payload)
	if !ok {
		panic(exception.Error{Name: "SyntaxError", Message: "Invalid JSON at offset " + strconv.Itoa(int(prim.CodeUnitLength(payload[:failedAt])))})
	}
	return value
}

// jsonSyntaxError ends a parse at the byte where the input stopped being JSON.
type jsonSyntaxError struct{ at int }

type jsonParser struct {
	input    string
	position int
	// values and keys are the members of the arrays and objects being read,
	// innermost last; each array or object takes its own exactly sized copy.
	values []any
	keys   []string
}

func parseJSON(text string) (value any, failedAt int, ok bool) {
	p := &jsonParser{input: text}
	defer func() {
		if failure := recover(); failure != nil {
			failed, isSyntaxError := failure.(jsonSyntaxError)
			if !isSyntaxError {
				panic(failure)
			}
			value, failedAt, ok = nil, failed.at, false
		}
	}()
	value = p.value()
	p.skipWhitespace()
	if p.position != len(p.input) {
		p.fail()
	}
	return value, 0, true
}

func (p *jsonParser) fail() {
	panic(jsonSyntaxError{min(p.position, len(p.input))})
}

func (p *jsonParser) value() any {
	p.skipWhitespace()
	if p.position >= len(p.input) {
		p.fail()
	}
	switch p.input[p.position] {
	case '{':
		return p.object()
	case '[':
		return p.array()
	case '"':
		return p.string()
	case 't':
		p.keyword("true")
		return true
	case 'f':
		p.keyword("false")
		return false
	case 'n':
		p.keyword("null")
		return nil
	}
	return p.number()
}

func (p *jsonParser) array() any {
	p.position++
	if p.skipWhitespace(); p.peek() == ']' {
		p.position++
		return []any{}
	}
	start := len(p.values)
	for {
		item := p.value()
		p.values = append(p.values, item)
		p.skipWhitespace()
		switch p.peek() {
		case ',':
			p.position++
		case ']':
			p.position++
			items := make([]any, len(p.values)-start)
			copy(items, p.values[start:])
			clear(p.values[start:])
			p.values = p.values[:start]
			return items
		default:
			p.fail()
		}
	}
}

func (p *jsonParser) object() any {
	p.position++
	if p.skipWhitespace(); p.peek() == '}' {
		p.position++
		return newObjectTable(nil, nil)
	}
	start := len(p.keys)
	for {
		p.skipWhitespace()
		if p.peek() != '"' {
			p.fail()
		}
		key := p.string()
		p.skipWhitespace()
		if p.peek() != ':' {
			p.fail()
		}
		p.position++
		member := p.value()
		p.keys = append(p.keys, key)
		p.values = append(p.values, member)
		p.skipWhitespace()
		switch p.peek() {
		case ',':
			p.position++
		case '}':
			p.position++
			table := newObjectTable(p.keys[start:], p.values[len(p.values)-(len(p.keys)-start):])
			clear(p.values[len(p.values)-(len(p.keys)-start):])
			p.values = p.values[:len(p.values)-(len(p.keys)-start)]
			p.keys = p.keys[:start]
			return table
		default:
			p.fail()
		}
	}
}

// newObjectTable is the `Object Foreign` of the members in order, a repeated
// key keeping its first slot and taking its last value.
func newObjectTable(keys []string, values []any) objectTable {
	table := objectTable(&struct {
		Index map[string]int
		Ints  int
		Keys  []string
		Vals  []any
	}{Index: make(map[string]int, len(keys)), Keys: make([]string, 0, len(keys)), Vals: make([]any, 0, len(keys))})
	for i, key := range keys {
		if slot, repeated := table.Index[key]; repeated {
			table.Vals[slot] = values[i]
			continue
		}
		table.Index[key] = len(table.Keys)
		table.Keys = append(table.Keys, key)
		table.Vals = append(table.Vals, values[i])
	}
	return table
}

func (p *jsonParser) keyword(word string) {
	if len(p.input)-p.position >= len(word) && p.input[p.position:p.position+len(word)] == word {
		p.position += len(word)
		return
	}
	for index := 0; index < len(word); index++ {
		if p.peek() != word[index] {
			p.fail()
		}
		p.position++
	}
}

// number reads `-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?` as JS does,
// correctly rounded; out of range literals become an infinity or a zero.
func (p *jsonParser) number() any {
	start := p.position
	negative := p.peek() == '-'
	if negative {
		p.position++
	}
	integerStart := p.position
	switch c := p.peek(); {
	case c == '0':
		p.position++
	case '1' <= c && c <= '9':
		p.position = digitRunEnd(p.input, p.position)
	default:
		p.fail()
	}
	integerEnd := p.position
	fractionEnd := integerEnd
	if p.peek() == '.' {
		p.position++
		p.digits()
		fractionEnd = p.position
	}
	exponent := 0
	if c := p.peek(); c == 'e' || c == 'E' {
		p.position++
		exponentNegative := false
		if c := p.peek(); c == '+' || c == '-' {
			exponentNegative = c == '-'
			p.position++
		}
		exponentStart := p.position
		p.digits()
		if p.position-exponentStart > 4 {
			return p.slowNumber(start)
		}
		for _, digit := range []byte(p.input[exponentStart:p.position]) {
			exponent = exponent*10 + int(digit-'0')
		}
		if exponentNegative {
			exponent = -exponent
		}
	}
	if number, exact := exactNumber(p.input[integerStart:integerEnd], p.input[min(integerEnd+1, fractionEnd):fractionEnd], exponent); exact {
		if negative {
			number = -number
		}
		return boxNumber(number)
	}
	return p.slowNumber(start)
}

func (p *jsonParser) slowNumber(start int) any {
	number, _ := strconv.ParseFloat(p.input[start:p.position], 64)
	return number
}

// exactNumber is the value of a literal whose digits, read as one integer,
// are exactly a float64 and whose power of ten is too: then one
// multiplication or division rounds correctly (Clinger's fast path).
func exactNumber(integer, fraction string, exponent int) (float64, bool) {
	if len(integer)+len(fraction) > 15 {
		return 0, false
	}
	mantissa := int64(0)
	for i := 0; i < len(integer); i++ {
		mantissa = mantissa*10 + int64(integer[i]-'0')
	}
	for i := 0; i < len(fraction); i++ {
		mantissa = mantissa*10 + int64(fraction[i]-'0')
	}
	exponent -= len(fraction)
	switch {
	case exponent == 0:
		return float64(mantissa), true
	case 0 < exponent && exponent <= 22:
		return float64(mantissa) * exactPowersOfTen[exponent], true
	case -22 <= exponent && exponent < 0:
		return float64(mantissa) / exactPowersOfTen[-exponent], true
	}
	return 0, false
}

var exactPowersOfTen = [23]float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9, 1e10, 1e11,
	1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19, 1e20, 1e21, 1e22,
}

// boxNumber is a number as a Foreign, sharing one box for each small
// non-negative integer, which Go would otherwise allocate each time.
func boxNumber(number float64) any {
	if index := int(number); number >= 0 && index < len(smallNumbers) && float64(index) == number && !math.Signbit(number) {
		return smallNumbers[index]
	}
	return number
}

var smallNumbers = func() (boxes [1024]any) {
	for i := range boxes {
		boxes[i] = float64(i)
	}
	return boxes
}()

// digits reads one or more decimal digits.
func (p *jsonParser) digits() {
	if c := p.peek(); c < '0' || c > '9' {
		p.fail()
	}
	p.position = digitRunEnd(p.input, p.position)
}

// string reads a string literal. Escaped surrogates join into one code
// point when they pair up and stay lone surrogates when they do not.
func (p *jsonParser) string() string {
	p.position++
	start := p.position
	// stringSpecial, by hand: the call costs more than a short string.
	end, stop := start, min(start+scalarPrefix, len(p.input))
	for end < stop && !stringSpecials[p.input[end]] {
		end++
	}
	if end == stop {
		end = vectorStringSpecial(p.input, end)
	}
	if end < len(p.input) && p.input[end] == '"' {
		p.position = end + 1
		return p.input[start:end]
	}
	var contents prim.Builder
	contents.Grow(end - start + 16)
	contents.WriteString(p.input[start:end])
	p.position = end
	for {
		if p.position >= len(p.input) {
			p.fail()
		}
		switch character := p.input[p.position]; {
		case character == '"':
			p.position++
			return contents.String()
		case character < 0x20:
			p.fail()
		case character != '\\':
			end := stringSpecial(p.input, p.position)
			contents.WriteString(p.input[p.position:end])
			p.position = end
			continue
		}
		p.position++
		escape := p.peek()
		if escape == 'u' {
			p.position++
			contents.WriteRune(p.unicodeEscape())
			continue
		}
		replacement := shortEscapes[escape]
		if replacement == 0 {
			p.fail()
		}
		contents.WriteByte(replacement)
		p.position++
	}
}

// shortEscapes is the character each one-letter escape stands for, or zero.
var shortEscapes = [256]byte{
	'"': '"', '\\': '\\', '/': '/', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t',
}

// unicodeEscape reads the four hex digits after `\u`.
func (p *jsonParser) unicodeEscape() rune {
	var value rune
	for range 4 {
		digit, valid := jsonHexDigit(p.peek())
		if !valid {
			p.fail()
		}
		value = value<<4 | digit
		p.position++
	}
	return value
}

func jsonHexDigit(character byte) (rune, bool) {
	switch {
	case '0' <= character && character <= '9':
		return rune(character - '0'), true
	case 'a' <= character && character <= 'f':
		return rune(character-'a') + 10, true
	case 'A' <= character && character <= 'F':
		return rune(character-'A') + 10, true
	}
	return 0, false
}

func (p *jsonParser) skipWhitespace() {
	if p.position < len(p.input) && p.input[p.position] > ' ' {
		return
	}
	p.position = whitespaceRunEnd(p.input, p.position)
}

// peek is the byte at the position, or zero at the end.
func (p *jsonParser) peek() byte {
	if p.position < len(p.input) {
		return p.input[p.position]
	}
	return 0
}
