package tcpSocket

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/wonderivan/logger"
)

const (
	HEADER          int16 = 1028
	HEARTBEATPACKET int16 = 1000
	REQUESTPACKET   int16 = 1001
)

type Socket struct {
}

type SocketPacket struct {
	Header     int16
	PacketType int16
	Length     int32
	Content    []byte
}

func (s *Socket) Handle(tcpServer *TcpServer, session *Session) {
	defer func() {
		tcpServer.SessionMap.RemoveSession(session.Id)
		tcpServer.Event.OnClose(session)
	}()

	buffer := make([]byte, 0)
	readBuffer := make([]byte, 1024)
	data := make([]byte, 20)
	packetType := int16(0)

	for {
		l, err := session.Conn.Read(readBuffer)
		session.UpdateTime()
		if err != nil {
			return
		}

		buffer = append(buffer, readBuffer[:l]...)
		buffer, data, packetType, err = s.Unpack(buffer)
		if err != nil {
			logger.Error(err)
			return
		}

		if len(data) == 0 {
			continue
		}

		if packetType != HEARTBEATPACKET {
			if tcpServer.Event.OnMessage(tcpServer, session, packetType, data) == false {
				return
			}
		} else {
			logger.Debug("receive heaetbeat packet form ", session.Conn.RemoteAddr().String())
		}
	}
}

func (s *Socket) Pack(message []byte) ([]byte, error) {
	socketPacket := &SocketPacket{
		Header:     HEADER,
		PacketType: REQUESTPACKET,
		Length:     int32(len(message)),
		Content:    message,
	}

	buffer := new(bytes.Buffer)
	var err error

	err = binary.Write(buffer, binary.BigEndian, &socketPacket.Header)
	err = binary.Write(buffer, binary.BigEndian, &socketPacket.PacketType)
	err = binary.Write(buffer, binary.BigEndian, &socketPacket.Length)
	err = binary.Write(buffer, binary.BigEndian, &socketPacket.Content)

	return buffer.Bytes(), err
}

func (s *Socket) Unpack(buffer []byte) ([]byte, []byte, int16, error) {
	length := len(buffer)

	if length < 8 {
		return buffer, nil, 0, nil
	}

	header := int16(0)
	binary.Read(bytes.NewReader(buffer[0:2]), binary.BigEndian, &header)
	if header != HEADER {
		return []byte{}, nil, 0, errors.New("header is illegal")
	}

	messageLength := int32(0)
	binary.Read(bytes.NewReader(buffer[4:8]), binary.BigEndian, &messageLength)
	if length < 8+int(messageLength) {
		return buffer, nil, 0, nil
	}

	packetType := int16(0)
	binary.Read(bytes.NewReader(buffer[2:4]), binary.BigEndian, &packetType)

	data := buffer[8 : messageLength+8]
	tbuffer := buffer[messageLength+8:]
	return tbuffer, data, packetType, nil
}
