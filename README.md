xiter
=====

[![Go Reference](https://pkg.go.dev/badge/deedles.dev/xiter.svg)](https://pkg.go.dev/deedles.dev/xiter)

xiter is a very simple implementation of iterator utility functions for Go's `rangefunc` GOEXPERIMENT introduced in Go 1.22. It is primarily intended to make it easier to play around with that experiment and not for actual usage. Although the module's functionality is compatible with the new experiment, all of its features should work just fine with a plain Go toolchain and should even work with any older version of Go that supports generics (1.18+).

Note that due to the lack of generic type aliases, this package's `Seq` type and the standard library's `iter.Seq` type need to be manually converted between. This should hopefully be resolved in Go 1.23. See https://github.com/golang/go/issues/46477.
