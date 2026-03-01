package persistence

import (
	"bytes"
	"os"
	"testing"

	"github.com/RoboCup-SSL/ssl-go-tools/internal/gc"
)

var (
	refboxType  = MessageType{Id: MessageSslRefbox2013, Name: "referee"}
	visionType  = MessageType{Id: MessageSslVision2014, Name: "vision"}
	trackerType = MessageType{Id: MessageSslVisionTracker2020, Name: "vision-tracker"}
)

func TestWriteAndReadMessages(t *testing.T) {
	messages := []*Message{
		{Timestamp: 1000, MessageType: refboxType, Message: newRefboxPayload(t, gc.Referee_NORMAL_FIRST_HALF, gc.Referee_HALT)},
		{Timestamp: 2000, MessageType: visionType, Message: newVisionDetectionPayload(t, 0, 3)},
		{Timestamp: 3000, MessageType: visionType, Message: newVisionGeometryPayload(t)},
		{Timestamp: 4000, MessageType: trackerType, Message: newTrackerPayload(t, "TIGERs")},
		{Timestamp: 5000, MessageType: trackerType, Message: newTrackerPayload(t, "ER-FORCE")},
	}

	path := createTestLogFile(t, messages)

	r, err := NewReader(path)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	ch := r.CreateChannel()
	var got []*Message
	for msg := range ch {
		got = append(got, msg)
	}

	if len(got) != len(messages) {
		t.Fatalf("got %d messages, want %d", len(got), len(messages))
	}
	for i, want := range messages {
		g := got[i]
		if g.Timestamp != want.Timestamp {
			t.Errorf("msg[%d] timestamp = %d, want %d", i, g.Timestamp, want.Timestamp)
		}
		if g.MessageType.Id != want.MessageType.Id {
			t.Errorf("msg[%d] type = %d, want %d", i, g.MessageType.Id, want.MessageType.Id)
		}
		if !bytes.Equal(g.Message, want.Message) {
			t.Errorf("msg[%d] payload mismatch", i)
		}
	}
}

func TestWriteAndReadSingleMessage(t *testing.T) {
	msg := &Message{
		Timestamp:   42000,
		MessageType: refboxType,
		Message:     newRefboxPayload(t, gc.Referee_NORMAL_SECOND_HALF, gc.Referee_STOP),
	}

	path := createTestLogFile(t, []*Message{msg})

	r, err := NewReader(path)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	got, err := r.ReadMessage()
	if err != nil {
		t.Fatalf("ReadMessage: %v", err)
	}
	if got.Timestamp != msg.Timestamp {
		t.Errorf("timestamp = %d, want %d", got.Timestamp, msg.Timestamp)
	}
	if got.MessageType.Id != msg.MessageType.Id {
		t.Errorf("type = %d, want %d", got.MessageType.Id, msg.MessageType.Id)
	}
	if !bytes.Equal(got.Message, msg.Message) {
		t.Error("payload mismatch")
	}
	if r.HasMessage() {
		t.Error("expected no more messages")
	}
}

func TestReadMessageAt(t *testing.T) {
	messages := []*Message{
		{Timestamp: 100, MessageType: refboxType, Message: newRefboxPayload(t, gc.Referee_NORMAL_FIRST_HALF, gc.Referee_HALT)},
		{Timestamp: 200, MessageType: visionType, Message: newVisionDetectionPayload(t, 1, 2)},
		{Timestamp: 300, MessageType: trackerType, Message: newTrackerPayload(t, "test-source")},
	}

	path := createTestLogFile(t, messages)

	// Compute expected offsets: HeaderSize, then each message adds 8+4+4+len(payload)
	offsets := make([]int64, len(messages))
	offset := int64(HeaderSize)
	for i, m := range messages {
		offsets[i] = offset
		offset += 8 + 4 + 4 + int64(len(m.Message))
	}

	r, err := NewReader(path)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	for i, want := range messages {
		got, err := r.ReadMessageAt(offsets[i])
		if err != nil {
			t.Fatalf("ReadMessageAt(%d): %v", offsets[i], err)
		}
		if got.Timestamp != want.Timestamp {
			t.Errorf("msg[%d] timestamp = %d, want %d", i, got.Timestamp, want.Timestamp)
		}
		if got.MessageType.Id != want.MessageType.Id {
			t.Errorf("msg[%d] type = %d, want %d", i, got.MessageType.Id, want.MessageType.Id)
		}
		if !bytes.Equal(got.Message, want.Message) {
			t.Errorf("msg[%d] payload mismatch", i)
		}
	}
}

func TestHeaderSize(t *testing.T) {
	path := createTestLogFile(t, nil)
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Stat: %v", err)
	}
	if info.Size() != int64(HeaderSize) {
		t.Errorf("empty log size = %d, want %d", info.Size(), HeaderSize)
	}
}

func TestWriteAndReadIndex(t *testing.T) {
	messages := []*Message{
		{Timestamp: 100, MessageType: refboxType, Message: newRefboxPayload(t, gc.Referee_NORMAL_FIRST_HALF, gc.Referee_HALT)},
		{Timestamp: 200, MessageType: visionType, Message: newVisionDetectionPayload(t, 0, 1)},
		{Timestamp: 300, MessageType: trackerType, Message: newTrackerPayload(t, "src")},
	}

	path := createTestLogFile(t, messages)

	// Compute offsets
	expectedOffsets := make([]int64, len(messages))
	off := int64(HeaderSize)
	for i, m := range messages {
		expectedOffsets[i] = off
		off += 8 + 4 + 4 + int64(len(m.Message))
	}

	// Write index by reopening in append mode
	w, err := NewWriter(path)
	if err != nil {
		t.Fatalf("NewWriter for index: %v", err)
	}
	if err := w.WriteIndex(expectedOffsets); err != nil {
		t.Fatalf("WriteIndex: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}

	// Verify IsIndexed
	r, err := NewReader(path)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}
	defer r.Close()

	if !r.IsIndexed() {
		t.Fatal("expected file to be indexed")
	}

	gotOffsets, err := r.ReadIndex()
	if err != nil {
		t.Fatalf("ReadIndex: %v", err)
	}

	if len(gotOffsets) != len(expectedOffsets) {
		t.Fatalf("got %d offsets, want %d", len(gotOffsets), len(expectedOffsets))
	}
	for i, want := range expectedOffsets {
		if gotOffsets[i] != want {
			t.Errorf("offset[%d] = %d, want %d", i, gotOffsets[i], want)
		}
	}
}
