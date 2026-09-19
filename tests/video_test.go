package tests

import (
	"encoding/json"
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

func (s *YumSuite) TestUpdateVideoSchedule() {
	scheduleAssert := func(req *http.Request) int {
		s.Equal("PUT", req.Method)
		s.Equal("https://youtube.googleapis.com/youtube/v3/videos?alt=json&part=status&prettyPrint=false", req.URL.String())
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))

		body, err := io.ReadAll(req.Body)
		s.NoError(err)
		// Every mutable status field must be echoed back explicitly. YouTube
		// deletes any mutable property the request omits, so the false booleans
		// have to appear on the wire rather than being dropped by omitempty.
		expectedJSON := `{
			"id": "video_id_12345",
			"status": {
				"privacyStatus": "private",
				"publishAt": "2026-09-20T18:00:00+05:30",
				"license": "creativeCommon",
				"embeddable": false,
				"publicStatsViewable": false,
				"selfDeclaredMadeForKids": false,
				"containsSyntheticMedia": false
			}
		}`
		s.JSONEq(expectedJSON, string(body))
		// Read-only fields must never be sent back.
		s.NotContains(string(body), "uploadStatus")
		s.NotContains(string(body), "madeForKids\":")
		return 0
	}

	s.mock.Add("get_video_request", "get_video_response").Respond(nil)
	s.mock.Add("schedule_video_request", "schedule_video_response").Respond(
		httpmatter.RequestResponse(scheduleAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
				"publish_at": "2026-09-20T18:00:00+05:30",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"video_id":"video_id_12345"`)
	s.Contains(text.Text, `"schedule_status":"success"`)
}

func (s *YumSuite) TestUpdateVideoMadeForKids() {
	madeForKidsAssert := func(req *http.Request) int {
		s.Equal("PUT", req.Method)
		s.Equal("https://youtube.googleapis.com/youtube/v3/videos?alt=json&part=status&prettyPrint=false", req.URL.String())

		body, err := io.ReadAll(req.Body)
		s.NoError(err)
		expectedJSON := `{
			"id": "video_id_12345",
			"status": {
				"privacyStatus": "private",
				"license": "creativeCommon",
				"embeddable": false,
				"publicStatsViewable": false,
				"selfDeclaredMadeForKids": true,
				"containsSyntheticMedia": false
			}
		}`
		s.JSONEq(expectedJSON, string(body))
		s.NotContains(string(body), "publishAt")
		return 0
	}

	s.mock.Add("get_video_request", "get_video_response").Respond(nil)
	s.mock.Add("schedule_video_request", "schedule_video_response").Respond(
		httpmatter.RequestResponse(madeForKidsAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id":     "mock-channel-id",
				"video_id":       "video_id_12345",
				"made_for_kids":  true,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"made_for_kids_status":"success"`)
}

func (s *YumSuite) TestUpdateVideoScheduleAndMadeForKids() {
	getCount := 0
	putCount := 0

	getAssert := func(req *http.Request) int {
		getCount++
		s.Equal("GET", req.Method)
		s.Contains(req.URL.String(), "youtube/v3/videos")
		return 0
	}
	combinedAssert := func(req *http.Request) int {
		putCount++
		s.Equal("PUT", req.Method)

		body, err := io.ReadAll(req.Body)
		s.NoError(err)
		expectedJSON := `{
			"id": "video_id_12345",
			"status": {
				"privacyStatus": "private",
				"publishAt": "2026-09-20T18:00:00+05:30",
				"license": "creativeCommon",
				"embeddable": false,
				"publicStatsViewable": false,
				"selfDeclaredMadeForKids": true,
				"containsSyntheticMedia": false
			}
		}`
		s.JSONEq(expectedJSON, string(body))
		return 0
	}

	s.mock.Add("get_video_request", "get_video_response").Respond(httpmatter.RequestResponse(getAssert))
	s.mock.Add("schedule_video_request", "schedule_video_response").Respond(
		httpmatter.RequestResponse(combinedAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id":     "mock-channel-id",
				"video_id":       "video_id_12345",
				"publish_at":     "2026-09-20T18:00:00+05:30",
				"made_for_kids":  true,
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Equal(1, getCount)
	s.Equal(1, putCount)
	s.Contains(text.Text, `"schedule_status":"success"`)
	s.Contains(text.Text, `"made_for_kids_status":"success"`)
}

func (s *YumSuite) TestUpdateVideoScheduleAbortsBeforeMutations() {
	s.mock.Add("get_video_request", "get_video_public_response").Respond(nil)
	s.mock.Init()

	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id":  "mock-channel-id",
				"video_id":    "video_id_12345",
				"playlist_id": "playlist_id_12345",
				"publish_at":  "2099-01-01T12:00:00Z",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)

	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Contains(text.Text, "already public")
	s.NotContains(text.Text, `"playlist_status"`)
}

func (s *YumSuite) TestUpdateVideoScheduleInvalidPublishAt() {
	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
				"publish_at": "not-a-date",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)
	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Contains(text.Text, "RFC3339")
}

