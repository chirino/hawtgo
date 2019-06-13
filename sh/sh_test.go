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

package sh_test

import (
    "github.com/chirino/hawtgo/sh"
    `github.com/stretchr/testify/assert`
    `os/exec`
    `testing`
)

var ENV = map[string]string{
    `hello`: `world`,
}

func last(v []string) string {
    return v[len(v)-1]
}


func TestExecPath(t *testing.T) {
    assert := assert.New(t)

    expected, _ := exec.LookPath(`go`)
    c := sh.New().Env(ENV).Line(`go version`).Cmd()
    assert.Equal(expected, c.Path)
}

func TestCmdImmutableUsage(t *testing.T) {

    assert := assert.New(t)
    c := &exec.Cmd{}

    // Let's configure som common settings like env and reuse it
    // for multiple commands
    cmdEnv := sh.New().Env(ENV)


    // Verify common settings
    c = cmdEnv.Cmd()
    assert.Equal("", c.Dir)
    assert.Equal("hello=world", last(c.Env))

    // Verify we can set working dir for one execution.
    c = cmdEnv.Dir("/test").Cmd()
    assert.Equal("/test", c.Dir)
    assert.Equal("hello=world", last(c.Env))

    // Verify that it did not change working dir of common settings.
    c = cmdEnv.Cmd()
    assert.Equal("", c.Dir)
    assert.Equal("hello=world", last(c.Env))
}

func TestCmdLineUsage(t *testing.T) {
    assert := assert.New(t)
    c := &exec.Cmd{}
    cmdEnv := sh.New().Env(ENV)

    // Double Quotes do allow expansion and nested whitespace for a single arg
    c = cmdEnv.Line(`go "${hello} world"`).Cmd()
    assert.Equal([]string{`go`, `world world`}, c.Args)
    return

    // We expand variables by default.
    c = cmdEnv.Line(`go ${hello}`).Cmd()
    assert.Equal([]string{`go`, `world`}, c.Args)

    // White  space around arguements is removed.  Use quotes to preserve the white space.
    c = cmdEnv.Line("  go \t   hello \n  world \r rocks   ").Cmd()
    assert.Equal([]string{`go`, `hello`, `world`, `rocks`}, c.Args)

    // Use Expand(sh.EXPAND_OFF) to disable variable expansion...
    c = cmdEnv.Expand(sh.ExpandDisabled()).Line(`go ${hello}`).Cmd()
    assert.Equal([]string{`go`, `${hello}`}, c.Args)

    // Double Quotes do allow expansion and nested whitespace for a single arg
    c = cmdEnv.Line(`go "${hello} world"`).Cmd()
    assert.Equal([]string{`go`, `world world`}, c.Args)

    // Single Quotes disable expansion for a single arg
    c = cmdEnv.Line(`go '${hello}' ${hello}`).Cmd()
    assert.Equal([]string{`go`, `${hello}`, `world`}, c.Args)

    // Quotes don't have start or end an arg, and you use mutiples and mix the types
    // in a single argument
    c = cmdEnv.Line(`go ab'c def 'hig" lmn "opq`).Cmd()
    assert.Equal([]string{`go`, `abc def hig lmn opq`}, c.Args)

    // Unclosed single quote treats used the rest of the line as a single arg
    c = cmdEnv.Line(`go hi 'this is an unclosed quote arg`).Cmd()
    assert.Equal([]string{`go`, `hi`, `this is an unclosed quote arg`}, c.Args)

    // Same for double quote
    c = cmdEnv.Line(`go hi "this is an unclosed quote arg`).Cmd()
    assert.Equal([]string{`go`, `hi`, `this is an unclosed quote arg`}, c.Args)

    // You can use \" to escape in double quotes
    c = cmdEnv.Line(`go "hi \" there"`).Cmd()
    assert.Equal([]string{`go`, `hi " there`}, c.Args)

    // There's no escaping in single quotes.
    c = cmdEnv.Line(`go 'hi \' the ' other`).Cmd()
    assert.Equal([]string{`go`, `hi \`,  `the`, ` other`}, c.Args)

    // single quotes don't need escaping with double quotes and vice versa
    c = cmdEnv.Line(`go '"arg1"' "'arg2'"`).Cmd()
    assert.Equal([]string{`go`, `"arg1"`,  `'arg2'`}, c.Args)

}

func TestCmdLineArgsUsage(t *testing.T) {
    assert := assert.New(t)
    c := &exec.Cmd{}

    // Let's configure som common settings like env and reuse it
    // for multiple commands
    cmdEnv := sh.New().Env(ENV)

    // use Expand(false) to disable expansion for all arguments
    c = cmdEnv.Expand(sh.ExpandDisabled()).LineArgs(`go`, `'"${hello}"'`).Cmd()
    assert.Equal([]string{`go`, `'"${hello}"'`}, c.Args)


    // LineArgs allows you to be explicit about the arguments passed.. no quote processing is done.
    c = cmdEnv.LineArgs(`go`, `'"arg1"'` ,`"'arg2'"`).Cmd()
    assert.Equal([]string{`go`, `'"arg1"'` ,`"'arg2'"`}, c.Args)

    // All args passed with way will allow variable expansion.
    c = cmdEnv.LineArgs(`go`, `'"${hello}"'`).Cmd()
    assert.Equal([]string{`go`, `'"world"'`}, c.Args)

}