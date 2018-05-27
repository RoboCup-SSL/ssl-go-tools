// Code generated by protoc-gen-go. DO NOT EDIT.
// source: messages_robocup_ssl_detection.proto

package sslproto

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type SSL_DetectionBall struct {
	Confidence           *float32 `protobuf:"fixed32,1,req,name=confidence" json:"confidence,omitempty"`
	Area                 *uint32  `protobuf:"varint,2,opt,name=area" json:"area,omitempty"`
	X                    *float32 `protobuf:"fixed32,3,req,name=x" json:"x,omitempty"`
	Y                    *float32 `protobuf:"fixed32,4,req,name=y" json:"y,omitempty"`
	Z                    *float32 `protobuf:"fixed32,5,opt,name=z" json:"z,omitempty"`
	PixelX               *float32 `protobuf:"fixed32,6,req,name=pixel_x,json=pixelX" json:"pixel_x,omitempty"`
	PixelY               *float32 `protobuf:"fixed32,7,req,name=pixel_y,json=pixelY" json:"pixel_y,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSL_DetectionBall) Reset()         { *m = SSL_DetectionBall{} }
func (m *SSL_DetectionBall) String() string { return proto.CompactTextString(m) }
func (*SSL_DetectionBall) ProtoMessage()    {}
func (*SSL_DetectionBall) Descriptor() ([]byte, []int) {
	return fileDescriptor_messages_robocup_ssl_detection_845148c4b3278070, []int{0}
}
func (m *SSL_DetectionBall) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSL_DetectionBall.Unmarshal(m, b)
}
func (m *SSL_DetectionBall) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSL_DetectionBall.Marshal(b, m, deterministic)
}
func (dst *SSL_DetectionBall) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSL_DetectionBall.Merge(dst, src)
}
func (m *SSL_DetectionBall) XXX_Size() int {
	return xxx_messageInfo_SSL_DetectionBall.Size(m)
}
func (m *SSL_DetectionBall) XXX_DiscardUnknown() {
	xxx_messageInfo_SSL_DetectionBall.DiscardUnknown(m)
}

var xxx_messageInfo_SSL_DetectionBall proto.InternalMessageInfo

func (m *SSL_DetectionBall) GetConfidence() float32 {
	if m != nil && m.Confidence != nil {
		return *m.Confidence
	}
	return 0
}

func (m *SSL_DetectionBall) GetArea() uint32 {
	if m != nil && m.Area != nil {
		return *m.Area
	}
	return 0
}

func (m *SSL_DetectionBall) GetX() float32 {
	if m != nil && m.X != nil {
		return *m.X
	}
	return 0
}

func (m *SSL_DetectionBall) GetY() float32 {
	if m != nil && m.Y != nil {
		return *m.Y
	}
	return 0
}

func (m *SSL_DetectionBall) GetZ() float32 {
	if m != nil && m.Z != nil {
		return *m.Z
	}
	return 0
}

func (m *SSL_DetectionBall) GetPixelX() float32 {
	if m != nil && m.PixelX != nil {
		return *m.PixelX
	}
	return 0
}

func (m *SSL_DetectionBall) GetPixelY() float32 {
	if m != nil && m.PixelY != nil {
		return *m.PixelY
	}
	return 0
}

type SSL_DetectionRobot struct {
	Confidence           *float32 `protobuf:"fixed32,1,req,name=confidence" json:"confidence,omitempty"`
	RobotId              *uint32  `protobuf:"varint,2,opt,name=robot_id,json=robotId" json:"robot_id,omitempty"`
	X                    *float32 `protobuf:"fixed32,3,req,name=x" json:"x,omitempty"`
	Y                    *float32 `protobuf:"fixed32,4,req,name=y" json:"y,omitempty"`
	Orientation          *float32 `protobuf:"fixed32,5,opt,name=orientation" json:"orientation,omitempty"`
	PixelX               *float32 `protobuf:"fixed32,6,req,name=pixel_x,json=pixelX" json:"pixel_x,omitempty"`
	PixelY               *float32 `protobuf:"fixed32,7,req,name=pixel_y,json=pixelY" json:"pixel_y,omitempty"`
	Height               *float32 `protobuf:"fixed32,8,opt,name=height" json:"height,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *SSL_DetectionRobot) Reset()         { *m = SSL_DetectionRobot{} }
func (m *SSL_DetectionRobot) String() string { return proto.CompactTextString(m) }
func (*SSL_DetectionRobot) ProtoMessage()    {}
func (*SSL_DetectionRobot) Descriptor() ([]byte, []int) {
	return fileDescriptor_messages_robocup_ssl_detection_845148c4b3278070, []int{1}
}
func (m *SSL_DetectionRobot) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSL_DetectionRobot.Unmarshal(m, b)
}
func (m *SSL_DetectionRobot) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSL_DetectionRobot.Marshal(b, m, deterministic)
}
func (dst *SSL_DetectionRobot) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSL_DetectionRobot.Merge(dst, src)
}
func (m *SSL_DetectionRobot) XXX_Size() int {
	return xxx_messageInfo_SSL_DetectionRobot.Size(m)
}
func (m *SSL_DetectionRobot) XXX_DiscardUnknown() {
	xxx_messageInfo_SSL_DetectionRobot.DiscardUnknown(m)
}

