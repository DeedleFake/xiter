package xiter

import (
	"context"
	"io"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"
	"unsafe"
)

// Generate returns a Seq that first yields start and then yields
// successive values by adding step to the previous continuously. The
// returned Seq does not end. To limit it to a specific number of
// returned elements, use [Seq.Limit].
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
	return S(slices.Values(vals))
}

// Bytes returns a Seq over the bytes of s.
func Bytes(s string) Seq[byte] {
	return func(yield func(byte) bool) {
		for i := 0; i < len(s); i++ {
			if !yield(s[i]) {
				return
			}
		}
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
	}
}

// StringSplit returns an iterator over the substrings of s that are
// separated by sep. It behaves very similarly to [strings.Split].
func StringSplit(s, sep string) Seq[string] {
	if sep == "" {
		return Runes(s).Map(func(c rune) string { return string(c) })
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
		for i, r := range Runes(s).Enumerate() {
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

// OfChan returns a Seq which yields values received from c. The
// sequence ends when c is closed. It is equivalent to range c.
func OfChan[T any](c <-chan T) Seq[T] {
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

// SliceChunksFunc is like [ChunksFunc] but operates on a slice
// instead of a [Seq]. When dealing with data that is in a
// slice, this is more effecient than using ChunksFunc as it can yield
// subslices of the underlying slice instead of having to allocate a
// moving window. The yielded subslices have their capacity clipped.
func SliceChunksFunc[T any, C comparable, S ~[]T](s S, chunker func(T) C) Seq[S] {
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
func reader[T byte | rune](read func() (T, error), done func()) Seq2[T, error] {
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
func ReadBytes(r io.ByteReader) Seq2[byte, error] {
	return reader(
		r.ReadByte,
		func() {},
	)
}

// ReadRunes returns an iterator over the runes of r. If reading the
// next rune returns an error, the iterator will yield a non-nil error
// and then exit.
func ReadRunes(r io.RuneReader) Seq2[rune, error] {
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
func ScanBytes(r io.ByteScanner) Seq2[byte, error] {
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
func ScanRunes(r io.RuneScanner) Seq2[rune, error] {
	return reader(
		func() (rune, error) {
			c, _, err := r.ReadRune()
			return c, err
		},
		func() { r.UnreadRune() },
	)
}
