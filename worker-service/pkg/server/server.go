package server

import (
	"context"
	"encoding/json"
	"github.com/lakshay2395/notification-engine/worker-service/pkg/queue"
	"github.com/pkg/errors"
	"log"
)

type Server interface {
	SubscribeToWorkerQueue(ctx context.Context) (chan bool,error)
	pushToDoneQueue(msg string) error
}

type server struct {
	workerQueue queue.Queue
	doneQueue   queue.Queue
}

func (s *server) SubscribeToWorkerQueue(ctx context.Context) (chan bool,error) {
	doneChan := make(chan bool)
	workerChan,err := s.workerQueue.Subscribe(ctx)
	if err != nil {
		return nil,errors.Wrap(err,"Error in opening subscription to worker channel")
	}
	go func() {
		for{
			select{
				case msg,ok := <- workerChan:
					log.Println("Message received from worker queue")
					if !ok {
						log.Println("worker channel closed")
						return
					}
					data := make(map[string]string)
					err = json.Unmarshal([]byte(msg),&data)
					if err != nil {
						log.Println("Error in unmarshalling message. Error: ",err.Error())
					}
					err = s.pushToDoneQueue(data["id"])
					if err != nil{
						log.Println("Error in pushing message to done queue. Error: ",err.Error())
					}
				case <-ctx.Done():
					close(doneChan)
					log.Println("subscription to worker closed")
					return
			}
		}
	}()
	return doneChan,nil
}


func (s *server) pushToDoneQueue(msg string) error {
	log.Println("Pushing msg to done queue")
	return s.doneQueue.Publish(msg)
}

func InitServer(workerQueue queue.Queue,doneQueue queue.Queue) Server{
	return &server{
		workerQueue: workerQueue,
		doneQueue: doneQueue,
	}
}