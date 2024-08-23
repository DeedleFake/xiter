package xiter

import (
	"bytes"
	"cmp"
	"iter"
	"slices"
	"testing"
)

func TestMap(t *testing.T) {
	s := slices.Values([]int{1, 2, 3})
	n := slices.Collect(Map(s, func(v int) float64 { return float64(v * 2) }))
	if [3]float64(n) != [...]float64{2, 4, 6} {
		t.Fatal(n)
	}
}

func TestFilter(t *testing.T) {
	s := slices.Values([]int{1, 2, 3})
	n := slices.Collect(Filter(s, func(v int) bool { return v%2 != 0 }))
	if [2]int(n) != [...]int{1, 3} {
		t.Fatal(n)
	}
}

func TestSkip(t *testing.T) {
	s := slices.Collect(Skip(Limit(Generate(
		0, 1),
		3),
		2),
	)
	if !Equal(slices.Values(s), Of(2)) {
		t.Fatal(s)
	}
}

func TestLimit(t *testing.T) {
	s := slices.Collect(Limit(Generate(
		0, 2),
		3),
	)
	if [3]int(s) != [...]int{0, 2, 4} {
		t.Fatal(s)
	}
}

func TestConcat(t *testing.T) {
	s := slices.Collect(Concat(slices.Values([]int{1, 2, 3}), slices.Values([]int{3, 2, 5})))
	if [6]int(s) != [...]int{1, 2, 3, 3, 2, 5} {
		t.Fatal(s)
	}

	s = Chain[int](slices.Values([]int{1, 2, 3})).Concat(slices.Values([]int{3, 2, 5})).Collect()
	if [6]int(s) != [...]int{1, 2, 3, 3, 2, 5} {
		t.Fatal(s)
	}
}

func TestZip(t *testing.T) {
	s1 := slices.Values([]int{1, 2, 3, 4, 5})
	s2 := slices.Values([]int{2, 3, 4, 5, 6})
	seq := Zip(s1, s2)
	seq(func(v Zipped[int, int]) bool {
		if v.V2-v.V1 != 1 {
			t.Fatalf("unexpected values: %+v", v)
		}
		return true
	})
}

func BenchmarkZip(b *testing.B) {
	slice1 := []int{1, 2, 3, 4, 5}
	slice2 := []int{2, 3, 4, 5, 6}

	for i := 0; i < b.N; i++ {
		s1 := slices.Values(slice1)
		s2 := slices.Values(slice2)
		seq := Zip(s1, s2)
		seq(func(v Zipped[int, int]) bool {
			return true
		})
	}
}

func TestIsSorted(t *testing.T) {
	if IsSorted(slices.Values([]int{1, 2, 3, 2})) {
		t.Fatal("is not sorted")
	}
	if !IsSorted(slices.Values([]int{1, 2, 3, 4, 5})) {
		t.Fatal("is sorted")
	}
	if !IsSorted(slices.Values([]int{48, 48})) {
		t.Fatal("is sorted")
	}
}

func TestMerge(t *testing.T) {
	s1 := slices.Values([]int{2, 3, 5})
	s2 := slices.Values([]int{1, 2, 3, 4, 5})
	r := slices.Collect(Merge(s1, s2))
	if [8]int(r) != [...]int{1, 2, 2, 3, 3, 4, 5, 5} {
		t.Fatal(r)
	}
}

func splitmerge[T cmp.Ordered](s []T) iter.Seq[T] {
	if len(s) <= 1 {
		return slices.Values(s)
	}

	left := splitmerge(s[:len(s)/2])
	right := splitmerge(s[len(s)/2:])
	return Merge(left, right)
}

func mergesort[T cmp.Ordered](s []T) {
	slices.AppendSeq(s[:0], splitmerge(s))
}

func TestMergeSort(t *testing.T) {
	s := []int{3, 2, 5, 1, 6, 2}
	mergesort(s)
	if [6]int(s) != [...]int{1, 2, 2, 3, 5, 6} {
		t.Fatal(s)
	}
}

func FuzzMergeSort(f *testing.F) {
	f.Add([]byte("The quick brown fox jumped over the lazy dog."))
	f.Fuzz(func(t *testing.T, s []byte) {
		check := bytes.Clone(s)
		slices.Sort(check)

		mergesort(s)
		if !Equal(slices.Values(s), slices.Values(check)) {
			t.Fatal(s)
		}
	})
}

func TestChunks(t *testing.T) {
	s := slices.Collect(Map(Chunks(slices.Values([]int{1, 2, 3, 4, 5}),
		2),
		slices.Clone),
	)
	if !slices.EqualFunc(s, [][]int{{1, 2}, {3, 4}, {5}}, slices.Equal) {
		t.Fatal(s)
	}
}

func TestChunksFunc(t *testing.T) {
	s := slices.Collect(Map(ChunksFunc(slices.Values([]int{0, 0, 0, 0, 1, 0, 1, 1, 0, 1}),
		func(v int) bool { return v%2 == 0 }),
		slices.Clone),
	)
	if !slices.EqualFunc(s, [][]int{{0, 0, 0, 0}, {1}, {0}, {1, 1}, {0}, {1}}, slices.Equal) {
		t.Fatal(s)
	}
}

func TestSplit2(t *testing.T) {
	s1, s2 := CollectSplit(Split2(FromPair(slices.Values([]Pair[int32, int64]{{1, 2}, {3, 4}, {5, 6}}))))
	if !slices.Equal(s1, []int32{1, 3, 5}) {
		t.Fatal(s1)
	}
	if !slices.Equal(s2, []int64{2, 4, 6}) {
		t.Fatal(s2)
	}
}

func TestCache(t *testing.T) {
	var i int
	f := func(yield func(int) bool) {
		yield(i)
		i++
		return
	}
	seq := Cache(f)
	if s := slices.Collect(seq); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
	if s := slices.Collect(seq); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
}

func TestEnumerate(t *testing.T) {
	s := slices.Collect(ToPair(Enumerate(Limit(Generate(0, 2), 3))))
	if !slices.Equal(s, []Pair[int, int]{{0, 0}, {1, 2}, {2, 4}}) {
		t.Fatal(s)
	}
}

func TestOr(t *testing.T) {
	s := slices.Collect(Or(Of[int](), nil, Of(1, 2, 3), Of(4, 5, 6)))
	if !slices.Equal(s, []int{1, 2, 3}) {
		t.Fatal(s)
	}

	s = Chain[int](Of[int]()).Or(nil, Of(1, 2, 3), Of(4, 5, 6)).Collect()
	if !slices.Equal(s, []int{1, 2, 3}) {
		t.Fatal(s)
	}
}
