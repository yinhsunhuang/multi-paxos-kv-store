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
	var leaders arrayPeers
	var clientPort int
	var paxosPort int
	flag.Int64Var(&seed, "seed", -1,
		"Seed for random number generator, values less than 0 result in use of time")
	flag.IntVar(&clientPort, "port", 3000,
		"Port on which server should listen to client requests")
	flag.IntVar(&paxosPort, "paxos", 3001,
		"Port on which server should listen to Paxos requests")
	flag.Var(&leaders, "leader", "A peer for this process")
	flag.Parse()

	// Initialize the random number generator
	if seed < 0 {
		r = rand.New(rand.NewSource(time.Now().UnixNano()))
	} else {
		r = rand.New(rand.NewSource(seed))
	}
	acceptor := NewAcceptor()

	// Get hostname
	name, err := os.Hostname()
	if err != nil {
		// Without a host name we can't really get an ID, so die.
		log.Fatalf("Could not get hostname")
	}

	id := fmt.Sprintf("%s:%d", name, paxosPort)
	log.Printf("Starting acceptor with ID %s", id)

	serve(acceptor, r, &leaders, id, paxosPort)
}
