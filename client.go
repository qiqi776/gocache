package gocache

import (
	"context"
	"fmt"
	pb "gocache/pb"
	"gocache/registry"
	"time"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type Client struct {
	addr string
	svcName string
	etcdCli *clientv3.Client
	conn *grpc.ClientConn
	grpcCli pb.GoCacheClient
}

var _ Peer = (*Client)(nil)

func NewClient(addr string, svcName string, etcdCli *clientv3.Client) (*Client, error) {
	var err error
	if etcdCli == nil {
		etcdCli, err = clientv3.New(clientv3.Config{
			Endpoints: []string{"localhost:2379"},
			DialTimeout: 5 * time.Second,
		})
		if err != nil {
			return nil, err
		}
	}
	conn, err := registry.EtcdDial(etcdCli, svcName, addr)
	if err != nil {
		return nil, err
	}
	grpcCli := pb.NewGoCacheClient(conn)
	client := &Client{
		addr: addr,
		svcName: svcName,
		etcdCli: etcdCli,
		conn: conn,
		grpcCli: grpcCli,
	}
	return client, nil
}

func (c *Client) Get(group, key string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.grpcCli.Get(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get value from lcache: %v", err)
	}

	return resp.GetValue(), nil
}

func (c *Client) Delete(group, key string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	resp, err := c.grpcCli.Delete(ctx, &pb.Request{
		Group: group,
		Key:   key,
	})
	if err != nil {
		return false, fmt.Errorf("failed to delete value from lcache: %v", err)
	}

	return resp.GetValue(), nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
