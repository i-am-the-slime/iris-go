package syntax

import "testing"

func TestParseErrorsFillTheirTemplates(t *testing.T) {
	cases := map[string]string{
		`(?z)`:       "unrecognized grouping construct: (?z",
		`(a)\2`:      "reference to undefined group number 2",
		`\k<x>`:      "reference to undefined group name x",
		`(?(1)a|b)`:  "(?(1) ) reference to undefined group",
		`(?(1a)a|b)`: "(?(1) ) malformed",
		`\p{Foo}`:    "unknown unicode category, script, or property 'Foo'",
		`\q`:         `unrecognized escape sequence \q`,
		`[a-\d]`:     `cannot include class \100 in character range`,
		`[z-a]`:      "[z-a] range in reverse order",
		`[ž-a]`:      "[ž-a] range in reverse order",
		`(`:          "missing closing )",
	}
	for pattern, reason := range cases {
		_, err := Parse(pattern, 0)
		want := "error parsing regexp: " + reason + " in `" + pattern + "`"
		if err == nil || err.Error() != want {
			t.Errorf("Parse(%q) = %v, want %q", pattern, err, want)
		}
	}
}
