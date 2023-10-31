package etcd

import (
	"context"
	"etcd-svc/wrr"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"strings"
)

func LoadBalance(redisCli *redis.Client, cli *clientv3.Client, name, filter string) (ban *wrr.Balancer, err error) {
	ban, err = getBalancer(redisCli, cli, name, filter)
	if err != nil {
		return
	}
	watch := cli.Watch(context.Background(), svcPrefix+name+".",
		clientv3.WithPrefix(), clientv3.WithPrevKV())
	go func() {
		for resp := range watch {
			for _, ev := range resp.Events {
				key := string(ev.Kv.Key)
				switch ev.Type {
				case mvccpb.PUT:
					value := string(ev.Kv.Value)
					split := strings.Split(key, ".")
					err := ban.Add(wrr.Server{
						Addr:   value,
						Name:   split[2],
						Weight: 1,
					})
					if err != nil {
						fmt.Printf("put key %v %v\n", key, err)
					}
				case mvccpb.DELETE:
					split := strings.Split(key, ".")
					err := ban.Remove(split[2])
					if err != nil {
						fmt.Printf("del key %v %v\n", key, err)
					}
				}
			}
		}
	}()
	return
}

func getBalancer(redisCli *redis.Client, cli *clientv3.Client, name, filter string) (wr *wrr.Balancer, err error) {
	list, err := SvcList(cli, name)
	if err != nil {
		return
	}
	ss := make([]wrr.Server, 0)
	for _, s := range list {
		if !strings.Contains(s[0], filter) {
			continue
		}
		sp := strings.Split(s[0], ".")
		ss = append(ss, wrr.Server{Name: sp[0], Addr: s[1], Weight: 1})
	}
	if len(ss) == 0 {
		err = fmt.Errorf("no match %s", filter)
		return
	}
	wr, err = wrr.NewBalancer(redisCli, "wrr_"+name+"_"+filter, ss)
	return
}
