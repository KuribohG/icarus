package pku_test

import (
	"flag"
	"os"
	"testing"

	"github.com/applepi-icpc/icarus/client/pku/satellite"
)

var (
	flagUserID   = flag.String("id", "12000XXXXX", "PKU User ID")
	flagPassword = flag.String("pw", "XXXXXX", "PKU Password")
)

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestLoginAndCaptcha(t *testing.T) {
	jsid, _, err := pku.LoginHelper([]string{*flagUserID, *flagPassword})
	if err != nil {
		t.Fatalf("Login error: %s", err.Error())
	} else {
		t.Logf("JSESSIONID: %s", jsid)
	}

	err = pku.FetchAndIdentify(jsid)
	if err != nil {
		t.Fatalf("Failed to pass captcha: %s", err.Error())
	}
}
