package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/sedyh/ebitengine-mcp/internal/event"
	"github.com/sedyh/ebitengine-mcp/internal/mod"
	"github.com/sedyh/ebitengine-mcp/internal/out"
)

func main() {
	const (
		pollURL  = ":0"
		pollPub  = "/pub"
		pollSub  = "/sub"
		logLevel = "debug"
	)
	url := flag.String("url", pollURL, "listen url")
	pub := flag.String("pub", pollPub, "publish url")
	sub := flag.String("sub", pollSub, "subscribe url")
	lvl := flag.String("log", logLevel, "log level")
	flag.Parse()

	out.Setup(out.Level(*lvl))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	poll, err := event.NewServer(*url, *pub, *sub)
	if err != nil {
		slog.Error("requester", "err", err)
		os.Exit(1)
	}
	go func() {
		if err := poll.Start(ctx); err != nil {
			slog.Error("requester", "err", err)
			os.Exit(1)
		}
	}()

	server := mod.NewServer(poll, poll.Host(), *pub, *sub)
	go func() {
		if err := server.Start(ctx); err != nil {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("started", "pub", *pub, "sub", *sub, "url", poll.Host())

	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := poll.Stop(ctx); err != nil {
		slog.Error("requester", "err", err)
		os.Exit(1)
	}

	slog.Info("stopped", "graceful", !out.Done(ctx))
}
