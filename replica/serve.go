package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func propose(r *Replica, command *pb.PaxosCommand, peerClients map[string]pb.LeaderServiceClient) {
	s := r.FindSlot()
	prop := &pb.Proposal{
		SlotIdx: s,
		Command: command}
	r.AddProposal(prop)
	for p, c := range peerClients {
		// Send in parallel so we don't wait for each client.
		go func(c pb.LeaderServiceClient, p string) {
			log.Printf("Send Propose RPC to %v", p)
			c.Propose(context.Background(), prop)
		}(c, p)
	}
}

func connectToPeer(peer string) (pb.LeaderServiceClient, error) {
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

//RunReplicaServiceServer launches a ReplicaService server
func RunReplicaServiceServer(r *Replica, port int) {
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

	pb.RegisterReplicaServiceServer(s, r)
	log.Printf("Going to listen on port %v", port)

	// Start serving, this will block this function and only return when done.
	if err := s.Serve(c); err != nil {
		log.Fatalf("Failed to serve %v", err)
	}
}

func serve(s *KVStore, r *rand.Rand, peers *arrayPeers, id string, port int) {
	replica := NewReplica()
	go RunReplicaServiceServer(replica, port)

	peerClients := make(map[string]pb.LeaderServiceClient)
	for _, peer := range *peers {
		client, err := connectToPeer(peer)
		if err != nil {
			log.Fatalf("Failed to connect to GRPC server %v", err)
		}

		peerClients[peer] = client
		log.Printf("Connected to %v", peer)
	}

	// serve loop
	for {
		select {
		case op := <-s.C:
			log.Printf("{%v}", op.command)
			propose(replica,
				&pb.PaxosCommand{
					ClientId:  "222",
					CommandId: 3,
					Operation: &op.command}, peerClients)
			s.HandleCommand(op)
		}
	}
}
