xiter
=====

[![Go Reference](https://pkg.go.dev/badge/deedles.dev/xiter.svg)](https://pkg.go.dev/deedles.dev/xiter)

xiter is a very simple implementation of iterator support functions compatible with CL 510541. It is primarily intended to make it easier to play around with that CL and not for actual usage.

Although the module's functionality is compatible with CL 510541, all of its features should work just fine with an unmodified Go toolchain. If you would like to try it with the CL, you can install it fairly easily by running

```bash
$ go install golang.org/dl/gotip@latest
$ gotip download 510541
```

This will set up the `gotip` command to run the modified toolchain, allowing for `gotip build`, `gotip run`, etc.
