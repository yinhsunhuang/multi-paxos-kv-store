package main

import (
	"log"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
	context "golang.org/x/net/context"
)

// Acceptor implements the AcceptorService grpc interface
type Acceptor struct {
	ballotNum *pb.BallotNum
	accepted  []*pb.Pvalue

	phaseOneChan chan *pb.PhaseOneArg
	phaseTwoChan chan *pb.PhaseTwoArg
}

// NewAcceptor TODO
func NewAcceptor() *Acceptor {
	ret := Acceptor{
		ballotNum:    &pb.BallotNum{BallotIdx: -1, LeaderId: ""},
		accepted:     make([]*pb.Pvalue, 0, 5),
		phaseOneChan: make(chan *pb.PhaseOneArg),
		phaseTwoChan: make(chan *pb.PhaseTwoArg)}
	return &ret
}

func (acc *Acceptor) AddAccepted(ac *pb.Pvalue) {
	acc.accepted = append(acc.accepted, ac)
}

func (acc *Acceptor) PhaseOneA(ctx context.Context, in *pb.PhaseOneArg) (*pb.Empty, error) {
	log.Printf("Receive PhaseOneA RPC call")
	acc.phaseOneChan <- in
	return &pb.Empty{}, nil
}

func (acc *Acceptor) PhaseTwoA(ctx context.Context, in *pb.PhaseTwoArg) (*pb.Empty, error) {
	log.Printf("Receive PhaseTwoA RPC call")
	acc.phaseTwoChan <- in
	return &pb.Empty{}, nil
}
