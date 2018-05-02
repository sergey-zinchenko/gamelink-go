package main

import (
	"gamelink-go/app"
	"gamelink-go/config"
	log "github.com/sirupsen/logrus"
)

func init() {
	if config.IsDevelopmentEnv() {
		log.SetLevel(log.DebugLevel)
	} else {
		log.SetLevel(log.WarnLevel)
	}
}

func main() {
	if a, err := app.NewApp(); err != nil {
		log.Fatal(err.Error())
	} else if err = a.Run(); err != nil {
		log.Fatal(err.Error())
	}
}
