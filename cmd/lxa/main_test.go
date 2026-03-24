package main

import (
	"bytes"
	"embed"
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

			if entry.Name() == "xdg.txtar" {
				setXattr(t, filepath.Join(dir, "xdg_file.txt"), "user.xdg.tags", "test,foo\x00")
				setXattr(t, filepath.Join(dir, "xdg_file.txt"), "user.xdg.comment", "a comment\x00")
			}

			out := new(bytes.Buffer)
			errOut := new(bytes.Buffer)

			// pass --recursive to ensure it enters the generated dirs
			code := cli.Run([]string{"--recursive", dir}, out, errOut)
			if code != 0 {
				t.Errorf("expected 0, got %d. errOut: %s", code, errOut.String())
			}

			output := out.String()
			for _, f := range archive.Files {
				if !strings.Contains(output, filepath.Base(f.Name)) {
					t.Errorf("expected output to contain %q, got:\n%s", filepath.Base(f.Name), output)
				}
			}
		})
	}
}

func TestCLI_Help(t *testing.T) {
	out := new(bytes.Buffer)
	errOut := new(bytes.Buffer)

	code := cli.Run([]string{"-h"}, out, errOut)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if !strings.Contains(errOut.String(), "Usage:") {
		t.Errorf("expected Usage in output")
	}
}
