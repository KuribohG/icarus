package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/applepi-icpc/icarus"

	log "github.com/Sirupsen/logrus"
	"github.com/applepi-icpc/icarus/client"
	_ "github.com/applepi-icpc/icarus/client/pku/satellite"
	_ "github.com/applepi-icpc/icarus/client/pku/server"
	"github.com/applepi-icpc/icarus/dispatcher/satellite"
	"github.com/applepi-icpc/icarus/dispatcher/server"
	"github.com/applepi-icpc/icarus/task"
)

var (
	flagPort = flag.Int("port", 13013, "Port used on localhost")
)

func main() {
	flag.Parse()

	fmt.Println("Icarus Personal")
	fmt.Println("---------------")

	localListen := fmt.Sprintf(":%d", *flagPort)
	localUrl := fmt.Sprintf("http://127.0.0.1:%d", *flagPort)
	delay := 200 * time.Millisecond

	regList := client.RegisteredList()
	if len(regList) == 0 {
		log.Errorf("This personal icarus is running with no actual logic.\n")
		os.Exit(-1)
	}

	server.InitDefaultDispatcher()
	go func() {
		log.Infof("Launching local server on %s...", localListen)
		mux := http.NewServeMux()
		mux.Handle("/", server.DefaultDispatcher)
		panic(http.ListenAndServe(localListen, mux))
	}()

	log.Infof("Launching local satellite on %s...", localUrl)
	satellite.SilentSatellite = true
	for i := 0; i < 8; i++ {
		go satellite.StandardSatellite(localUrl, delay)
	}

	var handler string
	if len(regList) > 1 {
		fmt.Printf("Select a handler in %v: ", regList)
		_, err := fmt.Scanf("%s", &handler)
		if err != nil {
			panic(err)
		}
	} else {
		handler = regList[0]
	}

	cli, err := client.GetHandle(handler)
	if err != nil {
		log.Errorf("Fatal error occured when getting handler %s: %s\n", handler, err.Error())
		os.Exit(1)
	}

	time.Sleep(time.Second)

	var username, password string
	fmt.Printf("Username: ")
	_, err = fmt.Scanln(&username)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Password: ")
	_, err = fmt.Scanln(&password)
	if err != nil {
		panic(err)
	}

	user, err := cli.MakeUser(username, password)
	if err != nil {
		log.Errorf("Fatal error occured when making user: %s\n", err.Error())
		os.Exit(1)
	}
	courses, err := user.ListCourse()
	if err != nil {
		log.Errorf("Fatal error occured when retrieving courses: %s\n", err.Error())
		os.Exit(1)
	}

	if len(courses) == 0 {
		log.Errorf("You have no courses to be supplemented.\n")
		os.Exit(1)
	}
	fmt.Printf("Select a course: \n")
	for k, v := range courses {
		fmt.Printf("%d) %s: %s\n", k, v.Name, v.Desc)
	}
	var idx int
	for {
		fmt.Printf("Choice (%d - %d): ", 0, len(courses)-1)
		_, err := fmt.Scanf("%d", &idx)
		if err != nil {
			fmt.Printf("Failed to read your choice: %s\n", err.Error())
		} else if idx < 0 || idx >= len(courses) {
			fmt.Printf("Out of range.\n")
		} else {
			break
		}
	}

	course, err := client.MakeCourceByData(cli, courses[idx])
	if err != nil {
		log.Errorf("Fatal error occured when making course: %s\n", err.Error())
		os.Exit(1)
	}

	// Make task
	task := task.NewTask(user, []icarus.Course{course})
	task.Start()

	// Wait for elected
	for {
		time.Sleep(2 * time.Second)
		stat := task.Statistics()
		if stat.Elected {
			os.Exit(0)
		}
	}
}
