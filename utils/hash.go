package utils

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"os"
)

// CalcFileSHA1 calculates SHA1 of a file using streaming (low memory).
func CalcFileSHA1(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	return CalcReaderSHA1(f)
}

// CalcReaderSHA1 calculates SHA1 of a reader using streaming.
func CalcReaderSHA1(reader io.Reader) (string, error) {
	h := sha1.New()
	if _, err := io.Copy(h, reader); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
