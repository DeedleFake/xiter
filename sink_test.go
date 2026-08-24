package xiter

import (
	"context"
	"slices"
	"testing"
)

func TestFind(t *testing.T) {
	s, _ := Windows(Generate(0, 1), 3).Find(func(win []int) bool { return Sum(Of(win...)) >= 100 })
	if [3]int(s) != [...]int{33, 34, 35} {
		t.Fatal(s)
	}
}

func TestContains(t *testing.T) {
	c := Contains(Of(1, 2, 3), 2)
	if !c {
		t.Fatal(c)
	}
}

func TestSum(t *testing.T) {
	s := Sum(Of("a", " ", "test"))
	if s != "a test" {
		t.Fatal(s)
	}
}

func TestProduct(t *testing.T) {
	p := Product(Of(3, 2, -5))
	if p != -30 {
		t.Fatal(p)
	}
}

func TestPartition(t *testing.T) {
	s1, s2 := Of(1, 2, 3, 4, 5).Partition(func(v int) bool { return v%2 == 0 })
	if !Equal(Of(s1...), Of(2, 4)) {
		t.Fatal(s1)
	}
	if !Equal(Of(s2...), Of(1, 3, 5)) {
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

func TestAny(t *testing.T) {
	r := Of(2, 4, 6, 7).Any(func(v int) bool { return v%2 != 0 })
	if !r {
		t.Fatal(r)
	}
}

func TestAll(t *testing.T) {
	even := func(v int) bool { return v%2 == 0 }

	r := Of(2, 4, 6).All(even)
	if !r {
		t.Fatal(r)
	}

	r = Of(2, 4, 6, 7).All(even)
	if r {
		t.Fatal(r)
	}

	r = Of(1, 3, 5).All(even)
	if r {
		t.Fatal(r)
	}

	r = Of[int]().All(even)
	if !r {
		t.Fatal(r)
	}
}

func TestSendContext(t *testing.T) {
	c := make(chan int, 3)
	Of(3, 2, 5).SendContext(context.Background(), c)
	s := []int{<-c, <-c, <-c}
	select {
	case v := <-c:
		t.Fatal(v)
	default:
	}
	if !slices.Equal(s, []int{3, 2, 5}) {
		t.Fatal(s)
	}
}

func FuzzSendRecvContext(f *testing.F) {
	f.Add([]byte("The quick brown fox jumps over the lazy dog."))

	f.Fuzz(func(t *testing.T, data []byte) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()

		c := make(chan byte, len(data))
		S(slices.Values(data)).SendContext(ctx, c)
		close(c)
		s := RecvContext(ctx, c).Collect()
		if !slices.Equal(data, s) {
			t.Fatal(s)
		}
	})
}

func TestDrain(t *testing.T) {
	v, ok := Of(3, 2, 5).Drain()
	if !ok || v != 5 {
		t.Fatalf("%v, %v", v, ok)
	}

	v, ok = Of[int]().Drain()
	if ok || v != 0 {
		t.Fatalf("%v, %v", v, ok)
	}
}

func TestStringJoin(t *testing.T) {
	s := StringJoin(Of("this", "is", "a", "test"), " ")
	if s != "this is a test" {
		t.Fatalf("%q", s)
	}
}
