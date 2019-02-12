// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/gloo/projects/gateway/api/v1/virtual_service.proto

package v1

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	v1 "github.com/solo-io/gloo/projects/gloo/pkg/api/v1"
	core "github.com/solo-io/solo-kit/pkg/api/v1/resources/core"
	math "math"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

//
//@solo-kit:resource.short_name=vs
//@solo-kit:resource.plural_name=virtual_services
//@solo-kit:resource.resource_groups=api.gateway.solo.io
//A virtual service describes the set of routes to match for a set of domains.
//Domains must be unique across all virtual services within a gateway (i.e. no overlap between sets).
type VirtualService struct {
	VirtualHost *v1.VirtualHost `protobuf:"bytes,1,opt,name=virtual_host,json=virtualHost,proto3" json:"virtual_host,omitempty"`
	// If provided, the Gateway will serve TLS/SSL traffic for this set of routes
	SslConfig *v1.SslConfig `protobuf:"bytes,2,opt,name=ssl_config,json=sslConfig,proto3" json:"ssl_config,omitempty"`
	// Display only, optional descriptive name.
	// Unlike metadata.name, DisplayName can be changed without deleting the resource.
	DisplayName string `protobuf:"bytes,3,opt,name=display_name,json=displayName,proto3" json:"display_name,omitempty"`
	// Status indicates the validation status of this resource.
	// Status is read-only by clients, and set by gloo during validation
	Status core.Status `protobuf:"bytes,6,opt,name=status,proto3" json:"status" testdiff:"ignore"`
	// Metadata contains the object metadata for this resource
	Metadata             core.Metadata `protobuf:"bytes,7,opt,name=metadata,proto3" json:"metadata"`
	XXX_NoUnkeyedLiteral struct{}      `json:"-"`
	XXX_unrecognized     []byte        `json:"-"`
	XXX_sizecache        int32         `json:"-"`
}

func (m *VirtualService) Reset()         { *m = VirtualService{} }
func (m *VirtualService) String() string { return proto.CompactTextString(m) }
func (*VirtualService) ProtoMessage()    {}
func (*VirtualService) Descriptor() ([]byte, []int) {
	return fileDescriptor_93fa9472926a2049, []int{0}
}
func (m *VirtualService) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_VirtualService.Unmarshal(m, b)
}
func (m *VirtualService) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_VirtualService.Marshal(b, m, deterministic)
}
func (m *VirtualService) XXX_Merge(src proto.Message) {
	xxx_messageInfo_VirtualService.Merge(m, src)
}
func (m *VirtualService) XXX_Size() int {
	return xxx_messageInfo_VirtualService.Size(m)
}
func (m *VirtualService) XXX_DiscardUnknown() {
	xxx_messageInfo_VirtualService.DiscardUnknown(m)
}

var xxx_messageInfo_VirtualService proto.InternalMessageInfo

func (m *VirtualService) GetVirtualHost() *v1.VirtualHost {
	if m != nil {
		return m.VirtualHost
	}
	return nil
}

func (m *VirtualService) GetSslConfig() *v1.SslConfig {
	if m != nil {
		return m.SslConfig
	}
	return nil
}

func (m *VirtualService) GetDisplayName() string {
	if m != nil {
		return m.DisplayName
	}
	return ""
}

func (m *VirtualService) GetStatus() core.Status {
	if m != nil {
		return m.Status
	}
	return core.Status{}
}

func (m *VirtualService) GetMetadata() core.Metadata {
	if m != nil {
		return m.Metadata
	}
	return core.Metadata{}
}

func init() {
	proto.RegisterType((*VirtualService)(nil), "gateway.solo.io.VirtualService")
}

func init() {
	proto.RegisterFile("github.com/solo-io/gloo/projects/gateway/api/v1/virtual_service.proto", fileDescriptor_93fa9472926a2049)
}

