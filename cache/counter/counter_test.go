package counter

import (
	"testing"
	"time"
)

func TestCount(t *testing.T) {
	c := New(10)

	for i := 0; i < 50; i++ {
		go func(seed int) {
			for {
				timestampMod := time.Now().UnixNano() / 1000000 % 10
				if timestampMod < 1 {
					idx, advanced := c.Advance()
					if advanced == false {
						t.Logf("still: %d, %d", idx, c.currentIdx())
					}
					t.Log(idx)
					if idx > 10 {
						t.Fatal(idx)
					}
				} else if timestampMod < 6 {
					c.Hit()
				} else {
					c.Miss()
				}
			}
		}(i)
	}

	time.Sleep(time.Second * 10)
	hits, misses, hitT, missT := c.snapshot()
	hitSum, missSum := int64(0), int64(0)
	for i := 0; i < 10; i++ {
		hitSum += hits[i]
		missSum += misses[i]
	}
	t.Log(hitSum, missSum, hitT, missT)
}
