package gateway

import (
	"context"
	"encoding/json"
	"etcd-svc/etcd"
	"fmt"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"net/http"
	"strings"
	"time"
)

type Svc struct {
	Name    string
	Routers []string
	Host    string
	Path    string
}

type router map[string]string

var Router = make(router)

var client *clientv3.Client

func InitClient(endpoints []string) error {
	if client != nil {
		return nil
	}
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: 5 * time.Second,
	})
	client = cli
	return err
}

func (r router) Add(svc Svc) error {
	grant, err := client.Grant(context.TODO(), 5)
	if err != nil {
		return err
	}
	key := "router/" + svc.Name
	apis := make([]string, 0)
	for _, s := range svc.Routers {
		api := svc.Host + s
		if svc.Path != "" {
			api = svc.Host + "/" + svc.Path + s
		}
		apis = append(apis, api)
	}
	marshal, err := json.Marshal(apis)
	if err != nil {
		return err
	}
	_, err = client.Put(context.TODO(), key, string(marshal), clientv3.WithLease(grant.ID))
	alive, err := client.KeepAlive(context.TODO(), grant.ID)
	if err != nil {
		return err
	}
	go func() {
		for range alive {
		}
	}()
	return err
}

func (r router) Del(svc string) error {
	key := "router/" + svc
	_, err := client.Delete(context.TODO(), key)
	return err
}

func (r router) Match(req *http.Request) string {
	host := req.Host
	ur := req.URL
	uri := host + ur.Path + "-" + req.Method
	svc := r[uri]
	return etcd.RegisterSvcs.Get(svc)
}

func RefreshRouter() error {
	key := "router/"
	response, err := client.Get(context.TODO(), key, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	for _, kv := range response.Kvs {
		key := string(kv.Key)
		split := strings.Split(key, "/")
		svc := split[1]
		value := kv.Value
		apis := make([]string, 0)
		err := json.Unmarshal(value, &apis)
		if err != nil {
			return err
		}
		for _, a := range apis {
			Router[a] = svc
		}
	}
	watch := client.Watch(context.TODO(), key, clientv3.WithPrefix())
	go func() {
		for resp := range watch {
			for _, ev := range resp.Events {
				key := string(ev.Kv.Key)
				split := strings.Split(key, "/")
				svc := split[1]
				switch ev.Type {
				case mvccpb.PUT:
					value := ev.Kv.Value
					apis := make([]string, 0)
					err := json.Unmarshal(value, &apis)
					if err != nil {
						fmt.Println(err)
						continue
					}
					for _, a := range apis {
						Router[a] = svc
					}
				case mvccpb.DELETE:
					get := etcd.RegisterSvcs.Get(svc)
					if get != "" {
						continue
					}
					for k, v := range Router {
						if v == svc {
							delete(Router, k)
						}
					}
				}
			}
		}
	}()
	return nil
}
