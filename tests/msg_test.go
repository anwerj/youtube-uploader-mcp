package tests

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Msg struct {
	mcp.JSONRPCRequest
	on *server.MCPServer
}

func (m *Msg) WithID(id int) *Msg {
	m.ID = mcp.NewRequestId(id)
	return m
}

func (m *Msg) WithMethod(method string) *Msg {
	m.Method = method
	return m
}

func (m *Msg) WithParams(params any) *Msg {
	m.Params = params
	return m
}

func (m *Msg) Send(ctx context.Context) (resp mcp.JSONRPCResponse, err error) {
	body, err := json.Marshal(m.JSONRPCRequest)
	if err != nil {
		return
	}

	raw := m.on.HandleMessage(ctx, json.RawMessage(body))
	resp, ok := raw.(mcp.JSONRPCResponse)
	if !ok {
		err = fmt.Errorf("Raw response is not a JSONRPCResponse: %v", raw)
	}
	return
}

func (m *Msg) Call(ctx context.Context, expContent int, expError bool) (result mcp.CallToolResult, err error) {
	resp, err := m.Send(ctx)
	if err != nil {
		return
	}
	result, ok := resp.Result.(mcp.CallToolResult)
	if !ok {
		err = fmt.Errorf("Expected CallToolResult but got %T for %v", resp.Result, m)
		return
	}
	if len(result.Content) != expContent {
		err = fmt.Errorf("Expected %d content, got %d for %v", expContent, len(result.Content), m)
		return
	}
	if result.IsError != expError {
		err = fmt.Errorf("Expected error: %v, got %v for %v", expError, result.Content, m)
		return
	}
	return
}

func (m *Msg) ExpectSuccessText(ctx context.Context) (text mcp.TextContent, err error) {
	result, err := m.Call(ctx, 1, false)
	if err != nil {
		return
	}
	text, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		err = fmt.Errorf("Expected TextContent but got %T for %v", result.Content[0], m)
		return
	}
	return
}
