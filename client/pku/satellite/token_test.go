package pku_test

import (
	"testing"

	"github.com/applepi-icpc/icarus/client/pku/satellite"
)

func ensureToken(a string, b string, c int, t *testing.T) {
	token := pku.EnToken(a, b, c)
	t.Logf("Token = %s", token)
	aa, bb, cc, err := pku.DeToken(token)
	if err != nil {
		t.Fatalf("failed to decode token: %s", err.Error())
	}
	t.Logf("Decoded = %s, %s, %d", aa, bb, cc)
	if a != aa || b != bb || c != cc {
		t.Fatalf("wrong answer")
	}
}

func TestToken(t *testing.T) {
	ensureToken("13", "BKPC001271892", 17, t)
	ensureToken("13", "BKPC$", 17, t)
	ensureToken("13", "BKPC$$", 17, t)
	ensureToken("13", "BKPC#", 17, t)
	ensureToken("13", "BKPC$#", 17, t)
	ensureToken("13", "BKPC$$#", 17, t)
}
