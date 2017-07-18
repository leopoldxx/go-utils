package main

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/leopoldxx/go-utils/lock"
)

func newClientv3() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://127.0.0.1:2389"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("create etcd v3 client failed:", err)
	}
	return cli
}

func main() {
	locker := lock.New(newClientv3())

	unlock, ctx, err := locker.Trylock(context.TODO(), "/lock", lock.WithTTL(time.Second*10))
	if err != nil {
		log.Printf("lock failed: %s", err)
	}
	log.Println("safe")

	select {
	case <-ctx.Done():
	case <-time.After(time.Minute):
	}
	log.Println("safe")
	unlock()

	unlock, ctx, err = locker.Trylock(context.TODO(), "/lock", lock.WithTTL(time.Minute*10))
	if err != nil {
		log.Printf("lock failed: %s", err)
	}
	log.Println("safe")
	unlock()
}
