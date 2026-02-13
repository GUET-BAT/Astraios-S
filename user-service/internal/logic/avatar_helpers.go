package logic

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

const (
	avatarObjectPrefix  = "avatars"
	avatarUploadExpiry  = 60 * time.Minute
	avatarDisplayExpiry = 30 * time.Minute
	ossOpTimeout        = 5 * time.Second
)

func buildAvatarObjectKey(userID string) (string, error) {
	timestamp := time.Now().UTC().Format("20060102T150405Z")
	suffix, err := randomHex(8)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s/%s/%s-%s", avatarObjectPrefix, userID, timestamp, suffix), nil
}

func randomHex(bytesLen int) (string, error) {
	if bytesLen <= 0 {
		return "", fmt.Errorf("bytesLen must be positive")
	}
	buf := make([]byte, bytesLen)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

func isHTTPURL(value string) bool {
	v := strings.ToLower(strings.TrimSpace(value))
	return strings.HasPrefix(v, "http://") || strings.HasPrefix(v, "https://")
}
