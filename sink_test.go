package xiter

import "testing"

func TestFind(t *testing.T) {
	s, _ := Find(Windows(Generate(
		0, 1),
		3),
		func(win []int) bool { return Sum(Slice(win)) >= 100 })
	if [3]int(s) != [...]int{33, 34, 35} {
		t.Fatal(s)
	}
}

func TestSum(t *testing.T) {
	s := Sum(Slice([]string{"a", " ", "test"}))
	if s != "a test" {
		t.Fatal(s)
	}
}

func TestPartition(t *testing.T) {
	s1, s2 := Partition(Of(1, 2, 3, 4, 5), func(v int) bool { return v%2 == 0 })
	if !Equal(Slice(s1), Of(2, 4)) {
		t.Fatal(s1)
	}
	if !Equal(Slice(s2), Of(1, 3, 5)) {
		t.Fatal(s2)
	}
}

func TestExtent(t *testing.T) {
	s := Of(3, 2, 5, 1, 6, -2, 10)
	min := Min(s)
	max := Max(s)
	if min != -2 {
		t.Fatal(min)
	}
	if max != 10 {
		t.Fatal(max)
	}
}
