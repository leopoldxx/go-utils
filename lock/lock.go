package lock

import (
	"context"
	"errors"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/clientv3/concurrency"
)

const (
	defaultLockTTL = time.Second
)

// Unlocker is a func that can unlock the lock from the outside pkg
type Unlocker func()

// Locker is an interface that have a trylock method.
// It will lock on the specific key, different key will not be interfered.
type Locker interface {
	Trylock(ctx context.Context, key string, ops ...Options) (Unlocker, context.Context, error)
}

// Options config Locker
type Options func(opt *options)

// WithTTL configs the Locker with a timeout value, if the Locker cant
// get the lock when the ttl expiration, then Lock will fail. If the Locker
// get the lock in ttl, the ttl will be ignored.
func WithTTL(ttl time.Duration) Options {
	return func(opt *options) {
		opt.timeout = ttl
	}
}

// New will create a new Locker
func New(cli *clientv3.Client, opts ...Options) Locker {
	ops := &options{timeout: defaultLockTTL}
	for _, opt := range opts {
		opt(ops)
	}

	return &locker{
		etcdCli: cli,
		opts:    ops,
	}
}

type options struct {
	timeout time.Duration
}

type locker struct {
	etcdCli *clientv3.Client
	opts    *options
}

func (l *locker) Trylock(ctx context.Context, key string, ops ...Options) (Unlocker, context.Context, error) {
	if l == nil {
		return nil, nil, errors.New("nil locker")
	}
	opts := *l.opts
	for _, op := range ops {
		op(&opts)
	}
	if opts.timeout == 0 {
		opts.timeout = defaultLockTTL
	}

	type result struct {
		f   Unlocker
		ctx context.Context
		e   error
	}

	tmpCh := make(chan result, 1)
	cancelCtx, cancelFunc := context.WithCancel(context.TODO())

	timeoutCtx, timeoutCancel := context.WithTimeout(cancelCtx, opts.timeout)
	defer timeoutCancel()

	go func() {
		s, err := concurrency.NewSession(l.etcdCli, concurrency.WithContext(cancelCtx), concurrency.WithTTL(1 /* 1s */))
		if err != nil {
			cancelFunc()
			tmpCh <- result{nil, nil, err}
			return
		}

		mtx := concurrency.NewMutex(s, key)
		err = mtx.Lock(timeoutCtx)
		if err != nil {
			s.Close()
			cancelFunc()
			tmpCh <- result{nil, nil, err}
			return
		}
		select {
		case <-s.Done():
			mtx.Unlock(cancelCtx)
			tmpCh <- result{nil, nil, errors.New("session closed")}
			return
		default:
		}

		go func() {
			select {
			case <-s.Done():
			case <-cancelCtx.Done():
			}
			cancelFunc()
			newCtx, cancel := context.WithTimeout(context.TODO(), time.Second)
			defer cancel()
			mtx.Unlock(newCtx)
			s.Close()
		}()

		tmpCh <- result{
			func() {
				cancelFunc()
			},
			cancelCtx,
			nil,
		}
	}()

	select {
	case <-ctx.Done():
		cancelFunc()
		return nil, nil, ctx.Err()
	case <-timeoutCtx.Done():
		cancelFunc()
		return nil, nil, timeoutCtx.Err()
	case res := <-tmpCh:
		return res.f, res.ctx, res.e
	}
}
