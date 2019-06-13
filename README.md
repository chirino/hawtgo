# A Hawt Go Library

Install with:

    go get github.com/chirino/hawtgo

## Package: github.com/chirino/hawtgo/sh 

The `sh` package makes executing processes from go almost as easy as using a shell.

For example see the following:

	 sh.New().Line(`cp "${HOME}/my file.txt" /tmp/target.txt`).MustExec()

Notice that you don't need to break up the command arguments into an array and 
that shell style argument quoting is supported. 

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