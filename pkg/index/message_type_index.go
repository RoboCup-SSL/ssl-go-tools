package index

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/RoboCup-SSL/ssl-go-tools/internal/tracked"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"github.com/pkg/errors"
	"google.golang.org/protobuf/proto"
)

const recordHeaderSize = 16 // 8-byte timestamp + 4-byte message type + 4-byte message length
const indexEntrySize = 16   // 8-byte timestamp + 8-byte offset

const (
	ManifestTypeRefbox          = "refbox"
	ManifestTypeVisionDetection = "vision_detection"
	ManifestTypeVisionGeometry  = "vision_geometry"
	ManifestTypeTracker         = "tracker"
)

// Manifest is the top-level structure of the manifest JSON file.
type Manifest struct {
	Indices []ManifestEntry `json:"indices"`
}

// ManifestEntry describes a single index file in the manifest.
type ManifestEntry struct {
	Type   string `json:"type"`
	Path   string `json:"path"`
	Source string `json:"source,omitempty"`
}

// TrackerIndexResult holds the paths of tracker index files keyed by source.
type TrackerIndexResult struct {
	Paths map[string]string
}

// WriteMessageTypeIndices creates separate index files for Vision, Refbox, and Tracker
// in a single pass over the log file.
func WriteMessageTypeIndices(logFilePath string) error {
	logReader, err := persistence.NewReader(logFilePath)
	if err != nil {
		return errors.Wrap(err, "could not create logfile reader")
	}
	defer func() {
		if err := logReader.Close(); err != nil {
			log.Println("could not close log file:", err)
		}
	}()

	channel := logReader.CreateChannel()
	currentOffset := int64(persistence.HeaderSize)

	var refboxEntries [][2]int64
	var visionGeometryEntries [][2]int64
	visionDetectionCameraEntries := make(map[uint32][][2]int64)
	trackerSourceEntries := make(map[string][][2]int64)

	for c := range channel {
		ts := c.Timestamp
		switch c.MessageType.Id {
		case persistence.MessageSslRefbox2013:
			refboxEntries = append(refboxEntries, [2]int64{ts, currentOffset})
		case persistence.MessageSslVision2014:
			isDetection, isGeometry, cameraId := classifyVisionMessage(c.Message)
			if isDetection {
				visionDetectionCameraEntries[cameraId] = append(visionDetectionCameraEntries[cameraId], [2]int64{ts, currentOffset})
			}
			if isGeometry {
				visionGeometryEntries = append(visionGeometryEntries, [2]int64{ts, currentOffset})
			}
		case persistence.MessageSslVisionTracker2020:
			source := extractTrackerSource(c.Message)
			trackerSourceEntries[source] = append(trackerSourceEntries[source], [2]int64{ts, currentOffset})
		}
		currentOffset += recordHeaderSize + int64(len(c.Message))
	}

	manifest := make([]ManifestEntry, 0)

	// Refbox index
	if len(refboxEntries) > 0 {
		refboxPath := getIndexPath(logFilePath, persistence.MessageSslRefbox2013)
		if err := writeIndexFile(refboxPath, refboxEntries); err != nil {
			return errors.Wrap(err, "failed to create index for refbox")
		}
		manifest = append(manifest, ManifestEntry{
			Type: ManifestTypeRefbox,
			Path: filepath.Base(refboxPath),
		})
		log.Printf("Found %d refbox messages in %v", len(refboxEntries), logFilePath)
	}

	// Vision detection indices (one per camera)
	cameraIds := make([]uint32, 0, len(visionDetectionCameraEntries))
	for cameraId := range visionDetectionCameraEntries {
		cameraIds = append(cameraIds, cameraId)
	}
	sort.Slice(cameraIds, func(i, j int) bool { return cameraIds[i] < cameraIds[j] })
	for _, cameraId := range cameraIds {
		entries := visionDetectionCameraEntries[cameraId]
		path := getVisionDetectionCameraIndexPath(logFilePath, cameraId)
		if err := writeIndexFile(path, entries); err != nil {
			return errors.Wrap(err, "failed to create vision detection index for camera")
		}
		manifest = append(manifest, ManifestEntry{
			Type:   ManifestTypeVisionDetection,
			Source: fmt.Sprintf("cam%d", cameraId),
			Path:   filepath.Base(path),
		})
	}

	// Vision geometry index
	visionGeometryPath := getVisionIndexPath(logFilePath, "geometry")
	if len(visionGeometryEntries) > 0 {
		if err := writeIndexFile(visionGeometryPath, visionGeometryEntries); err != nil {
			return errors.Wrap(err, "failed to create vision geometry index")
		}
		manifest = append(manifest, ManifestEntry{
			Type: ManifestTypeVisionGeometry,
			Path: filepath.Base(visionGeometryPath),
		})
	}

	// Tracker indices (one per source)
	sources := make([]string, 0, len(trackerSourceEntries))
	for source := range trackerSourceEntries {
		sources = append(sources, source)
	}
	sort.Strings(sources)
	for _, source := range sources {
		entries := trackerSourceEntries[source]
		path := getTrackerIndexPath(logFilePath, source)
		if err := writeIndexFile(path, entries); err != nil {
			return errors.Wrap(err, "failed to create tracker index for source "+source)
		}
		manifest = append(manifest, ManifestEntry{
			Type:   ManifestTypeTracker,
			Source: source,
			Path:   filepath.Base(path),
		})
	}

	// Write manifest JSON
	manifestPath := strings.TrimSuffix(logFilePath, ".gz") + ".manifest.json"
	if err := writeManifest(manifestPath, manifest); err != nil {
		return errors.Wrap(err, "failed to write manifest file")
	}

	return nil
}

