package xiter

import (
	"bytes"
	"cmp"
	"slices"
	"testing"
)

func TestMap(t *testing.T) {
	s := Of(1, 2, 3)
	n := s.Map(func(v int) float64 { return float64(v * 2) }).Collect()
	if [3]float64(n) != [...]float64{2, 4, 6} {
		t.Fatal(n)
	}
}

func TestFilter(t *testing.T) {
	s := Of(1, 2, 3)
	n := s.Filter(func(v int) bool { return v%2 != 0 }).Collect()
	if [2]int(n) != [...]int{1, 3} {
		t.Fatal(n)
	}
}

func TestFilterMap(t *testing.T) {
	s := OfSlice([]int{1, 2, 3})
	n := Collect(FilterMap(s, func(v int) float64 {
		if v%2 != 0 {
			return float64(v * 2)
		}

		return 0
	}))
	if [2]float64(n) != [...]float64{2, 6} {
		t.Fatal(n)
	}
}

func TestFilterMapPtr(t *testing.T) {
	type Foo struct{}
	type Bar struct{}

	s := Of[any](new(Foo), new(Bar), new(Foo), new(Bar))
	n := Collect(FilterMap(s, func(v any) *Foo {
		foo, _ := v.(*Foo)
		return foo
	}))
	if len(n) != 2 {
		t.Fatal(n)
	}
}

func TestFilterMap2(t *testing.T) {
	type Foo struct{}
	type Bar struct{}

	s := Of[any](new(Foo), new(Bar), new(Foo), new(Bar))
	n := Collect(FilterMap2(s, func(v any) (*Foo, bool) {
		foo, ok := v.(*Foo)
		return foo, ok
	}))
	if len(n) != 2 {
		t.Fatal(n)
	}
}

func TestSkip(t *testing.T) {
	s := Generate(0, 1).Limit(3).Skip(2).Collect()
	if !Equal(Of(s...), Of(2)) {
		t.Fatal(s)
	}
}

func TestLimit(t *testing.T) {
	s := Generate(0, 2).Limit(3).Collect()
	if [3]int(s) != [...]int{0, 2, 4} {
		t.Fatal(s)
	}
}

func TestConcat(t *testing.T) {
	s := Concat(Of(1, 2, 3), Of(3, 2, 5)).Collect()
	if [6]int(s) != [...]int{1, 2, 3, 3, 2, 5} {
		t.Fatal(s)
	}

	s = Concat(Of(1, 2, 3), Of(3, 2, 5)).Collect()
	if [6]int(s) != [...]int{1, 2, 3, 3, 2, 5} {
		t.Fatal(s)
	}
}

func TestZip(t *testing.T) {
	tests := []struct {
		name  string
		s1    []int
		s2    []int
		limit int
		want  []Zipped[int, int]
	}{
		{
			name: "equal",
			s1:   []int{1, 2, 3, 4, 5},
			s2:   []int{2, 3, 4, 5, 6},
			want: []Zipped[int, int]{
				{V1: 1, OK1: true, V2: 2, OK2: true},
				{V1: 2, OK1: true, V2: 3, OK2: true},
				{V1: 3, OK1: true, V2: 4, OK2: true},
				{V1: 4, OK1: true, V2: 5, OK2: true},
				{V1: 5, OK1: true, V2: 6, OK2: true},
			},
		},
		{
			name: "seq1 shorter",
			s1:   []int{1, 2, 3, 4},
			s2:   []int{2, 3, 4, 5, 10},
			want: []Zipped[int, int]{
				{V1: 1, OK1: true, V2: 2, OK2: true},
				{V1: 2, OK1: true, V2: 3, OK2: true},
				{V1: 3, OK1: true, V2: 4, OK2: true},
				{V1: 4, OK1: true, V2: 5, OK2: true},
				{V2: 10, OK2: true},
			},
		},
		{
			name: "seq2 shorter",
			s1:   []int{1, 2, 3, 4, 10},
			s2:   []int{2, 3, 4, 5},
			want: []Zipped[int, int]{
				{V1: 1, OK1: true, V2: 2, OK2: true},
				{V1: 2, OK1: true, V2: 3, OK2: true},
				{V1: 3, OK1: true, V2: 4, OK2: true},
				{V1: 4, OK1: true, V2: 5, OK2: true},
				{V1: 10, OK1: true},
			},
		},
		{
			name: "empty seq1",
			s2:   []int{1, 2, 3},
			want: []Zipped[int, int]{
				{V2: 1, OK2: true},
				{V2: 2, OK2: true},
				{V2: 3, OK2: true},
			},
		},
		{
			name: "empty seq2",
			s1:   []int{1, 2, 3},
			want: []Zipped[int, int]{
				{V1: 1, OK1: true},
				{V1: 2, OK1: true},
				{V1: 3, OK1: true},
			},
		},
		{
			name:  "stop with seq2 remaining",
			s1:    []int{1, 2},
			s2:    []int{10, 20, 30, 40},
			limit: 1,
			want: []Zipped[int, int]{
				{V1: 1, OK1: true, V2: 10, OK2: true},
			},
		},
		{
			name:  "stop leftover",
			s1:    []int{1},
			s2:    []int{10, 20, 30},
			limit: 2,
			want: []Zipped[int, int]{
				{V1: 1, OK1: true, V2: 10, OK2: true},
				{V2: 20, OK2: true},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seq := Zip(Of(tt.s1...), Of(tt.s2...))
			if tt.limit > 0 {
				seq = seq.Limit(tt.limit)
			}
			got := seq.Collect()
			if !slices.Equal(got, tt.want) {
				t.Fatal(got)
			}
		})
	}
}