var xxx_messageInfo_SSL_DetectionRobot proto.InternalMessageInfo

func (m *SSL_DetectionRobot) GetConfidence() float32 {
	if m != nil && m.Confidence != nil {
		return *m.Confidence
	}
	return 0
}

func (m *SSL_DetectionRobot) GetRobotId() uint32 {
	if m != nil && m.RobotId != nil {
		return *m.RobotId
	}
	return 0
}

func (m *SSL_DetectionRobot) GetX() float32 {
	if m != nil && m.X != nil {
		return *m.X
	}
	return 0
}

func (m *SSL_DetectionRobot) GetY() float32 {
	if m != nil && m.Y != nil {
		return *m.Y
	}
	return 0
}

func (m *SSL_DetectionRobot) GetOrientation() float32 {
	if m != nil && m.Orientation != nil {
		return *m.Orientation
	}
	return 0
}

func (m *SSL_DetectionRobot) GetPixelX() float32 {
	if m != nil && m.PixelX != nil {
		return *m.PixelX
	}
	return 0
}

func (m *SSL_DetectionRobot) GetPixelY() float32 {
	if m != nil && m.PixelY != nil {
		return *m.PixelY
	}
	return 0
}

func (m *SSL_DetectionRobot) GetHeight() float32 {
	if m != nil && m.Height != nil {
		return *m.Height
	}
	return 0
}

type SSL_DetectionFrame struct {
	FrameNumber          *uint32               `protobuf:"varint,1,req,name=frame_number,json=frameNumber" json:"frame_number,omitempty"`
	TCapture             *float64              `protobuf:"fixed64,2,req,name=t_capture,json=tCapture" json:"t_capture,omitempty"`
	TSent                *float64              `protobuf:"fixed64,3,req,name=t_sent,json=tSent" json:"t_sent,omitempty"`
	CameraId             *uint32               `protobuf:"varint,4,req,name=camera_id,json=cameraId" json:"camera_id,omitempty"`
	Balls                []*SSL_DetectionBall  `protobuf:"bytes,5,rep,name=balls" json:"balls,omitempty"`
	RobotsYellow         []*SSL_DetectionRobot `protobuf:"bytes,6,rep,name=robots_yellow,json=robotsYellow" json:"robots_yellow,omitempty"`
	RobotsBlue           []*SSL_DetectionRobot `protobuf:"bytes,7,rep,name=robots_blue,json=robotsBlue" json:"robots_blue,omitempty"`
	XXX_NoUnkeyedLiteral struct{}              `json:"-"`
	XXX_unrecognized     []byte                `json:"-"`
	XXX_sizecache        int32                 `json:"-"`
}

