package network

import (
	"errors"
	"net"
	"strings"
)

func LocalIPString() (string, error) {
	ip, err := localIP()
	if err != nil {
		return "", err
	}
	return ip.String(), nil
}

func LocalIP() ([]uint8, error) {
	ipv4, err := localIP()
	if err != nil {
		return nil, err
	}
	result := make([]uint8, 4)
	copy(result, ipv4)
	return result, nil
}

func localIP() (net.IP, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok &&
			!ipnet.IP.IsLoopback() {
			ipv4 := ipnet.IP.To4()
			if ipv4 != nil && strings.Index(ipv4.String(), "127") != 0 {
				return ipv4, nil
			}
		}
	}
	return nil, errors.New("cannot lookup local ip address")
}
