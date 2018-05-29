package main

import (
	"gamelink-go/app"
	"gamelink-go/config"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func init() {
	config.LoadEnvironment()
	if config.IsDevelopmentEnv() {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	log.Printf(
		"Starting the service...\ncommit: %s, build time: %s, release: %s",
		app.Commit, app.BuildTime, app.Release,
	)
	port := os.Getenv("SERVADDR")
	if port == "" {
		log.Fatal("Port is not set.")
	}
	a := app.NewApp()
	if err := a.ConnectDataBases(); err != nil {
		log.Fatal(err.Error())
	} else if err = a.Run(); err != nil {
		log.Fatal(err.Error())
	}
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	log.Print("The service is ready to listen and serve.")

	killSignal := <-interrupt
	switch killSignal {
	case os.Interrupt:
		log.Print("Got SIGINT...")
	case syscall.SIGTERM:
		log.Print("Got SIGTERM...")
	}

	log.Print("The service is shutting down...")
	//Close all connections
	a.Shutdown()
	log.Print("Done")
}
