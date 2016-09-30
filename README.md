# metago

Playing around with metaprogramming in go.

- https://www.youtube.com/watch?v=aI_XdjQwTCU
- http://www.lshift.net/blog/2011/04/30/using-the-syntax-tree-in-go/
- http://stackoverflow.com/questions/23923383/evaluate-formula-in-go
- See second comment, go has an rpc server thing??? That looks cool:
  http://stackoverflow.com/questions/37122401/execute-another-go-program-from-within-a-golang-program
- https://github.com/xtaci/goeval

## Recursive Visualization

The initial inspiration for creating this repository was that I wanted to
write a program to "visualize" the recursive calls that get made for an
arbitrary function. I just thought it would be a fun thing to do! For example,
if we have this fibonacci function:

```go
func fib(n int) int {
	if n < 2 {
		return n
	}
	return fib(n-1) + fib(n-2)
}
```

Then doing something like: `visualizeRecFnCalls("fib", "4")` will print out:

```
fib(4)
-fib(3)
--fib(2)
---fib(1)
---fib(0)
--fib(1)
-fib(2)
--fib(1)
--fib(0)
```

I suppose you could generalize this and visualize all function calls that get
made or perhaps a subset of functions. Maybe I'll do that someday but this is
fine for now. I wouldn't be surprised if someone has done that actually, I
should try to find it.

### TODO:

I finished the super MVP of the above description but there are a lot of other
things I could do related to it:

1. Look into how `go test` constructs a main package and executes it. After
   looking at it for a while I'm getting the feeling that they also do the
   equivalent of just calling `go run` like I'm doing but I'm not sure.
2. Generate a random file name that we will call `go run` on so we don't run
   into any file conflict issues.
3. Dynamically generate the imports of the generated file. Right now they are
   just statically there so if the function I am visualizing uses packages
   other than "fmt" and "strings" then it won't work.
4. Dynamically construct the "depth" identifier name just in case the
   recursive function uses that name in its definition.
5. If something goes wrong with the `go run` I don't think I'm getting stderr
   which makes it a bit harder to debug if something goes wrong. How do I get
   stderr when executing a program?
6. Put the "visualize" function into a library rather than just having it in
   the main package.
7. When searching through the `go test` code I saw something about examples:
   /usr/local/Cellar/go/1.6/libexec/src/cmd/go/test.go 1361. And after
   noticing that I did notice more that the standard library documentation has
   these "Examples" sections. Learn more about these.
