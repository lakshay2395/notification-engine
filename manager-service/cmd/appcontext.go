package main

import (
	"context"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/api"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/config"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/db"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/etcd"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/queue"
	"github.com/lakshay2395/notification-engine/manager-service/pkg/server"
	"log"
	"strconv"
)

func Init(ctx context.Context) {
	config.Load()
	etcdClient, err := etcd.InitEtcdClient(config.EtcdClusterHosts(), config.EtcdClusterDialTimeoutInSeconds())
	if err != nil {
		log.Fatal("Failed to initialize etcd client: ", err.Error())
	}
	log.Println("Connected to etcd successfully")
	dbClient, err := db.InitDB(config.DBHost(), config.DBPort(), config.DBUsername(), config.DBPassword(), config.DBName(),config.MaxShardSize(), config.DBConnectionTimeout(), config.DBMaxConnectionLifetime(), config.DBMaxIdleConnections(), config.DBMaxOpenConnections())
	if err != nil {
		log.Fatal("Failed to initialize db client: ", err.Error())
	}
	log.Println("Connected to db successfully")
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
	server := server.InitServer(config.NodeId(), config.NodeIp(), config.NodePrefix(), config.EtcdClusterElectionKey(), etcdClient, dbClient,doneQueue)
	_,err = initClusterOperations(ctx, server)
	if err != nil {
		log.Fatal("Failed to init cluster operations on server: ", err.Error())
	}
	log.Println("Cluster operations started successfully")
	api := api.Init(dbClient,workerQueue)
	log.Println("Starting API server...")
	err = api.InitRoutes(strconv.Itoa(config.APIPort()))
	if err != nil {
		log.Fatal("Failed to init API server: ", err.Error())
	}
}
