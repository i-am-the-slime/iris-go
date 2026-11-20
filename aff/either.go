package aff

import (
	"reflect"
	"sync"

	"github.com/i-am-the-slime/iris-go/exception"
)

// The Aff foreigns exchange `Either Error a` values with PureScript code, and
// each instantiation is a generated ADT struct go-support cannot name:
// `struct{ Tag EitherTag; LeftF1 L; RightF1 R }` with Left = 0, Right = 1.
// The codec reads and builds them by field name, once per instantiation.
type eitherCodec struct {
	tag   int
	left  int
	right int
}

var codecs sync.Map

func codecOf(t reflect.Type) *eitherCodec {
	if cached, ok := codecs.Load(t); ok {
		return cached.(*eitherCodec)
	}
	field := func(name string) int {
		if f, ok := t.FieldByName(name); ok {
			return f.Index[0]
		}
		return -1
	}
	codec := &eitherCodec{tag: field("Tag"), left: field("LeftF1"), right: field("RightF1")}
	if t.Kind() != reflect.Struct || codec.tag < 0 {
		panic("aff: " + t.String() + " is not an Either representation")
	}
	codecs.Store(t, codec)
	return codec
}

func eitherLeft[E any](failure exception.Error) E {
	var e E
	v := reflect.ValueOf(&e).Elem()
	codec := codecOf(v.Type())
	v.Field(codec.tag).SetInt(0)
	if codec.left >= 0 {
		v.Field(codec.left).Set(reflect.ValueOf(failure))
	}
	return e
}

func eitherRight[E, A any](value A) E {
	var e E
	v := reflect.ValueOf(&e).Elem()
	codec := codecOf(v.Type())
	v.Field(codec.tag).SetInt(1)
	if codec.right >= 0 {
		v.Field(codec.right).Set(reflect.ValueOf(&value).Elem())
	}
	return e
}

// eitherOutcome builds the `Either Error a` a fiber's outcome denotes.
func eitherOutcome[E, A any](result outcome) E {
	if result.failure != nil {
		return eitherLeft[E](*result.failure)
	}
	return eitherRight[E, A](result.value.(A))
}

// decodeEither splits an `Either Error a` into its failure or value.
func decodeEither[A any](e any) (A, *exception.Error) {
	var value A
	v := reflect.ValueOf(e)
	codec := codecOf(v.Type())
	if v.Field(codec.tag).Int() == 0 {
		failure := exception.Error{}
		if codec.left >= 0 {
			failure = v.Field(codec.left).Interface().(exception.Error)
		}
		return value, &failure
	}
	if codec.right >= 0 {
		value = v.Field(codec.right).Interface().(A)
	}
	return value, nil
}
