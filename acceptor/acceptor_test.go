package main

import (
	"math/rand"
	"testing"

	"time"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
	context "golang.org/x/net/context"
)

func TestPhaseOneACall(t *testing.T) {
	acc := NewAcceptor()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	leaders := arrayPeers([]string{})
	go serve(acc, r, &leaders, "acc0", 3001)
	_, err := acc.PhaseOneA(context.Background(), &pb.PhaseOneArg{
		LeaderId:  "lead0",
		BallotNum: &pb.BallotNum{BallotIdx: 1, LeaderId: "hahaha"}})
	if err != nil {
		t.Errorf("Rpc fails")
	}
}

func TestLessThan(t *testing.T) {
	pb1 := &pb.BallotNum{
		BallotIdx: 3,
		LeaderId:  "hahaha"}
	pb2 := &pb.BallotNum{
		BallotIdx: 3,
		LeaderId:  "hahaha"}
	if BallotNumLessThan(pb1, pb2) {
		t.Errorf("Should be false")
	}
	if BallotNumLessThan(pb2, pb1) {
		t.Errorf("Should be false")
	}
	pb2 = &pb.BallotNum{
		BallotIdx: 2,
		LeaderId:  "hahaha"}

	if !BallotNumLessThan(pb2, pb1) {
		t.Errorf("Should be true")
	}
	if BallotNumLessThan(pb1, pb2) {
		t.Errorf("Should be false")
	}

	pb1 = &pb.BallotNum{
		BallotIdx: 2,
		LeaderId:  "haha"}

	if BallotNumLessThan(pb2, pb1) {
		t.Errorf("Should be false")
	}
	if !BallotNumLessThan(pb1, pb2) {
		t.Errorf("Should be true")
	}

	pb1 = &pb.BallotNum{
		BallotIdx: 2,
		LeaderId:  ""}

	if BallotNumLessThan(pb2, pb1) {
		t.Errorf("Should be false")
	}
	if !BallotNumLessThan(pb1, pb2) {
		t.Errorf("Should be true")
	}
}
