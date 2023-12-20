package xiter

import "cmp"

// Map returns a Seq that yields the values of seq transformed via f.
func Map[T1, T2 any](seq Seq[T1], f func(T1) T2) Seq[T2] {
	return func(yield func(T2) bool) {
		seq(func(v T1) bool {
			return yield(f(v))
		})
	}
}

// Filter returns a Seq that yields only the values of seq for which
// f(value) returns true.
func Filter[T any](seq Seq[T], f func(T) bool) Seq[T] {
	return func(yield func(T) bool) {
		seq(func(v T) bool {
			if !f(v) {
				return true
			}
			return yield(v)
		})
	}
}

// Skip returns a Seq that skips over the first n elements of seq and
// then yields the rest normally.
func Skip[T any](seq Seq[T], n int) Seq[T] {
	return func(yield func(T) bool) {
		seq(func(v T) bool {
			if n > 0 {
				n--
				return true
			}
			return yield(v)
		})
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
		seq(func(v T, err error) bool {
			if err != nil {
				return f(err) && yield(v)
			}
			return yield(v)
		})
	}
}

// Limit returns a Seq that yields at most n values from seq.
func Limit[T any](seq Seq[T], n int) Seq[T] {
	return func(yield func(T) bool) {
		seq(func(v T) bool {
			if !yield(v) {
				return false
			}
			n--
			return n > 0
		})
	}
}

// Concat creates a new Seq that yields the values of each of the
// provided Seqs in turn.
func Concat[T any](seqs ...Seq[T]) Seq[T] {
	return Flatten(OfSlice(seqs))
}

// Flatten yields all of the elements of each Seq yielded from seq in
// turn.
func Flatten[T any](seq Seq[Seq[T]]) Seq[T] {
	return func(yield func(T) bool) {
		seq(func(s Seq[T]) bool {
			s(yield)
			return true
		})
		return
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
func Zip[T1, T2 any](seq1 Seq[T1], seq2 Seq[T2]) Seq[Zipped[T1, T2]] {
	return func(body func(Zipped[T1, T2]) bool) {
		p2, stop2 := Pull(seq2)
		defer stop2()
		done := false

		f := func(v T1) bool {
			var val Zipped[T1, T2]
			val.V1, val.OK1 = v, true
			val.V2, val.OK2 = p2()
			if !body(val) {
				done = true
				return false
			}
			return true
		}
		seq1(f)
		if done {
			return
		}
		// seq1 is exhausted
		for v2, ok2 := p2(); ok2; v2, ok2 = p2() {
			var v1 T1
			var val Zipped[T1, T2]
			val.V1, val.OK1 = v1, false
			val.V2, val.OK2 = v2, true
			if !body(val) {
				return
			}
		}
		return
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
		p1, stop := Pull(seq1)
		defer stop()
		p2, stop := Pull(seq2)
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

		return
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
// [Map] and [slices.Clone] may come in handy for dealing with
// situations where this is necessary.
func Windows[T any](seq Seq[T], n int) Seq[[]T] {
	return func(yield func([]T) bool) {
		win := make([]T, 0, n)

		seq(func(v T) bool {
			if len(win) < n-1 {
				win = append(win, v)
				return true
			}
			if len(win) < n {
				win = append(win, v)
				return yield(win)
			}

			copy(win, win[1:])
			win[len(win)-1] = v
			return yield(win)
		})
		if len(win) < n {
			yield(win)
		}
		return
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
func Chunks[T any](seq Seq[T], n int) Seq[[]T] {
	return func(yield func([]T) bool) {
		win := make([]T, 0, n)

		seq(func(v T) bool {
			if len(win) == n {
				clear(win)
				win = win[:0]
			}

			if len(win) < n-1 {
				win = append(win, v)
				return true
			}
			if len(win) < n {
				win = append(win, v)
				return yield(win)
			}

			// This should only be reachable if n is 0, so just yield a
			// bunch of empty slices because why not?
			return yield(win)
		})
		if len(win) < n {
			yield(win)
		}
		return
	}
}

// Split returns a SplitSeq which yields the values of seq for which
// f(value) is true to its first yield function and the rest to its
// second.
func Split[T any](seq Seq[T], f func(T) bool) SplitSeq[T, T] {
	return func(true, false func(T) bool) {
		seq(func(v T) bool {
			y := false
			if f(v) {
				y = true
			}
			return y(v)
		})
	}
}

// Split2 transforms a Seq2 into a SplitSeq. Every iteration of the
// Seq2 yields both values via the SplitSeq.
func Split2[T1, T2 any](seq Seq2[T1, T2]) SplitSeq[T1, T2] {
	return func(y1 func(T1) bool, y2 func(T2) bool) {
		seq(func(v1 T1, v2 T2) bool {
			return y1(v1) && y2(v2)
		})
	}
}

// Cache returns a Seq that can be iterated more than once. On the
// first iteration, it yields the values from seq and caches them. On
// subsequent iterations, it yields the cached values from the first
// iteration.
func Cache[T any](seq Seq[T]) Seq[T] {
	var cache []T
	return func(yield func(T) bool) {
		if cache != nil {
			OfSlice(cache)(yield)
			return
		}

		cache = []T{}
		seq(func(v T) bool {
			cache = append(cache, v)
			return yield(v)
		})
	}
}

// Enumerate returns a Seq2 that counts the number of iterations of
// seq as it yields elements from it, starting at 0.
func Enumerate[T any](seq Seq[T]) Seq2[int, T] {
	return func(yield func(int, T) bool) {
		i := -1
		seq(func(v T) bool {
			i++
			return yield(i, v)
		})
	}
}

// Or yields all of the values from the first Seq which yields at
// least one value and then stops.
func Or[T any](seqs ...Seq[T]) Seq[T] {
	ss := Filter(OfSlice(seqs), func(s Seq[T]) bool { return s != nil })
	return func(yield func(T) bool) {
		ss(func(seq Seq[T]) bool {
			cont := true
			seq(func(v T) bool {
				cont = false
				return yield(v)
			})
			return cont
		})
	}
}
