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
			for err != nil {
				time.Sleep(10 * time.Millisecond)
				_, err = c.PhaseOneA(context.Background(), &pb.PhaseOneArg{
					LeaderId:  id,
					BallotNum: &b})
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
			for err != nil {
				time.Sleep(10 * time.Millisecond)
				_, err = c.PhaseTwoA(context.Background(), &pb.PhaseTwoArg{
					LeaderId: id,
					Pv:       &pv})
			}
		}(c, p)
	}
}

func (l *Leader) sendReplicasDecision(id string, replicaClients map[string]pb.ReplicaServiceClient, prop pb.Proposal) {
	for p, c := range replicaClients {
		go func(c pb.ReplicaServiceClient, p string) {
			log.Printf("Send Decision RPC to %v", p)
			_, err := c.Decision(context.Background(), &prop)
			for err != nil {
				time.Sleep(10 * time.Millisecond)
				_, err = c.Decision(context.Background(), &prop)
			}
		}(c, p)
	}
}

func randomDuration(r *rand.Rand) time.Duration {
	// Constant
	const DurationMax = 4000
	const DurationMin = 1000
	return time.Duration(r.Intn(DurationMax-DurationMin)+DurationMin) * time.Millisecond
}

// Restart the supplied timer using a random timeout based on function above
func restartTimer(timer *time.Timer, r *rand.Rand) {
	log.Printf("Restart Timer")
	stopped := timer.Stop()
	// If stopped is false that means someone stopped before us, which could be due to the timer going off before this,
	// in which case we just drain notifications.
	if !stopped {
		// Loop for any queued notifications
		for len(timer.C) > 0 {
			<-timer.C
		}
	}
	timer.Reset(randomDuration(r))
}

