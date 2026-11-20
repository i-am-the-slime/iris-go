package tauri

import (
	"errors"
	"os"
	"reflect"
	"syscall"
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
)

func TestNumericWeightLeavesSizeAndFamilies(t *testing.T) {
	weight, ok, rest := splitNumericWeight("574 13px 'Victor Mono', ui-monospace, monospace")
	if !ok || weight != 574 || rest != "13px 'Victor Mono', ui-monospace, monospace" {
		t.Fatalf("got %d %v %q", weight, ok, rest)
	}
	if _, ok, rest := splitNumericWeight("bold 11.0px Ilisarniq"); ok || rest != "bold 11.0px Ilisarniq" {
		t.Fatalf("keyword weight was lifted: %q", rest)
	}
}

func TestFontShorthandAndFamilies(t *testing.T) {
	parsed, err := parseFontShorthand("italic bold 11.0px/1.2 'Victor Mono', ui-monospace, sans-serif")
	if err != nil {
		t.Fatal(err)
	}
	if parsed.style != "italic" || parsed.weight != "bold" || parsed.size != "11.0px" {
		t.Fatalf("parsed %+v", parsed)
	}
	families, err := parseFontFamilies(parsed.family)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"Victor Mono", "ui-monospace", "Arial"}; !reflect.DeepEqual(families, want) {
		t.Fatalf("families %q, want %q", families, want)
	}
	if size, _ := fontSize("12pt"); size != 16 {
		t.Fatalf("12pt is %v px", size)
	}
	if _, err := fontSize("2em"); err == nil {
		t.Fatal("relative sizes need a reference")
	}
}

func TestCSSFontMatching(t *testing.T) {
	demi := &faceInfo{families: []string{"Ilisarniq"}, weight: 500, stretch: 5}
	italic := &faceInfo{families: []string{"Ilisarniq-Italic", "Ilisarniq"}, weight: 400, style: styleItalic, stretch: 5}
	database := &fontDatabase{faces: []*faceInfo{demi, italic}}
	query := func(weight uint16, style faceStyle) *faceInfo {
		return database.query(fontQuery{families: []string{"Missing", "Ilisarniq"}, weight: weight, stretch: 5, style: style})
	}
	if query(700, styleNormal) != demi || query(400, styleItalic) != italic || query(500, styleOblique) != italic {
		t.Fatal("unexpected CSS font match")
	}
	if database.query(fontQuery{families: []string{"Missing"}, weight: 400, stretch: 5}) != nil {
		t.Fatal("an unknown family matched")
	}
}

func TestHostFailuresReadLikeRustIOErrors(t *testing.T) {
	defer func() {
		failure, ok := recover().(exception.Error)
		if !ok || failure.Message != "No such file or directory (os error 2)" {
			t.Fatalf("recovered %v", failure)
		}
	}()
	if rustErrno(syscall.ENOENT) != "No such file or directory (os error 2)" {
		t.Fatalf("got %q", rustErrno(syscall.ENOENT))
	}
	var errno syscall.Errno
	if _, err := os.ReadFile("/definitely/missing"); !errors.As(err, &errno) {
		t.Fatal("expected an errno")
	}
	ReadText("/definitely/missing")
}

func TestAbsolutePathIsLexical(t *testing.T) {
	directory, _ := os.Getwd()
	if got := AbsolutePath("a/./b/../c.svg"); got != directory+"/a/b/../c.svg" {
		t.Fatalf("got %q", got)
	}
	if got := AbsolutePath("/x//y/"); got != "/x/y/" {
		t.Fatalf("got %q", got)
	}
}
