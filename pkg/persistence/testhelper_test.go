package persistence

import (
	"testing"

	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/tracked"
	"github.com/RoboCup-SSL/ssl-go-tools/internal/vision"
	"google.golang.org/protobuf/proto"
)

// createTestLogFile writes the given messages to a temporary log file and returns its path.
func createTestLogFile(t *testing.T, messages []*Message) string {
	t.Helper()
	path := t.TempDir() + "/test.log"
	w, err := NewWriter(path)
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

func ptrU32(v uint32) *uint32   { return &v }
func ptrU64(v uint64) *uint64   { return &v }
func ptrI32(v int32) *int32     { return &v }
func ptrF32(v float32) *float32 { return &v }
func ptrF64(v float64) *float64 { return &v }
func ptrStr(v string) *string   { return &v }

func refStage(v gc.Referee_Stage) *gc.Referee_Stage   { return &v }
func refCmd(v gc.Referee_Command) *gc.Referee_Command { return &v }

// newRefboxPayload creates a marshalled gc.Referee message with all required fields.
func newRefboxPayload(t *testing.T, stage gc.Referee_Stage, command gc.Referee_Command) []byte {
	t.Helper()
	teamInfo := &gc.Referee_TeamInfo{
		Name:        ptrStr("test"),
		Score:       ptrU32(0),
		RedCards:    ptrU32(0),
		YellowCards: ptrU32(0),
		Timeouts:    ptrU32(0),
		TimeoutTime: ptrU32(0),
		Goalkeeper:  ptrU32(0),
	}
	msg := &gc.Referee{
		PacketTimestamp:  ptrU64(0),
		Stage:            refStage(stage),
		Command:          refCmd(command),
		CommandCounter:   ptrU32(0),
		CommandTimestamp: ptrU64(0),
		Yellow:           teamInfo,
		Blue:             teamInfo,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal Referee: %v", err)
	}
	return data
}

// newVisionDetectionPayload creates a marshalled SSL_WrapperPacket with a Detection frame.
func newVisionDetectionPayload(t *testing.T, cameraId uint32, numBalls int) []byte {
	t.Helper()
	var balls []*vision.SSL_DetectionBall
	for i := 0; i < numBalls; i++ {
		balls = append(balls, &vision.SSL_DetectionBall{
			Confidence: ptrF32(0.9),
			X:          ptrF32(float32(i * 100)),
			Y:          ptrF32(float32(i * 200)),
			PixelX:     ptrF32(0),
			PixelY:     ptrF32(0),
		})
	}
	msg := &vision.SSL_WrapperPacket{
		Detection: &vision.SSL_DetectionFrame{
			FrameNumber: ptrU32(1),
			TCapture:    ptrF64(0),
			TSent:       ptrF64(0),
			CameraId:    &cameraId,
			Balls:       balls,
		},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal SSL_WrapperPacket (detection): %v", err)
	}
	return data
}

// newVisionGeometryPayload creates a marshalled SSL_WrapperPacket with a Geometry field.
func newVisionGeometryPayload(t *testing.T) []byte {
	t.Helper()
	msg := &vision.SSL_WrapperPacket{
		Geometry: &vision.SSL_GeometryData{
			Field: &vision.SSL_GeometryFieldSize{
				FieldLength:   ptrI32(12000),
				FieldWidth:    ptrI32(9000),
				GoalWidth:     ptrI32(1800),
				GoalDepth:     ptrI32(180),
				BoundaryWidth: ptrI32(300),
			},
		},
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal SSL_WrapperPacket (geometry): %v", err)
	}
	return data
}

// newTrackerPayload creates a marshalled TrackerWrapperPacket with the given source name.
func newTrackerPayload(t *testing.T, sourceName string) []byte {
	t.Helper()
	uuid := "test-uuid"
	msg := &tracked.TrackerWrapperPacket{
		Uuid:       &uuid,
		SourceName: &sourceName,
	}
	data, err := proto.Marshal(msg)
	if err != nil {
		t.Fatalf("marshal TrackerWrapperPacket: %v", err)
	}
	return data
}
