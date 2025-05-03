package out

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/phsym/console-slog"
)

func Setup(level slog.Level) {
	slog.SetDefault(slog.New(
		console.NewHandler(
			Prefixed(os.Stderr, "\r"),
			&console.HandlerOptions{
				Level:      level,
				TimeFormat: time.TimeOnly,
			},
		),
	))
}

func Level(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type adapter struct {
	logger *slog.Logger
	msg    string
}

func DefaultLog(msg string) *log.Logger {
	adapter := &adapter{logger: slog.Default(), msg: msg}
	return log.New(adapter, "", 0)
}

func (a *adapter) Write(p []byte) (n int, err error) {
	s := strings.ToLower(string(p))
	s = strings.TrimSuffix(s, "\n")
	if idx := strings.Index(s, ":"); idx >= 0 {
		s = strings.TrimLeft(s[idx+1:], " ")
	}
	if strings.Contains(s, context.Canceled.Error()) {
		return len(p), nil
	}
	a.logger.Error(a.msg, "err", errors.New(s))
	return len(p), nil
}

type prefixed struct {
	w      io.Writer
	prefix string
	buf    []byte
}

func Prefixed(w io.Writer, prefix string) *prefixed {
	return &prefixed{
		w:      w,
		prefix: prefix,
		buf:    make([]byte, 0, 1024),
	}
}

func (r *prefixed) Write(p []byte) (n int, err error) {
	r.buf = append(r.buf, p...)

	for {
		idx := bytes.IndexByte(r.buf, '\n')
		if idx < 0 {
			break
		}
		line := r.buf[:idx+1]
		_, err = r.w.Write([]byte(r.prefix))
		if err != nil {
			return 0, err
		}
		_, err = r.w.Write(line)
		if err != nil {
			return 0, err
		}
		r.buf = r.buf[idx+1:]
	}

	return len(p), nil
}

func Done(ctx context.Context) bool {
	select {
	case <-ctx.Done():
		return true
	default:
		return false
	}
}
