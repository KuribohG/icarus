package pku

// #cgo CFLAGS: -O3
// #cgo LDFLAGS: -lm
// #include "nzkcaptcha.h"
import "C"
import (
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	"io/ioutil"
	"net/http"
	"strings"
	"unsafe"

	log "github.com/Sirupsen/logrus"
)

func Identify(im image.Image) string {
	rect := im.Bounds()
	baseW, baseH := rect.Min.X, rect.Min.Y
	w, h := rect.Max.X-baseW, rect.Max.Y-baseH

	imbit := make([]uint32, w)
	for x := 0; x < w; x++ {
		for y := 0; y < h; y++ {
			c, _, _, _ := im.At(x+baseW, y+baseH).RGBA()
			if c > 32767 {
				imbit[x] |= 1 << uint(y)
			}
		}
	}

	res := make([]byte, 5)
	ptr := (*C.char)(unsafe.Pointer(&res[0]))
	C.identify(C.int(h), C.int(w), (*C.int)(unsafe.Pointer(&imbit[0])), ptr)
	return C.GoString(ptr)
}

func FetchAndIdentify(jsessionid string) error {
	for {
		req, err := http.NewRequest("GET", fmt.Sprintf("%s/elective2008/DrawServlet", electRoot), strings.NewReader(""))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s", jsessionid))
		client := &http.Client{}
		res, err := client.Do(req)
		if err != nil {
			log.Warnf("Client PKU: Failed to fetch captcha: %s", err.Error())
			continue
		}
		defer res.Body.Close()
		im, _, err := image.Decode(res.Body)
		if err != nil {
			log.Warnf("Client PKU: Failed to decode captcha image: %s", err.Error())
			continue
		}
		s := Identify(im)

		req, err = http.NewRequest("GET", fmt.Sprintf("%s/elective2008/edu/pku/stu/elective/controller/supplement/validate.do?validCode=%s", electRoot, s), strings.NewReader(""))
		if err != nil {
			panic(err)
		}
		req.Header.Set("Host", "elective.pku.edu.cn")
		req.Header.Set("Connection", "keep-alive")
		req.Header.Set("Pragma", "no-cache")
		req.Header.Set("Cache-Control", "no-cache")
		req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2743.116 Safari/537.36")
		req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
		req.Header.Set("Referer", "http://elective.pku.edu.cn/elective2008/edu/pku/stu/elective/controller/supplement/SupplyCancel.do")
		req.Header.Set("Accept-Encoding", "gzip, deflate, sdch")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.8,en;q=0.6,ja;q=0.4,zh-TW;q=0.2")
		req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s", jsessionid))
		res, err = client.Do(req)
		if err != nil {
			log.Warnf("Client PKU: Failed to submit captcha result: %s", err.Error())
			continue
		}
		defer res.Body.Close()
		rawResBody, err := ioutil.ReadAll(res.Body)
		if err != nil {
			log.Warnf("Client PKU: Failed to read captcha submission result: %s", err.Error())
			continue
		}
		resBody := string(rawResBody)
		if strings.Index(resBody, "<title>") != -1 {
			return errors.New("session expired")
		} else if strings.Index(resBody, "<valid>2</valid>") != -1 {
			return nil
		}
	}
}