var fileDescriptor_93fa9472926a2049 = []byte{
	// 370 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x92, 0x41, 0x4e, 0xc2, 0x40,
	0x14, 0x86, 0x2d, 0x12, 0x94, 0x81, 0x68, 0x6c, 0x88, 0x16, 0x16, 0x42, 0xd8, 0xc8, 0xc6, 0x4e,
	0x90, 0xc4, 0xa0, 0x71, 0x55, 0x63, 0x74, 0xa3, 0x8b, 0x92, 0xb8, 0x70, 0x43, 0x86, 0x32, 0x0c,
	0x23, 0x2d, 0xaf, 0xe9, 0x0c, 0x55, 0x76, 0x1e, 0xc7, 0xb5, 0xa7, 0xf0, 0x14, 0x2c, 0xbc, 0x81,
	0x9e, 0xc0, 0x74, 0x3a, 0x95, 0x90, 0x98, 0x88, 0x2b, 0xe6, 0xe5, 0x7f, 0xdf, 0xff, 0x92, 0x8f,
	0xa2, 0x2b, 0xc6, 0xe5, 0x78, 0x36, 0xb0, 0x3d, 0x08, 0xb0, 0x00, 0x1f, 0x8e, 0x39, 0x60, 0xe6,
	0x03, 0xe0, 0x30, 0x82, 0x47, 0xea, 0x49, 0x81, 0x19, 0x91, 0xf4, 0x89, 0xcc, 0x31, 0x09, 0x39,
	0x8e, 0xdb, 0x38, 0xe6, 0x91, 0x9c, 0x11, 0xbf, 0x2f, 0x68, 0x14, 0x73, 0x8f, 0xda, 0x61, 0x04,
	0x12, 0xcc, 0x5d, 0xbd, 0x65, 0x27, 0x1d, 0x36, 0x87, 0x5a, 0x85, 0x01, 0x03, 0x95, 0xe1, 0xe4,
	0x95, 0xae, 0xd5, 0xda, 0xbf, 0x5c, 0x53, 0xbf, 0x13, 0x2e, 0xb3, 0x03, 0x01, 0x95, 0x64, 0x48,
	0x24, 0xd1, 0x08, 0x5e, 0x03, 0x11, 0x92, 0xc8, 0x99, 0xf8, 0xc7, 0x8d, 0x6c, 0xd6, 0x48, 0xf7,
	0x6f, 0x09, 0xc9, 0xa4, 0xe1, 0x30, 0x82, 0xe7, 0x79, 0x4a, 0x36, 0xdf, 0x72, 0x68, 0xe7, 0x3e,
	0x35, 0xd2, 0x4b, 0x85, 0x98, 0x17, 0xa8, 0x9c, 0x39, 0x1a, 0x83, 0x90, 0x96, 0xd1, 0x30, 0x5a,
	0xa5, 0x93, 0xaa, 0x9d, 0x54, 0x64, 0x7a, 0x6c, 0xcd, 0xdc, 0x80, 0x90, 0x6e, 0x29, 0x5e, 0x0e,
	0xe6, 0x29, 0x42, 0x42, 0xf8, 0x7d, 0x0f, 0xa6, 0x23, 0xce, 0xac, 0x9c, 0x62, 0x0f, 0x56, 0xd9,
	0x9e, 0xf0, 0x2f, 0x55, 0xec, 0x16, 0x45, 0xf6, 0x34, 0x8f, 0x50, 0x79, 0xc8, 0x45, 0xe8, 0x93,
	0x79, 0x7f, 0x4a, 0x02, 0x6a, 0x6d, 0x36, 0x8c, 0x56, 0xd1, 0xc9, 0xbf, 0x7c, 0xe6, 0x0d, 0xb7,
	0xa4, 0x93, 0x3b, 0x12, 0x50, 0xf3, 0x1a, 0x15, 0x52, 0x5d, 0x56, 0x41, 0x95, 0x57, 0x6c, 0x0f,
	0x22, 0xba, 0x2c, 0x57, 0x99, 0x53, 0x7d, 0x5f, 0xd4, 0x37, 0xbe, 0x16, 0xf5, 0x3d, 0x49, 0x85,
	0x1c, 0xf2, 0xd1, 0xe8, 0xbc, 0xc9, 0xd9, 0x14, 0x22, 0xda, 0x74, 0x35, 0x6e, 0x76, 0xd1, 0x76,
	0xf6, 0x57, 0x59, 0x5b, 0xaa, 0x6a, 0x7f, 0xb5, 0xea, 0x56, 0xa7, 0x4e, 0x3e, 0x29, 0x73, 0x7f,
	0xb6, 0x9d, 0xb3, 0xd7, 0x8f, 0x43, 0xe3, 0xa1, 0xb3, 0xf6, 0x97, 0x17, 0x4e, 0x98, 0x76, 0x3f,
	0x28, 0x28, 0xed, 0x9d, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0xfd, 0x3e, 0x16, 0x42, 0xb7, 0x02,
	0x00, 0x00,
}

func (this *VirtualService) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*VirtualService)
	if !ok {
		that2, ok := that.(VirtualService)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !this.VirtualHost.Equal(that1.VirtualHost) {
		return false
	}
	if !this.SslConfig.Equal(that1.SslConfig) {
		return false
	}
	if this.DisplayName != that1.DisplayName {
		return false
	}
	if !this.Status.Equal(&that1.Status) {
		return false
	}
	if !this.Metadata.Equal(&that1.Metadata) {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
