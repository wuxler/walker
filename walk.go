package walker

import (
	"archive/tar"
	"context"
	"io"
	"os"
	"path/filepath"
	"sync"

	"github.com/pkg/errors"
	"github.com/saracen/walker"
)

type Checker func(f File) (err error)
type Visiter func(f File) error
type ErrorFilter func(pathname string, err error) error

func New() *walk {
	return NewWithContext(context.Background())
}

func NewWithContext(ctx context.Context) *walk {
	newctx, cancel := context.WithCancel(ctx)
	return &walk{
		ctx:    newctx,
		cancel: cancel,
	}
}

type walk struct {
	ctx    context.Context
	cancel context.CancelFunc
	mu     sync.Mutex

	checkers     []Checker
	visiters     []Visiter
	errorFilters []ErrorFilter
}

func (w *walk) OnVisit(fn Visiter) *walk {
	w.mu.Lock()
	defer w.mu.Unlock()

	if fn != nil {
		w.visiters = append(w.visiters, fn)
	}
	return w
}

func (w *walk) Check(fn Checker) *walk {
	w.mu.Lock()
	defer w.mu.Unlock()

	if fn != nil {
		w.checkers = append(w.checkers, fn)
	}
	return w
}

func (w *walk) FilterError(fn ErrorFilter) *walk {
	w.mu.Lock()
	defer w.mu.Unlock()

	if fn != nil {
		w.errorFilters = append(w.errorFilters, fn)
	}
	return w
}

func (w *walk) Cancel() {
	if w.cancel != nil {
		w.cancel()
	}
}

func (w *walk) WalkDir(root string) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Multiple goroutines stat the filesystem concurrently. The provided
	// walkFn must be safe for concurrent use.
	var mu sync.Mutex
	walkFn := func(pathname string, fi os.FileInfo) error {
		f, err := NewDiskFile(pathname, fi)
		if err != nil {
			return err
		}

		mu.Lock()
		defer mu.Unlock()

		if err := w.handleChecker(f); err != nil {
			if err == ErrCheckSkipDir {
				return filepath.SkipDir
			}
			return nil
		}

		if err := w.handleVisiters(f); err != nil {
			return err
		}

		return nil
	}

	errFn := func(pathname string, err error) error {
		return w.handleErrorFilters(pathname, err)
	}

	return walker.WalkWithContext(w.ctx, root, walkFn, walker.WithErrorCallback(errFn))
}

func (w *walk) WalkTarFile(path string) error {
	fd, err := os.Open(path)
	if err != nil {
		return err
	}
	defer fd.Close()

	return w.WalkTarReader(fd)
}

func (w *walk) WalkTarReader(r io.Reader) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	pathsToSkip := []string{}
	tr := tar.NewReader(r)
	for {
		select {
		case <-w.ctx.Done():
			return nil
		default:
			hdr, err := tr.Next()
			if err == io.EOF {
				return nil
			}
			if err != nil {
				return errors.Wrap(err, "fail to read the tarball")
			}
			f, err := NewTarFile(tr, hdr)
			if err != nil {
				return err
			}

			if err := HasAnyPrefix(pathsToSkip...)(f); err != nil {
				continue
			}

			if err := w.handleChecker(f); err != nil {
				if err == ErrCheckSkipDir {
					pathsToSkip = append(pathsToSkip, f.Path())
				}
				continue
			}

			if err := w.handleVisiters(f); err != nil {
				if nerr := w.handleErrorFilters(f.Path(), err); nerr == nil {
					continue
				}
				return err
			}
		}
	}
}

func (w *walk) handleChecker(f File) error {
	for _, checker := range w.checkers {
		if err := checker(f); err != nil {
			return err
		}
	}
	return nil
}

func (w *walk) handleVisiters(f File) error {
	for _, visiter := range w.visiters {
		if err := visiter(f); err != nil {
			return err
		}
	}
	return nil
}

func (w *walk) handleErrorFilters(pathname string, err error) error {
	for _, filter := range w.errorFilters {
		if err := filter(pathname, err); err == nil {
			return nil
		}
	}
	return err
}

// Error filters

func SkipPermissionError(pathname string, err error) error {
	if os.IsPermission(err) {
		return nil
	}
	return err
}
