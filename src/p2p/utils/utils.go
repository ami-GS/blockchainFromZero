package utils

import (
	"net"
	"os"
	"strings"
)

func GetExternalIP() string {
	debugVal := os.Getenv("DEBUG_LOCAL_IP")
	if debugVal != "" {
		return "127.0.0.1"
	}

	ifaces, _ := net.Interfaces()
	for _, iface := range ifaces {
		addrs, _ := iface.Addrs()
		for _, addr := range addrs {
			tmp := addr.String()
			if strings.Count(tmp, ".") == 3 && !strings.HasPrefix(tmp, "127.0.0.1") {
				return strings.Split(tmp, "/")[0]
			}
		}
	}
	return ""

}