func serve(r *rand.Rand, replicas *arrayPeers, acceptors *arrayPeers, leaders *arrayPeers, id string, port int) {
	leader := NewLeader(id)
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

	leaderClients := make(map[string]pb.LeaderServiceClient)
	for _, peer := range *leaders {
		client, err := connectToLeader(peer)
		if err != nil {
			log.Fatalf("Failed to connect to GRPC server %v", err)
		}

		leaderClients[peer] = client
		log.Printf("Connected to %v", peer)
	}

	var scoutBallotNum *pb.BallotNum
	scoutPvalue := make([]*pb.Pvalue, 0, 0)
	scoutWaitFor := make(map[string]bool)

	commanderWaitFor := make(map[string]bool)
	timer := time.NewTimer(randomDuration(r))

	scoutTimer := time.NewTimer(randomDuration(r))

	foundLeaderChan := make(chan bool)

	// serve loop
	// Spawn scout
	leader.scoutArg = leader.ballotNum
	for k := range acceptorClients {
		scoutWaitFor[k] = true
	}
	leader.sendAcceptorsPhaseOneA(id, acceptorClients, *leader.ballotNum)
	log.Printf("Serve loop start")
	for {
		log.Printf("Waiting")
		select {
		case <-timer.C:
			log.Printf("Timeout")
			if !leader.active {
				log.Printf("Pinging")
				for p, c := range leaderClients {
					go func(c pb.LeaderServiceClient, p string) {
						log.Printf("Send Ping RPC to %v", p)
						ret, err := c.Ping(context.Background(), &pb.Empty{})
						for err != nil {
							time.Sleep(10 * time.Millisecond)
							ret, err = c.Ping(context.Background(), &pb.Empty{})
						}
						if ret.IsLeader {
							foundLeaderChan <- true
						}
					}(c, p)
				}
			}
			restartTimer(timer, r)
		case <-foundLeaderChan:
			log.Printf("Ping found active leader, resetting timer")
			restartTimer(scoutTimer, r)
		case <-scoutTimer.C:
			log.Printf("Timeout not finding active Leader")
			if !leader.active {
				if leader.scoutArg == nil {
					log.Printf("Starting Scout thread")
					leader.scoutArg = leader.ballotNum
					for k := range acceptorClients {
						scoutWaitFor[k] = true
					}
					leader.sendAcceptorsPhaseOneA(id, acceptorClients, *leader.ballotNum)
				}
			}
		case q := <-leader.pingChan:
			log.Printf("Pong")
			q.response <- leader.active
		case prop := <-leader.proposeChan:
			log.Printf("Processing  Propose Message: %v", prop)
			for _, v := range leader.proposals {
				if v.SlotIdx == prop.SlotIdx {
					log.Printf("Ignore proposal with same slotNum")
					break
				}
			}
			leader.AddProposal(prop)
			if leader.active {
				//Spawn Commander
				leader.commanderArg = []*pb.Pvalue{&pb.Pvalue{
					BallotIdx: leader.ballotNum,
					SlotIdx:   prop.SlotIdx,
					Command:   prop.Command}}
				for k := range acceptorClients {
					commanderWaitFor[k] = true
				}
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
			for k := range acceptorClients {
				commanderWaitFor[k] = true
			}
			if len(leader.commanderArg) > 0 {
				leader.sendAcceptorsPhaseTwoA(id, acceptorClients, *leader.commanderArg[0])
			}
			log.Printf("Leader Activated")
			leader.active = true
		case preempted := <-leader.preemptedChan:
			log.Printf("Processing Preempted Message %v", preempted)
			if BallotNumLessThan(leader.ballotNum, preempted) {
				log.Printf("Found greater ballot_num, return to passive mode")
				leader.active = false
				leader.ballotNum = &pb.BallotNum{BallotIdx: preempted.BallotIdx + 1,
					LeaderId: leader.ballotNum.LeaderId}
				log.Printf("Set BallotNum to %v", leader.ballotNum)
				// Spawn scout
				leader.scoutArg = leader.ballotNum
				for k := range acceptorClients {
					scoutWaitFor[k] = true
				}
				leader.sendAcceptorsPhaseOneA(id, acceptorClients, *leader.ballotNum)
			}
		case pOne := <-leader.phaseOneChan:
			log.Printf("Processing PhaseOneB %v", pOne)
			if leader.scoutArg != nil {
				// Scout is active
				if proto.Equal(pOne.BallotNum, leader.ballotNum) {
					for _, v := range pOne.Accepted {
						scoutPvalue = append(scoutPvalue, v)
					}
					log.Printf("Before deletion %v %v", scoutWaitFor, len(scoutWaitFor))
					delete(scoutWaitFor, pOne.AcceptorId)
					log.Printf("After deletion %v %v", scoutWaitFor, len(scoutWaitFor))
					if len(scoutWaitFor) < (acceptors.Num()+1)/2 {
						leader.adoptedChan <- AdoptedInputType{
							ballotNum: scoutBallotNum,
							pvals:     scoutPvalue}
						log.Printf("Adopt the ballotNum %v", leader.adoptedChan)

						leader.scoutArg = nil
					}
				} else {
					log.Printf("BallotNum not equal, sending preempted msg")
					leader.preemptedChan <- pOne.BallotNum
					leader.scoutArg = nil
				}
			} else {
				log.Printf("Scout is not active")
			}
		case pTwo := <-leader.phaseTwoChan:
			log.Printf("Processing PhaseTwoB %v", pTwo)
			if leader.commanderArg != nil {
				if proto.Equal(pTwo.BallotNum, leader.ballotNum) {
					log.Printf("Before deletion %v %v", commanderWaitFor, len(commanderWaitFor))
					delete(commanderWaitFor, pTwo.AcceptorId)
					log.Printf("After deletion %v %v", commanderWaitFor, len(commanderWaitFor))
					if len(commanderWaitFor) < (acceptors.Num()+1)/2 {
						// Send decision to replicas
						leader.sendReplicasDecision(id, replicaClients, pb.Proposal{
							SlotIdx: leader.commanderArg[0].SlotIdx,
							Command: leader.commanderArg[0].Command})
						if len(leader.commanderArg) != 1 {
							leader.commanderArg = leader.commanderArg[1:]
							for k := range acceptorClients {
								commanderWaitFor[k] = true
							}
							leader.sendAcceptorsPhaseTwoA(id, acceptorClients, *leader.commanderArg[0])
						} else {
							leader.commanderArg = nil
						}
					}
				} else {
					log.Printf("Stopping Commander thread")
					leader.preemptedChan <- pTwo.BallotNum
					if len(leader.commanderArg) > 1 {
						leader.commanderArg = leader.commanderArg[1:]
						for k := range acceptorClients {
							commanderWaitFor[k] = true
						}
						leader.sendAcceptorsPhaseTwoA(id, acceptorClients, *leader.commanderArg[0])
					} else {
						leader.commanderArg = nil
					}
				}
			} else {
				log.Printf("No commander is active")
			}
		}
	}
	log.Printf("Should not get to here")
}
