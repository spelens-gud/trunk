package server

import (
	"sync"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Cache struct {
	cli           *clientv3.Client
	leaseID       clientv3.LeaseID
	keepAliveChan <-chan *clientv3.LeaseKeepAliveResponse
	key           string
	val           string

	lock sync.RWMutex // 读写锁
}
