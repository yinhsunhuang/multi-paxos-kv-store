package main

import (
	"github.com/golang/protobuf/proto"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
	context "golang.org/x/net/context"
)

type Replica struct {
	slotNum   int64
	proposals []*pb.Proposal
	decisions []*pb.Proposal

	decisionChan chan DecisionInputType
}

type DecisionInputType struct {
	decision *pb.Proposal
}

func (r *Replica) FindSlot() int64 {
	var existsSlot map[int64]bool
	existsSlot = make(map[int64]bool)
	for _, v := range r.proposals {
		existsSlot[v.SlotIdx] = true
	}

	for _, v := range r.decisions {
		existsSlot[v.SlotIdx] = true
	}

	// find the smallest empty slot
	idx := int64(1)
	for {
		if _, ok := existsSlot[idx]; !ok {
			break
		}
		idx++
	}
	return idx
}

func (r *Replica) CheckDecision(cmd *pb.PaxosCommand) bool {
	for _, val := range r.decisions {
		if proto.Equal(val.Command, cmd) && val.SlotIdx < r.slotNum {
			return true
		}
	}
	return false
}

// CheckCurrentSlotNum returns true if \exists c s.t. <slotNum, c> \in decisions. Current Implementation is O(n)
func (r *Replica) CheckCurrentSlotNum() bool {
	for _, v := range r.decisions {
		if v.SlotIdx == r.slotNum {
			return true
		}
	}
	return false
}

func (r *Replica) AddProposal(prop *pb.Proposal) {
	r.proposals = append(r.proposals, prop)
}

func (r *Replica) AddDecision(prop *pb.Proposal) {
	r.decisions = append(r.decisions, prop)
}

// NewReplica returns a initialized Replica structure
func NewReplica() *Replica {
	ret := Replica{}
	ret.slotNum = 1
	ret.proposals = make([]*pb.Proposal, 0, 5)
	ret.decisions = make([]*pb.Proposal, 0, 5)

	ret.decisionChan = make(chan DecisionInputType)
	return &ret
}

func (r *Replica) Decision(ctx context.Context, in *pb.Proposal) (*pb.Empty, error) {
	r.decisionChan <- DecisionInputType{decision: in}
	return &pb.Empty{}, nil
}
