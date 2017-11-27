// Code generated by protoc-gen-go.
// source: isd_inv_info_container.proto
// DO NOT EDIT!

/*
Package cisco_ios_xr_spirit_install_instmgr_oper_software_install_issu_inventory is a generated protocol buffer package.

It is generated from these files:
	isd_inv_info_container.proto

It has these top-level messages:
	IsdInvInfoContainer_KEYS
	IsdInvInfoContainer
	IsdInvInfoSt
*/
package cisco_ios_xr_spirit_install_instmgr_oper_software_install_issu_inventory

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

// ISSU Inventory Information all nodes
type IsdInvInfoContainer_KEYS struct {
}

func (m *IsdInvInfoContainer_KEYS) Reset()                    { *m = IsdInvInfoContainer_KEYS{} }
func (m *IsdInvInfoContainer_KEYS) String() string            { return proto.CompactTextString(m) }
func (*IsdInvInfoContainer_KEYS) ProtoMessage()               {}
func (*IsdInvInfoContainer_KEYS) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type IsdInvInfoContainer struct {
	Invinfo []*IsdInvInfoSt `protobuf:"bytes,50,rep,name=invinfo" json:"invinfo,omitempty"`
}

func (m *IsdInvInfoContainer) Reset()                    { *m = IsdInvInfoContainer{} }
func (m *IsdInvInfoContainer) String() string            { return proto.CompactTextString(m) }
func (*IsdInvInfoContainer) ProtoMessage()               {}
func (*IsdInvInfoContainer) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *IsdInvInfoContainer) GetInvinfo() []*IsdInvInfoSt {
	if m != nil {
		return m.Invinfo
	}
	return nil
}

type IsdInvInfoSt struct {
	// Node ID
	NodeId int32 `protobuf:"zigzag32,1,opt,name=node_id,json=nodeId" json:"node_id,omitempty"`
	// Node Type
	NodeType string `protobuf:"bytes,2,opt,name=node_type,json=nodeType" json:"node_type,omitempty"`
	// ISSU Node Role
	IssuNodeRole string `protobuf:"bytes,3,opt,name=issu_node_role,json=issuNodeRole" json:"issu_node_role,omitempty"`
	// Node State
	NodeState string `protobuf:"bytes,4,opt,name=node_state,json=nodeState" json:"node_state,omitempty"`
	// Node role
	NodeRole string `protobuf:"bytes,5,opt,name=node_role,json=nodeRole" json:"node_role,omitempty"`
}

func (m *IsdInvInfoSt) Reset()                    { *m = IsdInvInfoSt{} }
func (m *IsdInvInfoSt) String() string            { return proto.CompactTextString(m) }
func (*IsdInvInfoSt) ProtoMessage()               {}
func (*IsdInvInfoSt) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *IsdInvInfoSt) GetNodeId() int32 {
	if m != nil {
		return m.NodeId
	}
	return 0
}

func (m *IsdInvInfoSt) GetNodeType() string {
	if m != nil {
		return m.NodeType
	}
	return ""
}

func (m *IsdInvInfoSt) GetIssuNodeRole() string {
	if m != nil {
		return m.IssuNodeRole
	}
	return ""
}

func (m *IsdInvInfoSt) GetNodeState() string {
	if m != nil {
		return m.NodeState
	}
	return ""
}

func (m *IsdInvInfoSt) GetNodeRole() string {
	if m != nil {
		return m.NodeRole
	}
	return ""
}

func init() {
	proto.RegisterType((*IsdInvInfoContainer_KEYS)(nil), "cisco_ios_xr_spirit_install_instmgr_oper.software_install.issu.inventory.isd_inv_info_container_KEYS")
	proto.RegisterType((*IsdInvInfoContainer)(nil), "cisco_ios_xr_spirit_install_instmgr_oper.software_install.issu.inventory.isd_inv_info_container")
	proto.RegisterType((*IsdInvInfoSt)(nil), "cisco_ios_xr_spirit_install_instmgr_oper.software_install.issu.inventory.isd_inv_info_st")
}

func init() { proto.RegisterFile("isd_inv_info_container.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 268 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xac, 0x91, 0xc1, 0x4a, 0x03, 0x31,
	0x10, 0x86, 0x59, 0xab, 0xad, 0x8d, 0xa2, 0x98, 0x83, 0x2e, 0xd4, 0x42, 0x59, 0x3c, 0xec, 0x29,
	0x87, 0xfa, 0x0c, 0x82, 0x22, 0x78, 0xd8, 0x7a, 0xe9, 0x69, 0x58, 0x77, 0xa7, 0x32, 0xb0, 0x66,
	0x42, 0x26, 0xae, 0xee, 0x43, 0xf8, 0x20, 0xbe, 0xa5, 0x24, 0x50, 0x45, 0xe9, 0xd1, 0x53, 0xc8,
	0xff, 0x7f, 0x93, 0x2f, 0x30, 0xea, 0x92, 0xa4, 0x05, 0xb2, 0x3d, 0x90, 0xdd, 0x30, 0x34, 0x6c,
	0x43, 0x4d, 0x16, 0xbd, 0x71, 0x9e, 0x03, 0xeb, 0xdb, 0x86, 0xa4, 0x61, 0x20, 0x16, 0x78, 0xf7,
	0x20, 0x8e, 0x3c, 0x05, 0x20, 0x2b, 0xa1, 0xee, 0xba, 0x74, 0xbe, 0x3c, 0x7b, 0x60, 0x87, 0xde,
	0x08, 0x6f, 0xc2, 0x5b, 0xed, 0x71, 0xdb, 0x1a, 0x12, 0x79, 0x35, 0x64, 0x7b, 0xb4, 0x81, 0xfd,
	0x50, 0xcc, 0xd5, 0x6c, 0xb7, 0x09, 0xee, 0x6f, 0xd6, 0xab, 0xe2, 0x23, 0x53, 0xe7, 0xbb, 0x7b,
	0x2d, 0x6a, 0x42, 0xb6, 0x8f, 0x61, 0xbe, 0x5c, 0x8c, 0xca, 0xa3, 0xe5, 0xda, 0xfc, 0xd7, 0xaf,
	0xcc, 0x2f, 0xa5, 0x84, 0x6a, 0x6b, 0x2a, 0x3e, 0x33, 0x75, 0xfa, 0xa7, 0xd4, 0x17, 0x6a, 0x62,
	0xb9, 0x45, 0xa0, 0x36, 0xcf, 0x16, 0x59, 0x79, 0x56, 0x8d, 0xe3, 0xf5, 0xae, 0xd5, 0x33, 0x35,
	0x4d, 0x45, 0x18, 0x1c, 0xe6, 0x7b, 0x8b, 0xac, 0x9c, 0x56, 0x87, 0x31, 0x78, 0x1c, 0x1c, 0xea,
	0x2b, 0x75, 0x12, 0xa5, 0x90, 0x08, 0xcf, 0x1d, 0xe6, 0xa3, 0x44, 0x1c, 0xc7, 0xf4, 0x81, 0x5b,
	0xac, 0xb8, 0x43, 0x3d, 0x57, 0x2a, 0x01, 0x12, 0xea, 0x80, 0xf9, 0x7e, 0x22, 0xd2, 0xa3, 0xab,
	0x18, 0x7c, 0x1b, 0xd2, 0xfc, 0xc1, 0x8f, 0x21, 0xce, 0x3e, 0x8d, 0xd3, 0xae, 0xae, 0xbf, 0x02,
	0x00, 0x00, 0xff, 0xff, 0xfc, 0x09, 0x05, 0x97, 0xcb, 0x01, 0x00, 0x00,
}