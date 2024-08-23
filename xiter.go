// Package xiter provides iterator-related functionality compatible
// with, but not requiring, Go 1.23.
package xiter

// A SplitSeq is like a Seq but can yield via either of two functions.
// It might not be useful, but is included anyways because it might
// be.
type SplitSeq[T1, T2 any] func(y1 func(T1) bool, y2 func(T2) bool)

// Pair contains two values of arbitrary types.
type Pair[T1, T2 any] struct {
	V1 T1
	V2 T2
}

// P returns a Pair containing v1 and v2.
func P[T1, T2 any](v1 T1, v2 T2) Pair[T1, T2] {
	return Pair[T1, T2]{V1: v1, V2: v2}
}

// Split is a convenience function that just returns the two values
// contained in the pair.
func (p Pair[T1, T2]) Split() (T1, T2) {
	return p.V1, p.V2
}

// Addable is a type that should probably exist in the standard
// library somewhere because it's quite command and kind of a pain to
// write every time I need it.
type Addable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | complex64 | complex128 | string
}

type Multiplyable interface {
	int | int8 | int16 | int32 | int64 | uint | uint8 | uint16 | uint32 | uint64 | uintptr | float32 | float64 | complex64 | complex128
}
