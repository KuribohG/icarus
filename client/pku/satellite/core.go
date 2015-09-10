package pku

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

const (
	iaaaRoot  = "https://iaaa.pku.edu.cn"
	electRoot = "http://elective.pku.edu.cn"
)

type PKUWorker struct{}

func (p PKUWorker) Login(data []string) []string {
	jsid, _, err := LoginHelper(data)
	if err != nil {
		return []string{"failed", err.Error()}
	} else {
		return []string{"succeeded", jsid}
	}
}

// Index, Seq, UBound
func enToken(a string, b string, c int) string {
	conv := func(s string) string {
		return strings.Replace(s, "$", "$$", -1)
	}
	return fmt.Sprintf("%s$#%s$#%d", conv(a), conv(b), c)
}

func deToken(s string) (string, string, int, error) {
	res := strings.Split(s, "$#")
	if len(res) != 3 {
		return "", "", 0, errors.New("wrong amount of parts in token")
	}
	for i, v := range res {
		res[i] = strings.Replace(v, "$$", "$", -1)
	}
	u, err := strconv.Atoi(res[2])
	if err != nil {
		return "", "", 0, err
	}
	return res[0], res[1], u, nil
}

func (p PKUWorker) ListCourse(data []string) []string {
	jsid, s, err := LoginHelper(data)
	if err != nil {
		return []string{err.Error()}
	}
	res, err := parseList(s)
	if err != nil {
		return []string{err.Error()}
	}
	tot, err := parseTotalPage(s)
	if err != nil {
		return []string{err.Error()}
	}
	for i := 1; i < tot; i++ {
		s, err := getOriginalPage(i, jsid)
		if err != nil {
			return []string{err.Error()}
		}
		resCont, err := parseList(s)
		if err != nil {
			return []string{err.Error()}
		}
		res = append(res, resCont...)
	}

	ret := []string{"succeeded"}
	ret = append(ret, fmt.Sprintf("%d", len(ret)))
	for _, v := range res {
		ret = append(ret, v.Name)
		ret = append(ret, fmt.Sprintf("%s 班: %s, %s", v.GroupID, v.Teacher, v.Msg))
		ret = append(ret, enToken(v.Index, v.Seq, v.UBound))
	}
	return ret
}
