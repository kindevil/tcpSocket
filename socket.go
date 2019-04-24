package socket

import (
	"bytes"
	"encoding/binary"
	"errors"

	"github.com/wonderivan/logger"
)

const (
	HEADER          int16 = 1028
	HeartBeatPacket int16 = 1000
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
	readBuffer := make([]byte, 512)
	data := make([]byte, 20)
	packetTypeByte := make([]byte, 4)

	for {
		l, err := session.Conn.Read(readBuffer)
		session.updateTime()
		if err != nil {
			return
		}

		buffer = append(buffer, readBuffer[:l]...)
		buffer, data, packetTypeByte, err = s.Depack(buffer)
		if err != nil {
			logger.Error(err)
			return
		}

		if len(data) == 0 {
			continue
		}

		packetType := s.BytesToInt16(packetTypeByte)
		if tcpServer.Event.OnMessage(session, packetType, data) == false {
			return
		}
	}
}

func (s *Socket) Pack(message []byte) []byte {
	header := s.Int16ToBytes(HEADER)
	packetType := s.Int16ToBytes(HeartBeatPacket)
	length := s.Int32ToBytes(len(message))
	return append(append(append(header, packetType...), length...), message...)
}

func (s *Socket) Depack(buffer []byte) ([]byte, []byte, []byte, error) {
	length := len(buffer)

	if length < 8 {
		return buffer, nil, nil, nil
	}

	if s.BytesToInt16(buffer[:2]) != HEADER {
		return []byte{}, nil, nil, errors.New("header is illegal")
	}

	messageLength := s.BytesToInt(buffer[4:8])
	if length < 8+messageLength {
		return buffer, nil, nil, nil
	}

	data := buffer[8 : messageLength+8]
	packetType := buffer[2:4]
	tbuffer := buffer[messageLength+8:]
	return tbuffer, data, packetType, nil
}

func (s *Socket) BytesToInt16(b []byte) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int16
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

func (s *Socket) BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}

func (s *Socket) Int16ToBytes(n int16) []byte {
	x := int16(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func (s *Socket) Int32ToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}
