package main

import (
	"encoding/binary"
	"fmt"
	"net"
	"reflect"
	"testing"
)

// GO111MODULE=off CGO_ENABLED=0 go test

const MagicCookie uint32 = 0x2112A442
const MethodBinding uint16 = 0x001

type StunHeader struct { // 20 bytes
	Type          uint16   // 2 bytes
	Length        uint16   // 2 bytes
	MessageCookie uint32   // 4 bytes
	TransactionID [12]byte // 12 bytes
}

func encodeStunHeaderToBytes(sh StunHeader) []byte {
	buf := make([]byte, 20)
	binary.BigEndian.PutUint16(buf[0:], sh.Type)
	binary.BigEndian.PutUint16(buf[2:], sh.Length)
	binary.BigEndian.PutUint32(buf[4:], sh.MessageCookie)
	copy(buf[8:], sh.TransactionID[:])
	return buf
}

func TestValidateRequest(t *testing.T) {
	type test struct {
		input []byte
		want  error
	}

	tests := []test{
		{input: make([]byte, 20), want: fmt.Errorf("Request is type 0, should be Binding Request (0x0001)")},
		{input: make([]byte, 2), want: fmt.Errorf("Request is 2 bytes, should be 20")},
		{input: encodeStunHeaderToBytes(StunHeader{Type: 0x1234,
			Length:        0,
			MessageCookie: MagicCookie,
			TransactionID: [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}}),
			want: fmt.Errorf("Request is type 1234, should be Binding Request (0x0001)")},
		{input: encodeStunHeaderToBytes(StunHeader{Type: MethodBinding,
			Length:        0,
			MessageCookie: MagicCookie,
			TransactionID: [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}}), want: nil},
	}

	for _, tc := range tests {
		got := validateRequest(tc.input, len(tc.input))
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}

func printGolangBuffer(bt []byte) {
	s := ""
	for _, b := range bt {
		s += fmt.Sprintf("%#.2x", b)
		s += ","
	}
	fmt.Printf("[]byte{%s}\n", s)
}

func TestMakeResponse(t *testing.T) {
	type test struct {
		packet []byte
		addr   *net.UDPAddr
		want   []byte
	}

	// TODO: IPv6
	addr1, _ := net.ResolveUDPAddr("udp", ":0")
	addr2, _ := net.ResolveUDPAddr("udp", "127.0.0.1:22")
	addr3, _ := net.ResolveUDPAddr("udp", "255.255.255.255:65535")

	tests := []test{
		{
			packet: encodeStunHeaderToBytes(StunHeader{
				Type:          0x1234,
				Length:        0,
				MessageCookie: MagicCookie,
				TransactionID: [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			}),
			addr: addr1,
			want: []byte{0x01, 0x01, 0x00, 0x18, 0x21, 0x12, 0xa4, 0x42, 0x01, 0x02, 0x03,
				0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x00, 0x20,
				0x00, 0x14, 0x00, 0x02, 0x21, 0x12, 0x21, 0x12, 0xa4, 0x42, 0x01,
				0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
		},
		{
			packet: encodeStunHeaderToBytes(StunHeader{
				Type:          0x1001,
				Length:        0,
				MessageCookie: MagicCookie,
				TransactionID: [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			}),
			addr: addr2,
			want: []byte{0x01, 0x01, 0x00, 0x0c, 0x21, 0x12, 0xa4, 0x42, 0x01, 0x02, 0x03,
				0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x00, 0x20,
				0x00, 0x08, 0x00, 0x01, 0x21, 0x04, 0x5e, 0x12, 0xa4, 0x43},
		},
		{
			packet: encodeStunHeaderToBytes(StunHeader{
				Type:          0x1001,
				Length:        65535,
				MessageCookie: MagicCookie,
				TransactionID: [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12},
			}),
			addr: addr3,
			want: []byte{0x01, 0x01, 0x00, 0x0c, 0x21, 0x12, 0xa4, 0x42, 0x01, 0x02, 0x03, 0x04, 0x05,
				0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x00, 0x20, 0x00, 0x08, 0x00, 0x01,
				0xde, 0xed, 0xde, 0xed, 0x5b, 0xbd},
		},
	}

	for _, tc := range tests {
		got := makeResponse(tc.packet, tc.addr)
		//printGolangBuffer(got)
		if !reflect.DeepEqual(tc.want, got) {
			t.Fatalf("expected: %v, got: %v", tc.want, got)
		}
	}
}
