package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/jcuga/golongpoll"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
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
	m, err := golongpoll.StartLongpoll(golongpoll.Options{})
	if err != nil {
		slog.Error("polling", "err", err)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc(*sub, m.SubscriptionHandler)
	mux.HandleFunc(*pub, m.PublishHandler)
	poll := &http.Server{Addr: *url, Handler: mux}

	go func() {
		err := poll.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("listening", "err", err)
			os.Exit(1)
		}
	}()

	// Model communication

	mod := server.NewMCPServer(mcpName, mcpVersion)
	mod.AddTool(
		mcp.NewTool(
			"record",
			mcp.WithDescription("record n frames every m milliseconds and exit"),
			mcp.WithNumber(
				"frames",
				mcp.Required(),
				mcp.Description("number of frames to record"),
			),
			mcp.WithNumber(
				"delay",
				mcp.Required(),
				mcp.Description("delay in milliseconds between frames"),
			),
		),
		func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			frames, ok := request.Params.Arguments["frames"].(float64)
			if !ok {
				return nil, errors.New("frames must be a number")
			}
			delay, ok := request.Params.Arguments["delay"].(float64)
			if !ok {
				return nil, errors.New("delay must be a number")
			}

			_, _ = int(frames), time.Duration(delay)*time.Millisecond
			call := &mcp.CallToolResult{}

			return call, nil
		})

	go func() {
		s := server.NewStdioServer(mod)
		s.SetErrorLogger(out.DefaultLog("serving"))
		err := s.Listen(ctx, os.Stdin, os.Stdout)
		if err != nil && !errors.Is(err, context.Canceled) {
			slog.Error("serving", "err", err)
			os.Exit(1)
		}
	}()

	slog.Info("listening", "name", mcpName, "version", mcpVersion)
	<-ctx.Done()

	ctx, cancel = context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := poll.Shutdown(ctx); err != nil {
		slog.Error("shutdown", "err", err)
		os.Exit(1)
	}

	slog.Info("stopped")
}
