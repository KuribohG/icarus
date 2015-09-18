package pku

import (
	"errors"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/applepi-icpc/icarus"
	"github.com/applepi-icpc/icarus/client"
	"github.com/applepi-icpc/icarus/dispatcher"
	"github.com/applepi-icpc/icarus/dispatcher/server"
	"github.com/applepi-icpc/icarus/task"
)

type PKUClient struct{}

type PKUUser struct {
	userID   string
	password string
}

type PKUCourse struct {
	name  string
	desc  string
	token string
}

type PKULoginSession string

func (pu PKUUser) Name() string {
	return pu.userID
}

func (pu PKUUser) Login() (icarus.LoginSession, error) {
	res := server.DefaultDispatcher.RunSubtask(&dispatcher.Subtask{
		Handler: "pku",
		Type:    dispatcher.SubtaskLogin,
		Data:    []string{pu.userID, pu.password},
	})
	if res.Error != nil {
		return nil, res.Error
	} else {
		// Data:
		// + "failed" / "succeeded"
		// + Reason / SessionID

		if len(res.Data) < 2 {
			log.Warnf("Client PKU: Invalid login data: %v", res.Data)
			return nil, server.ErrInvalidData
		} else if res.Data[0] == "failed" {
			log.Warnf("Client PKU: Failed to login: %s", res.Data[1])
			return nil, server.ErrFailedToLogin
		} else {
			if res.Data[0] != "succeeded" {
				log.Warnf("Client PKU: Invalid login data: %v", res.Data)
				return nil, server.ErrInvalidData
			}
			log.Infof("Client PKU: Successfully login.")
			return PKULoginSession(res.Data[1]), nil
		}
	}
}

func (pu PKUUser) ListCourse() ([]icarus.CourseData, error) {
	res := server.DefaultDispatcher.RunSubtask(&dispatcher.Subtask{
		Handler: "pku",
		Type:    dispatcher.SubtaskList,
		Data:    []string{pu.userID, pu.password},
	})
	if res.Error != nil {
		return nil, res.Error
	} else {
		// Data:
		// + "succeeded" / error
		// + Total amount of courses
		// (For each course)
		//     + Name
		//     + Desc
		//     + Token

		if len(res.Data) < 1 {
			log.Warnf("Client PKU (%s): Invalid course list: %v", pu.userID, res.Data)
			return nil, server.ErrInvalidData
		}
		if res.Data[0] != "succeeded" {
			return nil, errors.New(res.Data[0])
		}
		if len(res.Data) < 2 {
			log.Warnf("Client PKU (%s): Invalid course list: %v", pu.userID, res.Data)
			return nil, server.ErrInvalidData
		}
		count, err := strconv.Atoi(res.Data[1])
		if err != nil {
			log.Warnf("Client PKU (%s): Failed to parse course list count: %s (%v)", pu.userID, err.Error(), res.Data)
			return nil, server.ErrInvalidData
		}
		rawData := res.Data[2:]
		if len(rawData) < 3*count {
			log.Warnf("Client PKU (%s): Invalid course list: %v", pu.userID, res.Data)
			return nil, server.ErrInvalidData
		}
		res := make([]icarus.CourseData, count)
		for i := 0; i < count; i++ {
			res[i] = icarus.CourseData{
				Name:  rawData[i*3],
				Desc:  rawData[i*3+1],
				Token: rawData[i*3+2],
			}
		}
		return res, nil
	}
}

func (pc PKUCourse) Name() string {
	return pc.name
}

func (pc PKUCourse) Elect(session icarus.LoginSession) (bool, error) {
	s, ok := session.(PKULoginSession)
	if !ok {
		log.Warnf("Client PKU (%s): Wrong session type! Session should be a JSESSIONID string.", pc.name)
		return false, server.ErrWrongType
	}

	res := server.DefaultDispatcher.RunSubtask(&dispatcher.Subtask{
		Handler: "pku",
		Type:    dispatcher.SubtaskElect,
		Data:    []string{pc.token, string(s)},
	})
	if res.Error != nil {
		return false, res.Error
	} else {
		// Data:
		// + "succeeded" / "full" / error

		if len(res.Data) < 1 {
			log.Warnf("Client PKU (%s): Invalid elect response: %v", pc.name, res.Data)
			return false, server.ErrInvalidData
		}
		if res.Data[0] == "succeeded" {
			log.Infof("Client PKU (%s): Elected!", pc.name)
			return true, nil
		} else if res.Data[0] == "full" {
			log.Infof("Client PKU (%s): Full.", pc.name)
			return false, nil
		} else if res.Data[0] == "session expired" {
			return false, task.ErrSessionExpired
		} else {
			return false, errors.New(res.Data[0])
		}
	}
}

func (p PKUClient) MakeUser(userID string, password string) (icarus.User, error) {
	return PKUUser{
		userID:   userID,
		password: password,
	}, nil
}

func (p PKUClient) MakeCourse(name string, desc string, token string) (icarus.Course, error) {
	return PKUCourse{
		name:  name,
		desc:  desc,
		token: token,
	}, nil
}

func init() {
	if err := client.RegisterHandle("pku", PKUClient{}); err != nil {
		panic(err)
	}
}
