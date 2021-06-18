// +build linux

package walker

import (
	"os"
	"syscall"
)

func ExtractTimestat(fi os.FileInfo) Timestat {
	stat_t := fi.Sys().(*syscall.Stat_t)
	return Timestat{
		MTime: stat_t.Mtim,
		CTime: stat_t.Ctim,
		ATime: stat_t.Atim,
	}
}
