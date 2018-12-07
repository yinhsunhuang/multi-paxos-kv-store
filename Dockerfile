FROM golang:1.11.0
WORKDIR /go/src/github.com/nyu-distributed-systems-fa18/multi-paxos/
COPY get.sh .
RUN sh get.sh

COPY pb pb

WORKDIR /go/src/github.com/nyu-distributed-systems-fa18/multi-paxos/replica
COPY replica .

RUN go get -v ./...
RUN go install -v ./...

WORKDIR /go/src/github.com/nyu-distributed-systems-fa18/multi-paxos/leader
COPY leader .

RUN go get -v ./...
RUN go install -v ./...

WORKDIR /go/src/github.com/nyu-distributed-systems-fa18/multi-paxos/acceptor
COPY acceptor .

RUN go get -v ./...
RUN go install -v ./...

EXPOSE 3000 3001
CMD [""]
