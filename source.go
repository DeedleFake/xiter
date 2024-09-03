package xiter

import (
	"context"
	"io"
	"iter"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// Generate returns a Seq that first yields start and then yields
// successive values by adding step to the previous continuously. The
// returned Seq does not end. To limit it to a specific number of
// returned elements, use [Limit].
func Generate[T Addable](start, step T) iter.Seq[T] {
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
func Of[T any](vals ...T) iter.Seq[T] {
	return slices.Values(vals)
}

// Bytes returns a Seq over the bytes of s.
func Bytes(s string) iter.Seq[byte] {
	return func(yield func(byte) bool) {
		for i := 0; i < len(s); i++ {
			if !yield(s[i]) {
				return
			}
		}
	}
}

// Runes returns a Seq over the runes of s.
func Runes[T ~[]byte | ~string](s T) iter.Seq[rune] {
	return func(yield func(rune) bool) {
		b := unsafe.Slice(unsafe.StringData(*(*string)(unsafe.Pointer(&s))), len(s))
		for len(b) > 0 {
			r, size := utf8.DecodeRune(b)
			if !yield(r) {
				return
			}
			b = b[size:]
		}
	}
}

// StringSplit returns an iterator over the substrings of s that are
// separated by sep. It behaves very similarly to [strings.Split].
func StringSplit(s, sep string) iter.Seq[string] {
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
func StringFields(s string) iter.Seq[string] {
	return StringFieldsFunc(s, unicode.IsSpace)
}

// StringFieldsFunc returns an iterator over the substrings of s that
// are seperated by consecutive sections of runes for which sep
// returns true. It behaves very similarly to [strings.FieldsFunc].
func StringFieldsFunc(s string, sep func(rune) bool) iter.Seq[string] {
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

// ToPair takes a two-value iterator and produces a single-value
// iterator of pairs.
func ToPair[T1, T2 any](seq iter.Seq2[T1, T2]) iter.Seq[Pair[T1, T2]] {
	return func(yield func(Pair[T1, T2]) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(P(v1, v2))
		})
	}
}

// V1 returns a Seq which iterates over only the T1 elements of seq.
func V1[T1, T2 any](seq iter.Seq2[T1, T2]) iter.Seq[T1] {
	return func(yield func(T1) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(v1)
		})
	}
}

// V2 returns a Seq which iterates over only the T2 elements of seq.
func V2[T1, T2 any](seq iter.Seq2[T1, T2]) iter.Seq[T2] {
	return func(yield func(T2) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return yield(v2)
		})
	}
}

// OfChan returns a Seq which yields values received from c. The
// sequence ends when c is closed. It is equivalent to range c.
func OfChan[T any](c <-chan T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range c {
			if !yield(v) {
				return
			}
		}
	}
}

// RecvContext returns a Seq that receives from c continuously until
// either c is closed or the given context is canceled.
func RecvContext[T any](ctx context.Context, c <-chan T) iter.Seq[T] {
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

// SliceChunksFunc is like [ChunksFunc] but operates on a slice
// instead of an [iter.Seq]. When dealing with data that is in a
// slice, this is more effecient than using ChunksFunc as it can yield
// subslices of the underlying slice instead of having to allocate a
// moving window. The yielded subslices have their capacity clipped.
func SliceChunksFunc[T any, C comparable, S ~[]T](s S, chunker func(T) C) iter.Seq[S] {
	return func(yield func(S) bool) {
		if len(s) == 0 {
			return
		}

		prev := chunker(s[0])
		var start int
		for i := 1; i < len(s); i++ {
			v := s[i]
			cur := chunker(v)
			if cur == prev {
				continue
			}

			if !yield(slices.Clip(s[start:i])) {
				return
			}
			prev, start = cur, i
		}

		last := s[start:]
		if len(last) != 0 {
			if !yield(slices.Clip(last)) {
				return
			}
		}
	}
}

// reader returns an iterator that reads using the given function. If
// that function returns a non-nil error, the iterator will yield that
// error and then exit. If the iterator is terminated early, it will
// call the provided done function first.
func reader[T byte | rune](read func() (T, error), done func()) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for {
			c, err := read()
			if err != nil {
				yield(c, err)
				return
			}
			if !yield(c, nil) {
				done()
				return
			}
		}
	}
}

// ReadBytes returns an iterator over the bytes of r. If reading the
// next byte returns an error, the iterator will yield a non-nil error
// and then exit.
func ReadBytes(r io.ByteReader) iter.Seq2[byte, error] {
	return reader(
		r.ReadByte,
		func() {},
	)
}

// ReadRunes returns an iterator over the runes of r. If reading the
// next rune returns an error, the iterator will yield a non-nil error
// and then exit.
func ReadRunes(r io.RuneReader) iter.Seq2[rune, error] {
	return reader(
		func() (rune, error) {
			c, _, err := r.ReadRune()
			return c, err
		},
		func() {},
	)
}

// ScanBytes returns an iterator over the bytes of r. If reading the
// next byte returns an error, the iterator will yield a non-nil error
// and then exit.
//
// If the iterator is terminated early, it will unread
// the last byte read, allowing it to be used again to continue from
// where it left off. If this is not the desired behavior, use
// [ReadBytes] instead.
func ScanBytes(r io.ByteScanner) iter.Seq2[byte, error] {
	return reader(
		r.ReadByte,
		func() { r.UnreadByte() },
	)
}

// ScanRunes returns an iterator over the runes of r. If reading the
// next rune returns an error, the iterator will yield a non-nil error
// and then exit.
//
// If the iterator is terminated early, it will unread
// the last rune read, allowing it to be used again to continue from
// where it left off. If this is not the desired behavior, use
// [ReadRunes] instead.
func ScanRunes(r io.RuneScanner) iter.Seq2[rune, error] {
	return reader(
		func() (rune, error) {
			c, _, err := r.ReadRune()
			return c, err
		},
		func() { r.UnreadRune() },
	)
}
