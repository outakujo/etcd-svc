package main

import (
	"context"
	"etcd-svc/etcd"
	"etcd-svc/gateway/pb"
	"etcd-svc/svc"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	req := &pb.Req{
		Name: *sv,
		Routers: []string{
			"/user-GET",
			"/good-GET",
		},
		Host: "localhost:9001",
		Path: "",
	}
	gateway := etcd.RegisterSvcs.Get("gateway")
	con, err := grpc.Dial(gateway, grpc.WithTransportCredentials(insecure.
		NewCredentials()))
	if err != nil {
		panic(err)
	}
	client := pb.NewServiceClient(con)
	_, err = client.Add(context.TODO(), req)
	if err != nil {
		panic(err)
	}
	err = svc.Run(*sv, *port)
	panic(err)
}
