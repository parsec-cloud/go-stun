package main

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"
	"time"
)

const MagicCookie uint32 = 0x2112A442
const MethodBinding uint16 = 0x001

type StunHeader struct { // 20 bytes
	Type          uint16   // 2 bytes
	Length        uint16   // 2 bytes
	MessageCookie uint32   // 4 bytes
	TransactionID [12]byte // 12 bytes
}

var mu sync.RWMutex
var stunServerIsHealthy = false

type HttpHandler struct{}

func (h HttpHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if isHealthy() {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
}

func isHealthy() bool {
	mu.RLock()
	defer mu.RUnlock()
	return stunServerIsHealthy
}

func setHealthStatus(s bool) {
	mu.Lock()
	stunServerIsHealthy = s
	mu.Unlock()
}
func testLocalStunService() {
	log.Println("Running STUN test")
	err := makeLocalStunBindingRequest()
	if err != nil {
		log.Printf("An error was observed while testing the STUN service %v", err)
		setHealthStatus(false)
		return
	}
	setHealthStatus(true)
}

func encodeToBytes(sh StunHeader) []byte {
	buf := make([]byte, 20)
	binary.BigEndian.PutUint16(buf[0:], sh.Type)
	binary.BigEndian.PutUint16(buf[2:], sh.Length)
	binary.BigEndian.PutUint32(buf[4:], sh.MessageCookie)
	copy(buf[8:], sh.TransactionID[:])
	return buf
}

func makeLocalStunBindingRequest() error {

	// create a valid test binding request for the STUN server
	var b = encodeToBytes(StunHeader{Type: MethodBinding,
		Length:        0,
		MessageCookie: MagicCookie,
		TransactionID: [12]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}})

	// send the STUN message binding datagram
	conn, err := net.Dial("udp", "127.0.0.1:3478")
	defer conn.Close()
	if err != nil {
		log.Printf("An error was triggered while sending STUN request %v", err)
		return err
	}
	conn.Write(b)

	// receive the resonse from STUN server, or timeout and fail if it is slow
	rBuf := make([]byte, 128)
	conn.SetReadDeadline(time.Now().Add(1 * time.Second))
	_, err = bufio.NewReader(conn).Read(rBuf)
	if err != nil {
		log.Printf("An error was triggered while reading STUN response %v", err)
		return err
	}
	log.Println("Success")
	return nil
}

//

func main() {

	fmt.Println("Starting TCP healthcheck service.")
	// create a new http request handler
	handler := HttpHandler{}

	// initialize the healthy state
	testLocalStunService()

	// run the check on a timer from now on
	ticker := time.NewTicker(10 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				go testLocalStunService()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()

	// listen and serve the health service
	http.ListenAndServe(":8888", handler)
}
