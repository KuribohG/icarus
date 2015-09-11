package pku

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Index, Seq, UBound
func EnToken(a string, b string, c int) string {
	conv := func(s string) string {
		return strings.Replace(s, "$", "$$", -1)
	}
	return fmt.Sprintf("%s$#%s$#%d", conv(a), conv(b), c)
}

func DeToken(s string) (string, string, int, error) {
	res := make([]string, 0)
	cur := make([]byte, 0)
	bs := []byte(s)
	ll := len(bs)
	for k := 0; k < ll; k++ {
		c := bs[k]
		if c != '$' {
			cur = append(cur, c)
		} else {
			if k == ll-1 {
				return "", "", 0, errors.New("wrong format")
			}
			nc := bs[k+1]
			if nc == '$' {
				cur = append(cur, '$')
			} else if nc == '#' {
				res = append(res, string(cur))
				cur = make([]byte, 0)
			} else {
				return "", "", 0, errors.New("wrong format")
			}
			k++
		}
	}
	res = append(res, string(cur))

	if len(res) != 3 {
		return "", "", 0, errors.New("wrong amount of parts")
	}

	u, err := strconv.Atoi(res[2])
	if err != nil {
		return "", "", 0, err
	}
	return res[0], res[1], u, nil
}
