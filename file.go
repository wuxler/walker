package walker

import (
	"hash"
	"os"
	"time"
)

type File interface {
	Path() string
	FileInfo() os.FileInfo
	Opener() func() ([]byte, error)
	Hash(h hash.Hash) (string, error)

	ModTime() time.Time
	ChangeTime() time.Time
	AccessTime() time.Time
}
