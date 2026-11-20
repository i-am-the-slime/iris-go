package argonaut

import (
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/i-am-the-slime/iris-go/object"
	"github.com/i-am-the-slime/iris-go/prim"
)

// JsonParser implements Data.Argonaut.Parser._jsonParser: `JSON.parse(s)`
// passed to succ, or the message of the SyntaxError it throws passed to fail.
//
// The messages are JavaScriptCore's (Bun's), whose JSON parser this follows
// token by token: a lexical error reports the lexer's message, a grammatical
// one the parser's, and trailing input after a complete value a generic one.
func JsonParser[A any](fail func(string) A, succ func(Json) A, text string) A {
	value, message := parse(text)
	if message != "" {
		return fail("JSON Parse error: " + message)
	}
	return succ(value)
}

type tokenKind int

const (
	tokenEnd tokenKind = iota
	tokenError
	tokenPunctuation
	tokenString
	tokenNumber
	tokenTrue
	tokenFalse
	tokenNull
	tokenIdentifier
)

type token struct {
	kind tokenKind
	// text is a punctuation character, a string's contents or an identifier.
	text   string
	number float64
}

// syntaxError ends a parse with its message, the lexer's if it found an error.
type syntaxError struct{ message string }

type parser struct {
	input    string
	position int
	// lexical is the lexer's error, which outranks the parser's.
	lexical string
}

func parse(text string) (value Json, message string) {
	parser := &parser{input: text}
	defer func() {
		if failure := recover(); failure != nil {
			failed, ok := failure.(syntaxError)
			if !ok {
				panic(failure)
			}
			value, message = nil, failed.message
		}
	}()
	value = parser.value(parser.next())
	if parser.next().kind != tokenEnd {
		parser.fail("")
	}
	return value, ""
}

func (p *parser) fail(message string) {
	switch {
	case p.lexical != "":
		message = p.lexical
	case message == "":
		message = "Unable to parse JSON string"
	}
	panic(syntaxError{message})
}

func (p *parser) value(current token) Json {
	switch current.kind {
	case tokenString:
		return jsonString(current.text)
	case tokenNumber:
		return jsonNumber(current.number)
	case tokenTrue:
		return jsonBoolean(true)
	case tokenFalse:
		return jsonBoolean(false)
	case tokenNull:
		return nil
	case tokenEnd:
		p.fail("Unexpected EOF")
	case tokenIdentifier:
		p.fail(`Unexpected identifier "` + current.text + `"`)
	case tokenPunctuation:
		switch current.text {
		case "[":
			return p.array()
		case "{":
			return p.object()
		}
		p.fail("Unexpected token '" + current.text + "'")
	}
	p.fail("")
	return nil
}

func (p *parser) array() Json {
	values := jsonArray{}
	current := p.next()
	if isPunctuation(current, "]") {
		return values
	}
	for {
		values = append(values, p.value(current))
		current = p.next()
		switch {
		case isPunctuation(current, ","):
			current = p.next()
			if isPunctuation(current, "]") {
				p.fail("Unexpected comma at the end of array expression")
			}
		case isPunctuation(current, "]"):
			return values
		default:
			p.fail("Expected ']'")
		}
	}
}

func (p *parser) object() Json {
	table := object.New[Json]()()
	current := p.next()
	if isPunctuation(current, "}") {
		return jsonObject{table}
	}
	if current.kind != tokenString {
		p.fail("Expected '}'")
	}
	for {
		key := current.text
		if !isPunctuation(p.next(), ":") {
			p.fail("Expected ':' before value in object property definition")
		}
		object.Poke(key, p.value(p.next()), table)()
		current = p.next()
		switch {
		case isPunctuation(current, ","):
			current = p.next()
			if current.kind != tokenString {
				p.fail("Property name must be a string literal")
			}
		case isPunctuation(current, "}"):
			return jsonObject{table}
		default:
			p.fail("Expected '}'")
		}
	}
}

func isPunctuation(current token, text string) bool {
	return current.kind == tokenPunctuation && current.text == text
}

// next lexes one token, recording a lexical error's message.
func (p *parser) next() token {
	for p.position < len(p.input) && isWhitespace(p.input[p.position]) {
		p.position++
	}
	if p.position == len(p.input) {
		return token{kind: tokenEnd}
	}
	character := p.input[p.position]
	switch character {
	case '[', ']', '{', '}', ',', ':', '.', '(', ')', ';', '=':
		p.position++
		return token{kind: tokenPunctuation, text: string(character)}
	case '"':
		return p.string()
	case '\'':
		return p.lexicalError("Single quotes (') are not allowed in JSON")
	}
	if character == '-' || isDigit(character) {
		return p.number()
	}
	for _, keyword := range keywords {
		if strings.HasPrefix(p.input[p.position:], keyword.text) {
			p.position += len(keyword.text)
			return token{kind: keyword.kind}
		}
	}
	if isIdentifierStart(character) {
		start := p.position
		for p.position < len(p.input) && isIdentifierPart(p.input[p.position]) {
			p.position++
		}
		return token{kind: tokenIdentifier, text: p.input[start:p.position]}
	}
	return p.lexicalError("Unrecognized token '" + firstCodeUnit(p.input[p.position:]) + "'")
}

