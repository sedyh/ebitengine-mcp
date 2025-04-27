package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

const (
	name    = "ebitengine-mcp-recorder-client"
	version = "1.0.0"
)

func main() {
	var command string
	flag.StringVar(&command, "client", "./cmd/server", "golang client to run")
	flag.Parse()

	goroot := os.Getenv("GOROOT")
	if goroot == "" {
		log.Fatal("env variable GOROOT is not set")
	}
	gobin := filepath.Join(goroot, "bin", "go")
	if _, err := os.Stat(gobin); os.IsNotExist(err) {
		log.Fatalf("go binary not found at %q", gobin)
	}
	c, err := client.NewStdioMCPClient(
		gobin,
		nil,
		"run",
		command,
	)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	log.Println("connecting:", fmt.Sprintf("%s %s %s", gobin, "run", command))
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    name,
		Version: version,
	}
	initRes, err := c.Initialize(ctx, initReq)
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}
	log.Printf(
		"connected: %s-%s\n",
		initRes.ServerInfo.Name,
		initRes.ServerInfo.Version,
	)

	log.Println("listing tools")
	toolsReq := mcp.ListToolsRequest{}
	toolsRes, err := c.ListTools(ctx, toolsReq)
	if err != nil {
		log.Fatalf("failed to get tools list: %v", err)
	}
	log.Printf("found %d tools:\n", len(toolsRes.Tools))
	for _, tool := range toolsRes.Tools {
		log.Printf("  - %s: %s\n", tool.Name, tool.Description)
	}

	helloWorldFound := false
	for _, tool := range toolsRes.Tools {
		if tool.Name == "hello_world" {
			helloWorldFound = true
			break
		}
	}
	if helloWorldFound {
		log.Println("calling hello_world tool")
		callHelloReq := mcp.CallToolRequest{}
		callHelloReq.Params.Name = "hello_world"
		callHelloReq.Params.Arguments = map[string]any{
			"name": name,
		}
		helloResult, err := c.CallTool(ctx, callHelloReq)
		if err != nil {
			log.Printf("failed to call hello_world tool: %v", err)
		} else {
			for _, content := range helloResult.Content {
				if textContent, ok := mcp.AsTextContent(content); ok {
					log.Printf("hello_world: %s\n", textContent.Text)
				}
			}
		}
	} else {
		log.Println("hello_world tool not found on server")
	}
}
