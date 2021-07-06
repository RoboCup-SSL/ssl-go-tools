package main

import (
	"flag"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/player"
	"log"
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

var logFile = flag.String("file", "", "The log file to play")
var skipNonRunningStages = flag.Bool("skip", false, "Skip frames while not in a running stage")
var nif = flag.String("nif", "", "Network interface (host IP address) to publish all multicast traffic to. If empty, publish to all interfaces.")

var startTimestamp = flag.Int64("startTimestamp", 0, "The unix timestamp [ns] at which the log file should be started")

func main() {
	flag.Parse()

	if *logFile == "" {
		log.Fatal("Missing logfile")
	}

	broadcaster := player.NewBroadcaster()
	broadcaster.SkipNonRunningStages = *skipNonRunningStages
	addSlots(&broadcaster)

	defer broadcaster.Stop()
	if err := broadcaster.Start(*logFile, *startTimestamp); err != nil {
		log.Fatal(err)
	}
}

func addSlots(logger *player.Broadcaster) {
	if *visionLegacyEnabled {
		logger.AddSlot(VisionLegacyType, *addressVisionLegacy, *nif)
	}
	if *visionEnabled {
		logger.AddSlot(VisionType, *addressVision, *nif)
	}
	if *visionTrackerEnabled {
		logger.AddSlot(VisionTrackerType, *addressVisionTracker, *nif)
	}
	if *refereeEnabled {
		logger.AddSlot(RefereeType, *addressReferee, *nif)
	}
}
