package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/db"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/etcd"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/queue"
	"github.com/pkg/errors"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

type Server interface {
	CampaignForMasterCandidature(ctx context.Context) (chan *concurrency.Election, error)
	InitJobToSendServerUpdates(ctx context.Context) error
	InitJobForPullingTasks(ctx context.Context, tickerTimeInSeconds int, electionResultChan chan *concurrency.Election) error
	SubscribeToDoneQueue(ctx context.Context) (chan bool, error)
}

type server struct {
	identifier                     string `json:"identifier"`
	ipAddress                      string `json:"ipAddress"`
	nodePrefix                     string
	electionKey                    string
	detailsUpdateIntervalInSeconds int64
	etcdClient                     etcd.Etcd
	dbClient                       db.DB
	doneQueue                      queue.Queue
}

func (s *server) SubscribeToDoneQueue(ctx context.Context) (chan bool, error) {
	doneChan := make(chan bool)
	ch, err := s.doneQueue.Subscribe(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "Error in opening subscription to done channel")
	}
	go func() {
		for {
			select {
			case taskId, ok := <-ch:
				log.Println("Task id received as done task")
				if !ok {
					log.Println("done channel closed")
					return
				}
				uid, err := uuid.FromString(taskId)
				if err != nil {
					log.Println("Error while extracting uuid for task from string", err.Error())
					return
				}
				err = s.dbClient.UpdateTaskStatus(uid, db.DONE)
				if err != nil {
					log.Println("Error while updating task status", err.Error())
					return
				}
				log.Println(fmt.Sprintf("Task with id %s marked as done.", taskId))
			case <-ctx.Done():
				close(doneChan)
				log.Println("subscription to worker closed")
				return
			}
		}
	}()
	return doneChan, nil
}

func (s *server) CampaignForMasterCandidature(ctx context.Context) (chan *concurrency.Election, error) {
	electionChannel := make(chan *concurrency.Election)
	electionCandidate, err := s.etcdClient.NewElection(s.electionKey)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to initiate new election")
	}
	nodeId := s.identifier
	go func() {
		defer close(electionChannel)
		err := electionCandidate.Campaign(ctx, nodeId)
		if err != nil {
			log.Println(errors.Wrap(err, fmt.Sprintf("Campaign failed for node : %s", nodeId)))
			return
		} else {
			log.Println(fmt.Sprintf("Node %s elected as master", nodeId))
			electionChannel <- electionCandidate
		}
	}()
	return electionChannel, nil
}

func (s *server) InitJobToSendServerUpdates(ctx context.Context) error {
	key := s.nodePrefix + s.identifier
	value, err := json.Marshal(map[string]string{
		"identifier": s.identifier,
		"ipAddress":  s.ipAddress,
	})
	if err != nil {
		return errors.Wrap(err, "Server details marshalling failed")
	}
	log.Println(fmt.Sprintf("Updating server details to etcd: key = %s, value = %s", key, value))
	keepAliveChan, err := s.etcdClient.PutWithLease(ctx, key, string(value), s.detailsUpdateIntervalInSeconds)
	if err != nil {
		return errors.Wrap(err, "Failed to start server ping routine")
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				exit := <-keepAliveChan
				log.Println("Goroutine for sending server pings stopped. Exit: ", exit)
				return
			}
		}
	}()
	return nil
}

func (s *server) InitJobForPullingTasks(ctx context.Context, tickerTimeInSeconds int, electionResultChan chan *concurrency.Election) error {
	go func() {
		candidate := <-electionResultChan
		log.Println("Candidate won the election, starting ticker to pull jobs")
		candidateName := string((<-candidate.Observe(ctx)).Kvs[0].Value)
		log.Println(fmt.Sprintf("Candidate node %s starting cron job as master", candidateName))
		ticker := time.Tick(time.Duration(tickerTimeInSeconds) * time.Second)
		for {
			select {
			case <-ticker:
				current := time.Now().Unix()
				window := current - current%60
				log.Println("Pulling jobs for window : ", window)
				shards, err := s.dbClient.GetShards(window)
				if err != nil {
					log.Println("Error occurred while fetching jobs : ", err.Error())
					continue
				}
				servers, err := s.getActiveServers(ctx)
				if err != nil {
					log.Println("Error occurred while fetching active servers : ", err.Error())
					continue
				}
				err = s.sendShardsToSlaves(shards, servers)
				if err != nil {
					log.Println("Error occurred while sending shards to slaves : ", err.Error())
					continue
				}
			case <-ctx.Done():
				log.Println("Closing cron for pulling jobs")
				return
			}
		}
	}()
	return nil
}

func (s *server) getActiveServers(ctx context.Context) ([]map[string]string, error) {
	data, err := s.etcdClient.GetWithPrefix(ctx, s.nodePrefix)
	if err != nil {
		return nil, errors.Wrap(err, "Error in getting active servers from etcd")
	}
	var servers []map[string]string
	for _, v := range data {
		var server map[string]string
		err = json.Unmarshal([]byte(v), &server)
		if err != nil {
			return nil, errors.Wrap(err, "Error in unmarshalling server details")
		}
		servers = append(servers, server)
	}
	return servers, nil
}

func (s *server) sendShardsToSlaves(shards []db.Shard, servers []map[string]string) error {
	log.Println(fmt.Sprintf("sending shards = %v to slaves = %v", shards, servers))
	i := 0
	data := make(map[string][]uuid.UUID)
	for _, shard := range shards {
		server := servers[i]
		if _, ok := data[server["ipAddress"]]; !ok {
			data[server["ipAddress"]] = []uuid.UUID{}
		}
		data[server["ipAddress"]] = append(data[server["ipAddress"]], shard.Id)
		i = (i + 1) % len(servers)
	}
	for ipAddress, shards := range data {
		payload, _ := json.Marshal(map[string][]uuid.UUID{"shards": shards})
		req, err := http.NewRequest("PUT", "http://"+ipAddress+":8080/trigger-shards", bytes.NewReader(payload))
		if err != nil {
			return errors.Wrapf(err, "Error while making http request object")
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return errors.Wrapf(err, "Error in making request to slave")
		}
		if res.StatusCode == 200 {
			response, _ := ioutil.ReadAll(res.Body)
			log.Println(fmt.Sprintf("Success response for triggering shard for server %s : %s", ipAddress, string(response)))
		} else {
			response, _ := ioutil.ReadAll(res.Body)
			log.Println(fmt.Sprintf("Error response for triggering shard for server %s : %s", ipAddress, string(response)))
		}
	}
	return nil
}

func InitServer(identifier string, ipAddress string, nodePrefix string, electionKey string, etcdClient etcd.Etcd, dbClient db.DB, doneQueue queue.Queue) Server {
	return &server{
		identifier:  identifier,
		ipAddress:   ipAddress,
		etcdClient:  etcdClient,
		nodePrefix:  nodePrefix,
		electionKey: electionKey,
		dbClient:    dbClient,
		doneQueue:   doneQueue,
	}
}
