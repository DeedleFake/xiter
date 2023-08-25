// Package xiter provides iterator-related functionality compatible
// with, but not requiring, CL 510541.
package xiter

// Seq represents an iterator over a sequence of values. When called,
// the passed yield function is called for each successive value.
// Returning false from yield causes the iterator to stop, equivalent
// to a break statement. The return value of the Seq function itself
// is completely ignored, but present to be compatible with the CL
// 510541 prototype.
type Seq[T any] func(yield func(T) bool) bool
