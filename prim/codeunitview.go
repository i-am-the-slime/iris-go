package prim

import (
	"math"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// CodeUnitsView indexes a string by UTF-16 code units in constant time. The
// compiler builds one before a loop that indexes a string the loop never
// changes, where the code-unit operations would each walk the string from its
// start.
//
// Every method returns exactly what the code-unit operation it is named after
// returns for the viewed string, bytes included: a cut between the two
// surrogates of an astral character leaves a lone surrogate on each side, and
// a string holding invalid UTF-8 or a lone surrogate is not indexed at all but
// walked by those operations, so that their reading of it stays the only one.
type CodeUnitsView struct {
	text        string
	asciiLength uint
	form        viewForm
	// The code units of text, which a walk counts once, when the view is
	// built. It is CodeUnitLength(text), also for text too long for
	// int32 offsets, where no method but Length reads it.
	length int
	// For wide text, the code units, and for each the byte offset of the
	// character it belongs to, followed by len(text).
	units  []uint16
	starts []int32
}

type viewForm uint8

const (
	// Every code unit is one byte of text, at its own offset.
	asciiForm viewForm = iota
	wideForm
	// Text holding invalid UTF-8 or a lone surrogate, or too long for int32
	// offsets.
	walkedForm
)

// UTF16 views text by its UTF-16 code units.
func UTF16(text string) CodeUnitsView {
	length, valid := codeUnitLengthUpTo(text, len(text))
	switch {
	case valid && length == len(text):
		return CodeUnitsView{text: text, asciiLength: uint(length), form: asciiForm, length: length}
	case !valid || len(text) > math.MaxInt32 || HasSurrogates(text):
		return CodeUnitsView{text: text, form: walkedForm, length: length}
	}
	units, starts := wideCodeUnits(text, length)
	return CodeUnitsView{text: text, form: wideForm, length: int(length), units: units, starts: starts}
}

// scalarWideCodeUnits is the code units of valid UTF-8 text holding length of
// them, and for each the byte offset of the character it belongs to, followed
// by len(text), read a character at a time. A build with SIMD widens runs of
// ASCII sixteen bytes at a time.
func scalarWideCodeUnits(text string, length int) (units []uint16, starts []int32) {
	units = make([]uint16, 0, length)
	starts = make([]int32, 0, length+1)
	for offset, r := range text {
		if utf16.RuneLen(r) == 2 {
			high, low := utf16.EncodeRune(r)
			units = append(units, uint16(high), uint16(low))
			starts = append(starts, int32(offset), int32(offset))
			continue
		}
		units = append(units, uint16(r))
		starts = append(starts, int32(offset))
	}
	starts = append(starts, int32(len(text)))
	return units, starts
}

// ASCII is the viewed text when each of its code units is one of its bytes,
// and "" otherwise. A loop that indexes the view reads a code unit below
// len(view.ASCII()) as that byte, without a call.
func (view *CodeUnitsView) ASCII() string {
	if view.form == asciiForm {
		return view.text
	}
	return ""
}

// Length is CodeUnitLength.
func (view *CodeUnitsView) Length() int {
	return view.length
}

// CharAt is CodeUnitCharAt.
func (view *CodeUnitsView) CharAt(i int) (rune, bool) {
	return view.unitAt(int(i))
}

// UnsafeCharAt is CharAt, the code unit of Data.String.Unsafe.charAt.
func (view *CodeUnitsView) UnsafeCharAt(i int) rune {
	unit, ok := view.unitAt(int(i))
	if !ok {
		panic("Data.String.Unsafe.charAt: Invalid index.")
	}
	return unit
}

// Take is CodeUnitTake.
func (view *CodeUnitsView) Take(n int) string {
	before, _ := view.split(int(n))
	return before
}

// Drop is CodeUnitDrop.
func (view *CodeUnitsView) Drop(n int) string {
	_, after := view.split(int(n))
	return after
}

// Slice is CodeUnitSlice: Drop(begin, Take(end, text)), read off the view.
func (view *CodeUnitsView) Slice(begin int, end int) string {
	if uint(begin) <= uint(end) && uint(end) <= view.asciiLength {
		return view.text[begin:end]
	}
	return view.slice(begin, end)
}

func (view *CodeUnitsView) slice(begin32 int, end32 int) string {
	begin, end := int(begin32), int(end32)
	if view.form == walkedForm {
		return sliceCodeUnits(begin, end, view.text)
	}
	length := int(view.length)
	if begin < 0 || end < 0 {
		begin, end = sliceIndex(begin, length), sliceIndex(end, length)
	}
	begin, end = clampCodeUnit(begin, length), clampCodeUnit(end, length)
	if begin >= end {
		return ""
	}
	_, after := view.cut(begin)
	before, _ := view.cut(end)
	if after.prefix == "" && before.suffix == "" {
		return view.text[after.offset:before.offset]
	}
	return after.prefix + view.text[after.offset:before.offset] + before.suffix
}

// IndexOfStartingAt is CodeUnitIndexOfStartingAt.
func (view *CodeUnitsView) IndexOfStartingAt(pattern string, startAt int) (int, bool) {
	if view.form == walkedForm || !utf8.ValidString(pattern) {
		return CodeUnitIndexOfStartingAt(pattern, startAt, view.text)
	}
	if startAt < 0 || startAt > view.length {
		return 0, false
	}
	if pattern == "" {
		return startAt, true
	}
	_, after := view.cut(int(startAt))
	at := strings.Index(view.text[after.offset:], pattern)
	if at < 0 {
		return 0, false
	}
	return int(view.unitStartingAt(after.offset + at)), true
}

// LastIndexOfStartingAt is CodeUnitLastIndexOfStartingAt.
func (view *CodeUnitsView) LastIndexOfStartingAt(pattern string, startAt int) (int, bool) {
	if view.form == walkedForm || !utf8.ValidString(pattern) {
		return CodeUnitLastIndexOfStartingAt(pattern, startAt, view.text)
	}
	from := clampCodeUnit(int(startAt), int(view.length))
	if pattern == "" {
		return int(from), true
	}
	offset := view.start(from)
	at := strings.LastIndex(view.text[:min(len(view.text), offset+len(pattern))], pattern)
	if at < 0 {
		return 0, false
	}
	return int(view.unitStartingAt(at)), true
}

// The methods above rest on four readings of the view, for code unit
// positions already known to lie in [0, Length()] (and below it for unitAt).

func (view *CodeUnitsView) unitAt(i int) (rune, bool) {
	switch view.form {
	case asciiForm:
		if i < 0 || i >= len(view.text) {
			return 0, false
		}
		return rune(view.text[i]), true
	case wideForm:
		if i < 0 || i >= len(view.units) {
			return 0, false
		}
		return rune(view.units[i]), true
	default:
		return codeUnitAt(view.text, i)
	}
}

// start is the byte offset of the character holding code unit n.
func (view *CodeUnitsView) start(n int) int {
	if view.form == asciiForm {
		return n
	}
	return int(view.starts[n])
}

// unitStartingAt is the first code unit of the character at a byte offset.
func (view *CodeUnitsView) unitStartingAt(offset int) int {
	if view.form == asciiForm {
		return offset
	}
	unit, _ := slices.BinarySearch(view.starts, int32(offset))
	return unit
}

// splitsPair tells whether position n lies between the two surrogates of an
// astral character. A viewed text holds no lone surrogate, so a trailing
// surrogate always ends a pair.
func (view *CodeUnitsView) splitsPair(n int) bool {
	if view.form != wideForm || n >= len(view.units) {
		return false
	}
	unit := view.units[n]
	return unit >= 0xDC00 && unit <= 0xDFFF
}

// boundary is one side of a cut: the lone surrogate a split pair leaves, and where
// the text on that side ends (for the part before) or starts (for the part
// after).
type boundary struct {
	prefix string
	suffix string
	offset int
}

// cut is the part before and the part after position n, which lies in
// [0, Length()], as splitCodeUnits builds them.
func (view *CodeUnitsView) cut(n int) (before boundary, after boundary) {
	offset := view.start(n)
	if !view.splitsPair(n) {
		return boundary{offset: offset}, boundary{offset: offset}
	}
	return boundary{suffix: RuneString(rune(view.units[n-1])), offset: offset},
		boundary{prefix: RuneString(rune(view.units[n])), offset: view.start(n + 1)}
}

// split is splitCodeUnits.
func (view *CodeUnitsView) split(n int) (before string, after string) {
	if view.form == walkedForm {
		return splitCodeUnits(view.text, n)
	}
	head, tail := view.cut(clampCodeUnit(n, int(view.length)))
	return view.text[:head.offset] + head.suffix, tail.prefix + view.text[tail.offset:]
}
