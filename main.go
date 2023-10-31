package main

import (
	"etcd-svc/etcd"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

func main() {
	cli, err := etcd.InitClient([]string{"localhost:2379"})
	if err != nil {
		panic(err)
	}
	err = etcd.Register(cli, "http://123.123.123.123", "user", "node1", "http")
	if err != nil {
		fmt.Println(err)
		return
	}
	err = etcd.Register(cli, "http://123.123.123.124", "user", "node2", "http")
	if err != nil {
		fmt.Println(err)
		return
	}
	list, err := etcd.SvcList(cli, "user")
	if err != nil {
		panic(err)
	}
	fmt.Println(list)
	err = etcd.DeRegister(cli, "user", "node1")
	if err != nil {
		fmt.Println(err)
	}
	list, err = etcd.SvcList(cli, "user")
	if err != nil {
		panic(err)
	}
	fmt.Println(list)
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379"})
	userBalance, err := etcd.LoadBalance(client, cli, "user", "http")
	if err != nil {
		panic(err)
	}
	for i := 0; i < 6; i++ {
		next, err := userBalance.Next()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(next)
	}
	fmt.Println("---------")
	go func() {
		time.Sleep(3 * time.Second)
		err = etcd.Register(cli, "http://123.123.123.123", "user", "node1", "http")
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	go func() {
		time.Sleep(7 * time.Second)
		err = etcd.DeRegister(cli, "user", "node2")
		if err != nil {
			fmt.Println(err)
			return
		}
	}()
	for i := 0; i < 16; i++ {
		next, err := userBalance.Next()
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(next)
		time.Sleep(time.Second)
	}
}
