package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/tools/txtar"

	"github.com/lxa-project/lxa/internal/cli"
)

//go:embed testdata/*.txtar
var testdata embed.FS

type MockFileInfo struct {
	mode    fs.FileMode
	size    int64
	modTime time.Time
	uid     uint32
	gid     uint32
	sys     any
}

func (m *MockFileInfo) Name() string       { return "" }
func (m *MockFileInfo) Size() int64        { return m.size }
func (m *MockFileInfo) Mode() fs.FileMode  { return m.mode }
func (m *MockFileInfo) ModTime() time.Time { return m.modTime }
func (m *MockFileInfo) IsDir() bool        { return m.mode.IsDir() }
func (m *MockFileInfo) Sys() any {
	if m.sys != nil {
		return m.sys
	}
	return &syscall.Stat_t{
		Uid:   m.uid,
		Gid:   m.gid,
		Nlink: 1, // Add Nlink to show as '1' instead of '0'
	}
}

type MockDirEntry struct {
	name string
	info fs.FileInfo
}

func (m MockDirEntry) Name() string               { return m.name }
func (m MockDirEntry) IsDir() bool                { return m.info.IsDir() }
func (m MockDirEntry) Type() fs.FileMode          { return m.info.Mode().Type() }
func (m MockDirEntry) Info() (fs.FileInfo, error) { return m.info, nil }

type MockFS struct {
	files map[string]*MockFileInfo
}

func (m *MockFS) Lstat(name string) (fs.FileInfo, error) {
	if info, ok := m.files[name]; ok {
		return info, nil
	}
	return nil, fs.ErrNotExist
}

func (m *MockFS) ReadDir(name string) ([]fs.DirEntry, error) {
	var entries []fs.DirEntry
	for path, info := range m.files {
		dir := filepath.Dir(path)
		if dir == name && path != name {
			entries = append(entries, MockDirEntry{name: filepath.Base(path), info: info})
		}
	}
	return entries, nil
}

func (m *MockFS) WalkDir(root string, fn fs.WalkDirFunc) error {
	info, err := m.Lstat(root)
	if err != nil {
		return fn(root, nil, err)
	}
	err = fn(root, MockDirEntry{name: filepath.Base(root), info: info}, nil)
	if err != nil {
		if err == fs.SkipDir {
			return nil
		}
		return err
	}

	entries, _ := m.ReadDir(root)
	for _, entry := range entries {
		path := filepath.Join(root, entry.Name())
		if entry.IsDir() {
			err = m.WalkDir(path, fn)
			if err != nil && err != fs.SkipDir {
				return err
			}
		} else {
			info, _ := entry.Info()
			err = fn(path, MockDirEntry{name: entry.Name(), info: info}, nil)
			if err != nil && err != fs.SkipDir {
				return err
			}
		}
	}
	return nil
}

type mockXattrReader struct {
	xattrs map[string]map[string][]byte
}

func (m *mockXattrReader) List(path string) ([]string, error) {
	attrs, ok := m.xattrs[path]
	if !ok {
		return nil, nil
	}
	var names []string
	for k := range attrs {
		names = append(names, k)
	}
	return names, nil
}

func (m *mockXattrReader) Get(path, name string) ([]byte, error) {
	attrs, ok := m.xattrs[path]
	if !ok {
		return nil, syscall.ENODATA
	}
	val, ok := attrs[name]
	if !ok {
		return nil, syscall.ENODATA
	}
	return val, nil
}


type FileMetadata struct {
	Mode    uint32 `json:"mode"`
	Size    int64  `json:"size"`
	ModTime string `json:"modTime"`
	Uid     uint32 `json:"uid"`
	Gid     uint32 `json:"gid"`
}

type Options struct {
	Args         []string                  `json:"args"`
	Xattrs       map[string][]string       `json:"xattrs"`
	MockFiles    map[string]FileMetadata   `json:"mockFiles"`
	UseMockFS    bool                      `json:"useMockFS"`
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



			out := new(bytes.Buffer)
			errOut := new(bytes.Buffer)

			var targetDir string
			if opts.UseMockFS {
				// We still need to give it a root directory to start from
				targetDir = filepath.Join(dir, "input")

				// Change working dir so that relative paths in expected output are matched
				origWd, _ := os.Getwd()
				os.Chdir(targetDir)
				defer os.Chdir(origWd)
				targetDir = "."

				mockFS := &MockFS{files: make(map[string]*MockFileInfo)}
				mockReader := &mockXattrReader{xattrs: make(map[string]map[string][]byte)}

				// Add dir itself using "." because we are chdir'ing
				mockFS.files["."] = &MockFileInfo{
					mode: fs.ModeDir | 0755,
				}

				for relPath, md := range opts.MockFiles {
					// Use base name because we chdir into input
					fileName := filepath.Base(relPath)
					modTime := time.Now()
					if md.ModTime != "" {
						t, err := time.Parse(time.RFC3339, md.ModTime)
						if err == nil {
							modTime = t
						}
					}
					mockFS.files[fileName] = &MockFileInfo{
						mode:    fs.FileMode(md.Mode),
						size:    md.Size,
						modTime: modTime,
						uid:     md.Uid,
						gid:     md.Gid,
					}
				}

				if opts.Xattrs != nil {
					for relPath, attrs := range opts.Xattrs {
						fileName := filepath.Base(relPath)
						if mockReader.xattrs[fileName] == nil {
							mockReader.xattrs[fileName] = make(map[string][]byte)
						}
						for i := 0; i < len(attrs); i += 2 {
							if i+1 < len(attrs) {
								mockReader.xattrs[fileName][attrs[i]] = []byte(attrs[i+1])
							}
						}
					}
				}

				opts.Args = append(opts.Args, "-R", targetDir)
				err = cli.Run(opts.Args, out, errOut, cli.WithFS(mockFS), cli.WithXattrReader(mockReader))
			} else {
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
				targetDir = filepath.Join(dir, "input")
				args := append(opts.Args, "-R", targetDir)
				err = cli.Run(args, out, errOut)
			}

			if err != nil {
				t.Errorf("expected no error, got %v. errOut: %s", err, errOut.String())
			}

			// Clean whitespace padding introduced dynamically by text/tabwriter based on filepaths
			output := out.String()
			re := regexp.MustCompile(`[ \t]+`)
			normalizedOutput := re.ReplaceAllString(output, " ")

			expectedLines := strings.Split(strings.TrimSpace(expectedOutput), "\n")
			for _, line := range expectedLines {
				if line != "" {
					normalizedExpected := re.ReplaceAllString(line, " ")
					if !strings.Contains(normalizedOutput, normalizedExpected) {
						t.Errorf("expected output to contain %q, got:\n%s", normalizedExpected, normalizedOutput)
					}
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
