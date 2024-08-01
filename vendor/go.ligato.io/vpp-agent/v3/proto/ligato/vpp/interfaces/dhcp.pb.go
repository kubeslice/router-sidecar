// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.4
// source: ligato/vpp/interfaces/dhcp.proto

package vpp_interfaces

import (
	proto "github.com/golang/protobuf/proto"
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

// This is a compile-time assertion that a sufficiently up-to-date version
// of the legacy proto package is being used.
const _ = proto.ProtoPackageIsVersion4

// DHCPLease is a notification, i.e. flows from SB upwards
type DHCPLease struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	InterfaceName   string `protobuf:"bytes,1,opt,name=interface_name,json=interfaceName,proto3" json:"interface_name,omitempty"`
	HostName        string `protobuf:"bytes,2,opt,name=host_name,json=hostName,proto3" json:"host_name,omitempty"`
	IsIpv6          bool   `protobuf:"varint,3,opt,name=is_ipv6,json=isIpv6,proto3" json:"is_ipv6,omitempty"`
	HostPhysAddress string `protobuf:"bytes,4,opt,name=host_phys_address,json=hostPhysAddress,proto3" json:"host_phys_address,omitempty"`
	HostIpAddress   string `protobuf:"bytes,5,opt,name=host_ip_address,json=hostIpAddress,proto3" json:"host_ip_address,omitempty"`       // IP addresses in the format <ipAddress>/<ipPrefix>
	RouterIpAddress string `protobuf:"bytes,6,opt,name=router_ip_address,json=routerIpAddress,proto3" json:"router_ip_address,omitempty"` // IP addresses in the format <ipAddress>/<ipPrefix>
}

func (x *DHCPLease) Reset() {
	*x = DHCPLease{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_interfaces_dhcp_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *DHCPLease) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*DHCPLease) ProtoMessage() {}

func (x *DHCPLease) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_interfaces_dhcp_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use DHCPLease.ProtoReflect.Descriptor instead.
func (*DHCPLease) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_interfaces_dhcp_proto_rawDescGZIP(), []int{0}
}

func (x *DHCPLease) GetInterfaceName() string {
	if x != nil {
		return x.InterfaceName
	}
	return ""
}

func (x *DHCPLease) GetHostName() string {
	if x != nil {
		return x.HostName
	}
	return ""
}

func (x *DHCPLease) GetIsIpv6() bool {
	if x != nil {
		return x.IsIpv6
	}
	return false
}

func (x *DHCPLease) GetHostPhysAddress() string {
	if x != nil {
		return x.HostPhysAddress
	}
	return ""
}

func (x *DHCPLease) GetHostIpAddress() string {
	if x != nil {
		return x.HostIpAddress
	}
	return ""
}

func (x *DHCPLease) GetRouterIpAddress() string {
	if x != nil {
		return x.RouterIpAddress
	}
	return ""
}

var File_ligato_vpp_interfaces_dhcp_proto protoreflect.FileDescriptor

