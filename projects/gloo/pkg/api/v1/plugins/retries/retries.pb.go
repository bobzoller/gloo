// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: github.com/solo-io/gloo/projects/gloo/api/v1/plugins/retries/retries.proto

package retries

import (
	bytes "bytes"
	fmt "fmt"
	_ "github.com/gogo/protobuf/gogoproto"
	proto "github.com/gogo/protobuf/proto"
	_ "github.com/gogo/protobuf/types"
	math "math"
	time "time"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf
var _ = time.Kitchen

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion2 // please upgrade the proto package

// Retry Policy applied to a route
type RetryPolicy struct {
	// Specifies the conditions under which retry takes place. These are the same
	// conditions [documented for Envoy](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#config-http-filters-router-x-envoy-retry-on)
	RetryOn string `protobuf:"bytes,1,opt,name=retry_on,json=retryOn,proto3" json:"retry_on,omitempty"`
	// Specifies the allowed number of retries. This parameter is optional and
	// defaults to 1. These are the same conditions [documented for Envoy](https://www.envoyproxy.io/docs/envoy/latest/configuration/http_filters/router_filter#config-http-filters-router-x-envoy-retry-on)
	NumRetries uint32 `protobuf:"varint,2,opt,name=num_retries,json=numRetries,proto3" json:"num_retries,omitempty"`
	// Specifies a non-zero upstream timeout per retry attempt. This parameter is optional.
	PerTryTimeout        *time.Duration `protobuf:"bytes,3,opt,name=per_try_timeout,json=perTryTimeout,proto3,stdduration" json:"per_try_timeout,omitempty"`
	XXX_NoUnkeyedLiteral struct{}       `json:"-"`
	XXX_unrecognized     []byte         `json:"-"`
	XXX_sizecache        int32          `json:"-"`
}

func (m *RetryPolicy) Reset()         { *m = RetryPolicy{} }
func (m *RetryPolicy) String() string { return proto.CompactTextString(m) }
func (*RetryPolicy) ProtoMessage()    {}
func (*RetryPolicy) Descriptor() ([]byte, []int) {
	return fileDescriptor_224e5c9ce828f899, []int{0}
}
func (m *RetryPolicy) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_RetryPolicy.Unmarshal(m, b)
}
func (m *RetryPolicy) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_RetryPolicy.Marshal(b, m, deterministic)
}
func (m *RetryPolicy) XXX_Merge(src proto.Message) {
	xxx_messageInfo_RetryPolicy.Merge(m, src)
}
func (m *RetryPolicy) XXX_Size() int {
	return xxx_messageInfo_RetryPolicy.Size(m)
}
func (m *RetryPolicy) XXX_DiscardUnknown() {
	xxx_messageInfo_RetryPolicy.DiscardUnknown(m)
}

var xxx_messageInfo_RetryPolicy proto.InternalMessageInfo

func (m *RetryPolicy) GetRetryOn() string {
	if m != nil {
		return m.RetryOn
	}
	return ""
}

func (m *RetryPolicy) GetNumRetries() uint32 {
	if m != nil {
		return m.NumRetries
	}
	return 0
}

func (m *RetryPolicy) GetPerTryTimeout() *time.Duration {
	if m != nil {
		return m.PerTryTimeout
	}
	return nil
}

func init() {
	proto.RegisterType((*RetryPolicy)(nil), "retries.plugins.gloo.solo.io.RetryPolicy")
}

func init() {
	proto.RegisterFile("github.com/solo-io/gloo/projects/gloo/api/v1/plugins/retries/retries.proto", fileDescriptor_224e5c9ce828f899)
}

