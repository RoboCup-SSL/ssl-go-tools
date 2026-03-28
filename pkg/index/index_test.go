package index

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/tracked"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"github.com/RoboCup-SSL/ssl-go-tools/pkg/persistence"
	"google.golang.org/protobuf/proto"
)

var (
	refboxType  = persistence.MessageType{Id: persistence.MessageSslRefbox2013, Name: "referee"}
	visionType  = persistence.MessageType{Id: persistence.MessageSslVision2014, Name: "vision"}
	trackerType = persistence.MessageType{Id: persistence.MessageSslVisionTracker2020, Name: "vision-tracker"}
)

func ptrU32(v uint32) *uint32   { return &v }
func ptrU64(v uint64) *uint64   { return &v }
func ptrI32(v int32) *int32     { return &v }
func ptrF32(v float32) *float32 { return &v }
func ptrF64(v float64) *float64 { return &v }
func ptrStr(v string) *string   { return &v }

func createTestLogFile(t *testing.T, messages []*persistence.Message) string {
	t.Helper()
	path := t.TempDir() + "/test.log"
	w, err := persistence.NewWriter(path)
	if err != nil {
		t.Fatalf("NewWriter: %v", err)
	}
	for _, msg := range messages {
		if err := w.Write(msg); err != nil {
			t.Fatalf("Write: %v", err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	return path
}

func marshalProto(t *testing.T, msg proto.Message) []byte {
	t.Helper()
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("proto.Marshal: %v", err)
	}
	return data
}

func newRefboxPayload(t *testing.T) []byte {
	t.Helper()
	stage := gc.Referee_NORMAL_FIRST_HALF
	cmd := gc.Referee_HALT
	teamInfo := &gc.Referee_TeamInfo{
		Name:        ptrStr("test"),
		Score:       ptrU32(0),
		RedCards:    ptrU32(0),
		YellowCards: ptrU32(0),
		Timeouts:    ptrU32(0),
		TimeoutTime: ptrU32(0),
		Goalkeeper:  ptrU32(0),
	}
	return marshalProto(t, &gc.Referee{
		PacketTimestamp:  ptrU64(0),
		Stage:            &stage,
		Command:          &cmd,
		CommandCounter:   ptrU32(0),
		CommandTimestamp: ptrU64(0),
		Yellow:           teamInfo,
		Blue:             teamInfo,
	})
}

func newDetectionPayload(t *testing.T, cameraId uint32) []byte {
	t.Helper()
	return marshalProto(t, &vision.SSL_WrapperPacket{
		Detection: &vision.SSL_DetectionFrame{
			FrameNumber: ptrU32(1),
			TCapture:    ptrF64(0),
			TSent:       ptrF64(0),
			CameraId:    &cameraId,
			Balls: []*vision.SSL_DetectionBall{{
				Confidence: ptrF32(0.9),
				X:          ptrF32(100),
				Y:          ptrF32(200),
				PixelX:     ptrF32(0),
				PixelY:     ptrF32(0),
			}},
		},
	})
}

func newGeometryPayload(t *testing.T) []byte {
	t.Helper()
	return marshalProto(t, &vision.SSL_WrapperPacket{
		Geometry: &vision.SSL_GeometryData{
			Field: &vision.SSL_GeometryFieldSize{
				FieldLength:   ptrI32(12000),
				FieldWidth:    ptrI32(9000),
				GoalWidth:     ptrI32(1800),
				GoalDepth:     ptrI32(180),
				BoundaryWidth: ptrI32(300),
			},
		},
	})
}

func newTrackerPayload(t *testing.T, source string) []byte {
	t.Helper()
	uuid := "test-uuid"
	return marshalProto(t, &tracked.TrackerWrapperPacket{Uuid: &uuid, SourceName: &source})
}

func TestWriteMessageTypeIndices(t *testing.T) {
	messages := []*persistence.Message{
		{Timestamp: 1000, MessageType: refboxType, Message: newRefboxPayload(t)},
		{Timestamp: 2000, MessageType: visionType, Message: newDetectionPayload(t, 0)},
		{Timestamp: 3000, MessageType: visionType, Message: newGeometryPayload(t)},
		{Timestamp: 4000, MessageType: trackerType, Message: newTrackerPayload(t, "TIGERs")},
		{Timestamp: 5000, MessageType: trackerType, Message: newTrackerPayload(t, "ER-FORCE")},
		{Timestamp: 6000, MessageType: refboxType, Message: newRefboxPayload(t)},
		{Timestamp: 7000, MessageType: trackerType, Message: newTrackerPayload(t, "TIGERs")},
	}

	path := createTestLogFile(t, messages)

	if err := WriteMessageTypeIndices(path); err != nil {
		t.Fatalf("WriteMessageTypeIndices: %v", err)
	}

	// Read and verify manifest
	manifestPath := path + ".manifest.json"
	manifestData, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest Manifest
	if err := json.Unmarshal(manifestData, &manifest); err != nil {
		t.Fatalf("unmarshal manifest: %v", err)
	}

	// Expect: refbox, vision_detection, vision_geometry, tracker(ER-FORCE), tracker(TIGERs)
	if len(manifest.Indices) != 5 {
		t.Fatalf("manifest has %d entries, want 5", len(manifest.Indices))
	}

	typeCount := map[string]int{}
	for _, e := range manifest.Indices {
		typeCount[e.Type]++
	}
	if typeCount[ManifestTypeRefbox] != 1 {
		t.Errorf("expected 1 refbox entry, got %d", typeCount[ManifestTypeRefbox])
	}
	if typeCount[ManifestTypeVisionDetection] != 1 {
		t.Errorf("expected 1 vision_detection entry, got %d", typeCount[ManifestTypeVisionDetection])
	}
	if typeCount[ManifestTypeVisionGeometry] != 1 {
		t.Errorf("expected 1 vision_geometry entry, got %d", typeCount[ManifestTypeVisionGeometry])
	}
	if typeCount[ManifestTypeTracker] != 2 {
		t.Errorf("expected 2 tracker entries, got %d", typeCount[ManifestTypeTracker])
	}

	// Verify index entry counts by reading each index file
	logDir := filepath.Dir(path)
	for _, e := range manifest.Indices {
		entries, err := ReadMessageTypeIndex(filepath.Join(logDir, e.Path))
		if err != nil {
			t.Fatalf("ReadMessageTypeIndex(%s): %v", e.Path, err)
		}
		var wantCount int
		switch {
		case e.Type == ManifestTypeRefbox:
			wantCount = 2
		case e.Type == ManifestTypeVisionDetection:
			wantCount = 1
		case e.Type == ManifestTypeVisionGeometry:
			wantCount = 1
		case e.Type == ManifestTypeTracker && e.Source == "TIGERs":
			wantCount = 2
		case e.Type == ManifestTypeTracker && e.Source == "ER-FORCE":
			wantCount = 1
		}
		if len(entries) != wantCount {
			t.Errorf("index %s (source=%s): got %d entries, want %d", e.Type, e.Source, len(entries), wantCount)
		}

		// Verify each entry's offset points to a valid message
		reader, err := persistence.NewReader(path)
		if err != nil {
			t.Fatalf("NewReader for verification: %v", err)
		}
		for j, entry := range entries {
			msg, err := reader.ReadMessageAt(entry[1])
			if err != nil {
				t.Errorf("ReadMessageAt(offset=%d) for %s[%d]: %v", entry[1], e.Type, j, err)
				continue
			}
			if msg.Timestamp != entry[0] {
				t.Errorf("index %s[%d]: timestamp mismatch: index=%d, message=%d", e.Type, j, entry[0], msg.Timestamp)
			}
		}
		reader.Close()
	}
}

func TestReadMessageTypeIndex(t *testing.T) {
	entries := [][2]int64{
		{1000, 16},
		{2000, 48},
		{3000, 96},
	}
	path := t.TempDir() + "/test.idx"
	if err := writeIndexFile(path, entries); err != nil {
		t.Fatalf("writeIndexFile: %v", err)
	}

	got, err := ReadMessageTypeIndex(path)
	if err != nil {
		t.Fatalf("ReadMessageTypeIndex: %v", err)
	}
	if len(got) != len(entries) {
		t.Fatalf("got %d entries, want %d", len(got), len(entries))
	}
	for i, want := range entries {
		if got[i] != want {
			t.Errorf("entry[%d] = %v, want %v", i, got[i], want)
		}
	}
}

func TestSanitizeFileFragment(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"TIGERs", "TIGERs"},
		{"ER-FORCE", "ER-FORCE"},
		{"hello world", "hello_world"},
		{"a/b\\c:d", "a_b_c_d"},
		{"", "unknown"},
		{"___", "unknown"},
		{"  spaces  ", "spaces"},
		{"a..b", "a..b"},
		{"normal123", "normal123"},
		{"---dashes---", "---dashes---"},
		{"a@b#c$d", "a_b_c_d"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := sanitizeFileFragment(tt.input)
			if got != tt.want {
				t.Errorf("sanitizeFileFragment(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
