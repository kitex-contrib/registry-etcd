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

// Package etcd resolver
package etcd

import (
	"github.com/cloudwego-contrib/cwgo-pkg/registry/etcd/etcdkitex"

	"github.com/cloudwego/kitex/pkg/discovery"
)

const (
	defaultWeight = 10
)

// NewEtcdResolver creates a etcd based resolver.
func NewEtcdResolver(endpoints []string, opts ...Option) (discovery.Resolver, error) {
	return etcdkitex.NewEtcdResolver(endpoints, opts...)
}

// NewEtcdResolverWithAuth creates a etcd based resolver with given username and password.
// Deprecated: Use WithAuthOpt instead.
func NewEtcdResolverWithAuth(endpoints []string, username, password string) (discovery.Resolver, error) {
	return etcdkitex.NewEtcdResolverWithAuth(endpoints, username, password)
}
