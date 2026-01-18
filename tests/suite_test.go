package tests

import (
	"context"
	"testing"

	"github.com/anwerj/youtube-uploader-mcp/yum"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/stretchr/testify/suite"
	"github.com/therewardstore/httpmatter"
)

type YumSuite struct {
	suite.Suite
	servers map[string]*server.MCPServer
	mock    *httpmatter.HTTP
}

func (s *YumSuite) SetupSuite() {
	httpmatter.Init(&httpmatter.Config{
		BaseDir: "./data/",
	})
	s.servers = make(map[string]*server.MCPServer)
	s.SetServer("default", "./data/client_secret_1.json", "./data/tmp")
}

func (s *YumSuite) BeforeTest(_, _ string) {
	s.mock = httpmatter.NewHTTP(s.T(), "outgoing")
}

func (s *YumSuite) AfterTest(_, _ string) {
	s.mock.Destroy()
}

func (s *YumSuite) Ctx() context.Context {
	return s.T().Context()
}

func (s *YumSuite) SetServer(name, clientSecretFile, workingDir string) {
	server, err := yum.Build(context.Background(), clientSecretFile, workingDir)
	if err != nil {
		s.T().Error("Failed to build server: " + err.Error())
	}
	s.servers[name] = server
}

func (s *YumSuite) OnServer(name string) *Msg {
	return &Msg{
		on: s.servers[name],
		JSONRPCRequest: mcp.JSONRPCRequest{
			JSONRPC: mcp.JSONRPC_VERSION,
			ID:      mcp.NewRequestId(1),
		},
	}
}

func TestYumSuite(t *testing.T) {
	suite.Run(t, new(YumSuite))
}
