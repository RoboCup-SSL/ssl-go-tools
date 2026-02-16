package index

import (
	"encoding/binary"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// WriteMessageTypeIndices creates separate index files for Vision, Refbox, and Tracker
func WriteMessageTypeIndices(logFilePath string) error {
	messageTypes := []persistence.MessageId{
		persistence.MessageSslRefbox2013,
		persistence.MessageSslVision2014,
		persistence.MessageSslVisionTracker2020,
	}

	for _, msgType := range messageTypes {
		outputPath := getIndexPath(logFilePath, msgType)
		if err := WriteMessageTypeIndex(logFilePath, msgType, outputPath); err != nil {
			return errors.Wrapf(err, "Failed to create index for message type %v", msgType)
		}
	}

	return nil
}

// WriteMessageTypeIndex creates an index for a specific message type
func WriteMessageTypeIndex(logFilePath string, messageType persistence.MessageId, outputPath string) error {
	logReader, err := persistence.NewReader(logFilePath)
	if err != nil {
		return errors.Wrap(err, "Could not create logfile reader")
	}
	defer func() {
		if err := logReader.Close(); err != nil {
			log.Println("Could not close log file:", err)
		}
	}()

	channel := logReader.CreateChannel()

	// Track timestamp/offset pairs
	var entries [][2]int64
	currentOffset := int64(persistence.HeaderSize)

	for c := range channel {
		if c.MessageType.Id == messageType {
			entries = append(entries, [2]int64{c.Timestamp, currentOffset})
		}
		currentOffset += 16 + int64(len(c.Message))
	}

	if len(entries) == 0 {
		log.Printf("No messages of type %v found in %v", messageType, logFilePath)
		return nil
	}

	log.Printf("Found %d messages of type %v in %v", len(entries), messageType, logFilePath)

	// Write binary index file
	file, err := os.Create(outputPath)
	if err != nil {
		return errors.Wrap(err, "Could not create index file")
	}
	defer file.Close()

	for _, entry := range entries {
		if err := binary.Write(file, binary.BigEndian, entry[0]); err != nil {
			return errors.Wrap(err, "Could not write timestamp")
		}
		if err := binary.Write(file, binary.BigEndian, entry[1]); err != nil {
			return errors.Wrap(err, "Could not write offset")
		}
	}

	return nil
}

// ReadMessageTypeIndex reads a message type index file and returns timestamp/offset pairs
func ReadMessageTypeIndex(indexPath string) ([][2]int64, error) {
	file, err := os.Open(indexPath)
	if err != nil {
		return nil, errors.Wrap(err, "Could not open index file")
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "Could not stat index file")
	}

	entrySize := int64(16) // 8 bytes timestamp + 8 bytes offset
	numEntries := fileInfo.Size() / entrySize

	entries := make([][2]int64, numEntries)
	for i := range entries {
		if err := binary.Read(file, binary.BigEndian, &entries[i][0]); err != nil {
			return nil, errors.Wrap(err, "Could not read timestamp")
		}
		if err := binary.Read(file, binary.BigEndian, &entries[i][1]); err != nil {
			return nil, errors.Wrap(err, "Could not read offset")
		}
	}

	return entries, nil
}

// getIndexPath generates the output path for a message type index
func getIndexPath(logFilePath string, messageType persistence.MessageId) string {
	var suffix string
	switch messageType {
	case persistence.MessageSslRefbox2013:
		suffix = ".refbox.idx"
	case persistence.MessageSslVision2014:
		suffix = ".vision.idx"
	case persistence.MessageSslVisionTracker2020:
		suffix = ".tracker.idx"
	default:
		suffix = ".unknown.idx"
	}

	// Remove .gz extension if present
	basePath := logFilePath
	if strings.HasSuffix(logFilePath, ".gz") {
		basePath = strings.TrimSuffix(logFilePath, ".gz")
	}

	// Get the base name without extension
	dir := filepath.Dir(basePath)
	base := filepath.Base(basePath)

	return filepath.Join(dir, base+suffix)
}
