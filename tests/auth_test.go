package tests

import (
	"github.com/mark3labs/mcp-go/mcp"
)

func (s *YumSuite) TestAuthenticate() {
	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "authenticate",
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Equal(text.Text, "https://accounts.google.com/o/oauth2/auth?"+
		"access_type=offline&client_id=1234567890-mockclientid.apps.googleusercontent.com&prompt=consent&"+
		"redirect_uri=https%3A%2F%2Flocalhost%3A8080&response_type=code&"+
		"scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fyoutube.upload+"+
		"https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fyoutube.readonly&state=state-token")
}

func (s *YumSuite) TestAccessToken() {
	s.mock.Add("access_token_request", "access_token_response").Respond(nil)
	s.mock.Add("get_channel_request", "get_channel_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "accesstoken",
			"arguments": mcp.Params{
				"code": "mock-code",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `{"id":"mock-channel-id","name":"mock-channel-title","customer_url":"","token":{"access_token":"***","token_type":"Bearer","refresh_token":"***","expiry":"`)
}

func (s *YumSuite) TestGetChannels() {
	// s.mock.Add("get_channel_request", "get_channel_response").Respond(nil)
	// s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "channels",
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `{"id":"mock-channel-id","name":"mock-channel-title","customer_url":"","token":{"access_token":"***","token_type":"Bearer","refresh_token":"***","expiry":"`)
}

func (s *YumSuite) TestRefreshToken() {
	// s.mock.Add("refresh_token_request", "refresh_token_response").Respond(nil)
	// s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "refreshtoken",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `{"id":"mock-channel-id","name":"mock-channel-title","customer_url":"","token":{"access_token":"***","token_type":"Bearer","refresh_token":"***","expiry":"`)
}
