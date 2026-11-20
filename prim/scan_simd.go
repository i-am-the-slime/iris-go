//go:build goexperiment.simd && arm64

package prim

import (
	"math"
	"math/bits"
	"simd/archsimd"
	"unicode/utf16"
	"unicode/utf8"
	"unsafe"
)

// With SIMD the string scans read sixteen bytes at a time through NEON
// (simd/archsimd, an experimental API; this follows Go 1.27.0). The code-unit
// walks start after a prefix of whole characters that is counted in code
// units and validated as UTF-8 in vector registers, and finish the rest a
// character at a time: the bytes after the last whole chunk, and the last
// code units before a limit. A string that holds invalid UTF-8 in that prefix
// is walked from its start, so that the walk's reading of invalid bytes stays
// the only one.

// wideCodeUnits is scalarWideCodeUnits, widening each run of ASCII sixteen
// bytes at a time.
func wideCodeUnits(text string, length int) (units []uint16, starts []int32) {
	units = make([]uint16, length)
	starts = make([]int32, length+1)
	bytes := stringBytes(text)
	unit := 0
	for offset := 0; offset < len(text); {
		if text[offset] < utf8.RuneSelf && offset+chunkSize <= len(text) {
			chunk := archsimd.LoadUint8x16(bytes[offset : offset+chunkSize])
			run := asciiRun(chunk)
			if run == chunkSize {
				widenAscii(chunk, units[unit:unit+chunkSize], starts[unit:unit+chunkSize], offset)
			} else {
				for at := range run {
					units[unit+at] = uint16(text[offset+at])
					starts[unit+at] = int32(offset + at)
				}
			}
			unit += run
			offset += run
			continue
		}
		r, size := utf8.DecodeRuneInString(text[offset:])
		if utf16.RuneLen(r) == 2 {
			high, low := utf16.EncodeRune(r)
			units[unit], units[unit+1] = uint16(high), uint16(low)
			starts[unit], starts[unit+1] = int32(offset), int32(offset)
			unit += 2
		} else {
			units[unit] = uint16(r)
			starts[unit] = int32(offset)
			unit++
		}
		offset += size
	}
	starts[length] = int32(len(text))
	return units, starts
}

// A chunk is one vector of bytes; a block is four, which runs of ASCII skip
// at once.
const (
	chunkSize = 16
	blockSize = 4 * chunkSize
)

// wholeCharacterPrefix is the longest prefix of s, in whole chunks less a
// character the last chunk cuts short, that holds at most limit code units,
// as its length in bytes and in code units; or no prefix, where those bytes
// are not valid UTF-8, or where s or the limit is shorter than a block and
// setting up the vectors would cost more than they save.
//
// A chunk adds a code unit for each byte that starts a character and another
// for each that starts a four-byte one, an astral character with two
// surrogates. The count is known only at the end of a run of chunks, so each
// run takes no more chunks than the limit still allows if every byte were a
// code unit. A run passes the limit only by cutting a four-byte character
// after its first byte, and then it is the last run, and the character is
// dropped.
func wholeCharacterPrefix(s string, limit int) (offset int, units int) {
	if min(len(s), limit) >= blockSize {
		offset, units = vectorPrefix(s, limit)
	}
	return offset, units
}

func vectorPrefix(s string, limit int) (offset int, units int) {
	if limit >= len(s) {
		// No prefix of s holds more code units than bytes.
		limit = math.MaxInt
	}
	bytes := stringBytes(s)
	tables := loadUtf8Tables()
	var previous, unfinished, errors archsimd.Uint8x16
	for {
		chunks := min(limit-units, len(s)-offset) / chunkSize
		if chunks <= 0 {
			break
		}
		var counted archsimd.Uint8x16
		pending := 0
		end := offset + chunks*chunkSize
		for offset < end {
			if offset+blockSize <= end && asciiBlock(bytes[offset:offset+blockSize]) {
				errors = errors.Or(unfinished)
				unfinished = archsimd.Uint8x16{}
				previous = archsimd.LoadUint8x16(bytes[offset+blockSize-chunkSize : offset+blockSize])
				units += blockSize
				offset += blockSize
				continue
			}
			for stop := min(offset+blockSize, end); offset < stop; offset += chunkSize {
				chunk := archsimd.LoadUint8x16(bytes[offset : offset+chunkSize])
				if chunk.ReduceMax() < utf8.RuneSelf {
					errors = errors.Or(unfinished)
					unfinished = archsimd.Uint8x16{}
					units += chunkSize
					previous = chunk
					continue
				}
				errors = errors.Or(tables.errors(chunk, previous))
				unfinished = chunk.SubSaturated(tables.lastCharacterStarts)
				counted = counted.Sub(maskBytes(chunk.BitsToInt8().Greater(tables.lastContinuation)))
				counted = counted.Sub(maskBytes(chunk.GreaterEqual(tables.fourByteStart)))
				previous = chunk
				pending++
				if pending == maxPendingChunks {
					units += sumBytes(counted)
					counted = archsimd.Uint8x16{}
					pending = 0
				}
			}
		}
		units += sumBytes(counted)
	}
	if errors.ReduceMax() != 0 {
		return 0, 0
	}
	return beforeCutCharacter(s, offset, units)
}

