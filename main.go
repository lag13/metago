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
	// printer.Fprint(os.Stdout, token.NewFileSet(), fn.fnDecl)
	b := &bytes.Buffer{}
	printer.Fprint(b, token.NewFileSet(), fn.fnDecl)
	if err := visualizeRecFnCalls("fib", "4", b.String()); err != nil {
		log.Fatal(err)
	}
	fn = visualizeFn{fnName: "fact"}
	ast.Walk(&fn, file)
	b = &bytes.Buffer{}
	printer.Fprint(b, token.NewFileSet(), fn.fnDecl)
	if err := visualizeRecFnCalls("fact", "4", b.String()); err != nil {
		log.Fatal(err)
	}
}

const prg = `
package main

import (
	"fmt"
	"strings"
)

func main() {
	%s(%s, 0)
}

%s
`

func visualizeRecFnCalls(fnName string, args string, fnBody string) error {
	return generateRunRmGoFile("./pleaseWork.go", fmt.Sprintf(prg, fnName, args, fnBody))
}

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
	defer os.Remove(file)
	output, err := runGoProgram(file)
	if err != nil {
		return fmt.Errorf("%s: %v", output, err)
	}
	fmt.Print(string(output))
	return nil
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
			// Modify the recursive calls to add "depth+1" as the second
			// parameter to the recursive function call.
			addDepthParamToRecCalls(t.Name.Name, t.Body)
			// Modify the body to add a print statement
			printParams := []string{}
			printArgs := []string{`strings.Repeat("-", depth)`}
			for _, arg := range t.Type.Params.List {
				printParams = append(printParams, "%v")
				printArgs = append(printArgs, arg.Names[0].Name)
			}
			newBodyList := []ast.Stmt{}
			// I think if we were "properly" constructing the AST to add a
			// fmt.Printf statement we would need these types:
			// *ast.ExprStmt: &{0x82032ca80}
			// *ast.CallExpr: &{0x82032ee20 7545 [0x82032ee40 n] 0 7559}
			// *ast.SelectorExpr: &{fmt Print}
			// *ast.BasicLit: &{7546 STRING "hello %d"}
			// *ast.Ident: n
			printCall := &ast.BasicLit{token.NoPos, token.STRING, "fmt.Printf"}
			printCallArg := &ast.BasicLit{token.NoPos, token.STRING, fmt.Sprintf("\"%%s%s(%s)\\n\", %v", t.Name.Name, strings.Join(printParams, ", "), strings.Join(printArgs, ", "))}
			newPrintStmt := &ast.ExprStmt{&ast.CallExpr{printCall, token.NoPos, []ast.Expr{printCallArg}, token.NoPos, token.NoPos}}
			newBodyList = append(newBodyList, newPrintStmt)
			newBodyList = append(newBodyList, t.Body.List...)
			t.Body.List = newBodyList
			// Modify the function signature to accept a "depth" argument as the first parameter
			depthFieldName := &ast.Ident{token.NoPos, "depth", nil}
			depthFieldType := &ast.Ident{token.NoPos, "int", nil}
			depthField := &ast.Field{nil, []*ast.Ident{depthFieldName}, depthFieldType, nil, nil}
			t.Type.Params.List = append(t.Type.Params.List, depthField)
			// Store the modified function
			v.fnDecl = t
		}
	}
	return v
}

func addDepthParamToRecCalls(fnName string, node ast.Node) {
	// fmt.Printf("%T: %v\n", node, node)
	switch t := node.(type) {
	case *ast.BlockStmt:
		for _, stmt := range t.List {
			addDepthParamToRecCalls(fnName, stmt)
		}
	case *ast.ExprStmt:
		addDepthParamToRecCalls(fnName, t.X)
	case *ast.IfStmt:
		addDepthParamToRecCalls(fnName, t.Body)
	case *ast.ReturnStmt:
		for _, result := range t.Results {
			addDepthParamToRecCalls(fnName, result)
		}
	case *ast.BinaryExpr:
		addDepthParamToRecCalls(fnName, t.X)
		addDepthParamToRecCalls(fnName, t.Y)
	case *ast.CallExpr:
		addDepthParamToRecCalls(fnName, t.Fun)
		for _, arg := range t.Args {
			addDepthParamToRecCalls(fnName, arg)
		}
		name, ok := t.Fun.(*ast.Ident)
		if !ok || name.Name != fnName {
			return
		}
		depthIdent := &ast.Ident{token.NoPos, "depth", nil}
		oneLit := &ast.BasicLit{token.NoPos, token.INT, "1"}
		t.Args = append(t.Args, &ast.BinaryExpr{depthIdent, token.NoPos, token.ADD, oneLit})
	}
}

func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}

func fact(n int) int {
	if n == 0 {
		return 1
	}
	return n * fact(n-1)
}
