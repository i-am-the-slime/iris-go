package Data_Int64_Internal_ffi

import (
	"math"
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
)

func TestDivisionWrapsLikeTwosComplement(t *testing.T) {
	if got := Divide_(int64(math.MinInt64), -1); got != math.MinInt64 {
		t.Fatalf("min / -1 = %v", got)
	}
	if got := Modulo_(int64(math.MinInt64), -1); got != 0 {
		t.Fatalf("min %% -1 = %v", got)
	}
}

func TestDivisionReadsTheLongsSignedness(t *testing.T) {
	if got := ToString_(Divide_(uint64(math.MaxUint64), 10), 10); got != "1844674407370955161" {
		t.Fatalf("unsigned max / 10 = %s", got)
	}
	if got := ToString_(Divide_(int64(-7), 2), 10); got != "-3" {
		t.Fatalf("-7 / 2 = %s", got)
	}
	if got := Modulo_(int64(-7), 2); got != -1 {
		t.Fatalf("-7 %% 2 = %d", got)
	}
}

func TestDivisionByZeroThrows(t *testing.T) {
	defer func() {
		if thrown, ok := recover().(exception.Error); !ok || thrown.Message != "division by zero" {
			t.Fatalf("recovered %v", thrown)
		}
	}()
	Divide_(uint64(1), 0)
}

func TestCompareFollowsSignedness(t *testing.T) {
	if Compare_(int64(-1), 1) != -1 || Compare_(uint64(math.MaxUint64), 1) != 1 || Compare_(int64(3), 3) != 0 {
		t.Fatal("comparisons")
	}
}

func TestStringsInEveryRadix(t *testing.T) {
	if got := ToString_(int64(math.MinInt64), 16); got != "-8000000000000000" {
		t.Fatalf("min in hex = %s", got)
	}
	if got := ToString_(uint64(math.MaxUint64), 36); got != "3w5e11264sgsf" {
		t.Fatalf("unsigned max in base 36 = %s", got)
	}
	if got := SafeReadLong_UInt64("-000", true, 10); got == nil || *got != 0 {
		t.Fatalf("-000 unsigned = %v", got)
	}
	if got := SafeReadLong_Int64("1Z141Z3", false, 36); got == nil || ToString_(*got, 36) != "1z141z3" {
		t.Fatalf("1Z141Z3 = %v", got)
	}
	if got := SafeReadLong_Int64("1Z00000000000", false, 36); got != nil {
		t.Fatalf("uppercase past signed max = %v", got)
	}
	if got := SafeReadLong_UInt64("3W5E11264SGSG", true, 36); got != nil {
		t.Fatalf("one past unsigned max = %v", got)
	}
	if got := SafeReadLong_Int64("-9223372036854775808", false, 10); got == nil || *got != math.MinInt64 {
		t.Fatalf("signed min = %v", got)
	}
	if got := FromString_Int64("-ff", false, 16); got != -255 {
		t.Fatalf("fromString -ff = %v", got)
	}
	if got := FromString_UInt64("-1", true, 10); got != math.MaxUint64 {
		t.Fatalf("unsigned fromString -1 = %v", got)
	}
}

func TestNumbersConvertTruncatingAndSaturating(t *testing.T) {
	if FromNumber_Int64(-2.9, false) != -2 || FromNumber_Int64(math.NaN(), false) != 0 {
		t.Fatal("truncation and NaN")
	}
	if FromNumber_Int64(1e30, false) != math.MaxInt64 || FromNumber_UInt64(-1, true) != 0 || FromNumber_UInt64(1e30, true) != math.MaxUint64 {
		t.Fatal("saturation")
	}
	if NumberBitsToInt(4294967295) != -1 {
		t.Fatal("to Int")
	}
}

func TestBitsAndBytes(t *testing.T) {
	long := FromBits_Int64(-1, 1, false)
	if long != 1<<33-1 || FromBits_UInt64(-1, -1, true) != math.MaxUint64 {
		t.Fatal("bit halves")
	}
	if got := FromBytesBE_Int64(ToBytes_(long, false), false); got != long {
		t.Fatalf("big-endian round trip = %v", got)
	}
	if got := FromBytes_Int64(ToBytes_(long, true), false, true); got != long {
		t.Fatalf("little-endian round trip = %v", got)
	}
	if GetNumBitsAbs(int64(0)) != 1 || GetNumBitsAbs(int64(math.MinInt64)) != 64 || GetNumBitsAbs(int64(-1)) != 1 || GetNumBitsAbs(uint64(math.MaxUint64)) != 64 {
		t.Fatal("bit counts")
	}
	if RotateRight_(int64(1), 1) != math.MinInt64 || RotateLeft_(uint64(1), 65) != 2 {
		t.Fatal("rotations")
	}
}

func TestValuesConvert(t *testing.T) {
	if FromValue_Int64(-3.0, false) != -3 || FromValue_UInt64("42", true) != 42 || FromValue_UInt64(int64(-1), true) != math.MaxUint64 {
		t.Fatal("fromValue")
	}
	if !IsLong_(int64(1)) || !IsLong_(uint64(1)) || IsLong_(1.0) {
		t.Fatal("isLong")
	}
}
