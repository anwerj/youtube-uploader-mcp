package tests

import (
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/therewardstore/httpmatter"
)

const (
	configuredRedirectURI = "http://localhost:8080/oauth2callback"
	customRedirectURI     = "http://localhost:9191"
)

func (s *YumSuite) TestAuthenticate() {
	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "authenticate",
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Equal(text.Text, expectedAuthURL(configuredRedirectURI))
	s.assertNoRedirectFile("./data/tmp")
}

func (s *YumSuite) TestAuthenticateCustomRedirectURI() {
	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "authenticate",
			"arguments": mcp.Params{
				"redirect_uri": customRedirectURI,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Equal(text.Text, expectedAuthURL(customRedirectURI))
	s.assertNoRedirectFile("./data/tmp")
}

func (s *YumSuite) TestAccessToken() {
	s.mock.Add("access_token_request", "access_token_response").Respond(
		httpmatter.RequestResponse(s.assertTokenRedirect(configuredRedirectURI)))
	s.mock.Add("get_channel_request", "get_channel_response").Respond(nil)
	s.mock.Init()

	text, err := s.callAccessToken("default", "")
	s.NoError(err)

	s.Contains(text.Text, `{"id":"mock-channel-id","name":"mock-channel-title","customer_url":"","token":{"access_token":"***","token_type":"Bearer","refresh_token":"***","expiry":"`)
	s.assertNoRedirectFile("./data/tmp")
}

func (s *YumSuite) TestAccessTokenCustomRedirectURI() {
	s.mock.Add("access_token_request", "access_token_response").Respond(
		httpmatter.RequestResponse(s.assertTokenRedirect(customRedirectURI)))
	s.mock.Add("get_channel_request", "get_channel_response").Respond(nil)
	s.mock.Init()

	text, err := s.callAccessToken("default", customRedirectURI)
	s.NoError(err)

	s.Contains(text.Text, `{"id":"mock-channel-id","name":"mock-channel-title","customer_url":"","token":{"access_token":"***","token_type":"Bearer","refresh_token":"***","expiry":"`)
	s.assertNoRedirectFile("./data/tmp")
}

func (s *YumSuite) TestAccessTokenCustomRedirectOnNewCore() {
	dir := s.T().TempDir()
	s.SetServer("fresh-redirect", "./data/client_secret_1.json", dir)

	s.mock.Add("access_token_request", "access_token_response").Respond(
		httpmatter.RequestResponse(s.assertTokenRedirect(customRedirectURI)))
	s.mock.Add("get_channel_request", "get_channel_response").Respond(nil)
	s.mock.Add("access_token_request", "access_token_response").Respond(
		httpmatter.RequestResponse(s.assertTokenRedirect(configuredRedirectURI)))
	s.mock.Add("get_channel_request", "get_channel_response").Respond(nil)
	s.mock.Init()

	text, err := s.callAccessToken("fresh-redirect", customRedirectURI)
	s.NoError(err)
	s.Contains(text.Text, `"id":"mock-channel-id"`)

	text, err = s.callAccessToken("fresh-redirect", "")
	s.NoError(err)
	s.Contains(text.Text, `"id":"mock-channel-id"`)
	s.assertNoRedirectFile(dir)
}

func (s *YumSuite) callAccessToken(serverName, redirectURI string) (mcp.TextContent, error) {
	args := mcp.Params{"code": "mock-code"}
	if redirectURI != "" {
		args["redirect_uri"] = redirectURI
	}
	return s.OnServer(serverName).
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name":      "accesstoken",
			"arguments": args,
		}).
		ExpectSuccessText(s.Ctx())
}

func (s *YumSuite) assertTokenRedirect(want string) func(*http.Request) int {
	return func(req *http.Request) int {
		s.Equal(http.MethodPost, req.Method)
		body, err := io.ReadAll(req.Body)
		s.NoError(err)
		values, err := url.ParseQuery(string(body))
		s.NoError(err)
		s.Equal("authorization_code", values.Get("grant_type"))
		s.Equal(want, values.Get("redirect_uri"))
		return 0
	}
}

func (s *YumSuite) assertNoRedirectFile(dir string) {
	_, err := os.Stat(filepath.Join(dir, ".youtube_uploader_redirect_uri"))
	s.ErrorIs(err, os.ErrNotExist)
}

func expectedAuthURL(redirectURI string) string {
	return "https://accounts.google.com/o/oauth2/auth?" +
		"access_type=offline&client_id=1234567890-mockclientid.apps.googleusercontent.com&prompt=consent&" +
		"redirect_uri=" + url.QueryEscape(redirectURI) + "&response_type=code&" +
		"scope=https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fyoutube.upload+" +
		"https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fyoutube.force-ssl+" +
		"https%3A%2F%2Fwww.googleapis.com%2Fauth%2Fyoutube.readonly&state=state-token"
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
