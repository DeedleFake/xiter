package xiter

import (
	"fmt"
	"testing"
)

func TestCoroutine(t *testing.T) {
	yield, stop := Coroutine(func(first int, yield func(int) (int, bool)) int {
		if first != 0 {
			t.Fatal(first)
		}

		prev := first
		for {
			v, ok := yield(prev + 1)
			if !ok {
				if prev != 10 {
					t.Fatal(v)
				}
				return -1
			}
			if v != prev+2 {
				t.Fatal(v)
			}
			prev = v
		}
	})

	prev, ok := yield(0)
	if !ok {
		t.Fatal(prev)
	}
	for prev < 10 {
		v, ok := yield(prev + 1)
		if !ok {
			t.Fatal(v)
		}
		if v != prev+2 {
			t.Fatal(v)
		}
		prev = v
	}

	r := stop()
	if r != -1 {
		t.Fatal(r)
	}
}

func ExampleCoroutine() {
	next, stop := Coroutine(func(val int, next func(int) (int, bool)) int {
		for {
			v, ok := next(val * 2)
			if !ok {
				return v
			}
			val = v
		}
	})
	defer stop()

	var val int
	for range 10 {
		v, ok := next(val + 1)
		if !ok {
			break
		}
		val = v
	}
	fmt.Printf("result: %v\n", stop())
	// Output: result: 1023
}

func TestPush(t *testing.T) {
	next, stop := Push(Sum[int])
	defer stop()

	for n := range 10 {
		if !next(n) {
			t.Fatal(n)
		}
	}

	r := stop()
	if r != 45 {
		t.Fatal(r)
	}
}
