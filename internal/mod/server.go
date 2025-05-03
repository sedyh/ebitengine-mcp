package mod

import (
	"context"
	"errors"
	"os"

	"github.com/mark3labs/mcp-go/server"
	"github.com/sedyh/ebitengine-mcp/internal/event"
	"github.com/sedyh/ebitengine-mcp/internal/out"
)

const (
	name    = "ebitengine-mcp"
	version = "1.0.0"
)

type Server struct {
	mod *server.MCPServer
}

func NewServer(poll *event.Server, url, pub, sub string) *Server {
	mod := server.NewMCPServer(name, version)
	mod.AddTools(NewRecordTool(poll, url, pub, sub))
	return &Server{mod: mod}
}

func (s *Server) Start(ctx context.Context) error {
	stdio := server.NewStdioServer(s.mod)
	stdio.SetErrorLogger(out.DefaultLog("server"))
	err := stdio.Listen(ctx, os.Stdin, os.Stdout)
	if errors.Is(err, context.Canceled) {
		return nil
	}
	return err
}
