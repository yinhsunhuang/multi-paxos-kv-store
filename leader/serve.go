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

func connectToReplica(peer string) (pb.ReplicaServiceClient, error) {
	backoffConfig := grpc.DefaultBackoffConfig
	// Choose an aggressive backoff strategy here.
	backoffConfig.MaxDelay = 500 * time.Millisecond
	conn, err := grpc.Dial(peer, grpc.WithInsecure(), grpc.WithBackoffConfig(backoffConfig))
	// Ensure connection did not fail, which should not happen since this happens in the background
	if err != nil {
		return pb.NewReplicaServiceClient(nil), err
	}
	return pb.NewReplicaServiceClient(conn), nil
}

func connectToAcceptor(peer string) (pb.AcceptorServiceClient, error) {
	backoffConfig := grpc.DefaultBackoffConfig
	// Choose an aggressive backoff strategy here.
	backoffConfig.MaxDelay = 500 * time.Millisecond
	conn, err := grpc.Dial(peer, grpc.WithInsecure(), grpc.WithBackoffConfig(backoffConfig))
	// Ensure connection did not fail, which should not happen since this happens in the background
	if err != nil {
		return pb.NewAcceptorServiceClient(nil), err
	}
	return pb.NewAcceptorServiceClient(conn), nil
}

//RunLeaderServiceServer launches a ReplicaService server
func RunLeaderServiceServer(l *Leader, port int) {
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

	pb.RegisterLeaderServiceServer(s, l)
	log.Printf("Going to listen on port %v", port)

	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func serve(r *rand.Rand, replicas *arrayPeers, acceptors *arrayPeers, id string, port int) {
	leader := NewLeader()
	go RunLeaderServiceServer(leader, port)

	replicaClients := make(map[string]pb.ReplicaServiceClient)
	for _, peer := range *replicas {
		client, err := connectToReplica(peer)
		if err != nil {
			log.Fatalf("Failed to connect to GRPC server %v", err)
		}

		replicaClients[peer] = client
		log.Printf("Connected to %v", peer)
	}

	acceptorClients := make(map[string]pb.AcceptorServiceClient)
	for _, peer := range *acceptors {
		client, err := connectToAcceptor(peer)
		if err != nil {
			log.Fatalf("Failed to connect to GRPC server %v", err)
		}

		acceptorClients[peer] = client
		log.Printf("Connected to %v", peer)
	}

	// serve loop
	for {
		select {
		case prop := <-leader.proposeChan:

		}
	}
}
