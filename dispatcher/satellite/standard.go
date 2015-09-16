package satellite

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/applepi-icpc/icarus/client"
	"github.com/applepi-icpc/icarus/dispatcher"
)

var SilentSatellite = false

func StandardSatellite(root string, delay time.Duration) {
	p := NewPostOffice(root, client.RegisteredWorkerList())
	for {
		func() {
			defer time.Sleep(delay)

			pm, err := p.GetTask()
			if err != nil {
				if err == ErrNoNewTasks {
					if !SilentSatellite {
						log.Infof("No new task avaliable")
					}
				} else {
					log.Warnf("Error fetching new task: %s", err.Error())
				}
				return
			}

			sb := pm.Subtask
			if !SilentSatellite {
				log.Infof("Get task %d: handler %s, task type %d", sb.ID, sb.Handler, int(sb.Type))
			}

			w, err := client.GetWorker(sb.Handler)
			if err != nil {
				log.Errorf("Unknown handler: %s", sb.Handler)
				return
			}

			var res []string
			if sb.Type == dispatcher.SubtaskLogin {
				res = w.Login(sb.Data)
			} else if sb.Type == dispatcher.SubtaskList {
				res = w.ListCourse(sb.Data)
			} else if sb.Type == dispatcher.SubtaskElect {
				res = w.Elect(sb.Data)
			} else {
				log.Errorf("Unknown subtask type: %d", int(sb.Type))
				return
			}

			resp := &dispatcher.SubtaskResult{
				Data: res,
			}

			err = pm.SendResult(resp)
			if err == ErrTaskVanished {
				if !SilentSatellite {
					log.Warnf("Task %d has gone", sb.ID)
				}
			} else if err != nil {
				log.Warnf("Task %d error sending back result: %s", sb.ID, err.Error())
			}
		}()
	}
}
