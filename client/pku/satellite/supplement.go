package pku

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/qiniu/log"
)

var (
	reElectedNum = regexp.MustCompile(`(?i)<electedNum>(\d+)</electedNum>`)
	reSupError   = regexp.MustCompile(`<label class='message_error'>(.*?)</label>`)
)

func Refresh(jsessionid string, index string, seq string, ubound int) (bool, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/elective2008/edu/pku/stu/elective/controller/supplement/refreshLimit.do?index=%s&seq=%s", electRoot, index, seq), strings.NewReader(""))
	if err != nil {
		panic(err)
	}
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
