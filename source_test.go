package xiter

import (
	"cmp"
	"context"
	"slices"
	"testing"
)

func TestRunes(t *testing.T) {
	s := Collect(Runes("これはテストです。"))
	if [9]rune(s) != [9]rune([]rune("これはテストです。")) {
		t.Fatal(s)
	}
}

func TestMapEntries(t *testing.T) {
	s := Collect(MapEntries(map[string]string{"this": "is", "a": "test"}))
	slices.SortFunc(s, func(e1, e2 Pair[string, string]) int { return cmp.Compare(e1.V1, e2.V2) })
	if [2]Pair[string, string](s) != [...]Pair[string, string]{{"a", "test"}, {"this", "is"}} {
		t.Fatal(s)
	}
}

func TestRecvContext(t *testing.T) {
	c := make(chan int, 3)
	c <- 3
	c <- 2
	c <- 5
	close(c)

	s := Collect(RecvContext(context.Background(), c))
	if !slices.Equal(s, []int{3, 2, 5}) {
		t.Fatal(s)
	}
}
