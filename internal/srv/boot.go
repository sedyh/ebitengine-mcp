package srv

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	name    = "ebitengine-mcp-server"
	version = "1.0.0"
)

type Topic int

const (
	Record Topic = iota
)

type Events map[Topic]chan any

func NewEvents(topics ...Topic) Events {
	events := make(map[Topic]chan any)
	for _, topic := range topics {
		events[topic] = make(chan any, 1)
	}
	return events
}

type RecordReq struct {
	Frames int
	Delay  time.Duration
}

type RecordRes struct {
	Images []string
	Err    error
}

func Run(req, res Events) {
	s := server.NewMCPServer(
		name,
		version,
	)

	tool := mcp.NewTool(
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
	)
	s.AddTool(tool, func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		frames, ok := request.Params.Arguments["frames"].(float64)
		if !ok {
			return nil, errors.New("frames must be a number")
		}
		delay, ok := request.Params.Arguments["delay"].(float64)
		if !ok {
			return nil, errors.New("delay must be a number")
		}

		req[Record] <- RecordReq{
			Frames: int(frames),
			Delay:  time.Duration(delay) * time.Millisecond,
		}
		event := <-res[Record]
		res, ok := event.(RecordRes)
		if !ok {
			return nil, errors.New("invalid response")
		}

		call := &mcp.CallToolResult{}
		for _, png := range res.Images {
			call.Content = append(call.Content, mcp.ImageContent{
				MIMEType: "image/png",
				Type:     "image",
				Data:     png,
			})
		}
		return call, nil
	})

	if err := server.ServeStdio(s); err != nil {
		log.Fatal(err)
	}
}
