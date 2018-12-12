package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"time"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

type ReplicaServers struct {
	clientId  string
	commandId int64
	replicas  map[string]pb.KvStoreClient
}

func (r *ReplicaServers) sendExecCommand(req *pb.PaxosCommand) (*pb.Result, error) {
	C := make(chan *pb.Result, len(r.replicas))
	for peer, client := range r.replicas {
		go func(p string, c pb.KvStoreClient) {
			log.Printf("Sending command to %v, %v", p, c)
			res, err := c.ExecuteCommand(context.Background(), req)
			for {
				time.Sleep(10 * time.Millisecond)
				if err != nil {
					res, err = c.ExecuteCommand(context.Background(), req)
				} else if res.GetRedirect() != nil {
					log.Fatalf("Should never receive Redirect")
				} else {
					log.Printf("Receive execution from %v", p)
					C <- res
					break
				}
			}
		}(peer, client)
	}
	ret := <-C
	return ret, nil
}

func (r *ReplicaServers) kvGet(key string) string {
	r.commandId++
	req := &pb.Key{Key: key}

	paxosCmd := &pb.PaxosCommand{
		ClientId:  r.clientId,
		CommandId: r.commandId,
		KvOp: &pb.Command{
			Operation: pb.Op_GET,
			Arg: &pb.Command_Get{
				Get: req}}}
	res, err := r.sendExecCommand(paxosCmd)
	if err != nil {
		log.Fatalf("Request error %v", err)
	}

	log.Printf("Got Get response key: \"%v\" value:\"%v\"", res.GetKv().Key, res.GetKv().Value)

	if res.GetKv().Key != key {
		log.Fatalf("Get returned the wrong key response")
	}
	return res.GetKv().Value
}

func (r *ReplicaServers) kvCAS(key string, expected string, toSet string) string {
	r.commandId++
	casReq := &pb.CASArg{Kv: &pb.KeyValue{Key: key, Value: expected}, Value: &pb.Value{Value: toSet}}

	paxosCmd := &pb.PaxosCommand{
		ClientId:  r.clientId,
		CommandId: r.commandId,
		KvOp: &pb.Command{
			Operation: pb.Op_CAS,
			Arg: &pb.Command_Cas{
				Cas: casReq}}}
	res, err := r.sendExecCommand(paxosCmd)
	if err != nil {
		log.Fatalf("Request error %v", err)
	}

	log.Printf("Got CAS response key: \"%v\" value:\"%v\"", res.GetKv().Key, res.GetKv().Value)

	if res.GetKv().Key != key {
		log.Fatalf("Get returned the wrong key response")
	}
	return res.GetKv().Value
}

func usage() {
	fmt.Printf("Usage %s <endpoint>\n", os.Args[0])
	flag.PrintDefaults()
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandStringBytes(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}
func main() {
	// Take endpoint as input
	flag.Usage = usage
	flag.Parse()
	// If there is no endpoint fail
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	endpoints := flag.Args()
	rs := &ReplicaServers{}

	rand.Seed(time.Now().UTC().UnixNano())

	rs.clientId = RandStringBytes(6)
	replicas := make(map[string]pb.KvStoreClient)
	for _, endpoint := range endpoints {
		log.Printf("Connecting to %v", endpoint)
		// Connect to the server. We use WithInsecure since we do not configure https in this class.
		conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
		//Ensure connection did not fail.
		if err != nil {
			log.Fatalf("Failed to dial GRPC server %v", err)
		}
		log.Printf("Connected")
		// Create a KvStore client
		replicas[endpoint] = pb.NewKvStoreClient(conn)
	}

	rs.replicas = replicas

	// Request value for hello

	for i := 0; i < 3; i++ {
		prevVal := rs.kvGet("hello")
		// add 1
		val, _ := strconv.Atoi(prevVal)
		val += rand.Intn(10) + 1
		toPut := strconv.Itoa(val)
		log.Printf("Try putting %v", val)
		// Successfully CAS changing hello -> +1
		// time.Sleep(1 * time.Second)
		res := rs.kvCAS("hello", prevVal, toPut)
		log.Printf("New Value: %v", res)
	}
}
