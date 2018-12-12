package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/golang/protobuf/proto"
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

func BallotNumLessThan(bn *pb.BallotNum, other *pb.BallotNum) bool {
	if bn.BallotIdx != other.BallotIdx {
		return bn.BallotIdx < other.BallotIdx
	}
	for i := 0; i < len(bn.LeaderId) && i < len(other.LeaderId); i++ {
		if bn.LeaderId[i] < other.LeaderId[i] {
			return true
		}
	}
	return len(bn.LeaderId) < len(other.LeaderId)
}

func serve(acceptor *Acceptor, r *rand.Rand, leaders *arrayPeers, id string, port int) {
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
	log.Printf("Serve loop starting")
	for {
		select {
		case pOne := <-acceptor.phaseOneChan:
			log.Printf("Processing Phase One A Msg %v", pOne)
			if BallotNumLessThan(acceptor.ballotNum, pOne.BallotNum) &&
				!proto.Equal(acceptor.ballotNum, pOne.BallotNum) {
				log.Printf("Update ballot num")
				acceptor.ballotNum = pOne.BallotNum
			}
			arg := &pb.PhaseOneBArg{
				AcceptorId: id,
				BallotNum:  acceptor.ballotNum,
				Accepted:   acceptor.accepted}
			// Send in parallel so we don't wait for each client.
			if _, ok := leaderClients[pOne.LeaderId]; ok {
				go func(c pb.LeaderServiceClient, p string) {
					log.Printf("Send PhaseOneB RPC to %v", p)
					log.Printf("with arg: %v", arg)
					_, err := c.PhaseOneB(context.Background(), arg)
					for err != nil {
						_, err = c.PhaseOneB(context.Background(), arg)
					}
				}(leaderClients[pOne.LeaderId], pOne.LeaderId)

			}
		case pTwo := <-acceptor.phaseTwoChan:
			log.Printf("Processing Phase Two A Msg %v", pTwo)
			if !BallotNumLessThan(pTwo.Pv.BallotIdx, acceptor.ballotNum) {
				log.Printf("Update ballot num and accepted")
				acceptor.ballotNum = pTwo.Pv.BallotIdx
				log.Printf("Adding %v pvalue to accepted", pTwo.Pv)
				acceptor.AddAccepted(pTwo.Pv)
			}
			arg := &pb.PhaseTwoBArg{
				AcceptorId: id,
				BallotNum:  acceptor.ballotNum}
			if _, ok := leaderClients[pTwo.LeaderId]; ok {
				// Send in parallel so we don't wait for each client.
				go func(c pb.LeaderServiceClient, p string) {
					log.Printf("Send PhaseTwoB RPC to %v", p)
					log.Printf("with arg: %v", arg)
					_, err := c.PhaseTwoB(context.Background(), arg)
					for err != nil {
						_, err = c.PhaseTwoB(context.Background(), arg)
					}
				}(leaderClients[pTwo.LeaderId], pTwo.LeaderId)
			}
		}
	}
}
