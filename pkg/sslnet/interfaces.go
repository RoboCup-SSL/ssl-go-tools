package sslnet

import (
	"log"
	"net"
)

func interfaces(skipInterfaces []string) (interfaces []net.Interface) {
	interfaces = []net.Interface{}
	ifis, err := net.Interfaces()
	if err != nil {
		log.Println("Could not get available interfaces: ", err)
	}
	for _, ifi := range ifis {
		if skipInterface(ifi, skipInterfaces) {
			continue
		}
		interfaces = append(interfaces, ifi)
	}
	return
}

func hasIpv4Net(ifi net.Interface) bool {
	addrs, err := ifi.Addrs()
	if err != nil {
		log.Printf("Could not get addresses for interface %v: %v", ifi.Name, err)
		return false
	}

	for _, addr := range addrs {
		if ipNet, ok := addr.(*net.IPNet); ok {
			ip := ipNet.IP
			if ip.To4() != nil {
				return true
			}
		}
	}

	return false
}

func skipInterface(ifi net.Interface, skipInterfaces []string) bool {
	for _, skipIfi := range skipInterfaces {
		if skipIfi == ifi.Name {
			return true
		}
	}

	if ifi.Flags&net.FlagMulticast == 0 ||
		ifi.Flags&net.FlagUp == 0 {
		return true
	}

	if !hasIpv4Net(ifi) {
		return true
	}

	return false
}
