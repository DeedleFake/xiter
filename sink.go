package xiter

import (
	"cmp"
	"context"
	"iter"
	"slices"
)

// CollectSize pre-allocates the slice being collected into to the
// given size. It is provided purely for convenience.
func CollectSize[T any](seq iter.Seq[T], len int) []T {
	return slices.AppendSeq(make([]T, 0, len), seq)
}

// Find returns the first value of seq for which f(value) returns
// true.
func Find[T any](seq iter.Seq[T], f func(T) bool) (r T, ok bool) {
	seq(func(v T) bool {
		if !f(v) {
			return true
		}
		r = v
		ok = true
		return false
	})
	return r, ok
}

// Contains returns true if v is an element of seq.
func Contains[T comparable](seq iter.Seq[T], v T) bool {
	_, ok := Find(seq, func(e T) bool { return v == e })
	return ok
}

// Any returns true if f(element) is true for any elements of seq.
func Any[T any](seq iter.Seq[T], f func(T) bool) bool {
	_, ok := Find(seq, f)
	return ok
}

// All returns true if f(element) is true for every element of seq.
func All[T any](seq iter.Seq[T], f func(T) bool) bool {
	return !Any(seq, f)
}

// Reduce calls reducer on each value of seq, passing it initial as
// its first argument on the first call and then the result of the
// previous call for each call after that. It returns the final value
// returned by reducer.
//
// Reduce can be somewhat complicated to get the hang of, but very
// powerful. For example, a simple summation of values can be done as
//
//	sum := Reduce(seq, 0, func(total, v int) int { return total + v })
func Reduce[T, R any](seq iter.Seq[T], initial R, reducer func(R, T) R) R {
	seq(func(v T) bool {
		initial = reducer(initial, v)
		return true
	})
	return initial
}

// Fold performs a [Reduce] but uses the first value yielded by seq
// instead of a provided initial value. If seq doesn't yield any
// values, the zero value of T is returned.
func Fold[T any](seq iter.Seq[T], reducer func(T, T) T) T {
	var prev T
	r := func(v1, v2 T) T { return v2 }
	seq(func(v T) bool {
		prev = r(prev, v)
		r = reducer
		return true
	})
	return prev
}

// Sum returns the values of seq added together in the order that they
// are yielded.
func Sum[T Addable](seq iter.Seq[T]) T {
	return Fold(seq, func(total, v T) T { return total + v })
}

// Product returns the values of seq multiplied together. It returns
// 1 if no values are yielded.
func Product[T Multiplyable](seq iter.Seq[T]) T {
	return Reduce(seq, 1, func(product, v T) T { return product * v })
}

// IsSorted returns true if each element of seq is greater than or
// equal to the previous one.
func IsSorted[T cmp.Ordered](seq iter.Seq[T]) bool {
	return IsSortedFunc(seq, cmp.Compare)
}

// IsSortedFunc is like [IsSorted] but uses a custom comparison
// function.
func IsSortedFunc[T any](seq iter.Seq[T], compare func(T, T) int) bool {
	var prev T
	c := func(T, T) int { return -1 }

	sorted := true
	seq(func(v T) bool {
		sorted = c(prev, v) <= 0
		c, prev = compare, v
		return sorted
	})
	return sorted
}

// Equal returns true if seq1 and seq2 are the same length and each
// element of each is equal to the element at the same point in the
// sequence of the other.
func Equal[T cmp.Ordered](seq1, seq2 iter.Seq[T]) bool {
	return EqualFunc(seq1, seq2, func(v1, v2 T) bool { return v1 == v2 })
}

// EqualFunc is like [Equal] but uses a custom comparison function to
// determine the equivalence of the elements of each sequence.
func EqualFunc[T1, T2 any](seq1 iter.Seq[T1], seq2 iter.Seq[T2], equal func(T1, T2) bool) bool {
	p1, stop := iter.Pull(seq1)
	defer stop()
	p2, stop := iter.Pull(seq2)
	defer stop()

	for {
		v1, ok1 := p1()
		v2, ok2 := p2()
		if !ok1 && !ok2 {
			return true
		}
		if (ok1 != ok2) || !equal(v1, v2) {
			return false
		}
	}
}

// Drain empties seq, returning the last value yielded. If no values
// are yielded, ok will be false.
func Drain[T any](seq iter.Seq[T]) (v T, ok bool) {
	seq(func(val T) bool {
		v = val
		ok = true
		return true
	})
	return v, ok
}

// CollectSplit is like [Collect], but for a SplitSeq.
func CollectSplit[T1, T2 any](seq SplitSeq[T1, T2]) (y1 []T1, y2 []T2) {
	return AppendSplitTo(seq, y1, y2)
}

// AppendSplitTo collects the elements of seq by appending them to
// existing slices.
func AppendSplitTo[T1, T2 any](seq SplitSeq[T1, T2], s1 []T1, s2 []T2) ([]T1, []T2) {
	seq(
		func(v T1) bool {
			s1 = append(s1, v)
			return true
		},
		func(v T2) bool {
			s2 = append(s2, v)
			return true
		},
	)
	return s1, s2
}

// Partition returns two slices, one containing all of the elements of
// seq for which f(element) is true and one containing all of those
// for which it is false.
func Partition[T any](seq iter.Seq[T], f func(T) bool) (true, false []T) {
	return PartitionInto(seq, f, true, false)
}

// PartitionInto performs a [Partition] by appending to two existing
// slices.
func PartitionInto[T any](seq iter.Seq[T], f func(T) bool, true, false []T) ([]T, []T) {
	return AppendSplitTo(Split(seq, f), true, false)
}

// Min returns the minimum element yielded by seq or the zero value if
// seq doesn't yield anything.
func Min[T cmp.Ordered](seq iter.Seq[T]) T {
	return Fold(seq, func(v1, v2 T) T { return min(v1, v2) })
}

// Max returns maximum element yielded by seq or the zero value if seq
// doesn't yield anything.
func Max[T cmp.Ordered](seq iter.Seq[T]) T {
	return Fold(seq, func(v1, v2 T) T { return max(v1, v2) })
}

// FromPair converts a Seq of pairs to a two-value Seq.
func FromPair[T1, T2 any](seq iter.Seq[Pair[T1, T2]]) iter.Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		seq(func(v Pair[T1, T2]) bool {
			return yield(v.Split())
		})
	}
}

// SendContext sends values from seq to c repeatedly until either the
// sequence ends or ctx is canceled. It blocks until one of those two
// things happens.
func SendContext[T any](seq iter.Seq[T], ctx context.Context, c chan<- T) {
	seq(func(v T) bool {
		select {
		case <-ctx.Done():
			return false
		case c <- v:
			return true
		}
	})
}
