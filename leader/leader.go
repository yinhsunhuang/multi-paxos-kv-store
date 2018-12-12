package main

import (
	"context"
	"log"

	"github.com/golang/protobuf/proto"
	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

// Leader represent the leader role in paxos protocol
type Leader struct {
	ballotNum *pb.BallotNum
	active    bool
	proposals []*pb.Proposal

	proposeChan   chan *pb.Proposal
	adoptedChan   chan AdoptedInputType
	preemptedChan chan *pb.BallotNum

	phaseOneChan chan *pb.PhaseOneBArg
	phaseTwoChan chan *pb.PhaseTwoBArg

	scoutArg     *pb.BallotNum
	commanderArg []*pb.Pvalue
}

type AdoptedInputType struct {
	ballotNum *pb.BallotNum
	pvals     []*pb.Pvalue
}

// NewLeader TODO
func NewLeader(id string) *Leader {
	ret := Leader{ballotNum: &pb.BallotNum{BallotIdx: 0, LeaderId: id},
		proposeChan:   make(chan *pb.Proposal),
		active:        false,
		proposals:     make([]*pb.Proposal, 0, 5),
		adoptedChan:   make(chan AdoptedInputType, 1),
		preemptedChan: make(chan *pb.BallotNum, 1),
		phaseOneChan:  make(chan *pb.PhaseOneBArg),
		phaseTwoChan:  make(chan *pb.PhaseTwoBArg),
		scoutArg:      nil,
		commanderArg:  nil}
	return &ret
}

func Pmax(pvs []*pb.Pvalue) []*pb.Proposal {
	slotToPvalues := make(map[int64]*pb.Pvalue)
	for _, v := range pvs {
		val, ok := slotToPvalues[v.SlotIdx]
		if !ok {
			slotToPvalues[v.SlotIdx] = v
		} else {
			if BallotNumLessThan(val.BallotIdx, v.BallotIdx) {
				slotToPvalues[v.SlotIdx] = v
			}
		}
	}
	ret := make([]*pb.Proposal, 0, 0)
	for _, v := range slotToPvalues {
		ret = append(ret, &pb.Proposal{SlotIdx: v.SlotIdx, Command: v.Command})
	}
	return ret
}

func CombineOp(x []*pb.Proposal, y []*pb.Proposal) []*pb.Proposal {
	slotToCommand := make(map[int64]bool)
	ret := make([]*pb.Proposal, 0, 0)
	for _, v := range y {
		ret = append(ret, v)
		slotToCommand[v.SlotIdx] = true
	}
	for _, v := range x {
		if val, ok := slotToCommand[v.SlotIdx]; !ok || !val {
			ret = append(ret, v)
		}
	}
	return ret
}

// BallotNumLessThan is copyed from acceptor/ TODO should find a way to elimate copy-and-paste
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

func (l *Leader) AddProposal(prop *pb.Proposal) {
	l.proposals = append(l.proposals, prop)
}

func (l *Leader) CheckProposals(cmd *pb.PaxosCommand) bool {
	for _, v := range l.proposals {
		if proto.Equal(v.Command, cmd) {
			return true
		}
	}
	return false
}

// Propose TODO
func (l *Leader) Propose(ctx context.Context, in *pb.Proposal) (*pb.Empty, error) {
	log.Printf("Receive Propose RPC call")
	l.proposeChan <- in
	return &pb.Empty{}, nil
}

func (l *Leader) PhaseOneB(ctx context.Context, in *pb.PhaseOneBArg) (*pb.Empty, error) {
	log.Printf("Receive phaseOneB RPC call")
	l.phaseOneChan <- in
	log.Printf("Pushed to phaseOneChan")
	return &pb.Empty{}, nil
}

func (l *Leader) PhaseTwoB(ctx context.Context, in *pb.PhaseTwoBArg) (*pb.Empty, error) {
	log.Printf("Receive phaseTwoB RPC call")
	l.phaseTwoChan <- in
	return &pb.Empty{}, nil
}
