package main

import (
	"context"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/server"
	"github.com/pkg/errors"
	"log"
)

func multiplexElectionChan(electionChannel chan *concurrency.Election, channelCount int) *[]chan *concurrency.Election {
	channels := make([]chan *concurrency.Election, channelCount)
	for i := 0; i < channelCount; i++ {
		channels[i] = make(chan *concurrency.Election)
	}
	go func() {
		defer func() {
			for i := 0; i < channelCount; i++ {
				close(channels[i])
			}
		}()
		value, ok := <-electionChannel
		if !ok {
			log.Println("Election channel closed")
			return
		}
		log.Println("Received value from election channel")
		for i := 0; i < channelCount; i++ {
			channels[i] <- value
		}
	}()
	return &channels
}

func initClusterOperations(ctx context.Context,server server.Server) (chan bool,error) {
	err := server.InitJobToSendServerUpdates(ctx)
	if err != nil {
		return nil,errors.Wrapf(err,"Failed to initialize job for sending node updates: ",err.Error())
	}
	electionChan, err := server.CampaignForMasterCandidature(ctx)
	if err != nil {
		return nil,errors.Wrap(err, "Campaign initialization for master failed")
	}
	channels := multiplexElectionChan(electionChan, 1)
	err = server.InitJobForPullingTasks(ctx, 60, (*channels)[0])
	if err != nil {
		return nil,errors.Wrap(err, "Initializing cron job for pulling jobs failed")
	}
	doneChan,err := server.SubscribeToDoneQueue(ctx)
	if err != nil {
		return nil,errors.Wrap(err,"Failed to subscribe to worker queue: ")
	}
	return doneChan,nil
}
