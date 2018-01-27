package retry_test

import (
	"errors"
	"log"
	"testing"

	"github.com/leopoldxx/go-utils/retry"
	"github.com/stretchr/testify/assert"
)

func TestRetry(t *testing.T) {
	err := retry.Do(3, func() error {
		return errors.New("mean it")
	}, 0)
	log.Println(err)

	err = retry.Do(3, func() error {
		return retry.NewRetriableError("mean it")
	}, 0)
	assert.NotNil(t, err)
	log.Println(err)

	errs := []error{
		retry.NewRetriableError("mean it"),
		retry.NewRetriableError("mean it"),
		nil,
	}
	assert.NotNil(t, err)
	idx := 0

	err = retry.Do(3, func() error {
		defer func() { idx++ }()
		return errs[idx]
	}, 0)
	assert.Nil(t, err)
	log.Println(err)

}
