package util

import (
	"fmt"
	"net"
)

func ResolveHost(host string) (string, error) {
	ipaddr, err := net.ResolveIPAddr("ip4", host)
	if err != nil {
		return "", err
	}
	if ipaddr.IP.To4() != nil {
		return ipaddr.String(), nil
	}
	return "", fmt.Errorf("failed to lookup host")
}
