package walker

import (
	"hash"
	"os"
	"time"
)

func NewDiskFile(path string, info os.FileInfo) (File, error) {
	var err error
	if info == nil {
		info, err = os.Lstat(path)
		if err != nil {
			return nil, err
		}
	}
	return &diskfile{
		path: path,
		info: info,
	}, nil
}

type diskfile struct {
	path string
	info os.FileInfo
}

func (f *diskfile) Path() string {
	return f.path
}

func (f *diskfile) FileInfo() os.FileInfo {
	return f.info
}

func (f *diskfile) Opener() func() ([]byte, error) {
	return fileOnceOpener(f.path)
}

func (f *diskfile) Hash(h hash.Hash) (string, error) {
	return calcBytesHash(h, f.Opener())
}

func (f *diskfile) ModTime() time.Time {
	return f.info.ModTime()
}

func (f *diskfile) ChangeTime() time.Time {
	tstat := ExtractTimestat(f.info)
	return timespecToTime(tstat.CTime)
}

func (f *diskfile) AccessTime() time.Time {
	tstat := ExtractTimestat(f.info)
	return timespecToTime(tstat.ATime)
}
