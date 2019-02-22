// Code generated by protoc-gen-go. DO NOT EDIT.
// source: google/ads/googleads/v0/enums/ad_group_type.proto

package enums // import "google.golang.org/genproto/googleapis/ads/googleads/v0/enums"

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

// Enum listing the possible types of an ad group.
type AdGroupTypeEnum_AdGroupType int32

const (
	// The type has not been specified.
	AdGroupTypeEnum_UNSPECIFIED AdGroupTypeEnum_AdGroupType = 0
	// The received value is not known in this version.
	//
	// This is a response-only value.
	AdGroupTypeEnum_UNKNOWN AdGroupTypeEnum_AdGroupType = 1
	// The default ad group type for Search campaigns.
	AdGroupTypeEnum_SEARCH_STANDARD AdGroupTypeEnum_AdGroupType = 2
	// The default ad group type for Display campaigns.
	AdGroupTypeEnum_DISPLAY_STANDARD AdGroupTypeEnum_AdGroupType = 3
	// The ad group type for Shopping campaigns serving standard product ads.
	AdGroupTypeEnum_SHOPPING_PRODUCT_ADS AdGroupTypeEnum_AdGroupType = 4
	// The default ad group type for Hotel campaigns.
	AdGroupTypeEnum_HOTEL_ADS AdGroupTypeEnum_AdGroupType = 6
	// The type for ad groups in Smart Shopping campaigns.
	AdGroupTypeEnum_SHOPPING_SMART_ADS AdGroupTypeEnum_AdGroupType = 7
	// Short unskippable in-stream video ads.
	AdGroupTypeEnum_VIDEO_BUMPER AdGroupTypeEnum_AdGroupType = 8
	// TrueView (skippable) in-stream video ads.
	AdGroupTypeEnum_VIDEO_TRUE_VIEW_IN_STREAM AdGroupTypeEnum_AdGroupType = 9
	// TrueView in-display video ads.
	AdGroupTypeEnum_VIDEO_TRUE_VIEW_IN_DISPLAY AdGroupTypeEnum_AdGroupType = 10
	// Unskippable in-stream video ads.
	AdGroupTypeEnum_VIDEO_NON_SKIPPABLE_IN_STREAM AdGroupTypeEnum_AdGroupType = 11
	// Outstream video ads.
	AdGroupTypeEnum_VIDEO_OUTSTREAM AdGroupTypeEnum_AdGroupType = 12
)

var AdGroupTypeEnum_AdGroupType_name = map[int32]string{
	0:  "UNSPECIFIED",
	1:  "UNKNOWN",
	2:  "SEARCH_STANDARD",
	3:  "DISPLAY_STANDARD",
	4:  "SHOPPING_PRODUCT_ADS",
	6:  "HOTEL_ADS",
	7:  "SHOPPING_SMART_ADS",
	8:  "VIDEO_BUMPER",
	9:  "VIDEO_TRUE_VIEW_IN_STREAM",
	10: "VIDEO_TRUE_VIEW_IN_DISPLAY",
	11: "VIDEO_NON_SKIPPABLE_IN_STREAM",
	12: "VIDEO_OUTSTREAM",
}
var AdGroupTypeEnum_AdGroupType_value = map[string]int32{
	"UNSPECIFIED":                   0,
	"UNKNOWN":                       1,
	"SEARCH_STANDARD":               2,
	"DISPLAY_STANDARD":              3,
	"SHOPPING_PRODUCT_ADS":          4,
	"HOTEL_ADS":                     6,
	"SHOPPING_SMART_ADS":            7,
	"VIDEO_BUMPER":                  8,
	"VIDEO_TRUE_VIEW_IN_STREAM":     9,
	"VIDEO_TRUE_VIEW_IN_DISPLAY":    10,
	"VIDEO_NON_SKIPPABLE_IN_STREAM": 11,
	"VIDEO_OUTSTREAM":               12,
}

