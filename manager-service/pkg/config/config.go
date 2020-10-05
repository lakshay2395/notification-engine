package config

import (
	"fmt"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
	"log"
	"strings"
	"time"
)

type configuration struct {
	apiPort                            int
	nodePrefix                         string
	nodeIP                             string
	nodeDetailsUpdateIntervalInSeconds int
	nodeId                             string
	etcdClusterHosts                   string
	etcdClusterDialTimeoutInSeconds    int
	etcdClusterRequestTimeoutInSeconds int
	etcdClusterElectionKey             string
	dbDriver                           string
	dbHost                             string
	dbPort                             int
	dbName                             string
	dbUsername                         string
	dbPassword                         string
	dbConnectionTimeout                int
	dbMaxOpenConnections               int
	dbMaxIdleConnections               int
	dbMaxConnectionLifetime            int
	rabbitMQUsername                   string
	rabbitMQPassword                   string
	rabbitMQHost                       string
	rabbitMQPort                       int
	workerQueueName                    string
	doneQueueName                      string
	maxShardSize                       int
}

var appConfiguration *configuration

func LoadTestConfig() {
	load("application.test")
}

func Load() {
	load("application")
}

func load(configFileName string) {
	viper.AutomaticEnv()
	viper.SetConfigName(configFileName)
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("..")
	err := viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err.Error()))
	}
	uid := uuid.NewV4()
	ip, err := resolveHostIp()
	if err != nil {
		panic(fmt.Errorf("Fatal error in resolving ip address: %s \n", err.Error()))
	}
	log.Println("ip = ", ip)
	appConfiguration = &configuration{
		maxShardSize:                       getIntOrPanic("max_shard_size"),
		workerQueueName:                    getStringOrPanic("worker_queue_name"),
		doneQueueName:                      getStringOrPanic("done_queue_name"),
		apiPort:                            getIntOrPanic("api_port"),
		etcdClusterHosts:                   getStringOrPanic("etcd_cluster_hosts"),
		etcdClusterDialTimeoutInSeconds:    getIntOrPanic("etcd_cluster_dial_timeout_in_seconds"),
		etcdClusterRequestTimeoutInSeconds: getIntOrPanic("etcd_cluster_request_timeout_in_seconds"),
		etcdClusterElectionKey:             getStringOrPanic("etcd_cluster_election_key"),
		dbDriver:                           getStringOrPanic("db_driver"),
		dbHost:                             getStringOrPanic("db_host"),
		dbPort:                             getIntOrPanic("db_port"),
		dbName:                             getStringOrPanic("db_name"),
		dbUsername:                         getStringOrPanic("db_username"),
		dbPassword:                         getStringOrPanic("db_password"),
		dbConnectionTimeout:                getIntOrPanic("db_connection_timeout"),
		dbMaxOpenConnections:               getIntOrPanic("db_max_open_connections"),
		dbMaxIdleConnections:               getIntOrPanic("db_max_idle_connections"),
		dbMaxConnectionLifetime:            getIntOrPanic("db_max_connection_lifetime"),
		nodeId:                             uid.String(),
		nodeIP:                             ip,
		nodePrefix:                         getStringOrPanic("node_prefix"),
		nodeDetailsUpdateIntervalInSeconds: getIntOrPanic("node_details_update_interval_in_seconds"),
		rabbitMQHost:                       getStringOrPanic("rabbitmq_host"),
		rabbitMQUsername:                   getStringOrPanic("rabbitmq_username"),
		rabbitMQPassword:                   getStringOrPanic("rabbitmq_password"),
		rabbitMQPort:                       getIntOrPanic("rabbitmq_port"),
	}
}

func NodeId() string {
	return appConfiguration.nodeId
}

func NodeIp() string {
	return appConfiguration.nodeIP
}

func NodePrefix() string {
	return appConfiguration.nodePrefix
}

func NodeDetailsUpdateIntervalInSeconds() int64 {
	return int64(appConfiguration.nodeDetailsUpdateIntervalInSeconds)
}

func EtcdClusterHosts() []string {
	return strings.Split(appConfiguration.etcdClusterHosts, ",")
}

func EtcdClusterDialTimeoutInSeconds() time.Duration {
	return time.Duration(appConfiguration.etcdClusterDialTimeoutInSeconds) * time.Second
}

func EtcdClusterRequestTimeoutInSeconds() time.Duration {
	return time.Duration(appConfiguration.etcdClusterRequestTimeoutInSeconds) * time.Second
}

func EtcdClusterElectionKey() string {
	return appConfiguration.etcdClusterElectionKey
}

func DBDriver() string {
	return appConfiguration.dbDriver
}

func DBHost() string {
	return appConfiguration.dbHost
}

func DBPort() int {
	return appConfiguration.dbPort
}

func DBName() string {
	return appConfiguration.dbName
}

func DBUsername() string {
	return appConfiguration.dbUsername
}

func DBPassword() string {
	return appConfiguration.dbPassword
}

func DBConnectionTimeout() int {
	return appConfiguration.dbConnectionTimeout
}

func DBMaxOpenConnections() int {
	return appConfiguration.dbMaxOpenConnections
}

func DBMaxIdleConnections() int {
	return appConfiguration.dbMaxIdleConnections
}

func DBMaxConnectionLifetime() int {
	return appConfiguration.dbMaxConnectionLifetime
}

func RabbitMQHost() string {
	return appConfiguration.rabbitMQHost
}

func RabbitMQPort() int {
	return appConfiguration.rabbitMQPort
}

func RabbitMQUsername() string {
	return appConfiguration.rabbitMQUsername
}

func RabbitMQPassword() string {
	return appConfiguration.rabbitMQPassword
}

func APIPort() int {
	return appConfiguration.apiPort
}

func WorkerQueueName() string {
	return appConfiguration.workerQueueName
}

func DoneQueueName() string {
	return appConfiguration.doneQueueName
}

func MaxShardSize() int64 {
	return int64(appConfiguration.maxShardSize)
}
