// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.36.6
// 	protoc        (unknown)
// source: gc/ssl_gc_common.proto

package gc

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
	unsafe "unsafe"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

// Team is either blue or yellow
type Team int32

const (
	// team not set
	Team_UNKNOWN Team = 0
	// yellow team
	Team_YELLOW Team = 1
	// blue team
	Team_BLUE Team = 2
)

// Enum value maps for Team.
var (
	Team_name = map[int32]string{
		0: "UNKNOWN",
		1: "YELLOW",
		2: "BLUE",
	}
	Team_value = map[string]int32{
		"UNKNOWN": 0,
		"YELLOW":  1,
		"BLUE":    2,
	}
)

func (x Team) Enum() *Team {
	p := new(Team)
	*p = x
	return p
}

func (x Team) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Team) Descriptor() protoreflect.EnumDescriptor {
	return file_gc_ssl_gc_common_proto_enumTypes[0].Descriptor()
}

func (Team) Type() protoreflect.EnumType {
	return &file_gc_ssl_gc_common_proto_enumTypes[0]
}

func (x Team) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *Team) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = Team(num)
	return nil
}

// Deprecated: Use Team.Descriptor instead.
func (Team) EnumDescriptor() ([]byte, []int) {
	return file_gc_ssl_gc_common_proto_rawDescGZIP(), []int{0}
}

// Division denotes the current division, which influences some rules
type Division int32

const (
	Division_DIV_UNKNOWN Division = 0
	Division_DIV_A       Division = 1
	Division_DIV_B       Division = 2
)

// Enum value maps for Division.
var (
	Division_name = map[int32]string{
		0: "DIV_UNKNOWN",
		1: "DIV_A",
		2: "DIV_B",
	}
	Division_value = map[string]int32{
		"DIV_UNKNOWN": 0,
		"DIV_A":       1,
		"DIV_B":       2,
	}
)

func (x Division) Enum() *Division {
	p := new(Division)
	*p = x
	return p
}

func (x Division) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Division) Descriptor() protoreflect.EnumDescriptor {
	return file_gc_ssl_gc_common_proto_enumTypes[1].Descriptor()
}

func (Division) Type() protoreflect.EnumType {
	return &file_gc_ssl_gc_common_proto_enumTypes[1]
}

func (x Division) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Do not use.
func (x *Division) UnmarshalJSON(b []byte) error {
	num, err := protoimpl.X.UnmarshalJSONEnum(x.Descriptor(), b)
	if err != nil {
		return err
	}
	*x = Division(num)
	return nil
}

// Deprecated: Use Division.Descriptor instead.
func (Division) EnumDescriptor() ([]byte, []int) {
	return file_gc_ssl_gc_common_proto_rawDescGZIP(), []int{1}
}

// RobotId is the combination of a team and a robot id
type RobotId struct {
	state protoimpl.MessageState `protogen:"open.v1"`
	// the robot number
	Id *uint32 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	// the team that the robot belongs to
	Team          *Team `protobuf:"varint,2,opt,name=team,enum=Team" json:"team,omitempty"`
	unknownFields protoimpl.UnknownFields
	sizeCache     protoimpl.SizeCache
}

func (x *RobotId) Reset() {
	*x = RobotId{}
	mi := &file_gc_ssl_gc_common_proto_msgTypes[0]
	ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
	ms.StoreMessageInfo(mi)
}

func (x *RobotId) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RobotId) ProtoMessage() {}

func (x *RobotId) ProtoReflect() protoreflect.Message {
	mi := &file_gc_ssl_gc_common_proto_msgTypes[0]
	if x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RobotId.ProtoReflect.Descriptor instead.
func (*RobotId) Descriptor() ([]byte, []int) {
	return file_gc_ssl_gc_common_proto_rawDescGZIP(), []int{0}
}

func (x *RobotId) GetId() uint32 {
	if x != nil && x.Id != nil {
		return *x.Id
	}
	return 0
}

func (x *RobotId) GetTeam() Team {
	if x != nil && x.Team != nil {
		return *x.Team
	}
	return Team_UNKNOWN
}

var File_gc_ssl_gc_common_proto protoreflect.FileDescriptor

const file_gc_ssl_gc_common_proto_rawDesc = "" +
	"\n" +
	"\x16gc/ssl_gc_common.proto\"4\n" +
	"\aRobotId\x12\x0e\n" +
	"\x02id\x18\x01 \x01(\rR\x02id\x12\x19\n" +
	"\x04team\x18\x02 \x01(\x0e2\x05.TeamR\x04team*)\n" +
	"\x04Team\x12\v\n" +
	"\aUNKNOWN\x10\x00\x12\n" +
	"\n" +
	"\x06YELLOW\x10\x01\x12\b\n" +
	"\x04BLUE\x10\x02*1\n" +
	"\bDivision\x12\x0f\n" +
	"\vDIV_UNKNOWN\x10\x00\x12\t\n" +
	"\x05DIV_A\x10\x01\x12\t\n" +
	"\x05DIV_B\x10\x02BEB\x10SslGcCommonProtoP\x01Z/github.com/RoboCup-SSL/ssl-go-tools/internal/gc"

var (
	file_gc_ssl_gc_common_proto_rawDescOnce sync.Once
	file_gc_ssl_gc_common_proto_rawDescData []byte
)

func file_gc_ssl_gc_common_proto_rawDescGZIP() []byte {
	file_gc_ssl_gc_common_proto_rawDescOnce.Do(func() {
		file_gc_ssl_gc_common_proto_rawDescData = protoimpl.X.CompressGZIP(unsafe.Slice(unsafe.StringData(file_gc_ssl_gc_common_proto_rawDesc), len(file_gc_ssl_gc_common_proto_rawDesc)))
	})
	return file_gc_ssl_gc_common_proto_rawDescData
}

var file_gc_ssl_gc_common_proto_enumTypes = make([]protoimpl.EnumInfo, 2)
var file_gc_ssl_gc_common_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_gc_ssl_gc_common_proto_goTypes = []any{
	(Team)(0),       // 0: Team
	(Division)(0),   // 1: Division
	(*RobotId)(nil), // 2: RobotId
}
var file_gc_ssl_gc_common_proto_depIdxs = []int32{
	0, // 0: RobotId.team:type_name -> Team
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_gc_ssl_gc_common_proto_init() }
func file_gc_ssl_gc_common_proto_init() {
	if File_gc_ssl_gc_common_proto != nil {
		return
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: unsafe.Slice(unsafe.StringData(file_gc_ssl_gc_common_proto_rawDesc), len(file_gc_ssl_gc_common_proto_rawDesc)),
			NumEnums:      2,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_gc_ssl_gc_common_proto_goTypes,
		DependencyIndexes: file_gc_ssl_gc_common_proto_depIdxs,
		EnumInfos:         file_gc_ssl_gc_common_proto_enumTypes,
		MessageInfos:      file_gc_ssl_gc_common_proto_msgTypes,
	}.Build()
	File_gc_ssl_gc_common_proto = out.File
	file_gc_ssl_gc_common_proto_goTypes = nil
	file_gc_ssl_gc_common_proto_depIdxs = nil
}