func (s *YumSuite) TestUpdateVideoSchedulePastPublishAt() {
	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "update_video",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"video_id":   "video_id_12345",
				"publish_at": "2020-01-01T00:00:00Z",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)
	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Contains(text.Text, "future timestamp")
}

func (s *YumSuite) TestForbiddenListVideos() {
	playlistAssert := func(req *http.Request) int {
		s.Equal("GET", req.Method)
		s.Equal("https://youtube.googleapis.com/youtube/v3/playlistItems?alt=json&maxResults=50&part=snippet&part=contentDetails&part=status&playlistId=UUmock-channel-id&prettyPrint=false", req.URL.String())
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))
		return 0
	}

	s.mock.Add("uploads_playlist_request", "uploads_playlist_response").Respond(nil)
	s.mock.Add("list_videos_request", "forbidden_response").Respond(
		httpmatter.RequestResponse(playlistAssert))
	s.mock.Init()

	result, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "list_videos",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
			},
		}).
		Call(s.Ctx(), 1, true)
	s.NoError(err)

	text, ok := result.Content[0].(mcp.TextContent)
	s.True(ok)
	s.Contains(text.Text, "authenticate")
}

func (s *YumSuite) TestListVideos() {
	playlistAssert := func(req *http.Request) int {
		s.Equal("GET", req.Method)
		s.Equal("https://youtube.googleapis.com/youtube/v3/playlistItems?alt=json&maxResults=50&part=snippet&part=contentDetails&part=status&playlistId=UUmock-channel-id&prettyPrint=false", req.URL.String())
		s.Equal("Bearer mock-access-token", req.Header.Get("Authorization"))
		return 0
	}

	s.mock.Add("uploads_playlist_request", "uploads_playlist_response").Respond(nil)
	s.mock.Add("list_videos_request", "list_videos_response").Respond(
		httpmatter.RequestResponse(playlistAssert))
	s.mock.Init()

	text, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "list_videos",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	s.Contains(text.Text, `"video_private_1"`)
	s.Contains(text.Text, `"privacy_status":"private"`)
	s.True(strings.Index(text.Text, `"video_public_1"`) < strings.Index(text.Text, `"video_private_1"`))

	ascText, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "list_videos",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"direction":  "asc",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)
	s.True(strings.Index(ascText.Text, `"video_private_1"`) < strings.Index(ascText.Text, `"video_public_1"`))

	queryText, err := s.OnServer("default").
		WithMethod("tools/call").
		WithParams(mcp.Params{
			"name": "list_videos",
			"arguments": mcp.Params{
				"channel_id": "mock-channel-id",
				"query":      "private draft",
			},
		}).
		ExpectSuccessText(s.Ctx())
	s.NoError(err)

	var queryResult struct {
		Videos       []map[string]string `json:"videos"`
		TotalMatched int                 `json:"total_matched"`
	}
	s.NoError(json.Unmarshal([]byte(queryText.Text), &queryResult))
	s.Equal(1, queryResult.TotalMatched)
	s.Equal("video_private_1", queryResult.Videos[0]["id"])
}
