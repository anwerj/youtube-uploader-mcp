package tracker

import "testing"

func TestBuildKeyFromFieldsDeterministicRegardlessOfOrder(t *testing.T) {
	a := BuildKeyFromFields(map[string]string{"channel_id": "c1", "file_path": "f1"})
	b := BuildKeyFromFields(map[string]string{"file_path": "f1", "channel_id": "c1"})
	if a != b {
		t.Fatalf("expected identical key regardless of map iteration order, got %q vs %q", a, b)
	}
}

func TestBuildKeyFromFieldsDiffersOnValue(t *testing.T) {
	a := BuildKeyFromFields(map[string]string{"channel_id": "c1", "file_path": "f1"})
	b := BuildKeyFromFields(map[string]string{"channel_id": "c2", "file_path": "f1"})
	if a == b {
		t.Fatalf("expected different keys for different field values")
	}
}

func TestBuildKeyFromFieldsEmpty(t *testing.T) {
	if BuildKeyFromFields(map[string]string{}) == "" {
		t.Fatalf("expected a stable non-empty key even for empty fields")
	}
}
