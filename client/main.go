package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"

	context "golang.org/x/net/context"
	"google.golang.org/grpc"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func usage() {
	fmt.Printf("Usage %s <endpoint>\n", os.Args[0])
	flag.PrintDefaults()
}

// TODO: Deal with redirect
func main() {
	// Take endpoint as input
	flag.Usage = usage
	flag.Parse()
	// If there is no endpoint fail
	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(1)
	}
	endpoint := flag.Args()[0]
	log.Printf("Connecting to %v", endpoint)
	// Connect to the server. We use WithInsecure since we do not configure https in this class.
	conn, err := grpc.Dial(endpoint, grpc.WithInsecure())
	//Ensure connection did not fail.
	if err != nil {
		log.Fatalf("Failed to dial GRPC server %v", err)
	}
	log.Printf("Connected")
	// Create a KvStore client
	kvc := pb.NewKvStoreClient(conn)
	// Clear KVC
	// log.Printf("Clear kvstore")
	// res, err := kvc.Clear(context.Background(), &pb.Empty{})
	// if err != nil {
	// 	log.Fatalf("Could not clear")
	// }
	// if t := res.GetRedirect(); t != nil {
	// 	log.Fatalf("Got Redirect to %v", t.Server)
	// }

	// Put setting hello -> 1
	// putReq := &pb.KeyValue{Key: "hello", Value: "1"}
	// res, err = kvc.Set(context.Background(), putReq)
	// if err != nil {
	// 	log.Fatalf("Put error")
	// }
	// log.Printf("Got response key: \"%v\" value:\"%v\"", res.GetKv().Key, res.GetKv().Value)
	// if res.GetKv().Key != "hello" || res.GetKv().Value != "1" {
	// 	log.Fatalf("Put returned the wrong response")
	// }

	// Request value for hello

	for i := 0; i < 3; i++ {
		req := &pb.Key{Key: "hello"}
		res, err := kvc.Get(context.Background(), req)
		if err != nil {
			log.Fatalf("Request error %v", err)
		}

		if res.GetRedirect() != nil {
			log.Fatalf("Redirect to %v", res.GetRedirect().Server)
		}

		log.Printf("Got Get response key: \"%v\" value:\"%v\"", res.GetKv().Key, res.GetKv().Value)

		if res.GetKv().Key != "hello" {
			log.Fatalf("Get returned the wrong response")
		}

		prevVal := res.GetKv().Value
		// add 1
		val, _ := strconv.Atoi(prevVal)
		val += rand.Intn(10)
		toPut := strconv.Itoa(val)
		log.Printf("Try putting %v", val)
		// Successfully CAS changing hello -> +1
		casReq := &pb.CASArg{Kv: &pb.KeyValue{Key: "hello", Value: prevVal}, Value: &pb.Value{Value: toPut}}
		res, err = kvc.CAS(context.Background(), casReq)
		if err != nil {
			log.Fatalf("Request error %v", err)
		}

		if res.GetRedirect() != nil {
			log.Fatalf("Redirect to %v", res.GetRedirect().Server)
		}

		log.Printf("Got CAS response key: \"%v\" value:\"%v\"", res.GetKv().Key, res.GetKv().Value)
		if res.GetKv().Key != "hello" {
			log.Fatalf("Get returned the wrong response")
		}
	}
}
