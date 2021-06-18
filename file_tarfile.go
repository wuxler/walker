package walker

import (
	"archive/tar"
	"errors"
	"hash"
	"os"
	"time"
)

func NewTarFile(reader *tar.Reader, header *tar.Header) (File, error) {
	if reader == nil || header == nil {
		return nil, errors.New("tar reader or header is nil")
	}
	return &tarfile{
		reader: reader,
		header: header,
	}, nil
}

type tarfile struct {
	reader *tar.Reader
	header *tar.Header
}

func (f *tarfile) Path() string {
	return f.header.Name
}

func (f *tarfile) FileInfo() os.FileInfo {
	return f.header.FileInfo()
}

func (f *tarfile) Opener() func() ([]byte, error) {
	return tarOnceOpener(f.reader)
}

func (f *tarfile) Hash(h hash.Hash) (string, error) {
	return calcBytesHash(h, f.Opener())
}

func (f *tarfile) ModTime() time.Time {
	return f.header.ModTime
}

func (f *tarfile) ChangeTime() time.Time {
	return f.header.ChangeTime
}

func (f *tarfile) AccessTime() time.Time {
	return f.header.AccessTime
}
