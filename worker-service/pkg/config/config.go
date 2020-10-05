package config

import (
	"fmt"
	"github.com/spf13/viper"
	"log"
)

type configuration struct {
	rabbitMQUsername string
	rabbitMQPassword string
	rabbitMQHost     string
	rabbitMQPort     int
	workerQueueName  string
	doneQueueName    string
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
	ip, err := resolveHostIp()
	if err != nil {
		panic(fmt.Errorf("Fatal error in resolving ip address: %s \n", err.Error()))
	}
	log.Println("ip = ", ip)
	appConfiguration = &configuration{
		workerQueueName:  getStringOrPanic("worker_queue_name"),
		doneQueueName:    getStringOrPanic("done_queue_name"),
		rabbitMQHost:     getStringOrPanic("rabbitmq_host"),
		rabbitMQUsername: getStringOrPanic("rabbitmq_username"),
		rabbitMQPassword: getStringOrPanic("rabbitmq_password"),
		rabbitMQPort:     getIntOrPanic("rabbitmq_port"),
	}
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

func WorkerQueueName() string {
	return appConfiguration.workerQueueName
}

func DoneQueueName() string {
	return appConfiguration.doneQueueName
}
