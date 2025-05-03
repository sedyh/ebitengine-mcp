package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sedyh/ebitengine-mcp/internal/event"
	"github.com/sedyh/ebitengine-mcp/internal/out"
)

func main() {
	const (
		mcpName    = "ebitengine-mcp-server"
		mcpVersion = "1.0.0"
		pollURL    = ":8080"
		pollPub    = "/pub"
		pollSub    = "/sub"
		logLevel   = "debug"
	)

	url := flag.String("url", pollURL, "listen url")
	pub := flag.String("pub", pollPub, "publish url")
	sub := flag.String("sub", pollSub, "subscribe url")
	lvl := flag.String("log", logLevel, "log level")
	flag.Parse()

	out.Setup(out.Level(*lvl))

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	// Library communication

	slog.Info("started", "url", *url, "pub", *pub, "sub", *sub)
	poll, err := event.NewServer(*url, *pub, *sub)
	if err != nil {
		slog.Error("create-event-server", "err", err)
		os.Exit(1)
	}

	go func() {
		if err := poll.Start(ctx); err != nil {
			slog.Error("poll-server", "err", err)
			os.Exit(1)
		}
	}()

	// Model communication

	mod := server.NewMCPServer(mcpName, mcpVersion)
	mod.AddTool(
		mcp.NewTool(
			"record",
			mcp.WithDescription("record 1-5 frames with 100-1000 ms delay between them"),
			mcp.WithString(
				"target",
				mcp.Required(),
				mcp.Description("any type of path to main go package of the app, like: ./cmd/app, ../example/main.go, app/client"),
			),
			mcp.WithNumber(
				"frames",
				mcp.Required(),
				mcp.Description("number of frames to record, 1-5"),
			),
			mcp.WithNumber(
				"delay",
				mcp.Required(),
				mcp.Description("delay in milliseconds between frames, 100-1000"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			target, ok := request.Params.Arguments["target"].(string)
			if !ok {
				return nil, errors.New("target must be a string")
			}
			frames, ok := request.Params.Arguments["frames"].(float64)
			if !ok {
				return nil, errors.New("frames must be a number")
			}
			delay, ok := request.Params.Arguments["delay"].(float64)
			if !ok {
				return nil, errors.New("delay must be a number")
			}

			id := shortuuid.New()
			offset := time.Now().Add(-2 * time.Second)

			client, err := event.NewClient[*event.RecordResponse](*url, *pub, *sub, id)
			if err != nil {
				return nil, fmt.Errorf("create client: %w", err)
			}

			responses := client.Start(offset)
			defer client.Stop()

			if err := event.Publish(poll, id, &event.RecordRequest{
				Target: target,
				Frames: int(frames),
				Delay:  time.Duration(delay) * time.Millisecond,
			}); err != nil {
				return nil, fmt.Errorf("publish event: %w", err)
			}

			res := <-responses
			if res.Err != nil {
				return mcp.NewToolResultErrorFromErr("fail to record", res.Err), nil
			}

			call := &mcp.CallToolResult{}
			for _, png := range res.Images {
				content := mcp.NewImageContent(png, "image/png")
				call.Content = append(call.Content, content)
			}

			return call, nil
		})

	go func() {
		s := server.NewStdioServer(mod)
		s.SetErrorLogger(out.DefaultLog("serving"))
		err := s.Listen(ctx, os.Stdin, os.Stdout)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("mcp-server", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("listening", "name", mcpName, "version", mcpVersion)
	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := poll.Stop(ctx); err != nil {
		slog.Error("poll-server", "err", err)
		os.Exit(1)
	}

	slog.Info("stopped", "graceful", !out.Done(ctx))
}
