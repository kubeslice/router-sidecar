// Code generated by protoc-gen-go. DO NOT EDIT.
// versions:
// 	protoc-gen-go v1.25.0
// 	protoc        v3.12.4
// source: ligato/vpp/l3/route.proto

package vpp_l3

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

type Route_RouteType int32

const (
	// Forwarding is being done in the specified vrf_id only, or according to
	// the specified outgoing interface.
	Route_INTRA_VRF Route_RouteType = 0
	// Forwarding is being done by lookup into a different VRF,
	// specified as via_vrf_id field. In case of these routes, the outgoing
	// interface should not be specified. The next hop IP address
	// does not have to be specified either, in that case VPP does full
	// recursive lookup in the via_vrf_id VRF.
	Route_INTER_VRF Route_RouteType = 1
	// Drops the network communication designated for specific IP address.
	Route_DROP Route_RouteType = 2
)

// Enum value maps for Route_RouteType.
var (
	Route_RouteType_name = map[int32]string{
		0: "INTRA_VRF",
		1: "INTER_VRF",
		2: "DROP",
	}
	Route_RouteType_value = map[string]int32{
		"INTRA_VRF": 0,
		"INTER_VRF": 1,
		"DROP":      2,
	}
)

func (x Route_RouteType) Enum() *Route_RouteType {
	p := new(Route_RouteType)
	*p = x
	return p
}

func (x Route_RouteType) String() string {
	return protoimpl.X.EnumStringOf(x.Descriptor(), protoreflect.EnumNumber(x))
}

func (Route_RouteType) Descriptor() protoreflect.EnumDescriptor {
	return file_ligato_vpp_l3_route_proto_enumTypes[0].Descriptor()
}

func (Route_RouteType) Type() protoreflect.EnumType {
	return &file_ligato_vpp_l3_route_proto_enumTypes[0]
}

func (x Route_RouteType) Number() protoreflect.EnumNumber {
	return protoreflect.EnumNumber(x)
}

// Deprecated: Use Route_RouteType.Descriptor instead.
func (Route_RouteType) EnumDescriptor() ([]byte, []int) {
	return file_ligato_vpp_l3_route_proto_rawDescGZIP(), []int{0, 0}
}

type Route struct {
	state         protoimpl.MessageState
	sizeCache     protoimpl.SizeCache
	unknownFields protoimpl.UnknownFields

	Type Route_RouteType `protobuf:"varint,10,opt,name=type,proto3,enum=ligato.vpp.l3.Route_RouteType" json:"type,omitempty"`
	// VRF identifier, field required for remote client. This value should be
	// consistent with VRF ID in static route key. If it is not, value from
	// key will be preffered and this field will be overriden.
	// Non-zero VRF has to be explicitly created (see api/models/vpp/l3/vrf.proto)
	VrfId uint32 `protobuf:"varint,1,opt,name=vrf_id,json=vrfId,proto3" json:"vrf_id,omitempty"`
	// Destination network defined by IP address and prefix (format: <address>/<prefix>).
	DstNetwork string `protobuf:"bytes,3,opt,name=dst_network,json=dstNetwork,proto3" json:"dst_network,omitempty"`
	// Next hop address.
	NextHopAddr string `protobuf:"bytes,4,opt,name=next_hop_addr,json=nextHopAddr,proto3" json:"next_hop_addr,omitempty"`
	// Interface name of the outgoing interface.
	OutgoingInterface string `protobuf:"bytes,5,opt,name=outgoing_interface,json=outgoingInterface,proto3" json:"outgoing_interface,omitempty"`
	// Weight is used for unequal cost load balancing.
	Weight uint32 `protobuf:"varint,6,opt,name=weight,proto3" json:"weight,omitempty"`
	// Preference defines path preference. Lower preference is preferred.
	// Only paths with the best preference contribute to forwarding (a poor man's primary and backup).
	Preference uint32 `protobuf:"varint,7,opt,name=preference,proto3" json:"preference,omitempty"`
	// Specifies VRF ID for the next hop lookup / recursive lookup
	ViaVrfId uint32 `protobuf:"varint,8,opt,name=via_vrf_id,json=viaVrfId,proto3" json:"via_vrf_id,omitempty"`
}

func (x *Route) Reset() {
	*x = Route{}
	if protoimpl.UnsafeEnabled {
		mi := &file_ligato_vpp_l3_route_proto_msgTypes[0]
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		ms.StoreMessageInfo(mi)
	}
}

func (x *Route) String() string {
	return protoimpl.X.MessageStringOf(x)
}

func (*Route) ProtoMessage() {}

