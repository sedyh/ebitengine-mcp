package out

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"
)

const (
	debug = true
	path  = "debug.txt"
)

type deb struct {
	handler slog.Handler
}

func NewDebugHandler(handler slog.Handler) *deb {
	rem()
	return &deb{handler: handler}
}

func (h *deb) Enabled(ctx context.Context, level slog.Level) bool {
	return h.handler.Enabled(ctx, level)
}

func (h *deb) Handle(ctx context.Context, record slog.Record) error {
	if err := h.handler.Handle(ctx, record); err != nil {
		return err
	}

	var buf strings.Builder
	buf.WriteString(record.Time.Format(time.TimeOnly))
	buf.WriteString(" ")
	buf.WriteString(record.Level.String())
	buf.WriteString(" ")
	buf.WriteString(record.Message)

	record.Attrs(func(attr slog.Attr) bool {
		buf.WriteString(" ")
		buf.WriteString(attr.Key)
		buf.WriteString("=")
		buf.WriteString(attr.Value.String())
		return true
	})

	add(buf.String())

	return nil
}

func (h *deb) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &deb{handler: h.handler.WithAttrs(attrs)}
}

func (h *deb) WithGroup(name string) slog.Handler {
	return &deb{handler: h.handler.WithGroup(name)}
}

func add(log string) string {
	f, _ := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	_, _ = f.WriteString(log + "\n")
	return path
}

func rem() {
	_ = os.Remove(path)
}
