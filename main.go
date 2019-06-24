package main

import (
	"gamelink-go/app"
	"gamelink-go/config"
	log "github.com/sirupsen/logrus"
	"sync"
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

	var wg sync.WaitGroup
	wg.Add(1)
	go a.GenerateRanks(&wg)
	wg.Wait()

	//uncomment this if need get dummy tokkens for testing
	//a.GetDummyTokkens(25000)

	err = a.Run()
	if err != nil {
		log.Fatal(err.Error())
	}
}