var fileDescriptor_224e5c9ce828f899 = []byte{
	// 272 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x90, 0xcf, 0x4a, 0x03, 0x31,
	0x10, 0x87, 0x89, 0x8a, 0x7f, 0x52, 0x8a, 0xb0, 0x78, 0xd8, 0x16, 0x69, 0x17, 0x4f, 0x7b, 0x31,
	0x41, 0x7d, 0x01, 0x29, 0xa2, 0xe0, 0x45, 0x09, 0x3d, 0x79, 0x59, 0xba, 0x6b, 0x8c, 0xd1, 0x6c,
	0x26, 0xcc, 0x26, 0xc2, 0xbe, 0x86, 0x27, 0x1f, 0xc1, 0xb7, 0x12, 0x7c, 0x12, 0xc9, 0xa6, 0xf6,
	0x26, 0x78, 0xca, 0xfc, 0x86, 0xf9, 0x32, 0x1f, 0x43, 0x6f, 0x95, 0xf6, 0xcf, 0xa1, 0x66, 0x0d,
	0xb4, 0xbc, 0x03, 0x03, 0xa7, 0x1a, 0xb8, 0x32, 0x00, 0xdc, 0x21, 0xbc, 0xc8, 0xc6, 0x77, 0x29,
	0xad, 0x9c, 0xe6, 0x6f, 0x67, 0xdc, 0x99, 0xa0, 0xb4, 0xed, 0x38, 0x4a, 0x8f, 0x5a, 0x6e, 0x5e,
	0xe6, 0x10, 0x3c, 0x64, 0xc7, 0x9b, 0x98, 0xc6, 0x58, 0x44, 0x59, 0xfc, 0x95, 0x69, 0x98, 0xce,
	0x14, 0x80, 0x32, 0x92, 0x0f, 0xb3, 0x75, 0x78, 0xe2, 0x8f, 0x01, 0x57, 0x5e, 0x83, 0x4d, 0xf4,
	0xf4, 0x48, 0x81, 0x82, 0xa1, 0xe4, 0xb1, 0x4a, 0xdd, 0x93, 0x77, 0x42, 0x47, 0x42, 0x7a, 0xec,
	0xef, 0xc1, 0xe8, 0xa6, 0xcf, 0x26, 0x74, 0x3f, 0x6e, 0xe9, 0x2b, 0xb0, 0x39, 0x29, 0x48, 0x79,
	0x20, 0xf6, 0x86, 0x7c, 0x67, 0xb3, 0x39, 0x1d, 0xd9, 0xd0, 0x56, 0x6b, 0x89, 0x7c, 0xab, 0x20,
	0xe5, 0x58, 0x50, 0x1b, 0x5a, 0x91, 0x3a, 0xd9, 0x0d, 0x3d, 0x74, 0x12, 0xab, 0x48, 0x7b, 0xdd,
	0x4a, 0x08, 0x3e, 0xdf, 0x2e, 0x48, 0x39, 0x3a, 0x9f, 0xb0, 0xe4, 0xc6, 0x7e, 0xdd, 0xd8, 0xd5,
	0xda, 0x6d, 0xb1, 0xf3, 0xf1, 0x35, 0x27, 0x62, 0xec, 0x24, 0x2e, 0xb1, 0x5f, 0x26, 0x6a, 0x71,
	0xfd, 0xf9, 0x3d, 0x23, 0x0f, 0x97, 0xff, 0x3b, 0x9d, 0x7b, 0x55, 0x7f, 0x9c, 0xaf, 0xde, 0x1d,
	0xf6, 0x5d, 0xfc, 0x04, 0x00, 0x00, 0xff, 0xff, 0x7e, 0x01, 0xea, 0x90, 0x85, 0x01, 0x00, 0x00,
}

func (this *RetryPolicy) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*RetryPolicy)
	if !ok {
		that2, ok := that.(RetryPolicy)
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
	if this.RetryOn != that1.RetryOn {
		return false
	}
	if this.NumRetries != that1.NumRetries {
		return false
	}
	if this.PerTryTimeout != nil && that1.PerTryTimeout != nil {
		if *this.PerTryTimeout != *that1.PerTryTimeout {
			return false
		}
	} else if this.PerTryTimeout != nil {
		return false
	} else if that1.PerTryTimeout != nil {
		return false
	}
	if !bytes.Equal(this.XXX_unrecognized, that1.XXX_unrecognized) {
		return false
	}
	return true
}
