package cache

import "time"

const (
	DefaultMaxLen = 100000
)

type Key interface{}
type Value interface{}

type Cache interface {
	Put(key Key, value Value)
	PutWithTimeout(key Key, value Value, t time.Duration)
	Get(key Key) (Value, bool)
	Del(key Key) Value
	Len() int
	Close()
}
