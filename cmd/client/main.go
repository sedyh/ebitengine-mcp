package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"os"
	"time"
	"unicode/utf8"

	"github.com/sedyh/ebitengine-mcp/internal/cli"
	"github.com/sedyh/ebitengine-mcp/internal/mod"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	var dir string
	flag.StringVar(&dir, "client", "./examples/record", "golang client to run")
	flag.Parse()

	gobin, err := cli.Compiler()
	if err != nil {
		log.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		log.Fatal(err)
	}
	c, err := client.NewStdioMCPClient(gobin, nil, "run", ".", "-"+cli.DefaultFlag)
	if err != nil {
		log.Fatal("failed to create client: ", err)
	}
	defer c.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	log.Println("connecting:", fmt.Sprintf("%s %s %s", gobin, "run", dir))
	initReq := mcp.InitializeRequest{}
	initReq.Params.ProtocolVersion = mcp.LATEST_PROTOCOL_VERSION
	initReq.Params.ClientInfo = mcp.Implementation{
		Name:    mod.Name,
		Version: mod.Version,
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

	found := false
	for _, tool := range toolsRes.Tools {
		if tool.Name == "record" {
			found = true
			break
		}
	}
	if found {
		log.Println("calling record tool")
		callRecordReq := mcp.CallToolRequest{}
		callRecordReq.Params.Name = "record"
		callRecordReq.Params.Arguments = map[string]any{
			"frames": 3,
			"delay":  100,
		}
		recordResult, err := c.CallTool(ctx, callRecordReq)
		if err != nil {
			log.Printf("failed to call record tool: %v", err)
		} else {
			for _, content := range recordResult.Content {
				if img, ok := mcp.AsImageContent(content); ok {
					log.Printf("record: %s\n", Short(img.Data))
				}
			}
		}
	} else {
		log.Println("record tool not found on server")
	}
}

func Short(str string) string {
	return Trunc(Hash([]byte(str)), 10)
}

func Trunc(str string, length int) string {
	if length <= 0 {
		return ""
	}
	if utf8.RuneCountInString(str) < length {
		return str
	}
	return string([]rune(str)[:length])
}

func Hash(data []byte) string {
	h := sha256.Sum256(data)
	return hex.EncodeToString(h[:])
}