func (x *Route) ProtoReflect() protoreflect.Message {
	mi := &file_ligato_vpp_l3_route_proto_msgTypes[0]
	if protoimpl.UnsafeEnabled && x != nil {
		ms := protoimpl.X.MessageStateOf(protoimpl.Pointer(x))
		if ms.LoadMessageInfo() == nil {
			ms.StoreMessageInfo(mi)
		}
		return ms
	}
	return mi.MessageOf(x)
}

// Deprecated: Use Route.ProtoReflect.Descriptor instead.
func (*Route) Descriptor() ([]byte, []int) {
	return file_ligato_vpp_l3_route_proto_rawDescGZIP(), []int{0}
}

func (x *Route) GetType() Route_RouteType {
	if x != nil {
		return x.Type
	}
	return Route_INTRA_VRF
}

func (x *Route) GetVrfId() uint32 {
	if x != nil {
		return x.VrfId
	}
	return 0
}

func (x *Route) GetDstNetwork() string {
	if x != nil {
		return x.DstNetwork
	}
	return ""
}

func (x *Route) GetNextHopAddr() string {
	if x != nil {
		return x.NextHopAddr
	}
	return ""
}

func (x *Route) GetOutgoingInterface() string {
	if x != nil {
		return x.OutgoingInterface
	}
	return ""
}

func (x *Route) GetWeight() uint32 {
	if x != nil {
		return x.Weight
	}
	return 0
}

func (x *Route) GetPreference() uint32 {
	if x != nil {
		return x.Preference
	}
	return 0
}

func (x *Route) GetViaVrfId() uint32 {
	if x != nil {
		return x.ViaVrfId
	}
	return 0
}

var File_ligato_vpp_l3_route_proto protoreflect.FileDescriptor

var file_ligato_vpp_l3_route_proto_rawDesc = []byte{
	0x0a, 0x19, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x6c, 0x33, 0x2f,
	0x72, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x12, 0x0d, 0x6c, 0x69, 0x67,
	0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e, 0x6c, 0x33, 0x22, 0xd1, 0x02, 0x0a, 0x05, 0x52,
	0x6f, 0x75, 0x74, 0x65, 0x12, 0x32, 0x0a, 0x04, 0x74, 0x79, 0x70, 0x65, 0x18, 0x0a, 0x20, 0x01,
	0x28, 0x0e, 0x32, 0x1e, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x76, 0x70, 0x70, 0x2e,
	0x6c, 0x33, 0x2e, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x2e, 0x52, 0x6f, 0x75, 0x74, 0x65, 0x54, 0x79,
	0x70, 0x65, 0x52, 0x04, 0x74, 0x79, 0x70, 0x65, 0x12, 0x15, 0x0a, 0x06, 0x76, 0x72, 0x66, 0x5f,
	0x69, 0x64, 0x18, 0x01, 0x20, 0x01, 0x28, 0x0d, 0x52, 0x05, 0x76, 0x72, 0x66, 0x49, 0x64, 0x12,
	0x1f, 0x0a, 0x0b, 0x64, 0x73, 0x74, 0x5f, 0x6e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b, 0x18, 0x03,
	0x20, 0x01, 0x28, 0x09, 0x52, 0x0a, 0x64, 0x73, 0x74, 0x4e, 0x65, 0x74, 0x77, 0x6f, 0x72, 0x6b,
	0x12, 0x22, 0x0a, 0x0d, 0x6e, 0x65, 0x78, 0x74, 0x5f, 0x68, 0x6f, 0x70, 0x5f, 0x61, 0x64, 0x64,
	0x72, 0x18, 0x04, 0x20, 0x01, 0x28, 0x09, 0x52, 0x0b, 0x6e, 0x65, 0x78, 0x74, 0x48, 0x6f, 0x70,
	0x41, 0x64, 0x64, 0x72, 0x12, 0x2d, 0x0a, 0x12, 0x6f, 0x75, 0x74, 0x67, 0x6f, 0x69, 0x6e, 0x67,
	0x5f, 0x69, 0x6e, 0x74, 0x65, 0x72, 0x66, 0x61, 0x63, 0x65, 0x18, 0x05, 0x20, 0x01, 0x28, 0x09,
	0x52, 0x11, 0x6f, 0x75, 0x74, 0x67, 0x6f, 0x69, 0x6e, 0x67, 0x49, 0x6e, 0x74, 0x65, 0x72, 0x66,
	0x61, 0x63, 0x65, 0x12, 0x16, 0x0a, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x18, 0x06, 0x20,
	0x01, 0x28, 0x0d, 0x52, 0x06, 0x77, 0x65, 0x69, 0x67, 0x68, 0x74, 0x12, 0x1e, 0x0a, 0x0a, 0x70,
	0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x18, 0x07, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x0a, 0x70, 0x72, 0x65, 0x66, 0x65, 0x72, 0x65, 0x6e, 0x63, 0x65, 0x12, 0x1c, 0x0a, 0x0a, 0x76,
	0x69, 0x61, 0x5f, 0x76, 0x72, 0x66, 0x5f, 0x69, 0x64, 0x18, 0x08, 0x20, 0x01, 0x28, 0x0d, 0x52,
	0x08, 0x76, 0x69, 0x61, 0x56, 0x72, 0x66, 0x49, 0x64, 0x22, 0x33, 0x0a, 0x09, 0x52, 0x6f, 0x75,
	0x74, 0x65, 0x54, 0x79, 0x70, 0x65, 0x12, 0x0d, 0x0a, 0x09, 0x49, 0x4e, 0x54, 0x52, 0x41, 0x5f,
	0x56, 0x52, 0x46, 0x10, 0x00, 0x12, 0x0d, 0x0a, 0x09, 0x49, 0x4e, 0x54, 0x45, 0x52, 0x5f, 0x56,
	0x52, 0x46, 0x10, 0x01, 0x12, 0x08, 0x0a, 0x04, 0x44, 0x52, 0x4f, 0x50, 0x10, 0x02, 0x42, 0x36,
	0x5a, 0x34, 0x67, 0x6f, 0x2e, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2e, 0x69, 0x6f, 0x2f, 0x76,
	0x70, 0x70, 0x2d, 0x61, 0x67, 0x65, 0x6e, 0x74, 0x2f, 0x76, 0x33, 0x2f, 0x70, 0x72, 0x6f, 0x74,
	0x6f, 0x2f, 0x6c, 0x69, 0x67, 0x61, 0x74, 0x6f, 0x2f, 0x76, 0x70, 0x70, 0x2f, 0x6c, 0x33, 0x3b,
	0x76, 0x70, 0x70, 0x5f, 0x6c, 0x33, 0x62, 0x06, 0x70, 0x72, 0x6f, 0x74, 0x6f, 0x33,
}

