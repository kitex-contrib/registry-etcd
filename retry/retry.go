package retry

import "time"

type Config struct {
	// The maximum number of call attempt times, including the initial call
	MaxAttemptTimes uint

	// The delay time of observing etcd key
	ObserveDelay time.Duration

	// The retry delay time
	RetryDelay time.Duration
}

func (o *Config) Apply(opts []Option) {
	for _, op := range opts {
		op.F(o)
	}
}

func NewRetryConfig(opts ...Option) *Config {
	retryConfig := &Config{
		MaxAttemptTimes: 5,
		ObserveDelay:    30 * time.Second,
		RetryDelay:      10 * time.Second,
	}

	retryConfig.Apply(opts)

	return retryConfig
}
