package tests

import (
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/therewardstore/httpmatter"
)

func (s *YumSuite) TestUploadVideo() {
	reqAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Equal("https://youtube.googleapis.com/upload/youtube/v3/videos?alt=json&part=snippet&part=status&prettyPrint=false&uploadType=multipart", req.URL.String())
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		s.NoError(err)
		s.True(strings.HasPrefix(mediaType, "multipart/"))
		boundary := params["boundary"]

		mr := multipart.NewReader(req.Body, boundary)

		// Part 1: JSON Metadata
		p1, err := mr.NextPart()
		s.NoError(err)
		s.Equal("application/json", p1.Header.Get("Content-Type"))

		b, err := io.ReadAll(p1)
		s.NoError(err)

		expectedJSON := `{
			"snippet": {
				"title": "mock-title",
				"description": "mock-description",
				"tags": ["mock-tag1", "mock-tag2"],
				"categoryId": "mock-category-id"
			},
			"status": {
				"privacyStatus": "unlisted"
			}
		}`
		s.JSONEq(expectedJSON, string(b))

		// Part 2: Video Content
		p2, err := mr.NextPart()
		s.NoError(err)
		content, err := io.ReadAll(p2)
		s.NoError(err)
		s.NotEmpty(content)

		return 0
	}

	s.mock.Add("upload_video_request", "upload_video_response").Respond(
		httpmatter.RequestResponse(reqAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "upload_video",
			"arguments": mcp.Params{
				"channel_id":    "mock-channel-id",
				"file_path":     "./data/videos/video_1.mp4",
				"description":   "mock-description",
				"title":         "mock-title",
				"tags":          "mock-tag1,mock-tag2",
				"category_id":   "mock-category-id",
				"status":        "unlisted",
				"made_for_kids": false,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `{"id":"video_id_12345","path":"./data/videos/video_1.mp4","title":"mock-title","description":"mock-description","tags":["mock-tag1","mock-tag2"],"category_id":"mock-category-id","language":"","privacy_status":"unlisted","made_for_kids":false}`)
}

func (s *YumSuite) TestUploadVideoWithLanguage() {
	reqAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Equal("https://youtube.googleapis.com/upload/youtube/v3/videos?alt=json&part=snippet&part=status&prettyPrint=false&uploadType=multipart", req.URL.String())

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		s.NoError(err)
		s.True(strings.HasPrefix(mediaType, "multipart/"))
		boundary := params["boundary"]

		mr := multipart.NewReader(req.Body, boundary)

		// Part 1: JSON Metadata
		p1, err := mr.NextPart()
		s.NoError(err)
		b, err := io.ReadAll(p1)
		s.NoError(err)

		expectedJSON := `{
			"snippet": {
				"title": "mock-title",
				"description": "mock-description",
				"tags": ["mock-tag1", "mock-tag2"],
				"categoryId": "mock-category-id",
				"defaultLanguage": "fr",
				"defaultAudioLanguage": "fr"
			},
			"status": {
				"privacyStatus": "unlisted"
			}
		}`
		s.JSONEq(expectedJSON, string(b))

		return 0
	}

	s.mock.Add("upload_video_request", "upload_video_response").Respond(
		httpmatter.RequestResponse(reqAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "upload_video",
			"arguments": mcp.Params{
				"channel_id":     "mock-channel-id",
				"file_path":      "./data/videos/video_1.mp4",
				"description":    "mock-description",
				"title":          "mock-title",
				"tags":           "mock-tag1,mock-tag2",
				"category_id":    "mock-category-id",
				"video_language": "fr",
				"status":         "unlisted",
				"made_for_kids":  false,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"language":"fr"`)
}

func (s *YumSuite) TestUploadScheduledVideo() {
	reqAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Equal("https://youtube.googleapis.com/upload/youtube/v3/videos?alt=json&part=snippet&part=status&prettyPrint=false&uploadType=multipart", req.URL.String())
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))

		mediaType, params, err := mime.ParseMediaType(req.Header.Get("Content-Type"))
		s.NoError(err)
		s.True(strings.HasPrefix(mediaType, "multipart/"))
		boundary := params["boundary"]

		mr := multipart.NewReader(req.Body, boundary)

		// Part 1: JSON Metadata
		p1, err := mr.NextPart()
		s.NoError(err)
		b, err := io.ReadAll(p1)
		s.NoError(err)

		expectedJSON := `{
			"snippet": {
				"title": "mock-title",
				"description": "mock-description",
				"tags": ["mock-tag1", "mock-tag2"],
				"categoryId": "mock-category-id"
			},
			"status": {
				"privacyStatus": "private",
				"publishAt": "2026-01-20T12:00:00Z"
			}
		}`
		s.JSONEq(expectedJSON, string(b))

		return 0
	}

	s.mock.Add("upload_video_request", "upload_video_response").Respond(
		httpmatter.RequestResponse(reqAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "upload_video",
			"arguments": mcp.Params{
				"channel_id":    "mock-channel-id",
				"file_path":     "./data/videos/video_1.mp4",
				"description":   "mock-description",
				"title":         "mock-title",
				"tags":          "mock-tag1,mock-tag2",
				"category_id":   "mock-category-id",
				"status":        "public",
				"publish_at":    "2026-01-20T12:00:00Z",
				"made_for_kids": false,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"publish_at":"2026-01-20T12:00:00Z"`)
}

func (s *YumSuite) TestUpdateVideoSuccess() {
	playlistAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/playlistItems")
		return 0
	}
	captionsAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/captions")
		return 0
	}
	thumbnailAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/thumbnails/set")
		return 0
	}

	s.mock.Add("add_playlist_request", "add_playlist_response").Respond(
		httpmatter.RequestResponse(playlistAssert))
	s.mock.Add("add_captions_request", "add_captions_response").Respond(
		httpmatter.RequestResponse(captionsAssert))
	s.mock.Add("set_thumbnail_request", "set_thumbnail_response").Respond(
		httpmatter.RequestResponse(thumbnailAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id":         "mock-channel-id",
				"video_id":           "video_id_12345",
				"playlist_id":        "playlist_id_12345",
				"subtitle_path":      "./data/subtitles.srt",
				"subtitle_language":  "en",
				"thumbnail_path":     "./data/thumbnail.png",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"video_id":"video_id_12345"`)
	s.Contains(text.Text, `"playlist_status":"success"`)
	s.Contains(text.Text, `"subtitles_status":"success"`)
	s.Contains(text.Text, `"thumbnail_status":"success"`)
}

func (s *YumSuite) TestUpdateVideoPartialFailure() {
	playlistAssert := func(req *http.Request) int {
		s.Equal("POST", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/playlistItems")
		return 0
	}

	s.mock.Add("add_playlist_request", "add_playlist_response").Respond(
		httpmatter.RequestResponse(playlistAssert))
	s.mock.Add("add_captions_request", "error_response").Respond(nil)
	s.mock.Add("set_thumbnail_request", "error_response").Respond(nil)
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id":         "mock-channel-id",
				"video_id":           "video_id_12345",
				"playlist_id":        "playlist_id_12345",
				"subtitle_path":      "./data/subtitles.srt",
				"subtitle_language":  "en",
				"thumbnail_path":     "./data/thumbnail.png",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"video_id":"video_id_12345"`)
	s.Contains(text.Text, `"playlist_status":"success"`)
	s.Contains(text.Text, `"subtitles_status":"failed"`)
	s.Contains(text.Text, `"thumbnail_status":"failed"`)
	s.Contains(text.Text, `"errors"`)
}
