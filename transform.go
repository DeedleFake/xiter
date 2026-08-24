package xiter

import (
	"cmp"
	"iter"
	"slices"

	"deedles.dev/xiter/internal/xheap"
)

// Map returns a Seq that yields the values of seq transformed via f.
func (seq Seq[T1]) Map[T2 any](f func(T1) T2) Seq[T2] {
	return func(yield func(T2) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}

// Filter returns a Seq that yields only the values of seq for which
// f(value) returns true.
func (seq Seq[T]) Filter(f func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !f(v) {
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Skip returns a Seq that skips over the first n elements of seq and
// then yields the rest normally.
func (seq Seq[T]) Skip(n int) Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if n > 0 {
				n--
				continue
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Handle splits seq by calling f for any non-nil errors yielded by
// seq. If f returns false, iteration stops. If an iteration's error
// is nil or f returns true, the other value is yielded by the
// returned Seq.
//
// TODO: This is significantly less useful than it could be. For
// example, there's no way to tell it to skip the yield but continue
// iteration anyways.
func Handle[T any](seq Seq2[T, error], f func(error) bool) Seq[T] {
	return func(yield func(T) bool) {
		for v, err := range seq {
			if err != nil && !f(err) {
				return
			}
			if !yield(v) {
				return
			}
		}
	}
}

// Limit returns a Seq that yields at most n values from seq.
func (seq Seq[T]) Limit(n int) Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !yield(v) {
				return
			}
			n--
			if n <= 0 {
				return
			}
		}
	}
}

// Concat creates a new Seq that yields the values of each of the
// provided Seqs in turn.
func Concat[T any, S SeqLike[T]](seqs ...S) Seq[T] {
	return Flatten(Of(seqs...))
}

// Flatten yields all of the elements of each Seq yielded from seq in
// turn.
func Flatten[T any, S SeqLike[T]](seq Seq[S]) Seq[T] {
	return func(yield func(T) bool) {
		for s := range seq {
			for v := range s {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// Zipped holds values from an iteration of a Seq returned by [Zip].
type Zipped[T1, T2 any] struct {
	V1  T1
	OK1 bool

	V2  T2
	OK2 bool
}

// Zip returns a new Seq that yields the values of seq1 and seq2
// simultaneously.
//
// Must not be a method. See https://github.com/golang/go/issues/80172.
func Zip[T1, T2 any](seq1 Seq[T1], seq2 Seq[T2]) Seq[Zipped[T1, T2]] {
	return func(yield func(Zipped[T1, T2]) bool) {
		next2, stop2 := iter.Pull(seq2.Seq())
		defer stop2()

		for v1 := range seq1 {
			v2, ok2 := next2()
			val := Zipped[T1, T2]{
				V1: v1, OK1: true,
				V2: v2, OK2: ok2,
			}
			if !yield(val) {
				return
			}
		}

		for v2, ok2 := next2(); ok2; v2, ok2 = next2() {
			val := Zipped[T1, T2]{V2: v2, OK2: ok2}
			if !yield(val) {
				return
			}
		}
	}
}

// Merge returns a sequence that yields values from the ordered
// sequences seq1 and seq2 one at a time to produce a new ordered
// sequence made up of all of the elements of both seq1 and seq2.
func Merge[T cmp.Ordered](seq1, seq2 Seq[T]) Seq[T] {
	return MergeFunc(seq1, seq2, cmp.Compare)
}

// MergeFunc is like [Merge], but uses a custom comparison function
// for determining the order of values.
func MergeFunc[T any](seq1, seq2 Seq[T], compare func(T, T) int) Seq[T] {
	return func(yield func(T) bool) {
		p1, stop := iter.Pull(seq1.Seq())
		defer stop()
		p2, stop := iter.Pull(seq2.Seq())
		defer stop()

		v1, ok1 := p1()
		v2, ok2 := p2()
		for ok1 || ok2 {
			var c int
			if ok1 && ok2 {
				c = compare(v1, v2)
			}

			switch {
			case !ok2 || c < 0:
				if !yield(v1) {
					return
				}
				v1, ok1 = p1()
			case !ok1 || c > 0:
				if !yield(v2) {
					return
				}
				v2, ok2 = p2()
			default:
				if !yield(v1) || !yield(v2) {
					return
				}
				v1, ok1 = p1()
				v2, ok2 = p2()
			}
		}
	}
}

// Windows returns a slice over successive overlapping portions of
// size n of the values yielded by seq. In other words,
//
//	Windows(Generate(0, 1), 3)
//
// will yield
//
//	[0, 1, 2]
//	[1, 2, 3]
//	[2, 3, 4]
//
// and so on. The slice yielded is reused from one iteration to the
// next, so it should not be held onto after each iteration has ended.
// [Seq.Map] and [slices.Clone] may come in handy for dealing with
// situations where this is necessary.
//
// Must not be a method. See https://github.com/golang/go/issues/80172.
func Windows[T any](seq Seq[T], n int) Seq[[]T] {
	return func(yield func([]T) bool) {
		win := make([]T, 0, n)

		for v := range seq {
			if len(win) < n-1 {
				win = append(win, v)
				continue
			}
			if len(win) < n {
				win = append(win, v)
				if !yield(win) {
					return
				}
				continue
			}

			copy(win, win[1:])
			win[len(win)-1] = v
			if !yield(win) {
				return
			}
		}
		if len(win) < n {
			yield(win)
		}
	}
}

// Chunks works just like [Windows] except that the yielded slices of
// elements do not overlap. In other words,
//
//	Chunks(Generate(0, 1), 3)
//
// will yield
//
//	[0, 1, 2]
//	[3, 4, 5]
//	[6, 7, 8]
//
// Like with Windows, the slice is reused between iterations.
//
// Must not be a method. See https://github.com/golang/go/issues/80172.
func Chunks[T any](seq Seq[T], n int) Seq[[]T] {
	return func(yield func([]T) bool) {
		win := make([]T, 0, n)

		for v := range seq {
			if len(win) == n {
				clear(win)
				win = win[:0]
			}

			if len(win) < n-1 {
				win = append(win, v)
				continue
			}
			if len(win) < n {
				win = append(win, v)
				if !yield(win) {
					return
				}
				continue
			}

			// This should only be reachable if n is 0, so just yield a
			// bunch of empty slices because why not?
			if !yield(win) {
				return
			}
		}
		if len(win) < n {
			yield(win)
		}
	}
}

// ChunksFunc is like [Chunks], except chunk boundaries are determined
// by calling chunker on successive elements. When the return value of
// the function changes from the previous call, a new chunk is started.
//
// Like with Chunks, the slice is reused between iterations.
//
// Must not be a method. See https://github.com/golang/go/issues/80172.
func ChunksFunc[T any, C comparable](seq Seq[T], chunker func(T) C) Seq[[]T] {
	return func(yield func([]T) bool) {
		next, stop := iter.Pull(seq.Seq())
		defer stop()

		cur, ok := next()
		if !ok {
			return
		}
		prev := chunker(cur)
		win := []T{cur}

		for {
			cur, ok := next()
			if !ok {
				if len(win) != 0 {
					yield(win)
				}
				return
			}

			check := chunker(cur)
			if check == prev {
				win = append(win, cur)
				continue
			}

			if !yield(slices.Clip(win)) {
				return
			}
			clear(win)
			win = win[:1]
			win[0] = cur

			prev = check
		}
	}
}

// Split returns a SplitSeq which yields the values of seq for which
// f(value) is true to its first yield function and the rest to its
// second.
func (seq Seq[T]) Split(f func(T) bool) SplitSeq[T, T] {
	return func(true, false func(T) bool) {
		for v := range seq {
			y := false
			if f(v) {
				y = true
			}
			if !y(v) {
				return
			}
		}
	}
}

// Split2 transforms a Seq2 into a SplitSeq. Every iteration of the
// Seq2 yields both values via the SplitSeq.
func (seq Seq2[T1, T2]) Split2() SplitSeq[T1, T2] {
	return func(y1 func(T1) bool, y2 func(T2) bool) {
		for v1, v2 := range seq {
			if !y1(v1) || !y2(v2) {
				return
			}
		}
	}
}

// Cache returns a Seq that can be iterated more than once. On the
// first iteration, it yields the values from seq and caches them. On
// subsequent iterations, it yields the cached values from the first
// iteration.
func (seq Seq[T]) Cache() Seq[T] {
	var cache []T
	return func(yield func(T) bool) {
		if cache != nil {
			for _, v := range cache {
				if !yield(v) {
					return
				}
			}
			return
		}

		cache = []T{}
		for v := range seq {
			cache = append(cache, v)
			if !yield(v) {
				return
			}
		}
	}
}

// Enumerate returns a Seq2 that counts the number of iterations of
// seq as it yields elements from it, starting at 0.
func (seq Seq[T]) Enumerate() Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := -1
		for v := range seq {
			i++
			if !yield(i, v) {
				return
			}
		}
	}
}

// Or yields all of the values from the first Seq which yields at
// least one value and then stops.
func Or[T any, S SeqLike[T]](seqs ...S) Seq[T] {
	ss := Of(seqs...).Filter(func(s S) bool { return s != nil })
	return func(yield func(T) bool) {
		for seq := range ss {
			empty := true
			for v := range seq {
				empty = false
				if !yield(v) {
					return
				}
			}
			if !empty {
				return
			}
		}
	}
}

// Dedup returns an iterator that only yields each unique element from
// seq once. Note that to do this, it stores a set of all elements
// that have been seen, so this iterator can use a large amount of
// memory if seq yields a very large number of unique elements.
func Dedup[T comparable](seq Seq[T]) Seq[T] {
	return func(yield func(T) bool) {
		found := make(map[T]struct{})
		for v := range seq {
			if _, ok := found[v]; ok {
				continue
			}
			found[v] = struct{}{}
			if !yield(v) {
				return
			}
		}
	}
}

// Uniq returns an iterator that removes consecutive duplicates from
// seq. It is similar to [Dedup], but non-consecutive duplicates are
// not filtered out. Unlike Dedup, it does not store all found values,
// and so does not have the same performance implication that Dedup
// does.
func Uniq[T comparable](seq Seq[T]) Seq[T] {
	return seq.UniqFunc(func(v1, v2 T) bool { return v1 == v2 })
}

// UniqFunc is like [Uniq] but uses the provided comparison function
// to check for duplicates.
func (seq Seq[T]) UniqFunc(eq func(T, T) bool) Seq[T] {
	return func(yield func(T) bool) {
		next, stop := iter.Pull(seq.Seq())
		defer stop()

		cur, ok := next()
		if !ok || !yield(cur) {
			return
		}

		for {
			v, ok := next()
			if !ok {
				return
			}
			if eq(v, cur) {
				continue
			}

			if !yield(v) {
				return
			}
			cur = v
		}
	}
}

// Sorted collects the entirety of seq and then returns a one-time use
// iterator which yields the elements of seq in a sorted order.
func Sorted[T cmp.Ordered](seq Seq[T]) Seq[T] {
	s := seq.Collect()
	xheap.Init(s)

	return func(yield func(T) bool) {
		for len(s) > 0 {
			var v T
			v, s = xheap.Pop(s)
			if !yield(v) {
				return
			}
		}
	}
}

// SortedFunc collects the entirety of seq and then returns a one-time
// use iterator which yields the elements of seq in a sorted order
// determined by the provided comparison function.
func (seq Seq[T]) SortedFunc(compare func(T, T) int) Seq[T] {
	s := seq.Collect()
	xheap.InitFunc(s, compare)

	return func(yield func(T) bool) {
		for len(s) > 0 {
			var v T
			v, s = xheap.PopFunc(s, compare)
			if !yield(v) {
				return
			}
		}
	}
}

// ToPair takes a two-value iterator and produces a single-value
// iterator of pairs.
//
// Must not be a method. See https://github.com/golang/go/issues/80172.
func ToPair[T1, T2 any](seq Seq2[T1, T2]) Seq[Pair[T1, T2]] {
	return func(yield func(Pair[T1, T2]) bool) {
		for v1, v2 := range seq {
			if !yield(P(v1, v2)) {
				return
			}
		}
	}
}

// V1 returns a Seq which iterates over only the T1 elements of seq.
func (seq Seq2[T1, T2]) V1() Seq[T1] {
	return func(yield func(T1) bool) {
		for v1 := range seq {
			if !yield(v1) {
				return
			}
		}
	}
}

// V2 returns a Seq which iterates over only the T2 elements of seq.
func (seq Seq2[T1, T2]) V2() Seq[T2] {
	return func(yield func(T2) bool) {
		for _, v2 := range seq {
			if !yield(v2) {
				return
			}
		}
	}
}

// FromPair converts a Seq of pairs to a two-value Seq.
//
// Must not be a method. See https://github.com/golang/go/issues/80172.
func FromPair[T1, T2 any](seq Seq[Pair[T1, T2]]) Seq2[T1, T2] {
	return func(yield func(T1, T2) bool) {
		for v := range seq {
			if !yield(v.Split()) {
				return
			}
		}
	}
}
