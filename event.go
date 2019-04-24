package tcpSocket

import (
	"net"

	"github.com/wonderivan/logger"
)

type events interface {
	OnHandel(id string, conn net.Conn) bool
	OnClose(session *Session)
	OnMessage(tcpserver *TcpServer, session *Session, packetType int16, message []byte) bool
}

type DefaultEvents struct {
}

func (d *DefaultEvents) OnHandel(id string, conn net.Conn) bool {
	logger.Info(id, "new clinet connected.")
	return true
}

func (d *DefaultEvents) OnClose(session *Session) {
	logger.Info(session.Id, "clinet closed.")
}

func (d *DefaultEvents) OnMessage(tcpserver *TcpServer, session *Session, packetType int16, message []byte) bool {
	logger.Info(string(message[:]))
	tcpserver.SessionMap.WriteTo(session.Id, message)
	return true
}
