package generators

import (
	cr "crypto/rand"
	"encoding/base64"
	mr "math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

func RandString(length int8, ss mr.Source) string {
	r := mr.New(ss)

	// Create a byte slice to hold the random string
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[r.Intn(len(charset))]
	}
	return string(b)
}

func RandSecString(length int) (string, error) {
	// Create a byte slice of the desired length
	bytes := make([]byte, length)

	// Read random bytes into the slice
	_, err := cr.Read(bytes)
	if err != nil {
		return "", err
	}

	// Encode the bytes as a base64 string
	return base64.URLEncoding.EncodeToString(bytes), nil
}