func (m *SSL_DetectionFrame) Reset()         { *m = SSL_DetectionFrame{} }
func (m *SSL_DetectionFrame) String() string { return proto.CompactTextString(m) }
func (*SSL_DetectionFrame) ProtoMessage()    {}
func (*SSL_DetectionFrame) Descriptor() ([]byte, []int) {
	return fileDescriptor_messages_robocup_ssl_detection_845148c4b3278070, []int{2}
}
func (m *SSL_DetectionFrame) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_SSL_DetectionFrame.Unmarshal(m, b)
}
func (m *SSL_DetectionFrame) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_SSL_DetectionFrame.Marshal(b, m, deterministic)
}
func (dst *SSL_DetectionFrame) XXX_Merge(src proto.Message) {
	xxx_messageInfo_SSL_DetectionFrame.Merge(dst, src)
}
func (m *SSL_DetectionFrame) XXX_Size() int {
	return xxx_messageInfo_SSL_DetectionFrame.Size(m)
}
func (m *SSL_DetectionFrame) XXX_DiscardUnknown() {
	xxx_messageInfo_SSL_DetectionFrame.DiscardUnknown(m)
}

var xxx_messageInfo_SSL_DetectionFrame proto.InternalMessageInfo

func (m *SSL_DetectionFrame) GetFrameNumber() uint32 {
	if m != nil && m.FrameNumber != nil {
		return *m.FrameNumber
	}
	return 0
}

func (m *SSL_DetectionFrame) GetTCapture() float64 {
	if m != nil && m.TCapture != nil {
		return *m.TCapture
	}
	return 0
}

func (m *SSL_DetectionFrame) GetTSent() float64 {
	if m != nil && m.TSent != nil {
		return *m.TSent
	}
	return 0
}

func (m *SSL_DetectionFrame) GetCameraId() uint32 {
	if m != nil && m.CameraId != nil {
		return *m.CameraId
	}
	return 0
}

func (m *SSL_DetectionFrame) GetBalls() []*SSL_DetectionBall {
	if m != nil {
		return m.Balls
	}
	return nil
}

func (m *SSL_DetectionFrame) GetRobotsYellow() []*SSL_DetectionRobot {
	if m != nil {
		return m.RobotsYellow
	}
	return nil
}

func (m *SSL_DetectionFrame) GetRobotsBlue() []*SSL_DetectionRobot {
	if m != nil {
		return m.RobotsBlue
	}
	return nil
}

func init() {
	proto.RegisterType((*SSL_DetectionBall)(nil), "SSL_DetectionBall")
	proto.RegisterType((*SSL_DetectionRobot)(nil), "SSL_DetectionRobot")
	proto.RegisterType((*SSL_DetectionFrame)(nil), "SSL_DetectionFrame")
}

func init() {
	proto.RegisterFile("messages_robocup_ssl_detection.proto", fileDescriptor_messages_robocup_ssl_detection_845148c4b3278070)
}

