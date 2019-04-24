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

type HeartBeat struct {
	Counter int
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

	for {
		l, err := t.Conn.Read(readBuffer)
		if err != nil {
			return
		}

		buffer = append(buffer, readBuffer[:l]...)
		buffer, data, err = t.Depack(buffer)
		if err != nil {
			logger.Error(err)
			return
		}

		if len(data) == 0 {
			continue
		}

		logger.Info(string(data))
	}
}

func (t *TcpClinet) Depack(buffer []byte) ([]byte, []byte, error) {
	length := len(buffer)

	if length < 8 {
		return buffer, nil, nil
	}

	if t.BytesToInt16(buffer[:2]) != HEADER {
		return []byte{}, nil, errors.New("header is illegal")
	}

	messageLength := t.BytesToInt(buffer[4:8])
	if length < 8+messageLength {
		return buffer, nil, nil
	}

	data := buffer[8 : messageLength+8]
	tbuffer := buffer[messageLength+8:]
	return tbuffer, data, nil
}

func (t *TcpClinet) Pack(message []byte, packetType int16) []byte {
	header := t.Int16ToBytes(HEADER)
	packetTypeByte := t.Int16ToBytes(packetType)
	length := t.Int32ToBytes(len(message))
	return append(append(append(header, packetTypeByte...), length...), message...)
}

func (t *TcpClinet) Int16ToBytes(n int16) []byte {
	x := int16(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func (t *TcpClinet) Int32ToBytes(n int) []byte {
	x := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, x)
	return bytesBuffer.Bytes()
}

func (t *TcpClinet) BytesToInt16(b []byte) int16 {
	bytesBuffer := bytes.NewBuffer(b)
	var x int16
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return x
}

func (t *TcpClinet) BytesToInt(b []byte) int {
	bytesBuffer := bytes.NewBuffer(b)
	var x int32
	binary.Read(bytesBuffer, binary.BigEndian, &x)
	return int(x)
}
