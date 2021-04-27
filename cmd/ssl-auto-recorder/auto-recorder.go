package main

import (
	"flag"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/auto"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"os"
	"os/signal"
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

	refereeServer := auto.NewMulticastServer(*addressReferee)
	autoRecorder := auto.NewRecorder(refereeServer)

	addSlots(autoRecorder.Recorder)
	autoRecorder.Start()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	for range c {
		autoRecorder.Stop()
		os.Exit(0)
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
