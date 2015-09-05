package server

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"sync"

	"github.com/applepi-icpc/icarus/dispatcher"
)

// アリス・マーガトロイド

var (
	flagPrivateKey = flag.String("priv", "private.pem", "Path of private key")

	genkeyOnce sync.Once
	privateKey *rsa.PrivateKey

	ErrNoPrivateKeyFound = errors.New("no private key found")
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// This function could be called any times you want, explicitly or implicitly.
func InitPrivKey() {
	genkeyOnce.Do(func() {
		f, err := os.Open(*flagPrivateKey)
		checkErr(err)
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		checkErr(err)

		block, _ := pem.Decode(data)
		if block == nil {
			panic(ErrNoPrivateKeyFound)
		}

		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		checkErr(err)
	})
}

const (
	// AES-256
	KeyLength = 32
)

func GetNakedKey(cipher string) ([]byte, error) {
	rawCipher, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return nil, err
	}

	InitPrivKey()
	return rsa.DecryptPKCS1v15(rand.Reader, privateKey, rawCipher)
}

func NakedEncrypt(content string, orig []byte) (string, error) {
	return dispatcher.NakedEncrypt(content, orig)
}

func Decrypt(content string, orig []byte) (string, error) {
	return dispatcher.Decrypt(content, orig)
}

func Encrypt(content string, key string) (string, error) {
	r, err := GetNakedKey(key)
	if err != nil {
		return "", err
	}
	return NakedEncrypt(content, r)
}
