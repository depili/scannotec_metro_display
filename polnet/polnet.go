package polnet

import (
	"fmt"
	"time"
)

const (
	addSetType        = 0x05
	rtcSetType        = 0x09
	uType             = 0x55
	vType             = 0x56
	wType             = 0x59
	pingType          = 0x81
	analogMeasureType = 0x87
	startMagic        = 0x82
	scrollCode        = 0x09
	normalFontCode    = 0x0e
	boldFontCode      = 0x0f
	blinkOnCode       = 0x11
	blinkOffCode      = 0x12
	enableDynamicCode = 0x13
	timedMsgCode      = 0x14
	tempReplaceCode   = 0x15
	todReplaceCode    = 0x16
	rowSetCode        = 0x1b
)

// Message is the Polnet message representation
type Message struct {
	MsgType     byte
	Addr        byte
	Payload     []byte
	decodeState int
	checksum    byte
	len         int
}

// PingPacket is the '0x81' packet
func PingPacket(addr byte) []byte {
	m := Message{
		MsgType: pingType,
		Addr:    addr,
	}
	return m.Encode()
}

// EnablePacket is the 'V' packet that triggers message reprosessing
func EnablePacket(addr byte) []byte {
	m := Message{
		MsgType: vType,
		Addr:    addr,
	}
	return m.Encode()
}

// EnableWithTimeoutPacket is the 'W' packet, that triggers message reprosessing and then shows stuff with a timeout
func EnableWithTimeoutPacket(addr byte) []byte {
	m := Message{
		MsgType: wType,
		Addr:    addr,
	}
	return m.Encode()
}

// SetTimePacket return a byte slice representing a RTC time set packet
func SetTimePacket(addr byte, t time.Time) []byte {
	msg := Message{
		MsgType: rtcSetType,
		Addr:    addr,
	}
	h, m, s := t.Clock()
	msg.Append([]byte{byte(h), byte(m), byte(s)})
	return msg.Encode()
}

// Append adds the given bytes to message payload
func (m *Message) Append(data []byte) {
	m.Payload = append(m.Payload, data...)
}

// AddRow selects the given row and adds some text to the payload
func (m *Message) AddRow(row int, str string) {
	encoded := rowText(byte(row), []byte(str))
	m.Append(encoded)
}

// AddBlink adds a piece of blinking text to the payload
func (m *Message) AddBlink(str string) {
	enc := []byte{blinkOnCode}
	enc = append(enc, []byte(str)...)
	enc = append(enc, blinkOffCode)
	m.Append(enc)
}

// AddTimed adds a piece of timed text to the payload
func (m *Message) AddTimed(timer byte, str string) {
	s := fmt.Sprintf("%02X%s", timer, str)
	enc := []byte{enableDynamicCode, timedMsgCode}
	enc = append(enc, []byte(s)...)
	m.Append(enc)
}

// AddScroll adds a scrolling text to the payload
func (m *Message) AddScroll(offset byte, str string) {
	s := fmt.Sprintf("%02X%s", offset, str)
	enc := []byte{enableDynamicCode, scrollCode}
	enc = append(enc, []byte(s)...)
	m.Append(enc)
}

// AddBold adds piece of bold text to the payload
func (m *Message) AddBold(str string) {
	enc := []byte{boldFontCode}
	enc = append(enc, []byte(str)...)
	enc = append(enc, normalFontCode)
	m.Append(enc)
}

// AddTod adds the 0x16 special character that gets replaced with time-of-day from RTC
func (m *Message) AddTod() {
	m.Append([]byte{todReplaceCode})
}

// AddTemp adds the 0x15 special character that gets replaced with 4 dynamic characters
func (m *Message) AddTemp() {
	m.Append([]byte{tempReplaceCode})
}

func rowText(row byte, text []byte) []byte {
	ret := make([]byte, 2)
	ret[0] = rowSetCode
	ret[1] = row&0x7 + 0x30
	ret = append(ret, text...)
	return ret
}

func (m *Message) resetState() {
	m.decodeState = 0
	m.MsgType = 0
	m.len = 0
	m.Addr = 0
	m.Payload = make([]byte, 0)
	m.checksum = 0
}

// Encode returns the message encoded into a []byte slice for sending over serial
func (m *Message) Encode() []byte {
	m.len = len(m.Payload) + 1

	ret := make([]byte, 4)
	ret[0] = 0x82
	ret[1] = 0x0
	ret[2] = m.Addr
	ret[3] = m.MsgType
	if m.len != 1 {
		ret = append(ret, byte(m.len>>8))
		ret = append(ret, byte(m.len&0xff))
		m.checksum = 0
		for _, b := range m.Payload {
			m.checksum += b
		}
		ret = append(ret, m.Payload...)
		ret = append(ret, m.checksum)
	}

	return ret
}

func (m *Message) parse(in byte) bool {
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
			fmt.Printf("V message\n")
			m.MsgType = 0x3
			m.print()
			m.resetState()
			return true
		case 0x57: // 'W'
			fmt.Printf("W message\n")
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
			if m.MsgType == 0x55 {
				fmt.Printf("U message\n")
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

func (m *Message) print() {
	fmt.Printf("Message: address: 0x%X type: 0x%X len: 0x%X payload: %X\n", m.Addr, m.MsgType, m.len, m.Payload)
}
