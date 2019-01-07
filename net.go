// Package net has messages and functions that enable various nodes
// in a p2p network, or clients/ servers to communicate with each other.
package net

import (
	"errors"
	"net"

	log "github.com/openspock/log"
	netm "github.com/openspock/net/proto"
)

var _ = log.Debug
var _ = netm.SessionType_MULTI_CALL

var ip string

func init() {
	ip, _ := ipv4()
	log.SysInfo("Host ip: " + ip)
}

func ipv4() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch t := addr.(type) {
			case *net.IPNet:
				ip = t.IP
			case *net.IPAddr:
				ip = t.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("No network conn found")
}
