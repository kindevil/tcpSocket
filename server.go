package socket

import (
	"net"

	"github.com/wonderivan/logger"
)

type SocketTypes interface {
	Handle(tcpServer *TcpServer, session *Session)
	Pack(data []byte) []byte
}

type TcpServer struct {
	SessionMap    *SessionMap
	SocketType    SocketTypes
	Event         events
	HeartBeatTime int64
}

func NewServer(socketType SocketTypes) *TcpServer {
	tcpServer := &TcpServer{
		SocketType:    socketType,
		HeartBeatTime: 6,
	}
	tcpServer.SessionMap = NewSessionMap(tcpServer)

	if tcpServer.Event == nil {
		tcpServer.RegisterEvent(&DefaultEvents{})
	}

	return tcpServer
}

func (t *TcpServer) Start(address string) {
	listen, err := net.Listen("tcp", address)

	if err != nil {
		panic(err)
	}

	go t.SessionMap.HeartBeat(t.HeartBeatTime)

	for {
		conn, err := listen.Accept()
		if err != nil {
			logger.Error(err)
			continue
		}

		id := RandString(12)

		if t.Event.OnHandel(id, conn) == false {
			continue
		}

		t.SessionMap.SetSession(id, conn)
		go t.SocketType.Handle(t, t.SessionMap.GetSession(id))
	}
}

func (t *TcpServer) RegisterEvent(event events) {
	t.Event = event
}

func (t *TcpServer) SetHeartBeat(time int64) {
	t.HeartBeatTime = time
}
