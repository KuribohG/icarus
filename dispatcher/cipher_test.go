package dispatcher_test

import (
	"encoding/base64"
	"flag"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/applepi-icpc/icarus/dispatcher/satellite"
	"github.com/applepi-icpc/icarus/dispatcher/server"
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestEncryptAndDecrypt(t *testing.T) {
	orig, key := satellite.GenKey()
	plain := "魔理沙の…バカ"
	t.Logf("Key: %s", key)

	encrypted, err := server.Encrypt(plain, key)
	if err != nil {
		t.Fatalf("Error occured when encrypting: %s", err.Error())
	}
	t.Logf("Encrypted: %s", encrypted)

	decrypted, err := satellite.Decrypt(encrypted, orig)
	if err != nil {
		t.Fatalf("Error occured when decrypting: %s", err.Error())
	}
	t.Logf("Decrypted: %s", decrypted)

	if decrypted != plain {
		t.Fatalf("Content corrupted after decrypting.")
	}

	corruptedIDs := []string{
		"",
		encrypted[:8],
		encrypted + encrypted,
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	ts, _ := base64.StdEncoding.DecodeString(encrypted)
	for i := 0; i < 500; i++ {
		ts[r.Intn(len(ts))] ^= 1 << byte(r.Intn(8))
		corrupted_id := base64.StdEncoding.EncodeToString(ts)
		corruptedIDs = append(corruptedIDs, corrupted_id)
	}

	for _, id := range corruptedIDs {
		_, err = satellite.Decrypt(id, orig)
		if err == nil {
			t.Fatalf("A random corrupted cipher can be decrypted.")
		}
	}
}
