# GoCache 🚀

**GoCache** 是一个用 Go 语言实现的轻量级分布式缓存系统。它支持 LRU 淘汰策略、基于 gRPC 的节点通信、一致性哈希路由以及防缓存击穿机制。

## ✨ 特性 (Features)

- **核心存储**: 基于 **LRU** (Least Recently Used) 的内存淘汰策略。
- **并发控制**: 使用 **Singleflight** 机制防止缓存击穿（Thundering Herd）。
- **分布式**: 实现了 **一致性哈希 (Consistent Hashing)** 进行节点选择和负载均衡。
- **通信**: 高性能的 **gRPC** 节点间通信。
- **易用性**: 简单的 Group 命名空间管理和回调回源机制。

## 📦 目录结构

```text
.
├── consistenthash/  # 一致性哈希算法
├── pb/              # gRPC Protobuf 定义及生成代码
├── singleflight/    # 请求合并机制
├── store/           # 核心 LRU 存储实现
├── byteview.go      # 不可变字节视图
├── group.go         # 核心调度逻辑 (Cache Miss/Hit 处理)
├── server.go        # gRPC 服务端实现
└── peers.go         # 节点抽象接口

```
