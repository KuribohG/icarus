package main

import (
	"flag"
	"fmt"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/applepi-icpc/icarus/client"
	_ "github.com/applepi-icpc/icarus/client/pku/satellite"
	"github.com/applepi-icpc/icarus/dispatcher/satellite"
)

var (
	flagRoot     = flag.String("server", "http://127.0.0.1:8001", "URL of Icarus server")
	flagDelay    = flag.Int("delay", 200, "Delay before fetching the next task (millisecond)")
	flagRoutines = flag.Int("r", 8, "Concurrent routines")
)

func main() {
	flag.Parse()

	delay := time.Duration(*flagDelay) * time.Millisecond

	fmt.Println("Icarus Satellite")
	fmt.Println("----------------")

	log.Infof("Fetching tasks from %s", *flagRoot)
	log.Infof("Avaliable handlers: %v", client.RegisteredWorkerList())

	for i := 1; i < *flagRoutines; i++ {
		go satellite.StandardSatellite(*flagRoot, delay)
	}

	// And do it together
	satellite.StandardSatellite(*flagRoot, delay)
}
