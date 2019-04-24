package main

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"
	"time"

	"github.com/wonderivan/logger"
)

const (
	HEADER          int16 = 1028
	HEARTBEATPACKET int16 = 1000
	REQUESTPACKET   int16 = 1001
)

type TcpClinet struct {
	Resolve *net.TCPAddr
	Conn    net.Conn
}

type SocketPacket struct {
	Header     int16
	PacketType int16
	Length     int32
	Content    []byte
}

type HeartBeat struct {
	Counter int
}

func main() {
	clinet := NewClinet("127.0.0.1:9000")
	clinet.Start()
}

func NewClinet(address string) *TcpClinet {
	resolve, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		logger.Error(err)
	}

	conn, err := net.DialTCP("tcp", nil, resolve)
	if err != nil {
		logger.Error(err)
	}

	tcpClinet := &TcpClinet{
		Resolve: resolve,
		Conn:    conn,
	}

	return tcpClinet
}

func (t *TcpClinet) Start() {
	go t.heartbeat()
	go t.read()

	for {
		text, err := bufio.NewReader(os.Stdin).ReadString('\n')
		if err != nil {
			break
		}

		if len(text) > 0 {
			text := strings.TrimSpace(text)
			msg := t.Pack([]byte(text), REQUESTPACKET)
			if err != nil {
				break
			}
			_, err := t.Conn.Write(msg)
			if err != nil {
				fmt.Println(err)
				break
			}

			time.Sleep(time.Millisecond * 100)
		}
	}
}

func (t *TcpClinet) heartbeat() {
	counter := 1
	for {
		time.Sleep(time.Second * 5)
		heartBeat := HeartBeat{
			Counter: counter,
		}

		msg, err := json.Marshal(&heartBeat)
		if err != nil {
			break
		}

		msg = t.Pack(msg, HEARTBEATPACKET)
		_, err = t.Conn.Write(msg)
		if err != nil {
			fmt.Println(err)
			break
		}
	}
}

func (t *TcpClinet) read() {
	buffer := make([]byte, 0)
	readBuffer := make([]byte, 512)
	data := make([]byte, 20)
	packetType := int16(0)

	for {
		l, err := t.Conn.Read(readBuffer)
		if err != nil {
			return
		}

		buffer = append(buffer, readBuffer[:l]...)
		buffer, data, packetType, err = t.Unpack(buffer)
		if err != nil {
			logger.Error(err)
			return
		}

		if len(data) == 0 {
			continue
		}

		if packetType != HEARTBEATPACKET {
			// do something
		}

		logger.Info(string(data))
	}
}

func (s *TcpClinet) Unpack(buffer []byte) ([]byte, []byte, int16, error) {
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

func (s *TcpClinet) Pack(message []byte, packetType int16) []byte {
	socketPacket := &SocketPacket{
		Header:     HEADER,
		PacketType: packetType,
		Length:     int32(len(message)),
		Content:    message,
	}

	buffer := new(bytes.Buffer)
	var err error

	err = binary.Write(buffer, binary.BigEndian, &socketPacket.Header)
	err = binary.Write(buffer, binary.BigEndian, &socketPacket.PacketType)
	err = binary.Write(buffer, binary.BigEndian, &socketPacket.Length)
	err = binary.Write(buffer, binary.BigEndian, &socketPacket.Content)

	if err != nil {
		logger.Error(err)
	}

	return buffer.Bytes()
}
