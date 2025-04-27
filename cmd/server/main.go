package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	const (
		name    = "ebitengine-mcp-recorder-server"
		version = "1.0.0"
	)

	s := server.NewMCPServer(
		name,
		version,
	)

	tool := mcp.NewTool(
		"hello_world",
		mcp.WithDescription("Say hello to someone"),
		mcp.WithString(
			"name",
			mcp.Required(),
			mcp.Description("Name of the person to greet"),
		),
	)
	s.AddTool(tool, helloHandler)

	if err := server.ServeStdio(s); err != nil {
		fmt.Println("Server error:", err)
	}
}

func helloHandler(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	name, ok := request.Params.Arguments["name"].(string)
	if !ok {
		return nil, errors.New("name must be a string")
	}

	return mcp.NewToolResultText(fmt.Sprintf("hello, %s!", name)), nil
}
