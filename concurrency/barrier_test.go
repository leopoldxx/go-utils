package concurrency

import (
	"log"
	"sync"
	"time"
)

func ExampleBarrier() {
	b := NewBarrier(15)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	wg := &sync.WaitGroup{}
	for i := 0; i < 100; i++ {

		b.Advance()
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			defer b.Done()

			time.Sleep(time.Millisecond * 500)
			log.Printf("done %d", i)
		}(i)
	}
	wg.Done()

	// Output:
}
