package main

import (
	"flag"
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"
)

func main() {
	// Argument parsing
	var r *rand.Rand
	var seed int64
	var replicas arrayPeers
	var leaders arrayPeers
	var acceptors arrayPeers
	var clientPort int
	var paxosPort int
	flag.Int64Var(&seed, "seed", -1,
		"Seed for random number generator, values less than 0 result in use of time")
	flag.IntVar(&clientPort, "port", 3000,
		"Port on which server should listen to client requests")
	flag.IntVar(&paxosPort, "paxos", 3001,
		"Port on which server should listen to Paxos requests")
	flag.Var(&replicas, "replica", "A replica for this process")
	flag.Var(&acceptors, "acceptor", "An acceptor for this process")
	flag.Var(&leaders, "leader", "An acceptor for this process")
	flag.Parse()

	// Initialize the random number generator
	if seed < 0 {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		r = rand.New(rand.NewSource(seed))
	}

	// Get hostname
	name, err := os.Hostname()
	if err != nil {
		// Without a host name we can't really get an ID, so die.
		log.Fatalf("Could not get hostname")
	}

	id := fmt.Sprintf("%s:%d", name, paxosPort)
	log.Printf("Starting peer with ID %s", id)

	serve(r, &replicas, &acceptors, &leaders, id, paxosPort)
}