// ReadMessageTypeIndex reads a message type index file and returns timestamp/offset pairs
func ReadMessageTypeIndex(indexPath string) ([][2]int64, error) {
	file, err := os.Open(indexPath)
	if err != nil {
		return nil, errors.Wrap(err, "could not open index file")
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println("could not close index file:", err)
		}
	}()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, errors.Wrap(err, "could not stat index file")
	}

	entrySize := int64(indexEntrySize)
	numEntries := fileInfo.Size() / entrySize

	entries := make([][2]int64, numEntries)
	for i := range entries {
		if err := binary.Read(file, binary.BigEndian, &entries[i][0]); err != nil {
			return nil, errors.Wrap(err, "could not read timestamp")
		}
		if err := binary.Read(file, binary.BigEndian, &entries[i][1]); err != nil {
			return nil, errors.Wrap(err, "could not read offset")
		}
	}

	return entries, nil
}

// logBasePath strips .gz suffix and splits into directory and base filename.
func logBasePath(logFilePath string) (dir, base string) {
	p := strings.TrimSuffix(logFilePath, ".gz")
	return filepath.Dir(p), filepath.Base(p)
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

	dir, base := logBasePath(logFilePath)
	return filepath.Join(dir, base+suffix)
}

// Helper: write index file
func writeIndexFile(path string, entries [][2]int64) error {
	file, err := os.Create(path)
	if err != nil {
		return errors.Wrap(err, "could not create index file")
	}
	defer func() {
		if err := file.Close(); err != nil {
			log.Println("could not close index file:", err)
		}
	}()
	for _, entry := range entries {
		if err := binary.Write(file, binary.BigEndian, entry[0]); err != nil {
			return errors.Wrap(err, "could not write timestamp")
		}
		if err := binary.Write(file, binary.BigEndian, entry[1]); err != nil {
			return errors.Wrap(err, "could not write offset")
		}
	}
	return nil
}

// Helper: write manifest JSON
func writeManifest(path string, manifest []ManifestEntry) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func() {
		if err := f.Close(); err != nil {
			log.Println("could not close manifest file:", err)
		}
	}()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	return encoder.Encode(Manifest{Indices: manifest})
}

// Helper: get vision index path
func getVisionIndexPath(logFilePath, visionType string) string {
	dir, base := logBasePath(logFilePath)
	return filepath.Join(dir, base+".vision."+visionType+".idx")
}

// Helper: get tracker index path
func getTrackerIndexPath(logFilePath, source string) string {
	dir, base := logBasePath(logFilePath)
	return filepath.Join(dir, base+".tracker."+source+".idx")
}

// Helper: extract tracker source from message
func extractTrackerSource(msg []byte) string {
	packet := &tracked.TrackerWrapperPacket{}
	if err := proto.Unmarshal(msg, packet); err != nil {
		return "unknown"
	}
	if source := packet.GetSourceName(); source != "" {
		return sanitizeFileFragment(source)
	}
	if uuid := packet.GetUuid(); uuid != "" {
		return sanitizeFileFragment(uuid)
	}
	return "unknown"
}

// Helper: classify vision message as detection and/or geometry, returning the camera ID for detections
func classifyVisionMessage(msg []byte) (bool, bool, uint32) {
	packet := &vision.SSL_WrapperPacket{}
	if err := proto.Unmarshal(msg, packet); err != nil {
		return false, false, 0
	}
	var cameraId uint32
	if packet.GetDetection() != nil {
		cameraId = packet.GetDetection().GetCameraId()
	}
	return packet.GetDetection() != nil, packet.GetGeometry() != nil, cameraId
}

// Helper: get vision detection camera index path
func getVisionDetectionCameraIndexPath(logFilePath string, cameraId uint32) string {
	dir, base := logBasePath(logFilePath)
	return filepath.Join(dir, fmt.Sprintf("%s.vision.detection.cam%d.idx", base, cameraId))
}

// Helper: sanitize a string for use in file paths
func sanitizeFileFragment(value string) string {
	if value == "" {
		return "unknown"
	}
	buf := make([]byte, 0, len(value))
	lastUnderscore := false
	for i := 0; i < len(value); i++ {
		c := value[i]
		isSafe := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.'
		if isSafe {
			buf = append(buf, c)
			lastUnderscore = false
			continue
		}
		if !lastUnderscore {
			buf = append(buf, '_')
			lastUnderscore = true
		}
	}
	result := strings.Trim(string(buf), "_")
	if result == "" {
		return "unknown"
	}
	return result
}
