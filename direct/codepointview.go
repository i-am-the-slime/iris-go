package direct

import (
	"math"
	"unicode/utf8"

	"github.com/i-am-the-slime/iris-go/prim"
)

// CodePointsView indexes a string by code points in constant time. The
// compiler builds one before a loop that indexes a string the loop never
// changes, where the code point operations would each walk the string from
// its start.
//
// Every method returns exactly what the code point operation it is named
// after returns for the viewed string: the view reads the string as those
// operations range over it, each invalid byte one U+FFFD.
type CodePointsView struct {
	text  string
	form  pointsForm
	valid bool
	// For wide text, the code points, and the byte offset at which each
	// starts.
	points []int32
	starts []int32
}

type pointsForm uint8

const (
	// Every code point is one byte of text, at its own offset.
	asciiPoints pointsForm = iota
	widePoints
	// Text holding a lone surrogate, or too long for int32 offsets, which
	// the operations walk.
	walkedPoints
)

// CodePoints views text by its code points.
func CodePoints(text string) CodePointsView {
	if len(text) > math.MaxInt32 || prim.HasSurrogates(text) {
		return CodePointsView{text: text, form: walkedPoints}
	}
	if prim.IsASCII(text) {
		return CodePointsView{text: text, form: asciiPoints, valid: true}
	}
	count := utf8.RuneCountInString(text)
	valid := utf8.ValidString(text)
	points := make([]int32, 0, count)
	starts := make([]int32, 0, count)
	for offset, point := range text {
		points = append(points, point)
		starts = append(starts, int32(offset))
	}
	return CodePointsView{text: text, form: widePoints, valid: valid, points: points, starts: starts}
}

// Length is CodePointLength.
func (view *CodePointsView) Length() int32 {
	switch view.form {
	case asciiPoints:
		return int32(len(view.text))
	case widePoints:
		return int32(len(view.points))
	default:
		return CodePointLength(view.text)
	}
}

// CodePointAt is CodePointAt.
func (view *CodePointsView) CodePointAt[R any](just func(int32) R, nothing R, index int32) R {
	switch {
	case view.form == walkedPoints:
		return CodePointAt(nil, just, nothing, nil, index, view.text)
	case index < 0 || index >= view.Length():
		return nothing
	case view.form == asciiPoints:
		return just(int32(view.text[index]))
	default:
		return just(view.points[index])
	}
}

// Take is CodePointTake.
func (view *CodePointsView) Take(count int32) string {
	switch {
	case view.form == walkedPoints:
		return CodePointTake(nil, count, view.text)
	case count <= 0:
		return ""
	case count >= view.Length():
		return view.text
	case view.form == asciiPoints:
		return view.text[:count]
	case view.valid:
		return view.text[:view.starts[count]]
	default:
		return validPrefix(view.text[:view.starts[count]])
	}
}
