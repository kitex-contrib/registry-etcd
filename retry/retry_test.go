// Copyright 2021 CloudWeGo Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
