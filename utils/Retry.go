package utils

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/avast/retry-go"
)

type RetryFunc func(ctx context.Context) error

type retryableError struct {
	err error
}

// RetryableError marks an error as retryable.
func RetryableError(err error) error {
	if err == nil {
		return nil
	}
	return &retryableError{err}
}

// Unwrap implements error wrapping.
func (e *retryableError) Unwrap() error {
	return e.err
}

// Error returns the error string.
func (e *retryableError) Error() string {
	if e.err == nil {
		return "retryable: <nil>"
	}
	return "retryable: " + e.err.Error()
}

// DoRetry wraps a function with a backoff to retry. The provided context is the same
// context passed to the RetryFunc.
func DoRetry(ctx context.Context, f RetryFunc) error {
	var retryCount uint
	maxRetries := 3

	err := retry.Do(
		func() error {
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
			}
			err := f(ctx)
			if err != nil {
				retryCount++
			}
			return err
		},
		retry.Attempts(uint(maxRetries)),
		retry.Delay(5*time.Second),
		retry.OnRetry(func(n uint, err error) {
			fmt.Printf("Retrying request after error: %v\n", err)
		}),
	)

	// Tambahkan penanganan khusus jika mencapai batas percobaan maksimum
	if retryCount >= uint(maxRetries) {
		return errors.New("timeout: maximum retries exceeded")
	}

	return err
}

// func DoRetry(ctx context.Context, f RetryFunc) error {
// 	return retry.Do(
// 		func() error {
// 			select {
// 			case <-ctx.Done():
// 				return ctx.Err()
// 			default:
// 			}
// 			return f(ctx)
// 		},
// 		retry.Attempts(3),
// 		retry.Delay(5 * time.Second),
// 		retry.OnRetry(func(n uint, err error) {
// 			fmt.Printf("Retrying request after error: %v\n", err)
// 		}),
// 	)
// }