func (x AdGroupTypeEnum_AdGroupType) String() string {
	return proto.EnumName(AdGroupTypeEnum_AdGroupType_name, int32(x))
}
func (AdGroupTypeEnum_AdGroupType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor_ad_group_type_3b9233b105e8c706, []int{0, 0}
}

// Defines types of an ad group, specific to a particular campaign channel
// type. This type drives validations that restrict which entities can be
// added to the ad group.
type AdGroupTypeEnum struct {
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *AdGroupTypeEnum) Reset()         { *m = AdGroupTypeEnum{} }
func (m *AdGroupTypeEnum) String() string { return proto.CompactTextString(m) }
func (*AdGroupTypeEnum) ProtoMessage()    {}
func (*AdGroupTypeEnum) Descriptor() ([]byte, []int) {
	return fileDescriptor_ad_group_type_3b9233b105e8c706, []int{0}
}
func (m *AdGroupTypeEnum) XXX_Unmarshal(b []byte) error {
	return xxx_messageInfo_AdGroupTypeEnum.Unmarshal(m, b)
}
func (m *AdGroupTypeEnum) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	return xxx_messageInfo_AdGroupTypeEnum.Marshal(b, m, deterministic)
}
func (dst *AdGroupTypeEnum) XXX_Merge(src proto.Message) {
	xxx_messageInfo_AdGroupTypeEnum.Merge(dst, src)
}
func (m *AdGroupTypeEnum) XXX_Size() int {
	return xxx_messageInfo_AdGroupTypeEnum.Size(m)
}
func (m *AdGroupTypeEnum) XXX_DiscardUnknown() {
	xxx_messageInfo_AdGroupTypeEnum.DiscardUnknown(m)
}

var xxx_messageInfo_AdGroupTypeEnum proto.InternalMessageInfo

func init() {
	proto.RegisterType((*AdGroupTypeEnum)(nil), "google.ads.googleads.v0.enums.AdGroupTypeEnum")
	proto.RegisterEnum("google.ads.googleads.v0.enums.AdGroupTypeEnum_AdGroupType", AdGroupTypeEnum_AdGroupType_name, AdGroupTypeEnum_AdGroupType_value)
}

func init() {
	proto.RegisterFile("google/ads/googleads/v0/enums/ad_group_type.proto", fileDescriptor_ad_group_type_3b9233b105e8c706)
}

