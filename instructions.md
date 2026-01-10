# Project Context and Instructions

## Project Overview
`youtube-uploader-mcp` is a Model Context Protocol (MCP) server written in Go. It enables AI agents (like Claude, Cursor, etc.) to securely upload videos to YouTube using OAuth2 for authentication, without exposing credentials to the LLM.

## Architecture
The project follows a clean separation of concerns:
- **`main.go`**: Entry point. Parses flags, initializes the YouTube client, and starts the MCP server using `github.com/mark3labs/mcp-go`. Registers all available tools.
- **`tool/`**: Contains the `Tool` interface and concrete implementations of MCP tools. Each tool is in its own subdirectory.
  - **Interface**: defined in `tool/tool.go`.
- **`youtube/`**: Service layer interacting with the YouTube Data API v3. Handles OAuth2 tokens (`oauth.go`), channel logic (`channel.go`), and video operations (`video.go`).

## Code Conventions

### 1. Tool Implementation
To add a new tool, create a new package under `tool/` (e.g., `tool/mytool`) and implement the `Tool` interface:

```go
type Tool interface {
    Name() string
    Define(ctx context.Context) mcp.Tool
    Handle(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)
}
```

*   **Define**: Use `mcp.NewTool` to define the schema. Use helper methods like `mcp.WithString`, `mcp.WithDescription`.
*   **Handle**:
    *   Extract arguments using `request.GetString(key, default)` or `request.GetBool(...)`.
    *   Validate inputs explicitly.
    *   Use `mcp.NewToolResultError(msg)` for errors.
    *   Use `mcp.NewToolResultText(content)` for success. If returning an object, marshal it to JSON first.

### 2. Service Layer (`youtube/`)
*   Business logic resides here, not in the tool handler.
*   Functions should accept `context.Context` and `*oauth2.Token` (if auth is needed).
*   Return standard Go errors wrapped with context (e.g., `fmt.Errorf("failed to ...: %w", err)`).

### 3. Error Handling
*   Return friendly error messages to the LLM via `mcp.NewToolResultError`.
*   Log internal details using `logn.Infof` if necessary (but keep user output clean).

### 4. Style
*   Follow standard Go idioms (fmt, linting).
*   Use the internal `logn` package for logging.

## Extensibility Guide

### Adding a New Feature
1.  **Service Logic**: Implement the core logic in the `youtube/` package.
2.  **Tool Wrapper**: Create a new directory in `tool/` (e.g., `tool/comment`).
3.  **Implementation**: Create the struct implementing `Tool` interface.
4.  **Registration**: Add the new tool to the `tools` slice in `main.go`.

### Example: Adding a "List Comments" Tool
1.  Add `GetComments` function in `youtube/comment.go`.
2.  Create `tool/comment/list.go`.
3.  Implement `Define` to take `video_id`.
4.  Implement `Handle` to call `youtube.GetComments` and return JSON.
5.  Register `&comment.ListTool{}` in `main.go`.
