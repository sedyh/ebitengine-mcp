package mod

import (
	"context"
	"encoding/base64"
	"log/slog"
	"os"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sedyh/ebitengine-mcp/internal/cli"
	"github.com/sedyh/ebitengine-mcp/internal/out"
)

type Client struct {
	c *client.Client
}

func NewClient(dir string) (*Client, error) {
	gobin, err := cli.Compiler()
	if err != nil {
		return nil, err
	}

	c, err := client.NewStdioMCPClient(gobin, nil, "run", dir)
	if err != nil {
		return nil, err
	}

	return &Client{c}, nil
}

func (c *Client) Init(ctx context.Context) error {
	req := mcp.InitializeRequest{}
	req.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	req.Params.ClientInfo = mcp.Implementation{Name: Name, Version: Version}

	_, err := c.c.Initialize(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Tools(ctx context.Context) ([]mcp.Tool, error) {
	req := mcp.ListToolsRequest{}
	res, err := c.c.ListTools(ctx, req)
	if err != nil {
		return nil, err
	}

	return res.Tools, nil
}

func (c *Client) Record(
	ctx context.Context,
	target string, frames int, delay int,
) (hashes []string, logs []string, e error) {
	req := mcp.CallToolRequest{}
	req.Params.Name = "record"
	req.Params.Arguments = map[string]any{
		"target": target,
		"frames": frames,
		"delay":  delay,
	}

	res, err := c.c.CallTool(ctx, req)
	if err != nil {
		return nil, nil, err
	}

	hashes = make([]string, 0, len(res.Content))
	logs = make([]string, 0, len(res.Content))

	for _, content := range res.Content {
		if img, ok := mcp.AsImageContent(content); ok {
			hash := out.Short(img.Data)
			data, err := base64.StdEncoding.DecodeString(img.Data)
			if err != nil {
				slog.Error("decode image", "hash", hash, "err", err)
				continue
			}
			if err := os.WriteFile(hash+".png", data, 0644); err != nil {
				slog.Error("write image", "hash", hash, "err", err)
				continue
			}
			hashes = append(hashes, hash)
		}
		if txt, ok := mcp.AsTextContent(content); ok {
			logs = append(logs, txt.Text)
		}
	}

	return hashes, logs, nil
}

func (c *Client) Close() error {
	return c.c.Close()
}
