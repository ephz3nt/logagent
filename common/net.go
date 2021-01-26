package common

import (
	"log"
	"net"
	"strings"
)

func GetOutboundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "1.1.1.1:53")
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}
