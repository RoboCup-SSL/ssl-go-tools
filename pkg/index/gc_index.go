package index

import (
	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/logs"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
	"log"
	"os"
	"sort"
)

func WriteGcIndex(filename string) error {
	logReader, err := persistence.NewReader(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create logfile reader")
	}

	channel := logReader.CreateChannel()

	gcIndexer := NewGcIndexer()
	for c := range channel {
		if c.MessageType.Id == persistence.MessageSslRefbox2013 {
			var msg gc.Referee
			if err := proto.Unmarshal(c.Message, &msg); err != nil {
				log.Println("Could not parse referee message:", err)
				continue
			}
			gcIndexer.AddRefereeMessage(&msg)
		}
	}

	if err := logReader.Close(); err != nil {
		return errors.Wrap(err, "Could not close logfile reader")
	}

	logIndex := gcIndexer.CreateLogIndex()

	gcIndexFilename := filename[:len(filename)-4] + "_gc_index.json"
	if err := WriteJson(gcIndexFilename, logIndex); err != nil {
		return errors.Wrap(err, "Could not write json index")
	}

	return nil
}

type GcIndexer struct {
	gameEvents        map[string]*gc.GameEvent
	currentGameEvents []*gc.GameEvent
}

func NewGcIndexer() *GcIndexer {
	return &GcIndexer{
		gameEvents: make(map[string]*gc.GameEvent),
	}
}

func (g *GcIndexer) AddRefereeMessage(msg *gc.Referee) {
	for _, gameEvent := range msg.GameEvents {
		if gameEvent.Id != nil {
			g.gameEvents[*gameEvent.Id] = gameEvent
		}
	}
}

func (g *GcIndexer) CreateLogIndex() *logs.SslLogIndex {
	logIndex := logs.SslLogIndex{}
	for _, gameEvent := range g.gameEvents {
		if gameEvent.CreatedTimestamp == nil {
			log.Println("Found game event without createdTimestamp, skipping")
			continue
		}
		logIndex.GameEvents = append(logIndex.GameEvents, &logs.GameEventIndex{
			Timestamp: gameEvent.CreatedTimestamp,
			GameEvent: gameEvent.Type,
		})
	}
	sort.Slice(logIndex.GameEvents, func(i, j int) bool {
		return *logIndex.GameEvents[i].Timestamp < *logIndex.GameEvents[j].Timestamp
	})
	return &logIndex
}

func WriteJson(filename string, logIndex *logs.SslLogIndex) error {
	f, err := os.Create(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create JSON output file")
	}

	jsonMarsh := protojson.MarshalOptions{EmitUnpopulated: true, Indent: "  "}
	if data, err := jsonMarsh.Marshal(logIndex); err != nil {
		return errors.Wrap(err, "Could not marshal log index to json")
	} else if _, err := f.Write(data); err != nil {
		return errors.Wrap(err, "Could write marshaled data to file")
	}
	return f.Close()
}