func BenchmarkZip(b *testing.B) {
	benchmarkZip(b, Zip)
}

func benchmarkZip(b *testing.B, zip func(Seq[int], Seq[int]) Seq[Zipped[int, int]]) {
	shapes := []struct {
		name   string
		n1, n2 int
	}{
		{"n5_equal", 5, 5},
		{"n1000_equal", 1000, 1000},
		{"seq1_short", 10, 1000},
	}

	for _, shape := range shapes {
		b.Run(shape.name, func(b *testing.B) {
			s1 := make([]int, shape.n1)
			s2 := make([]int, shape.n2)
			for i := range s1 {
				s1[i] = i
			}
			for i := range s2 {
				s2[i] = i + 1
			}

			for b.Loop() {
				seq := zip(S(slices.Values(s1)), S(slices.Values(s2)))
				seq(func(v Zipped[int, int]) bool {
					return true
				})
			}
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
	r := Merge(s1, s2).Collect()
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
	s := Chunks(Of(1, 2, 3, 4, 5), 2).Map(slices.Clone).Collect()
	if !slices.EqualFunc(s, [][]int{{1, 2}, {3, 4}, {5}}, slices.Equal) {
		t.Fatal(s)
	}
}

func TestChunksFunc(t *testing.T) {
	s := ChunksFunc(Of(0, 0, 0, 0, 1, 0, 1, 1, 0, 1),
		func(v int) bool { return v%2 == 0 }).Map(slices.Clone).Collect()
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
	if s := seq.Collect(); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
	if s := seq.Collect(); !slices.Equal(s, []int{0}) {
		t.Fatal(s)
	}
}

func TestEnumerate(t *testing.T) {
	s := ToPair(Generate(0, 2).Limit(3).Enumerate()).Collect()
	if !slices.Equal(s, []Pair[int, int]{{0, 0}, {1, 2}, {2, 4}}) {
		t.Fatal(s)
	}
}

func TestOr(t *testing.T) {
	s := Or(Of[int](), nil, Of(1, 2, 3), Of(4, 5, 6)).Collect()
	if !slices.Equal(s, []int{1, 2, 3}) {
		t.Fatal(s)
	}

	s = Or(Of[int](), Of(1, 2, 3), Of(4, 5, 6)).Collect()
	if !slices.Equal(s, []int{1, 2, 3}) {
		t.Fatal(s)
	}
}

func TestDedup(t *testing.T) {
	s := Dedup(Of(1, 2, 3, 1, 2, 3, 4, 5, 3, 3, 3, 1, 2, 10)).Collect()
	if !slices.Equal(s, []int{1, 2, 3, 4, 5, 10}) {
		t.Fatal(s)
	}
}

func TestUniq(t *testing.T) {
	s := Uniq(Of(1, 2, 3, 1, 2, 3, 4, 5, 3, 3, 3, 1, 2, 10)).Collect()
	if !slices.Equal(s, []int{1, 2, 3, 1, 2, 3, 4, 5, 3, 1, 2, 10}) {
		t.Fatal(s)
	}
}

func TestSorted(t *testing.T) {
	s := Sorted(Of(3, 2, 5, 1, 7, 7, 8, 2)).Collect()
	if !slices.Equal(s, []int{1, 2, 2, 3, 5, 7, 7, 8}) {
		t.Fatal(s)
	}
}

func TestSortedFunc(t *testing.T) {
	compare := func(v1, v2 int) int { return v2 - v1 }
	s := Of(3, 2, 5, 1, 7, 7, 8, 2).SortedFunc(compare).Collect()
	if !slices.Equal(s, []int{8, 7, 7, 5, 3, 2, 2, 1}) {
		t.Fatal(s)
	}
}
