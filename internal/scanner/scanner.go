package scanner

import (
	"io/fs"
	"os"
	"path/filepath"

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

// Options configure the scanner.
type Options struct {
	Recursive bool
	Filter    string
	XDGOnly   bool // shorthand for filter "xdg"
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
	info, err := os.Lstat(root)
	if err != nil {
		out <- FileInfo{Path: root, Error: err}
		return
	}

	// If it's a file, or if we're not recursive and it's not a directory we intend to list
	if !info.IsDir() {
		s.emitIfMatches(root, info, out)
		return
	}

	if !s.opts.Recursive {
		// Read top-level directory contents only
		entries, err := os.ReadDir(root)
		if err != nil {
			out <- FileInfo{Path: root, Error: err}
			return
		}
		for _, e := range entries {
			childPath := filepath.Join(root, e.Name())
			childInfo, err := e.Info()
			if err != nil {
				out <- FileInfo{Path: childPath, Error: err}
				continue
			}
			s.emitIfMatches(childPath, childInfo, out)
		}
		return
	}

	// Recursive walk
	filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
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
		s.emitIfMatches(path, childInfo, out)
		return nil
	})
}

func (s *Scanner) emitIfMatches(path string, info fs.FileInfo, out chan<- FileInfo) {
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
