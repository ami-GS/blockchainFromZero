package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"strings"
)

func handleMessage(conn net.Conn) {
	b, err := ioutil.ReadAll(conn)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", b)
}

func getSelfExternalIP() string {
	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			tmp := addr.String()
			if strings.Count(tmp, ".") == 3 && !strings.HasPrefix(tmp, "127.0.0.1") {
				return tmp
			}
		}
	}
	return ""
}

func main() {

	listen, _ := net.Listen("tcp", ":50030")
	for {
		conn, _ := listen.Accept()
		defer conn.Close()
		go handleMessage(conn)
	}
}
