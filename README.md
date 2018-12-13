# Multi-Creed Paxos Key Store
This project implements a Key-Store replicated state machine upon multi-creed paxos.

## Compile and Run
```shell
protoc -I pb pb/kv.proto --go_out=plugins=grpc:pb
./create-docker-image.sh
```

## Boot the pods
```shell
./launch-tool/launch.py boot 2 1 3
```
for booting 2 replicas, 1 leader and 3 acceptors

## Run client
use
```shell
kubectl get svc
```
to see the port of rep0, rep1, ...

run
```shell
go run ./client/ [ip]:port1, [ip]:port2
```
where [ip] is the result of 
```shell
minikube ip
```

## Kill one pod
```shell
./launch-tool/launch.py kill acc0
```
to kill acc0 pod