package main

import (
	"bytes"
	"flag"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"slices"

	"golang.org/x/tools/go/ast/astutil"
)

var DEFAULT_BUILD_SUPPRESSIONS = []string{"-gcflags=runtime/cgo=-d=libfuzzer=0",
	"-gcflags=runtime/pprof=-d=libfuzzer=0",
	"-gcflags=runtime/race=-d=libfuzzer=0",
	"-gcflags=syscall=-d=libfuzzer=0"}

var INSTRUMENTATION_FLAGS = []string{"-gcflags=all=-d=libfuzzer", "-tags=libfuzzer,gofuzz"}

func patchMainPkgImports(filename string, patchAction func(fset *token.FileSet, f *ast.File, name string, path string) bool) {
	var pkg string
	if debugInfo, ok := debug.ReadBuildInfo(); ok {
		pkg = debugInfo.Main.Path + "/agent"
	} else {
		log.Fatal("cannot read build info")
	}
	var fset = token.NewFileSet()
	file, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)

	if err != nil {
		log.Fatalf("failed to parse %s: %v", filename, err)
	}

	isPatched := patchAction(fset, file, "_", pkg)

	if !isPatched {
		log.Fatalf("failed to patch imports in %s", filename)
	}

	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		log.Fatalf("failed to format %s: %v", filename, err)
	}

	if err := os.WriteFile(filename, buf.Bytes(), 0644); err != nil {
		log.Fatalf("failed to write %s: %v", filename, err)
	}
}

func patchBuildArgs(buildArgs []string) []string {
	if buildArgs[0] != "build" {
		log.Fatalf("expected first argument to be \"build\", got %q", buildArgs[0])
	}
	newArgs := slices.Concat(INSTRUMENTATION_FLAGS, DEFAULT_BUILD_SUPPRESSIONS)
	return append(buildArgs[:1], append(newArgs, buildArgs[1:]...)...)
}

func buildProject(buildArgs []string) {
	newArgs := patchBuildArgs(buildArgs)

	cmd := exec.Command("go", newArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "CGO_ENABLED=1")

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		log.Fatalf("failed to execute go build: %v", err)
	}
}

func main() {
	flag.Parse()
	buildArgs := flag.Args()

	if len(buildArgs) == 0 {
		log.Fatal("no build arguments provided")
	}

	pathToMainPkg := filepath.Join(buildArgs[len(buildArgs)-1], "main.go")

	log.Print("patching main package imports")
	patchMainPkgImports(pathToMainPkg, astutil.AddNamedImport)
	log.Print("building project with instrumentation flags")
	buildProject(buildArgs)
	log.Print("restoring main package imports")
	patchMainPkgImports(pathToMainPkg, astutil.DeleteNamedImport)

}
