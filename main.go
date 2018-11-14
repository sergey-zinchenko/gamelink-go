package main

import (
	"gamelink-go/app"
	"gamelink-go/config"
	"gamelink-go/version"
	log "github.com/sirupsen/logrus"
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
		version.Commit, version.BuildTime, version.Release,
	)
}