var fileDescriptor_messages_robocup_ssl_detection_845148c4b3278070 = []byte{
	// 383 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x4f, 0x8b, 0xd4, 0x30,
	0x18, 0xc6, 0x49, 0x77, 0xfa, 0xc7, 0xb7, 0x33, 0x07, 0x23, 0x6a, 0x44, 0x90, 0x3a, 0x78, 0xe8,
	0x69, 0x0e, 0xe2, 0xc1, 0xf3, 0x2a, 0xc2, 0x82, 0x78, 0xc8, 0x5c, 0xdc, 0x53, 0x48, 0xdb, 0x77,
	0x77, 0x0b, 0x69, 0x53, 0x92, 0x14, 0xdb, 0xfd, 0x38, 0xfa, 0xa1, 0xfc, 0x3a, 0x92, 0xb4, 0xe2,
	0xe8, 0xa2, 0xb2, 0xb7, 0xfc, 0x9e, 0x3c, 0x6f, 0xfa, 0x3c, 0x69, 0xe0, 0x55, 0x87, 0xd6, 0xca,
	0x6b, 0xb4, 0xc2, 0xe8, 0x4a, 0xd7, 0xe3, 0x20, 0xac, 0x55, 0xa2, 0x41, 0x87, 0xb5, 0x6b, 0x75,
	0x7f, 0x18, 0x8c, 0x76, 0x7a, 0xff, 0x8d, 0xc0, 0xc3, 0xe3, 0xf1, 0xa3, 0x78, 0xff, 0x53, 0x3f,
	0x97, 0x4a, 0xd1, 0x17, 0x00, 0xb5, 0xee, 0xaf, 0xda, 0x06, 0xfb, 0x1a, 0x19, 0x29, 0xa2, 0x32,
	0xe2, 0x27, 0x0a, 0xa5, 0xb0, 0x91, 0x06, 0x25, 0x8b, 0x0a, 0x52, 0xee, 0x78, 0x58, 0xd3, 0x2d,
	0x90, 0x89, 0x9d, 0x05, 0x2b, 0x99, 0x3c, 0xcd, 0x6c, 0xb3, 0xd0, 0xec, 0xe9, 0x96, 0xc5, 0x05,
	0xf1, 0x74, 0x4b, 0x9f, 0x42, 0x3a, 0xb4, 0x13, 0x2a, 0x31, 0xb1, 0x24, 0x38, 0x92, 0x80, 0x9f,
	0x7f, 0x6d, 0xcc, 0x2c, 0x3d, 0xd9, 0xb8, 0xdc, 0x7f, 0x27, 0x40, 0x7f, 0x4b, 0xc9, 0x75, 0xa5,
	0xdd, 0x7f, 0x63, 0x3e, 0x83, 0xcc, 0x77, 0x77, 0xa2, 0x6d, 0xd6, 0xa8, 0x69, 0xe0, 0x8b, 0xe6,
	0x9f, 0x69, 0x0b, 0xc8, 0xb5, 0x69, 0xb1, 0x77, 0xd2, 0x7f, 0x6a, 0xcd, 0x7d, 0x2a, 0xdd, 0xbf,
	0x01, 0x7d, 0x02, 0xc9, 0x0d, 0xb6, 0xd7, 0x37, 0x8e, 0x65, 0xe1, 0xb8, 0x95, 0xf6, 0x5f, 0xa3,
	0x3f, 0x9a, 0x7d, 0x30, 0xb2, 0x43, 0xfa, 0x12, 0xb6, 0x57, 0x7e, 0x21, 0xfa, 0xb1, 0xab, 0xd0,
	0x84, 0x6e, 0x3b, 0x9e, 0x07, 0xed, 0x53, 0x90, 0xe8, 0x73, 0x78, 0xe0, 0x44, 0x2d, 0x07, 0x37,
	0x1a, 0x64, 0x51, 0x11, 0x95, 0x84, 0x67, 0xee, 0xdd, 0xc2, 0xf4, 0x31, 0x24, 0x4e, 0x58, 0xec,
	0x5d, 0xe8, 0x48, 0x78, 0xec, 0x8e, 0xd8, 0x3b, 0x3f, 0x53, 0xcb, 0x0e, 0x8d, 0xf4, 0x37, 0xb2,
	0x09, 0x67, 0x66, 0x8b, 0x70, 0xd1, 0xd0, 0x12, 0xe2, 0x4a, 0x2a, 0x65, 0x59, 0x5c, 0x9c, 0x95,
	0xf9, 0x6b, 0x7a, 0xb8, 0xf3, 0x2e, 0xf8, 0x62, 0xa0, 0x6f, 0x61, 0x17, 0xee, 0xd1, 0x8a, 0x19,
	0x95, 0xd2, 0x5f, 0x58, 0x12, 0x26, 0x1e, 0x1d, 0xee, 0xfe, 0x23, 0xbe, 0x5d, 0x9c, 0x97, 0xc1,
	0x48, 0xdf, 0x40, 0xbe, 0x4e, 0x56, 0x6a, 0x44, 0x96, 0xfe, 0x7d, 0x0e, 0x16, 0xdf, 0xb9, 0x1a,
	0xf1, 0x47, 0x00, 0x00, 0x00, 0xff, 0xff, 0xfd, 0xff, 0xb2, 0xcd, 0xcb, 0x02, 0x00, 0x00,
}
