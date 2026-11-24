package bigints

import (
	"encoding/json"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/i-am-the-slime/iris-go/exception"
)

// cases are purescript-bigints' results under Bun, recorded by
// testdata/generate.js.
type cases struct {
	FromBase []struct {
		Base         int32
		Text, Result string
	}
	FromNumber []struct{ Number, Result string }
	Unary      []struct {
		Value, ToNumber, Abs, Not string
		Even, Odd                 bool
		Bases, Digits             []string
		Shifts                    [][2]string
	}
	Binary []struct {
		Left, Right string
		Results     []string
		Equals      bool
		Compare     int32
	}
	Powers []struct{ Base, Exponent, Result string }
	Primes []struct {
		Value string
		Prime bool
	}
}

func load(t *testing.T) cases {
	t.Helper()
	contents, err := os.ReadFile("testdata/cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var recorded cases
	if err := json.Unmarshal(contents, &recorded); err != nil {
		t.Fatal(err)
	}
	return recorded
}

func value(t *testing.T, text string) *big.Int {
	t.Helper()
	parsed, ok := new(big.Int).SetString(text, 10)
	if !ok {
		t.Fatalf("bad test value %q", text)
	}
	return parsed
}

// attempt renders a result as the generator does: a BigInt as its decimal
// digits, anything else as JSON, a thrown Error as its name and message.
func attempt[A any](f func() A) (rendered string) {
	defer func() {
		if recovered := recover(); recovered != nil {
			failure, ok := exception.AsError(recovered)
			if !ok {
				panic(recovered)
			}
			name := failure.Name
			if name == "" {
				name = "Error"
			}
			rendered = "THROW " + name + ": " + failure.Message
		}
	}()
	result := f()
	if number, ok := any(result).(*big.Int); ok {
		return number.String()
	}
	var encoded strings.Builder
	encoder := json.NewEncoder(&encoded)
	encoder.SetEscapeHTML(false)
	if err := encoder.Encode(result); err != nil {
		panic(err)
	}
	return strings.TrimSuffix(encoded.String(), "\n")
}

func maybe(value *big.Int, found bool) string {
	if !found {
		return "Nothing"
	}
	return value.String()
}

func TestFromBaseMatchesBigInteger(t *testing.T) {
	for _, recorded := range load(t).FromBase {
		if got := maybe(FromBaseImpl(recorded.Base, recorded.Text)); got != recorded.Result {
			t.Errorf("fromBase %d %q = %s, JavaScript %s", recorded.Base, recorded.Text, got, recorded.Result)
		}
	}
}

func TestFromNumberTruncates(t *testing.T) {
	for _, recorded := range load(t).FromNumber {
		number := math.Copysign(0, -1)
		if recorded.Number != "-0" {
			number, _ = strconv.ParseFloat(recorded.Number, 64)
		}
		if got := maybe(FromNumberImpl(number)); got != recorded.Result {
			t.Errorf("fromNumber %s = %s, JavaScript %s", recorded.Number, got, recorded.Result)
		}
	}
}

var bases = []int32{10, 2, 8, 16, 36, 1, 0, -1, -2, -10, 37, 100, 3, -16, 62, 5, 7}

var digitBases = []int32{2, 10, 16, 100, -3, 1, 0}

var shifts = []float64{0, 1, 7, 63, 64, 65, 200, -1, -7, -64, -200, 1.5, math.NaN(), math.Inf(1), 1e7, 10000001, -1e8}

type digits struct {
	Value      []int32 `json:"value"`
	IsNegative bool    `json:"isNegative"`
}

func TestUnaryOperationsMatchBigInteger(t *testing.T) {
	for _, recorded := range load(t).Unary {
		x := value(t, recorded.Value)
		if got := strconv.FormatFloat(ToNumber(x), 'g', -1, 64); !sameNumber(got, recorded.ToNumber) {
			t.Errorf("toNumber %s = %s, JavaScript %s", recorded.Value, got, recorded.ToNumber)
		}
		if got := Abs(x).String(); got != recorded.Abs {
			t.Errorf("abs %s = %s", recorded.Value, got)
		}
		if got := Not(x).String(); got != recorded.Not {
			t.Errorf("not %s = %s, JavaScript %s", recorded.Value, got, recorded.Not)
		}
		if Even(x) != recorded.Even || Odd(x) != recorded.Odd {
			t.Errorf("even/odd %s", recorded.Value)
		}
		for index, base := range bases {
			if recorded.Bases[index] == javaScriptStackOverflow {
				continue
			}
			if got := attempt(func() string { return ToBase(base, x) }); got != recorded.Bases[index] {
				t.Errorf("toBase %d %s = %s, JavaScript %s", base, recorded.Value, got, recorded.Bases[index])
			}
		}
		for index, base := range digitBases {
			if recorded.Digits[index] == javaScriptStackOverflow {
				continue
			}
			got := attempt(func() digits {
				negative, values := DigitsInBase(base, x)
				return digits{values, negative}
			})
			if want := recorded.Digits[index]; got != want {
				t.Errorf("digitsInBase %d %s = %s, JavaScript %s", base, recorded.Value, got, want)
			}
		}
		for index, n := range shifts {
			for side, shift := range []func(*big.Int, float64) *big.Int{Shl, Shr} {
				want := recorded.Shifts[index][side]
				if strings.HasPrefix(want, javaScriptCoreSizeLimit) {
					beyondJavaScriptCore(t, shift(x, n))
					continue
				}
				if got := attempt(func() *big.Int { return shift(x, n) }); got != want {
					t.Errorf("shift %d %s %v = %.80s, JavaScript %.80s", side, recorded.Value, n, got, want)
				}
			}
		}
	}
}

// sameNumber compares Go's and JavaScript's spelling of a Number.
func sameNumber(goText, javaScriptText string) bool {
	goNumber, _ := strconv.ParseFloat(goText, 64)
	javaScriptNumber, _ := strconv.ParseFloat(javaScriptText, 64)
	return goNumber == javaScriptNumber
}

func TestBinaryOperationsMatchBigInteger(t *testing.T) {
	operations := []func(*big.Int, *big.Int) *big.Int{BiAdd, BiSub, BiMul, BiDiv, BiMod, And, Or, Xor}
	names := []string{"add", "sub", "mul", "div", "mod", "and", "or", "xor"}
	for _, recorded := range load(t).Binary {
		x, y := value(t, recorded.Left), value(t, recorded.Right)
		for index, operation := range operations {
			if got := attempt(func() *big.Int { return operation(x, y) }); got != recorded.Results[index] {
				t.Errorf("%s %s %s = %s, JavaScript %s", names[index], recorded.Left, recorded.Right, got, recorded.Results[index])
			}
		}
		if BiEquals(x, y) != recorded.Equals || BiCompare(x, y) != recorded.Compare {
			t.Errorf("compare %s %s", recorded.Left, recorded.Right)
		}
	}
}

func TestPowMatchesBigInteger(t *testing.T) {
	for _, recorded := range load(t).Powers {
		if got := attempt(func() *big.Int { return Pow(value(t, recorded.Base), value(t, recorded.Exponent)) }); got != recorded.Result {
			t.Errorf("pow %s %s = %s, JavaScript %s", recorded.Base, recorded.Exponent, got, recorded.Result)
		}
	}
}

func TestPrimeIsBigIntegersMillerRabin(t *testing.T) {
	for _, recorded := range load(t).Primes {
		if got := Prime(value(t, recorded.Value)); got != recorded.Prime {
			t.Errorf("prime %s = %t, JavaScript %t", recorded.Value, got, recorded.Prime)
		}
	}
}

// javaScriptStackOverflow is JavaScriptCore running out of stack while
// spreading a long unary digit array, which Go builds instead.
const javaScriptStackOverflow = "THROW RangeError: Maximum call stack size exceeded."

func TestUnaryDigitsHaveNoStackLimit(t *testing.T) {
	if got := ToBase(1, big.NewInt(700000)); got != strings.Repeat("1", 700000) {
		t.Errorf("toBase 1 700000 has %d characters", len(got))
	}
}

// javaScriptCoreSizeLimit is JavaScriptCore's refusal to allocate a BigInt
// beyond 2^20 bits, a resource limit of the engine that Go does not share.
const javaScriptCoreSizeLimit = "THROW RangeError: Out of memory"

func beyondJavaScriptCore(t *testing.T, result *big.Int) {
	t.Helper()
	if result.BitLen() < 1<<20-64 {
		t.Errorf("JavaScriptCore refused a result of %d bits", result.BitLen())
	}
}
