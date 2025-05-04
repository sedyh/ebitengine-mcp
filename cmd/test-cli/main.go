package main

import (
	"context"
	"flag"
	"log/slog"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/sedyh/ebitengine-mcp/internal/cli"
	"github.com/sedyh/ebitengine-mcp/internal/out"
)

func main() {
	target := flag.String("target", "./cmd/test-bin", "target to run the test")
	url := flag.String(cli.FlagURL, "0.0.0.0:8080", "url to run the test")
	pub := flag.String(cli.FlagPub, "/pub", "pub to run the test")
	sub := flag.String(cli.FlagSub, "/sub", "sub to run the test")
	id := flag.String(cli.FlagID, shortuuid.New(), "id to run the test")
	flag.Parse()

	out.Setup(out.DefaultLevel)

	slog.Info("cli started", "target", *target, "url", *url, "pub", *pub, "sub", *sub, "id", *id)

	timeout := 1 * time.Millisecond
	slog.Info("run", "timeout", timeout)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := cli.Run(ctx, *target, *url, *pub, *sub, *id); err != nil {
		slog.Error("run", "err", err)
	}

	timeout = 500 * time.Millisecond
	slog.Info("run", "timeout", timeout)
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := cli.Run(ctx, *target, *url, *pub, *sub, *id); err != nil {
		slog.Error("run", "err", err)
	}

	timeout = 1500 * time.Millisecond
	slog.Info("run", "timeout", timeout)
	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	defer cancel()
	if err := cli.Run(ctx, *target, *url, *pub, *sub, *id); err != nil {
		slog.Error("run", "err", err)
	}

	slog.Info("cli stopped")
}
