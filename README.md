# Multi-Creed Paxos Key Store
This project implements a Key-Store replicated state machine upon multi-creed paxos.

# Messages

## Kv Store, a.k.a. Command 
Defined in ```pb/kv.proto```. These are the command that a client can
issue to replicas.

## Replica
### Sending
* &lt;**propose**, s, c\&gt;
* &lt;**response**, cid, result&gt;
### Receiving
* &lt;**request**, c&gt;
* &lt;**decision**, s, c&gt;

## Acceptor
### Sending
* &lt;**p1a**, l, b&gt;
* &lt;**p2a**, l, &lt;b,s,c&gt;&gt;
### Receiving
* &lt;**p2b**, self(), *ballot_num*, *accepted*&gt;
* &lt;**p2b**, self(), *ballot_num*&gt;



# Data Structure

# Notes
To be consistent with lab2, each client request is treated as a unique one

# Program Logic
## Replica
Index response by Key as ClientId + CommandId