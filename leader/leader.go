package main

import (
	"context"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

// Leader represent the leader role in paxos protocol
type Leader struct {
}

// NewLeader TODO
func NewLeader() *Leader {
	ret := Leader{}
	return &ret
}

// Propose TODO
func (r *Leader) Propose(ctx context.Context, in *pb.Proposal) (*pb.Empty, error) {
	return &pb.Empty{}, nil
}
