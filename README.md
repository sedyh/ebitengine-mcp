# Ebitengine Recorder MCP for Cursor

Just wrap your game before passing it to `ebiten.RunGame`.

```go
if err := ebiten.RunGame(mcp.Wrap(&Game{})); err != nil {
	log.Fatal(err)
}
```

And add this config to your mcp servers:

```json
{
  "mcpServers": {
    "ebitengine-mcp": {
      "command": "go run github.com/sedyh/ebitengine-mcp/cmd/trun",
      "args": [".", "-mcp"]
    }
  }
}
```
