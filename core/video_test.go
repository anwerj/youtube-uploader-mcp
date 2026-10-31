package core

import (
	"strings"
	"testing"

	"google.golang.org/api/youtube/v3"
)

func baseStatus() *youtube.VideoStatus {
	return &youtube.VideoStatus{
		PrivacyStatus:           "unlisted",
		License:                 "youtube",
		Embeddable:              true,
		PublicStatsViewable:     true,
		SelfDeclaredMadeForKids: false,
		ContainsSyntheticMedia:  false,
	}
}

func TestBuildStatusUpdateKeepsUnspecifiedFields(t *testing.T) {
	// videos.update replaces the whole status part, so anything the request omits is reset by
	// YouTube. Setting one field must never clear embeddable/publicStatsViewable.
	got, err := BuildStatusUpdate(baseStatus(), StatusUpdateOptions{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Embeddable || !got.PublicStatsViewable || got.License != "youtube" {
		t.Fatalf("unspecified fields were not carried over: %+v", got)
	}
	if got.PrivacyStatus != "unlisted" {
		t.Fatalf("privacy changed unexpectedly: %q", got.PrivacyStatus)
	}
}

func TestBuildStatusUpdatePrivacyStatus(t *testing.T) {
	for _, want := range []string{"public", "unlisted", "private"} {
		got, err := BuildStatusUpdate(baseStatus(), StatusUpdateOptions{PrivacyStatus: &want})
		if err != nil {
			t.Fatalf("privacy %q: unexpected error: %v", want, err)
		}
		if got.PrivacyStatus != want {
			t.Fatalf("privacy: got %q, want %q", got.PrivacyStatus, want)
		}
	}
}

func TestBuildStatusUpdateRejectsUnknownPrivacy(t *testing.T) {
	bad := "everyone"
	_, err := BuildStatusUpdate(baseStatus(), StatusUpdateOptions{PrivacyStatus: &bad})
	if err == nil || !strings.Contains(err.Error(), "privacy_status") {
		t.Fatalf("expected a privacy_status error, got %v", err)
	}
}

func TestBuildStatusUpdateSyntheticMediaFlag(t *testing.T) {
	yes := true
	got, err := BuildStatusUpdate(baseStatus(), StatusUpdateOptions{ContainsSyntheticMedia: &yes})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.ContainsSyntheticMedia {
		t.Fatal("containsSyntheticMedia was not set")
	}
	// It must be force-sent: Go's omitempty would otherwise drop it and YouTube would keep false.
	found := false
	for _, f := range got.ForceSendFields {
		if f == "ContainsSyntheticMedia" {
			found = true
		}
	}
	if !found {
		t.Fatalf("ContainsSyntheticMedia missing from ForceSendFields: %v", got.ForceSendFields)
	}

	no := false
	got, err = BuildStatusUpdate(&youtube.VideoStatus{PrivacyStatus: "public", ContainsSyntheticMedia: true},
		StatusUpdateOptions{ContainsSyntheticMedia: &no})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ContainsSyntheticMedia {
		t.Fatal("containsSyntheticMedia was not cleared")
	}
}

func TestBuildStatusUpdatePublishAtForcesPrivate(t *testing.T) {
	when := "2026-12-01T10:00:00Z"
	got, err := BuildStatusUpdate(baseStatus(), StatusUpdateOptions{PublishAt: &when})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PrivacyStatus != "private" || got.PublishAt != when {
		t.Fatalf("scheduling did not force private: %+v", got)
	}

	public := "public"
	if _, err := BuildStatusUpdate(baseStatus(), StatusUpdateOptions{PublishAt: &when, PrivacyStatus: &public}); err == nil {
		t.Fatal("expected an error when scheduling a non-private video")
	}
}

func TestBuildStatusUpdatePublishAtCappedWhenPublic(t *testing.T) {
	when := "2026-12-01T10:00:00Z"
	cur := baseStatus()
	cur.PublishAt = when
	if _, err := BuildStatusUpdate(cur, StatusUpdateOptions{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Promoting a scheduled private video to public must drop the stale publish time, otherwise
	// YouTube rejects the request.
	got, err := BuildStatusUpdate(cur, StatusUpdateOptions{PrivacyStatus: strPtr("public")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublishAt != "" {
		t.Fatalf("publish time should be cleared for a public video, got %q", got.PublishAt)
	}
}

func TestBuildStatusUpdateMadeForKids(t *testing.T) {
	yes := true
	got, err := BuildStatusUpdate(baseStatus(), StatusUpdateOptions{MadeForKids: &yes})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.SelfDeclaredMadeForKids {
		t.Fatal("selfDeclaredMadeForKids was not set")
	}
}

func TestBuildStatusUpdateRequiresCurrentStatus(t *testing.T) {
	if _, err := BuildStatusUpdate(nil, StatusUpdateOptions{}); err == nil {
		t.Fatal("expected an error for a nil current status")
	}
}

func strPtr(s string) *string { return &s }
