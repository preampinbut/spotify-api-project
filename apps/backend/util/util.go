// Package util provides utility functions for application.
package util

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomBytes generates a slice of cryptographically secure random bytes of the given length.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// GenerateRandomString generates a cryptographically secure random string of a given length.
func GenerateRandomString(n int) (string, error) {
	bytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil // Encode bytes to a URL-safe base64 string
}
