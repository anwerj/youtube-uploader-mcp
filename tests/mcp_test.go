package tests

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *YumSuite) TestInitialize() {
	resp, err := s.OnServer("default").
		WithMethod("initialize").
		Send(s.Ctx())

	s.NoError(err)

	s.IsType(mcp.InitializeResult{}, resp.Result)
	result := resp.Result.(mcp.InitializeResult)

	s.NotNil(result.ServerInfo)
	s.Equal("Youtube Uploader MCP", result.ServerInfo.Name)
}

func (s *YumSuite) TestListTools() {
	resp, err := s.OnServer("default").
		WithMethod("tools/list").
		Send(s.Ctx())

	s.NoError(err)

	s.IsType(mcp.ListToolsResult{}, resp.Result)
	result := resp.Result.(mcp.ListToolsResult)

	s.Len(result.Tools, 5)
	toolsName := []string{}
	for _, tool := range result.Tools {
		toolsName = append(toolsName, tool.Name)
	}

	s.Equal([]string{"accesstoken", "authenticate", "channels", "refreshtoken", "upload_video"}, toolsName)
}
