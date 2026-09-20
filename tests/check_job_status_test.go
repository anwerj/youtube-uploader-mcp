package tests

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *YumSuite) TestCheckJobStatusRequiresJobKey() {
	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name":      "check_job_status",
			"arguments": mcp.Params{},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)

	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Contains(text.Text, "job_key is required")
}

func (s *YumSuite) TestCheckJobStatusNotPresentInRegistry() {
	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "check_job_status",
			"arguments": mcp.Params{
				"job_key": "unknown-job-key",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)

	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Contains(text.Text, "not present in registry")
}
