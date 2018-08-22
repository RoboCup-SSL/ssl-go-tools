package main

import (
	"fmt"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/sslproto"
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
	logReader, err := persistence.NewReader(filename)
	if err != nil {
		return
	}
	defer logReader.Close()

	channel := logReader.CreateChannel()

	var lastRefereeMsg *sslproto.SSL_Referee
	for c := range channel {
		if c.MessageType.Id != persistence.MessageSslRefbox2013 {
			continue
		}
		lastRefereeMsg, err = c.ParseReferee()
		if err != nil {
			return err
		}
	}

	yellowCards[*lastRefereeMsg.Yellow.Name] += *lastRefereeMsg.Yellow.YellowCards
	yellowCards[*lastRefereeMsg.Blue.Name] += *lastRefereeMsg.Blue.YellowCards
	redCards[*lastRefereeMsg.Yellow.Name] += *lastRefereeMsg.Yellow.RedCards
	redCards[*lastRefereeMsg.Blue.Name] += *lastRefereeMsg.Blue.RedCards
	return
}
