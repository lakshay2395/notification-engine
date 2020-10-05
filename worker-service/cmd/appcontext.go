package main

import (
	"context"
	"github.com/lakshay2395/notification-engine/worker-service/pkg/config"
	"github.com/lakshay2395/notification-engine/worker-service/pkg/queue"
	"github.com/lakshay2395/notification-engine/worker-service/pkg/server"
	"log"
)

func Init(ctx context.Context) {
	config.Load()
	workerQueue,err := queue.InitQueue(config.RabbitMQUsername(),config.RabbitMQPassword(),config.RabbitMQHost(),config.RabbitMQPort(),config.WorkerQueueName())
	if err != nil {
		log.Fatal("Failed to init worker queue on rabbitmq: ", err.Error())
	}
	log.Println("Worker queue initialized successfully")
	doneQueue,err := queue.InitQueue(config.RabbitMQUsername(),config.RabbitMQPassword(),config.RabbitMQHost(),config.RabbitMQPort(),config.DoneQueueName())
	if err != nil {
		log.Fatal("Failed to init done queue on rabbitmq: ", err.Error())
	}
	log.Println("Done queue initialized successfully")
	server := server.InitServer(workerQueue,doneQueue)
	doneChan,err := server.SubscribeToWorkerQueue(ctx)
	if err != nil {
		log.Fatal("Failed to subscribe to worker queue: ", err.Error())
	}
	<-doneChan
}
