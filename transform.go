package xiter

// Map returns a Seq that yields the values of seq transformed via f.
func Map[T1, T2 any](seq Seq[T1], f func(T1) T2) Seq[T2] {
	return func(yield func(T2) bool) bool {
		return seq(func(v T1) bool {
			return yield(f(v))
		})
	}
}

// Filter returns a Seq that yields only the values of seq for which
// f(value) returns true.
func Filter[T any](seq Seq[T], f func(T) bool) Seq[T] {
	return func(yield func(T) bool) bool {
		return seq(func(v T) bool {
			if !f(v) {
				return true
			}
			return yield(v)
		})
	}
}

// Limit returns a Seq that yields at most n values from seq.
func Limit[T any](seq Seq[T], n int) Seq[T] {
	return func(yield func(T) bool) bool {
		return seq(func(v T) bool {
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
	return func(yield func(T) bool) bool {
		for _, seq := range seqs {
			seq(yield)
		}
		return false
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
	return func(yield func(Zipped[T1, T2]) bool) bool {
		p1, stop := Pull(seq1)
		defer stop()
		p2, stop := Pull(seq2)
		defer stop()

		for {
			var val Zipped[T1, T2]
			val.V1, val.OK1 = p1()
			val.V2, val.OK2 = p2()
			if (!val.OK1 && !val.OK2) || !yield(val) {
				return false
			}
		}
	}
}
