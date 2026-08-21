package xiter

import (
	"bytes"
	"cmp"
	"slices"
	"testing"
)

func TestMap(t *testing.T) {
	s := Of(1, 2, 3)
	n := slices.Collect(s.Map(func(v int) float64 { return float64(v * 2) }).Seq())
	if [3]float64(n) != [...]float64{2, 4, 6} {
		t.Fatal(n)
	}
}

func TestFilter(t *testing.T) {
	s := Of(1, 2, 3)
	n := slices.Collect(s.Filter(func(v int) bool { return v%2 != 0 }).Seq())
	if [2]int(n) != [...]int{1, 3} {
		t.Fatal(n)
	}
}

func TestSkip(t *testing.T) {
	s := slices.Collect(Generate(0, 1).Limit(3).Skip(2).Seq())
	if !Equal(Of(s...), Of(2)) {
		t.Fatal(s)
	}
}

func TestLimit(t *testing.T) {
	s := slices.Collect(Generate(0, 2).Limit(3).Seq())
	if [3]int(s) != [...]int{0, 2, 4} {
		t.Fatal(s)
	}
}

func TestConcat(t *testing.T) {
	s := slices.Collect(Concat(Of(1, 2, 3), Of(3, 2, 5)).Seq())
	if [6]int(s) != [...]int{1, 2, 3, 3, 2, 5} {
		t.Fatal(s)
	}

	s = slices.Collect(Concat(Of(1, 2, 3), Of(3, 2, 5)).Seq())
	if [6]int(s) != [...]int{1, 2, 3, 3, 2, 5} {
		t.Fatal(s)
	}
}

func TestZip(t *testing.T) {
	s1 := Of(1, 2, 3, 4, 5)
	s2 := Of(2, 3, 4, 5, 6)
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

	for b.Loop() {
		s1 := S(slices.Values(slice1))
		s2 := S(slices.Values(slice2))
		seq := Zip(s1, s2)
		seq(func(v Zipped[int, int]) bool {
			return true
		})
	}
}

func TestIsSorted(t *testing.T) {
	if IsSorted(Of(1, 2, 3, 2)) {
		t.Fatal("is not sorted")
	}
	if !IsSorted(Of(1, 2, 3, 4, 5)) {
		t.Fatal("is sorted")
	}
	if !IsSorted(Of(48, 48)) {
		t.Fatal("is sorted")
	}
}

func TestMerge(t *testing.T) {
	s1 := Of(2, 3, 5)
	s2 := Of(1, 2, 3, 4, 5)
	r := slices.Collect(Merge(s1, s2).Seq())
	if [8]int(r) != [...]int{1, 2, 2, 3, 3, 4, 5, 5} {
		t.Fatal(r)
	}
}

func splitmerge[T cmp.Ordered](s []T) Seq[T] {
	if len(s) <= 1 {
		return S(slices.Values(s))
	}

	left := splitmerge(s[:len(s)/2])
	right := splitmerge(s[len(s)/2:])
	return Merge(left, right)
}

func mergesort[T cmp.Ordered](s []T) {
	_ = slices.AppendSeq(s[:0], splitmerge(s).Seq())
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
		if !Equal(S(slices.Values(s)), S(slices.Values(check))) {
			t.Fatal(s)
		}
	})
}

func TestChunks(t *testing.T) {
	s := slices.Collect(Chunks(Of(1, 2, 3, 4, 5), 2).Map(slices.Clone).Seq())
	if !slices.EqualFunc(s, [][]int{{1, 2}, {3, 4}, {5}}, slices.Equal) {
		t.Fatal(s)
	}
}

func TestChunksFunc(t *testing.T) {
	s := slices.Collect(ChunksFunc(Of(0, 0, 0, 0, 1, 0, 1, 1, 0, 1),
		func(v int) bool { return v%2 == 0 }).Map(slices.Clone).Seq())
	if !slices.EqualFunc(s, [][]int{{0, 0, 0, 0}, {1}, {0}, {1, 1}, {0}, {1}}, slices.Equal) {
		t.Fatal(s)
	}
}

func TestSplit2(t *testing.T) {
	s1, s2 := FromPair(Of(Pair[int32, int64]{1, 2}, Pair[int32, int64]{3, 4}, Pair[int32, int64]{5, 6})).Split2().CollectSplit()
	if !slices.Equal(s1, []int32{1, 3, 5}) {
		t.Fatal(s1)
	}
	if !slices.Equal(s2, []int64{2, 4, 6}) {
		t.Fatal(s2)
	}
}

func TestCache(t *testing.T) {
	var i int
	f := S(func(yield func(int) bool) {
		yield(i)
		i++
	})
	seq := f.Cache()
	if s := slices.Collect(seq.Seq()); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
	if s := slices.Collect(seq.Seq()); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
}

func TestEnumerate(t *testing.T) {
	s := slices.Collect(ToPair(Generate(0, 2).Limit(3).Enumerate()).Seq())
	if !slices.Equal(s, []Pair[int, int]{{0, 0}, {1, 2}, {2, 4}}) {
		t.Fatal(s)
	}
}

func TestOr(t *testing.T) {
	s := slices.Collect(Or(Of[int](), nil, Of(1, 2, 3), Of(4, 5, 6)).Seq())
	if !slices.Equal(s, []int{1, 2, 3}) {
		t.Fatal(s)
	}

	s = slices.Collect(Or(Of[int](), Of(1, 2, 3), Of(4, 5, 6)).Seq())
	if !slices.Equal(s, []int{1, 2, 3}) {
		t.Fatal(s)
	}
}

func TestDedup(t *testing.T) {
	s := slices.Collect(Dedup(Of(1, 2, 3, 1, 2, 3, 4, 5, 3, 3, 3, 1, 2, 10)).Seq())
	if !slices.Equal(s, []int{1, 2, 3, 4, 5, 10}) {
		t.Fatal(s)
	}
}

func TestUniq(t *testing.T) {
	s := slices.Collect(Uniq(Of(1, 2, 3, 1, 2, 3, 4, 5, 3, 3, 3, 1, 2, 10)).Seq())
	if !slices.Equal(s, []int{1, 2, 3, 1, 2, 3, 4, 5, 3, 1, 2, 10}) {
		t.Fatal(s)
	}
}

func TestSorted(t *testing.T) {
	s := slices.Collect(Sorted(Of(3, 2, 5, 1, 7, 7, 8, 2)).Seq())
	if !slices.Equal(s, []int{1, 2, 2, 3, 5, 7, 7, 8}) {
		t.Fatal(s)
	}
}

func TestSortedFunc(t *testing.T) {
	compare := func(v1, v2 int) int { return v2 - v1 }
	s := slices.Collect(Of(3, 2, 5, 1, 7, 7, 8, 2).SortedFunc(compare).Seq())
	if !slices.Equal(s, []int{8, 7, 7, 5, 3, 2, 2, 1}) {
		t.Fatal(s)
	}
}
