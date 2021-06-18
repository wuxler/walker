package walker

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

var (
	scaffolingRoot string
	prepandRoot    = func(s string) string {
		return filepath.Join(scaffolingRoot, s)
	}
	trimRoot = func(s string) string {
		p := strings.TrimPrefix(s, scaffolingRoot)
		if p == "" {
			return string(filepath.Separator)
		}
		return p
	}
	fileEntries []creater = []creater{
		file{"/d0/f1"},               //
		file{"/d0/d1/f2"},            //
		file{"/d0/skips/d2/f3"},      // node precedes skip
		file{"/d0/skips/d2/skip"},    // skip is non-directory
		file{"/d0/skips/d2/z1"},      // node follows skip non-directory: should never be visited
		file{"/d0/skips/d3/f4"},      // node precedes skip
		file{"/d0/skips/d3/skip/f5"}, // skip is directory: this node should never be visited
		file{"/d0/skips/d3/z2"},      // node follows skip directory: should be visited
	}
	linkEntries []creater = []creater{
		link{"/d0/symlinks/nothing", "../f0"},    // referent will be deleted
		link{"/d0/symlinks/toF1", "../f1"},       //
		link{"/d0/symlinks/toD1", "../d1"},       //
		link{"/d0/symlinks/d4/toSD1", "../toD1"}, // chained symbolic links
		link{"/d0/symlinks/d4/toSF1", "../toF1"}, // chained symbolic links
	}
)

const (
	exitAsSuccess = 0
	exitAsFail    = 1
)

// TestMain will setup a filetree relative to "/tmp/walker-602940484" which is a temporary directory create with ioutil.TempDir:
// 		drwx------ /
// 		drwxrwxr-x /d0
// 		drwxrwxr-x /d0/d1
// 		-rwxrwxr-x /d0/d1/f2
// 		-rwxrwxr-x /d0/f1
// 		drwxrwxr-x /d0/skips
// 		drwxrwxr-x /d0/skips/d2
// 		-rwxrwxr-x /d0/skips/d2/f3
// 		-rwxrwxr-x /d0/skips/d2/skip
// 		-rwxrwxr-x /d0/skips/d2/z1
// 		drwxrwxr-x /d0/skips/d3
// 		-rwxrwxr-x /d0/skips/d3/f4
// 		drwxrwxr-x /d0/skips/d3/skip
// 		-rwxrwxr-x /d0/skips/d3/skip/f5
// 		-rwxrwxr-x /d0/skips/d3/z2
// 		drwxrwxr-x /d0/symlinks
// 		drwxrwxr-x /d0/symlinks/d4
// 		Lrwxrwxrwx /d0/symlinks/d4/toSD1 -> ../toD1
// 		Lrwxrwxrwx /d0/symlinks/d4/toSF1 -> ../toF1
// 		Lrwxrwxrwx /d0/symlinks/nothing -> ../f0
// 		Lrwxrwxrwx /d0/symlinks/toD1 -> ../d1
// 		Lrwxrwxrwx /d0/symlinks/toF1 -> ../f1
func TestMain(m *testing.M) {
	code := exitAsSuccess

	defer func() {
		if err := teardown(); err != nil {
			fmt.Fprintf(os.Stderr, "worker teardown fail: %s\n", err)
			code = exitAsFail
		}
		os.Exit(code)
	}()

	if err := setup(); err != nil {
		fmt.Fprintf(os.Stderr, "worker setup fail: %s\n", err)
		dumpDirectory()
		code = exitAsFail
		return
	}

	dumpDirectory()

	code = m.Run()

	if code != exitAsSuccess {
		dumpDirectory()
	}
}

func setup() error {
	var err error
	scaffolingRoot, err = ioutil.TempDir(os.TempDir(), "walker-")
	if err != nil {
		return err
	}

	entries := []creater{
		file{"/d0/f0"}, // will be deleted after symlink for it created
	}
	entries = append(entries, fileEntries...)
	entries = append(entries, linkEntries...)

	for _, entry := range entries {
		if err := entry.Create(); err != nil {
			return fmt.Errorf("cannot create scaffolding entry: %s", err)
		}
	}

	// oldname, err := filepath.Abs(prepandRoot("/d0/f1"))
	// if err != nil {
	// 	return fmt.Errorf("cannot create scaffolding entry: %s", err)
	// }
	// if err := (link{"/d0/symlinks/toAbs", oldname}).Create(); err != nil {
	// 	return fmt.Errorf("cannot create scaffolding entry: %s", err)
	// }

	if err := os.Remove(prepandRoot("/d0/f0")); err != nil {
		return fmt.Errorf("cannot remove file from test scaffolding: %s", err)
	}

	return nil
}

func teardown() error {
	if scaffolingRoot == "" {
		return nil
	}
	return os.RemoveAll(scaffolingRoot)
}

func dumpDirectory() {
	trim := len(scaffolingRoot) // trim rootDir from prefix of strings
	err := filepath.Walk(scaffolingRoot, func(pathname string, info os.FileInfo, err error) error {
		if err != nil {
			// we have no info, so get it
			info, err2 := os.Lstat(pathname)
			if err2 != nil {
				fmt.Fprintf(os.Stderr, "?--------- %s: %s\n", pathname[trim:], err2)
			} else {
				fmt.Fprintf(os.Stderr, "%s %s: %s\n", info.Mode(), pathname[trim:], err)
			}
			return nil
		}

		var suffix string

		if info.Mode()&os.ModeSymlink != 0 {
			referent, err := os.Readlink(pathname)
			if err != nil {
				suffix = fmt.Sprintf(": cannot read symlink: %s", err)
				err = nil
			} else {
				suffix = fmt.Sprintf(" -> %s", referent)
			}
		}
		p := pathname[trim:]
		if p == "" {
			p = "/"
		}
		fmt.Fprintf(os.Stderr, "%s %s%s\n", info.Mode(), p, suffix)
		return nil
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "cannot walk test directory: %s\n", err)
	}
}

type creater interface {
	Create() error
	Name() string
}

type file struct {
	name string
}

func (f file) Create() error {
	absname := prepandRoot(filepath.FromSlash(f.name))
	if err := os.MkdirAll(filepath.Dir(absname), os.ModePerm); err != nil {
		return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
	}
	if err := ioutil.WriteFile(absname, []byte(f.name+"\n"), os.ModePerm); err != nil {
		return fmt.Errorf("cannot create file for test scaffolding: %s", err)
	}
	return nil
}

func (f file) Name() string {
	return f.name
}

type link struct {
	name     string
	referent string
}

func (l link) Create() error {
	newname := prepandRoot(filepath.FromSlash(l.name))
	if err := os.MkdirAll(filepath.Dir(newname), os.ModePerm); err != nil {
		return fmt.Errorf("cannot create directory for test scaffolding: %s", err)
	}
	oldname := filepath.FromSlash(l.referent)
	if err := os.Symlink(oldname, newname); err != nil {
		return fmt.Errorf("cannot create symbolic link for test scaffolding: %s", err)
	}
	return nil
}

func (l link) Name() string {
	return l.name
}
