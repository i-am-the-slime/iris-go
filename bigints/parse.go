package bigints

import (
	"math"
	"math/big"
	"strconv"
	"strings"
	"unicode"

	"github.com/i-am-the-slime/iris-go/exception"
	"github.com/i-am-the-slime/iris-go/prim"
)

// maxSafe is big-integer's MAX_INT, 2^53: below it a Number holds every
// integer exactly.
const maxSafe = 9007199254740992

func invalid(text string) *exception.Error {
	return &exception.Error{Message: "Invalid integer: " + text}
}

// parseValue is big-integer's parseStringValue: a string that JavaScript's
// Number conversion reads as a safe integer is that integer; any other string
// must be decimal digits, after an optional minus sign, with an optional
// non-negative exponent that the digits after a decimal point count against.
func parseValue(text string) (*big.Int, *exception.Error) {
	if number := stringToNumber(text); -maxSafe < number && number < maxSafe {
		if number != math.Trunc(number) {
			return nil, invalid(text)
		}
		return big.NewInt(int64(number)), nil
	}
	negative := strings.HasPrefix(text, "-")
	if negative {
		text = text[1:]
	}
	parts := splitExponent(text)
	if len(parts) > 2 {
		return nil, invalid(strings.Join(parts, "e"))
	}
	if len(parts) == 2 {
		digits, failure := withExponent(parts[0], parts[1])
		if failure != nil {
			return nil, failure
		}
		text = digits
	}
	if text == "" || strings.TrimLeft(text, "0123456789") != "" {
		return nil, invalid(text)
	}
	value, _ := new(big.Int).SetString(text, 10)
	if negative {
		value.Neg(value)
	}
	return value, nil
}

// splitExponent is `text.split(/e/i)`, empty parts included.
func splitExponent(text string) []string {
	return strings.Split(strings.ReplaceAll(text, "E", "e"), "e")
}

// withExponent writes a mantissa and its exponent as plain digits, as
// big-integer does: `exponent` may carry a plus sign and must be an integer,
// and the zeros it appends must make an array JavaScript can allocate.
func withExponent(mantissa string, exponent string) (string, *exception.Error) {
	exponent = strings.TrimPrefix(exponent, "+")
	shift := stringToNumber(exponent)
	if shift != math.Trunc(shift) || !(-maxSafe < shift && shift < maxSafe) {
		return "", &exception.Error{Message: "Invalid integer: " + prim.NumberToString(shift) + " is not a valid exponent."}
	}
	if point := strings.IndexByte(mantissa, '.'); point >= 0 {
		shift -= float64(len(mantissa) - point - 1)
		mantissa = mantissa[:point] + mantissa[point+1:]
	}
	if shift < 0 {
		return "", &exception.Error{Message: "Cannot include negative exponent part for integers"}
	}
	if shift+1 > math.MaxUint32 {
		return "", &exception.Error{Name: "RangeError", Message: "Array length must be a positive integer of safe magnitude."}
	}
	return mantissa + strings.Repeat("0", int(shift)), nil
}

