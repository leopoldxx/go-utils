package concurrency

// Barrier for limited go-routines
type Barrier struct {
	c chan struct{}
}

// NewBarrier creates new object and inits it
func NewBarrier(num int) *Barrier {
	b := &Barrier{}
	b.c = make(chan struct{}, num)
	for i := 0; i < num; i++ {
		b.c <- struct{}{}
	}
	return b
}

// Advance 1 step if there still is a unused go-routine
func (b *Barrier) Advance() {
	if b == nil {
		return
	}
	<-b.c
}

// Done means outside will release the go routine
func (b *Barrier) Done() {
	if b == nil {
		return
	}
	select {
	case b.c <- struct{}{}:
	default:
	}
}
