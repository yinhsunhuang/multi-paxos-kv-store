package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"google.golang.org/grpc"

	"github.com/golang/protobuf/proto"
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

func (l *Leader) sendAcceptorsPhaseOneA(id string, acceptorClients map[string]pb.AcceptorServiceClient, b pb.BallotNum) {
	for p, c := range acceptorClients {
		go func(c pb.AcceptorServiceClient, p string) {
			log.Printf("Send PhaseOneA RPC to %v", p)
			_, err := c.PhaseOneA(context.Background(), &pb.PhaseOneArg{
				LeaderId:  id,
				BallotNum: &b})
			if err != nil {
				log.Printf("Error during PhaseOneA RPC call to %v", p)
			}
		}(c, p)
	}
}

func (l *Leader) sendAcceptorsPhaseTwoA(id string, acceptorClients map[string]pb.AcceptorServiceClient, pv pb.Pvalue) {
	for p, c := range acceptorClients {
		go func(c pb.AcceptorServiceClient, p string) {
			log.Printf("Send PhaseTwoA RPC to %v", p)
			_, err := c.PhaseTwoA(context.Background(), &pb.PhaseTwoArg{
				LeaderId: id,
				Pv:       &pv})
			if err != nil {
				log.Printf("Error during PhaseTwoA RPC call to %v", p)
			}
		}(c, p)
	}
}

func (l *Leader) sendReplicasDecision(id string, replicaClients map[string]pb.ReplicaServiceClient, prop pb.Proposal) {
	for p, c := range replicaClients {
		go func(c pb.ReplicaServiceClient, p string) {
			log.Printf("Send PhaseTwoA RPC to %v", p)
			_, err := c.Decision(context.Background(), &prop)
			if err != nil {
				log.Printf("Error during ReplicaDecision RPC call to %v", p)
			}
		}(c, p)
	}
}

func serve(r *rand.Rand, replicas *arrayPeers, acceptors *arrayPeers, id string, port int) {
	leader := NewLeader(id)
	go RunLeaderServiceServer(leader, port)

	log.Printf("Sleep for two second so that all node is up")
	time.Sleep(4 * time.Second)
	log.Printf("Sleep Done")

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
	var scoutBallotNum *pb.BallotNum
	scoutPvalue := make([]*pb.Pvalue, 0, 0)
	scoutWaitFor := make(map[string]bool)

	commanderWaitFor := make(map[string]bool)

	// serve loop
	// Spawn scout
	leader.scoutArg = leader.ballotNum
	leader.sendAcceptorsPhaseOneA(id, acceptorClients, *leader.ballotNum)
	for {
		select {
		case prop := <-leader.proposeChan:
			log.Printf("Processing  Propose Message: %v", prop)
			leader.AddProposal(prop)
			if leader.active {
				//Spawn Commander
				leader.commanderArg = []*pb.Pvalue{&pb.Pvalue{
					BallotIdx: leader.ballotNum,
					SlotIdx:   prop.SlotIdx,
					Command:   prop.Command}}
				leader.sendAcceptorsPhaseTwoA(id, acceptorClients, *leader.commanderArg[0])
			}
		case arg := <-leader.adoptedChan:
			log.Printf("Processing Adopted Message")
			leader.proposals = CombineOp(leader.proposals, Pmax(arg.pvals))
			// Spawn Commander arrays
			leader.commanderArg = make([]*pb.Pvalue, 0, 0)
			for _, v := range arg.pvals {
				leader.commanderArg = append(leader.commanderArg, v)
			}
			leader.active = true
		case preempted := <-leader.preemptedChan:
			log.Printf("Processing Preempted Message %v", preempted)
			if BallotNumLessThan(leader.ballotNum, preempted) {
				log.Printf("Found greater ballot_num, return to passive mode")
				leader.active = false
				leader.ballotNum = &pb.BallotNum{BallotIdx: leader.ballotNum.BallotIdx + 1,
					LeaderId: leader.ballotNum.LeaderId}
				// Spawn scout
				leader.scoutArg = leader.ballotNum
				leader.sendAcceptorsPhaseOneA(id, acceptorClients, *leader.ballotNum)
			}
		case pOne := <-leader.phaseOneChan:
			log.Printf("Processing PhaseOneB")
			if leader.scoutArg != nil {
				// Scout is active
				if proto.Equal(pOne.BallotNum, leader.ballotNum) {
					for _, v := range pOne.Accepted {
						scoutPvalue = append(scoutPvalue, v)
					}
					delete(scoutWaitFor, pOne.AcceptorId)
					if len(scoutWaitFor) < acceptors.Num()/2 {
						leader.adoptedChan <- AdoptedInputType{
							ballotNum: scoutBallotNum,
							pvals:     scoutPvalue}
						leader.scoutArg = nil
					}
				} else {
					leader.preemptedChan <- pOne.BallotNum
					leader.scoutArg = nil
				}
			}
		case pTwo := <-leader.phaseTwoChan:
			log.Printf("Processing PhaseTwoB")
			if leader.commanderArg != nil {
				if proto.Equal(pTwo.BallotNum, leader.ballotNum) {
					delete(commanderWaitFor, pTwo.AcceptorId)
					if len(commanderWaitFor) < acceptors.Num()/2 {
						// Send decision to replicas
						leader.sendReplicasDecision(id, replicaClients, pb.Proposal{
							SlotIdx: leader.commanderArg[0].SlotIdx,
							Command: leader.commanderArg[0].Command})
						if len(leader.commanderArg) != 1 {
							leader.commanderArg = leader.commanderArg[1:]
							leader.sendAcceptorsPhaseTwoA(id, acceptorClients, *leader.commanderArg[0])
						} else {
							leader.commanderArg = nil
						}
					}
				} else {
					log.Printf("Stopping Commander thread")
					leader.preemptedChan <- pTwo.BallotNum
					if len(leader.commanderArg) != 1 {
						leader.commanderArg = leader.commanderArg[1:]
						leader.sendAcceptorsPhaseTwoA(id, acceptorClients, *leader.commanderArg[0])
					} else {
						leader.commanderArg = nil
					}
				}
			}
		}
	}
}
