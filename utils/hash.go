package utils

import (
	"crypto/md5"
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

	h := sha1.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalcFileMD5 calculates MD5 of a file.
func CalcFileMD5(filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := md5.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalcReaderMD5 calculates MD5 from any io.Reader. Does NOT seek back.
func CalcReaderMD5(r io.Reader) (string, error) {
	h := md5.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

// CalcReaderSHA1 calculates SHA1 from any io.Reader. Does NOT seek back.
func CalcReaderSHA1(r io.Reader) (string, error) {
	h := sha1.New()
	if _, err := io.Copy(h, r); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
