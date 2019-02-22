// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v0/resources/campaign_shared_set.proto

package resources // import "google.golang.org/genproto/googleapis/ads/googleads/v0/resources"

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import wrappers "github.com/golang/protobuf/ptypes/wrappers"
import enums "google.golang.org/genproto/googleapis/ads/googleads/v0/enums"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

// CampaignSharedSets are used for managing the shared sets associated with a
// campaign.
type CampaignSharedSet struct {
	// The resource name of the campaign shared set.
	// Campaign shared set resource names have the form:
	//
	// `customers/{customer_id}/campaignSharedSets/{campaign_id}_{shared_set_id}`
	ResourceName string `protobuf:"bytes,1,opt,name=resource_name,json=resourceName,proto3" json:"resource_name,omitempty"`
	// The campaign to which the campaign shared set belongs.
	Campaign *wrappers.StringValue `protobuf:"bytes,3,opt,name=campaign,proto3" json:"campaign,omitempty"`
	// The shared set associated with the campaign. This may be a negative keyword
	// shared set of another customer. This customer should be a manager of the
	// other customer, otherwise the campaign shared set will exist but have no
	// serving effect. Only negative keyword shared sets can be associated with
	// Shopping campaigns. Only negative placement shared sets can be associated
	// with Display mobile app campaigns.
	SharedSet *wrappers.StringValue `protobuf:"bytes,4,opt,name=shared_set,json=sharedSet,proto3" json:"shared_set,omitempty"`
	// The status of this campaign shared set. Read only.
	Status               enums.CampaignSharedSetStatusEnum_CampaignSharedSetStatus `protobuf:"varint,2,opt,name=status,proto3,enum=google.ads.googleads.v0.enums.CampaignSharedSetStatusEnum_CampaignSharedSetStatus" json:"status,omitempty"`
	XXX_NoUnkeyedLiteral struct{}                                                  `json:"-"`
	XXX_unrecognized     []byte                                                    `json:"-"`
	XXX_sizecache        int32                                                     `json:"-"`
}

func (m *CampaignSharedSet) Reset()         { *m = CampaignSharedSet{} }
func (m *CampaignSharedSet) String() string { return proto.CompactTextString(m) }
func (*CampaignSharedSet) ProtoMessage()    {}
func (*CampaignSharedSet) Descriptor() ([]byte, []int) {
	return fileDescriptor_campaign_shared_set_5be2acea5bdaff30, []int{0}
}
func (m *CampaignSharedSet) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_CampaignSharedSet.Unmarshal(m, b)
}
func (m *CampaignSharedSet) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_CampaignSharedSet.Marshal(b, m, deterministic)
}
func (dst *CampaignSharedSet) XXX_Merge(src proto.Message) {
	xxx_messageInfo_CampaignSharedSet.Merge(dst, src)
}
func (m *CampaignSharedSet) XXX_Size() int {
	return xxx_messageInfo_CampaignSharedSet.Size(m)
}
func (m *CampaignSharedSet) XXX_DiscardUnknown() {
	xxx_messageInfo_CampaignSharedSet.DiscardUnknown(m)
}

var xxx_messageInfo_CampaignSharedSet proto.InternalMessageInfo

func (m *CampaignSharedSet) GetResourceName() string {
	if m != nil {
		return m.ResourceName
	}
	return ""
}

func (m *CampaignSharedSet) GetCampaign() *wrappers.StringValue {
	if m != nil {
		return m.Campaign
	}
	return nil
}

func (m *CampaignSharedSet) GetSharedSet() *wrappers.StringValue {
	if m != nil {
		return m.SharedSet
	}
	return nil
}

func (m *CampaignSharedSet) GetStatus() enums.CampaignSharedSetStatusEnum_CampaignSharedSetStatus {
	if m != nil {
		return m.Status
	}
	return enums.CampaignSharedSetStatusEnum_UNSPECIFIED
}

func init() {
	proto.RegisterType((*CampaignSharedSet)(nil), "google.ads.googleads.v0.resources.CampaignSharedSet")
}

func init() {
	proto.RegisterFile("google/ads/googleads/v0/resources/campaign_shared_set.proto", fileDescriptor_campaign_shared_set_5be2acea5bdaff30)
}

