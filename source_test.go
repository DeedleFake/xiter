package xiter

import (
	"cmp"
	"context"
	"maps"
	"slices"
	"strings"
	"testing"
	"unicode"
)

func TestBytes(t *testing.T) {
	s := Bytes("テスト").Collect()
	if !slices.Equal(s, []byte("テスト")) {
		t.Fatal(s)
	}
}

func TestRunes(t *testing.T) {
	s := Runes("これはテストです。").Collect()
	if [9]rune(s) != [9]rune([]rune("これはテストです。")) {
		t.Fatal(s)
	}
}

func TestMapEntries(t *testing.T) {
	s := ToPair(S2(maps.All(map[string]string{"this": "is", "a": "test"}))).Collect()
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

	s := RecvContext(context.Background(), c).Collect()
	if !slices.Equal(s, []int{3, 2, 5}) {
		t.Fatal(s)
	}
}

func TestStringSplit(t *testing.T) {
	s := StringSplit("this is a test", " ").Collect()
	if !slices.Equal(s, []string{"this", "is", "a", "test"}) {
		t.Fatal(s)
	}
}

func TestStringFields(t *testing.T) {
	s := StringFields("  this is a  test ").Collect()
	if !slices.Equal(s, []string{"this", "is", "a", "test"}) {
		t.Fatal(s)
	}

	s = StringFields("  this is a  test").Collect()
	if !slices.Equal(s, []string{"this", "is", "a", "test"}) {
		t.Fatal(s)
	}
}

func TestSliceChunkBy(t *testing.T) {
	s := SliceChunksFunc([]int{-1, -2, -3, 1, 2, 3, -1, -2, 3}, func(v int) int { return cmp.Compare(v, 0) }).Collect()
	if !slices.EqualFunc(s, [][]int{{-1, -2, -3}, {1, 2, 3}, {-1, -2}, {3}}, slices.Equal) {
		t.Fatal(s)
	}
}

func TestScanBytes(t *testing.T) {
	r := strings.NewReader("te st")
	b := ScanRunes(r)

	var buf []rune
	for c := range b {
		if unicode.IsSpace(c) {
			break
		}
		buf = append(buf, c)
	}
	if !slices.Equal(buf, []rune("te")) {
		t.Fatal(string(buf))
	}
	if r.Len() != 3 {
		t.Fatal(r.Len())
	}
	if c, err := r.ReadByte(); err != nil || c != ' ' {
		t.Fatalf("%q, %v", c, err)
	}
}
