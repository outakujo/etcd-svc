package gateway

import (
	"etcd-svc/etcd"
	"net/http"
	"sync"
)

type Svc struct {
	Name    string
	Routers []string
	Host    string
	Path    string
}

type router map[string]Svc

var Router = make(router)

var mut sync.Mutex

var uriSvc = make(map[string]string)

func (r router) Add(svc Svc) {
	mut.Lock()
	defer mut.Unlock()
	_, ok := r[svc.Name]
	if ok {
		return
	}
	for _, s := range svc.Routers {
		key := svc.Host + s
		if svc.Path != "" {
			key = svc.Host + "/" + svc.Path + s
		}
		uriSvc[key] = svc.Name
	}
	r[svc.Name] = svc
}

func (r router) Del(svc string) {
	mut.Lock()
	defer mut.Unlock()
	_, ok := r[svc]
	if !ok {
		return
	}
	delete(r, svc)
	for k, s := range uriSvc {
		if s == svc {
			delete(uriSvc, k)
		}
	}
}

func Match(req *http.Request) string {
	host := req.Host
	ur := req.URL
	uri := host + ur.Path + "-" + req.Method
	return etcd.RegisterSvcs.Get(uriSvc[uri])
}
