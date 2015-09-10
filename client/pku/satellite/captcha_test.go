package pku_test

import (
	"image"
	_ "image/jpeg"
	"os"
	"testing"

	"github.com/applepi-icpc/icarus/client/pku/satellite"
)

func TestCaptcha(t *testing.T) {
	f, err := os.Open("testdata/test.jpg")
	if err != nil {
		t.Fatalf("error opening file: %s", err.Error())
	}
	defer f.Close()
	im, _, err := image.Decode(f)
	if err != nil {
		t.Fatalf("error decoding file: %s", err.Error())
	}
	s := pku.Identify(im)
	t.Logf("identified as %s", s)
	if s != "5JFU" {
		t.Fatalf("wrong answer")
	}
}
