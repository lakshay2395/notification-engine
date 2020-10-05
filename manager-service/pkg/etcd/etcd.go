package etcd

import (
	"context"
	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
	"github.com/pkg/errors"
	"time"
)

type Etcd interface {
	PutWithLease(ctx context.Context,key string,value string,ttl int64) (<-chan *clientv3.LeaseKeepAliveResponse,error)
	Get(ctx context.Context,key string) (string,error)
	GetWithPrefix(ctx context.Context,prefix string) (map[string]string,error)
	NewElection(electionKey string) (*concurrency.Election,error)
}

type etcd struct {
	connection *clientv3.Client
}

func (e *etcd) Get(ctx context.Context,key string) (string,error) {
	response, err := e.connection.Get(ctx, key)
	if err != nil {
		return "",errors.Wrap(err,"Error in getting prefixed value")
	}
	return string(response.Kvs[0].Value),nil
}

func (e *etcd) GetWithPrefix(ctx context.Context,prefix string) (map[string]string,error) {
	response, err := e.connection.Get(ctx, prefix, clientv3.WithPrefix(), clientv3.WithSort(clientv3.SortByKey, clientv3.SortDescend))
	if err != nil {
		return nil,errors.Wrap(err,"Error in getting prefixed value")
	}
	data := make(map[string]string)
	for _,kv := range response.Kvs {
		data[string(kv.Key)] = string(kv.Value)
	}
	return data,nil
}

func (e *etcd) PutWithLease(ctx context.Context,key string,value string,ttl int64) (<-chan *clientv3.LeaseKeepAliveResponse,error) {
	lease, err := e.connection.Grant(ctx, ttl)
	if err != nil {
		return nil,errors.Wrap(err,"Lease grant failed")
	}
	_, err = e.connection.Put(ctx, key, value, clientv3.WithLease(lease.ID))
	if err != nil {
		return nil,errors.Wrap(err,"Put request failed")
	}
	keepAliveChannel, err := e.connection.KeepAlive(ctx, lease.ID)
	if err != nil {
		return nil,errors.Wrap(err,"Keep-alive request failed")
	}
	return keepAliveChannel,nil
}

func (e *etcd) NewElection(electionKey string) (*concurrency.Election,error) {
	session, err := concurrency.NewSession(e.connection)
	if err != nil {
		return nil,err
	}
	election := concurrency.NewElection(session,electionKey)
	return election,nil
}

func InitEtcdClient(hosts []string,dialTimeoutInSeconds time.Duration) (Etcd,error) {
	client,err := clientv3.New(clientv3.Config{
		Endpoints:   hosts,
		DialTimeout: dialTimeoutInSeconds,
	})
	if err != nil{
		return nil,errors.Wrapf(err, "Unable to open connection to etcd cluster")
	}
	return &etcd{
		connection: client,
	},nil
}