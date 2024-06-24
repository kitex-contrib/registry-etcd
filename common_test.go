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

package etcd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_serviceKeyPrefix(t *testing.T) {
	assert.Equal(t,
		"kitex/registry-etcd/serviceName",
		serviceKeyPrefix("", "serviceName"),
	)

	assert.Equal(t,
		"tmp/serviceName",
		serviceKeyPrefix("tmp", "serviceName"),
	)
}

func Test_serviceKey(t *testing.T) {
	assert.Equal(t,
		"kitex/registry-etcd/serviceName/addr",
		serviceKey("", "serviceName", "addr"),
	)

	assert.Equal(t,
		"tmp/serviceName/addr",
		serviceKey("tmp", "serviceName", "addr"),
	)
}
