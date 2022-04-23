package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"
)

var (
	UpdateLog *log.Logger
)

func initLogger(fileName string) {
	target := ioutil.Discard
	if fileName != "" {
		file, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalln("Failed to open log file", err)
		}
		target = file
	}
	UpdateLog = log.New(target, "", log.Ldate|log.Ltime)
}

func printHeader() {
	log.Printf("Compose Updater %s\n", BuildVersion)
}

func mainLoop() {
	updater := &Updater{}
	for {
		updater.PerformComposeUpdates()
		if GlobalSettings.Once {
			return
		}
		log.Printf("Waiting %d minutes until next execution...\n", GlobalSettings.Interval)
		time.Sleep(time.Duration(GlobalSettings.Interval) * time.Minute)
	}
}

func main() {
	printHeader()
	ReadSettings()
	if GlobalSettings.Help {
		flag.Usage()
		return
	}
	if GlobalSettings.PrintSettings {
		GlobalSettings.Print()
	}
	CreateEventBus()
	initLogger(GlobalSettings.UpdateLog)
	mainLoop()
}
