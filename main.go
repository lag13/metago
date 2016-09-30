// Playing around with metaprogramming
package main

import (
	"bytes"
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
)

func main() {
	// I'm not really sure what this FileSet variable does
	fset := token.NewFileSet()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("could not get current file name")
	}
	file, err := parser.ParseFile(fset, thisFile, nil, 0)
	if err != nil {
		log.Fatal(err)
	}
	fn := visualizeFn{fnName: "fib"}
	ast.Walk(&fn, file)
	b := &bytes.Buffer{}
	printer.Fprint(b, token.NewFileSet(), fn.fnDecl)
	generateRunRmGoFile("./pleaseWork.go", fmt.Sprintf(prg, b.String()))
}

const prg = `
package main

import (
	"fmt"
)

func main() {
	fib(4)
}

%s
`

func runGoProgram(file string) ([]byte, error) {
	return exec.Command("go", "run", file).Output()
}

func writeGoProgram(file string, contents string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := f.Write([]byte(contents)); err != nil {
		return err
	}
	return nil
}

func generateRunRmGoFile(file string, contents string) error {
	if err := writeGoProgram(file, contents); err != nil {
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

type modifyFn string

func (v *modifyFn) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.FuncDecl:
		if t.Name.Name == string(*v) {
			args := t.Type.Params.List
			printParams := []string{}
			printArgs := []string{}
			for _, arg := range args {
				printParams = append(printParams, "%v")
				printArgs = append(printArgs, arg.Names[0].Name)
			}
			fmt.Println(args[0].Names[0])
			newBodyList := make([]ast.Stmt, 1, len(t.Body.List)+1)
			printCall := &ast.BasicLit{token.NoPos, token.STRING, "fmt.Printf"}
			printCallArg := &ast.BasicLit{token.NoPos, token.STRING, fmt.Sprintf("\"%s(%s)\\n\", %v", t.Name.Name, strings.Join(printParams, ", "), strings.Join(printArgs, ", "))}
			newPrintStmt := &ast.ExprStmt{&ast.CallExpr{printCall, token.NoPos, []ast.Expr{printCallArg}, token.NoPos, token.NoPos}}
			newBodyList[0] = newPrintStmt
			newBodyList = append(newBodyList, t.Body.List...)
			t.Body.List = newBodyList
		}
	}
	return v
}

type visualizeFn struct {
	fnName string
	fnDecl *ast.FuncDecl
}

func (v *visualizeFn) Visit(node ast.Node) (w ast.Visitor) {
	switch t := node.(type) {
	case *ast.FuncDecl:
		if t.Name.Name == v.fnName {
			args := t.Type.Params.List
			printParams := []string{}
			printArgs := []string{}
			for _, arg := range args {
				printParams = append(printParams, "%v")
				printArgs = append(printArgs, arg.Names[0].Name)
			}
			newBodyList := []ast.Stmt{}
			printCall := &ast.BasicLit{token.NoPos, token.STRING, "fmt.Printf"}
			printCallArg := &ast.BasicLit{token.NoPos, token.STRING, fmt.Sprintf("\"%s(%s)\\n\", %v", t.Name.Name, strings.Join(printParams, ", "), strings.Join(printArgs, ", "))}
			newPrintStmt := &ast.ExprStmt{&ast.CallExpr{printCall, token.NoPos, []ast.Expr{printCallArg}, token.NoPos, token.NoPos}}
			newBodyList = append(newBodyList, newPrintStmt)
			newBodyList = append(newBodyList, t.Body.List...)
			t.Body.List = newBodyList
			v.fnDecl = t
		}
	}
	return v
}

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}
