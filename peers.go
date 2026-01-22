package gocache

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"
	"gocache/consistenthash"
	"gocache/registry"
	"github.com/sirupsen/logrus"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type PeerPicker interface {
	PickPeer(key string) (peer Peer, ok bool, isSelf bool)
}

type Peer interface {
	Get(group string, key string) ([]byte, error)
	Delete(group string, key string) (bool, error)
}

type ClientPicker struct {
	selfAddr string
	svcName  string
	mu       sync.RWMutex
	consHash *consistenthash.Map
	clients  map[string]*Client
	etcdCli  *clientv3.Client 
}

type PickerOptions func(*ClientPicker)

func NewClientPicker(addr string, opts ...PickerOptions) (*ClientPicker, error) {
	picker := &ClientPicker{
		selfAddr: addr,
		svcName:  defaultSvcName,
		clients:  make(map[string]*Client),
		consHash: consistenthash.New(),
	}
	for _, opt := range opts {
		opt(picker)
	}
	
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   registry.DefaultConfig.Endpoints,
		DialTimeout: registry.DefaultConfig.DialTimeout,
	})
	if err != nil {
		logrus.Errorf("failed to create etcd client: %v", err)
		return nil, fmt.Errorf("failed to create etcd client: %v", err)
	}
	picker.etcdCli = cli

	go picker.watchServiceChanges()
	go picker.fetchAllService()
	return picker, nil
}

// watchServiceChanges 监听服务实例的增量变化
func (p *ClientPicker) watchServiceChanges() {
	watcher := clientv3.NewWatcher(p.etcdCli)
	watchChan := watcher.Watch(context.Background(), p.svcName, clientv3.WithPrefix())

	for {
		select {
		case a := <-watchChan:
			p.mu.Lock()
			for _, event := range a.Events {
				addr := parseAddrFromKey(string(event.Kv.Key), p.svcName)
				if addr == p.selfAddr {
					continue
				}
				if event.IsCreate() {
					if _, ok := p.clients[addr]; !ok {
						p.set(addr)
					}
				} else if event.Type == clientv3.EventTypeDelete {
					if _, ok := p.clients[addr]; ok {
						p.remove(addr)
					}
				}
			}
			p.mu.Unlock()
		}
	}
}

// fetchAllServices 获取所有服务实例并更新
func (p *ClientPicker) fetchAllService() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := p.etcdCli.Get(ctx, p.svcName, clientv3.WithPrefix())
	if err != nil {
		logrus.Errorf("failed to get all service: %v", err)
		return fmt.Errorf("failed to get all service: %v", err)
	}

	for _, kv := range resp.Kvs {
		addr := parseAddrFromKey(string(kv.Key), p.svcName)
		if _, ok := p.clients[addr]; !ok {
			p.set(addr)
		}
	}
	return nil
}

// set 添加一个服务实例到哈希环
func (p *ClientPicker) set(addr string) error {
	p.consHash.Add(addr)
	client, err := NewClient(addr, p.svcName, p.etcdCli)
	if err != nil {
		logrus.Errorf("failed to create client: %v", err)
		return fmt.Errorf("failed to create client: %v", err)
	}
	p.clients[addr] = client
	return nil
}

// remove 从哈希环中移除一个服务实例
func (p *ClientPicker) remove(addr string) error {
	p.consHash.Remove(addr)
	delete(p.clients, addr)
	return nil
}

// parseAddrFromKey 从 etcd 的 key 中提取出服务实例的地址
func parseAddrFromKey(key, svcName string) string {
	idx := strings.Index(key, svcName)
	if idx == -1 {
		return ""
	}
	return key[idx+len(svcName)+1:]
}

// PickPeer 根据一致性哈希算法选择一个 peer
func (p *ClientPicker) PickPeer(key string) (Peer, bool, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if peer := p.consHash.Get(key); peer != "" {
		logrus.Infof("pick peer: %s", peer)
		return p.clients[peer], true, peer == p.selfAddr
	}
	return nil, false, false
}
