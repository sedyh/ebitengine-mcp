package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	"github.com/sedyh/ebitengine-mcp/internal/mod"
	"github.com/sedyh/ebitengine-mcp/internal/out"
)

func main() {
	server := flag.String("server", "./cmd/server", "mcp stdio server to run")
	target := flag.String("target", "./examples/final-record", "ebitengine game to run")
	lvl := flag.String("log", "debug", "log level")
	flag.Parse()

	out.Setup(out.Level(*lvl))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	client, err := mod.NewClient(*server)
	if err != nil {
		slog.Error("client", "err", err)
		os.Exit(1)
	}
	defer client.Close()

	if err := client.Init(ctx); err != nil {
		slog.Error("client", "err", err)
		os.Exit(1)
	}

	slog.Info("connected")

	tools, err := client.Tools(ctx)
	if err != nil {
		slog.Error("client", "err", err)
		os.Exit(1)
	}
	for _, tool := range tools {
		slog.Info("found tool", "name", tool.Name, "description", tool.Description)
	}

	hashes, logs, err := client.Record(ctx, *target, 3, 100)
	if err != nil {
		slog.Error("client", "err", err)
		os.Exit(1)
	}
	for _, hash := range hashes {
		slog.Info("image", "hash", hash)
	}
	for _, log := range logs {
		slog.Info("log", "msg", log)
	}

	<-ctx.Done()

	slog.Info("stopped", "graceful", !out.Done(ctx))
}
