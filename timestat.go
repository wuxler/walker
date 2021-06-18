package walker

import "syscall"

type Timestat struct {
	MTime syscall.Timespec
	CTime syscall.Timespec
	ATime syscall.Timespec
}
