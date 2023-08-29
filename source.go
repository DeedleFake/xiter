package xiter

import (
	"context"
	"strings"
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
	return OfSlice(vals)
}

// OfSlice returns a Seq over the elements of s.
func OfSlice[T any, S ~[]T](s S) Seq[T] {
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

// StringSplit returns an iterator over the substrings of s that are
// separated by sep. It behaves very similarly to [strings.Split].
func StringSplit(s, sep string) Seq[string] {
	if sep == "" {
		return Map(Runes(s), func(c rune) string { return string(c) })
	}

	return func(yield func(string) bool) bool {
		for {
			m := strings.Index(s, sep)
			if m < 0 {
				return yield(s)
			}
			if !yield(s[:m]) {
				return false
			}
			s = s[m+len(sep):]
		}
	}
}

// OfMap returns a Seq over the key-value pairs of m.
func OfMap[K comparable, V any, M ~map[K]V](m M) Seq2[K, V] {
	return func(yield func(K, V) bool) bool {
		for k, v := range m {
			if !yield(k, v) {
				return false
			}
		}
		return false
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
	return func(yield func(Pair[T1, T2]) bool) bool {
		return seq(func(v1 T1, v2 T2) bool {
			return yield(Pair[T1, T2]{v1, v2})
		})
	}
}

// V1 returns a Seq which iterates over only the T1 elements of seq.
func V1[T1, T2 any](seq Seq2[T1, T2]) Seq[T1] {
	return func(yield func(T1) bool) bool {
		return seq(func(v1 T1, v2 T2) bool {
			return yield(v1)
		})
	}
}

// V2 returns a Seq which iterates over only the T2 elements of seq.
func V2[T1, T2 any](seq Seq2[T1, T2]) Seq[T2] {
	return func(yield func(T2) bool) bool {
		return seq(func(v1 T1, v2 T2) bool {
			return yield(v2)
		})
	}
}

// RecvContext returns a Seq that receives from c continuously until
// either c is closed or the given context is canceled.
func RecvContext[T any](ctx context.Context, c <-chan T) Seq[T] {
	return func(yield func(T) bool) bool {
		for {
			select {
			case <-ctx.Done():
				return false
			case v, ok := <-c:
				if !ok || !yield(v) {
					return false
				}
			}
		}
	}
}
