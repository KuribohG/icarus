package pku

import "fmt"

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
		ret = append(ret, EnToken(v.Index, v.Seq, v.UBound))
	}
	return ret
}

func (p PKUWorker) Elect(data []string) []string {
	if len(data) < 2 {
		return []string{"datum are not sufficient"}
	}
	token := data[0]
	jsid := data[1]

	index, seq, ubound, err := DeToken(token)
	if err != nil {
		return []string{err.Error()}
	}

	electable, err := Refresh(jsid, index, seq, ubound)
	if err != nil {
		return []string{err.Error()}
	}
	if electable {
		res, err := Supplement(jsid, index, seq)
		if err != nil {
			return []string{err.Error()}
		}
		if res {
			return []string{"succeeded"}
		}
	}

	return []string{"full"}
}
