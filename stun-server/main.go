package main

import (
	"encoding/binary"
	"fmt"
	"net"
)

func isv4(ip []byte) bool {
	return binary.BigEndian.Uint32(ip[0:4]) == 0 && binary.BigEndian.Uint64(ip[4:12]) == 0xffff
}

func xorv4(ip []byte, messageCookie []byte) {
	for x := 0; x < 4; x++ {
		ip[x] ^= messageCookie[x]
	}
}

func xorv6(ip []byte, messageCookie []byte, transactionID []byte) {
	xorv4(ip, messageCookie)

	for x := 4; x < 16; x++ {
		ip[x] ^= transactionID[x-4]
	}
}

func validateRequest(packet []byte, n int) error {
	if n != 20 {
		return fmt.Errorf("Request is %d bytes, should be 20", n)
	}

	messageType := binary.BigEndian.Uint16(packet[0:2])

	if messageType != 0x0001 {
		return fmt.Errorf("Request is type %x, should be Binding Request (0x0001)", messageType)
	}

	return nil
}

func makeResponse(packet []byte, addr *net.UDPAddr) []byte {
	messageCookie := packet[4:8]
	magicCookie := uint16(binary.BigEndian.Uint32(messageCookie) >> 16)
	transactionID := packet[8:20]

	size := uint16(44)
	response := make([]byte, size)

	if isv4(addr.IP) {
		size = 32
		response[25] = 0x01 //IPv4
		copy(response[28:32], addr.IP[12:16])
		xorv4(response[28:32], messageCookie)

	} else {
		response[25] = 0x02 //IPv6
		copy(response[28:44], addr.IP)
		xorv6(response[28:44], messageCookie, transactionID)
	}

	//header
	binary.BigEndian.PutUint16(response[0:2], 0x0101)
	binary.BigEndian.PutUint16(response[2:4], size-20)
	copy(response[4:8], messageCookie)
	copy(response[8:20], transactionID)

	//XOR-MAPPED-ADDRESS
	binary.BigEndian.PutUint16(response[20:22], 0x0020)
	binary.BigEndian.PutUint16(response[22:24], size-20-4)
	binary.BigEndian.PutUint16(response[26:28], uint16(addr.Port)^magicCookie)

	return response[0:size]
}

func main() {
	addr := net.UDPAddr{
		Port: 3478,
		IP:   net.ParseIP("::"),
	}

	srv, err := net.ListenUDP("udp", &addr)

	if err == nil {
		for {
			packet := make([]byte, 128)
			n, peerAddr, err := srv.ReadFromUDP(packet)

			if err == nil {
				err = validateRequest(packet, n)

				if err == nil {
					_, err = srv.WriteToUDP(makeResponse(packet, peerAddr), peerAddr)
				}
			}

			if err != nil {
				fmt.Printf("%v %v\n", peerAddr, err)
			}
		}

	} else {
		fmt.Printf("%v\n", err)
	}

	srv.Close()
}
