package main

import (
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"log"
	"os"
	"os/signal"
	"time"
)

var addressVisionLegacy = flag.String("vision-legacy-address", "224.5.23.2:10005", "Multicast address for vision 2010 (legacy)")
var addressVision = flag.String("vision-address", "224.5.23.2:10006", "Multicast address for vision 2014")
var addressVisionTracker = flag.String("vision-tracker-address", "224.5.23.2:10010", "Multicast address for vision tracker 2020")
var addressReferee = flag.String("referee-address", "224.5.23.1:10003", "Multicast address for referee 2013")

var visionLegacyEnabled = flag.Bool("vision-legacy-enabled", true, "Record legacy vision packages")
var visionEnabled = flag.Bool("vision-enabled", true, "Record vision packages")
var visionTrackerEnabled = flag.Bool("vision-tracker-enabled", true, "Record vision tracker packages")
var refereeEnabled = flag.Bool("referee-enabled", true, "Record referee packages")

var VisionLegacyType = persistence.MessageType{Id: persistence.MessageSslVision2010, Name: "vision-legacy"}
var VisionType = persistence.MessageType{Id: persistence.MessageSslVision2014, Name: "vision"}
var VisionTrackerType = persistence.MessageType{Id: persistence.MessageSslVisionTracker2020, Name: "vision-tracker"}
var RefereeType = persistence.MessageType{Id: persistence.MessageSslRefbox2013, Name: "referee"}

func main() {
	flag.Parse()

	if flag.NArg() > 0 {
		log.Fatal("Unexpected arguments: ", flag.Args())
	}

	logger := persistence.NewRecorder()
	addSlots(&logger)
	fileName := time.Now().Format("2006-01-02_15-04-05") + ".log.gz"
	err := logger.StartRecording(fileName)
	if err != nil {
		fmt.Println(err)
		return
	}
	logger.StartReceiving()

	registerToInterrupt(&logger)

	for {
		fmt.Print("\r")
		for _, slot := range logger.Slots {
			fmt.Printf(" | %v: %7d", slot.MessageType.Name, slot.ReceivedMessages)
		}
		fmt.Print(" |")

		time.Sleep(time.Millisecond * 500)
	}
}

func addSlots(logger *persistence.Recorder) {
	if *visionLegacyEnabled {
		logger.AddSlot(VisionLegacyType, *addressVisionLegacy)
	}
	if *visionEnabled {
		logger.AddSlot(VisionType, *addressVision)
	}
	if *visionTrackerEnabled {
		logger.AddSlot(VisionTrackerType, *addressVisionTracker)
	}
	if *refereeEnabled {
		logger.AddSlot(RefereeType, *addressReferee)
	}
}

func registerToInterrupt(recorder *persistence.Recorder) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		for range c {
			err := recorder.StopRecording()
			if err != nil {
				log.Println("Could not stop recorder: ", err)
			}
			os.Exit(0)
		}
	}()
}
