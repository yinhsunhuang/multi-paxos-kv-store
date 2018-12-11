package main

import (
	"context"
	"log"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

// Leader represent the leader role in paxos protocol
type Leader struct {
	proposeChan chan *pb.Proposal
}

// NewLeader TODO
func NewLeader() *Leader {
	ret := Leader{}
	return &ret
}

// Propose TODO
func (r *Leader) Propose(ctx context.Context, in *pb.Proposal) (*pb.Empty, error) {
	log.Printf("Receive Propose RPC call")
	r.proposeChan <- in
	return &pb.Empty{}, nil
}

func (r *Leader) phaseOneB(ctx context.Context, in *pb.PhaseOneBArg) (*pb.Empty, error) {
	log.Printf("Receive phaseOneB RPC call")
	return &pb.Empty{}, nil
}

func (r *Leader) phaseTwoB(ctx context.Context, in *pb.PhaseTwoBArg) (*pb.Empty, error) {
	log.Printf("Receive phaseTwoB RPC call")
	return &pb.Empty{}, nil
}
