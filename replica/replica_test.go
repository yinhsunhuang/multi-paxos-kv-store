package main

import (
	"testing"

	"github.com/golang/protobuf/proto"
	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func TestPbEqual(t *testing.T) {
	pb1 := &pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: "sdfqer",
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}}
	pb2 := &pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: "sdfqer",
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}}
	if !proto.Equal(pb1, pb2) {
		t.Errorf("pb1, pb2 not equal")
	}

	if pb1 == pb2 {
		t.Errorf("pb1, pb2 should not be literally equal")
	}

}

func TestNewReplica(t *testing.T) {
	r := NewReplica()
	if r.slotNum != 1 {
		t.Errorf("initial slotNum should be 1")
	}
}

func TestAddProposal(t *testing.T) {
	r := NewReplica()
	r.AddProposal(&pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: "sdfqer",
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	if len(r.proposals) != 1 {
		t.Errorf("should add to proposals")
	}
}

func TestAddDecision(t *testing.T) {
	r := NewReplica()
	r.AddDecision(&pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: "sdfqer",
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	if len(r.decisions) != 1 {
		t.Errorf("should add to decisions")
	}
}

func TestFindSlot(t *testing.T) {
	r := NewReplica()
	clientId := "ffsdf"
	r.AddDecision(&pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	r.AddDecision(&pb.Proposal{SlotIdx: 1,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 2,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	r.AddProposal(&pb.Proposal{SlotIdx: 2,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 3,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})

	slot := r.FindSlot()
	if slot != 4 {
		t.Errorf("Min slot should be 4")
	}

	r.AddProposal(&pb.Proposal{SlotIdx: 4,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 4,
			KvOp: &pb.Command{Operation: pb.Op_SET, Arg: &pb.Command_Set{
				Set: &pb.KeyValue{Key: "Hello", Value: "World"}}}}})

	slot = r.FindSlot()
	if slot != 5 {
		t.Errorf("Min slot should be 5")
	}
}

func TestCheckDecision(t *testing.T) {
	r := NewReplica()
	clientId := "j32kc"
	pro1 := &pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}}
	pro1_clone := &pb.Proposal{SlotIdx: 3,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 1,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}}
	r.AddDecision(pro1)
	r.AddDecision(&pb.Proposal{SlotIdx: 1,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 2,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})
	r.AddProposal(&pb.Proposal{SlotIdx: 2,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 3,
			KvOp: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
				Get: &pb.Key{Key: "Hello"}}}}})

	r.slotNum = 5
	exists := r.CheckDecision(pro1_clone.Command)
	if !exists {
		t.Errorf("should return true")
	}
	pro2 := &pb.Proposal{SlotIdx: 4,
		Command: &pb.PaxosCommand{ClientId: clientId,
			CommandId: 4,
			KvOp: &pb.Command{Operation: pb.Op_SET, Arg: &pb.Command_Set{
				Set: &pb.KeyValue{Key: "Hello", Value: "World"}}}}}
	r.AddProposal(pro2)

	exists = r.CheckDecision(pro2.Command)
	if exists {
		t.Errorf("should be false")
	}
}
