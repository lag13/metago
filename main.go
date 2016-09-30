// Playing around with metaprogramming
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"text/template"
)

func main() {
	fset := token.NewFileSet()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("could not get current file name")
	}
	file, err := parser.ParseFile(fset, thisFile, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	ast.Walk(&fnNameVisitor{}, file)
	ast.Walk(&importVisitor{}, file)
	ast.Walk(&fnBodyVisitor{}, file)
	printer.Fprint(os.Stdout, fset, file)
	visualizeCall(`thisIsNotExported()`)

	generateRunRmGoFile("./pleaseWork.go", testmainTmpl)
}

func runGoProgram(file string) ([]byte, error) {
	return exec.Command("go", "run", file).Output()
}

func writeGoProgram(file string, tmpl *template.Template) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := testmainTmpl.Execute(f, nil); err != nil {
		return err
	}
	return nil
}

func generateRunRmGoFile(file string, tmpl *template.Template) error {
	if err := writeGoProgram(file, tmpl); err != nil {
		return err
	}
	output, err := runGoProgram(file)
	if err != nil {
		return err
	}
	fmt.Print(string(output))
	return os.Remove(file)
}

type fnNameVisitor struct{}

func (v *fnNameVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.FuncDecl:
		t.Name = ast.NewIdent(strings.Title(t.Name.Name))
	}
	return v
}

type importVisitor struct{}

func (v *importVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.GenDecl:
		if t.Tok == token.IMPORT {
			newSpecs := make([]ast.Spec, len(t.Specs)+1)
			copy(newSpecs, t.Specs)
			newPackage := &ast.BasicLit{token.NoPos, token.STRING, `"woweee, this cannot be a real import can it?"`}
			newSpecs[len(t.Specs)] = &ast.ImportSpec{nil, nil, newPackage, nil, token.NoPos}
			t.Specs = newSpecs
			return nil
		}
	}
	return v
}

type fnBodyVisitor struct{}

func (v *fnBodyVisitor) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.FuncDecl:
		newBodyList := make([]ast.Stmt, 1, len(t.Body.List)+1)
		fnCall := &ast.BasicLit{token.NoPos, token.STRING, "fmt.Println"}
		arg := &ast.BasicLit{token.NoPos, token.STRING, fmt.Sprintf("\"%s\"", t.Name.Name)}
		newStmt := &ast.ExprStmt{&ast.CallExpr{fnCall, token.NoPos, []ast.Expr{arg}, token.NoPos, token.NoPos}}
		newBodyList[0] = newStmt
		newBodyList = append(newBodyList, t.Body.List...)
		t.Body.List = newBodyList
	}
	return v
}

func thisIsNotExported() {
	fmt.Println("hello")
}

// I think to get something like this to work we need to do something similar
// to go test (/usr/local/Cellar/go/1.6/libexec/src/cmd/go/test.go) where we (I
// think) write a program file and then execute that file. I have no idea how
// that would get done though. The simple case I want right now is just
// fibonacci. So I just need a file with a main package and a main function
// which calls the fibonacci function (but altered so the first line prints out
// the arguments).

// 1. Run a different go program from the current one. CHECK
// 2. Generate a go program and then call that go program. CHECK
// 3. Be able to modify a function by adding a fmt.Print() statement at the top
// of the function which prints out the function's name and its arguments.
// 4. Write a function which takes a function name and prints out that modified
// function name using 3.
// 5. Put 1-4 together in one program and I think we'll have everything we need.
func visualizeCall(s string) {
	expr, err := parser.ParseExpr(s)
	if err != nil {
		log.Fatal(err)
	}
	switch t := expr.(type) {
	case *ast.CallExpr:
		fnIdent := t.Fun.(*ast.Ident)
		fmt.Println(fnIdent.Name)
		fnRef := reflect.ValueOf(fnIdent.Name)
		fmt.Println(fnRef.Kind())
		fmt.Println(reflect.Func)
		// fmt.Printf("%T\n", t.Fun)
		// t.Fun is of type *ast.Ident
		// visualizeCall(`thisIsNotExported()`)
		// t.Fun is of type *ast.SelectorExpr
		//visualizeCall(`fmt.Println("hello there buddy")`)
	default:
		log.Fatal("could not parse string into a function call")
	}
}

var testmainTmpl = template.Must(template.New("main").Parse(`
package main

import (
	"fmt"
)

func main() {
	fmt.Println("hello world! I am a generated file!!!")
}

`))
