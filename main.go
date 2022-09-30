package main

import (
	"etcd-svc/etcd"
	"etcd-svc/svc"
	"flag"
)

func main() {
	sv := flag.String("svc", "svc", "svc name")
	addr := flag.String("addr", "addr", "svc addr")
	port := flag.Int("port", 8080, "svc port")
	flag.Parse()
	err := etcd.Register([]string{"localhost:2379"}, *sv, *addr)
	if err != nil {
		panic(err)
	}
	err = svc.Run(*sv, *port)
	panic(err)
}
