#! /bin/bash

bin=etcd-svc
go build
chmod +x $bin

./$bin -svc user -port 8081 -addr localhost:8081 &
pids[0]=$!
./$bin -svc user -port 8082 -addr localhost:8082 &
pids[1]=$!
./$bin -svc good -port 8083 -addr localhost:8083 &
pids[2]=$!
./$bin -svc good -port 8084 -addr localhost:8084 &
pids[3]=$!
echo ${pids[*]} >pids
