package etcdv3

import (
	"context"
	"fmt"
	"github.com/coreos/etcd/clientv3"
	"net"
	"strings"
)

// Prefix should start and end with no slash
var Deregister = make(chan struct{})

// Register
// target:etcdAddr,use "," to split
// service:service name
// host:service host
// port:service port
// ttl:seconds
func Register(target, service, host, port string, ttl int) error {
	serviceValue := net.JoinHostPort(host, port)
	serviceKey := fmt.Sprintf("/%s/%s/%s", schema, service, serviceValue)

	// get endpoints for register dial address
	var err error
	cli, err := clientv3.New(clientv3.Config{
		Endpoints: strings.Split(target, ","),
	})
	if err != nil {
		return fmt.Errorf("gateway: create clientv3 client failed: %v", err)
	}
	resp, err := cli.Grant(context.TODO(), int64(ttl))
	if err != nil {
		return fmt.Errorf("gateway: create clientv3 lease failed: %v", err)
	}

	if _, err := cli.Put(context.TODO(), serviceKey, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
		return fmt.Errorf("gateway: set service '%s' with ttl to clientv3 failed: %s", service, err.Error())
	}

	keepaliveCtx, keepaliveCancel := context.WithCancel(context.TODO())
	keepAliveResponse, err := cli.KeepAlive(keepaliveCtx, resp.ID)
	if err != nil {
		return fmt.Errorf("gateway: refresh service '%s' with ttl to clientv3 failed: %s", service, err.Error())
	}

	// 进行保活返回值的消费，避免堵塞后以每秒一次的频率告诉续期
	go func(keepAliveResponse <-chan *clientv3.LeaseKeepAliveResponse) {
		for x := range keepAliveResponse {
			_ = x
		}
	}(keepAliveResponse)

	// wait deregister then delete
	go func() {
		<-Deregister
		cli.Delete(context.Background(), serviceKey)
		keepaliveCancel()
		Deregister <- struct{}{}
	}()

	return nil
}

// UnRegister delete registered service from etcd
// 一进一出，避免堵塞
func UnRegister() {
	Deregister <- struct{}{}
	<-Deregister
}
