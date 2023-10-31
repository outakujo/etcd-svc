package wrr

import (
	"context"
	"embed"
	"github.com/redis/go-redis/v9"
	"sync"
)

//go:embed wrr.lua
var fs embed.FS

type Balancer struct {
	cli      *redis.Client
	key      string
	servers  []Server
	script   string
	scriptId string
	mut      sync.Mutex
}

type Server struct {
	Addr   string
	Name   string
	Weight int
}

func NewBalancer(cli *redis.Client, key string, servers []Server) (*Balancer, error) {
	file, err := fs.ReadFile("wrr.lua")
	if err != nil {
		return nil, err
	}
	b := &Balancer{
		cli:     cli,
		key:     key,
		script:  string(file),
		servers: servers,
	}
	hss := make([]interface{}, 0)
	zs := make([]redis.Z, 0)
	for _, s := range servers {
		hss = append(hss, s.Name+"_weight", s.Weight)
		hss = append(hss, s.Name+"_addr", s.Addr)
		zs = append(zs, redis.Z{
			Member: s.Name,
		})
	}
	err = cli.Del(context.Background(), key+"_meta", key+"_servers").Err()
	if err != nil {
		return nil, err
	}
	err = cli.HMSet(context.Background(), key+"_meta", hss...).Err()
	if err != nil {
		return nil, err
	}
	err = cli.ZAdd(context.Background(), key+"_servers", zs...).Err()
	if err != nil {
		return nil, err
	}
	b.scriptId, err = cli.ScriptLoad(context.Background(), b.script).Result()
	return b, err
}

func (r *Balancer) Add(s Server) error {
	r.mut.Lock()
	defer r.mut.Unlock()
	hss := make([]interface{}, 0)
	hss = append(hss, s.Name+"_weight", s.Weight)
	hss = append(hss, s.Name+"_addr", s.Addr)
	err := r.cli.HMSet(context.Background(), r.key+"_meta", hss...).Err()
	if err != nil {
		return err
	}
	err = r.cli.ZAdd(context.Background(), r.key+"_servers",
		redis.Z{Member: s.Name}).Err()
	return err
}

func (r *Balancer) Remove(name string) error {
	r.mut.Lock()
	defer r.mut.Unlock()
	err := r.cli.HDel(context.Background(), r.key+"_meta", name+"_weight",
		name+"_addr").Err()
	if err != nil {
		return err
	}
	err = r.cli.ZRem(context.Background(), r.key+"_servers",
		name).Err()
	return err
}

func (r *Balancer) Next() (next Server, err error) {
	sli, err := r.cli.EvalSha(context.Background(), r.scriptId,
		[]string{r.key + "_meta", r.key + "_servers"}).Slice()
	if err != nil {
		return
	}
	next.Addr = sli[0].(string)
	next.Name = sli[1].(string)
	next.Weight = int(sli[2].(int64))
	return
}
