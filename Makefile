BINARY_NAME=youtube-uploader-mcp
LDFLAGS=-s -w

.PHONY: all clean linux-amd64 darwin-amd64 darwin-arm64 windows-amd64

all: linux-amd64 darwin-amd64 darwin-arm64 windows-amd64

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
