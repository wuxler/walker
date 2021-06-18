package walker

import (
	"errors"
	"path"
	"strings"
)

var (
	ErrCheckFail    error = errors.New("fail")
	ErrCheckSkipDir error = errors.New("skip dir")
)

func hasAnySubs(has func(s, sub string) bool, subs ...string) Checker {
	return func(f File) error {
		if len(subs) == 0 {
			return nil
		}
		for _, sub := range subs {
			if has(f.Path(), sub) {
				return nil
			}
		}
		return ErrCheckFail
	}
}

func hasAllSubs(has func(s, sub string) bool, subs ...string) Checker {
	return func(f File) error {
		for _, sub := range subs {
			if !has(f.Path(), sub) {
				return ErrCheckFail
			}
		}
		return nil
	}
}

func HasPrefix(prefix string) Checker {
	return HasAnyPrefix(prefix)
}

func HasAnyPrefix(prefixes ...string) Checker {
	return hasAnySubs(strings.HasPrefix, prefixes...)
}

func HasAllPrefix(prefixes ...string) Checker {
	return hasAllSubs(strings.HasPrefix, prefixes...)
}

func HasSuffix(suffix string) Checker {
	return HasAnySuffix(suffix)
}

func HasAnySuffix(suffixes ...string) Checker {
	return hasAnySubs(strings.HasSuffix, suffixes...)
}

func HasAllSuffix(suffixes ...string) Checker {
	return hasAllSubs(strings.HasSuffix, suffixes...)
}

func IsRegular() Checker {
	return func(f File) error {
		if !f.FileInfo().Mode().IsRegular() {
			return ErrCheckFail
		}
		return nil
	}
}

func SkipDir(dirname string) Checker {
	return func(f File) error {
		if f.FileInfo().IsDir() && path.Base(f.Path()) == dirname {
			return ErrCheckSkipDir
		}
		return nil
	}
}
