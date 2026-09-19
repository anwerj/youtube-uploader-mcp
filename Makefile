BINARY_NAME=youtube-uploader-mcp
VERSION ?= 0.1.3
BUILD_ID ?= $(shell date -u +%Y%m%d%H%M%S)
LDFLAGS=-s -w -X github.com/anwerj/youtube-uploader-mcp/yum.Version=$(VERSION)+$(BUILD_ID)

.PHONY: all clean dev linux-amd64 darwin-amd64 darwin-arm64 windows-amd64

all: linux-amd64 darwin-amd64 darwin-arm64 windows-amd64

# Build only for the machine running make (typical local Cursor MCP workflow).
dev:
	@case "$$(uname -s)/$$(uname -m)" in \
		Darwin/arm64)  $(MAKE) darwin-arm64 ;; \
		Darwin/x86_64) $(MAKE) darwin-amd64 ;; \
		Linux/x86_64|Linux/amd64) $(MAKE) linux-amd64 ;; \
		*) echo "Unsupported platform for dev target: $$(uname -s)/$$(uname -m)"; exit 1 ;; \
	esac

linux-amd64:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-linux-amd64 .

darwin-amd64:
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-amd64 .

darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-darwin-arm64 .

windows-amd64:
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o $(BINARY_NAME)-windows-amd64.exe .

clean:
	rm -f $(BINARY_NAME)-linux-amd64
	rm -f $(BINARY_NAME)-darwin-amd64
	rm -f $(BINARY_NAME)-darwin-arm64
	rm -f $(BINARY_NAME)-windows-amd64.exe
