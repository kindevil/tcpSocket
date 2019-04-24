package socket

import (
	"net"

	"github.com/wonderivan/logger"
)

type events interface {
	OnHandel(id string, conn net.Conn) bool
	OnClose(session *Session)
	OnMessage(session *Session, packetType int16, message []byte) bool
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

func (d *DefaultEvents) OnMessage(session *Session, packetType int16, message []byte) bool {
	logger.Info(packetType)
	logger.Info(len(message))
	logger.Info(string(message[:]))
	return true
}
