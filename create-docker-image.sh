#!/bin/sh
set -E
docker build -t local/paxos-peer -f Dockerfile .
