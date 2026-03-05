# TweetStorm ⚡🐦  
*A Distributed Twitter Simulation Inspired by Apache Storm*

## Overview

TweetStorm is a distributed systems simulation that mimics how real-time tweet streams could be processed in a cluster similar to Apache Storm.

Multiple clients send tweets to a distributed processing system where worker nodes analyze the tweets and update shared results. The system demonstrates core distributed systems algorithms such as logical clocks, leader election, mutual exclusion, and consensus.

This project was built as a Distributed Systems case study to illustrate how coordination and fault tolerance are handled in multi-node architectures.

---

## System Architecture

The system follows a **multi-client, multi-server architecture**.

**Clients → Leader → Worker Nodes → Shared Storage**

### Components

| Component | Role |
|-----------|------|
| Clients | Generate tweets/events |
| Leader | Coordinates the cluster |
| Workers | Process tweets |
| Shared Storage | Stores global results |

### Inspired by Apache Storm

| Our System | Storm Equivalent |
|------------|------------------|
| Client | Spout |
| Worker | Bolt |
| Leader | Nimbus |
| Coordination | ZooKeeper-like behavior |

---

## Example Tweet Processing

Clients generate tweets such as:

- "distributed systems are fun"  
- "ai is the future"  
- "storm processes tweets"

Workers process tweets and update global word counts stored in shared storage.

---

## Distributed Algorithms Implemented

This project demonstrates four core distributed system protocols.

### 1. Lamport Logical Clock

Maintains **event ordering across distributed nodes**.

Problem:  
Different machines may process events at different times, making it difficult to determine the correct order.

Solution:  
Logical timestamps ensure events are processed consistently across the distributed system.

---

### 2. Ricart–Agrawala Mutual Exclusion

Ensures **only one worker updates shared storage at a time**.

Problem:  
Multiple workers may attempt to update the same shared data simultaneously.

Solution:  
Workers request permission from other nodes before entering the critical section where shared data is updated.

---

### 3. Bully Leader Election

Elects a **new leader if the current coordinator fails**.

Problem:  
If the leader node crashes, the system must continue operating.

Solution:  
Nodes initiate an election process where the node with the highest ID becomes the new leader.

---

### 4. Consensus Algorithm

Ensures **all nodes agree on the final processed result**.

Problem:  
Different workers may compute slightly different values due to timing or ordering differences.

Solution:  
Workers propose values and the system selects the majority result, ensuring consistent system state.

---

## Workflow

1. Clients send tweets to the distributed system.
2. Lamport logical clocks assign timestamps to maintain event ordering.
3. The leader assigns tasks to worker nodes.
4. Workers process tweets and request mutual exclusion before updating shared storage.
5. Consensus ensures agreement on final results.
6. If the leader fails, the Bully algorithm elects a new leader.

---

## Project Structure

tweetstorm  
- client/  
- leader/  
- worker/  
- algorithms/  
- shared/  
- main

---

## Technologies Used

- Distributed Systems Concepts  
- Golang  
- Message Passing / Socket Communication  
- Apache Storm–inspired architecture  

---

## Learning Objectives

This project demonstrates:

- Distributed coordination mechanisms  
- Event ordering using logical clocks  
- Mutual exclusion in distributed environments  
- Leader election and fault tolerance  
- Consensus and distributed agreement  

---

## Demo Features

- Multiple clients sending tweets  
- Distributed tweet processing  
- Logical timestamp ordering  
- Safe shared storage updates using mutual exclusion  
- Agreement on final results using consensus  
- Leader failure simulation and election  

---

## Authors

Distributed Systems Case Study Project
