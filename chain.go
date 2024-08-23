package xiter

import (
	"context"
	"iter"
	"slices"
)

// Chain is a wrapper around a iter.Seq that provides what
// functionality of this package it's possible to provide as methods,
// allowing calls to them to be chained. As a general rule of thumb,
// methods that are available are ones that don't introduce new type
// parameters, though a few others are missing as well.
type Chain[T any] iter.Seq[T]

func (chain Chain[T]) Seq() iter.Seq[T] { return iter.Seq[T](chain) }

func (chain Chain[T]) All(f func(T) bool) bool { return All(chain.Seq(), f) }

func (chain Chain[T]) Any(f func(T) bool) bool { return Any(chain.Seq(), f) }

func (chain Chain[T]) Collect() []T { return slices.Collect[T](chain.Seq()) }

func (chain Chain[T]) CollectSize(len int) []T { return CollectSize[T](chain.Seq(), len) }

func (chain Chain[T]) Drain() { Drain[T](chain.Seq()) }

func (chain Chain[T]) Find(f func(T) bool) (r T, ok bool) { return Find(chain.Seq(), f) }

func (chain Chain[T]) Fold(reducer func(T, T) T) T { return Fold(chain.Seq(), reducer) }

func (chain Chain[T]) IsSortedFunc(compare func(T, T) int) bool {
	return IsSortedFunc(chain.Seq(), compare)
}

func (chain Chain[T]) Partition(f func(T) bool) (true, false []T) {
	return Partition(chain.Seq(), f)
}

func (chain Chain[T]) PartitionInto(f func(T) bool, true, false []T) ([]T, []T) {
	return PartitionInto(chain.Seq(), f, true, false)
}

func (chain Chain[T]) Pull() (iterator func() (T, bool), stop func()) {
	return iter.Pull(chain.Seq())
}

func (chain Chain[T]) SendContext(ctx context.Context, c chan<- T) {
	SendContext(chain.Seq(), ctx, c)
}

func (chain Chain[T]) Cache() Chain[T] { return Chain[T](Cache[T](chain.Seq())) }

// TODO: Is it possible to get around the instantiation cycle here?
//func (chain Chain[T]) Chunks(n int) Chain[[]T] { return Chain[[]T](Chunks[T](chain.Seq(), n)) }

func (chain Chain[T]) Concat(seqs ...iter.Seq[T]) Chain[T] {
	return func(yield func(T) bool) {
		cont := true
		wrap := func(v T) bool {
			cont = yield(v)
			return cont
		}

		chain(wrap)
		if !cont {
			return
		}

		for _, seq := range seqs {
			seq(wrap)
			if !cont {
				return
			}
		}
	}
}

func (chain Chain[T]) Filter(f func(T) bool) Chain[T] { return Chain[T](Filter(chain.Seq(), f)) }

func (chain Chain[T]) Limit(n int) Chain[T] { return Chain[T](Limit[T](chain.Seq(), n)) }

func (chain Chain[T]) MergeFunc(seq2 iter.Seq[T], compare func(T, T) int) Chain[T] {
	return Chain[T](MergeFunc(chain.Seq(), seq2, compare))
}

func (chain Chain[T]) Or(seqs ...iter.Seq[T]) Chain[T] {
	return func(yield func(T) bool) {
		cont := true
		wrap := func(v T) bool {
			cont = false
			return yield(v)
		}

		if chain != nil {
			chain(wrap)
			if !cont {
				return
			}
		}

		for _, seq := range seqs {
			if seq == nil {
				continue
			}

			seq(wrap)
			if !cont {
				return
			}
		}
	}
}

func (chain Chain[T]) Skip(n int) Chain[T] { return Chain[T](Skip[T](chain.Seq(), n)) }

// TODO: Is it possible to get around the instantiation cycle here?
//func (chain Chain[T]) Windows(n int) Chain[[]T] { return Chain[[]T](Windows[T](chain.Seq(), n)) }

func (chain Chain[T]) Enumerate() iter.Seq2[int, T] { return Enumerate[T](chain.Seq()) }

func (chain Chain[T]) Split(f func(T) bool) SplitSeq[T, T] { return Split(chain.Seq(), f) }
