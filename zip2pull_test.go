package xiter

import (
	"iter"
	"testing"
)

func BenchmarkZip2Pull(b *testing.B) {
	benchmarkZip(b, zip2Pull)
}

// zip2Pull returns a new Seq that yields the values of seq1 and seq2
// simultaneously.  This is the straightforward 2-Pull version.
func zip2Pull[T1, T2 any](seq1 Seq[T1], seq2 Seq[T2]) Seq[Zipped[T1, T2]] {
	return func(yield func(Zipped[T1, T2]) bool) {
		p1, stop := iter.Pull(seq1.Seq())
		defer stop()
		p2, stop := iter.Pull(seq2.Seq())
		defer stop()

		for {
			var val Zipped[T1, T2]
			val.V1, val.OK1 = p1()
			val.V2, val.OK2 = p2()
			if (!val.OK1 && !val.OK2) || !yield(val) {
				return
			}
		}
	}
}
