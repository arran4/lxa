package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"

	"golang.org/x/tools/txtar"

	"github.com/lxa-project/lxa/internal/cli"
)

//go:embed testdata/*.txtar
var testdata embed.FS

type Options struct {
	Args   []string            `json:"args"`
	Xattrs map[string][]string `json:"xattrs"`
}

func extractTxtar(t *testing.T, archive *txtar.Archive) string {
	t.Helper()
	dir := t.TempDir()
	for _, f := range archive.Files {
		path := filepath.Join(dir, f.Name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, f.Data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func setXattr(t *testing.T, path, name, value string) {
	t.Helper()
	err := syscall.Setxattr(path, name, []byte(value), 0)
	if err != nil && err != syscall.ENOTSUP {
		t.Logf("Setxattr failed (maybe unsupported fs): %v", err)
	}
}

func TestScenarios(t *testing.T) {
	entries, err := testdata.ReadDir("testdata")
	if err != nil {
		t.Fatal(err)
	}

	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".txtar") {
			continue
		}

		t.Run(entry.Name(), func(t *testing.T) {
			b, err := testdata.ReadFile(filepath.Join("testdata", entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			archive := txtar.Parse(b)
			dir := extractTxtar(t, archive)

			var opts Options
			var expectedOutput string
			for _, f := range archive.Files {
				if f.Name == "options.json" {
					if err := json.Unmarshal(f.Data, &opts); err != nil {
						t.Fatal(err)
					}
				} else if f.Name == "expected/output.txt" {
					expectedOutput = string(f.Data)
				}
			}

			if opts.Xattrs != nil {
				for relPath, attrs := range opts.Xattrs {
					fullPath := filepath.Join(dir, relPath)
					for i := 0; i < len(attrs); i += 2 {
						if i+1 < len(attrs) {
							setXattr(t, fullPath, attrs[i], attrs[i+1])
						}
					}
				}
			}

			out := new(bytes.Buffer)
			errOut := new(bytes.Buffer)

			args := append(opts.Args, "-R", filepath.Join(dir, "input"))
			err = cli.Run(args, out, errOut)
			if err != nil {
				t.Errorf("expected no error, got %v. errOut: %s", err, errOut.String())
			}

			output := out.String()

			expectedLines := strings.Split(strings.TrimSpace(expectedOutput), "\n")
			for _, line := range expectedLines {
				if line != "" && !strings.Contains(output, line) {
					t.Errorf("expected output to contain %q, got:\n%s", line, output)
				}
			}
		})
	}
}

func TestCLI_Help(t *testing.T) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)

	err := cli.Run([]string{"-h"}, out, errOut)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if !strings.Contains(errOut.String(), "Usage:") {
		t.Errorf("expected Usage in output")
	}
}
