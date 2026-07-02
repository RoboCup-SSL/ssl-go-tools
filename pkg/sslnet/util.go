package sslnet

import (
	"fmt"
	"strconv"
	"strings"
)

func PortFromAddress(address string) (int, error) {
	strings.Split(address, ":")
	if len(strings.Split(address, ":")) != 2 {
		return 0, fmt.Errorf("invalid address: %s", address)
	}
	portStr := strings.Split(address, ":")[1]
	port, err := strconv.Atoi(portStr)
	if err != nil {
		return 0, fmt.Errorf("invalid port: %s", address)
	}
	return port, nil
}
