package backoff

import "time"

type Parameters struct {
	// BaseTimeout is the timeout to apply to the individual invocations of
	// the payload.
	BaseTimeout time.Duration

	// Multiplier is the factor that is applied to the timeout of each
	// consecutive retry.
	//
	// Values that are less than 1.0 are interpreted as 1.0 (no increase of
	// timeouts).
	Multiplier float64

	// MaximumTimeout is the cap at which the iterative increase of timeouts
	// stops.
	MaximumTimeout time.Duration

	// TotalTimeout is the overall period of time over which the retries will
	// be attempted.
	TotalTimeout time.Duration
}

var defaultParameters = Parameters{
	BaseTimeout:    5000 * time.Millisecond,
	Multiplier:     1.5,
	MaximumTimeout: 5 * time.Second,
	TotalTimeout:   30 * time.Second,
}

func DefaultParameters() *Parameters {
	res := defaultParameters // copy
	return &res
}

func WithDefaults(params *Parameters) *Parameters {
	if params == nil {
		res := defaultParameters // copy
		return &res
	}
	res := *params // copy
	if res.BaseTimeout == 0 {
		res.BaseTimeout = defaultParameters.BaseTimeout
	}
	if res.MaximumTimeout == 0 {
		res.MaximumTimeout = defaultParameters.MaximumTimeout
	}
	if res.TotalTimeout == 0 {
		res.TotalTimeout = defaultParameters.TotalTimeout
	}
	if res.Multiplier < 1.0 {
		res.Multiplier = defaultParameters.Multiplier
	}
	return &res
}
