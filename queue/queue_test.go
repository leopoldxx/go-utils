package queue_test

import (
	"sync"
	"testing"
	"time"

	"context"

	. "github.com/leopoldxx/go-utils/queue"
)

type fakeHandle struct {
	name  string
	t     *testing.T
	count int
	sync.Mutex
}

func (fh *fakeHandle) NeedWrapHandle(ctx context.Context, data []byte) error {
	fh.Lock()
	defer fh.Unlock()
	fh.count++
	fh.t.Logf("%s -> %s", fh.name, string(data))
	return nil
}

func (fh *fakeHandle) Handle(ctx context.Context, data []byte) error {
	fh.Lock()
	defer fh.Unlock()
	fh.count++
	fh.t.Logf("%s -> %s", fh.name, string(data))
	return nil
}

func TestSubPub(t *testing.T) {
	que := NewMsgQueue()

	go que.Run()

	fh1 := &fakeHandle{name: "fh1", t: t}
	fh2 := &fakeHandle{name: "fh2", t: t}

	unsub1, _ := que.Sub("topic-all", &HandlerWrapper{fh1.NeedWrapHandle})
	unsub2, _ := que.Sub("topic-all", HandlerWrap(fh2.NeedWrapHandle))
	unsub3, _ := que.Sub("topic-add", fh1)

	checkNum := func(n1, n2, n3 int) {
		if len(que.Handlers()) != n1 {
			t.Fatal("sub topic failed")
		}

		if len(que.Handlers()["topic-all"]) != n2 {
			t.Fatal("sub topic failed")
		}

		if len(que.Handlers()["topic-add"]) != n3 {
			t.Fatal("sub topic failed")
		}
	}

	checkCount := func(fh *fakeHandle, n int) {
		fh.Lock()
		defer fh.Unlock()
		if fh.count != n {
			t.Fatalf("consume data failed, expect %d, got %d", n, fh.count)
		}
	}

	que.Pub("topic-all", []byte("msg1"))
	time.Sleep(10 * time.Millisecond)
	checkCount(fh1, 1)
	checkCount(fh2, 1)

	que.Pub("topic-all", []byte("msg2"))
	time.Sleep(10 * time.Millisecond)
	checkCount(fh1, 2)
	checkCount(fh2, 2)

	que.Pub("topic-add", []byte("msg3"))
	time.Sleep(10 * time.Millisecond)
	checkCount(fh1, 3)
	checkCount(fh2, 2)

	que.Pub("topic-add", []byte("msg4"))
	time.Sleep(10 * time.Millisecond)
	checkCount(fh1, 4)
	checkCount(fh2, 2)

	que.Pub("topic-del", []byte("msg5"))
	time.Sleep(10 * time.Millisecond)
	checkCount(fh1, 4)
	checkCount(fh2, 2)

	que.Stop()
	que.Stop() // multi close
	// won't process from here

	que.Pub("topic-all", []byte("msg2"))
	time.Sleep(10 * time.Millisecond)
	checkCount(fh1, 4)
	checkCount(fh2, 2)

	que.Pub("topic-del", []byte("msg5"))
	time.Sleep(10 * time.Millisecond)
	checkCount(fh1, 4)
	checkCount(fh2, 2)

	checkNum(2, 2, 1)
	unsub3()
	checkNum(1, 2, 0)
	unsub1()
	checkNum(1, 1, 0)
	unsub2()
	checkNum(0, 0, 0)
}
