package main

import (
	"github.com/kindevil/newcontrol/socket"
)

func main() {
	var server = socket.NewServer(&socket.Socket{})
	server.Start(":9000")
}
