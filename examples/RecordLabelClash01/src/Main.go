// Foreign implementations for module Main. The records' structs name each
// field as Iris does: a label starting in lower case keeps the plain name,
// one starting in upper case that clashes with it takes a `_` behind it.
package Main_ffi

import "strconv"

type Pair struct {
	A_ int `purescript:"A"`
	A  int
}

type Triple struct {
	A_ int `purescript:"A"`
	A  string
	A2 float64
}

type Header struct {
	Content_type2 string `purescript:"content-type"`
	Content_type  string
}

func DescribePair(pair Pair) string {
	return "A_=" + strconv.Itoa(int(pair.A_)) + " A=" + strconv.Itoa(int(pair.A))
}

func DescribeTriple(triple Triple) string {
	return "A_=" + strconv.Itoa(int(triple.A_)) + " A=" + triple.A + " A2=" + strconv.FormatFloat(triple.A2, 'f', -1, 64)
}

func DescribeHeader(header Header) string {
	return "Content_type2=" + header.Content_type2 + " Content_type=" + header.Content_type
}

func MakePair(upper int, lower int) Pair {
	return Pair{A_: upper, A: lower}
}
