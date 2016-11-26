package driver

import (
	"net"
	"github.com/yunify/docker-plugin-hostnic/log"
)

func GetInterfaceIPAddr(ifi net.Interface) string {
	addrs, err := ifi.Addrs()
	if err != nil {
		log.Error("Get interface [%+v] addr error: %s", ifi, err.Error())
		return ""
	}
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok {
			if ipnet.IP.To4() != nil {
				return ipnet.String()
			}
		}
	}
	return ""
}