// parseBase is big-integer's parseBase for the default alphabet, case
// insensitively: digits of the alphabet below the base's magnitude, `<n>`
// for any other digit, and a leading minus sign.
func parseBase(text string, base int64) (*big.Int, *exception.Error) {
	text = lowerCase(text)
	magnitude := base
	if magnitude < 0 {
		magnitude = -magnitude
	}
	for _, character := range text {
		value := strings.IndexRune(alphabet, character)
		if value >= 0 && int64(value) >= magnitude && !(character == '1' && magnitude == 1) {
			return nil, &exception.Error{Message: string(character) + " is not a valid digit in base " + strconv.FormatInt(base, 10) + "."}
		}
	}
	negative := strings.HasPrefix(text, "-")
	rest := text
	if negative {
		rest = rest[1:]
	}
	var digits []*big.Int
	for index := 0; index < len(rest); {
		character := rest[index]
		if value := strings.IndexByte(alphabet, character); value >= 0 {
			digits = append(digits, big.NewInt(int64(value)))
			index++
			continue
		}
		if character != '<' {
			return nil, &exception.Error{Message: "not a valid character"}
		}
		end := strings.IndexByte(rest[index:], '>')
		inner := rest[index+1:]
		if end >= 0 {
			inner = rest[index+1 : index+end]
		}
		digit, failure := parseValue(inner)
		if failure != nil {
			return nil, failure
		}
		digits = append(digits, digit)
		if end < 0 {
			break
		}
		index += end + 1
	}
	value, power, radix := new(big.Int), big.NewInt(1), big.NewInt(base)
	for index := len(digits) - 1; index >= 0; index-- {
		value.Add(value, new(big.Int).Mul(digits[index], power))
		power.Mul(power, radix)
	}
	if negative {
		value.Neg(value)
	}
	return value, nil
}

// lowerCase is JavaScript's `toLowerCase`, as far as it decides which
// characters are digits: U+0130 lowers to two characters there.
func lowerCase(text string) string {
	return strings.Map(func(r rune) rune {
		if r == 'İ' {
			return -1
		}
		return unicode.ToLower(r)
	}, strings.ReplaceAll(text, "İ", "i̇"))
}

// stringToNumber is JavaScript's Number conversion of a string: surrounding
// white space is ignored, an empty string is zero, and the rest must be a
// decimal literal (with an optional sign, or Infinity) or an unsigned 0x, 0o
// or 0b integer; anything else is NaN.
func stringToNumber(text string) float64 {
	text = strings.TrimFunc(text, isWhitespace)
	if text == "" {
		return 0
	}
	if len(text) > 2 && text[0] == '0' {
		if base := prefixBase(text[1]); base != 0 {
			value, ok := new(big.Int).SetString(text[2:], base)
			if !ok || strings.ContainsAny(text[2:], "_+-") {
				return math.NaN()
			}
			number, _ := new(big.Float).SetInt(value).Float64()
			return number
		}
	}
	unsigned := strings.TrimLeft(text[:1], "+-") + text[1:]
	if unsigned == "Infinity" {
		if text[0] == '-' {
			return math.Inf(-1)
		}
		return math.Inf(1)
	}
	if !isDecimalLiteral(unsigned) {
		return math.NaN()
	}
	number, _ := strconv.ParseFloat(text, 64)
	return number
}

func prefixBase(letter byte) int {
	switch letter {
	case 'x', 'X':
		return 16
	case 'o', 'O':
		return 8
	case 'b', 'B':
		return 2
	}
	return 0
}

// isDecimalLiteral is StrUnsignedDecimalLiteral without Infinity: digits with
// an optional fraction, or a fraction alone, then an optional exponent.
func isDecimalLiteral(text string) bool {
	integer := digitRun(text, 0)
	position := integer
	fraction := 0
	if position < len(text) && text[position] == '.' {
		fraction = digitRun(text, position+1) - position - 1
		position += 1 + fraction
	}
	if integer == 0 && fraction == 0 {
		return false
	}
	if position < len(text) && (text[position] == 'e' || text[position] == 'E') {
		position++
		if position < len(text) && (text[position] == '+' || text[position] == '-') {
			position++
		}
		exponent := digitRun(text, position)
		if exponent == position {
			return false
		}
		position = exponent
	}
	return position == len(text)
}

func digitRun(text string, from int) int {
	for from < len(text) && '0' <= text[from] && text[from] <= '9' {
		from++
	}
	return from
}

// isWhitespace is ECMAScript WhiteSpace or LineTerminator.
func isWhitespace(r rune) bool {
	switch r {
	case '\t', '\n', '\v', '\f', '\r', ' ', '\u00a0', '\u2028', '\u2029', '\ufeff':
		return true
	}
	return unicode.Is(unicode.Zs, r)
}
