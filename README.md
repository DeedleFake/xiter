xiter
=====

[![Go Reference](https://pkg.go.dev/badge/deedles.dev/xiter.svg)](https://pkg.go.dev/deedles.dev/xiter)

xiter is a very simple implementation of iterator utility functions for Go's iterators that were introduced in 1.23. Although the module's functionality is compatible with Go 1.23, it may also work with older versions, though this is not guaranteed.

Note that due to the lack of generic type aliases, this package's `Seq` type and the standard library's `iter.Seq` type need to be manually converted between on versions prior to Go 1.23. In 1.23, enabling the `aliastypeparams` experiment and setting `GODEBUG=gotypesalias=1` (Note the pluralization of "types".) will avoid the need for this. This should hopefully become unnecessary in Go 1.24. See https://github.com/golang/go/issues/46477.
