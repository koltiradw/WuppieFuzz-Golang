# WuppieFuzz

TNO developed WuppieFuzz, a coverage-guided REST API fuzzer developed on top of
LibAFL, targeting a wide audience of end-users, with a strong focus on
ease-of-use, explainability of the discovered flaws and modularity. WuppieFuzz
supports all three settings of testing (black box, grey box and white box).

## WuppieFuzz-Golang

A build-time coverage agent for Go programs. It wraps `go build` to transparently
link a [SANCov](https://clang.llvm.org/docs/SanitizerCoverage.html) 8-bit
counters coverage agent into any Go project and expose the live coverage map
over TCP for a fuzzer / coverage consumer (such as WuppieFuzz) to read.

## How it works

`wuppie-go` is a drop-in replacement for `go build`. For a build command like
`wuppie-go build ... ./cmd/server`, it:

1. **Patches** the target package's `main.go` to blank-import the agent package,
   so the agent's `init()` runs when the program starts.
2. **Builds** the project with libfuzzer instrumentation flags
   (`-gcflags=all=-d=libfuzzer`, `-tags=libfuzzer,gofuzz`) and `CGO_ENABLED=1`.
3. **Restores** `main.go` to its original state once the build succeeds.

At runtime the patched binary starts a TCP server on `0.0.0.0:1337` from the
imported agent. The server hooks `__sanitizer_cov_8bit_counters_init` via cgo to
capture the 8-bit counters map, then answers coverage dump requests from a
connected consumer.

## Usage

Install the CLI wrapper:

```bash
go install github.com/TNO-S3/WuppieFuzz-Golang/cmd/wuppie-go@latest
```

Add the agent package to your project (so `go mod` resolves it at build time):

```bash
go get github.com/TNO-S3/WuppieFuzz-Golang/agent
```

Build your project with `wuppie-go` instead of `go`:

```bash
wuppie-go build <your go build args> -o ./bin ./<path-to-main-package>
```

## Example

```bash
user@host ~/my-awesome-project (main)> wuppie-go build -a -o server ./cmd/server/
2026/06/02 16:52:35 patching main package imports
2026/06/02 16:52:35 building project with instrumentation flags
2026/06/02 16:53:10 restoring main package imports
```

## Configuration

The coverage agent listens on `0.0.0.0:1337` by default. Override the port by
setting the `WUPPIE_COVERAGE_PORT` environment variable before running the
patched binary:

```bash
WUPPIE_COVERAGE_PORT=8080 ./server
```

The chosen port is logged on startup, e.g.
`coverage agent listening on 0.0.0.0:8080`.

## Wire protocol

The agent speaks a simple length-prefixed binary protocol over a persistent TCP
connection. Each frame is:

```
| Magic (4) | Version (1) | Type (1) | Length (4, LE) | Payload (Length) |
```

- **Magic:** `"WGCA"` (`57 47 43 41`) — identifies the protocol on every frame.
- **Version:** `0x01`.
- **Types:** `0x01 REQUEST_DUMP` (payload = 1-byte reset flag) from the consumer;
  `0x02 RESPONSE_DUMP` (payload = the coverage map) from the agent.
- The map size is carried by `Length`; there is no separate `size` field.

On `REQUEST_DUMP` the agent copies the current counters map and, if the reset
flag is set, zeroes the counters afterwards.

## Credits

PT Labs, Ivan Kapranov.
