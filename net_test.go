package net

import (
	"strconv"
	"sync"
	"testing"
)

func TestSameNewNodeTwice(t *testing.T) {
	node, err := NewNode("127.0.0.1", "8080")
	if err != nil {
		t.Error("Shouldn't throw an error when creating a first node")
	}
	t.Logf("New node created: %s", node.id)
	_, err = NewNode("127.0.0.1", "8080")
	if err == nil {
		t.Error("Shouldn't permit creation of a new node when a node for same host port exists")
	}
}

func TestSameNewNodeConcurrency(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(10)
	for i := 10; i > 0; i-- {
		go func(i int) {
			defer wg.Done()
			_, err := NewNode("127.0.0.1", strconv.Itoa(i))
			if err != nil {
				t.Error("Shouldn't throw an error when creating a first node")
			}
		}(i)
	}
	wg.Wait()

	nodes := []Node{}
	OnEveryNode(func(n Node) {
		nodes = append(nodes, n)
	})
	if len(nodes) != 10 {
		t.Error("expected size 10")
	}
	for i := 10; i > 0; i-- {
		avail := false
		for _, n := range nodes {
			t.Logf("found a node with port %s", n.Port)
			if n.Host == "127.0.0.1" && n.Port == strconv.Itoa(i) {
				avail = true
			}
		}
		if !avail {
			t.Errorf("expected a node with port %d", i)
		}
	}
}
