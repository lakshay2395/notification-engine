package queue

import (
	"context"
	"fmt"
	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"log"
)

var client *amqp.Connection

type Queue interface {
	Publish(message string) error
	Subscribe(ctx context.Context) (<-chan string, error)
}

type queue struct {
	name   string
}

func (q *queue) Publish(message string) error {
	ch, err := client.Channel()
	if err != nil {
		return errors.Wrapf(err, "Unable to open channel on rabbitmq connection")
	}
	defer ch.Close()
	return ch.Publish(
		"",     // exchange
		q.name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
}

func (q *queue) Subscribe(ctx context.Context) (<-chan string, error) {
	ch, err := client.Channel()
	if err != nil {
		return nil,errors.Wrapf(err, "Unable to open channel on rabbitmq connection")
	}
	msgs, err := ch.Consume(
		q.name, // queue
		"",     // consumer
		true,   // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	if err != nil {
		return nil,errors.Wrapf(err,"Unable to start subscription to rabbitmq queue")
	}
	channel := make(chan string)
	go func() {
		for {
			select {
			case delivery,ok := <- msgs:
				if !ok {
					log.Println("Subscription closed")
					close(channel)
				}
				channel <- string(delivery.Body)
			case <-ctx.Done():
				close(channel)
			}
		}
	}()
	return channel,nil
}

func initRabbitMqClient(username string,password string,host string,port int) error {
	if client == nil {
		conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s:%d/",username,password,host,port))
		if err != nil {
			return errors.Wrapf(err,"Failed to initialize connection to rabbitmq")
		}
		client = conn
	}
	return nil
}

func InitQueue(username,password,host string,port int, queueName string) (Queue, error) {
	err := initRabbitMqClient(username,password,host,port)
	if err != nil {
		return nil,errors.Wrapf(err,"Failed to initialize rabbitmq connection singleton")
	}
	ch, err := client.Channel()
	if err != nil {
		return nil, errors.Wrapf(err, "Unable to open channel on rabbitmq connection")
	}
	_, err = ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	if err != nil {
		return nil, errors.Wrapf(err, fmt.Sprintf("Unable to declare queue `%s` on rabbitmq channel", queueName))
	}
	return &queue{
		name:   queueName,
	}, nil
}
