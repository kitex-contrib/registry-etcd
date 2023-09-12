package retry

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRetryConfig(t *testing.T) {
	retryConfig := NewRetryConfig()
	assert.Equal(t, uint(5), retryConfig.MaxAttemptTimes)
	assert.Equal(t, 30*time.Second, retryConfig.ObserveDelay)
	assert.Equal(t, 10*time.Second, retryConfig.RetryDelay)
}

func TestRetryCustomConfig(t *testing.T) {
	retryConfig := NewRetryConfig(
		WithMaxAttemptTimes(10),
		WithObserveDelay(20*time.Second),
		WithRetryDelay(5*time.Second),
	)
	assert.Equal(t, uint(10), retryConfig.MaxAttemptTimes)
	assert.Equal(t, 20*time.Second, retryConfig.ObserveDelay)
	assert.Equal(t, 5*time.Second, retryConfig.RetryDelay)
}
