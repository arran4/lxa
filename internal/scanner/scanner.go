package scanner

import (
	"io/fs"
	"os"
	"path/filepath"

	"strings"

	"github.com/lxa-project/lxa/internal/filter"
	"github.com/lxa-project/lxa/internal/xattr"
)

// FileInfo contains standard fs.FileInfo and XDG metadata.
type FileInfo struct {
	Path     string
	Info     fs.FileInfo
	Metadata xattr.Metadata
	Error    error
}

// Scanner traverses directories and collects file metadata.
type Scanner struct {
	reader xattr.Reader
	eval   *filter.Evaluator
	opts   Options
}

// FileSystem interface allows mocking the file system for tests.
type FileSystem interface {
	Lstat(name string) (fs.FileInfo, error)
	Stat(name string) (fs.FileInfo, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	WalkDir(root string, fn fs.WalkDirFunc) error
}

type osFS struct{}

func (osFS) Lstat(name string) (fs.FileInfo, error) {
	return os.Lstat(name)
}

func (osFS) Stat(name string) (fs.FileInfo, error) {
	return os.Stat(name)
}

func (osFS) ReadDir(name string) ([]fs.DirEntry, error) {
	return os.ReadDir(name)
}

func (osFS) WalkDir(root string, fn fs.WalkDirFunc) error {
	return filepath.WalkDir(root, fn)
}

// Options configure the scanner.
type Options struct {
	Recursive             bool
	Filter                string
	XDGOnly               bool // shorthand for filter "xdg"
	FS                    FileSystem
	ShowHidden            bool
	AlmostAll             bool
	IgnoreBackups         bool
	DirectoryOnly         bool
	HidePattern           string
	IgnorePattern         string
	DereferenceCmdLine    bool
	DereferenceCmdLineDir bool
	Dereference           bool
}

// New creates a new Scanner.
func New(r xattr.Reader, opts Options) (*Scanner, error) {
	var eval *filter.Evaluator
	var err error
	if opts.Filter != "" {
		eval, err = filter.NewEvaluator(opts.Filter)
		if err != nil {
			return nil, err
		}
	}

	if opts.FS == nil {
		opts.FS = osFS{}
	}

	return &Scanner{
		reader: r,
		eval:   eval,
		opts:   opts,
	}, nil
}

// Scan path(s). Returns a channel of FileInfo.
func (s *Scanner) Scan(paths []string) <-chan FileInfo {
	out := make(chan FileInfo)

	go func() {
		defer close(out)
		for _, path := range paths {
			s.scanPath(path, out)
		}
	}()

	return out
}

func (s *Scanner) scanPath(root string, out chan<- FileInfo) {
	var info fs.FileInfo
	var err error

	if s.opts.Dereference || s.opts.DereferenceCmdLine {
		info, err = s.opts.FS.Stat(root)
		if err != nil {
			// fallback to Lstat if Stat fails (e.g. broken symlink)
			info, err = s.opts.FS.Lstat(root)
		}
	} else if s.opts.DereferenceCmdLineDir {
		info, err = s.opts.FS.Lstat(root)
		if err == nil && info.Mode()&fs.ModeSymlink != 0 {
			targetInfo, targetErr := s.opts.FS.Stat(root)
			if targetErr == nil && targetInfo.IsDir() {
				info = targetInfo
			}
		}
	} else {
		info, err = s.opts.FS.Lstat(root)
	}

	if err != nil {
		out <- FileInfo{Path: root, Error: err}
		return
	}

	if s.opts.DirectoryOnly {
		s.emitIfMatches(root, info, out)
		return
	}

	// If it's a file, or if we're not recursive and it's not a directory we intend to list
	if !info.IsDir() {
		s.emitIfMatches(root, info, out)
		return
	}

	if !s.opts.Recursive {
		// Read top-level directory contents only
		entries, err := s.opts.FS.ReadDir(root)
		if err != nil {
			out <- FileInfo{Path: root, Error: err}
			return
		}

		// Always emit . and .. for the directory if requested by -a, but not -A
		if s.opts.ShowHidden && !s.opts.AlmostAll {
			dotInfo, err := s.opts.FS.Lstat(root)
			if err == nil {
				s.emitIfMatches(filepath.Join(root, "."), dotInfo, out)
			}
			parentInfo, err := s.opts.FS.Lstat(filepath.Join(root, ".."))
			if err == nil {
				s.emitIfMatches(filepath.Join(root, ".."), parentInfo, out)
			}
		}

		for _, e := range entries {
			childPath := filepath.Join(root, e.Name())

			childInfo, err := e.Info()
			if err != nil {
				out <- FileInfo{Path: childPath, Error: err}
				continue
			}

			if s.opts.Dereference {
				statInfo, err := s.opts.FS.Stat(childPath)
				if err == nil {
					childInfo = statInfo
				}
			}

			s.emitIfMatches(childPath, childInfo, out)
		}
		return
	}

	// Recursive walk
	s.opts.FS.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			out <- FileInfo{Path: path, Error: err}
			return nil // continue
		}

		if path == root {
			return nil // skip the root dir itself in recursive walk
		}

		childInfo, err := d.Info()
		if err != nil {
			out <- FileInfo{Path: path, Error: err}
			return nil
		}

		if s.opts.Dereference {
			statInfo, err := s.opts.FS.Stat(path)
			if err == nil {
				childInfo = statInfo
			}
		}

		s.emitIfMatches(path, childInfo, out)
		return nil
	})
}

func (s *Scanner) emitIfMatches(path string, info fs.FileInfo, out chan<- FileInfo) {
	base := filepath.Base(path)

	// -B, --ignore-backups
	if s.opts.IgnoreBackups && strings.HasSuffix(base, "~") {
		return
	}

	// Pattern ignores
	if s.opts.HidePattern != "" {
		if matched, _ := filepath.Match(s.opts.HidePattern, base); matched {
			// overridden by -a or -A
			if !s.opts.ShowHidden && !s.opts.AlmostAll {
				return
			}
		}
	}
	if s.opts.IgnorePattern != "" {
		if matched, _ := filepath.Match(s.opts.IgnorePattern, base); matched {
			return
		}
	}

	if !s.opts.ShowHidden && !s.opts.AlmostAll && strings.HasPrefix(base, ".") && base != "." && base != ".." {
		return
	}

	// If AlmostAll is set, do not list . and ..
	if s.opts.AlmostAll && (base == "." || base == "..") {
		return
	}

	md, err := xattr.ReadMetadata(s.reader, path)
	if err != nil {
		// Emit even on xattr error to show the file, but indicate error
		out <- FileInfo{Path: path, Info: info, Metadata: md, Error: err}
		return
	}

	if s.opts.XDGOnly && !md.HasXDG {
		return
	}

	if s.eval != nil && !s.eval.Eval(md) {
		return
	}

	out <- FileInfo{Path: path, Info: info, Metadata: md}
}
