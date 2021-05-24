package index

import (
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/pkg/errors"
	"log"
)

func WriteIndex(filename string) error {
	logReader, err := persistence.NewReader(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create logfile reader")
	}

	if logReader.IsIndexed() {
		return errors.New("File is already indexed")
	}

	channel := logReader.CreateChannel()

	var offsets []int64
	currentOffset := int64(persistence.HeaderSize)
	for c := range channel {
		if c.MessageType.Id == persistence.MessageIndex2021 {
			log.Println("File is already indexed")
			return nil
		}
		offsets = append(offsets, currentOffset)
		currentOffset += 16 + int64(len(c.Message))
	}

	if err := logReader.Close(); err != nil {
		log.Println("Could not close log file:", err)
	}

	log.Printf("Found %d messages in %v", len(offsets), filename)

	l, err := persistence.NewWriter(filename)
	if err != nil {
		return errors.Wrap(err, "Could not create logfile writer")
	}

	if err := l.WriteIndex(offsets); err != nil {
		return errors.Wrap(err, "Could not write index message")
	}

	return l.Close()
}
