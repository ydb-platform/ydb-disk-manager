// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.28.1
// 	protoc        v4.23.3
// source: locks/locks.proto

package locks

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

type LocksRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LocksRequest) Reset() {
	*x = LocksRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_locks_locks_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LocksRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocksRequest) ProtoMessage() {}

func (x *LocksRequest) ProtoReflect() protoreflect.Message {
	mi := &file_locks_locks_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocksRequest.ProtoReflect.Descriptor instead.
func (*LocksRequest) Descriptor() ([]byte, []int) {
	return file_locks_locks_proto_rawDescGZIP(), []int{0}
}

type LocksResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields
}

func (x *LocksResponse) Reset() {
	*x = LocksResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_locks_locks_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LocksResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LocksResponse) ProtoMessage() {}

func (x *LocksResponse) ProtoReflect() protoreflect.Message {
	mi := &file_locks_locks_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LocksResponse.ProtoReflect.Descriptor instead.
func (*LocksResponse) Descriptor() ([]byte, []int) {
	return file_locks_locks_proto_rawDescGZIP(), []int{1}
}

var File_locks_locks_proto protoreflect.FileDescriptor

var file_locks_locks_proto_rawDesc = []byte{
	0x0a, 0x11, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x2f, 0x6c, 0x6f, 0x63, 0x6b, 0x73, 0x2e, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x12, 0x02, 0x70, 0x62, 0x22, 0x0e, 0x0a, 0x0c, 0x4c, 0x6f, 0x63, 0x6b, 0x73,
	0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x22, 0x0f, 0x0a, 0x0d, 0x4c, 0x6f, 0x63, 0x6b, 0x73,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x32, 0x71, 0x0a, 0x05, 0x4c, 0x6f, 0x63, 0x6b,
	0x73, 0x12, 0x31, 0x0a, 0x08, 0x53, 0x65, 0x74, 0x4c, 0x6f, 0x63, 0x6b, 0x73, 0x12, 0x10, 0x2e,
	0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a,
	0x11, 0x2e, 0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x63, 0x6b, 0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e,
	0x73, 0x65, 0x22, 0x00, 0x12, 0x35, 0x0a, 0x0c, 0x52, 0x65, 0x6c, 0x65, 0x61, 0x73, 0x65, 0x4c,
	0x6f, 0x63, 0x6b, 0x73, 0x12, 0x10, 0x2e, 0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x63, 0x6b, 0x73, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x11, 0x2e, 0x70, 0x62, 0x2e, 0x4c, 0x6f, 0x63, 0x6b,
	0x73, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x22, 0x00, 0x42, 0x36, 0x5a, 0x34, 0x67,
	0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x79, 0x64, 0x62, 0x2d, 0x70, 0x6c,
	0x61, 0x74, 0x66, 0x6f, 0x72, 0x6d, 0x2f, 0x79, 0x64, 0x62, 0x2d, 0x64, 0x69, 0x73, 0x6b, 0x2d,
	0x6d, 0x61, 0x6e, 0x61, 0x67, 0x65, 0x72, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x6f,
	0x63, 0x6b, 0x73, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_locks_locks_proto_rawDescOnce sync.Once
	file_locks_locks_proto_rawDescData = file_locks_locks_proto_rawDesc
)

func file_locks_locks_proto_rawDescGZIP() []byte {
	file_locks_locks_proto_rawDescOnce.Do(func() {
		file_locks_locks_proto_rawDescData = protoimpl.X.CompressGZIP(file_locks_locks_proto_rawDescData)
	})
	return file_locks_locks_proto_rawDescData
}

var file_locks_locks_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_locks_locks_proto_goTypes = []interface{}{
	(*LocksRequest)(nil),  // 0: pb.LocksRequest
	(*LocksResponse)(nil), // 1: pb.LocksResponse
}
var file_locks_locks_proto_depIdxs = []int32{
	0, // 0: pb.Locks.SetLocks:input_type -> pb.LocksRequest
	0, // 1: pb.Locks.ReleaseLocks:input_type -> pb.LocksRequest
	1, // 2: pb.Locks.SetLocks:output_type -> pb.LocksResponse
	1, // 3: pb.Locks.ReleaseLocks:output_type -> pb.LocksResponse
	2, // [2:4] is the sub-list for method output_type
	0, // [0:2] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_locks_locks_proto_init() }
func file_locks_locks_proto_init() {
	if File_locks_locks_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_locks_locks_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LocksRequest); i {
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
		file_locks_locks_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LocksResponse); i {
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
			RawDescriptor: file_locks_locks_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_locks_locks_proto_goTypes,
		DependencyIndexes: file_locks_locks_proto_depIdxs,
		MessageInfos:      file_locks_locks_proto_msgTypes,
	}.Build()
	File_locks_locks_proto = out.File
	file_locks_locks_proto_rawDesc = nil
	file_locks_locks_proto_goTypes = nil
	file_locks_locks_proto_depIdxs = nil
}