func asciiBlock(block []byte) bool {
	first := archsimd.LoadUint8x16(block[:chunkSize])
	second := archsimd.LoadUint8x16(block[chunkSize : 2*chunkSize])
	third := archsimd.LoadUint8x16(block[2*chunkSize : 3*chunkSize])
	fourth := archsimd.LoadUint8x16(block[3*chunkSize : blockSize])
	return first.Or(second).Or(third.Or(fourth)).ReduceMax() < utf8.RuneSelf
}

// beforeCutCharacter moves the end of a valid prefix that ends inside a
// character back to that character's start.
func beforeCutCharacter(s string, offset int, units int) (int, int) {
	for back := 1; back <= min(utf8.UTFMax-1, offset); back++ {
		b := s[offset-back]
		if b < utf8.RuneSelf {
			break
		}
		if b < 0xC0 {
			continue
		}
		if back < sequenceLength(b) {
			offset -= back
			units -= startedUnits(b)
		}
		break
	}
	return offset, units
}

func sequenceLength(start byte) int {
	switch {
	case start < 0xE0:
		return 2
	case start < 0xF0:
		return 3
	default:
		return 4
	}
}

// startedUnits is the code units counted for a byte that starts a character.
func startedUnits(start byte) int {
	if start >= 0xF0 {
		return 2
	}
	return 1
}

// A lane of the count grows by at most two a chunk, so it is summed before
// it can pass 255.
const maxPendingChunks = 127

func sumBytes(counted archsimd.Uint8x16) int {
	low := counted.ExtendLo8ToUint16()
	high := counted.InterleaveHi(archsimd.Uint8x16{}).ReshapeToUint16s()
	return int(low.Add(high).ReduceSum())
}

// maskBytes is 0xFF in each lane the mask holds and 0 elsewhere.
func maskBytes(mask archsimd.Mask8x16) archsimd.Uint8x16 {
	return mask.ToInt8x16().ToBits()
}

// asciiRun is the number of bytes the chunk starts with that are ASCII.
func asciiRun(chunk archsimd.Uint8x16) int {
	high := maskBytes(chunk.GreaterEqual(archsimd.BroadcastUint8x16(utf8.RuneSelf))).ReshapeToUint64s()
	if low := high.GetElem(0); low != 0 {
		return bits.TrailingZeros64(low) / 8
	}
	if upper := high.GetElem(1); upper != 0 {
		return 8 + bits.TrailingZeros64(upper)/8
	}
	return chunkSize
}

// widenAscii writes a chunk of ASCII text starting at offset as its code
// units, and for each its own offset.
func widenAscii(chunk archsimd.Uint8x16, units []uint16, starts []int32, offset int) {
	chunk.ExtendLo8ToUint16().Store(units[:8])
	chunk.InterleaveHi(archsimd.Uint8x16{}).ReshapeToUint16s().Store(units[8:16])
	first := archsimd.BroadcastInt32x4(int32(offset)).Add(archsimd.LoadInt32x4Array(&firstFourOffsets))
	four := archsimd.BroadcastInt32x4(4)
	for quarter := 0; quarter < chunkSize; quarter += 4 {
		first.Store(starts[quarter : quarter+4])
		first = first.Add(four)
	}
}

var firstFourOffsets = [4]int32{0, 1, 2, 3}

// stringBytes reads a string's bytes as a slice for vector loads, which take
// slices; nothing writes through it.
func stringBytes(s string) []byte {
	return unsafe.Slice(unsafe.StringData(s), len(s))
}

