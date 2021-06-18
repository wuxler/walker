package walker

import (
	"encoding/hex"
	"hash"
	"io"
	"io/ioutil"
	"sync"
	"syscall"
	"time"

	"github.com/pkg/errors"
)

func timespecToTime(ts syscall.Timespec) time.Time {
	return time.Unix(int64(ts.Sec), int64(ts.Nsec))
}

// fileOnceOpener opens a file once and the content is shared so that some analyzers can use the same data
func fileOnceOpener(filePath string) func() ([]byte, error) {
	var once sync.Once
	var b []byte
	var err error

	return func() ([]byte, error) {
		once.Do(func() {
			b, err = ioutil.ReadFile(filePath)
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to read file")
		}
		return b, nil
	}
}

// tarOnceOpener reads a file once and the content is shared so that some analyzers can use the same data
func tarOnceOpener(r io.Reader) func() ([]byte, error) {
	var once sync.Once
	var b []byte
	var err error

	return func() ([]byte, error) {
		once.Do(func() {
			b, err = ioutil.ReadAll(r)
		})
		if err != nil {
			return nil, errors.Wrap(err, "unable to read tar file")
		}
		return b, nil
	}
}

func calcBytesHash(h hash.Hash, opener func() ([]byte, error)) (string, error) {
	content, err := opener()
	if err != nil {
		return "", err
	}

	_, err = h.Write(content)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(h.Sum(nil)), nil
}
