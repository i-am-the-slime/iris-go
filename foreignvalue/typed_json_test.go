package foreignvalue

import (
	"math"
	"strconv"
	"testing"
)

type point struct {
	X    int
	Tags []string
}

func decodePoint(d *Decoder) point {
	var p point
	hasX := false
	if d.Object() {
		for {
			switch d.Key() {
			case "x":
				p.X = d.Int()
				hasX = true
			case "tags":
				p.Tags = []string{}
				if d.Array() {
					for {
						p.Tags = append(p.Tags, d.String())
						if !d.MoreElements() {
							break
						}
					}
				}
			default:
				d.Skip()
			}
			if !d.MoreMembers() {
				break
			}
		}
	}
	if !hasX {
		d.Fail()
	}
	return p
}

func TestDecodeReadsMembersInAnyOrderAndSkipsUnknownOnes(t *testing.T) {
	p, ok := Decode(` { "tags" : ["a","é"], "skip": {"deep":[1,{"n":null}]}, "x" : 1e2 } `, decodePoint)
	if !ok || p.X != 100 || len(p.Tags) != 2 || p.Tags[1] != "é" {
		t.Fatalf("got %+v %v", p, ok)
	}
}

func TestDecodeRefusesTextOfAnotherType(t *testing.T) {
	for _, text := range []string{
		`{"x":1.5}`, `{"x":9223372036854775808}`, `{"x":"1"}`, `{}`, `[1]`, `null`,
		`{"x":1,"tags":[1]}`, `{"x":1} 2`, `{"x":1,"skip":[1,]}`, `{"x":01}`, ``,
	} {
		if p, ok := Decode(text, decodePoint); ok {
			t.Errorf("%q decoded as %+v", text, p)
		}
	}
}

func TestDecodeKeepsARepeatedMembersLastValue(t *testing.T) {
	if p, ok := Decode(`{"x":1,"x":-0}`, decodePoint); !ok || p.X != 0 {
		t.Fatalf("got %+v %v", p, ok)
	}
}

func TestDecodeLetsOtherPanicsThrough(t *testing.T) {
	defer func() {
		if recover() != "boom" {
			t.Fatal("the panic was swallowed")
		}
	}()
	Decode(`1`, func(*Decoder) int { panic("boom") })
}

func TestEncodeWritesAsJSONStringify(t *testing.T) {
	text := Encode(func(e *Encoder, values []float64) {
		e.Raw(`{"n":[`)
		for i, value := range values {
			if i > 0 {
				e.Raw(",")
			}
			e.Number(value)
		}
		e.Raw("]")
		e.Member(`"s":`)
		e.String("q\"\\\n\x01é")
		e.Key("k\"")
		e.Int(math.MinInt32)
		e.Raw(",\"b\":")
		e.Boolean(false)
		e.Raw("}")
	}, []float64{0.30000000000000004, math.Inf(1), math.NaN(), 1e21, -0.0, 5e-324})
	want := `{"n":[0.30000000000000004,null,null,1e+21,0,5e-324],"s":"q\"\\\n\u0001é","k\"":-2147483648,"b":false}`
	if text != want {
		t.Fatalf("got  %s\nwant %s", text, want)
	}
}

func TestEncodeMembersTakeACommaOnlyAfterTheFirst(t *testing.T) {
	text := Encode(func(e *Encoder, _ int) {
		e.Raw("{")
		e.Member(`"a":`)
		e.Int(1)
		e.Key("b")
		e.Int(2)
		e.Raw("}")
	}, 0)
	if text != `{"a":1,"b":2}` {
		t.Fatalf("got %s", text)
	}
}

func TestEncoderWritesEveryDigitOfAnInt(t *testing.T) {
	for _, n := range []int{0, 7, -7, 2147483648, math.MaxInt, math.MinInt} {
		var e Encoder
		e.Int(n)
		if got, want := string(e.out), strconv.Itoa(n); got != want {
			t.Errorf("Int(%d) wrote %q, want %q", n, got, want)
		}
	}
}
