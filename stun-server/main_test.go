package main

import (
	"encoding/binary"
	"fmt"
	"reflect"
	"testing"
)

const MagicCookie uint32 = 0x2112A442
const MethodBinding uint16 = 0x001

type StunHeader struct { // 20 bytes
	Type          uint16   // 2 bytes
	Length        uint16   // 2 bytes
	MessageCookie uint32   // 4 bytes
	TransactionID [12]byte // 12 bytes
}

func encodeToBytes(sh StunHeader) []byte {
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
		{input: encodeToBytes(StunHeader{Type: 0x1234,
			Length:        0,
			MessageCookie: MagicCookie,
			TransactionID: [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}}),
			want: fmt.Errorf("Request is type 1234, should be Binding Request (0x0001)")},
		{input: encodeToBytes(StunHeader{Type: MethodBinding,
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
