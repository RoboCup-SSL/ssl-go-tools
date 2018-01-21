package main

import (
	"fmt"
	"github.com/g3force/ssl-log-analyzer/sslreader"
	"io/ioutil"
	"log"
	"os"
)

var yellowCards = make(map[string]uint32)
var redCards = make(map[string]uint32)

func main() {
	if len(os.Args) != 2 {
		log.Fatal("Please pass a directory to analyse")
	}
	dir := os.Args[1]

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}

	for _, f := range files {
		fmt.Println(f.Name())
		err := findCards(dir + "/" + f.Name())
		if err != nil {
			fmt.Printf("%v: %v\n", f.Name(), err)
		}
	}

	fmt.Println("yellow cards")
	for name, cards := range yellowCards {
		fmt.Printf("%20s: %d\n", name, cards)
	}

	fmt.Println("red cards")
	for name, cards := range redCards {
		fmt.Printf("%20s: %d\n", name, cards)
	}
}

func findCards(filename string) (err error) {
	logReader, err := sslreader.NewLogReader(filename)
	if err != nil {
		return
	}
	defer logReader.Close()

	channel := make(chan *sslreader.SSL_Referee, 100)
	go logReader.CreateRefereeChannel(channel)

	var lastRefereeMsg *sslreader.SSL_Referee
	for r := range channel {
		lastRefereeMsg = r
	}

	yellowCards[*lastRefereeMsg.Yellow.Name] += *lastRefereeMsg.Yellow.YellowCards
	yellowCards[*lastRefereeMsg.Blue.Name] += *lastRefereeMsg.Blue.YellowCards
	redCards[*lastRefereeMsg.Yellow.Name] += *lastRefereeMsg.Yellow.RedCards
	redCards[*lastRefereeMsg.Blue.Name] += *lastRefereeMsg.Blue.RedCards
	return
}
