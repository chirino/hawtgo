# My Hawt set of Go Packages

Install with:

    go get github.com/chirino/hawtgo

## Pacakge: sh [![GoDoc](https://godoc.org/github.com/chirino/hawtgo/sh?status.svg)](https://godoc.org/github.com/chirino/hawtgo/sh)

The sh package makes executing processes from go almost as easy as using a shell.

See the following example that takes care of splitting up the command arguments
and doing variable replacement:

	 sh.New().Line(`cp "${HOME}/my file.txt" /tmp/target.txt`).MustExec()

sh will can run command string like the ones you would enter into your shell.

The `sh.New()` function give you back a new immutable builder object. You can safely
reuse it for multiple command invocations.

    var mysh = sh.New().
        CommandLog(os.Stdout).
        CommandLogPrefix("> ").
        Env(map[string]string{
            "CGO_ENABLED":     "0",
            "GOOS":            "linux",
            "GOARCH":          "amd64",
        }).
        Dir("./target")
        
    func BuildExecutable() {
        mysh.Line(`go build -o "my app" github.com/chirino/cmd/myapp`).MustExec()
        mysh.Line(`go build -o otherapp github.com/chirino/cmd/otherapp`).MustExec()
    }