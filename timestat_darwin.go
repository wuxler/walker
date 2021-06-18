// +build darwin

package walker

import (
	"os"
	"syscall"
)

func ExtractTimestat(fi os.FileInfo) Timestat {
	stat_t := fi.Sys().(*syscall.Stat_t)
	return Timestat{
		CTime: stat_t.Ctimespec,
		MTime: stat_t.Mtimespec,
		ATime: stat_t.Atimespec,
	}
}
