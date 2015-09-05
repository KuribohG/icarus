package satellite

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

/*
 Server would hold a private key and each satellite would have its corresponding public key.

 1. Satellite generates a random key, then encrypts it with public key and sends it along with the task request. (`TaskRequest`)

 2. Server decrypts the cipher in the task request with its private key, then encrypts the subtask (a json string) with the random key the satellite first generated and sends it back. (`TaskResponse`)

 3. Satellite do the actual work, and sent back the result encrypted in the same cipher. (`WorkResponse`)

 In this process, server is a HTTP server to handle 2 actions of satellite (client):
    1. Request a task.
        Server <- `TaskRequest` (POST)
        `TaskResponse` -> Satellite (Status 200)
        -> Satellite (Status 400, 500)
    2. Report a work result.
        Server <- `WorkResponse` (POST)
        -> Satellite (Statue 200, 400, 500)
*/

// 霧雨　魔理沙

var (
	flagPublicKey = flag.String("pub", "public.pem", "Path of public key")

	genkeyOnce sync.Once
	publicKey  *rsa.PublicKey

	ErrNoPublicKeyFound = errors.New("no public key found")
)

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

// This function could be called any times you want, explicitly or implicitly.
func InitPubkey() {
	genkeyOnce.Do(func() {
		f, err := os.Open(*flagPublicKey)
		checkErr(err)
		defer f.Close()

		data, err := ioutil.ReadAll(f)
		checkErr(err)

		block, _ := pem.Decode(data)
		if block == nil {
			panic(ErrNoPublicKeyFound)
		}

		pub, err := x509.ParsePKIXPublicKey(block.Bytes)
		checkErr(err)
		publicKey = pub.(*rsa.PublicKey)
	})
}

const (
	// AES-256
	KeyLength = 32
)

func GenKey() (orig []byte, cipher string) {
	orig = make([]byte, KeyLength)
	_, err := rand.Read(orig)
	checkErr(err)

	InitPubkey()
	r, err := rsa.EncryptPKCS1v15(rand.Reader, publicKey, orig)
	checkErr(err)

	cipher = base64.StdEncoding.EncodeToString(r)
	return
}

func NakedEncrypt(content string, orig []byte) (string, error) {
	return dispatcher.NakedEncrypt(content, orig)
}

func Decrypt(content string, orig []byte) (string, error) {
	return dispatcher.Decrypt(content, orig)
}
