package cli

import (
	"context"
	"log/slog"
)

type Options struct {
	Target string
	URL    string
	Pub    string
	Sub    string
	ID     string
}

type Result struct {
	Logs  []string
	Error error
}

func Go(ctx context.Context, opts Options) (res chan Result) {
	res = make(chan Result, 1)
	go func() {
		logs, err := Run(ctx, opts.Target, opts.URL, opts.Pub, opts.Sub, opts.ID)
		res <- Result{
			Logs:  Unwrap(logs),
			Error: err,
		}
		slog.Info("executed", "logs", len(logs), "err", err)
	}()
	return res
}
