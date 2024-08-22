package xiter

import (
	"context"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// Generate returns a Seq that first yields start and then yields
// successive values by adding step to the previous continuously. The
// returned Seq does not end. To limit it to a specific number of
// returned elements, use [Limit].
func Generate[T Addable](start, step T) Seq[T] {
	return func(yield func(T) bool) {
		for {
			if !yield(start) {
				return
			}
			start += step
		}
	}
}

// Of returns a Seq that yields the provided values.
func Of[T any](vals ...T) Seq[T] {
	return OfSlice(vals)
}

// OfSlice returns a Seq over the elements of s. It is equivalent to
// range s with the index ignored.
func OfSlice[T any, S ~[]T](s S) Seq[T] {
	return V2(OfSliceIndex(s))
}

// OfSliceIndex returns a Seq over the elements of s. It is equivalent
// to range s.
func OfSliceIndex[T any, S ~[]T](s S) Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range s {
			if !yield(i, v) {
				return
			}
		}
		return
	}
}

// Bytes returns a Seq over the bytes of s.
func Bytes(s string) Seq[byte] {
	return func(yield func(byte) bool) {
		for i := 0; i < len(s); i++ {
			if !yield(s[i]) {
				return
			}
		}
		return
	}
}

// Runes returns a Seq over the runes of s.
func Runes[T ~[]byte | ~string](s T) Seq[rune] {
	return func(yield func(rune) bool) {
		b := unsafe.Slice(unsafe.StringData(*(*string)(unsafe.Pointer(&s))), len(s))
		for len(b) > 0 {
			r, size := utf8.DecodeRune(b)
			if !yield(r) {
				return
			}
			b = b[size:]
		}
		return
	}
}

// StringSplit returns an iterator over the substrings of s that are
// separated by sep. It behaves very similarly to [strings.Split].
func StringSplit(s, sep string) Seq[string] {
	if sep == "" {
		return Map(Runes(s), func(c rune) string { return string(c) })
	}

	return func(yield func(string) bool) {
		for {
			m := strings.Index(s, sep)
			if m < 0 {
				yield(s)
				return
			}
			if !yield(s[:m]) {
				return
			}
			s = s[m+len(sep):]
		}
	}
}

// StringFields returns an iterator over the substrings of s that are
// seperated by consecutive whitespace as determined by
// [unicode.IsSpace]. It is very similar to [strings.Fields].
func StringFields(s string) Seq[string] {
	return StringFieldsFunc(s, unicode.IsSpace)
}

// StringFieldsFunc returns an iterator over the substrings of s that
// are seperated by consecutive sections of runes for which sep
// returns true. It behaves very similarly to [strings.FieldsFunc].
func StringFieldsFunc(s string, sep func(rune) bool) Seq[string] {
	return func(yield func(string) bool) {
		start := 0
		for i, r := range Enumerate(Runes(s)) {
			if !sep(r) {
				continue
			}

			field := s[start:i]
			start = i + 1
			if field == "" {
				continue
			}
			if !yield(field) {
				return
			}

		}

		field := s[start:]
		if field == "" {
			return
		}
		if !yield(field) {
			return
		}
	}
}

// OfMap returns a Seq over the key-value pairs of m.
func OfMap[K comparable, V any, M ~map[K]V](m M) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
		return
	}
}

// MapKeys returns a Seq over the keys of m.
func MapKeys[K comparable, V any, M ~map[K]V](m M) Seq[K] {
	return V1(OfMap(m))
}

// MapValues returns a Seq over the values of m.
func MapValues[K comparable, V any, M ~map[K]V](m M) Seq[V] {
	return V2(OfMap(m))
}

// ToPair takes a two-value iterator and produces a single-value
// iterator of pairs.
func ToPair[T1, T2 any](seq Seq2[T1, T2]) Seq[Pair[T1, T2]] {
	return func(yield func(Pair[T1, T2]) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(P(v1, v2))
		})
	}
}

// V1 returns a Seq which iterates over only the T1 elements of seq.
func V1[T1, T2 any](seq Seq2[T1, T2]) Seq[T1] {
	return func(yield func(T1) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(v1)
		})
	}
}

// V2 returns a Seq which iterates over only the T2 elements of seq.
func V2[T1, T2 any](seq Seq2[T1, T2]) Seq[T2] {
	return func(yield func(T2) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(v2)
		})
	}
}

// OfChan returns a Seq which yields values received from c. The
// sequence ends when c is closed. It is equivalent to range c.
func OfChan[T any](c <-chan T) Seq[T] {
	return func(yield func(T) bool) {
		for v := range c {
			if !yield(v) {
				return
			}
		}
		return
	}
}

// RecvContext returns a Seq that receives from c continuously until
// either c is closed or the given context is canceled.
func RecvContext[T any](ctx context.Context, c <-chan T) Seq[T] {
	return func(yield func(T) bool) {
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-c:
				if !ok || !yield(v) {
					return
				}
			}
		}
	}
}
