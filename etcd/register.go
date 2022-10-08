package etcd

import (
	"context"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"
)

func Register(endpoints []string, svc, addr string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	locker, err := NewLocker(endpoints, "register")
	if err != nil {
		return err
	}
	err = refreshRegister(endpoints, svc)
	if err != nil {
		return err
	}
	locker.Lock()
	defer locker.Unlock()
	response, err := cli.Get(context.TODO(), "svc/"+svc, clientv3.WithPrefix(),
		clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return err
	}
	count := response.Count
	grant, err := cli.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}
	key := "svc/" + svc + "/" + strconv.Itoa(int(count))
	if count != 0 {
		bytes := response.Kvs[0].Key
		split := strings.Split(string(bytes), "/")
		atoi, err := strconv.Atoi(split[2])
		if err != nil {
			return err
		}
		key = "svc/" + svc + "/" + strconv.Itoa(atoi+1)
	}
	_, err = cli.Put(context.TODO(), key, addr, clientv3.WithLease(grant.ID))
	if err != nil {
		return err
	}
	alive, err := cli.KeepAlive(context.TODO(), grant.ID)
	if err != nil {
		return err
	}
	go func() {
		for range alive {
		}
	}()
	return nil
}

func NewLocker(endpoints []string, prefix string) (sync.Locker, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	session, err := concurrency.NewSession(cli)
	if err != nil {
		return nil, err
	}
	return concurrency.NewLocker(session, prefix), nil
}

type registerSvcs map[string][]string

var RegisterSvcs = make(registerSvcs)

var mut sync.Mutex

func (r registerSvcs) Get(svc string) string {
	mut.Lock()
	defer mut.Unlock()
	ss, ok := r[svc]
	if !ok {
		return ""
	}
	ln := len(ss)
	return ss[rand.Intn(ln)]
}

func (r registerSvcs) Put(key, val string) {
	mut.Lock()
	defer mut.Unlock()
	split := strings.Split(key, "/")
	ss, ok := r[split[1]]
	if !ok {
		ss = make([]string, 0)
	}
	ss = append(ss, val)
	r[split[1]] = ss
}

func (r registerSvcs) Del(key, val string) {
	mut.Lock()
	defer mut.Unlock()
	split := strings.Split(key, "/")
	ss, ok := r[split[1]]
	if !ok {
		return
	}
	ns := make([]string, 0)
	for _, s := range ss {
		if s != val {
			ns = append(ns, s)
		}
	}
	r[split[1]] = ns
}

func refreshRegister(endpoints []string, svc string) error {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	response, err := cli.Get(context.TODO(), "svc/", clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kv := range response.Kvs {
		key := string(kv.Key)
		split := strings.Split(key, "/")
		if split[1] == svc {
			continue
		}
		RegisterSvcs.Put(key, string(kv.Value))
	}
	watch := cli.Watch(context.TODO(), "svc/", clientv3.WithPrefix(), clientv3.WithPrevKV())
	go func() {
		for resp := range watch {
			for _, ev := range resp.Events {
				key := string(ev.Kv.Key)
				split := strings.Split(key, "/")
				if split[1] == svc {
					continue
				}
				switch ev.Type {
				case mvccpb.PUT:
					RegisterSvcs.Put(key, string(ev.Kv.Value))
				case mvccpb.DELETE:
					RegisterSvcs.Del(key, string(ev.PrevKv.Value))
				}
			}
		}
	}()
	return nil
}
