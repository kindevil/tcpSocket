package tcpSocket

import (
	"net"

	"github.com/wonderivan/logger"
)

type SocketTypes interface {
	Handle(tcpServer *TcpServer, session *Session)
	Pack(message []byte) ([]byte, error)
}

type TcpServer struct {
	SessionMap    	*SessionMap
	SocketType    	SocketTypes
	Event         	events
	EnableHeartBeat	bool
	HeartBeatTime 	int64
}

func NewServer(socketType SocketTypes) *TcpServer {
	tcpServer := &TcpServer{
		SocketType:    	socketType,
		EnableHeartBeat:false,	
		HeartBeatTime: 	6,
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

	t.Event.AfterStart(t, address)

	if t.EnableHeartBeat {
		go t.SessionMap.HeartBeat(t.HeartBeatTime)
	}

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
	t.EnableHeartBeat = true
	t.HeartBeatTime = time
}
