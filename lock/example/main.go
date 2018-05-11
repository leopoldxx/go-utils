package main

import (
	"context"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/leopoldxx/go-utils/lock"
	"github.com/leopoldxx/go-utils/trace"
)

func newClientv3() *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://10.0.2.15:2389"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Fatal("create etcd v3 client failed:", err)
	}
	return cli
}

func main() {
	locker := lock.New(newClientv3())
	ctx := trace.WithTraceForContext(context.TODO(), "test-main")
	tracer := trace.GetTraceFromContext(ctx)
	tracer.Info("begin test")

	unlock, ctx2, err := locker.Trylock(ctx, "/lock", lock.WithTTL(time.Second*10))
	if err != nil {
		tracer.Warnf("lock failed: %s", err)
		return
	}

	select {
	case <-ctx.Done():
	case <-time.After(10 * time.Second):
	}
	go func(ctx context.Context) {
		tracer := trace.GetTraceFromContext(ctx)
		tracer.Info("test context")
	}(ctx2)

	tracer.Info("safe")
	unlock()

	unlock, ctx3, err := locker.Trylock(ctx2, "/lock", lock.WithTTL(time.Minute*10))
	if err != nil {
		tracer.Infof("lock failed: %s", err)
		return
	}
	go func(ctx context.Context) {
		tracer := trace.GetTraceFromContext(ctx)
		tracer.Info("test context")
	}(ctx3)
	unlock()
}
