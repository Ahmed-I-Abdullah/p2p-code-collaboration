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

type RepoPullRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *RepoPullRequest) Reset() {
	*x = RepoPullRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[2]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepoPullRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepoPullRequest) ProtoMessage() {}

func (x *RepoPullRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[2]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepoPullRequest.ProtoReflect.Descriptor instead.
func (*RepoPullRequest) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{2}
}

func (x *RepoPullRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type RepoPullResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success     bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	RepoAddress string `protobuf:"bytes,2,opt,name=repoAddress,proto3" json:"repoAddress,omitempty"`
}

func (x *RepoPullResponse) Reset() {
	*x = RepoPullResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[3]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepoPullResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepoPullResponse) ProtoMessage() {}

func (x *RepoPullResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[3]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepoPullResponse.ProtoReflect.Descriptor instead.
func (*RepoPullResponse) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{3}
}

func (x *RepoPullResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *RepoPullResponse) GetRepoAddress() string {
	if x != nil {
		return x.RepoAddress
	}
	return ""
}

type RepoPushRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *RepoPushRequest) Reset() {
	*x = RepoPushRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[4]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepoPushRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepoPushRequest) ProtoMessage() {}

func (x *RepoPushRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[4]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepoPushRequest.ProtoReflect.Descriptor instead.
func (*RepoPushRequest) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{4}
}

