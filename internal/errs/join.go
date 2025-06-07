package errs

import (
	"errors"
	"io"
	"os"
)

type CloserFunc func() error

func (f CloserFunc) Close() error {
	return f()
}

func Closer(dest *error, c io.Closer) {
	if c == nil {
		return
	}
	err := c.Close()
	if errors.Is(err, os.ErrClosed) {
		return
	}
	*dest = errors.Join(*dest, err)
}
