package main

import (
	"testing"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func TestNewReplica(t *testing.T) {
	r := NewReplica()
	if r.slotNum != 1 {
		t.Errorf("initial slotNum should be 1")
	}
}

func TestAddProposal(t *testing.T) {
	r := NewReplica()
	r.AddProposal(&pb.Proposal{SlotIdx: 3,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}})
	if len(r.proposals) != 1 {
		t.Errorf("should add to proposals")
	}
}

func TestAddDecision(t *testing.T) {
	r := NewReplica()
	r.AddDecision(&pb.Proposal{SlotIdx: 3,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}})
	if len(r.decisions) != 1 {
		t.Errorf("should add to decisions")
	}
}

func TestFindSlot(t *testing.T) {
	r := NewReplica()
	r.AddDecision(&pb.Proposal{SlotIdx: 3,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}})
	r.AddDecision(&pb.Proposal{SlotIdx: 1,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}})
	r.AddProposal(&pb.Proposal{SlotIdx: 2,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}})

	slot := r.FindSlot()
	if slot != 4 {
		t.Errorf("Min slot should be 4")
	}

	r.AddProposal(&pb.Proposal{SlotIdx: 4,
		Command: &pb.Command{Operation: pb.Op_SET, Arg: &pb.Command_Set{
			Set: &pb.KeyValue{Key: "Hello", Value: "World"}}}})

	slot = r.FindSlot()
	if slot != 5 {
		t.Errorf("Min slot should be 5")
	}
}

func TestCheckDecision(t *testing.T) {
	r := NewReplica()
	pro1 := &pb.Proposal{SlotIdx: 3,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}}
	r.AddDecision(pro1)
	r.AddDecision(&pb.Proposal{SlotIdx: 1,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}})
	r.AddProposal(&pb.Proposal{SlotIdx: 2,
		Command: &pb.Command{Operation: pb.Op_GET, Arg: &pb.Command_Get{
			Get: &pb.Key{Key: "Hello"}}}})

	r.slotNum = 5
	exists := r.CheckDecision(pro1.Command)
	if !exists {
		t.Errorf("should return true")
	}
	pro2 := &pb.Proposal{SlotIdx: 4,
		Command: &pb.Command{Operation: pb.Op_SET, Arg: &pb.Command_Set{
			Set: &pb.KeyValue{Key: "Hello", Value: "World"}}}}
	r.AddProposal(pro2)

	exists = r.CheckDecision(pro2.Command)
	if exists {
		t.Errorf("should be false")
	}
}
