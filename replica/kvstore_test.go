package main

import (
	"testing"

	context "golang.org/x/net/context"

	"github.com/nyu-distributed-systems-fa18/multi-paxos/pb"
)

func TestGet(t *testing.T) {
	store := KVStore{C: make(chan InputChannelType), store: make(map[string]string)}
	go func() {
		for i := 0; i < 3; i++ {
			select {
			case op := <-store.C:
				store.HandleCommand(op)
			}
		}
	}()
	req := &pb.Key{Key: "Hello"}
	res, err := store.Get(context.Background(), req)
	if err != nil {
		t.Errorf("Test Failed")
	}
	v := res.GetKv()
	if v == nil {
		t.Errorf("Incorrect return type")
	}
	if v.Key != "Hello" || v.Value != "" {
		t.Errorf("return KV is wrong")
	}

	req2 := &pb.KeyValue{Key: "Hello", Value: "World"}
	store.Set(context.Background(), req2)

	req3 := &pb.Key{Key: "Hello"}
	res3, err3 := store.Get(context.Background(), req3)
	if err3 != nil {
		t.Errorf("Test Failed")
	}
	v2 := res3.GetKv()
	if v2 == nil {
		t.Errorf("Incorrect return type")
	}
	if v2.Key != "Hello" || v2.Value != "World" {
		t.Errorf("return KV is wrong")
	}
}