var (
	file_ligato_vpp_l3_route_proto_rawDescOnce sync.Once
	file_ligato_vpp_l3_route_proto_rawDescData = file_ligato_vpp_l3_route_proto_rawDesc
)

func file_ligato_vpp_l3_route_proto_rawDescGZIP() []byte {
	file_ligato_vpp_l3_route_proto_rawDescOnce.Do(func() {
		file_ligato_vpp_l3_route_proto_rawDescData = protoimpl.X.CompressGZIP(file_ligato_vpp_l3_route_proto_rawDescData)
	})
	return file_ligato_vpp_l3_route_proto_rawDescData
}

var file_ligato_vpp_l3_route_proto_enumTypes = make([]protoimpl.EnumInfo, 1)
var file_ligato_vpp_l3_route_proto_msgTypes = make([]protoimpl.MessageInfo, 1)
var file_ligato_vpp_l3_route_proto_goTypes = []interface{}{
	(Route_RouteType)(0), // 0: ligato.vpp.l3.Route.RouteType
	(*Route)(nil),        // 1: ligato.vpp.l3.Route
}
var file_ligato_vpp_l3_route_proto_depIdxs = []int32{
	0, // 0: ligato.vpp.l3.Route.type:type_name -> ligato.vpp.l3.Route.RouteType
	1, // [1:1] is the sub-list for method output_type
	1, // [1:1] is the sub-list for method input_type
	1, // [1:1] is the sub-list for extension type_name
	1, // [1:1] is the sub-list for extension extendee
	0, // [0:1] is the sub-list for field type_name
}

func init() { file_ligato_vpp_l3_route_proto_init() }
func file_ligato_vpp_l3_route_proto_init() {
	if File_ligato_vpp_l3_route_proto != nil {
		return
	}
	if !protoimpl.UnsafeEnabled {
		file_ligato_vpp_l3_route_proto_msgTypes[0].Exporter = func(v interface{}, i int) interface{} {
			switch v := v.(*Route); i {
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
			RawDescriptor: file_ligato_vpp_l3_route_proto_rawDesc,
			NumEnums:      1,
			NumMessages:   1,
			NumExtensions: 0,
			NumServices:   0,
		},
		GoTypes:           file_ligato_vpp_l3_route_proto_goTypes,
		DependencyIndexes: file_ligato_vpp_l3_route_proto_depIdxs,
		EnumInfos:         file_ligato_vpp_l3_route_proto_enumTypes,
		MessageInfos:      file_ligato_vpp_l3_route_proto_msgTypes,
	}.Build()
	File_ligato_vpp_l3_route_proto = out.File
	file_ligato_vpp_l3_route_proto_rawDesc = nil
	file_ligato_vpp_l3_route_proto_goTypes = nil
	file_ligato_vpp_l3_route_proto_depIdxs = nil
}