package consistenthash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

var (
	GlobalReplicas  = defaultReplicas
	defaultReplicas = 150
	defaultHash     = crc32.ChecksumIEEE
)

type Hash func(data []byte) uint32

type Map struct {
	hash 	 Hash
	replicas int
	keys  	 []int
	hashMap  map[int]string
}

type ConsOptions func(*Map)

func New(opts ...ConsOptions) *Map {
	m := Map{
		hash: defaultHash,
		replicas: defaultReplicas,
		hashMap: make(map[int]string),
	}
	for _, opt := range opts {
		opt(&m)
	}
	return &m
}

func Replicas(replicas int) ConsOptions {
	return func(m *Map) {
		m.replicas = replicas
	}
}

func HashFunc(hash Hash) ConsOptions {
	return func(m *Map) {
		m.hash = hash
	}
}

func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}
	hash := int(m.hash([]byte(key)))
	// 顺时针找到第一个匹配的虚拟节点
	idx := sort.Search(len(m.keys), func(i int) bool {
		return m.keys[i] >= hash
	})
	// 如果没有找到匹配的虚拟节点，返回哈希环上的第一个节点
	return m.hashMap[m.keys[idx%len(m.keys)]]
}

func (m *Map) Remove (key string) {
	for i := 0; i < m.replicas; i++ {
		hash := int(m.hash([]byte(strconv.Itoa(i) + key)))
		idx := sort.SearchInts(m.keys, hash)
		m.keys = append(m.keys[:idx], m.keys[idx+1:]...)
		delete(m.hashMap, hash)
	}
}