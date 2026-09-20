package tracker

import (
	"crypto/sha256"
	"encoding/hex"
)

// JobKey returns a stable hash for tracker map lookups from one or more parts.
// Parts are joined with a NUL separator so collisions across boundaries are avoided.
func JobKey(parts ...string) string {
	h := sha256.New()
	for i, p := range parts {
		if i > 0 {
			h.Write([]byte{0})
		}
		h.Write([]byte(p))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// UploadJobKey is JobKey(channelID, filePath) for upload_video / verify_upload.
func UploadJobKey(channelID, filePath string) string {
	return JobKey(channelID, filePath)
}
