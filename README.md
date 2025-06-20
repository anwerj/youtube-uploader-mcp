# YouTube Uploader MCP

AI‑powered YouTube uploader—no CLI, no YouTube Studio, and no secrets ever shared with LLMs or third‑party apps and all free of cost! It includes OAuth2 authentication, token management, and video upload functionality.

## Features
- Upload videos to YouTube from MCP Client(Claude/Cursor/VS Code)
- OAuth2 authentication flow
- Access token and refresh token management
- Multi Channel Support

## Demo
![output](https://github.com/user-attachments/assets/f8c2c303-ef77-4fa9-99a6-5de7f120ffac)


## Getting Started

Visit the [Releases](https://github.com/anwerj/youtube-uploader-mcp/releases) page and download the appropriate binary for your operating system:

- `youtube-uploader-mcp-linux-amd64`
- `youtube-uploader-mcp-darwin-arm64`
- `youtube-uploader-mcp-windows-amd64.exe`
- etc.

> You can use the latest versioned tag, e.g., `v1.0.0`.

---

### 2. Make it Executable (Linux/macOS)

```bash
chmod +x path/to/youtube-uploader-mcp-<os>-<arch>
```

### 3. Configure MCP (e.g., in Claude Desktop or Cursor)
```json
{
  "mcpServers": {
    "youtube-uploader-mcp": {
      "command": "/absolute/path/to/youtube-uploader-mcp-<os>-<arch>",
      "args": [
        "-client_secret_file",
        "/absolute/path/to/client_secret.json(See Below)"
      ]
    }
  }
}
```
### 4. Set Up Google OAuth 2.0
To upload to YouTube, you must configure OAuth and get a client_secret.json file from the Google Developer Console.

➡️ Follow the guide in [youtube_oauth2_setup.md](./youtube_oauth2_setup.md) for a step-by-step walkthrough.

### Usage

- `main.go`: Entry point for the CLI
- `main/`: Additional main package files
- `youtube/`: YouTube API integration (OAuth, video upload, config)
- `tool/`: Command-line tools for authentication, token, and upload
- `hook/`, `logn/`: Supporting packages
