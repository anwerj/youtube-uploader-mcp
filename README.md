<p align="center"> <img src="https://github.com/user-attachments/assets/21a9baa2-06e8-4af4-9bcd-1dbce52a2733"/> </p>


# YouTube Uploader MCP

AI‑powered YouTube uploader—no CLI, no YouTube Studio, and no secrets ever shared with LLMs or third‑party apps and all free of cost! It includes OAuth2 authentication, token management, and video upload functionality.

## Features
- Upload videos to YouTube from MCP Client(Claude/Cursor/VS Code)
- OAuth2 authentication flow
- Access token and refresh token management
- Multi Channel Support

## Single Command Installation

### For Mac and Linux
```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/anwerj/youtube-uploader-mcp/master/scripts/install.sh)"
```


### For Windows(Powershell)
```Powershell
Invoke-WebRequest -UseBasicParsing "https://raw.githubusercontent.com/anwerj/youtube-uploader-mcp/master/scripts/install.ps1" -OutFile "$env:TEMP\install.ps1"; PowerShell -NoProfile -ExecutionPolicy Bypass -File "$env:TEMP\install.ps1"
```
### Expected result

This single command will

1. Help in downloading oAuth client secret files, if not downloaded,
2. Download the MCP server,
3. Set minimum required permission to run mcp server,
4. Auto update **Cluade Desktop config** with youtube-uploader-mcp server and
5. At last print exact MCP config for any other clients **VS Code/Cursor/AnythingLLM etc**.

<small>
<pre>
[INFO] Detecting OS and architecture...
[INFO] Detected OS/Arch: darwin/arm64
[INFO] Checking location of client secret file
[INFO] If client_secret not downloaded yet, Please watch https://youtu.be/fcywz5FIUpM for very detailed steps to download
[INFO] Client secret stored at: /Users/itclear/.config/youtube-uploader-mcp/client_secret_10000000000-lvgrjhofjbnd110eouaasaasdasdavapc.apps.googleusercontent.com.json
[INFO] Installing anwerj/youtube-uploader-mcp (latest)
[INFO] Downloading https://github.com/anwerj/youtube-uploader-mcp/releases/latest/download/youtube-uploader-mcp-darwin-arm64
  % Total    % Received % Xferd  Average Speed   Time    Time     Time  Current
                                 Dload  Upload   Total   Spent    Left  Speed
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
  0     0    0     0    0     0      0      0 --:--:-- --:--:-- --:--:--     0
100 14.0M  100 14.0M    0     0  1734k      0  0:00:08  0:00:08 --:--:-- 3107k
[INFO] Installed binary: /Users/itclear/youtube-uploader-mcp/youtube-uploader-mcp-darwin-arm64

Integrate with Claude Desktop now? [y/N]: y
[INFO] Claude config updated at: /Users/itclear/Library/Application Support/Claude/claude_desktop_config.json
[INFO] If Claude is running, restart it to load the new MCP server.

Show VS Code/Cursor MCP config JSON now? [y/N]: y

[INFO] Add the following to your MCP config:
"youtube-uploader-mcp": {
  "command": "/Users/itclear/youtube-uploader-mcp/youtube-uploader-mcp-darwin-arm64",
  "args": [
    "-client_secret_file",
    "/Users/itclear/.config/youtube-uploader-mcp/client_secret_10000000000-lvgrjhofjbnd110eouaasaasdasdavapc.apps.googleusercontent.com.json"
  ],
  "env": {},
  "disabled": false,
  "autoApprove": []
}

[INFO] Setup complete.
</pre>
</small>

## Demo
### Setup and Demo Video
<p align="center"> <a href="https://youtu.be/fcywz5FIUpM" target="_blank"><img src="https://img.youtube.com/vi/fcywz5FIUpM/0.jpg"/></a> </p>

![output](https://github.com/user-attachments/assets/f8c2c303-ef77-4fa9-99a6-5de7f120ffac)

## Manual Installation
Please check [Single Command Installation](#single-command-installation), proceed if you prefer manual installation.

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
- `youtube/`: YouTube API integration (OAuth, video upload, config)
- `tool/`: Command-line tools for authentication, token, and upload
- `hook/`, `logn/`: Supporting packages
