package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strconv"
	"time"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/proto"
	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func propose(r *Replica, command *pb.PaxosCommand, leaders map[string]pb.LeaderServiceClient) {
	for _, v := range r.decisions {
		if proto.Equal(command, v) {
			return
		}
	}
	s := r.FindSlot()
	prop := &pb.Proposal{
		SlotIdx: s,
		Command: command}
	r.AddProposal(prop)
	for p, c := range leaders {
		// Send in parallel so we don't wait for each client.
		go func(c pb.LeaderServiceClient, p string) {
			log.Printf("Send Propose RPC to %v", p)
			c.Propose(context.Background(), prop)
		}(c, p)
	}
}

func perform(r *Replica, s *KVStore, command *pb.PaxosCommand, leaders map[string]pb.LeaderServiceClient, response chan pb.Result) {
	if r.CheckDecision(command) {
		r.slotNum++
	} else {
		// Execute the command, and return the response
		s.HandleCommand(InputChannelType{
			command:  *command.KvOp,
			response: response})
		r.slotNum++
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

func getStringRepresent(cmd *pb.PaxosCommand) string {
	return cmd.ClientId + strconv.FormatInt(cmd.CommandId, 10)
}

func serve(s *KVStore, replica *Replica, r *rand.Rand, peers *arrayPeers, id string, port int) {
	go RunReplicaServiceServer(replica, port)

	leaders := make(map[string]pb.LeaderServiceClient)
	for _, peer := range *peers {
		client, err := connectToPeer(peer)
		if err != nil {
			log.Fatalf("Failed to connect to GRPC server %v", err)
		}

		leaders[peer] = client
		log.Printf("Connected to %v", peer)
	}

	// response map
	responseMap := make(map[string]chan pb.Result)

	// serve loop
	for {
		select {
		case op := <-s.C:
			log.Fatalf("Obsolete RPC calls %v", op.command)
		case pOp := <-s.Q:
			log.Printf("Receive %v", pOp.command)
			responseMap[getStringRepresent(pOp.command)] = pOp.response
			propose(replica,
				pOp.command, leaders)
		case decision := <-replica.decisionChan:
			log.Printf("Receive decision %v", decision.decision)
			for replica.CheckCurrentSlotNum() {
				for _, v := range replica.proposals {
					if v.SlotIdx == replica.slotNum && !proto.Equal(v.Command, decision.decision) {
						propose(replica, v.Command, leaders)
						break
					}
				}
			}
			respChan, ok := responseMap[getStringRepresent(decision.decision.Command)]
			if !ok {
				log.Fatalf("No such response channel")
			}
			perform(replica, s, decision.decision.Command, leaders, respChan)
		case <-replica.quitChan:
			log.Printf("Quit the serve loop")
			break
		}
	}
}
