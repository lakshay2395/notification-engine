# Notification Engine

Distributed & scalable generic notification engine implementation in Golang. This is inspired from bigben by Walmart labs.

## Pre-requisites
- Docker

## Setup

Start a sample 3-node manager and single-node worker system using following command - 

```
make start
```

## Basic Architecture

### Major Components
- PostgresDB as data store
- Etcd v3 for cluster synchronization
- Rabbitmq for job queues

### Services
- Manager Service(s) : Master-slave cluster of a manager to pick up scheduled job from data store and push to worker queue.   
- Worker Service(s) : Idempotent worker for picking jobs from worker queue and executing them.