// Package net has messages and functions that enable various nodes
// in a p2p network, or clients/ servers to communicate with each other.
package net

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"sync"

	proto "github.com/golang/protobuf/proto"
	log "github.com/openspock/log"
	netm "github.com/openspock/net/proto"
	xid "github.com/rs/xid"
)

var _ = netm.SessionType_MULTI_CALL

var ip string
var nodeTable = struct {
	sync.RWMutex
	m map[string]Node
}{m: make(map[string]Node)}

func init() {
	ip, _ := ipv4()
	log.Info("Host ip: "+ip, log.AppLog, map[string]interface{}{})
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

// Listen starts a tcp listener on a port. The port is secured with TLSv1.2
func Listen(port string) error {
	log.Info(fmt.Sprintf("secure listener started on port: %s", port), log.AppLog, map[string]interface{}{})

	cer, err := tls.LoadX509KeyPair("server.crt", "server.key")
	if err != nil {
		log.Error(err.Error(), log.AppLog, map[string]interface{}{})
		return err
	}

	cfg := &tls.Config{
		MinVersion:   tls.VersionTLS12,
		Certificates: []tls.Certificate{cer},
	}

	listener, err := tls.Listen("tcp", fmt.Sprintf(":%s", port), cfg)
	if err != nil {
		log.Error(err.Error(), log.AppLog, map[string]interface{}{})
		return err
	}
	defer listener.Close()

	for {
		if conn, err := listener.Accept(); err != nil {
			log.Error(err.Error(), log.AppLog, map[string]interface{}{})
		} else {
			handleConnection(conn)
		}
	}
}

func handleConnection(conn net.Conn) {
	// handle connection -
	// receive proto message and choose handler based on type
	defer conn.Close()
	s := bufio.NewScanner(conn)
	req := &netm.Request{}
	if err := proto.Unmarshal(s.Bytes(), req); err != nil {
		log.Error(err.Error(), log.AppLog, map[string]interface{}{})
	}

	switch req.GetType() {
	case netm.Request_ASYNC:
		response := handleAsync(req)
		responseB, err := proto.Marshal(&response)
		if err != nil {
			log.Error(err.Error(), log.AppLog, map[string]interface{}{})
		} else {
			conn.Write(responseB)
		}
	}
}

//handler for async requests. netm.RequestType
func handleAsync(req *netm.Request) netm.Response {
	return netm.Response{Status: &netm.Status{Type: netm.Status_NO_ERROR, Message: "Success"}}
}

func handleLogin(netm.Login) {

}

// Node represents an active listening process on a host/ port.
type Node struct {
	id   string
	Host string
	Port string
}

// Dial dials a secure connection to a Node and returns it.
// The connection, if established will be secured with a TLSv1.2 cert.
//
// The caller of this method should ensure that the connection is closed.
func (node *Node) Dial() (net.Conn, error) {
	log.Info(fmt.Sprintf("dial connection to node %s:%s", node.Host, node.Port), log.AppLog, map[string]interface{}{})

	cert, err := ioutil.ReadFile("server.crt")
	if err != nil {
		return nil, err
	}
	certPool := x509.NewCertPool()
	certPool.AppendCertsFromPEM(cert)

	conf := &tls.Config{
		MinVersion: tls.VersionTLS12,
		RootCAs:    certPool,
	}
	conn, err := tls.Dial("tcp", fmt.Sprintf("%s:%s", node.Host, node.Port), conf)
	if err != nil {
		return nil, err
	}
	return conn, nil
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
