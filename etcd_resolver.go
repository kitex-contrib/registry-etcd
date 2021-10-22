// Copyright 2021 CloudWeGo authors.
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

// package etcd resolver
package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cloudwego/kitex/pkg/discovery"
	"github.com/cloudwego/kitex/pkg/klog"
	"github.com/cloudwego/kitex/pkg/rpcinfo"
	clientv3 "go.etcd.io/etcd/client/v3"
)

const (
	defaultWeight = 10
)

// etcdResolver is a resolver using etcd.
type etcdResolver struct {
	etcdClient *clientv3.Client
}

// NewEtcdResolver creates a etcd based resolver.
func NewEtcdResolver(endpoints []string) (discovery.Resolver, error) {
	etcdClient, err := clientv3.New(clientv3.Config{
		Endpoints: endpoints,
	})
	if err != nil {
		return nil, err
	}
	return &etcdResolver{
		etcdClient: etcdClient,
	}, nil
}

// Target implements the Resolver interface.
func (e *etcdResolver) Target(ctx context.Context, target rpcinfo.EndpointInfo) (description string) {
	return target.ServiceName()
}

// Resolve implements the Resolver interface.
func (e *etcdResolver) Resolve(ctx context.Context, desc string) (discovery.Result, error) {
	prefix := serviceKeyPrefix(desc)
	resp, err := e.etcdClient.Get(ctx, prefix, clientv3.WithPrefix())
	if err != nil {
		return discovery.Result{}, err
	}
	var (
		info serviceInfo
		eps  []discovery.Instance
	)
	for _, kv := range resp.Kvs {
		err := json.Unmarshal(kv.Value, &info)
		if err != nil {
			klog.Warnf("fail to unmarshal with err: %v, ignore key: %v", err, string(kv.Key))
			continue
		}
		weight := info.Weight
		if weight <= 0 {
			weight = defaultWeight
		}
		eps = append(eps, discovery.NewInstance(info.Network, info.Address, weight, info.Tags))
	}
	if len(eps) == 0 {
		return discovery.Result{}, fmt.Errorf("no instance remains for %v", desc)
	}
	return discovery.Result{
		Cacheable: true,
		CacheKey:  desc,
		Instances: eps,
	}, nil
}

// Diff implements the Resolver interface.
func (e *etcdResolver) Diff(cacheKey string, prev, next discovery.Result) (discovery.Change, bool) {
	return discovery.DefaultDiff(cacheKey, prev, next)
}

// Name implements the Resolver interface.
func (e *etcdResolver) Name() string {
	return "etcd"
}
