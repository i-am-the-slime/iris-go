package regex

import (
	"strings"
)

// substitute writes the replacement with its patterns expanded for one match,
// following ECMAScript's GetSubstitution. A pattern naming a group that does
// not exist is written literally; a group that did not participate is empty.
func substitute(out *strings.Builder, rx *Regex, replacement string, value string, match []int) {
	groups := len(match)/2 - 1
	group := func(index int) string {
		if match[2*index] < 0 {
			return ""
		}
		return value[match[2*index]:match[2*index+1]]
	}
	for position := 0; position < len(replacement); position++ {
		character := replacement[position]
		if character != '$' || position+1 == len(replacement) {
			out.WriteByte(character)
			continue
		}
		next := replacement[position+1]
		switch {
		case next == '$':
			out.WriteByte('$')
			position++
		case next == '&':
			out.WriteString(value[match[0]:match[1]])
			position++
		case next == '`':
			out.WriteString(value[:match[0]])
			position++
		case next == '\'':
			out.WriteString(value[match[1]:])
			position++
		case isDigit(next):
			index, width := groupReference(replacement[position+1:], groups)
			if width == 0 {
				out.WriteByte('$')
				continue
			}
			out.WriteString(group(index))
			position += width
		case next == '<' && hasNamedGroups(rx.pattern):
			end := strings.IndexByte(replacement[position+2:], '>')
			if end < 0 {
				out.WriteByte('$')
				continue
			}
			if index := rx.pattern.SubexpIndex(replacement[position+2 : position+2+end]); index > 0 {
				out.WriteString(group(index))
			}
			position += end + 2
		default:
			out.WriteByte('$')
		}
	}
}

// groupReference reads `n` or `nn` after a `$`: the two-digit group when it
// exists, else the one-digit group, else no reference (width 0).
func groupReference(digits string, groups int) (index int, width int) {
	if len(digits) >= 2 && isDigit(digits[1]) {
		if two := int(digits[0]-'0')*10 + int(digits[1]-'0'); two >= 1 && two <= groups {
			return two, 2
		}
	}
	if one := int(digits[0] - '0'); one >= 1 && one <= groups {
		return one, 1
	}
	return 0, 0
}

func hasNamedGroups(re engine) bool {
	for _, name := range re.SubexpNames() {
		if name != "" {
			return true
		}
	}
	return false
}

func isDigit(character byte) bool {
	return character >= '0' && character <= '9'
}
