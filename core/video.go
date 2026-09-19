package core

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/oauth2"
	"google.golang.org/api/youtube/v3"
)

const maxUploadsPages = 100

type Video struct {
	ID            string   `json:"id"`
	Path          string   `json:"path"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	Tags          []string `json:"tags"`
	CategoryID    string   `json:"category_id"`
	Language      string   `json:"language"`
	PrivacyStatus string   `json:"privacy_status"`
	MadeForKids   bool     `json:"made_for_kids"`
	PublishAt     string   `json:"publish_at,omitempty"`
}

type VideoSummary struct {
	ID            string `json:"id"`
	Title         string `json:"title"`
	Description   string `json:"description"`
	PublishedAt   string `json:"published_at"`
	PrivacyStatus string `json:"privacy_status"`
	ThumbnailURL  string `json:"thumbnail_url"`
}

type ListVideosOptions struct {
	Query         string
	PrivacyStatus string
	OrderBy       string
	Direction     string
	Max           int
	Offset        int
}

type ListVideosResult struct {
	Videos       []VideoSummary `json:"videos"`
	TotalMatched int            `json:"total_matched"`
	Offset       int            `json:"offset"`
	Max          int            `json:"max"`
	Truncated    bool           `json:"truncated"`
}

func (v *Video) toUpload() (*youtube.Video, error) {
	privacy := "private"
	if v.PrivacyStatus != "" {
		privacy = v.PrivacyStatus
	}

	// If PublishAt is set, privacy status must be private
	if v.PublishAt != "" {
		privacy = "private"
	}

	snippet := &youtube.VideoSnippet{
		Title:       v.Title,
		Description: v.Description,
		Tags:        v.Tags,
		CategoryId:  v.CategoryID,
	}
	if v.Language != "" {
		snippet.DefaultLanguage = v.Language
		snippet.DefaultAudioLanguage = v.Language
	}

	upload := &youtube.Video{
		Snippet: snippet,
		Status: &youtube.VideoStatus{
			PrivacyStatus:           privacy,
			SelfDeclaredMadeForKids: v.MadeForKids,
			PublishAt:               v.PublishAt,
			ForceSendFields: []string{
				"SelfDeclaredMadeForKids",
			},
		},
	}

	return upload, nil
}

func (c *Core) UploadVideo(ctx context.Context, video *Video, token *oauth2.Token) (string, error) {

	// First open the video file and verify it exists
	file, err := os.Open(video.Path)
	if err != nil {
		return "", fmt.Errorf("failed to open video file %s: %w", video.Path, err)
	}
	defer file.Close()

	service, err := c.Service(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube service: %w", err)
	}

	upload, err := video.toUpload()
	if err != nil {
		return "", fmt.Errorf("failed to convert video to upload format: %w", err)
	}

	call := service.Videos.Insert([]string{"snippet", "status"}, upload)
	resp, err := call.Media(file).Do()
	if err != nil {
		return "", fmt.Errorf("failed to upload video: %w", err)
	}

	return resp.Id, nil
}

// GetVideo fetches a video's snippet, status, and contentDetails in a single
// videos.list call. It performs no mutation.
func (c *Core) GetVideo(ctx context.Context, videoID string, token *oauth2.Token) (*youtube.Video, error) {
	if videoID == "" {
		return nil, fmt.Errorf("video ID must be provided")
	}
	if token == nil {
		return nil, fmt.Errorf("token must be provided")
	}

	service, err := c.Service(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	resp, err := service.Videos.List([]string{"snippet", "status", "contentDetails"}).Id(videoID).Do()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch video: %w", err)
	}
	if len(resp.Items) == 0 {
		return nil, fmt.Errorf("video %s not found", videoID)
	}

	return resp.Items[0], nil
}

// StatusUpdateOptions describes the mutable status fields a caller wants to
// change. A nil pointer means "leave this field exactly as it currently is".
type StatusUpdateOptions struct {
	PublishAt   *string
	MadeForKids *bool
}

// BuildStatusUpdate builds the full status object to send to videos.update,
// applying whichever fields in opts are non-nil on top of current. It
// performs no API call. videos.update replaces every mutable property in the
// status part, and any property the request omits is deleted by YouTube, so
// every existing mutable value from current is carried over and the booleans
// are forced onto the wire, since Go's omitempty would otherwise drop the
// false ones and let YouTube reset them to defaults.
func BuildStatusUpdate(current *youtube.VideoStatus, opts StatusUpdateOptions) (*youtube.VideoStatus, error) {
	if current == nil {
		return nil, fmt.Errorf("current video status must be provided")
	}

	privacyStatus := current.PrivacyStatus
	publishAt := current.PublishAt
	selfDeclaredMadeForKids := current.SelfDeclaredMadeForKids

	if opts.PublishAt != nil {
		if current.PrivacyStatus == "public" {
			return nil, fmt.Errorf("video is already public; scheduling would unpublish it")
		}
		privacyStatus = "private"
		publishAt = *opts.PublishAt
	}
	if opts.MadeForKids != nil {
		selfDeclaredMadeForKids = *opts.MadeForKids
	}

	return &youtube.VideoStatus{
		PrivacyStatus:           privacyStatus,
		PublishAt:               publishAt,
		License:                 current.License,
		Embeddable:              current.Embeddable,
		PublicStatsViewable:     current.PublicStatsViewable,
		SelfDeclaredMadeForKids: selfDeclaredMadeForKids,
		ContainsSyntheticMedia:  current.ContainsSyntheticMedia,
		ForceSendFields: []string{
			"Embeddable",
			"PublicStatsViewable",
			"SelfDeclaredMadeForKids",
			"ContainsSyntheticMedia",
		},
	}, nil
}

// UpdateVideoStatus writes a fully-built status object to an existing video.
// It performs exactly one API call (videos.update) and does not fetch or
// mutate the status itself; the caller is responsible for building a status
// that carries forward any mutable fields it wants to preserve.
func (c *Core) UpdateVideoStatus(ctx context.Context, videoID string, status *youtube.VideoStatus, token *oauth2.Token) error {
	if videoID == "" {
		return fmt.Errorf("video ID must be provided")
	}
	if status == nil {
		return fmt.Errorf("status must be provided")
	}
	if token == nil {
		return fmt.Errorf("token must be provided")
	}

	service, err := c.Service(ctx, token)
	if err != nil {
		return fmt.Errorf("failed to create YouTube service: %w", err)
	}

	update := &youtube.Video{Id: videoID, Status: status}
	if _, err := service.Videos.Update([]string{"status"}, update).Do(); err != nil {
		return fmt.Errorf("failed to update video status: %w", err)
	}

	return nil
}

func (c *Core) InvalidateVideoCatalog(channelID string) {
	if channelID == "" {
		return
	}
	c.catalogMu.Lock()
	if c.catalog != nil {
		delete(c.catalog, channelID)
	}
	c.catalogMu.Unlock()
}

func (c *Core) ListVideos(ctx context.Context, channel *Channel, opts ListVideosOptions) (*ListVideosResult, error) {
	if channel == nil || channel.Token == nil {
		return nil, fmt.Errorf("channel or token is nil")
	}

	catalog, truncated, err := c.loadVideoCatalog(ctx, channel)
	if err != nil {
		return nil, err
	}

	filtered := filterVideoSummaries(catalog, opts.Query, opts.PrivacyStatus)
	sortVideoSummaries(filtered, opts.OrderBy, opts.Direction)

	total := len(filtered)
	start := opts.Offset
	if start > total {
		start = total
	}
	end := start + opts.Max
	if end > total {
		end = total
	}

	return &ListVideosResult{
		Videos:       filtered[start:end],
		TotalMatched: total,
		Offset:       opts.Offset,
		Max:          opts.Max,
		Truncated:    truncated,
	}, nil
}

func (c *Core) loadVideoCatalog(ctx context.Context, channel *Channel) ([]VideoSummary, bool, error) {
	now := time.Now()
	c.catalogMu.RLock()
	if c.catalog != nil {
		entry, ok := c.catalog[channel.ID]
		if ok && now.Sub(entry.fetchedAt) < videoCatalogTTL {
			videos := entry.videos
			truncated := entry.truncated
			c.catalogMu.RUnlock()
			return videos, truncated, nil
		}
	}
	c.catalogMu.RUnlock()

	playlistID, err := c.uploadsPlaylistID(ctx, channel.Token)
	if err != nil {
		return nil, false, err
	}

	videos, truncated, err := c.fetchUploads(ctx, channel.Token, playlistID)
	if err != nil {
		return nil, false, err
	}

	c.catalogMu.Lock()
	if c.catalog == nil {
		c.catalog = make(map[string]videoCatalogEntry)
	}
	c.catalog[channel.ID] = videoCatalogEntry{
		uploadsPlaylistID: playlistID,
		videos:            videos,
		fetchedAt:         now,
		truncated:         truncated,
	}
	c.catalogMu.Unlock()

	return videos, truncated, nil
}

func (c *Core) uploadsPlaylistID(ctx context.Context, token *oauth2.Token) (string, error) {
	service, err := c.Service(ctx, token)
	if err != nil {
		return "", fmt.Errorf("failed to create YouTube service: %w", err)
	}

	resp, err := service.Channels.List([]string{"contentDetails"}).Mine(true).Do()
	if err != nil {
		return "", err
	}
	if len(resp.Items) == 0 {
		return "", fmt.Errorf("no channels found for token")
	}
	playlistID := resp.Items[0].ContentDetails.RelatedPlaylists.Uploads
	if playlistID == "" {
		return "", fmt.Errorf("uploads playlist not found for channel")
	}
	return playlistID, nil
}

func (c *Core) fetchUploads(ctx context.Context, token *oauth2.Token, playlistID string) ([]VideoSummary, bool, error) {
	service, err := c.Service(ctx, token)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create YouTube service: %w", err)
	}

	var all []VideoSummary
	truncated := false
	pageToken := ""

	for page := 0; page < maxUploadsPages; page++ {
		call := service.PlaylistItems.List([]string{"snippet", "contentDetails", "status"}).
			PlaylistId(playlistID).
			MaxResults(50)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}

		resp, err := call.Do()
		if err != nil {
			return nil, false, err
		}

		for _, item := range resp.Items {
			summary, ok := playlistItemToSummary(item)
			if ok {
				all = append(all, summary)
			}
		}

		pageToken = resp.NextPageToken
		if pageToken == "" {
			return all, truncated, nil
		}
	}

	truncated = true
	return all, truncated, nil
}

func playlistItemToSummary(item *youtube.PlaylistItem) (VideoSummary, bool) {
	if item == nil {
		return VideoSummary{}, false
	}

	videoID := ""
	if item.ContentDetails != nil {
		videoID = item.ContentDetails.VideoId
	}
	if videoID == "" && item.Snippet != nil && item.Snippet.ResourceId != nil {
		videoID = item.Snippet.ResourceId.VideoId
	}
	if videoID == "" {
		return VideoSummary{}, false
	}

	title := ""
	description := ""
	publishedAt := ""
	thumbnailURL := ""
	if item.Snippet != nil {
		title = item.Snippet.Title
		description = item.Snippet.Description
		publishedAt = item.Snippet.PublishedAt
		thumbnailURL = pickThumbnailURL(item.Snippet.Thumbnails)
	}
	if item.ContentDetails != nil && item.ContentDetails.VideoPublishedAt != "" {
		publishedAt = item.ContentDetails.VideoPublishedAt
	}

	privacy := ""
	if item.Status != nil {
		privacy = item.Status.PrivacyStatus
	}

	return VideoSummary{
		ID:            videoID,
		Title:         title,
		Description:   description,
		PublishedAt:   publishedAt,
		PrivacyStatus: privacy,
		ThumbnailURL:  thumbnailURL,
	}, true
}

func pickThumbnailURL(thumbs *youtube.ThumbnailDetails) string {
	if thumbs == nil {
		return ""
	}
	if thumbs.Medium != nil && thumbs.Medium.Url != "" {
		return thumbs.Medium.Url
	}
	if thumbs.Default != nil && thumbs.Default.Url != "" {
		return thumbs.Default.Url
	}
	if thumbs.High != nil && thumbs.High.Url != "" {
		return thumbs.High.Url
	}
	return ""
}

func filterVideoSummaries(videos []VideoSummary, query, privacyStatus string) []VideoSummary {
	query = strings.TrimSpace(strings.ToLower(query))
	privacyStatus = strings.TrimSpace(strings.ToLower(privacyStatus))

	if query == "" && privacyStatus == "" {
		out := make([]VideoSummary, len(videos))
		copy(out, videos)
		return out
	}

	var out []VideoSummary
	for _, v := range videos {
		if privacyStatus != "" && strings.ToLower(v.PrivacyStatus) != privacyStatus {
			continue
		}
		if query != "" {
			title := strings.ToLower(v.Title)
			desc := strings.ToLower(v.Description)
			if !strings.Contains(title, query) && !strings.Contains(desc, query) {
				continue
			}
		}
		out = append(out, v)
	}
	return out
}

func sortVideoSummaries(videos []VideoSummary, orderBy, direction string) {
	if orderBy == "" {
		orderBy = "date"
	}
	if direction == "" {
		direction = "desc"
	}
	if orderBy != "date" {
		return
	}

	desc := direction != "asc"
	sort.SliceStable(videos, func(i, j int) bool {
		ti := videos[i].PublishedAt
		tj := videos[j].PublishedAt
		if ti == tj {
			if desc {
				return videos[i].ID > videos[j].ID
			}
			return videos[i].ID < videos[j].ID
		}
		if desc {
			return ti > tj
		}
		return ti < tj
	})
}
