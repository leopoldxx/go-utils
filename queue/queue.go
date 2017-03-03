package queue

import (
	"fmt"
	"sync"
	"time"

	"github.com/leopoldxx/go-utils/trace"

	"context"
)

// const
const (
	HandleTimeout = 5 * time.Second
)

// HandlerWrap function for Handler interface
func HandlerWrap(f func(ctx context.Context, data []byte) error) *HandlerWrapper {
	return &HandlerWrapper{f}
}

// HandlerWrapper for Handler
type HandlerWrapper struct {
	Impl func(ctx context.Context, data []byte) error
}

// Handle of hw
func (hw *HandlerWrapper) Handle(ctx context.Context, data []byte) error {
	return hw.Impl(ctx, data)
}

// Handler of MsgQue
type Handler interface {
	Handle(ctx context.Context, data []byte) error
}

type msgBody struct {
	topic string
	body  []byte
}

// UnSub the handler
type UnSub func()

// MsgQueue struct
type MsgQueue struct {
	stopCtx    context.Context
	stopCancel context.CancelFunc
	data       chan msgBody
	mu         sync.Mutex
	handlers   map[string]map[Handler]UnSub
}

// NewMsgQueue creats new MsaQueue
func NewMsgQueue() *MsgQueue {
	ctx, cancel := context.WithCancel(context.TODO())
	msgque := &MsgQueue{
		stopCtx:    ctx,
		stopCancel: cancel,
		data:       make(chan msgBody, 10000),
		handlers:   map[string]map[Handler]UnSub{},
	}
	return msgque
}

// Handlers of all topics
func (mq *MsgQueue) Handlers() map[string]map[Handler]UnSub {
	return mq.handlers
}

// Stop the message queue
func (mq *MsgQueue) Stop() {
	mq.stopCancel()
}

// Run the background processor
func (mq *MsgQueue) Run() {
	handle := func(mq *MsgQueue, mb *msgBody) {
		var wg sync.WaitGroup
		mq.mu.Lock()
		if hs, ok := mq.handlers[mb.topic]; ok {
			ctx, cancel := context.WithTimeout(mq.stopCtx, HandleTimeout)
			for h := range hs {
				wg.Add(1)
				go func(h Handler, ctx context.Context, body []byte) {
					tracer := trace.GetTraceFromContext(ctx)
					defer wg.Done()

					defer func() {
						if r := recover(); r != nil {
							tracer.Errorf("panic: %v\n", r)
						}
					}()

					tmpCh := make(chan error, 1)
					defer close(tmpCh)
					select {
					case tmpCh <- h.Handle(ctx, body):
					case <-ctx.Done():
					}
				}(h, ctx, mb.body)
			}
			mq.mu.Unlock()
			wg.Wait()
			cancel()
		} else {
			mq.mu.Unlock()
		}
	}

	for {
		select {
		case <-mq.stopCtx.Done():
			close(mq.data)
			return
		case msg := <-mq.data:
			handle(mq, &msg)
		}
	}
}

// Publisher xxx
type Publisher interface {
	Pub(topic string, data []byte) error
}

// Subscriber xxx
type Subscriber interface {
	Sub(topic string, handler Handler) (UnSub, error)
}

// Pub data into the topic
func (mq *MsgQueue) Pub(topic string, data []byte) error {
	select {
	case <-mq.stopCtx.Done():
		return fmt.Errorf("the message queue has beed closed")
	default:
		mq.data <- msgBody{topic: topic, body: data}
		return nil
	}
}

// Sub topic with specific handler
func (mq *MsgQueue) Sub(topic string, handler Handler) (UnSub, error) {
	select {
	case <-mq.stopCtx.Done():
		return nil, fmt.Errorf("the message queue has beed closed")
	default:
		mq.mu.Lock()
		defer mq.mu.Unlock()
		if _, ok := mq.handlers[topic]; !ok {
			mq.handlers[topic] = make(map[Handler]UnSub)
		}
		mq.handlers[topic][handler] = func() {
			mq.mu.Lock()
			delete(mq.handlers[topic], handler)
			if len(mq.handlers[topic]) == 0 {
				delete(mq.handlers, topic)
			}
			mq.mu.Unlock()
		}
		return mq.handlers[topic][handler], nil
	}
}
