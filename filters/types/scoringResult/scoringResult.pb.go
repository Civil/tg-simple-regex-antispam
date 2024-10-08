// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.34.2
// 	protoc        v5.27.3
// source: scoringResult.proto

package scoringResult

import (
	protoreflect "google.golang.org/protobuf/reflect/protoreflect"
	protoimpl "google.golang.org/protobuf/runtime/protoimpl"
	reflect "reflect"
	sync "sync"
)

const (
	// Verify that this generated code is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(20 - protoimpl.MinVersion)
	// Verify that runtime/protoimpl is sufficiently up-to-date.
	_ = protoimpl.EnforceVersion(protoimpl.MaxVersion - 20)
)

type ScoringResult struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Score  int32  `protobuf:"varint,1,opt,name=score,proto3" json:"score,omitempty"`
	Reason string `protobuf:"bytes,2,opt,name=reason,proto3" json:"reason,omitempty"`
}

func (x *ScoringResult) Reset() {
	*x = ScoringResult{}
	if protoimpl.UnsafeEnabled {
		mi := &file_scoringResult_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *ScoringResult) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*ScoringResult) ProtoMessage() {}

func (x *ScoringResult) ProtoReflect() protoreflect.Message {
	mi := &file_scoringResult_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use ScoringResult.ProtoReflect.Descriptor instead.
func (*ScoringResult) Descriptor() ([]byte, []int) {
	return file_scoringResult_proto_rawDescGZIP(), []int{0}
}

func (x *ScoringResult) GetScore() int32 {
	if x != nil {
		return x.Score
	}
	return 0
}

func (x *ScoringResult) GetReason() string {
	if x != nil {
		return x.Reason
	}
	return ""
}

var File_scoringResult_proto protoreflect.FileDescriptor

var file_scoringResult_proto_rawDesc = []byte{
	0x0a, 0x13, 0x73, 0x63, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x2e,
	0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x73, 0x63, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65,
	0x73, 0x75, 0x6c, 0x74, 0x22, 0x3d, 0x0a, 0x0d, 0x53, 0x63, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x52,
	0x65, 0x73, 0x75, 0x6c, 0x74, 0x12, 0x14, 0x0a, 0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x05, 0x52, 0x05, 0x73, 0x63, 0x6f, 0x72, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x72,
	0x65, 0x61, 0x73, 0x6f, 0x6e, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x06, 0x72, 0x65, 0x61,
	0x73, 0x6f, 0x6e, 0x42, 0x47, 0x5a, 0x45, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f,
	0x6d, 0x2f, 0x43, 0x69, 0x76, 0x69, 0x6c, 0x2f, 0x74, 0x67, 0x2d, 0x73, 0x69, 0x6d, 0x70, 0x6c,
	0x65, 0x2d, 0x72, 0x65, 0x67, 0x65, 0x78, 0x2d, 0x61, 0x6e, 0x74, 0x69, 0x73, 0x61, 0x70, 0x6d,
	0x2f, 0x66, 0x69, 0x6c, 0x74, 0x65, 0x72, 0x73, 0x2f, 0x74, 0x79, 0x70, 0x65, 0x73, 0x2f, 0x73,
	0x63, 0x6f, 0x72, 0x69, 0x6e, 0x67, 0x52, 0x65, 0x73, 0x75, 0x6c, 0x74, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_scoringResult_proto_rawDescOnce sync.Once
	file_scoringResult_proto_rawDescData = file_scoringResult_proto_rawDesc
)

func file_scoringResult_proto_rawDescGZIP() []byte {
	file_scoringResult_proto_rawDescOnce.Do(func() {
		file_scoringResult_proto_rawDescData = protoimpl.X.CompressGZIP(file_scoringResult_proto_rawDescData)
	})
	return file_scoringResult_proto_rawDescData
}

var file_scoringResult_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_scoringResult_proto_goTypes = []any{
	(*ScoringResult)(nil), // 0: scoringResult.ScoringResult
}
var file_scoringResult_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_scoringResult_proto_init() }
func file_scoringResult_proto_init() {
	if File_scoringResult_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_scoringResult_proto_msgTypes[0].Exporter = func(v any, i int) any {
			switch v := v.(*ScoringResult); i {
			case 0:
				return &v.state
			case 1:
				return &v.sizeCache
			case 2:
				return &v.unknownFields
			default:
				return nil
			}
		}
	}
	type x struct{}
	out := protoimpl.TypeBuilder{
		File: protoimpl.DescBuilder{
			GoPackagePath: reflect.TypeOf(x{}).PkgPath(),
			RawDescriptor: file_scoringResult_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_scoringResult_proto_goTypes,
		DependencyIndexes: file_scoringResult_proto_depIdxs,
		MessageInfos:      file_scoringResult_proto_msgTypes,
	}.Build()
	File_scoringResult_proto = out.File
	file_scoringResult_proto_rawDesc = nil
	file_scoringResult_proto_goTypes = nil
	file_scoringResult_proto_depIdxs = nil
}
