### MCP Server For Ebitengine

Download this package.

```bash
go get github.com/sedyh/ebitengine-mcp@v1.1.0
```

Wrap your game to record its state.

`examples/plain-record/main.go`
```go
import "github.com/sedyh/ebitengine-mcp/mcp"

func main() {
  if err := ebiten.RunGame(mcp.Wrap(NewGame())); err != nil {
    log.Fatal(err)
  }
}
```

And add this config to your mcp servers.

`~/.cursor/mcp.json`
```json
{
  "mcpServers": {
    "ebitengine-mcp": {
      "command": "go run github.com/sedyh/ebitengine-mcp/cmd/server@v1.1.0"
    }
  }
}
```

<details><summary>Other editors</summary>
  <br>
  <details><summary>VS Code</summary>
    <br>
    <code>~/.vscode/mcp.json</code>
    <br>
    <br>
    <pre><code lang="json">
    {
      "servers": {
        "ebitengine-mcp": {
          "type": "stdio",
          "command": "go",
          "args": ["run", "github.com/sedyh/ebitengine-mcp/cmd/server@v1.1.0"]
        }
      }
    }
    </code></pre>
  </details>
  <details><summary>Windsurf</summary>
    <br>
    <code>~/.codeium/windsurf/mcp_config.json</code>  
    <br>
    <br>
    <pre><code lang="json">
    {
      "mcpServers": {
        "ebitengine-mcp": {
          "command": "go",
          "args": ["run", "github.com/sedyh/ebitengine-mcp/cmd/server@v1.1.0"]
        }
      }
    }
    </code></pre>
  </details>
</details>

Ask your agent to debug something in your game, you can use `example/record` as an example.

<img src="https://github.com/user-attachments/assets/ef277f53-3fcd-4e83-a49a-f28eda7043bb" width="400">

### Available tools

#### All
- Capture build and launch logs.
- Capture app errors.
#### Record
- Capture N frames with M delay in milliseconds.

### Special cases

- ✅ **DrawFinalScreen**
- ❌ **LayoutF**

### Supported plugins and editors

[Feature support matrix: check tools tab.](https://modelcontextprotocol.io/clients)

- ✅ **Cursor**
- ✅ **Windsurf**
- ✅ **VS Code**
- ✅ **Claude Code**
- ✅ **Claude Desktop**
- ✅ **Cline**
- ✅ **Emacs MCP**
- ✅ **Neovim MCP**
- ❓ **Continue**
- ❓ **OpenSumi**
- ❓ **TheiaAI**
- ❓ **Roo Code**
- ❌ **Zed**
- ❌ **Trae**

### Architecture

Your llm-based editor runs a stdio mcp server that provides various tools for working with the game in your project. The editor specifies the settings and location for running the project, and the server assembles it and passes certain flags on startup, which are picked up by the decorator embedded in the game. The decorator listens for requests to run tools, executes them, and returns a response via a reverse connection to the server, after which it closes. The server supplements the response with application logs and adapts the response to the mcp context. The server remains running as long as editor wants.

![](https://github.com/user-attachments/assets/df2e9025-7927-4c1d-b3e8-40a434ddffbd)

### Commands

#### `test-server`, `test-client`

Checking the operation of a message via a reverse connection based on long polling.

#### `test-cli`, `test-bin`

Testing a universal builder that can run a project from anywhere outside the working directory.

#### `server`, `client`

Testing the work via mcp together with the entire chain of message processing.
