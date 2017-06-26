package queue

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/leopoldxx/go-utils/concurrency"
	"github.com/leopoldxx/go-utils/trace"

	"context"
)

// const
const (
	HandleTimeout = 5 * time.Second
)

// HandlerWrap function for Handler interface
func HandlerWrap(name string, f func(ctx context.Context, data []byte) error) *HandlerWrapper {
	return &HandlerWrapper{name, f}
}

// HandlerWrapper for Handler
type HandlerWrapper struct {
	NameValue string
	Impl      func(ctx context.Context, data []byte) error
}

// Handle of hw
func (hw *HandlerWrapper) Handle(ctx context.Context, data []byte) error {
	if hw == nil {
		return errors.New("nil handler")
	}
	return hw.Impl(ctx, data)
}

// Name of the handler
func (hw *HandlerWrapper) Name() string {
	if hw == nil {
		return "<nil>"
	}
	return hw.NameValue
}

// Handler of MsgQue
type Handler interface {
	Name() string
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
	handleBarrier := concurrency.NewBarrier(100)
	handle := func(mq *MsgQueue, mb *msgBody) {
		mq.mu.Lock()
		hs, ok := mq.handlers[mb.topic]
		tmphs := make([]Handler, 0, len(hs))
		if ok {
			for h := range hs {
				tmphs = append(tmphs, h)
			}
		}
		mq.mu.Unlock()
		defer handleBarrier.Done()

		var wg sync.WaitGroup
		ctx, cancel := context.WithTimeout(mq.stopCtx, HandleTimeout)
		defer cancel()
		for h := range tmphs {
			wg.Add(1)
			go func(hd Handler, ctx context.Context, body []byte) {
				ctx = trace.WithTraceForContext(ctx, hd.Name())
				tracer := trace.GetTraceFromContext(ctx)
				defer func() {
					if r := recover(); r != nil {
						tracer.Errorf("handler %s panic: %v\n", hd.Name(), r)
					}
				}()
				defer wg.Done()

				tmpCh := make(chan error, 1)
				go func() {
					defer func() {
						if r := recover(); r != nil {
							tracer.Errorf("handler %s panic: %v\n", hd.Name(), r)
						}
					}()
					tmpCh <- hd.Handle(ctx, body)
				}()

				select {
				case <-tmpCh:
				case <-ctx.Done():
				}
			}(tmphs[h], ctx, mb.body)
		}
		wg.Wait()
	}

	for {
		select {
		case <-mq.stopCtx.Done():
			close(mq.data)
			return
		case msg := <-mq.data:
			handleBarrier.Advance()
			go handle(mq, &msg)
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
