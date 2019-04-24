package main

import (
	"github.com/kindevil/tcpSocket"
)

func main() {
	var server = socket.NewServer(&socket.Socket{})
	server.Start(":9000")
}
