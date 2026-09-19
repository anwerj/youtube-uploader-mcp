package yum

// Version is the MCP server version returned in initialize (serverInfo.version).
// Release builds set this via -ldflags; default is used for plain `go test` / `go run`.
var Version = "0.1.3"
