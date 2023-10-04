xiter
=====

[![Go Reference](https://pkg.go.dev/badge/deedles.dev/xiter.svg)](https://pkg.go.dev/deedles.dev/xiter)

xiter is a very simple implementation of iterator support functions now supported by GOEXPERIMENT=range in the 1.22dev compiler. It is primarily intended to make it easier to play around with that experiment and not for actual usage.

Although the module's functionality is compatible with the new experiment, all of its features should work just fine with a plain Go toolchain. If you would like to try it with the development toolchain, you can install that fairly easily by running

```bash
$ go install golang.org/dl/gotip@latest
$ gotip download
$ GOEXPERIMENT=range gotip ... 
```

This will set up the `gotip` command to run the development version of go, allowing for `gotip build`, `gotip run`, etc.
