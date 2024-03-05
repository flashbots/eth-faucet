package backoff

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/flashbots/eth-faucet/logutils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

var (
	ErrBackoffContextError                 = errors.New("context error")
	ErrBackoffDownstreamOperationCancelled = errors.New("downstream canceled the operation")
	ErrBackoffFunctionalError              = errors.New("functional error")
	ErrBackoffNonRetryableError            = errors.New("non-retryable error")
	ErrBackoffTotalTimeoutExpired          = errors.New("total timeout expired")
)

func Backoff(
	ctx context.Context,
	params *Parameters,
	payload func(ctx context.Context) error,
) error {
	params = WithDefaults(params)
	deadline := time.Now().Add(params.TotalTimeout)

	l := logutils.LoggerFromContext(ctx)
	if l != nil {
		l = l.With(zap.String(
			"operation_uuid", uuid.Must(uuid.NewUUID()).String(),
		))
		ctx = logutils.ContextWithLogger(ctx, l)
	}

	errs := make([]error, 0)
	attempt := 1
	timeout := params.BaseTimeout

	for time.Now().Before(deadline) {
		start := time.Now()
		if l != nil {
			l.Debug("Running backoff-wrapped operation...",
				zap.Duration("timeout", timeout),
				zap.Int("attempt", attempt),
			)
		}

		if timeout > params.MaximumTimeout {
			timeout = params.MaximumTimeout
		}
		_ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		err := payload(_ctx)
		ctxErr := _ctx.Err()

		switch {
		case err == nil && ctxErr == nil:
			// happy path
			return nil
		case err != nil:
			// error during execution
			errs = append(errs, err)

			// operation encountered non-retryable error
			if !IsRetryable(err) {
				slices.Reverse(errs)
				return fmt.Errorf("%w: attempt %d: error log:\n%w",
					ErrBackoffNonRetryableError, attempt,
					errors.Join(errs...),
				)
			}
		case ctxErr != nil:
			// no execution error, but something wrong with context
			errs = append(errs, ctxErr)

			// downstream canceled the operation
			if errors.Is(ctxErr, context.Canceled) {
				slices.Reverse(errs)
				return fmt.Errorf("%w: attempt %d: error log:\n%w",
					ErrBackoffDownstreamOperationCancelled, attempt,
					errors.Join(errs...),
				)
			}
		}

		time.Sleep(timeout - time.Since(start))

		timeout = time.Duration(params.Multiplier * float64(timeout))
		attempt++
	}

	slices.Reverse(errs)
	return fmt.Errorf("%w: error log:\n%w",
		ErrBackoffTotalTimeoutExpired,
		errors.Join(errs...),
	)
}
