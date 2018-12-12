package main

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func TestCheckProposals(t *testing.T) {
	leader := NewLeader("hxcvx")
	pb1 := &pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: "sdfqer",
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}}
	leader.AddProposal(pb1)
	pb2 := &pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: "sdfqer",
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}}
	if !proto.Equal(pb1, pb2) {
		t.Errorf("pb1, pb2 should be equal")
	}

	if leader.CheckProposals(pb2.Command) != true {
		t.Errorf("CheckProposals should be true")
	}

}

func TestPmax(t *testing.T) {
	arg := make([]*pb.Pvalue, 0, 0)
	for i := 0; i < 5; i++ {
		arg = append(arg, &pb.Pvalue{
			BallotIdx: &pb.BallotNum{BallotIdx: 0, LeaderId: "lead0"},
			SlotIdx:   int64(i + 1),
			Command: &pb.PaxosCommand{ClientId: "jjjjjj", CommandId: int64(i), KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	}

	for i := 0; i < 3; i++ {
		arg = append(arg, &pb.Pvalue{
			BallotIdx: &pb.BallotNum{BallotIdx: 1, LeaderId: "lead1"},
			SlotIdx:   int64(i + 1),
			Command: &pb.PaxosCommand{ClientId: "111cvcg", CommandId: int64(i), KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	}

	ret := Pmax(arg)
	if len(ret) != 5 {
		t.Errorf("Wrong return length")
	}
	for i, v := range ret {
		if v.SlotIdx <= 3 {
			if v.Command.ClientId != "111cvcg" {
				t.Errorf("Wrong Pmax %v, %v", i, v)
			}
		} else {
			if v.Command.ClientId != "jjjjjj" {
				t.Errorf("Wrong Pmax %v, %v", i, v)
			}
		}
	}
}

func TestOperation(t *testing.T) {
	arg := make([]*pb.Pvalue, 0, 0)
	for i := 0; i < 5; i++ {
		arg = append(arg, &pb.Pvalue{
			BallotIdx: &pb.BallotNum{BallotIdx: 0, LeaderId: "lead0"},
			SlotIdx:   int64(i + 1),
			Command: &pb.PaxosCommand{ClientId: "jjjjjj", CommandId: int64(i), KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	}

	for i := 0; i < 3; i++ {
		arg = append(arg, &pb.Pvalue{
			BallotIdx: &pb.BallotNum{BallotIdx: 1, LeaderId: "lead1"},
			SlotIdx:   int64(i + 1),
			Command: &pb.PaxosCommand{ClientId: "111cvcg", CommandId: int64(i), KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	}

	ret := Pmax(arg)
	if len(ret) != 5 {
		t.Errorf("Wrong return length")
	}

	proposals1 := []*pb.Proposal{&pb.Proposal{SlotIdx: 6, Command: &pb.PaxosCommand{ClientId: "111cvcg", CommandId: 6, KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
		Get: &pb.Key{Key: "World"}}}}},
		&pb.Proposal{SlotIdx: 2, Command: &pb.PaxosCommand{ClientId: "cxe324", CommandId: 8, KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "World"}}}}},
		&pb.Proposal{SlotIdx: 7, Command: &pb.PaxosCommand{ClientId: "cxe324", CommandId: 9, KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "World"}}}}}}

	combined := CombineOp(proposals1, ret)
	if len(combined) != 7 {
		t.Errorf("Wrong return length")
	}

	proposals2 := []*pb.Proposal{&pb.Proposal{SlotIdx: 1, Command: &pb.PaxosCommand{ClientId: "111cvcg", CommandId: 6, KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
		Get: &pb.Key{Key: "World"}}}}},
		&pb.Proposal{SlotIdx: 2, Command: &pb.PaxosCommand{ClientId: "cxe324", CommandId: 8, KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "World"}}}}},
		&pb.Proposal{SlotIdx: 3, Command: &pb.PaxosCommand{ClientId: "cxe324", CommandId: 9, KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "World"}}}}}}
	combined = CombineOp(proposals2, ret)
	if len(combined) != 5 {
		t.Errorf("Wrong return length")
	}
}