var fileDescriptor_ad_group_type_3b9233b105e8c706 = []byte{
	// 395 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x7c, 0x51, 0xcf, 0x6e, 0xd3, 0x30,
	0x18, 0xa7, 0x19, 0xda, 0xd8, 0xd7, 0xa1, 0x5a, 0x66, 0x42, 0x80, 0x14, 0xa4, 0xed, 0x01, 0x9c,
	0x20, 0x8e, 0x9c, 0x9c, 0xc6, 0xa4, 0xd6, 0x5a, 0xc7, 0x8a, 0x93, 0x4c, 0xa0, 0x48, 0x56, 0x20,
	0x51, 0x84, 0xb4, 0xc6, 0x51, 0xb3, 0x4e, 0xda, 0xdb, 0x20, 0x8e, 0x9c, 0x79, 0x01, 0xae, 0x3c,
	0x15, 0x4a, 0xdc, 0x95, 0x1e, 0x80, 0x4b, 0xf4, 0xcb, 0xef, 0xcf, 0x67, 0xfb, 0xfb, 0xc1, 0x9b,
	0xc6, 0x98, 0xe6, 0xa6, 0xf6, 0xca, 0xaa, 0xf7, 0x2c, 0x1c, 0xd0, 0x9d, 0xef, 0xd5, 0xed, 0x76,
	0xdd, 0x7b, 0x65, 0xa5, 0x9b, 0x8d, 0xd9, 0x76, 0xfa, 0xf6, 0xbe, 0xab, 0x49, 0xb7, 0x31, 0xb7,
	0x06, 0xbb, 0xd6, 0x47, 0xca, 0xaa, 0x27, 0xfb, 0x08, 0xb9, 0xf3, 0xc9, 0x18, 0xb9, 0xfc, 0xe1,
	0xc0, 0x8c, 0x56, 0xd1, 0x90, 0x4a, 0xef, 0xbb, 0x9a, 0xb5, 0xdb, 0xf5, 0xe5, 0x57, 0x07, 0xa6,
	0x07, 0x1c, 0x9e, 0xc1, 0x34, 0x13, 0x4a, 0xb2, 0x39, 0x7f, 0xcf, 0x59, 0x88, 0x1e, 0xe1, 0x29,
	0x9c, 0x64, 0xe2, 0x4a, 0xc4, 0xd7, 0x02, 0x4d, 0xf0, 0x33, 0x98, 0x29, 0x46, 0x93, 0xf9, 0x42,
	0xab, 0x94, 0x8a, 0x90, 0x26, 0x21, 0x72, 0xf0, 0x39, 0xa0, 0x90, 0x2b, 0xb9, 0xa4, 0x1f, 0xfe,
	0xb0, 0x47, 0xf8, 0x05, 0x9c, 0xab, 0x45, 0x2c, 0x25, 0x17, 0x91, 0x96, 0x49, 0x1c, 0x66, 0xf3,
	0x54, 0xd3, 0x50, 0xa1, 0xc7, 0xf8, 0x29, 0x9c, 0x2e, 0xe2, 0x94, 0x2d, 0xc7, 0xdf, 0x63, 0xfc,
	0x1c, 0xf0, 0xde, 0xa8, 0x56, 0x34, 0xb1, 0xb6, 0x13, 0x8c, 0xe0, 0x2c, 0xe7, 0x21, 0x8b, 0x75,
	0x90, 0xad, 0x24, 0x4b, 0xd0, 0x13, 0xec, 0xc2, 0x4b, 0xcb, 0xa4, 0x49, 0xc6, 0x74, 0xce, 0xd9,
	0xb5, 0xe6, 0x42, 0xab, 0x34, 0x61, 0x74, 0x85, 0x4e, 0xf1, 0x6b, 0x78, 0xf5, 0x17, 0x79, 0x77,
	0x35, 0x04, 0xf8, 0x02, 0x5c, 0xab, 0x8b, 0x58, 0x68, 0x75, 0xc5, 0xa5, 0xa4, 0xc1, 0x92, 0x1d,
	0x8c, 0x98, 0x0e, 0xef, 0xb3, 0x96, 0x38, 0x4b, 0x77, 0xe4, 0x59, 0xf0, 0x73, 0x02, 0x17, 0x9f,
	0xcd, 0x9a, 0xfc, 0x77, 0xb9, 0x01, 0x3a, 0xd8, 0xa2, 0x1c, 0xda, 0x90, 0x93, 0x8f, 0xc1, 0x2e,
	0xd2, 0x98, 0x9b, 0xb2, 0x6d, 0x88, 0xd9, 0x34, 0x5e, 0x53, 0xb7, 0x63, 0x57, 0x0f, 0x95, 0x76,
	0x5f, 0xfa, 0x7f, 0x34, 0xfc, 0x6e, 0xfc, 0x7e, 0x73, 0x8e, 0x22, 0x4a, 0xbf, 0x3b, 0x6e, 0x64,
	0x47, 0xd1, 0xaa, 0x27, 0x16, 0x0e, 0x28, 0xf7, 0xc9, 0xd0, 0x62, 0xff, 0xeb, 0x41, 0x2f, 0x68,
	0xd5, 0x17, 0x7b, 0xbd, 0xc8, 0xfd, 0x62, 0xd4, 0x3f, 0x1d, 0x8f, 0x87, 0xbe, 0xfd, 0x1d, 0x00,
	0x00, 0xff, 0xff, 0xb8, 0x57, 0x8e, 0xbb, 0x55, 0x02, 0x00, 0x00,
}
