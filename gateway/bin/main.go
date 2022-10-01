package main

import (
	"context"
	"etcd-svc/etcd"
	"etcd-svc/gateway"
	"etcd-svc/gateway/pb"
	"flag"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
	"net"
	"strconv"
)

func main() {
	port := flag.Int("port", 9001, "port")
	flag.Parse()
	err := etcd.Register([]string{"localhost:2379"}, "gateway",
		"localhost:"+strconv.Itoa(*port+1))
	if err != nil {
		panic(err)
	}
	go rpcServer(*port)
	err = gateway.Run(*port)
	panic(err)
}

type Server struct {
	pb.UnimplementedServiceServer
}

func (r *Server) Add(_ context.Context, req *pb.Req) (*emptypb.Empty, error) {
	gateway.Router.Add(gateway.Svc{
		Name:    req.Name,
		Routers: req.Routers,
		Host:    req.Host,
		Path:    req.Path,
	})
	return &emptypb.Empty{}, nil
}

func rpcServer(port int) {
	var ser Server
	server := grpc.NewServer()
	pb.RegisterServiceServer(server, &ser)
	listen, err := net.Listen("tcp", ":"+strconv.Itoa(port+1))
	if err != nil {
		panic(err)
	}
	err = server.Serve(listen)
	if err != nil {
		panic(err)
	}
}
