package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func connectToLeader(peer string) (pb.LeaderServiceClient, error) {
	backoffConfig := grpc.DefaultBackoffConfig
	// Choose an aggressive backoff strategy here.
	backoffConfig.MaxDelay = 500 * time.Millisecond
	conn, err := grpc.Dial(peer, grpc.WithInsecure(), grpc.WithBackoffConfig(backoffConfig))
	// Ensure connection did not fail, which should not happen since this happens in the background
	if err != nil {
		return pb.NewLeaderServiceClient(nil), err
	}
	return pb.NewLeaderServiceClient(conn), nil
}

//RunAcceptorServiceServer launches a ReplicaService server
func RunAcceptorServiceServer(acc *Acceptor, port int) {
	// Convert port to a string form
	portString := fmt.Sprintf(":%d", port)
	// Create socket that listens on the supplied port
	c, err := net.Listen("tcp", portString)
	if err != nil {
		// Note the use of Fatalf which will exit the program after reporting the error.
		log.Fatalf("Could not create listening socket %v", err)
	}
	// Create a new GRPC server
	s := grpc.NewServer()

	pb.RegisterAcceptorServiceServer(s, acc)
	log.Printf("Going to listen on port %v", port)

	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func serve(r *rand.Rand, leaders *arrayPeers, id string, port int) {
	acceptor := NewAcceptor()
	go RunAcceptorServiceServer(acceptor, port)

	leaderClients := make(map[string]pb.LeaderServiceClient)
	for _, peer := range *leaders {
		client, err := connectToLeader(peer)
		if err != nil {
			log.Fatalf("Failed to connect to GRPC server %v", err)
		}

		leaderClients[peer] = client
		log.Printf("Connected to %v", peer)
	}

	// serve loop
	for {
		select {}
	}
}
