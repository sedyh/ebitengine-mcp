package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/sedyh/ebitengine-mcp/internal/event"
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
	id := flag.String("id", "", "id")
	flag.Parse()

	out.Setup(out.Level(*lvl))

	if *id == "" {
		slog.Error("id is required")
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	requester, err := event.NewClient[*event.RecordRequest](*url, *pub, *sub, *id)
	if err != nil {
		slog.Error("requester", "err", err)
		os.Exit(1)
	}

	responser, err := event.NewClient[*event.RecordResponse](*url, *pub, *sub, *id)
	if err != nil {
		slog.Error("responser", "err", err)
		os.Exit(1)
	}

	requests := requester.Start(time.Now().Add(-2 * time.Minute))
	defer requester.Stop()

	slog.Info("waiting for request")
	req := &event.RecordRequest{}
	select {
	case e := <-requests:
		req = e
	case <-ctx.Done():
		req.SetError(ctx.Err())
	}
	if errors.Is(req.Error(), context.Canceled) {
		slog.Info("stopped")
		return
	}
	if req.Error() != nil {
		slog.Error("requester", "err", req.Error())
		os.Exit(1)
	}
	slog.Info("request", "target", req.Target, "frames", req.Frames, "delay", req.Delay)
	if err := responser.Publish(*id, &event.RecordResponse{
		Images: []string{
			"0123456789",
			"3210987654",
			"8765432101",
		},
	}); err != nil {
		slog.Error("responser", "err", err)
		os.Exit(1)
	}
	slog.Info("responsed")
}
