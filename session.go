package tcpSocket

import (
	"net"
	"sync"
	"time"

	"github.com/wonderivan/logger"
)

type Session struct {
	Id    string
	Conn  net.Conn
	times int64
	lock  sync.Mutex
}

func NewSession(id string, conn net.Conn) *Session {
	return &Session{
		Id:    id,
		Conn:  conn,
		times: time.Now().Unix(),
	}
}

func (s *Session) write(msg string) error {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, err := s.Conn.Write([]byte(msg))
	return err
}

func (s *Session) close() {
	s.Conn.Close()
}

func (s *Session) UpdateTime() {
	s.times = time.Now().Unix()
}

type SessionMap struct {
	server   *TcpServer
	sessions sync.Map
}

func NewSessionMap(tcpServer *TcpServer) *SessionMap {
	if tcpServer == nil {
		return nil
	}

	return &SessionMap{
		server: tcpServer,
	}
}

func (s *SessionMap) GetSession(id string) *Session {
	value, ok := s.sessions.Load(id)
	if ok {
		if session, ok := value.(*Session); ok {
			return session
		}
	}

	return nil
}

func (s *SessionMap) SetSession(id string, conn net.Conn) {
	session := NewSession(id, conn)
	s.sessions.Store(id, session)
}

func (s *SessionMap) RemoveSession(id string) {
	value, ok := s.sessions.Load(id)
	if ok {
		if session, ok := value.(*Session); ok {
			session.close()
		}
	}
	s.sessions.Delete(id)
}

func (s *SessionMap) Broadcast(message []byte) {
	message, err := s.server.SocketType.Pack(message)
	if err != nil {
		return
	}
	s.sessions.Range(func(key, value interface{}) bool {
		if value, ok := value.(*Session); ok {
			if err := value.write(string(message)); err != nil {
				s.RemoveSession(key.(string))
				logger.Error(err)
			}
		}
		return true
	})
}

func (s *SessionMap) WriteTo(id string, message []byte) bool {
	message, err := s.server.SocketType.Pack(message)
	if err != nil {
		return false
	}
	value, ok := s.sessions.Load(id)
	if ok {
		if session, ok := value.(*Session); ok {
			if err := session.write(string(message)); err == nil {
				return true
			}
		}
	}
	s.RemoveSession(id)
	return false
}

func (s *SessionMap) HeartBeat(num int64) {
	if num != 0 {
		for {
			time.Sleep(time.Second)
			s.sessions.Range(func(key, value interface{}) bool {
				val, ok := value.(*Session)
				if !ok {
					return true
				}

				if time.Now().Unix()-val.times > num {
					s.RemoveSession(key.(string))
				}
				return true
			})
		}
	}
}
