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
 package line_test

import (
    "github.com/chirino/hawtgo/sh/line"
    "github.com/stretchr/testify/assert"
    "testing"
)
var env = map[string]string{
    "hello": "world",
}


func TestParse(t *testing.T) {
    assert := assert.New(t)

    assert.Equal([]line.Arg{
        line.Arg{line.ArgPart{ "go", true}},
        line.Arg{line.ArgPart{ "${hello}", false}},
    }, line.Parse("go '${hello}'"))

    assert.Equal([]line.Arg{
        line.Arg{line.ArgPart{ "echo", true}},
        line.Arg{line.ArgPart{ "${hello}", true}},
    }, line.Parse("echo ${hello}"))

}
