package xiter

import (
	"unicode/utf8"
	"unsafe"
)

// Generate returns a Seq that first yields start and then yields
// successive values by adding step to the previous continuously. The
// returned Seq does not end. To limit it to a specific number of
// returned elements, use [Limit].
func Generate[T Addable](start, step T) Seq[T] {
	return func(yield func(T) bool) bool {
		for {
			if !yield(start) {
				return false
			}
			start += step
		}
	}
}

// Of returns a Seq that yields the provided values.
func Of[T any](vals ...T) Seq[T] {
	return Slice(vals)
}

// Slice returns a Seq over the elements of s.
func Slice[T any, S ~[]T](s S) Seq[T] {
	return func(yield func(T) bool) bool {
		for _, v := range s {
			if !yield(v) {
				return false
			}
		}
		return false
	}
}

// Bytes returns a Seq over the bytes of s.
func Bytes(s string) Seq[byte] {
	return func(yield func(byte) bool) bool {
		for i := 0; i < len(s); i++ {
			if !yield(s[i]) {
				return false
			}
		}
		return false
	}
}

// Runes returns a Seq over the runes of s.
func Runes[T ~[]byte | ~string](s T) Seq[rune] {
	return func(yield func(rune) bool) bool {
		b := unsafe.Slice(unsafe.StringData(*(*string)(unsafe.Pointer(&s))), len(s))
		for len(b) > 0 {
			r, size := utf8.DecodeRune(b)
			if !yield(r) {
				return false
			}
			b = b[size:]
		}
		return false
	}
}

// MapEntries returns a Seq over the key-value pairs of m.
func MapEntries[K comparable, V any, M ~map[K]V](m M) Seq[Pair[K, V]] {
	return func(yield func(Pair[K, V]) bool) bool {
		for k, v := range m {
			if !yield(Pair[K, V]{k, v}) {
				return false
			}
		}
		return false
	}
}

// MapKeys returns a Seq over the keys of m.
func MapKeys[K comparable, V any, M ~map[K]V](m M) Seq[K] {
	return Map(MapEntries(m), func(v Pair[K, V]) K { return v.V1 })
}

// MapValues returns a Seq over the values of m.
func MapValues[K comparable, V any, M ~map[K]V](m M) Seq[V] {
	return Map(MapEntries(m), func(v Pair[K, V]) V { return v.V2 })
}

// ToPair takes a two-value iterator and produces a single-value
// iterator of pairs.
func ToPair[T1, T2 any](seq Seq2[T1, T2]) Seq[Pair[T1, T2]] {
	return func(yield func(Pair[T1, T2]) bool) bool {
		return seq(func(v1 T1, v2 T2) bool {
			return yield(Pair[T1, T2]{v1, v2})
		})
	}
}
