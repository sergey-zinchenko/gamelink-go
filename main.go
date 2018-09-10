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
	if err := a.ConnectDataBases(); err != nil {
		log.Fatal(err.Error())
	} else if err = a.ConnetcGRPC(); err != nil {
		log.Fatal(err.Error())
	} else if err = a.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
