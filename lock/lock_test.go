package lock

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/leopoldxx/go-utils/trace"
)

func newClientv3(t *testing.T) *clientv3.Client {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{"http://10.0.2.15:2389"},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		t.Fatal("create etcd v3 client failed:", err)
	}
	return cli
}
func TestLockCancel(t *testing.T) {
	locker := New(newClientv3(t))
	ctx, cancel := context.WithTimeout(context.TODO(), time.Millisecond)
	defer cancel()
	_, _, err := locker.Trylock(ctx, "/test/distributed/locker/key")
	if err != context.DeadlineExceeded {
		t.Fatalf("lock must failed by context.DeadlineExceeded: %s", err)
	}
}

func TestLock(t *testing.T) {
	locker := New(newClientv3(t))

	testCases := []struct {
		lockTime          []time.Duration
		sleepTime         time.Duration
		tryTimes          int
		expectFailedCount int
	}{
		{
			lockTime:          []time.Duration{0},
			sleepTime:         1,
			tryTimes:          1,
			expectFailedCount: 0,
		},
		{
			lockTime:          []time.Duration{0, 0},
			sleepTime:         1,
			tryTimes:          2,
			expectFailedCount: 0,
		},
		{
			lockTime:          []time.Duration{0, 0},
			sleepTime:         1100,
			tryTimes:          2,
			expectFailedCount: 1,
		},
		{
			lockTime:          []time.Duration{0, 0, 0},
			sleepTime:         2000,
			tryTimes:          3,
			expectFailedCount: 2,
		},
		{
			lockTime:          []time.Duration{1000, 1000, 1000},
			sleepTime:         100,
			tryTimes:          3,
			expectFailedCount: 0,
		},
		{
			lockTime:          []time.Duration{1000, 1000, 1000},
			sleepTime:         600,
			tryTimes:          3,
			expectFailedCount: 1,
		},
		{
			lockTime:          []time.Duration{100, 100, 100},
			sleepTime:         210,
			tryTimes:          3,
			expectFailedCount: 2,
		},
		{
			lockTime:          []time.Duration{100, 300, 300},
			sleepTime:         210,
			tryTimes:          3,
			expectFailedCount: 1,
		},
		{
			lockTime:          []time.Duration{1100, 2000, 3000},
			sleepTime:         5000,
			tryTimes:          3,
			expectFailedCount: 2,
		},
	}

	for i, test := range testCases {
		failedCountChan := make(chan int, 3)

		var wg sync.WaitGroup
		for idx := 0; idx < test.tryTimes; idx++ {
			wg.Add(1)
			go func(i, idx int) {
				time.Sleep(time.Millisecond * time.Duration(idx*10))
				t.Logf("test case:%d : %v, lock #%d, %v", i, test, idx, time.Now())
				defer wg.Done()

				ctx, cancel := context.WithCancel(context.TODO())
				defer cancel()

				unlock, ctx, err := locker.Trylock(ctx, fmt.Sprintf("/test/distributed/locker/key%d", i), WithTTL(test.lockTime[idx]*time.Millisecond))
				if err != nil {
					t.Logf("lock %d,%v failed: %s, %v", idx, test, err, time.Now())
					failedCountChan <- 1
				} else {
					t.Logf("lock %d,%v success, %v", idx, test, time.Now())
					time.Sleep(time.Millisecond * test.sleepTime)
					unlock()
					t.Logf("unlock %d,%v success, %v", idx, test, time.Now())
				}
			}(i, idx)
		}
		wg.Wait()
		close(failedCountChan)
		failedCount := 0
		for i := range failedCountChan {
			failedCount += i
		}
		if failedCount != test.expectFailedCount {
			t.Fatalf("expect %d, got %d", test.expectFailedCount, failedCount)
		}
	}
}

func TestSessionClosed(t *testing.T) {
	locker := New(newClientv3(t))
	locker2 := New(newClientv3(t))

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()
	tracer := trace.GetTraceFromContext(ctx)

	_, lctx, _ := locker.Trylock(ctx, "/test/distributed/locker/key", WithTTL(time.Second))

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_, lctx, err := locker2.Trylock(ctx, "/test/distributed/locker/key", WithTTL(time.Second*10))
		tracer.Info(err)
		if err != nil {
			tracer.Errorf("locker2 lock failed: %s", err)
		} else {
			tracer.Info("locker2 get the lock ownership")
		}
		_, lctx, err = locker2.Trylock(ctx, "/test/distributed/locker/key", WithTTL(time.Second))
		tracer.Info(err)
		if err == nil {
			tracer.Info("locker2 get the lock ownership")
		}
		select {
		case <-lctx.Done():
			tracer.Info("locker2 lose the lock ownership")
		}

	}()

	tracer.Info("locker1 get the lock ownership")

	select {
	case <-lctx.Done():
		tracer.Info("locker1 lose the lock ownership by session closed")
	}
	wg.Wait()
}
