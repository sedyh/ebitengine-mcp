package errs

import (
	"errors"
	"io"
	"os"
)

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
