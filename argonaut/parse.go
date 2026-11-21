package argonaut

import (
	"strconv"

	"github.com/i-am-the-slime/iris-go/object"
	"github.com/i-am-the-slime/iris-go/prim"
)

// JsonParser implements Data.Argonaut.Parser._jsonParser: the parsed value
// passed to succ, or "Invalid JSON at offset N" passed to fail, N counting
// UTF-16 code units up to the first character that cannot continue a JSON
// text.
func JsonParser[A any](fail func(string) A, succ func(Json) A, text string) A {
	value, offset, ok := parse(text)
	if !ok {
		return fail("Invalid JSON at offset " + strconv.Itoa(int(offset)))
	}
	return succ(value)
}

// syntaxError ends a parse at the byte where the input stopped being JSON.
type syntaxError struct{ at int }

type parser struct {
	input    string
	position int
}

func parse(text string) (value Json, offset int32, ok bool) {
	p := &parser{input: text}
	defer func() {
		if failure := recover(); failure != nil {
			failed, isSyntaxError := failure.(syntaxError)
			if !isSyntaxError {
				panic(failure)
			}
			value, offset, ok = nil, prim.CodeUnitLength(text[:failed.at]), false
		}
	}()
	value = p.value()
	p.skipWhitespace()
	if p.position != len(p.input) {
		p.fail()
	}
	return value, 0, true
}

func (p *parser) fail() {
	panic(syntaxError{p.position})
}

func (p *parser) value() Json {
	p.skipWhitespace()
	switch p.peek() {
	case '{':
		return p.object()
	case '[':
		return p.array()
	case '"':
		return jsonString(p.string())
	case 't':
		p.keyword("true")
		return jsonBoolean(true)
	case 'f':
		p.keyword("false")
		return jsonBoolean(false)
	case 'n':
		p.keyword("null")
		return nil
	}
	return jsonNumber(p.number())
}

func (p *parser) array() Json {
	p.position++
	values := jsonArray{}
	if p.skipWhitespace(); p.peek() == ']' {
		p.position++
		return values
	}
	for {
		values = append(values, p.value())
		p.skipWhitespace()
		switch p.peek() {
		case ',':
			p.position++
		case ']':
			p.position++
			return values
		default:
			p.fail()
		}
	}
}

func (p *parser) object() Json {
	p.position++
	table := object.New[Json]()()
	if p.skipWhitespace(); p.peek() == '}' {
		p.position++
		return jsonObject{table}
	}
	for {
		p.skipWhitespace()
		if p.peek() != '"' {
			p.fail()
		}
		key := p.string()
		p.skipWhitespace()
		p.expect(':')
		object.Poke(key, p.value(), table)()
		p.skipWhitespace()
		switch p.peek() {
		case ',':
			p.position++
		case '}':
			p.position++
			return jsonObject{table}
		default:
			p.fail()
		}
	}
}

func (p *parser) keyword(word string) {
	for index := 0; index < len(word); index++ {
		p.expect(word[index])
	}
}

func (p *parser) expect(character byte) {
	if p.peek() != character || p.position == len(p.input) {
		p.fail()
	}
	p.position++
}

// number reads `-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?`. Out of
// range literals become an infinity or a zero.
func (p *parser) number() float64 {
	start := p.position
	if p.peek() == '-' {
		p.position++
	}
	switch {
	case p.peek() == '0':
		p.position++
	case isDigit(p.peek()):
		p.digits()
	default:
		p.fail()
	}
	if p.peek() == '.' {
		p.position++
		p.digits()
	}
	if p.peek() == 'e' || p.peek() == 'E' {
		p.position++
		if p.peek() == '+' || p.peek() == '-' {
			p.position++
		}
		p.digits()
	}
	number, _ := strconv.ParseFloat(p.input[start:p.position], 64)
	return number
}

// digits reads one or more decimal digits.
func (p *parser) digits() {
	if !isDigit(p.peek()) {
		p.fail()
	}
	for isDigit(p.peek()) {
		p.position++
	}
}

// string reads a string literal. Escaped surrogates join into one code
// point when they pair up and stay lone surrogates when they do not.
func (p *parser) string() string {
	p.position++
	var contents prim.Builder
	for {
		character := p.peek()
		switch {
		case p.position == len(p.input) || character < 0x20:
			p.fail()
		case character == '"':
			p.position++
			return contents.String()
		case character != '\\':
			contents.WriteByte(character)
			p.position++
			continue
		}
		p.position++
		escape := p.peek()
		if escape == 'u' {
			p.position++
			contents.WriteRune(p.unicodeEscape())
			continue
		}
		replacement, valid := shortEscapes[escape]
		if !valid || p.position == len(p.input) {
			p.fail()
		}
		contents.WriteByte(replacement)
		p.position++
	}
}

var shortEscapes = map[byte]byte{
	'"': '"', '\\': '\\', '/': '/', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t',
}

// unicodeEscape reads the four hex digits after `\u`.
func (p *parser) unicodeEscape() rune {
	var value rune
	for range 4 {
		digit, valid := hexDigit(p.peek())
		if !valid || p.position == len(p.input) {
			p.fail()
		}
		value = value<<4 | digit
		p.position++
	}
	return value
}

func hexDigit(character byte) (rune, bool) {
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

func (p *parser) skipWhitespace() {
	for p.position < len(p.input) && isWhitespace(p.input[p.position]) {
		p.position++
	}
}

// peek is the byte at the position, or zero at the end.
func (p *parser) peek() byte {
	if p.position < len(p.input) {
		return p.input[p.position]
	}
	return 0
}

func isWhitespace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\n' || character == '\r'
}

func isDigit(character byte) bool {
	return '0' <= character && character <= '9'
}
