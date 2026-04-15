package upstream

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
)

// BuildVisitorIDHeader returns the Rita VisitorId header value.
func BuildVisitorIDHeader(visitorID, secret string) (string, error) {
	if strings.TrimSpace(visitorID) == "" {
		return "", errors.New("visitor id is required")
	}

	if strings.TrimSpace(secret) == "" {
		return "", errors.New("secret is required")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	if _, err := mac.Write([]byte(visitorID)); err != nil {
		return "", err
	}

	signature := hex.EncodeToString(mac.Sum(nil))
	return visitorID + ":" + signature, nil
}
