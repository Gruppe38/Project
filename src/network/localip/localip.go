package localip

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strings"
)

var localIP string

func LocalIP() (string, error) {
	if localIP == "" {
		conn, err := net.DialTCP("tcp4", nil, &net.TCPAddr{IP: []byte{8, 8, 8, 8}, Port: 53})
		if err != nil {
			return "", err
		}
		defer conn.Close()
		localIP = strings.Split(conn.LocalAddr().String(), ":")[0]
	}
	return localIP, nil
}

func GetProcessID() string {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		IP, err := LocalIP()
		if err != nil {
			fmt.Println(err)
			IP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", IP, os.Getpid())
	}
	return id
}
