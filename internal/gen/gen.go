package main

import (
	"fmt"
	"go/types"
	"regexp"

	"golang.org/x/tools/go/packages"
)

var (
	namePattern = regexp.MustCompile(`^_[A-Z]`)
)

func main() {
	config := packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(&config, "deedles.dev/xiter")
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]

	for _, def := range pkg.TypesInfo.Defs {
		f, ok := def.(*types.Func)
		if !ok {
			continue
		}
		if !namePattern.MatchString(f.Name()) {
			continue
		}

		fmt.Println(f)
	}
}
