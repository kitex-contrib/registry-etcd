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