var keywords = []struct {
	text string
	kind tokenKind
}{{"true", tokenTrue}, {"false", tokenFalse}, {"null", tokenNull}}

func (p *parser) lexicalError(message string) token {
	p.lexical = message
	return token{kind: tokenError}
}

// number lexes `-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?`.
func (p *parser) number() token {
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
		return p.lexicalError("Invalid number")
	}
	if p.peek() == '.' {
		p.position++
		if !isDigit(p.peek()) {
			return p.lexicalError("Invalid digits after decimal point")
		}
		p.digits()
	}
	if p.peek() == 'e' || p.peek() == 'E' {
		p.position++
		if p.peek() == '+' || p.peek() == '-' {
			p.position++
		}
		if !isDigit(p.peek()) {
			return p.lexicalError(
				"Exponent symbols should be followed by an optional '+' or '-' and then by at least one number",
			)
		}
		p.digits()
	}
	// The grammar is ParseFloat's, whose only possible error is a range error
	// that still returns JavaScript's result: an infinity or a zero.
	number, _ := strconv.ParseFloat(p.input[start:p.position], 64)
	return token{kind: tokenNumber, number: number}
}

func (p *parser) digits() {
	for isDigit(p.peek()) {
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

// string lexes a string literal. Escaped surrogates join into one code
// point when they pair up and stay lone surrogates when they do not.
func (p *parser) string() token {
	p.position++
	var contents prim.Builder
	for {
		if p.position == len(p.input) {
			return p.lexicalError("Unterminated string")
		}
		character := p.input[p.position]
		switch {
		case character == '"':
			p.position++
			return token{kind: tokenString, text: contents.String()}
		case character < 0x20:
			return p.lexicalError("Unterminated string")
		case character != '\\':
			contents.WriteByte(character)
			p.position++
			continue
		}
		p.position++
		if p.position == len(p.input) {
			return p.lexicalError("Unterminated string")
		}
		escape := p.input[p.position]
		if escape != 'u' {
			replacement, valid := shortEscapes[escape]
			if !valid {
				return p.lexicalError("Invalid escape character " + firstCodeUnit(p.input[p.position:]))
			}
			contents.WriteByte(replacement)
			p.position++
			continue
		}
		p.position++
		unit, message := p.unicodeEscape()
		if message != "" {
			return p.lexicalError(message)
		}
		contents.WriteRune(unit)
	}
}

var shortEscapes = map[byte]byte{
	'"': '"', '\\': '\\', '/': '/', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t',
}

// unicodeEscape reads the four hex digits after `\u`. JavaScriptCore counts
// them in UTF-16 code units.
func (p *parser) unicodeEscape() (rune, string) {
	rest := p.input[p.position:]
	units, length := 0, 0
	for units < 4 && length < len(rest) {
		character, size := utf8.DecodeRuneInString(rest[length:])
		units += utf16.RuneLen(character)
		length += size
	}
	if units < 4 {
		return 0, `\u must be followed by 4 hex digits`
	}
	digits := rest[:length]
	value, err := strconv.ParseUint(digits, 16, 16)
	if err != nil || len(digits) != 4 {
		return 0, `"\u` + firstCodeUnits(digits, 4) + `" is not a valid unicode escape`
	}
	p.position += 4
	return rune(value), ""
}

// firstCodeUnit is the first UTF-16 code unit of a non-empty string, which
// JavaScriptCore's messages quote: a whole character unless it is astral.
func firstCodeUnit(text string) string {
	return firstCodeUnits(text, 1)
}

func firstCodeUnits(text string, count int) string {
	var out strings.Builder
	units := 0
	for _, character := range text {
		if units == count {
			break
		}
		units += utf16.RuneLen(character)
		if units > count {
			character = utf8.RuneError
		}
		out.WriteRune(character)
	}
	return out.String()
}

func isWhitespace(character byte) bool {
	return character == ' ' || character == '\t' || character == '\n' || character == '\r'
}

func isDigit(character byte) bool {
	return '0' <= character && character <= '9'
}

func isIdentifierStart(character byte) bool {
	return 'a' <= character && character <= 'z' || 'A' <= character && character <= 'Z' ||
		character == '_' || character == '$'
}

func isIdentifierPart(character byte) bool {
	return isIdentifierStart(character) || isDigit(character)
}
