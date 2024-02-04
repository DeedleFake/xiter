package xiter

import (
	"context"
	"strings"
	"unicode/utf8"
	"unsafe"
)

// _Generate returns a Seq that first yields start and then yields
// successive values by adding step to the previous continuously. The
// returned Seq does not end. To limit it to a specific number of
// returned elements, use [Limit].
func _Generate[T Addable](start, step T) _Seq[T] {
	return func(yield func(T) bool) {
		for {
			if !yield(start) {
				return
			}
			start += step
		}
	}
}

// _Of returns a Seq that yields the provided values.
func _Of[T any](vals ...T) _Seq[T] {
	return _OfSlice(vals)
}

// _OfSlice returns a Seq over the elements of s. It is equivalent to
// range s with the index ignored.
func _OfSlice[T any, S ~[]T](s S) _Seq[T] {
	return _V2(_OfSliceIndex(s))
}

// _OfSliceIndex returns a Seq over the elements of s. It is equivalent
// to range s.
func _OfSliceIndex[T any, S ~[]T](s S) _Seq2[int, T] {
	return func(yield func(int, T) bool) {
		for i, v := range s {
			if !yield(i, v) {
				return
			}
		}
		return
	}
}

// _Bytes returns a Seq over the bytes of s.
func _Bytes(s string) _Seq[byte] {
	return func(yield func(byte) bool) {
		for i := 0; i < len(s); i++ {
			if !yield(s[i]) {
				return
			}
		}
		return
	}
}

// _Runes returns a Seq over the runes of s.
func _Runes[T ~[]byte | ~string](s T) _Seq[rune] {
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

// _StringSplit returns an iterator over the substrings of s that are
// separated by sep. It behaves very similarly to [strings.Split].
func _StringSplit(s, sep string) _Seq[string] {
	if sep == "" {
		return _Map(_Runes(s), func(c rune) string { return string(c) })
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

// _OfMap returns a Seq over the key-value pairs of m.
func _OfMap[K comparable, V any, M ~map[K]V](m M) _Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
		return
	}
}

// _MapKeys returns a Seq over the keys of m.
func _MapKeys[K comparable, V any, M ~map[K]V](m M) _Seq[K] {
	return _V1(_OfMap(m))
}

// _MapValues returns a Seq over the values of m.
func _MapValues[K comparable, V any, M ~map[K]V](m M) _Seq[V] {
	return _V2(_OfMap(m))
}

// _ToPair takes a two-value iterator and produces a single-value
// iterator of pairs.
func _ToPair[T1, T2 any](seq _Seq2[T1, T2]) _Seq[Pair[T1, T2]] {
	return func(yield func(Pair[T1, T2]) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(P(v1, v2))
		})
	}
}

// _V1 returns a Seq which iterates over only the T1 elements of seq.
func _V1[T1, T2 any](seq _Seq2[T1, T2]) _Seq[T1] {
	return func(yield func(T1) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(v1)
		})
	}
}

// _V2 returns a Seq which iterates over only the T2 elements of seq.
func _V2[T1, T2 any](seq _Seq2[T1, T2]) _Seq[T2] {
	return func(yield func(T2) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(v2)
		})
	}
}

// _OfChan returns a Seq which yields values received from c. The
// sequence ends when c is closed. It is equivalent to range c.
func _OfChan[T any](c <-chan T) _Seq[T] {
	return func(yield func(T) bool) {
		for v := range c {
			if !yield(v) {
				return
			}
		}
		return
	}
}

// _RecvContext returns a Seq that receives from c continuously until
// either c is closed or the given context is canceled.
func _RecvContext[T any](ctx context.Context, c <-chan T) _Seq[T] {
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
