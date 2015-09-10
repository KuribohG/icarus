package pku

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

var (
	regexCookieHdr = regexp.MustCompile("JSESSIONID=([^;]+)")
	regexMsg       = regexp.MustCompile(`"msg":"([^"]*)"`)
	regexToken     = regexp.MustCompile(`"token":"(\w+)"`)
)

// Return jsessionid, supplement page, error
func LoginHelper(data []string) (string, string, error) {
	if len(data) < 2 {
		return "", "", errors.New("datum are not sufficient")
	}
	userid := data[0]
	password := data[1]

	// Step 1: Get IAAA Session
	res, err := http.Get(fmt.Sprintf("%s/iaaa/oauth.jsp?appID=syllabus&appName=学生选课系统&redirectUrl=http://elective.pku.edu.cn:80/elective2008/agent4Iaaa.jsp/../ssoLogin.do", iaaaRoot))
	if err != nil {
		return "", "", errors.New(fmt.Sprintf("failed on step 1 (get iaaa session) #1: %s", err.Error()))
	}
	defer res.Body.Close()
	cookieHdrRaw, ok := res.Header["Set-Cookie"]
	if !ok || len(cookieHdrRaw) < 1 {
		return "", "", errors.New("failed on step 1 (get iaaa session) #2: no set-cookie header")
	}
	cookieHdr := cookieHdrRaw[0]
	match := regexCookieHdr.FindStringSubmatch(cookieHdr)
	if match == nil {
		return "", "", errors.New("failed on step 1 (get iaaa session) #3: no jsessionid set")
	}
	jsessionid := match[1]

	// Step 2: Get IAAA Token
	v := url.Values{}
	v.Set("appid", "syllabus")
	v.Set("userName", userid)
	v.Set("password", password)
	v.Set("redirUrl", "http://elective.pku.edu.cn:80/elective2008/agent4Iaaa.jsp/../ssoLogin.do")
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/iaaa/oauthlogin.do", iaaaRoot), strings.NewReader(v.Encode()))
	req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s", jsessionid))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{}
	res, err = client.Do(req)
	if err != nil {
		return "", "", errors.New(fmt.Sprintf("failed on step 2 (get iaaa token) #1: %s", err.Error()))
	}
	defer res.Body.Close()
	rawResBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", "", errors.New(fmt.Sprintf("failed on step 2 (get iaaa token) #2: %s", err.Error()))
	}
	resBody := string(rawResBody)
	if strings.Index(resBody, `"success":true`) == -1 {
		match = regexMsg.FindStringSubmatch(resBody)
		if match == nil {
			return "", "", errors.New("failed on step 2 (get iaaa token) #3: unknown reason")
		} else {
			return "", "", errors.New(fmt.Sprintf("failed on step 2 (get iaaa token) #3: %s", match[1]))
		}
	}
	match = regexToken.FindStringSubmatch(resBody)
	if match == nil {
		return "", "", errors.New("failed on step 2 (get iaaa token) #4: succeeded, but no token provided")
	}
	token := match[1]

	// Step 3: Get Elective Session
	res, err = http.Get(fmt.Sprintf("%s/elective2008/ssoLogin.do?token=%s", electRoot, token))
	if err != nil {
		return "", "", errors.New(fmt.Sprintf("failed on step 3 (get elective session) #1: %s", err.Error()))
	}
	defer res.Body.Close()
	cookieHdrRaw, ok = res.Header["Set-Cookie"]
	if !ok || len(cookieHdrRaw) < 1 {
		return "", "", errors.New("failed on step 3 (get elective session) #2: no set-cookie header")
	}
	cookieHdr = cookieHdrRaw[0]
	match = regexCookieHdr.FindStringSubmatch(cookieHdr)
	if match == nil {
		return "", "", errors.New("failed on step 3 (get elective session) #3: no jsessionid set")
	}
	jsessionid = match[1]

	// Step: Activate
	req, err = http.NewRequest("GET", fmt.Sprintf("%s/elective2008/edu/pku/stu/elective/controller/supplement/SupplyCancel.do", electRoot), strings.NewReader(""))
	req.Header.Set("Cookie", fmt.Sprintf("JSESSIONID=%s", jsessionid))
	res, err = client.Do(req)
	if err != nil {
		return "", "", errors.New(fmt.Sprintf("failed on step 4 (activate) #1: %s", err.Error()))
	}
	defer res.Body.Close()
	s, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return jsessionid, "", nil
	} else {
		return jsessionid, string(s), nil
	}
}
