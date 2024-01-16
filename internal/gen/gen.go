package main

import (
	"bytes"
	_ "embed"
	"go/format"
	"go/types"
	"os"
	"regexp"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

var (
	namePattern = regexp.MustCompile(`^_[A-Z]`)

	//go:embed output.go.tmpl
	outputTemplate string
	tmpl           = template.Must(template.New("output").Funcs(funcMap).Parse(outputTemplate))
	funcMap        = map[string]any{"convertName": convertName}
)

func convertName(name string) string {
	return strings.TrimPrefix(name, "_")
}

func load() []*types.Func {
	config := packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo}
	pkgs, err := packages.Load(&config, "deedles.dev/xiter")
	if err != nil {
		panic(err)
	}
	pkg := pkgs[0]

	var funcs []*types.Func
	for _, def := range pkg.TypesInfo.Defs {
		f, ok := def.(*types.Func)
		if !ok {
			continue
		}
		if !namePattern.MatchString(f.Name()) {
			continue
		}

		funcs = append(funcs, f)
	}

	return funcs
}

func write(name string, funcs []*types.Func, rangefunc bool) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]any{"RangeFunc": rangefunc, "Funcs": funcs})
	if err != nil {
		panic(err)
	}

	b := buf.Bytes()
	b, err = format.Source(b)
	if err != nil {
		panic(err)
	}

	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.Write(b)
	if err != nil {
		panic(err)
	}
}

func main() {
	funcs := load()
	write("gen_rangefunc.go", funcs, true)
	write("gen_norangefunc.go", funcs, false)
}
