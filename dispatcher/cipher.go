package dispatcher

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"encoding/base64"
	"errors"
)

// Common part of satellite and server's cipher.

var (
	ErrValidation = errors.New("validation error")
)

const (
	// SHA-1 Validation
	ShaLength = 8
)

func NakedEncrypt(content string, orig []byte) (string, error) {
	block, err := aes.NewCipher(orig)
	if err != nil {
		return "", err
	}

	iv := make([]byte, block.BlockSize())
	stream := cipher.NewCFBEncrypter(block, iv)

	s := sha1.New()
	// Add validation
	rawContent := append([]byte(content), s.Sum([]byte(content))[:ShaLength]...)

	res := make([]byte, len(rawContent))
	stream.XORKeyStream(res, []byte(rawContent))
	return base64.StdEncoding.EncodeToString(res), nil
}

func Decrypt(content string, orig []byte) (string, error) {
	rawContent, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(orig)
	if err != nil {
		return "", err
	}

	// As cipher is random, IV is out of use.
	iv := make([]byte, block.BlockSize())
	stream := cipher.NewCFBDecrypter(block, iv)

	res := make([]byte, len(rawContent))
	stream.XORKeyStream(res, rawContent)

	l := len(res) - ShaLength
	if l < 0 {
		return "", ErrValidation
	}

	s := sha1.New()
	v := s.Sum(res[:l])[:ShaLength]
	if !bytes.Equal(v, res[l:]) {
		return "", ErrValidation
	}

	return string(res[:l]), nil
}
