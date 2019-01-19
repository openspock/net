// Package net has messages and functions that enable various nodes
// in a p2p network, or clients/ servers to communicate with each other.
package net

import (
	"errors"
	"fmt"
	"net"
	"sync"

	log "github.com/openspock/log"
	netm "github.com/openspock/net/proto"
	xid "github.com/rs/xid"
)

var _ = log.Debug
var _ = netm.SessionType_MULTI_CALL

var ip string
var nodeTable = struct {
	sync.RWMutex
	m map[string]Node
}{m: make(map[string]Node)}

func init() {
	ip, _ := ipv4()
	log.SysInfo("Host ip: " + ip)
}

func readNode(key string) (*Node, bool) {
	nodeTable.RLock()
	node, keyFound := nodeTable.m[key]
	nodeTable.RUnlock()
	return &node, keyFound
}

func writeNode(key string, node Node) {
	nodeTable.Lock()
	nodeTable.m[key] = node
	nodeTable.Unlock()
}

// Iter is a func type that receives a deep copy of a Node,
// usually in an iterative fashion.
type iter func(node Node)

// OnEveryNode takes a func of type iter, which is called for
// every node in the Node table.
func OnEveryNode(i iter) {
	nodeTable.RLock()
	for _, v := range nodeTable.m {
		n := Node{v.id, v.Host, v.Port}
		i(n)
	}
	nodeTable.RUnlock()
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

// Node represents an active listening process on a host/ port.
type Node struct {
	id   string
	Host string
	Port string
}

// Dial dials a connections to a Node. The connections are always secure.
//
// The connection, if established will be secured with a TLSv1.2 cert.
func (node *Node) Dial() (net.Conn, error) {
	return nil, nil
}

// NewNode creates a new Node if there isn't one existing already,
// and returns it.
func NewNode(host, port string) (*Node, error) {
	key := fmt.Sprintf("%s:%s", host, port)
	if _, keyFound := readNode(key); keyFound {
		return nil, errors.New("Node already exists at this host and port")
	}
	node := Node{
		id:   xid.New().String(),
		Host: host,
		Port: port,
	}
	writeNode(key, node)
	return &node, nil
}
