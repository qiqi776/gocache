package gocache

import (
	"gocache/consistenthash"
	"sync"
)

type PeerPicker interface {
	PickPeer(key string) (peer PeerGetter, ok bool, isSelf bool)
}

type PeerGetter interface {
	Get(group string, key string) ([]byte, error)
	Delete(group string, key string) (bool, error)
}

type ClientPicker struct {
	self        string
	serviceName string
	mu  		sync.RWMutex
	consHash    *consistenthash.Map
}