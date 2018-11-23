package main

import "net"

func do() {
	sock, _ := net.Dial("tcp", ":50030")
	defer sock.Close()
	_, _ = sock.Write([]byte("Hello! This is test message from my sample client!"))
}

func main() {
	do()
}
