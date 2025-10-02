package network

import (
	"errors"
	"net"
	"strings"
)

func LocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() {
			ipv4 := ipnet.IP.To4()
			if ipv4 != nil && strings.Index(ipv4.String(), "127") != 0 {
				return ipv4.String(), nil
			}
		}
	}
	return "", errors.New("cannot lookup local ip address")
}
