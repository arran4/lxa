# Bug Report: `gosubc` Import Cycle and Dead Code Generation

## Issue 1: Import Cycle When Using `--dir`
When running `gosubc generate` within the `cmd/lxa` package containing the `commands.go` file (e.g. `go generate .`), `gosubc` correctly parses `func Lxa` and `func Inspect`. However, the generated `root.go` and `inspect.go` files inject standard imports referring to `github.com/lxa-project/lxa/cmd/lxa/cmd` or the package itself natively.
Because these generated output files reside in `cmd/lxa` but import the very same namespace, the Go compiler immediately panics stating `import cycle not allowed`.

## Issue 2: Generation target folder discrepancy
Running `gosubc` with `--dir internal/cli` and producing the internal structures successfully outputs `commands.go` into `internal/cli`, but the generated files (`root.go`, `inspect.go`, etc) are randomly dumped into `internal/cli/cmd` under a completely different folder structure instead of inline, which causes dependency breaking.

## Issue 3: Missing Support for Arbitrary Positional Arguments without Value Error
When compiling positional arguments using `gosubc`, manual flags inside commands like `--max-tags-width` lacking explicit values immediately result in `flag requires a value` panicking due to `gosubc` overriding standard Go `flag` package behaviors.
