// Playing around with the go syntax parsing packages.
// http://www.lshift.net/blog/2011/04/30/using-the-syntax-tree-in-go/
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"log"
	"os"
	"runtime"
	"strings"
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
