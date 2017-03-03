package counter

import "sync/atomic"

type count struct {
	hit  int64
	miss int64
}

func (c *count) h(n int64) {
	atomic.AddInt64(&c.hit, n)
}

func (c *count) m(n int64) {
	atomic.AddInt64(&c.miss, n)
}

func (c *count) value() (int64, int64) {
	return atomic.LoadInt64(&c.hit), atomic.LoadInt64(&c.miss)
}

type Counter struct {
	counts []count
	total  count
	idx    int32
}

func New(group uint16) *Counter {
	return &Counter{
		counts: make([]count, group),
		total:  count{},
	}
}
func (counter *Counter) currentIdx() int32 {
	return atomic.LoadInt32(&counter.idx)
}
func (counter *Counter) Advance() (int32, bool) {
	idx := atomic.LoadInt32(&counter.idx)
	nextIdx := (idx + 1) % int32(len(counter.counts))
	if swapped := atomic.CompareAndSwapInt32(&counter.idx, idx, nextIdx); swapped {
		(&counter.total).h(
			-atomic.SwapInt64(
				&counter.counts[nextIdx].hit,
				0,
			),
		)
		(&counter.total).m(
			-atomic.SwapInt64(
				&counter.counts[nextIdx].miss,
				0,
			),
		)
		return nextIdx, true
	}
	return idx, false
}
func (counter *Counter) Hit() {
	idx := atomic.LoadInt32(&counter.idx)
	(&counter.counts[idx]).h(1)
	(&counter.total).h(1)
}
func (counter *Counter) Miss() {
	idx := atomic.LoadInt32(&counter.idx)
	(&counter.counts[idx]).m(1)
	(&counter.total).m(1)
}

func (counter *Counter) snapshot() ([]int64, []int64, int64, int64) {
	hit := make([]int64, len(counter.counts))
	miss := make([]int64, len(counter.counts))
	for idx := 0; idx < len(counter.counts); idx++ {
		hit[idx], miss[idx] = counter.counts[idx].value()
	}
	hitTotal, missTotal := counter.total.value()
	return hit, miss, hitTotal, missTotal
}
func (counter *Counter) Value() (int64, int64) {
	return counter.total.value()
}
