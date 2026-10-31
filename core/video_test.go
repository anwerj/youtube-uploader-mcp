package core

import (
	"encoding/json"
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
	current := baseStatus()
	current.ContainsSyntheticMedia = true
	private := "private"
	got, err := BuildStatusUpdate(current, StatusUpdateOptions{PrivacyStatus: &private})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !got.Embeddable || !got.PublicStatsViewable || got.License != "youtube" {
		t.Fatalf("unspecified fields were not carried over: %+v", got)
	}
	if got.PrivacyStatus != "private" {
		t.Fatalf("privacy: got %q, want private", got.PrivacyStatus)
	}
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if strings.Contains(string(body), "containsSyntheticMedia") {
		t.Fatalf("omitted synthetic media was serialized: %s", body)
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
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if !strings.Contains(string(body), `"containsSyntheticMedia":false`) {
		t.Fatalf("explicit false was dropped from the request: %s", body)
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
	cur := &youtube.VideoStatus{
		PrivacyStatus: "private",
		PublishAt:     when,
	}
	got, err := BuildStatusUpdate(cur, StatusUpdateOptions{PrivacyStatus: strPtr("public")})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.PublishAt != "" {
		t.Fatalf("publish time should be cleared for a public video, got %q", got.PublishAt)
	}
	if got.PrivacyStatus != "public" {
		t.Fatalf("privacy: got %q, want public", got.PrivacyStatus)
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
