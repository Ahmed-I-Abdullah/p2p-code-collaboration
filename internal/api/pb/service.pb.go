// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.32.0
// 	protoc        v4.25.2
// source: pb/service.proto

package pb

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

type RepoInitRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name    string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	FromCli bool   `protobuf:"varint,2,opt,name=from_cli,json=fromCli,proto3" json:"from_cli,omitempty"`
}

func (x *RepoInitRequest) Reset() {
	*x = RepoInitRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepoInitRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepoInitRequest) ProtoMessage() {}

func (x *RepoInitRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepoInitRequest.ProtoReflect.Descriptor instead.
func (*RepoInitRequest) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{0}
}

func (x *RepoInitRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *RepoInitRequest) GetFromCli() bool {
	if x != nil {
		return x.FromCli
	}
	return false
}

type RepoInitResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *RepoInitResponse) Reset() {
	*x = RepoInitResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[1]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepoInitResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepoInitResponse) ProtoMessage() {}

func (x *RepoInitResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[1]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepoInitResponse.ProtoReflect.Descriptor instead.
func (*RepoInitResponse) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{1}
}

func (x *RepoInitResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *RepoInitResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

var File_pb_service_proto protoreflect.FileDescriptor

var file_pb_service_proto_rawDesc = []byte{
	0x0a, 0x10, 0x70, 0x62, 0x2f, 0x73, 0x65, 0x72, 0x76, 0x69, 0x63, 0x65, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x03, 0x61, 0x70, 0x69, 0x22, 0x40, 0x0a, 0x0f, 0x52, 0x65, 0x70, 0x6f, 0x49,
	0x6e, 0x69, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61,
	0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x12, 0x19,
	0x0a, 0x08, 0x66, 0x72, 0x6f, 0x6d, 0x5f, 0x63, 0x6c, 0x69, 0x18, 0x02, 0x20, 0x01, 0x28, 0x08,
	0x52, 0x07, 0x66, 0x72, 0x6f, 0x6d, 0x43, 0x6c, 0x69, 0x22, 0x46, 0x0a, 0x10, 0x52, 0x65, 0x70,
	0x6f, 0x49, 0x6e, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a,
	0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61,
	0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67,
	0x65, 0x32, 0x41, 0x0a, 0x0a, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12,
	0x33, 0x0a, 0x04, 0x49, 0x6e, 0x69, 0x74, 0x12, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x65,
	0x70, 0x6f, 0x49, 0x6e, 0x69, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x49, 0x6e, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x42, 0x44, 0x5a, 0x42, 0x67, 0x69, 0x74, 0x68, 0x75, 0x62, 0x2e, 0x63,
	0x6f, 0x6d, 0x2f, 0x41, 0x68, 0x6d, 0x65, 0x64, 0x2d, 0x49, 0x2d, 0x41, 0x62, 0x64, 0x75, 0x6c,
	0x6c, 0x61, 0x68, 0x2f, 0x70, 0x32, 0x70, 0x2d, 0x63, 0x6f, 0x64, 0x65, 0x2d, 0x63, 0x6f, 0x6c,
	0x6c, 0x61, 0x62, 0x6f, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x6e, 0x61, 0x6c, 0x2f, 0x61, 0x70, 0x69, 0x2f, 0x70, 0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x33,
}

var (
	file_pb_service_proto_rawDescOnce sync.Once
	file_pb_service_proto_rawDescData = file_pb_service_proto_rawDesc
)

func file_pb_service_proto_rawDescGZIP() []byte {
	file_pb_service_proto_rawDescOnce.Do(func() {
		file_pb_service_proto_rawDescData = protoimpl.X.CompressGZIP(file_pb_service_proto_rawDescData)
	})
	return file_pb_service_proto_rawDescData
}

var file_pb_service_proto_msgTypes = make([]protoimpl.MessageInfo, 2)
var file_pb_service_proto_goTypes = []interface{}{
	(*RepoInitRequest)(nil),  // 0: api.RepoInitRequest
	(*RepoInitResponse)(nil), // 1: api.RepoInitResponse
}
var file_pb_service_proto_depIdxs = []int32{
	0, // 0: api.Repository.Init:input_type -> api.RepoInitRequest
	1, // 1: api.Repository.Init:output_type -> api.RepoInitResponse
	1, // [1:2] is the sub-list for method output_type
	0, // [0:1] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_pb_service_proto_init() }
func file_pb_service_proto_init() {
	if File_pb_service_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_pb_service_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepoInitRequest); i {
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
		file_pb_service_proto_msgTypes[1].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepoInitResponse); i {
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
			RawDescriptor: file_pb_service_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   2,
			NumExtensions: 0,
			NumServices:   1,
		},
		GoTypes:           file_pb_service_proto_goTypes,
		DependencyIndexes: file_pb_service_proto_depIdxs,
		MessageInfos:      file_pb_service_proto_msgTypes,
	}.Build()
	File_pb_service_proto = out.File
	file_pb_service_proto_rawDesc = nil
	file_pb_service_proto_goTypes = nil
	file_pb_service_proto_depIdxs = nil
}
