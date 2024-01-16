package main

import (
	"bytes"
	"cmp"
	_ "embed"
	"fmt"
	"go/format"
	"go/types"
	"log/slog"
	"os"
	"regexp"
	"slices"
	"strings"
	"text/template"

	"golang.org/x/tools/go/packages"
)

var (
	namePattern = regexp.MustCompile(`^_[A-Z]`)

	//go:embed output.go.tmpl
	outputTemplate string
	tmpl           = template.Must(template.New("output").Funcs(funcMap).Parse(outputTemplate))
	funcMap        = map[string]any{
		"typeParamSlice":  listToSlice[*types.TypeParam],
		"tupleSlice":      listToSlice[*types.Var],
		"convertFuncName": convertFuncName,
		"convertType":     convertType,
		"convertArg":      convertArg,
		"convertReturn":   convertReturn,
	}
)

func convertFuncName(name string) string {
	return strings.TrimPrefix(name, "_")
}

type List[T any] interface {
	At(int) T
	Len() int
}

func listToSlice[T any](list List[T]) []T {
	length := list.Len()
	s := make([]T, 0, length)
	for i := 0; i < length; i++ {
		s = append(s, list.At(i))
	}
	return s
}

func convertType(rangefunc bool, t types.Type) string {
	switch t := t.(type) {
	case *types.Named:
		var pkg string
		if t.Obj().Pkg() != nil && t.Obj().Pkg().Path() != "deedles.dev/xiter" {
			pkg = t.Obj().Pkg().Name() + "."
		}

		name := convertTypeName(rangefunc, t.Obj().Name())

		if t.TypeArgs().Len() == 0 {
			return pkg + name
		}

		typeArgs := make([]string, 0, t.TypeArgs().Len())
		for _, arg := range listToSlice(t.TypeArgs()) {
			typeArgs = append(typeArgs, convertType(rangefunc, arg))
		}
		return fmt.Sprintf("%v%v[%v]", pkg, name, strings.Join(typeArgs, ","))

	case *types.Slice:
		return fmt.Sprintf("[]%v", convertType(rangefunc, t.Elem()))

	case *types.Interface, *types.Basic, *types.TypeParam, *types.Signature, *types.Chan:
		return t.String()

	default:
		return fmt.Sprintf("\"%T\"", t)
	}
}

func convertTypeName(rangefunc bool, name string) string {
	cut, ok := strings.CutPrefix(name, "_")
	if !ok {
		return name
	}

	if !rangefunc {
		return cut
	}

	return "iter." + cut
}

func convertArgType(rangefunc bool, t types.Type) (from string, to string, ok bool) {
	switch t := t.(type) {
	case *types.Named:
		to := t.Obj().Name()

		if t.Obj().Pkg() == nil || t.Obj().Pkg().Path() != "deedles.dev/xiter" {
			return t.String(), t.String(), false
		}

		if t.TypeArgs().Len() != 0 {
			typeArgs := make([]string, 0, t.TypeArgs().Len())
			for _, arg := range listToSlice(t.TypeArgs()) {
				_, to, _ := convertArgType(rangefunc, arg)
				typeArgs = append(typeArgs, to)
			}
			to = fmt.Sprintf("%v[%v]", to, strings.Join(typeArgs, ","))
		}

		from, ok := strings.CutPrefix(to, "_")
		return from, to, ok

	default:
		return t.String(), t.String(), false
	}
}

func convertArg(rangefunc bool, t types.Type, name string) string {
	switch t := t.(type) {
	case *types.Named:
		_, to, ok := convertArgType(rangefunc, t)
		if !ok {
			return name
		}
		return fmt.Sprintf("%v(%v)", to, name)

	case *types.Slice:
		from, to, ok := convertArgType(rangefunc, t.Elem())
		if !ok {
			return name
		}
		return fmt.Sprintf("xslices.Map(%v, func(v %v) %v { return %v(v) })", name, from, to, to)

	default:
		return name
	}
}

func convertReturn(rangefunc bool, t types.Type, name string) string {
	switch t := t.(type) {
	case *types.Named:
		if t.Obj().Pkg() == nil || t.Obj().Pkg().Path() != "deedles.dev/xiter" {
			return name
		}

		tname, ok := strings.CutPrefix(t.Obj().Name(), "_")
		if !ok {
			return name
		}
		if rangefunc {
			tname = "iter." + tname
		}

		if t.TypeArgs().Len() == 0 {
			return fmt.Sprintf("%v(%v)", tname, name)
		}

		typeArgs := make([]string, 0, t.TypeArgs().Len())
		for _, arg := range listToSlice(t.TypeArgs()) {
			typeArgs = append(typeArgs, convertType(rangefunc, arg))
		}
		return fmt.Sprintf("%v[%v](%v)", tname, strings.Join(typeArgs, ","), name)

	default:
		return name
	}
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
	slices.SortFunc(funcs, func(f1, f2 *types.Func) int { return cmp.Compare(f1.Name(), f2.Name()) })

	return funcs
}

func write(name string, funcs []*types.Func, rangefunc bool) {
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, map[string]any{"RangeFunc": rangefunc, "Funcs": funcs})
	if err != nil {
		panic(err)
	}

	b := buf.Bytes()
	formatted, err := format.Source(b)
	if err != nil {
		slog.Error("format", "file", name, "err", err)
		formatted = b
	}

	file, err := os.Create(name)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_, err = file.Write(formatted)
	if err != nil {
		panic(err)
	}
}

func main() {
	funcs := load()
	write("gen_rangefunc.go", funcs, true)
	write("gen_norangefunc.go", funcs, false)
}
