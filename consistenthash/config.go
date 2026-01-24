package consistenthash

import "hash/crc32"

// Config 一致性哈希配置
type Config struct {
	DefaultReplicas int               // 每个真实节点对应的虚拟节点数
	MinReplicas int					  // 最小虚拟节点数
	MaxReplicas int					  // 最大虚拟节点数
	HashFunc func(data []byte) uint32 // 哈希函数
	LoadBalanceThreshold float64      // 负载均衡阈值
}

// DefaultConfig 默认配置
var DefaultConfig = &Config{
	DefaultReplicas:      50,
	MinReplicas:          10,
	MaxReplicas:          200,
	HashFunc:             crc32.ChecksumIEEE,
	LoadBalanceThreshold: 0.25,
}
