package gocache

import (
	"fmt"
	"gocache/singleflight"
	"log"
	"sync"
	"time"
)

var (
	lock   sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string
	getter    Getter              
	mainCache cache              
	peers     PeerPicker     
	loader    *singleflight.Group
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called multiple times")
	}
	g.peers = peers
}

// NewGroup 新创建一个Group
func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil Getter")
	}
	lock.Lock()
	defer lock.Unlock()
	g := &Group{
		name:   name,
		getter: getter,
		mainCache: cache{
			cacheBytes: cacheBytes,
		},
		loader: &singleflight.Group{},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group {
	lock.RLock()
	g := groups[name]
	lock.RUnlock()
	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}
	return g.load(key)
}

func (g *Group) load(key string) (ByteView, error) {
	v, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil {
			if peer, ok, isSelf := g.peers.PickPeer(key); ok {
				if isSelf {
					if v, ok := g.mainCache.get(key); ok {
						log.Println("[Geek-Cache] hit")
						return v, nil
					}
				} else {
					if value, err := g.getFromPeer(peer, key); err == nil {
						return value, nil
					} else {
						log.Println("[Geek-Cache] Failed to get from peer", err)
					}
				}
			}
		}
		return g.getLocally(key)
	})

	if err == nil {
		return v.(ByteView), nil
	}
	return ByteView{}, err
}

func (g *Group) Delete(key string) (bool, error) {
	if key == "" {
		return true, fmt.Errorf("key is required")
	}
	if g.peers == nil {
		return g.mainCache.delete(key), nil
	}
	peer, ok, isSelf := g.peers.PickPeer(key)
	if !ok {
		return false, nil
	}
	if isSelf {
		return g.mainCache.delete(key), nil
	} else {
		success, err := g.deleteFromPeer(peer, key)
		return success, err
	}
}

func (g *Group) getFromPeer(peer Peer, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{
		b: cloneBytes(bytes),
	}, nil
}

func (g *Group) deleteFromPeer(peer Peer, key string) (bool, error) {
	success, err := peer.Delete(g.name, key)
	if err != nil {
		return false, err
	}
	return success, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	if v, ok := g.mainCache.get(key); ok {
		log.Println("[Geek-Cache] hit")
		return v, nil
	}
	bytes, f, expirationTime := g.getter.Get(key)
	if !f {
		return ByteView{}, fmt.Errorf("data not found")
	}
	bw := ByteView{cloneBytes(bytes)}
	if !expirationTime.IsZero() {
		g.mainCache.addWithExpiration(key, bw, expirationTime)
	} else {
		g.mainCache.add(key, bw)
	}
	return bw, nil
}

type Getter interface {
	Get(key string) ([]byte, bool, time.Time)
}

type GetterFunc func(key string) ([]byte, bool, time.Time)

func (f GetterFunc) Get(key string) ([]byte, bool, time.Time) {
	return f(key)
}

func DestroyGroup(name string) {
	g := GetGroup(name)
	if g != nil {
		delete(groups, name)
		log.Printf("Destroy store [%s]", name)
	}
}
