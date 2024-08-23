package xiter

import (
	"fmt"
	"iter"
	"reflect"
)

// OfValue returns a Seq2 that iterates over any iterable type using
// reflection. If the type is one which only produces a single value
// per iteration, such as a channel, the second value yielded each
// iteration will just be reflect.Value{}.
func OfValue(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	switch v.Kind() {
	case reflect.Array, reflect.Slice:
		return ofValueIndexable(v)
	case reflect.String:
		return ofValueString(v)
	case reflect.Chan:
		return ofValueChan(v)
	case reflect.Map:
		return ofValueMap(v)
	case reflect.Pointer:
		if v.Type().Elem().Kind() == reflect.Array {
			return ofValuePointerToArray(v)
		}
	case reflect.Func:
		if isValueSeq(v) {
			return ofValueFunc(v)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return ofValueInt(v)
	}

	panic(fmt.Errorf("not rangeable type: %v", v.Type()))
}

func ofValueIndexable(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	return func(yield func(v1, v2 reflect.Value) bool) {
		for i := 0; i < v.Len(); i++ {
			if !yield(reflect.ValueOf(i), v.Index(i)) {
				return
			}
		}
		return
	}
}

func ofValueString(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	return FromPair(Map(ToPair(Enumerate(Runes(v.String()))),
		func(v Pair[int, rune]) Pair[reflect.Value, reflect.Value] {
			return P(reflect.ValueOf(v.V1), reflect.ValueOf(v.V2))
		}))
}

func ofValueChan(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	var zero reflect.Value
	return func(yield func(v1, v2 reflect.Value) bool) {
		for {
			val, ok := v.Recv()
			if !ok {
				return
			}
			if !yield(val, zero) {
				return
			}
		}
	}
}

func ofValueMap(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	return func(yield func(v1, v2 reflect.Value) bool) {
		iter := v.MapRange()
		defer iter.Reset(reflect.Value{})
		for iter.Next() {
			if !yield(iter.Key(), iter.Value()) {
				return
			}
		}
		return
	}
}

func ofValuePointerToArray(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	sv := v.Elem()
	if !sv.IsValid() {
		return func(func(v1, v2 reflect.Value) bool) { return }
	}
	return ofValueIndexable(sv)
}

func isValueSeq(v reflect.Value) bool {
	t := v.Type()
	if t.NumIn() != 1 {
		return false
	}
	if t.NumOut() != 1 || t.Out(0).Kind() != reflect.Bool {
		return false
	}

	yt := t.In(0)
	if yt.Kind() != reflect.Func {
		return false
	}
	if yt.NumIn() != 1 && yt.NumIn() != 2 {
		return false
	}
	if yt.NumOut() != 1 || t.Out(0).Kind() != reflect.Bool {
		return false
	}

	return true
}

func ofValueFunc(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	return func(yield func(v1, v2 reflect.Value) bool) {
		yv := reflect.MakeFunc(v.Type().In(0), func(vals []reflect.Value) []reflect.Value {
			v1, v2 := vals[0], reflect.Value{}
			if len(vals) == 2 {
				v2 = vals[1]
			}
			return []reflect.Value{reflect.ValueOf(yield(v1, v2))}
		})
		v.Call([]reflect.Value{yv})
		return
	}
}

func ofValueInt(v reflect.Value) iter.Seq2[reflect.Value, reflect.Value] {
	inc := func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Int() + 1) }
	if v.CanInt() && v.Int() < 0 {
		panic(fmt.Errorf("%v < 0", v.Int()))
	}
	if v.CanUint() {
		if v.Uint() < 0 {
			panic(fmt.Errorf("%v < 0", v.Int()))
		}
		inc = func(v reflect.Value) reflect.Value { return reflect.ValueOf(v.Uint() + 1) }
	}

	var zero reflect.Value
	return func(yield func(v1, v2 reflect.Value) bool) {
		for i := reflect.Zero(v.Type()); i.Interface() != v.Interface(); i = inc(i) {
			if !yield(i, zero) {
				return
			}
		}
		return
	}
}
