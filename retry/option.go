package retry

import "time"

// Option is the only struct that can be used to set Retry Config.
type Option struct {
	F func(o *Config)
}

// WithMaxAttemptTimes sets MaxAttemptTimes
func WithMaxAttemptTimes(maxAttemptTimes uint) Option {
	return Option{F: func(o *Config) {
		o.MaxAttemptTimes = maxAttemptTimes
	}}
}

// WithObserveDelay sets ObserveDelay
func WithObserveDelay(observeDelay time.Duration) Option {
	return Option{F: func(o *Config) {
		o.ObserveDelay = observeDelay
	}}
}

// WithRetryDelay sets RetryDelay
func WithRetryDelay(retryDelay time.Duration) Option {
	return Option{F: func(o *Config) {
		o.RetryDelay = retryDelay
	}}
}
