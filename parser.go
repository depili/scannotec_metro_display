package main

import (
	"fmt"
	"github.com/tarm/serial"
	"log"
	"os"
	"time"
)

const (
	headerLen = 6
	address   = 0x02
)

func main() {
	testMessage := []byte{0x82, 0x00, 0x30, 0x01, 0x00, 0x04, 0x31, 0x32, 0x33, 0x34, 0xca}

	vMessage := []byte{0x82, 0x00, address, 'V'}
	wMessage := []byte{0x82, 0x00, address, 'W'}
	pingMessage := []byte{0x82, 0x00, address, 0x81}

	m := message{}

	for _, b := range testMessage {
		m.parse(b)
	}

	m = message{
		Addr:    0x02,
		MsgType: 'U',
		Payload: make([]byte, 0),
	}

	for i := 0; i < 8; i++ {
		str := fmt.Sprintf("%d Hacklab ry", i)
		encoded := rowText(byte(i), []byte(str))
		m.Payload = append(m.Payload, encoded...)
	}
	dataPacket := m.encode()
	fmt.Printf("Encoded message:\n%X\n", dataPacket)

	m.resetState()
	for _, b := range dataPacket {
		m.parse(b)
	}

	bauds := []int{4800, 9600, 2400, 19200, 1200, 600}
	for _, b := range bauds {
		sendAll(b)
	}
}

func sendAll(baud int) {
	log.Printf("Trying with baudrate: %d", baud)
	c := &serial.Config{Name: os.Args[1], Baud: baud}

	s, err := serial.OpenPort(c)
	if err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)

	/*
		log.Printf("Sending first ping %x", pingPacket)
		n, err := s.Write(pingPacket)
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("Wrote %d bytes...", n)

		time.Sleep(500 * time.Millisecond)
	*/

	send(s, wMessage)

	log.Printf("Sending data...")
	send(s, dataPacket)

	time.Sleep(50 * time.Millisecond)

	log.Printf("Sending enable")
	send(s, vMessage)

	time.Sleep(50 * time.Millisecond)

	send(s, pingMessage)

	s.Flush()
	s.Close()
	time.Sleep(2 * time.Second)
}

func send(s *serial.Port, data []byte) {
	log.Printf("Sending %x", data)
	n, err = s.Write(data)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Wrote %d bytes...", n)
}

type message struct {
	MsgType     byte
	Addr        byte
	Payload     []byte
	decodeState int
	checksum    byte
	len         int
}

func (m *message) parse(in byte) bool {
	switch m.decodeState {
	case 0:
		if in != 0x82 {
			return false
		}
		fmt.Printf("got start\n")
		m.resetState()
	case 1:
		if in != 0x00 {
			m.resetState()
			return false
		}
		fmt.Printf("got null\n")
	case 2:
		m.Addr = in
	case 3:
		switch in {
		case 0x5:
			fmt.Printf("set addresses\n")
		case 0x9:
			fmt.Printf("set time\n")
		case 0x56: // 'V'
			m.MsgType = 0x3
		case 0x57: // 'W'
			m.MsgType = 0x4
			m.print()
			m.resetState()
			return true
		case 0x81:
			// Ping?
			fmt.Printf("ping?\n")
			m.resetState()
			return true
		case 0x87:
			// Set RTC
			fmt.Printf("RTC something\n")
		default:
			m.MsgType = in
		}
	case 4:
		if in >= 0x8 {
			m.resetState()
			return false
		}
		m.len = int(in) << 8
	case 5:
		m.len += int(in)
		fmt.Printf("got length: %v\n", m.len)
		m.decodeState = 0x7f
	case 0x80:
		if len(m.Payload) == m.len-1 {
			if in == m.checksum {
				fmt.Printf("checksum ok!\n")
			} else {
				fmt.Printf("checksum error: got %x expected %x\n", m.checksum, in)
			}
			m.print()
			m.resetState()
			return true
		} else {
			m.Payload = append(m.Payload, in)
			m.checksum += in
			return false
		}
	}
	m.decodeState++
	return false
}

func rowText(row byte, text []byte) []byte {
	ret := make([]byte, 2)
	ret[0] = 0x1b
	ret[1] = row & 0x7
	ret = append(ret, text...)
	return ret
}

func (m *message) resetState() {
	m.decodeState = 0
	m.MsgType = 0
	m.len = 0
	m.Addr = 0
	m.Payload = make([]byte, 0)
	m.checksum = 0
}

func (m *message) print() {
	fmt.Printf("Message: address: 0x%X type: 0x%X len: 0x%X payload: %s\n", m.Addr, m.MsgType, m.len, m.Payload)
}

func (m *message) encode() []byte {
	m.len = len(m.Payload) + 1

	ret := make([]byte, 4)
	ret[0] = 0x82
	ret[1] = 0x0
	ret[2] = m.Addr
	ret[3] = m.MsgType
	if m.len != 1 {
		ret = append(ret, byte(m.len>>8))
		ret = append(byte(m.len & 0xff))
		m.checksum = 0
		for _, b := range m.Payload {
			m.checksum += b
		}
		ret = append(ret, m.Payload...)
		ret = append(ret, m.checksum)
	}

	return ret
}
