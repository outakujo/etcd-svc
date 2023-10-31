package etcd

import (
	"context"
	"errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"strings"
	"sync"
	"time"
)

const (
	svcPrefix = "svc."
)

func InitClient(endpoints []string) (cli *clientv3.Client, err error) {
	cli, err = clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	return
}

func Register(cli *clientv3.Client, value string, keys ...string) error {
	if len(keys) == 0 {
		return errors.New("info not can be empty")
	}
	locker, err := NewLocker(cli, "register")
	if err != nil {
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	join := strings.Join(keys, ".")
	key := svcPrefix + join
	response, err := cli.Get(context.Background(), key)
	if err != nil {
		return err
	}
	if response.Count != 0 {
		return errors.New("already register")
	}
	grant, err := cli.Grant(context.Background(), 5)
	if err != nil {
		return err
	}
	_, err = cli.Put(context.Background(), key, value, clientv3.WithLease(grant.ID))
	if err != nil {
		return err
	}
	alive, err := cli.KeepAlive(context.Background(), grant.ID)
	if err != nil {
		return err
	}
	go func() {
		for range alive {
		}
	}()
	return nil
}

func NewLocker(cli *clientv3.Client, prefix string) (sync.Locker, error) {
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return nil, err
	}
	return concurrency.NewLocker(session, prefix), nil
}

func DeRegister(cli *clientv3.Client, keys ...string) error {
	join := strings.Join(keys, ".")
	key := svcPrefix + join
	response, err := cli.Delete(context.Background(), key, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	if response.Deleted == 0 {
		return errors.New("deRegister failed")
	}
	return nil
}

func SvcList(cli *clientv3.Client, name string) (rs [][]string, err error) {
	response, err := cli.Get(context.TODO(), svcPrefix+name+".", clientv3.WithPrefix())
	if err != nil {
		return
	}
	rs = make([][]string, 0)
	for _, kv := range response.Kvs {
		key := string(kv.Key)
		split := strings.Split(key, ".")
		join := strings.Join(split[2:], ".")
		rs = append(rs, []string{join, string(kv.Value)})
	}
	return
}
