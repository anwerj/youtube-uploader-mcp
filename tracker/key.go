package tracker

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
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

// BuildKeyFromFields derives a stable job key from a set of named fields.
// Field names are sorted before hashing so the resulting key is deterministic
// regardless of Go's randomized map iteration order.
func BuildKeyFromFields(fields map[string]string) string {
	names := make([]string, 0, len(fields))
	for k := range fields {
		names = append(names, k)
	}
	sort.Strings(names)

	parts := make([]string, 0, len(names)*2)
	for _, k := range names {
		parts = append(parts, k, fields[k])
	}
	return JobKey(parts...)
}
