package main

import (
	"github.com/kindevil/tcpSocket"
)

func main() {
	var server = tcpSocket.NewServer(&tcpSocket.Socket{})
	server.Start(":9000")
}
