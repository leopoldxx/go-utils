package trace

import (
	"context"
	"fmt"
	"log"
	"testing"
)

func TestHandleCrash(t *testing.T) {
	crash()
	t.Log("still ok")
	f()
}

func crash() {
	ctx := WithTraceForContext(context.TODO(), "crash-test")
	defer HandleCrash(func(r interface{}) {
		LogCrashStack(ctx, r)
	})

	func() {
		func() {
			panic("do it")
		}()
	}()
}

func f() {
	defer HandleCrash(func(r interface{}) {
		log.Printf("recover from: %v", r)
	})

	fmt.Println("Calling g.")
	g(0)
	fmt.Println("Returned normally from g.")
}

func g(i int) {
	if i > 3 {
		fmt.Println("Panicking!")
		panic(fmt.Sprintf("%v", i))
	}
	defer fmt.Println("Defer in g", i)
	fmt.Println("Printing in g", i)
	g(i + 1)
}
