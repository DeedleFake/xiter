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

func send[T any](done <-chan struct{}, c chan<- T) func(v T) bool {
	return func(v T) bool {
		select {
		case <-done:
			return false
		case c <- v:
			return true
		}
	}
}

// Zip returns a new Seq that yields the values of seq1 and seq2
// simultaneously.
//
// Important thing to note: Because of the lack of corountine support,
// seq1 and seq2 are called in new threads. This means not only is
// there some performance overhead for simple cases, but also that
// they should not do anything that isn't thread-safe without some
// other form of synchronization.
func Zip[T1, T2 any](seq1 Seq[T1], seq2 Seq[T2]) Seq[Zipped[T1, T2]] {
	done := make(chan struct{})

	c1 := make(chan T1)
	go func() {
		defer close(c1)
		seq1(send(done, c1))
	}()

	c2 := make(chan T2)
	go func() {
		defer close(c2)
		seq2(send(done, c2))
	}()

	return func(yield func(Zipped[T1, T2]) bool) bool {
		defer close(done)

		var c1hold chan T1
		var c2hold chan T2
		var val Zipped[T1, T2]

		for {
			select {
			case v, ok := <-c1:
				if !ok {
					c1 = nil
					goto send
				}
				c1, c1hold = nil, c1
				val.V1 = v
				val.OK1 = true
			case v, ok := <-c2:
				if !ok {
					c2 = nil
					goto send
				}
				c2, c2hold = nil, c2
				val.V2 = v
				val.OK2 = true
			}

		send: // Dear lord, what is even going on here?
			if (val.OK1 || (c1 == c1hold)) && (val.OK2 || (c2 == c2hold)) && !(c1==c1hold && c2==c2hold) {
				if !yield(val) {
					return false
				}
				c1, c1hold, c2, c2hold = c1hold, nil, c2hold, nil
				val = Zipped[T1, T2]{}
			}

			if (c1 == c1hold) && (c2 == c2hold) {
				return false
			}
		}
	}
}
