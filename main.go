package main

import (
	"gamelink-go/app"
	"gamelink-go/config"
	log "github.com/sirupsen/logrus"
)

var (
	// BuildTime is a time label of the moment when the binary was built
	BuildTime = "unset"
	// Commit is a last commit hash at the moment when the binary was built
	Commit = "unset"
	// Release is a semantic version of current build
	Release = "unset"
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
	a := app.NewApp()
	err := a.ConnectDataBases()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = a.ConnectNats()
	if err != nil {
		log.Fatal(err.Error())
	}
	go func() {
		err = a.ConnectGrpc()
		if err != nil {
			log.Fatal(err.Error())
		}
	}()
	err = a.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
	log.Printf(
		"Starting the service...\ncommit: %s, build time: %s, release: %s",
		Commit, BuildTime, Release,
	)
}
