package pku

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
)

var (
    reElectedNum = regexp.MustCompile(`(?i){"electedNum":"(\d+)"}`)
	reSupError   = regexp.MustCompile(`<label class='message_error'>(.*?)</label>`)
)

func Refresh(jsessionid string, index string, seq string, ubound int) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/elective2008/edu/pku/stu/elective/controller/supplement/refreshLimit.do?index=%s&seq=%s", electRoot, index, seq), strings.NewReader(""))
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
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Warnf("Client PKU: Failed to refresh: %s", err.Error())
		return false, err
	}
	defer res.Body.Close()
	rawResBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warnf("Client PKU: Failed to read refresh result: %s", err.Error())
		return false, err
	}
	resBody := string(rawResBody)
	match := reElectedNum.FindStringSubmatch(resBody)
	var elected int
	if match == nil {
		err = errors.New("refresh result: wrong format")
	} else {
		elected, err = strconv.Atoi(match[1])
	}
	if err != nil {
		log.Warnf("Client PKU: %s", err.Error())
		return false, err
	}

	if elected >= ubound {
		return false, nil
	} else {
		return true, nil
	}
}

func Supplement(jsessionid string, index string, seq string) (bool, error) {
	err := FetchAndIdentify(jsessionid)
	if err != nil {
		return false, err
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s/elective2008/edu/pku/stu/elective/controller/supplement/electSupplement.do?index=%s&seq=%s", electRoot, index, seq), strings.NewReader(""))
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
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Warnf("Client PKU: Failed to supplement: %s", err.Error())
		return false, err
	}
	defer res.Body.Close()
	rawResBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Warnf("Client PKU: Failed to read supplement result: %s", err.Error())
		return false, err
	}
	resBody := string(rawResBody)
    fmt.Println(resBody)
	if strings.Index(resBody, "success.gif") != -1 {
		// めでたしめでたし
		return true, nil
	} else {
		match := reSupError.FindStringSubmatch(resBody)
		if match == nil {
			err = errors.New("unknown error")
		} else {
			err = errors.New(match[1])
		}
		log.Warnf("Client PKU: Failed to supplement: Server response: %s", err.Error())
		return false, err
	}
}
