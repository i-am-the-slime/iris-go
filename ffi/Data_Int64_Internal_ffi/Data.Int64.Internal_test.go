package Data_Int64_Internal_ffi

import (
	"math"
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
)

func TestArithmeticWrapsLikeTwosComplement(t *testing.T) {
	if got := Add_(MaxValue_, One_); got != MinValue_ {
		t.Fatalf("max + 1 = %v", got)
	}
	if got := Subtract_(Uzero_, Uone_); got != MaxUnsignedValue_ {
		t.Fatalf("0u - 1u = %v", got)
	}
	if got := Divide_(MinValue_, NegOne_); got != MinValue_ {
		t.Fatalf("min / -1 = %v", got)
	}
	if got := Modulo_(MinValue_, NegOne_); got != Zero_ {
		t.Fatalf("min %% -1 = %v", got)
	}
}

func TestDivisionReadsTheDividendsSignedness(t *testing.T) {
	if got := ToString_(Divide_(MaxUnsignedValue_, FromInt_(10, true)), 10); got != "1844674407370955161" {
		t.Fatalf("unsigned max / 10 = %s", got)
	}
	if got := ToString_(Divide_(FromInt_(-7, false), FromInt_(2, false)), 10); got != "-3" {
		t.Fatalf("-7 / 2 = %s", got)
	}
}

func TestDivisionByZeroThrows(t *testing.T) {
	defer func() {
		if thrown, ok := recover().(exception.Error); !ok || thrown.Message != "division by zero" {
			t.Fatalf("recovered %v", thrown)
		}
	}()
	Divide_(One_, Zero_)
}

func TestCompareFollowsSignedness(t *testing.T) {
	if Compare_(NegOne_, One_) != -1 || Compare_(MaxUnsignedValue_, Uone_) != 1 {
		t.Fatal("same-signedness comparisons")
	}
	if Compare_(NegOne_, MaxUnsignedValue_) != -1 || Equals_(NegOne_, MaxUnsignedValue_) {
		t.Fatal("a negative signed Long is below every unsigned one")
	}
}

func TestStringsInEveryRadix(t *testing.T) {
	if got := ToString_(MinValue_, 16); got != "-8000000000000000" {
		t.Fatalf("min in hex = %s", got)
	}
	if got := ToString_(MaxUnsignedValue_, 36); got != "3w5e11264sgsf" {
		t.Fatalf("unsigned max in base 36 = %s", got)
	}
	if got := SafeReadLong_("-000", true, 10); got == nil || *got != Uzero_ {
		t.Fatalf("-000 unsigned = %v", got)
	}
	if got := SafeReadLong_("1Z141Z3", false, 36); got == nil || ToString_(*got, 36) != "1z141z3" {
		t.Fatalf("1Z141Z3 = %v", got)
	}
	if got := SafeReadLong_("1Z00000000000", false, 36); got != nil {
		t.Fatalf("uppercase past signed max = %v", got)
	}
	if got := SafeReadLong_("3W5E11264SGSG", true, 36); got != nil {
		t.Fatalf("one past unsigned max = %v", got)
	}
	if got := FromString_("-ff", false, 16); got != FromInt_(-255, false) {
		t.Fatalf("fromString -ff = %v", got)
	}
}

func TestNumbersConvertTruncatingAndSaturating(t *testing.T) {
	if FromNumber_(-2.9, false) != FromInt_(-2, false) || FromNumber_(math.NaN(), false) != Zero_ {
		t.Fatal("truncation and NaN")
	}
	if FromNumber_(1e30, false) != MaxValue_ || FromNumber_(-1, true) != Uzero_ || FromNumber_(1e30, true) != MaxUnsignedValue_ {
		t.Fatal("saturation")
	}
	if ToNumber_(MaxUnsignedValue_) != 18446744073709551616 || NumberBitsToInt(4294967295) != -1 {
		t.Fatal("to Number")
	}
}

func TestBitsAndBytes(t *testing.T) {
	long := FromBits_(-1, 1, false)
	if GetLowBits_(long) != -1 || GetHighBits_(long) != 1 || GetLowBitsUnsigned_(long) != 4294967295 {
		t.Fatal("bit halves")
	}
	if got := FromBytesBE_(ToBytes_(long, false), false); got != long {
		t.Fatalf("big-endian round trip = %v", got)
	}
	if got := FromBytes_(ToBytes_(long, true), false, true); got != long {
		t.Fatalf("little-endian round trip = %v", got)
	}
	if GetNumBitsAbs(Zero_) != 1 || GetNumBitsAbs(MinValue_) != 64 || GetNumBitsAbs(NegOne_) != 1 {
		t.Fatal("bit counts")
	}
	if RotateRight_(One_, One_) != MinValue_ || ShiftRight_(MinValue_, FromInt_(63, false)) != NegOne_ {
		t.Fatal("rotations and shifts")
	}
}
