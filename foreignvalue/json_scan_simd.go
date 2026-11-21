//go:build goexperiment.simd && arm64

package foreignvalue

import (
	"math/bits"
	"simd/archsimd"
	"unsafe"
)

// With SIMD the scans through string bodies read sixteen bytes at a time
// through NEON (simd/archsimd, an experimental API; this follows Go 1.27.0)
// and finish with the chunk that ends the string, overlapping bytes already
// read. Whole-input structural bitmasks (simdjson's first stage) do not fit a
// one-pass recursive descent that stops at every short token.

const jsonChunk = 16

// A string scan looks at its first scalarPrefix bytes one by one (in
// json_scan.go, inlined) and loads vectors only after them: on the benchmark
// documents, strings of up to about thirty bytes are scanned faster without.
const scalarPrefix = 32

// vectorStringSpecial is scalarStringSpecial sixteen bytes at a time.
func vectorStringSpecial(s string, at int) int {
	if len(s)-at < jsonChunk {
		return scalarStringSpecial(s, at)
	}
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	quote, backslash, space := archsimd.BroadcastUint8x16('"'), archsimd.BroadcastUint8x16('\\'), archsimd.BroadcastUint8x16(0x20)
	for ; at+jsonChunk <= len(s); at += jsonChunk {
		chunk := archsimd.LoadUint8x16(b[at : at+jsonChunk])
		found := chunk.Equal(quote).Or(chunk.Equal(backslash)).Or(chunk.Less(space))
		if offset, ok := firstLane(found); ok {
			return at + offset
		}
	}
	if at == len(s) {
		return at
	}
	last := archsimd.LoadUint8x16(b[len(s)-jsonChunk:])
	found := last.Equal(quote).Or(last.Equal(backslash)).Or(last.Less(space))
	return tailLane(found, s, at)
}

// vectorEscapeSpecial is scalarEscapeSpecial sixteen bytes at a time.
func vectorEscapeSpecial(s string, at int) int {
	if len(s)-at < jsonChunk {
		return scalarEscapeSpecial(s, at)
	}
	b := unsafe.Slice(unsafe.StringData(s), len(s))
	quote, backslash := archsimd.BroadcastUint8x16('"'), archsimd.BroadcastUint8x16('\\')
	// A byte is a control character or not ASCII exactly when it is below
	// 0x20 or above 0x7F, so one compare of byte-0x20 (wrapping) below 0x60
	// catches the safe ones.
	space, printable := archsimd.BroadcastUint8x16(0x20), archsimd.BroadcastUint8x16(0x7F-0x20)
	for ; at+jsonChunk <= len(s); at += jsonChunk {
		chunk := archsimd.LoadUint8x16(b[at : at+jsonChunk])
		found := chunk.Equal(quote).Or(chunk.Equal(backslash)).Or(chunk.Sub(space).Greater(printable))
		if offset, ok := firstLane(found); ok {
			return at + offset
		}
	}
	if at == len(s) {
		return at
	}
	last := archsimd.LoadUint8x16(b[len(s)-jsonChunk:])
	found := last.Equal(quote).Or(last.Equal(backslash)).Or(last.Sub(space).Greater(printable))
	return tailLane(found, s, at)
}

// Whitespace and digit runs stay scalar: in JSON they are a few bytes long,
// and the vector versions measured slower on the benchmark documents, even
// pretty-printed ones.
func whitespaceRunEnd(s string, at int) int {
	return scalarWhitespaceRunEnd(s, at)
}

func digitRunEnd(s string, at int) int {
	return scalarDigitRunEnd(s, at)
}

// tailLane is the first offset at or after at that the mask of the last
// chunk of s holds, which overlaps bytes before at, or len(s).
func tailLane(mask archsimd.Mask8x16, s string, at int) int {
	lanes := mask.ToInt8x16().ToBits().ReshapeToUint64s()
	low, high := lanes.GetElem(0), lanes.GetElem(1)
	start := len(s) - jsonChunk
	if skip := at - start; skip >= 8 {
		low, high = 0, high&(^uint64(0)<<(8*(skip-8)))
	} else {
		low &= ^uint64(0) << (8 * skip)
	}
	switch {
	case low != 0:
		return start + bits.TrailingZeros64(low)/8
	case high != 0:
		return start + 8 + bits.TrailingZeros64(high)/8
	}
	return len(s)
}

// firstLane is the first lane the mask holds.
func firstLane(mask archsimd.Mask8x16) (int, bool) {
	lanes := mask.ToInt8x16().ToBits().ReshapeToUint64s()
	if low := lanes.GetElem(0); low != 0 {
		return bits.TrailingZeros64(low) / 8, true
	}
	if high := lanes.GetElem(1); high != 0 {
		return 8 + bits.TrailingZeros64(high)/8, true
	}
	return 0, false
}
