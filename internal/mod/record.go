package mod

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/lithammer/shortuuid"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/sedyh/ebitengine-mcp/internal/cli"
	"github.com/sedyh/ebitengine-mcp/internal/event"
)

type RecordTool struct {
	server.ServerTool
	poll *event.Server
	url  string
	pub  string
	sub  string
}

func NewRecordTool(poll *event.Server, url, pub, sub string) server.ServerTool {
	r := RecordTool{
		poll: poll,
		url:  url,
		pub:  pub,
		sub:  sub,
	}

	return server.ServerTool{
		Tool:    r.Tool(),
		Handler: r.Handle,
	}
}

func (r *RecordTool) Tool() mcp.Tool {
	return mcp.NewTool(
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
	)
}

func (r *RecordTool) Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	res := r.Call(ctx, &event.RecordRequest{
		Target: target,
		Frames: int(frames),
		Delay:  time.Duration(delay) * time.Millisecond,
	})
	if res.Err != nil {
		return mcp.NewToolResultErrorFromErr("fail to record", res.Err), nil
	}

	call := &mcp.CallToolResult{}
	for _, png := range res.Images {
		content := mcp.NewImageContent(png, "image/png")
		call.Content = append(call.Content, content)
	}

	return call, nil
}

func (r *RecordTool) Call(ctx context.Context, req *event.RecordRequest) (res *event.RecordResponse) {
	id := shortuuid.New()
	go cli.Run(ctx, req.Target, r.url, r.pub, r.sub, id)

	responser, err := event.NewClient[*event.RecordResponse](r.url, r.pub, r.sub, id)
	if err != nil {
		res.Error(fmt.Errorf("create client: %w", err))
		return res
	}

	responses := responser.Start(time.Now())
	defer responser.Stop()

	if err := event.Publish(r.poll, id, req); err != nil {
		res.Error(fmt.Errorf("publish event: %w", err))
		return res
	}

	select {
	case res = <-responses:
	case <-ctx.Done():
		res.Error(ctx.Err())
	}

	return res
}