var fileDescriptor_campaign_shared_set_5be2acea5bdaff30 = []byte{
	// 350 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x8c, 0x52, 0x4f, 0x4b, 0xc3, 0x30,
	0x1c, 0xa5, 0x9d, 0x0c, 0x17, 0xff, 0x80, 0x3d, 0x48, 0x19, 0x22, 0x9b, 0x22, 0xec, 0xf4, 0x6b,
	0x99, 0x17, 0x61, 0x20, 0x74, 0x22, 0x03, 0x0f, 0x32, 0x5a, 0xd8, 0x41, 0x0a, 0x25, 0x5b, 0x63,
	0x9c, 0xac, 0x49, 0x49, 0xda, 0xf9, 0x61, 0xbc, 0x79, 0xf4, 0xa3, 0xf8, 0x15, 0xfc, 0x32, 0xb2,
	0x36, 0x89, 0x87, 0x59, 0xf5, 0xf6, 0xda, 0xbc, 0xf7, 0x7e, 0xef, 0xe5, 0x17, 0x34, 0xa2, 0x9c,
	0xd3, 0x15, 0xf1, 0x70, 0x2a, 0xbd, 0x1a, 0x6e, 0xd0, 0xda, 0xf7, 0x04, 0x91, 0xbc, 0x14, 0x0b,
	0x22, 0xbd, 0x05, 0xce, 0x72, 0xbc, 0xa4, 0x2c, 0x91, 0x4f, 0x58, 0x90, 0x34, 0x91, 0xa4, 0x80,
	0x5c, 0xf0, 0x82, 0x3b, 0xfd, 0x5a, 0x01, 0x38, 0x95, 0x60, 0xc4, 0xb0, 0xf6, 0xc1, 0x88, 0xbb,
	0xd7, 0x4d, 0xfe, 0x84, 0x95, 0xd9, 0x8f, 0xde, 0x89, 0x2c, 0x70, 0x51, 0xca, 0x7a, 0x44, 0xf7,
	0x54, 0xe9, 0xab, 0xaf, 0x79, 0xf9, 0xe8, 0xbd, 0x08, 0x9c, 0xe7, 0x44, 0xa8, 0xf3, 0xb3, 0x57,
	0x1b, 0x1d, 0xdd, 0x28, 0x93, 0xa8, 0xf2, 0x88, 0x48, 0xe1, 0x9c, 0xa3, 0x03, 0x1d, 0x21, 0x61,
	0x38, 0x23, 0xae, 0xd5, 0xb3, 0x06, 0x9d, 0x70, 0x5f, 0xff, 0xbc, 0xc7, 0x19, 0x71, 0xae, 0xd0,
	0xae, 0x1e, 0xef, 0xb6, 0x7a, 0xd6, 0x60, 0x6f, 0x78, 0xa2, 0x5a, 0x80, 0x9e, 0x06, 0x51, 0x21,
	0x96, 0x8c, 0xce, 0xf0, 0xaa, 0x24, 0xa1, 0x61, 0x3b, 0x23, 0x84, 0xbe, 0xf3, 0xba, 0x3b, 0xff,
	0xd0, 0x76, 0xa4, 0xc9, 0xf6, 0x8c, 0xda, 0x75, 0x43, 0xd7, 0xee, 0x59, 0x83, 0xc3, 0x61, 0x08,
	0x4d, 0xb7, 0x58, 0x5d, 0x11, 0x6c, 0xb5, 0x8b, 0x2a, 0xf5, 0x2d, 0x2b, 0xb3, 0xa6, 0xb3, 0x50,
	0x4d, 0x18, 0x7f, 0x5a, 0xe8, 0x62, 0xc1, 0x33, 0xf8, 0x73, 0x4f, 0xe3, 0xe3, 0x2d, 0xab, 0xe9,
	0xa6, 0xc7, 0xd4, 0x7a, 0xb8, 0x53, 0x62, 0xca, 0x57, 0x98, 0x51, 0xe0, 0x82, 0x7a, 0x94, 0xb0,
	0xaa, 0xa5, 0xde, 0x68, 0xbe, 0x94, 0xbf, 0x3c, 0xa0, 0x91, 0x41, 0x6f, 0x76, 0x6b, 0x12, 0x04,
	0xef, 0x76, 0x7f, 0x52, 0x5b, 0x06, 0xa9, 0x84, 0x1a, 0x6e, 0xd0, 0xcc, 0x87, 0x50, 0x33, 0x3f,
	0x34, 0x27, 0x0e, 0x52, 0x19, 0x1b, 0x4e, 0x3c, 0xf3, 0x63, 0xc3, 0x99, 0xb7, 0xab, 0x10, 0x97,
	0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xbe, 0xc4, 0xbd, 0xff, 0xc4, 0x02, 0x00, 0x00,
}