// UTF-8 is validated by classifying each pair of adjacent bytes through three
// table lookups, by the high nibble of the first, its low nibble and the high
// nibble of the second, whose AND is nonzero exactly where the pair is
// invalid; and by requiring a continuation byte two and three bytes after
// the start of a three- and four-byte character. This is the lookup algorithm
// of Keiser and Lemire, "Validating UTF-8 In Less Than One Instruction Per
// Byte" (2021), which accepts exactly what unicode/utf8 accepts.
const (
	tooShort     = 1 << 0 // a start or ASCII byte after a start byte
	tooLong      = 1 << 1 // a continuation byte after ASCII
	overlong3    = 1 << 2 // E0 followed by 80..9F
	tooLarge     = 1 << 3 // F4 followed by 90..BF, or F5..FF
	surrogate    = 1 << 4 // ED followed by A0..BF
	overlong2    = 1 << 5 // C0 or C1
	tooLarge1000 = 1 << 6 // F5..FF followed by 80..8F
	overlong4    = 1 << 6 // F0 followed by 80..8F
	twoContinues = 1 << 7 // a continuation byte after a continuation byte
	carry        = tooShort | tooLong | twoContinues
)

var firstHighNibble = [16]uint8{
	tooLong, tooLong, tooLong, tooLong, tooLong, tooLong, tooLong, tooLong,
	twoContinues, twoContinues, twoContinues, twoContinues,
	tooShort | overlong2,
	tooShort,
	tooShort | overlong3 | surrogate,
	tooShort | tooLarge | tooLarge1000 | overlong4,
}

var firstLowNibble = [16]uint8{
	carry | overlong3 | overlong2 | overlong4,
	carry | overlong2,
	carry,
	carry,
	carry | tooLarge,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000 | surrogate,
	carry | tooLarge | tooLarge1000,
	carry | tooLarge | tooLarge1000,
}

var secondHighNibble = [16]uint8{
	tooShort, tooShort, tooShort, tooShort, tooShort, tooShort, tooShort, tooShort,
	tooLong | overlong2 | twoContinues | overlong3 | tooLarge1000 | overlong4,
	tooLong | overlong2 | twoContinues | overlong3 | tooLarge,
	tooLong | overlong2 | twoContinues | surrogate | tooLarge,
	tooLong | overlong2 | twoContinues | surrogate | tooLarge,
	tooShort, tooShort, tooShort, tooShort,
}

// The largest byte that may stand at each position of a chunk with no
// character running past its end: three, two and one bytes from the end, no
// start of a four-, three- or two-byte character.
var lastCharacterStarts = [16]uint8{
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF,
	0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xEF, 0xDF, 0xBF,
}

type utf8Tables struct {
	firstHighNibble     archsimd.Uint8x16
	firstLowNibble      archsimd.Uint8x16
	secondHighNibble    archsimd.Uint8x16
	lastCharacterStarts archsimd.Uint8x16
	lowNibble           archsimd.Uint8x16
	lastContinuation    archsimd.Int8x16
	fourByteStart       archsimd.Uint8x16
	thirdByteStart      archsimd.Uint8x16
	fourthByteStart     archsimd.Uint8x16
	highBit             archsimd.Uint8x16
}

func loadUtf8Tables() utf8Tables {
	return utf8Tables{
		firstHighNibble:     archsimd.LoadUint8x16Array(&firstHighNibble),
		firstLowNibble:      archsimd.LoadUint8x16Array(&firstLowNibble),
		secondHighNibble:    archsimd.LoadUint8x16Array(&secondHighNibble),
		lastCharacterStarts: archsimd.LoadUint8x16Array(&lastCharacterStarts),
		lowNibble:           archsimd.BroadcastUint8x16(0x0F),
		lastContinuation:    archsimd.BroadcastInt8x16(-0x41), // 0xBF
		fourByteStart:       archsimd.BroadcastUint8x16(0xF0),
		thirdByteStart:      archsimd.BroadcastUint8x16(0xE0 - 0x80),
		fourthByteStart:     archsimd.BroadcastUint8x16(0xF0 - 0x80),
		highBit:             archsimd.BroadcastUint8x16(0x80),
	}
}

// errors is nonzero in each lane of chunk that is not valid UTF-8 after the
// chunk before it.
func (tables *utf8Tables) errors(chunk archsimd.Uint8x16, previous archsimd.Uint8x16) archsimd.Uint8x16 {
	before1 := chunk.ConcatShiftBytesRight(previous, chunkSize-1)
	before2 := chunk.ConcatShiftBytesRight(previous, chunkSize-2)
	before3 := chunk.ConcatShiftBytesRight(previous, chunkSize-3)
	pairs := tables.firstHighNibble.LookupOrZero(before1.ShiftAllRight(4)).
		And(tables.firstLowNibble.LookupOrZero(before1.And(tables.lowNibble))).
		And(tables.secondHighNibble.LookupOrZero(chunk.ShiftAllRight(4)))
	continues := before2.SubSaturated(tables.thirdByteStart).
		Or(before3.SubSaturated(tables.fourthByteStart)).
		And(tables.highBit)
	return continues.Xor(pairs)
}
