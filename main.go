package main

import (
	"gamelink-go/app"
	"gamelink-go/config"
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
	go a.NewAdminService()
	go a.ConnectNats()
	err = a.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