var file_ligato_vpp_interfaces_dhcp_proto_rawDesc = []byte{
	0x0a, 0x20, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x69, 0x6e, 0x74,
	0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73, 0x2f, 0x64, 0x68, 0x63, 0x70, 0x2e, 0x70, 0x72, 0x6f,
	0x74, 0x6f, 0x12, 0x15, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x69,
	0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73, 0x22, 0xe8, 0x01, 0x0a, 0x09, 0x44, 0x48,
	0x43, 0x50, 0x4c, 0x65, 0x61, 0x73, 0x65, 0x12, 0x25, 0x0a, 0x0e, 0x69, 0x6e, 0x74, 0x65, 0x72,
	0x66, 0x61, 0x63, 0x65, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x01, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0d, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x1b,
	0x0a, 0x09, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x6e, 0x61, 0x6d, 0x65, 0x18, 0x02, 0x20, 0x01, 0x28,
	0x09, 0x52, 0x08, 0x68, 0x6f, 0x73, 0x74, 0x4e, 0x61, 0x6d, 0x65, 0x12, 0x17, 0x0a, 0x07, 0x69,
	0x73, 0x5f, 0x69, 0x70, 0x76, 0x36, 0x18, 0x03, 0x20, 0x01, 0x28, 0x08, 0x52, 0x06, 0x69, 0x73,
	0x49, 0x70, 0x76, 0x36, 0x12, 0x2a, 0x0a, 0x11, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x70, 0x68, 0x79,
	0x73, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52,
	0x0f, 0x68, 0x6f, 0x73, 0x74, 0x50, 0x68, 0x79, 0x73, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73,
	0x12, 0x26, 0x0a, 0x0f, 0x68, 0x6f, 0x73, 0x74, 0x5f, 0x69, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72,
	0x65, 0x73, 0x73, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0d, 0x68, 0x6f, 0x73, 0x74, 0x49,
	0x70, 0x41, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x12, 0x2a, 0x0a, 0x11, 0x72, 0x6f, 0x75, 0x74,
	0x65, 0x72, 0x5f, 0x69, 0x70, 0x5f, 0x61, 0x64, 0x64, 0x72, 0x65, 0x73, 0x73, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x09, 0x52, 0x0f, 0x72, 0x6f, 0x75, 0x74, 0x65, 0x72, 0x49, 0x70, 0x41, 0x64, 0x64,
	0x72, 0x65, 0x73, 0x73, 0x42, 0x46, 0x5a, 0x44, 0x67, 0x6f, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74,
	0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2d, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76,
	0x33, 0x2f, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x2f, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76,
	0x70, 0x70, 0x2f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73, 0x3b, 0x76, 0x70,
	0x70, 0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x73, 0x62, 0x06, 0x70, 0x72,
	0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ligato_vpp_interfaces_dhcp_proto_rawDescOnce sync.Once
	file_ligato_vpp_interfaces_dhcp_proto_rawDescData = file_ligato_vpp_interfaces_dhcp_proto_rawDesc
)

func file_ligato_vpp_interfaces_dhcp_proto_rawDescGZIP() []byte {
	file_ligato_vpp_interfaces_dhcp_proto_rawDescOnce.Do(func() {
		file_ligato_vpp_interfaces_dhcp_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_vpp_interfaces_dhcp_proto_rawDescData)
	})
	return file_ligato_vpp_interfaces_dhcp_proto_rawDescData
}

var file_ligato_vpp_interfaces_dhcp_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ligato_vpp_interfaces_dhcp_proto_goTypes = []interface{}{
	(*DHCPLease)(nil), // 0: ligato.vpp.interfaces.DHCPLease
}
var file_ligato_vpp_interfaces_dhcp_proto_depIdxs = []int32{
	0, // [0:0] is the sub-list for method output_type
	0, // [0:0] is the sub-list for method input_type
	0, // [0:0] is the sub-list for extension type_name
	0, // [0:0] is the sub-list for extension extendee
	0, // [0:0] is the sub-list for field type_name
}

func init() { file_ligato_vpp_interfaces_dhcp_proto_init() }
func file_ligato_vpp_interfaces_dhcp_proto_init() {
	if File_ligato_vpp_interfaces_dhcp_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_vpp_interfaces_dhcp_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*DHCPLease); i {
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
			RawDescriptor: file_ligato_vpp_interfaces_dhcp_proto_rawDesc,
			NumEnums:      0,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ligato_vpp_interfaces_dhcp_proto_goTypes,
		DependencyIndexes: file_ligato_vpp_interfaces_dhcp_proto_depIdxs,
		MessageInfos:      file_ligato_vpp_interfaces_dhcp_proto_msgTypes,
	}.Build()
	File_ligato_vpp_interfaces_dhcp_proto = out.File
	file_ligato_vpp_interfaces_dhcp_proto_rawDesc = nil
	file_ligato_vpp_interfaces_dhcp_proto_goTypes = nil
	file_ligato_vpp_interfaces_dhcp_proto_depIdxs = nil
}