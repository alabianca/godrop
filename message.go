package godrop

import (
	"bytes"
	"encoding/binary"
)

type Header struct {
	Type   int
	Length uint16
}

func (h *Header) Encode() []byte {
	packet := make([]byte, 0)
	packet = append(packet, byte(h.Type))

	length, _ := uint16ToBytes(h.Length)

	packet = append(packet, length...)
	packet = append(packet, 0)

	return packet

}

func uint16ToBytes(num uint16) ([]byte, error) {
	buf := new(bytes.Buffer)

	if err := binary.Write(buf, binary.BigEndian, num); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

type Message struct {
	header  Header
	payload []byte
}

func NewMessage(mType int, payload []byte) *Message {
	return &Message{
		header: Header{
			Type:   mType,
			Length: uint16(len(payload)),
		},
		payload: payload,
	}
}

func (m *Message) Encode() []byte {
	packet := make([]byte, 0)

	packet = append(packet, m.header.Encode()...)
	packet = append(packet, m.payload...)
	packet = append(packet, END_OF_TEXT)

	return packet
}

func (m *Message) Decode(payload []byte) {
	header := Header{
		Type:   int(payload[0]),
		Length: binary.BigEndian.Uint16(payload[1:3]),
	}

	m.header = header

	m.payload = payload[3 : 3+header.Length]
}
