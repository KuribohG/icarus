package main

import (
	"flag"
	"fmt"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/applepi-icpc/icarus/client"
	_ "github.com/applepi-icpc/icarus/client/pku/server"
	"github.com/applepi-icpc/icarus/dispatcher/server"
	"github.com/applepi-icpc/icarus/task/manager"
	"github.com/gorilla/context"
	"github.com/rs/cors"
)

var (
	flagCORS     = flag.Bool("cors", false, "Enable CORS")
	flagAPIBind  = flag.String("api", ":8000", "API bind address")
	flagTaskBind = flag.String("task", ":8001", "Task bind address")
)

func main() {
	flag.Parse()

	fmt.Println("Icarus Server")
	fmt.Println("-------------")

	disp := server.NewDispatcher(1024, client.RegisteredList())

	go func() {
		log.Infof("Task Handler at %s", *flagTaskBind)
		mux := http.NewServeMux()
		mux.Handle("/", disp)
		panic(http.ListenAndServe(*flagTaskBind, mux))
	}()

	// Start all tasks
	manager.InitManager()
	{
		tasks := manager.ListTasks()
		for _, task := range tasks {
			task.Start()
		}
	}

	log.Infof("API Handler at %s", *flagAPIBind)
	var handler http.Handler = http.DefaultServeMux
	if *flagCORS {
		c := cors.New(cors.Options{
			AllowedOrigins:   []string{"*"},
			AllowedMethods:   []string{"GET", "PUT", "POST", "PATCH", "DELETE"},
			AllowedHeaders:   []string{"*"},
			AllowCredentials: true,
		})
		handler = c.Handler(handler)
	}
	panic(http.ListenAndServe(*flagAPIBind, context.ClearHandler(handler)))
}
