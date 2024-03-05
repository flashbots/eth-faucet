package backoff

import (
	"errors"
	"fmt"
	"net"
)

type RetryableError struct {
	err error
}

func NewRetryableError(text string) error {
	return &RetryableError{
		err: errors.New(text),
	}
}

func RetryableErrorf(format string, a ...any) error {
	if len(a) == 0 {
		return &RetryableError{
			err: errors.New(format),
		}
	}
	return &RetryableError{
		err: fmt.Errorf(format, a...),
	}
}

func Retryable(err error) error {
	return &RetryableError{err: err}
}

func (e RetryableError) Error() string {
	return e.err.Error()
}

func (e RetryableError) Unwrap() error {
	return errors.Unwrap(e.err)
}

func IsRetryable(err error) bool {
	return iterateWrappedErrors(err, func(err error) bool {
		switch err.(type) {
		case *net.OpError, *RetryableError:
			return true
		default:
			return false
		}
	})
}

func iterateWrappedErrors(err error, iterate func(error) bool) bool {
	if stop := iterate(err); stop {
		return stop
	}

	switch x := err.(type) {
	case interface{ Unwrap() error }:
		err = x.Unwrap()
		if stop := iterateWrappedErrors(err, iterate); stop {
			return stop
		}
	case interface{ Unwrap() []error }:
		for _, err := range x.Unwrap() {
			if err != nil {
				if stop := iterateWrappedErrors(err, iterate); stop {
					return stop
				}
			}
		}
	}

	return false
}
