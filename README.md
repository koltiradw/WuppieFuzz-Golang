# WuppieFuzz

TNO developed WuppieFuzz, a coverage-guided REST API fuzzer developed on top of
LibAFL, targeting a wide audience of end-users, with a strong focus on
ease-of-use, explainability of the discovered flaws and modularity. WuppieFuzz
supports all three settings of testing (black box, grey box and white box).

## WuppieFuzz-Golang

This package is based on the standard [feature](https://go.dev/doc/build-cover) for coverage profiling. 
It utilizes the external API of the Go standard library for dumping and resetting coverage data(WriteCounters and CleanCounters from `runtime/coverage` pkg) and 
the internal [API](./coverage/defs.go) for parsing the coverage format and transforming it into LCOV format.

### Usage

#### Setup

Install WuppieFuzz-Golang pkg, e.g:

```console
$ go get github.com/TNO-S3/WuppieFuzz-Golang
```

Import to main:

```golang
package main

import (
	"net/http"

	"github.com/gin-gonic/gin"

	_ "github.com/TNO-S3/WuppieFuzz-Golang"
)
```
Before build add cov flags, e.g:

```console
$ go build -cover -covemode=atomic
```

#### Starting PUT

You can run the program as usual without any additional steps. The library will automatically launch a 
separate goroutine at the start of the program to handle interactions with WuppieFuzz.

## Credits

PT Labs, Ivan Kapranov.