func (x *RepoPushRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type RepoPushResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool   `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
	Message string `protobuf:"bytes,2,opt,name=message,proto3" json:"message,omitempty"`
}

func (x *RepoPushResponse) Reset() {
	*x = RepoPushResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[5]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RepoPushResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RepoPushResponse) ProtoMessage() {}

func (x *RepoPushResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[5]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RepoPushResponse.ProtoReflect.Descriptor instead.
func (*RepoPushResponse) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{5}
}

func (x *RepoPushResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

func (x *RepoPushResponse) GetMessage() string {
	if x != nil {
		return x.Message
	}
	return ""
}

type LeaderUrlRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *LeaderUrlRequest) Reset() {
	*x = LeaderUrlRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[6]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaderUrlRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaderUrlRequest) ProtoMessage() {}

func (x *LeaderUrlRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[6]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaderUrlRequest.ProtoReflect.Descriptor instead.
func (*LeaderUrlRequest) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{6}
}

func (x *LeaderUrlRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type LeaderUrlResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success        string `protobuf:"bytes,1,opt,name=success,proto3" json:"success,omitempty"`
	Name           string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
	GitRepoAddress string `protobuf:"bytes,3,opt,name=gitRepoAddress,proto3" json:"gitRepoAddress,omitempty"`
	GrpcPort       string `protobuf:"bytes,4,opt,name=grpcPort,proto3" json:"grpcPort,omitempty"`
}

func (x *LeaderUrlResponse) Reset() {
	*x = LeaderUrlResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[7]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *LeaderUrlResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*LeaderUrlResponse) ProtoMessage() {}

func (x *LeaderUrlResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[7]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use LeaderUrlResponse.ProtoReflect.Descriptor instead.
func (*LeaderUrlResponse) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{7}
}

func (x *LeaderUrlResponse) GetSuccess() string {
	if x != nil {
		return x.Success
	}
	return ""
}

func (x *LeaderUrlResponse) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *LeaderUrlResponse) GetGitRepoAddress() string {
	if x != nil {
		return x.GitRepoAddress
	}
	return ""
}

func (x *LeaderUrlResponse) GetGrpcPort() string {
	if x != nil {
		return x.GrpcPort
	}
	return ""
}

type NotifyPushCompletionRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
}

func (x *NotifyPushCompletionRequest) Reset() {
	*x = NotifyPushCompletionRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[8]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyPushCompletionRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyPushCompletionRequest) ProtoMessage() {}

func (x *NotifyPushCompletionRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[8]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyPushCompletionRequest.ProtoReflect.Descriptor instead.
func (*NotifyPushCompletionRequest) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{8}
}

func (x *NotifyPushCompletionRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

type NotifyPushCompletionResponse struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Success bool `protobuf:"varint,1,opt,name=success,proto3" json:"success,omitempty"`
}

func (x *NotifyPushCompletionResponse) Reset() {
	*x = NotifyPushCompletionResponse{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[9]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *NotifyPushCompletionResponse) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*NotifyPushCompletionResponse) ProtoMessage() {}

func (x *NotifyPushCompletionResponse) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[9]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use NotifyPushCompletionResponse.ProtoReflect.Descriptor instead.
func (*NotifyPushCompletionResponse) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{9}
}

func (x *NotifyPushCompletionResponse) GetSuccess() bool {
	if x != nil {
		return x.Success
	}
	return false
}

type RequestToPullRequest struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Name       string `protobuf:"bytes,1,opt,name=name,proto3" json:"name,omitempty"`
	GitAddress string `protobuf:"bytes,2,opt,name=gitAddress,proto3" json:"gitAddress,omitempty"`
}

func (x *RequestToPullRequest) Reset() {
	*x = RequestToPullRequest{}
	if protoimpl.UnsafeEnabled {
		mi := &file_pb_service_proto_msgTypes[10]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *RequestToPullRequest) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*RequestToPullRequest) ProtoMessage() {}

func (x *RequestToPullRequest) ProtoReflect() protoreflect.Message {
	mi := &file_pb_service_proto_msgTypes[10]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use RequestToPullRequest.ProtoReflect.Descriptor instead.
func (*RequestToPullRequest) Descriptor() ([]byte, []int) {
	return file_pb_service_proto_rawDescGZIP(), []int{10}
}

func (x *RequestToPullRequest) GetName() string {
	if x != nil {
		return x.Name
	}
	return ""
}

func (x *RequestToPullRequest) GetGitAddress() string {
	if x != nil {
		return x.GitAddress
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
	0x65, 0x22, 0x25, 0x0a, 0x0f, 0x52, 0x65, 0x70, 0x6f, 0x50, 0x75, 0x6c, 0x6c, 0x52, 0x65, 0x71,
	0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01,
	0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x4e, 0x0a, 0x10, 0x52, 0x65, 0x70, 0x6f,
	0x50, 0x75, 0x6c, 0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07,
	0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73,
	0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x20, 0x0a, 0x0b, 0x72, 0x65, 0x70, 0x6f, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x72, 0x65, 0x70,
	0x6f, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x22, 0x25, 0x0a, 0x0f, 0x52, 0x65, 0x70, 0x6f,
	0x50, 0x75, 0x73, 0x68, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22,
	0x46, 0x0a, 0x10, 0x52, 0x65, 0x70, 0x6f, 0x50, 0x75, 0x73, 0x68, 0x52, 0x65, 0x73, 0x70, 0x6f,
	0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12, 0x18, 0x0a,
	0x07, 0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07,
	0x6d, 0x65, 0x73, 0x73, 0x61, 0x67, 0x65, 0x22, 0x26, 0x0a, 0x10, 0x4c, 0x65, 0x61, 0x64, 0x65,
	0x72, 0x55, 0x72, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22,
	0x85, 0x01, 0x0a, 0x11, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x55, 0x72, 0x6c, 0x52, 0x65, 0x73,
	0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73,
	0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x07, 0x73, 0x75, 0x63, 0x63, 0x65, 0x73, 0x73, 0x12,
	0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e,
	0x61, 0x6d, 0x65, 0x12, 0x26, 0x0a, 0x0e, 0x67, 0x69, 0x74, 0x52, 0x65, 0x70, 0x6f, 0x41, 0x64,
	0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x03, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0e, 0x67, 0x69, 0x74,
	0x52, 0x65, 0x70, 0x6f, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x1a, 0x0a, 0x08, 0x67,
	0x72, 0x70, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x08, 0x67,
	0x72, 0x70, 0x63, 0x50, 0x6f, 0x72, 0x74, 0x22, 0x31, 0x0a, 0x1b, 0x4e, 0x6f, 0x74, 0x69, 0x66,
	0x79, 0x50, 0x75, 0x73, 0x68, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69, 0x6f, 0x6e, 0x52,
	0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65, 0x22, 0x38, 0x0a, 0x1c, 0x4e, 0x6f,
	0x74, 0x69, 0x66, 0x79, 0x50, 0x75, 0x73, 0x68, 0x43, 0x6f, 0x6d, 0x70, 0x6c, 0x65, 0x74, 0x69,
	0x6f, 0x6e, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x18, 0x0a, 0x07, 0x73, 0x75,
	0x63, 0x63, 0x65, 0x73, 0x73, 0x18, 0x01, 0x20, 0x01, 0x28, 0x08, 0x52, 0x07, 0x73, 0x75, 0x63,
	0x63, 0x65, 0x73, 0x73, 0x22, 0x4a, 0x0a, 0x14, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x54,
	0x6f, 0x50, 0x75, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x12, 0x12, 0x0a, 0x04,
	0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52, 0x04, 0x6e, 0x61, 0x6d, 0x65,
	0x12, 0x1e, 0x0a, 0x0a, 0x67, 0x69, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x02,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x67, 0x69, 0x74, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x32, 0xb5, 0x01, 0x0a, 0x0a, 0x52, 0x65, 0x70, 0x6f, 0x73, 0x69, 0x74, 0x6f, 0x72, 0x79, 0x12,
	0x33, 0x0a, 0x04, 0x49, 0x6e, 0x69, 0x74, 0x12, 0x14, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x65,
	0x70, 0x6f, 0x49, 0x6e, 0x69, 0x74, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74, 0x1a, 0x15, 0x2e,
	0x61, 0x70, 0x69, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x49, 0x6e, 0x69, 0x74, 0x52, 0x65, 0x73, 0x70,
	0x6f, 0x6e, 0x73, 0x65, 0x12, 0x33, 0x0a, 0x04, 0x50, 0x75, 0x6c, 0x6c, 0x12, 0x14, 0x2e, 0x61,
	0x70, 0x69, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x50, 0x75, 0x6c, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65,
	0x73, 0x74, 0x1a, 0x15, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x52, 0x65, 0x70, 0x6f, 0x50, 0x75, 0x6c,
	0x6c, 0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x12, 0x3d, 0x0a, 0x0c, 0x47, 0x65, 0x74,
	0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x55, 0x72, 0x6c, 0x12, 0x15, 0x2e, 0x61, 0x70, 0x69, 0x2e,
	0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x55, 0x72, 0x6c, 0x52, 0x65, 0x71, 0x75, 0x65, 0x73, 0x74,
	0x1a, 0x16, 0x2e, 0x61, 0x70, 0x69, 0x2e, 0x4c, 0x65, 0x61, 0x64, 0x65, 0x72, 0x55, 0x72, 0x6c,
	0x52, 0x65, 0x73, 0x70, 0x6f, 0x6e, 0x73, 0x65, 0x42, 0x37, 0x5a, 0x35, 0x67, 0x69, 0x74, 0x68,
	0x75, 0x62, 0x2e, 0x63, 0x6f, 0x6d, 0x2f, 0x41, 0x68, 0x6d, 0x65, 0x64, 0x2d, 0x49, 0x2d, 0x41,
	0x62, 0x64, 0x75, 0x6c, 0x6c, 0x61, 0x68, 0x2f, 0x70, 0x32, 0x70, 0x2d, 0x63, 0x6f, 0x64, 0x65,
	0x2d, 0x63, 0x6f, 0x6c, 0x6c, 0x61, 0x62, 0x6f, 0x72, 0x61, 0x74, 0x69, 0x6f, 0x6e, 0x2f, 0x70,
	0x62, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
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

var file_pb_service_proto_msgTypes = make([]protoimpl.MessageInfo, 11)
var file_pb_service_proto_goTypes = []interface{}{
	(*RepoInitRequest)(nil),              // 0: api.RepoInitRequest
	(*RepoInitResponse)(nil),             // 1: api.RepoInitResponse
	(*RepoPullRequest)(nil),              // 2: api.RepoPullRequest
	(*RepoPullResponse)(nil),             // 3: api.RepoPullResponse
	(*RepoPushRequest)(nil),              // 4: api.RepoPushRequest
	(*RepoPushResponse)(nil),             // 5: api.RepoPushResponse
	(*LeaderUrlRequest)(nil),             // 6: api.LeaderUrlRequest
	(*LeaderUrlResponse)(nil),            // 7: api.LeaderUrlResponse
	(*NotifyPushCompletionRequest)(nil),  // 8: api.NotifyPushCompletionRequest
	(*NotifyPushCompletionResponse)(nil), // 9: api.NotifyPushCompletionResponse
	(*RequestToPullRequest)(nil),         // 10: api.RequestToPullRequest
}
var file_pb_service_proto_depIdxs = []int32{
	0, // 0: api.Repository.Init:input_type -> api.RepoInitRequest
	2, // 1: api.Repository.Pull:input_type -> api.RepoPullRequest
	6, // 2: api.Repository.GetLeaderUrl:input_type -> api.LeaderUrlRequest
	1, // 3: api.Repository.Init:output_type -> api.RepoInitResponse
	3, // 4: api.Repository.Pull:output_type -> api.RepoPullResponse
	7, // 5: api.Repository.GetLeaderUrl:output_type -> api.LeaderUrlResponse
	3, // [3:6] is the sub-list for method output_type
	0, // [0:3] is the sub-list for method input_type
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
		file_pb_service_proto_msgTypes[2].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepoPullRequest); i {
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
		file_pb_service_proto_msgTypes[3].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepoPullResponse); i {
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
		file_pb_service_proto_msgTypes[4].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepoPushRequest); i {
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
		file_pb_service_proto_msgTypes[5].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RepoPushResponse); i {
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
		file_pb_service_proto_msgTypes[6].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaderUrlRequest); i {
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
		file_pb_service_proto_msgTypes[7].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*LeaderUrlResponse); i {
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
		file_pb_service_proto_msgTypes[8].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyPushCompletionRequest); i {
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
		file_pb_service_proto_msgTypes[9].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*NotifyPushCompletionResponse); i {
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
		file_pb_service_proto_msgTypes[10].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*RequestToPullRequest); i {
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
			NumMessages:   11,
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
