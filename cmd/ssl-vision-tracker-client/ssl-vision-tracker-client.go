package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/tracked"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
	"log"
	"net"
	"time"
)

const maxDatagramSize = 8192

var address = flag.String("address", "224.5.23.2:10010", "The multicast address of the tracking source")
var fullScreen = flag.Bool("fullScreen", false, "Print the formatted frame to the console, clearing the screen during print")
var oneFrame = flag.Bool("oneFrame", false, "Print the formatted frame to the console, exit after a single frame was received")
var noBalls = flag.Bool("noBalls", false, "Do not print balls")
var noRobots = flag.Bool("noRobots", false, "Do not print robots")

func main() {
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", *address)
	if err != nil {
		log.Fatal(err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, addr)
	if err != nil {
		log.Fatal(err)
	}

	if err := conn.SetReadBuffer(maxDatagramSize); err != nil {
		log.Printf("Could not set read buffer to %v.", maxDatagramSize)
	}
	log.Println("Receiving from", *address)

	b := make([]byte, maxDatagramSize)
	for {
		n, err := conn.Read(b)
		if err != nil {
			log.Print("Could not read", err)
			time.Sleep(1 * time.Second)
			continue
		}
		if n >= maxDatagramSize {
			log.Fatal("Buffer size too small")
		}
		frame := tracked.TrackerWrapperPacket{}
		if err := proto.Unmarshal(b[0:n], &frame); err != nil {
			log.Println("Could not unmarshal frame")
			continue
		}

		if frame.TrackedFrame != nil {
			if *noBalls {
				frame.TrackedFrame.Balls = []*tracked.TrackedBall{}
				frame.TrackedFrame.KickedBall = nil
			}
			if *noRobots {
				frame.TrackedFrame.Robots = []*tracked.TrackedRobot{}
			}
		}

		if *fullScreen || *oneFrame {
			// clear screen, move cursor to upper left corner
			fmt.Print("\033[H\033[2J")

			// print frame formatted with line breaks
			fmt.Print(prototext.MarshalOptions{Multiline: true}.Format(&frame))
			if *oneFrame {
				return
			}
		} else {
			b, err := json.Marshal(&frame)
			if err != nil {
				log.Fatal(err)
			}
			log.Print(string(b))
		}
	}
}
