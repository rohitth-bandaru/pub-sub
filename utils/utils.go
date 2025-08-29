package utils

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// GenerateClientID generates a unique client identifier
func GenerateClientID() string {
	timestamp := time.Now().Format("20060102150405.000000000")
	randomPart := RandomString(8)
	return timestamp + "-" + randomPart
}

// RandomString generates a random string of specified length
func RandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// GenerateRequestID generates a unique request identifier
func GenerateRequestID() string {
	return time.Now().Format("20060102150405.000000000") + "-" + RandomString(6)
}
