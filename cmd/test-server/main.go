package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/lithammer/shortuuid"
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

	id := shortuuid.New()
	offset := time.Now().Add(-2 * time.Second)

	slog.Info("started", "pub", *pub, "sub", *sub, "url", poll.Host(), "id", id)

	client, err := event.NewClient[*event.RecordResponse](poll.Host(), *pub, *sub, id)
	if err != nil {
		slog.Error("responser", "err", err)
		os.Exit(1)
	}

	responses := client.Start(offset)
	defer client.Stop()

	if err := event.Publish(poll, id, &event.RecordRequest{
		Target: "./cmd/app",
		Frames: 2,
		Delay:  300 * time.Millisecond,
	}); err != nil {
		slog.Error("requester", "err", err)
		os.Exit(1)
	}
	slog.Info("published", "id", id)

	res := &event.RecordResponse{}
	select {
	case e := <-responses:
		res = e
	case <-ctx.Done():
		res.SetError(ctx.Err())
	}
	if res.Error() != nil && !errors.Is(res.Error(), context.Canceled) {
		slog.Error("responser", "err", res.Error())
	}
	for _, img := range res.Images {
		slog.Info("received", "image", out.Trunc(out.Hash([]byte(img)), 10))
	}
	cancel()

	ctx, cancel = context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	if err := poll.Stop(ctx); err != nil {
		slog.Error("requester", "err", err)
		os.Exit(1)
	}

	slog.Info("stopped", "graceful", !out.Done(ctx))
}
