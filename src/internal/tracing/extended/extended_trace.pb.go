// Code generated by protoc-gen-gogo. DO NOT EDIT.
// source: internal/tracing/extended/extended_trace.proto

package extended

import (
	fmt "fmt"
	proto "github.com/gogo/protobuf/proto"
	io "io"
	math "math"
	math_bits "math/bits"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.GoGoProtoPackageIsVersion3 // please upgrade the proto package

// TraceProto contains information identifying a Jaeger trace. It's used to
// propagate traces that follow the lifetime of a long operation (e.g. creating
// a pipeline or running a job), and which live longer than any single RPC.
type TraceProto struct {
	// serialized_trace contains the info identifying a trace in Jaeger (a
	// (trace ID, span ID, sampled) tuple, basically)
	SerializedTrace map[string]string `protobuf:"bytes,1,rep,name=serialized_trace,json=serializedTrace,proto3" json:"serialized_trace,omitempty" protobuf_key:"bytes,1,opt,name=key,proto3" protobuf_val:"bytes,2,opt,name=value,proto3"`
	// key is an identifying key of this trace. For example, for a trace created
	// by 'pachctl create-pipeline', this would be the pipeline name and project;
	// future PPS master and worker events would be associated with the same
	// pipeline.
	Key                  string   `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
	XXX_NoUnkeyedLiteral struct{} `json:"-"`
	XXX_unrecognized     []byte   `json:"-"`
	XXX_sizecache        int32    `json:"-"`
}

func (m *TraceProto) Reset()         { *m = TraceProto{} }
func (m *TraceProto) String() string { return proto.CompactTextString(m) }
func (*TraceProto) ProtoMessage()    {}
func (*TraceProto) Descriptor() ([]byte, []int) {
	return fileDescriptor_041fc5114ea6a2f6, []int{0}
}
func (m *TraceProto) XXX_Unmarshal(b []byte) error {
	return m.Unmarshal(b)
}
func (m *TraceProto) XXX_Marshal(b []byte, deterministic bool) ([]byte, error) {
	if deterministic {
		return xxx_messageInfo_TraceProto.Marshal(b, m, deterministic)
	} else {
		b = b[:cap(b)]
		n, err := m.MarshalToSizedBuffer(b)
		if err != nil {
			return nil, err
		}
		return b[:n], nil
	}
}
func (m *TraceProto) XXX_Merge(src proto.Message) {
	xxx_messageInfo_TraceProto.Merge(m, src)
}
func (m *TraceProto) XXX_Size() int {
	return m.Size()
}
func (m *TraceProto) XXX_DiscardUnknown() {
	xxx_messageInfo_TraceProto.DiscardUnknown(m)
}

var xxx_messageInfo_TraceProto proto.InternalMessageInfo

func (m *TraceProto) GetSerializedTrace() map[string]string {
	if m != nil {
		return m.SerializedTrace
	}
	return nil
}

func (m *TraceProto) GetKey() string {
	if m != nil {
		return m.Key
	}
	return ""
}

func init() {
	proto.RegisterType((*TraceProto)(nil), "extended.TraceProto")
	proto.RegisterMapType((map[string]string)(nil), "extended.TraceProto.SerializedTraceEntry")
}

func init() {
	proto.RegisterFile("internal/tracing/extended/extended_trace.proto", fileDescriptor_041fc5114ea6a2f6)
}

var fileDescriptor_041fc5114ea6a2f6 = []byte{
	// 217 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0xe2, 0xd2, 0xcb, 0xcc, 0x2b, 0x49,
	0x2d, 0xca, 0x4b, 0xcc, 0xd1, 0x2f, 0x29, 0x4a, 0x4c, 0xce, 0xcc, 0x4b, 0xd7, 0x4f, 0xad, 0x28,
	0x49, 0xcd, 0x4b, 0x49, 0x4d, 0x81, 0x33, 0xe2, 0x41, 0x32, 0xa9, 0x7a, 0x05, 0x45, 0xf9, 0x25,
	0xf9, 0x42, 0x1c, 0x30, 0x51, 0xa5, 0x1d, 0x8c, 0x5c, 0x5c, 0x21, 0x20, 0x99, 0x00, 0xb0, 0x44,
	0x08, 0x97, 0x40, 0x71, 0x6a, 0x51, 0x66, 0x62, 0x4e, 0x66, 0x15, 0x4c, 0x8b, 0x04, 0xa3, 0x02,
	0xb3, 0x06, 0xb7, 0x91, 0xa6, 0x1e, 0x4c, 0x8f, 0x1e, 0x42, 0xbd, 0x5e, 0x30, 0x5c, 0x31, 0x58,
	0xd0, 0x35, 0xaf, 0xa4, 0xa8, 0x32, 0x88, 0xbf, 0x18, 0x55, 0x54, 0x48, 0x80, 0x8b, 0x39, 0x3b,
	0xb5, 0x52, 0x82, 0x49, 0x81, 0x51, 0x83, 0x33, 0x08, 0xc4, 0x94, 0x72, 0xe2, 0x12, 0xc1, 0xa6,
	0x15, 0xa6, 0x92, 0x11, 0xae, 0x52, 0x48, 0x84, 0x8b, 0xb5, 0x2c, 0x31, 0xa7, 0x34, 0x15, 0xaa,
	0x1b, 0xc2, 0xb1, 0x62, 0xb2, 0x60, 0x74, 0xf2, 0x3d, 0xf1, 0x48, 0x8e, 0xf1, 0xc2, 0x23, 0x39,
	0xc6, 0x07, 0x8f, 0xe4, 0x18, 0xa3, 0xec, 0xd3, 0x33, 0x4b, 0x32, 0x4a, 0x93, 0xf4, 0x92, 0xf3,
	0x73, 0xf5, 0x0b, 0x12, 0x93, 0x33, 0x2a, 0x53, 0x52, 0x8b, 0x90, 0x59, 0x65, 0x46, 0xfa, 0xc5,
	0x45, 0xc9, 0xfa, 0x38, 0x03, 0x2a, 0x89, 0x0d, 0x1c, 0x34, 0xc6, 0x80, 0x00, 0x00, 0x00, 0xff,
	0xff, 0x0c, 0x38, 0xf0, 0xdb, 0x4c, 0x01, 0x00, 0x00,
}

func (m *TraceProto) Marshal() (dAtA []byte, err error) {
	size := m.Size()
	dAtA = make([]byte, size)
	n, err := m.MarshalToSizedBuffer(dAtA[:size])
	if err != nil {
		return nil, err
	}
	return dAtA[:n], nil
}

func (m *TraceProto) MarshalTo(dAtA []byte) (int, error) {
	size := m.Size()
	return m.MarshalToSizedBuffer(dAtA[:size])
}

func (m *TraceProto) MarshalToSizedBuffer(dAtA []byte) (int, error) {
	i := len(dAtA)
	_ = i
	var l int
	_ = l
	if m.XXX_unrecognized != nil {
		i -= len(m.XXX_unrecognized)
		copy(dAtA[i:], m.XXX_unrecognized)
	}
	if len(m.Key) > 0 {
		i -= len(m.Key)
		copy(dAtA[i:], m.Key)
		i = encodeVarintExtendedTrace(dAtA, i, uint64(len(m.Key)))
		i--
		dAtA[i] = 0x12
	}
	if len(m.SerializedTrace) > 0 {
		for k := range m.SerializedTrace {
			v := m.SerializedTrace[k]
			baseI := i
			i -= len(v)
			copy(dAtA[i:], v)
			i = encodeVarintExtendedTrace(dAtA, i, uint64(len(v)))
			i--
			dAtA[i] = 0x12
			i -= len(k)
			copy(dAtA[i:], k)
			i = encodeVarintExtendedTrace(dAtA, i, uint64(len(k)))
			i--
			dAtA[i] = 0xa
			i = encodeVarintExtendedTrace(dAtA, i, uint64(baseI-i))
			i--
			dAtA[i] = 0xa
		}
	}
	return len(dAtA) - i, nil
}

func encodeVarintExtendedTrace(dAtA []byte, offset int, v uint64) int {
	offset -= sovExtendedTrace(v)
	base := offset
	for v >= 1<<7 {
		dAtA[offset] = uint8(v&0x7f | 0x80)
		v >>= 7
		offset++
	}
	dAtA[offset] = uint8(v)
	return base
}
func (m *TraceProto) Size() (n int) {
	if m == nil {
		return 0
	}
	var l int
	_ = l
	if len(m.SerializedTrace) > 0 {
		for k, v := range m.SerializedTrace {
			_ = k
			_ = v
			mapEntrySize := 1 + len(k) + sovExtendedTrace(uint64(len(k))) + 1 + len(v) + sovExtendedTrace(uint64(len(v)))
			n += mapEntrySize + 1 + sovExtendedTrace(uint64(mapEntrySize))
		}
	}
	l = len(m.Key)
	if l > 0 {
		n += 1 + l + sovExtendedTrace(uint64(l))
	}
	if m.XXX_unrecognized != nil {
		n += len(m.XXX_unrecognized)
	}
	return n
}

func sovExtendedTrace(x uint64) (n int) {
	return (math_bits.Len64(x|1) + 6) / 7
}
func sozExtendedTrace(x uint64) (n int) {
	return sovExtendedTrace(uint64((x << 1) ^ uint64((int64(x) >> 63))))
}
func (m *TraceProto) Unmarshal(dAtA []byte) error {
	l := len(dAtA)
	iNdEx := 0
	for iNdEx < l {
		preIndex := iNdEx
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return ErrIntOverflowExtendedTrace
			}
			if iNdEx >= l {
				return io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= uint64(b&0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		fieldNum := int32(wire >> 3)
		wireType := int(wire & 0x7)
		if wireType == 4 {
			return fmt.Errorf("proto: TraceProto: wiretype end group for non-group")
		}
		if fieldNum <= 0 {
			return fmt.Errorf("proto: TraceProto: illegal tag %d (wire type %d)", fieldNum, wire)
		}
		switch fieldNum {
		case 1:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field SerializedTrace", wireType)
			}
			var msglen int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowExtendedTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				msglen |= int(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if msglen < 0 {
				return ErrInvalidLengthExtendedTrace
			}
			postIndex := iNdEx + msglen
			if postIndex < 0 {
				return ErrInvalidLengthExtendedTrace
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			if m.SerializedTrace == nil {
				m.SerializedTrace = make(map[string]string)
			}
			var mapkey string
			var mapvalue string
			for iNdEx < postIndex {
				entryPreIndex := iNdEx
				var wire uint64
				for shift := uint(0); ; shift += 7 {
					if shift >= 64 {
						return ErrIntOverflowExtendedTrace
					}
					if iNdEx >= l {
						return io.ErrUnexpectedEOF
					}
					b := dAtA[iNdEx]
					iNdEx++
					wire |= uint64(b&0x7F) << shift
					if b < 0x80 {
						break
					}
				}
				fieldNum := int32(wire >> 3)
				if fieldNum == 1 {
					var stringLenmapkey uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowExtendedTrace
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapkey |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapkey := int(stringLenmapkey)
					if intStringLenmapkey < 0 {
						return ErrInvalidLengthExtendedTrace
					}
					postStringIndexmapkey := iNdEx + intStringLenmapkey
					if postStringIndexmapkey < 0 {
						return ErrInvalidLengthExtendedTrace
					}
					if postStringIndexmapkey > l {
						return io.ErrUnexpectedEOF
					}
					mapkey = string(dAtA[iNdEx:postStringIndexmapkey])
					iNdEx = postStringIndexmapkey
				} else if fieldNum == 2 {
					var stringLenmapvalue uint64
					for shift := uint(0); ; shift += 7 {
						if shift >= 64 {
							return ErrIntOverflowExtendedTrace
						}
						if iNdEx >= l {
							return io.ErrUnexpectedEOF
						}
						b := dAtA[iNdEx]
						iNdEx++
						stringLenmapvalue |= uint64(b&0x7F) << shift
						if b < 0x80 {
							break
						}
					}
					intStringLenmapvalue := int(stringLenmapvalue)
					if intStringLenmapvalue < 0 {
						return ErrInvalidLengthExtendedTrace
					}
					postStringIndexmapvalue := iNdEx + intStringLenmapvalue
					if postStringIndexmapvalue < 0 {
						return ErrInvalidLengthExtendedTrace
					}
					if postStringIndexmapvalue > l {
						return io.ErrUnexpectedEOF
					}
					mapvalue = string(dAtA[iNdEx:postStringIndexmapvalue])
					iNdEx = postStringIndexmapvalue
				} else {
					iNdEx = entryPreIndex
					skippy, err := skipExtendedTrace(dAtA[iNdEx:])
					if err != nil {
						return err
					}
					if (skippy < 0) || (iNdEx+skippy) < 0 {
						return ErrInvalidLengthExtendedTrace
					}
					if (iNdEx + skippy) > postIndex {
						return io.ErrUnexpectedEOF
					}
					iNdEx += skippy
				}
			}
			m.SerializedTrace[mapkey] = mapvalue
			iNdEx = postIndex
		case 2:
			if wireType != 2 {
				return fmt.Errorf("proto: wrong wireType = %d for field Key", wireType)
			}
			var stringLen uint64
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return ErrIntOverflowExtendedTrace
				}
				if iNdEx >= l {
					return io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				stringLen |= uint64(b&0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			intStringLen := int(stringLen)
			if intStringLen < 0 {
				return ErrInvalidLengthExtendedTrace
			}
			postIndex := iNdEx + intStringLen
			if postIndex < 0 {
				return ErrInvalidLengthExtendedTrace
			}
			if postIndex > l {
				return io.ErrUnexpectedEOF
			}
			m.Key = string(dAtA[iNdEx:postIndex])
			iNdEx = postIndex
		default:
			iNdEx = preIndex
			skippy, err := skipExtendedTrace(dAtA[iNdEx:])
			if err != nil {
				return err
			}
			if (skippy < 0) || (iNdEx+skippy) < 0 {
				return ErrInvalidLengthExtendedTrace
			}
			if (iNdEx + skippy) > l {
				return io.ErrUnexpectedEOF
			}
			m.XXX_unrecognized = append(m.XXX_unrecognized, dAtA[iNdEx:iNdEx+skippy]...)
			iNdEx += skippy
		}
	}

	if iNdEx > l {
		return io.ErrUnexpectedEOF
	}
	return nil
}
func skipExtendedTrace(dAtA []byte) (n int, err error) {
	l := len(dAtA)
	iNdEx := 0
	depth := 0
	for iNdEx < l {
		var wire uint64
		for shift := uint(0); ; shift += 7 {
			if shift >= 64 {
				return 0, ErrIntOverflowExtendedTrace
			}
			if iNdEx >= l {
				return 0, io.ErrUnexpectedEOF
			}
			b := dAtA[iNdEx]
			iNdEx++
			wire |= (uint64(b) & 0x7F) << shift
			if b < 0x80 {
				break
			}
		}
		wireType := int(wire & 0x7)
		switch wireType {
		case 0:
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowExtendedTrace
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				iNdEx++
				if dAtA[iNdEx-1] < 0x80 {
					break
				}
			}
		case 1:
			iNdEx += 8
		case 2:
			var length int
			for shift := uint(0); ; shift += 7 {
				if shift >= 64 {
					return 0, ErrIntOverflowExtendedTrace
				}
				if iNdEx >= l {
					return 0, io.ErrUnexpectedEOF
				}
				b := dAtA[iNdEx]
				iNdEx++
				length |= (int(b) & 0x7F) << shift
				if b < 0x80 {
					break
				}
			}
			if length < 0 {
				return 0, ErrInvalidLengthExtendedTrace
			}
			iNdEx += length
		case 3:
			depth++
		case 4:
			if depth == 0 {
				return 0, ErrUnexpectedEndOfGroupExtendedTrace
			}
			depth--
		case 5:
			iNdEx += 4
		default:
			return 0, fmt.Errorf("proto: illegal wireType %d", wireType)
		}
		if iNdEx < 0 {
			return 0, ErrInvalidLengthExtendedTrace
		}
		if depth == 0 {
			return iNdEx, nil
		}
	}
	return 0, io.ErrUnexpectedEOF
}

var (
	ErrInvalidLengthExtendedTrace        = fmt.Errorf("proto: negative length found during unmarshaling")
	ErrIntOverflowExtendedTrace          = fmt.Errorf("proto: integer overflow")
	ErrUnexpectedEndOfGroupExtendedTrace = fmt.Errorf("proto: unexpected end of group")
)
