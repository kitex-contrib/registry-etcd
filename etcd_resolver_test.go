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
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"testing"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	"github.com/cloudwego/kitex/pkg/utils"
	"github.com/stretchr/testify/require"
	"go.etcd.io/etcd/server/v3/embed"
)

const (
	serviceName = "registry-etcd-test"
)

func TestEtcdResolver(t *testing.T) {
	s, endpoint := setupEmbedEtcd(t)

	rg, err := NewEtcdRegistry([]string{endpoint})
	require.Nil(t, err)
	rs, err := NewEtcdResolver([]string{endpoint})
	require.Nil(t, err)

	info := registry.Info{
		ServiceName: serviceName,
		Addr:        utils.NewNetAddr("tcp", "127.0.0.1:8888"),
		Weight:      66,
		Tags: map[string]string{
			"hello": "world",
		},
	}

	// test register service
	{
		err = rg.Register(&info)
		require.Nil(t, err)
		desc := rs.Target(context.TODO(), rpcinfo.NewEndpointInfo(serviceName, "", nil, nil))
		result, err := rs.Resolve(context.TODO(), desc)
		require.Nil(t, err)
		expected := discovery.Result{
			Cacheable: true,
			CacheKey:  serviceName,
			Instances: []discovery.Instance{
				discovery.NewInstance(info.Addr.Network(), info.Addr.String(), info.Weight, info.Tags),
			},
		}
		require.Equal(t, expected, result)
	}

	// test deregister service
	{
		err = rg.Deregister(&info)
		require.Nil(t, err)
		desc := rs.Target(context.TODO(), rpcinfo.NewEndpointInfo(serviceName, "", nil, nil))
		_, err = rs.Resolve(context.TODO(), desc)
		require.NotNil(t, err)
	}

	teardownEmbedEtcd(s)
}

func TestEmptyEndpoints(t *testing.T) {
	_, err := NewEtcdResolver([]string{})
	require.NotNil(t, err)
}

func setupEmbedEtcd(t *testing.T) (*embed.Etcd, string) {
	endpoint := fmt.Sprintf("unix://localhost:%06d", os.Getpid())
	u, err := url.Parse(endpoint)
	require.Nil(t, err)
	dir, err := ioutil.TempDir("", "etcd_resolver_test")
	require.Nil(t, err)

	cfg := embed.NewConfig()
	cfg.LCUrls = []url.URL{*u}
	// disable etcd log
	cfg.LogLevel = "panic"
	cfg.Dir = dir

	s, err := embed.StartEtcd(cfg)
	require.Nil(t, err)

	<-s.Server.ReadyNotify()
	return s, endpoint
}

func teardownEmbedEtcd(s *embed.Etcd) {
	s.Close()
	_ = os.RemoveAll(s.Config().Dir)
}
