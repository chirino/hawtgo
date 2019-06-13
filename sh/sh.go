/*
 * Copyright (C) 2018 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

//
// Package sh helps you to more easily execute processes.
package sh

import (
    "fmt"
    "github.com/chirino/hawtgo/sh/line"
    magesh "github.com/magefile/mage/sh"
    "io"
    "os"
    "os/exec"
    "regexp"
    "strings"
)

var needsQuote = regexp.MustCompile(`'|"| |\t|\r|\n`)

/////////////////////////////////////////////////////////////////////////
//
// Expander related bits..
//
/////////////////////////////////////////////////////////////////////////

// An Expander is used to expand/resolve a variable name to a value
type Expander interface {
    // Expand retrieves the value of the variable named
    // by the key. If the variable is found the
    // value (which may be empty) is returned and the boolean is true.
    // Otherwise the returned value will be empty and the boolean will
    // be false.
    Expand(key string) (value string, ok bool)
}

// Expand replaces ${var} or $var in the string based on the Expander.
func Expand(value string, expander Expander) string {
    return os.Expand(value, func(v string) string {
        if v, ok := expander.Expand(v); ok {
            return v
        }
        return ""
    })
}

// ExpandNotFound returns an Expander that never finds the value
// being expanded.
func ExpandNotFound() Expander {
    return notFound(1)
}
type notFound byte
func (notFound) Expand(key string) (string, bool) {
    return "", false
}

// ExpandDisabled returns an Expander that evaluates to the same string
// that describes the expansion.
func ExpandDisabled() Expander {
    return expandDisabled(2)
}
type expandDisabled byte
func (expandDisabled) Expand(key string) (string, bool) {
    return "${" + key + "}", true
}

// ExpandEnv returns an Expander that expands values from the
// operating system environment.
func ExpandEnv() Expander {
return expandEnv(3)
}
type expandEnv byte
func (expandEnv) Expand(key string) (string, bool) {
    return os.LookupEnv(key)
}

// ExpandPanic returns an Expander that panics when used.
func ExpandPanic() Expander {
    return expandPanic(4)
}
type expandPanic byte
func (expandPanic) Expand(key string) (string, bool) {
    panic(fmt.Errorf("can not find value to expand '${%s}'", key))
}

// ExpandMap returns an Expander that values found found in the map.
func ExpandMap(m map[string]string) Expander {
    return expandMap(m)
}
type expandMap map[string]string
func (m expandMap) Expand(key string) (string, bool) {
    v, ok := m[key]
    return v, ok
}

// Expanders creates an Expander that expands using
// the provided list of Expanders in order.
//
// You can use this to customize how key not found scenarios are handled.
// If you want to panic if the key is not found in the OS Env you could
// build that expander like:
//
// exp := ChainExpanders(ExpandEnv(), ExpandPanic())
//
func ChainExpanders(v ...Expander) Expander {
    return expanders(v)
}
type expanders []Expander
func (next expanders) Expand(key string) (string, bool) {
    for _, f := range next {
        if v, ok := f.Expand(key); ok {
            return v, ok
        }
    }
    return "", false
}

/////////////////////////////////////////////////////////////////////////
//
// Sh struct  bits
//
/////////////////////////////////////////////////////////////////////////

// Sh contains all the settings needed to execute process. It is
// guarded by an immutable builder access model.
type Sh struct {
    args                []line.Arg
    expanders           Expander
    env                 map[string]string
    dir                 string
    commandLog          io.Writer
    commandLogPrefix    string
}

// New returns a new sh.Sh
func New() *Sh {
    return &Sh{expanders: ExpandEnv()}
}

// Line returns a new sh.Sh with the command specified as a single command.  The command line is
// parsed into command line arguments.  You can use single and double quotes like you do in bash
// to group command line arguments.  Single quoted strings will have variable expansion disabled.
func (this *Sh) Line(commandLine string) *Sh {
    var sh = *this;
    sh.args = line.Parse(commandLine)
    return &sh
}

// Line returns a new sh.Sh with the specified command line arguments.
func (this *Sh) LineArgs(commandLIne ...string) *Sh {
    var sh = *this;
    sh.args = make([]line.Arg, len(commandLIne))
    for i, value := range commandLIne {
        arg := line.Arg{}
        arg = append(arg, line.ArgPart{value, true})
        sh.args[i] = arg
    }
    return &sh
}

// Line returns a new sh.Sh configured with an Expander to control variable expansion.
// Use Expand(ExpandDisabled()) to disable expanding variables.
func (this *Sh) Expand(expander Expander) *Sh {
    var sh = *this;
    sh.expanders = expander
    return &sh
}

// Line returns a new sh.Sh configured with additional env variables to pass to the executed process
func (this *Sh) Env(env map[string]string) *Sh {
    var sh = *this;
    sh.env = env
    return &sh
}

// Dir returns a new sh.Sh configured with the directory to run the executed process.
func (this *Sh) Dir(dir string) *Sh {
    var sh = *this;
    sh.dir = dir
    return &sh
}

// CommandLog returns a new sh.Sh configured io.Writer that will receive the fully expanded command when the process is executed.
func (this *Sh) CommandLog(commandLog io.Writer) *Sh {
    var sh = *this;
    sh.commandLog = commandLog
    return &sh
}

// CommandLogPrefix returns a new sh.Sh configured with a prfefix to use when logging executed commands.
func (this *Sh) CommandLogPrefix(prefix string) *Sh {
    var sh = *this;
    sh.commandLogPrefix = prefix
    return &sh
}

// Cmd returns a new exec.Cmd configured with all the settings collected in the sh.Sh
func (sh *Sh) Cmd() *exec.Cmd {
    args := sh.expandArgs()
    path := ""
    if len(args) >= 1 {
        path = args[0]
        args = args[1:]
    }
    c := exec.Command(path, args...)
    c.Env = os.Environ()

    if sh.env != nil {
        for k, v := range sh.env {
            c.Env = append(c.Env, k+"="+v)
        }
    }
    c.Dir = sh.dir
    c.Stderr = os.Stderr
    c.Stdout = os.Stdout
    c.Stdin = os.Stdin
    return c
}

func (sh *Sh) expandArgs() []string {
    args := make([]string, len(sh.args))

    // Should we expand variables?
    if sh.expanders == ExpandDisabled() {
        for i, value := range sh.args {
            args[i] = value.String()
        }
    } else {
        exp := sh.expanders

        // If there is an env, then lets resolve from that first.
        if ( sh.env!=nil ) {
            exp = ChainExpanders(ExpandMap(sh.env), sh.expanders)
        }

        for i, value := range sh.args {
            args[i] = value.Expand(exp.Expand)
        }
    }
    return args
}

func (sh *Sh) String() string {
    args := sh.expandArgs()
    t := make([]string, len(args))
    for i, arg := range args {
        if needsQuote.MatchString(arg) {
            arg = strings.ReplaceAll(arg, `\`, `\\`)
            arg = strings.ReplaceAll(arg, "\r", `\r`)
            arg = strings.ReplaceAll(arg, "\n", `\n`)
            arg = strings.ReplaceAll(arg, "\t", `\t`)
            arg = strings.ReplaceAll(arg, `"`, `\"`)
            t[i] = `""` + arg + `""`
        }
    }
    return strings.Join(t, " ")
}

// Exec runs the command and returns the process exit code, and any error encountered when running the command.
func (sh *Sh) Exec() (rc int, err error) {
    c := sh.Cmd()
    if sh.commandLog != nil {
        fmt.Fprintln(sh.commandLog, sh.commandLogPrefix, sh.String())
    }
    err = c.Run()
    return magesh.ExitStatus(err), err
}

// MustExec runs the process and panics if it returns a non zero exit code..
func (sh *Sh) MustExec() {
    rc, err := sh.Exec()
    if rc != 0 {
        panic(fmt.Errorf("<%s> failed: return code=%d, error: %s", sh.String(), rc, err))
    }
}
