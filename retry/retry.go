package retry

import (
	"errors"
	"strings"
	"time"
)

type retriableError struct {
	err string
}

func (e *retriableError) Error() string {
	return e.err
}

// NewRetriableError create a retryable error
func NewRetriableError(err string) error {
	return &retriableError{err}
}

type merrs []error

func (e merrs) Err() error {
	if len(e) == 0 {
		return nil
	}
	if len(e) == 1 {
		return e[0]
	}
	errStr := make([]string, 0, len(e))
	for i := range e {
		errStr = append(errStr, e[i].Error())
	}
	return errors.New(strings.Join(errStr, ";"))
}

// Do will retry attempts time after callback failed, and wait for d duration between each callback
func Do(attempts int, callback func() error, d time.Duration) error {
	var errs merrs
	if attempts == -1 {
		attempts = ^int(0)
	}
	for i := 0; i < attempts; i++ {
		err := callback()
		if err == nil {
			return nil
		}
		errs = append(errs, err)
		if _, ok := err.(*retriableError); !ok {
			return errs.Err()
		}
		if int(d) > 0 {
			<-time.After(d)
		}
	}
	return errs.Err()
}
