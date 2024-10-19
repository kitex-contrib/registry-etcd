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
	"net"

	"github.com/cloudwego-contrib/cwgo-pkg/registry/etcd/etcdkitex"
	"github.com/cloudwego/kitex/pkg/registry"
	"github.com/kitex-contrib/registry-etcd/retry"
)

// NewEtcdRegistry creates an etcd based registry.
func NewEtcdRegistry(endpoints []string, opts ...Option) (registry.Registry, error) {
	return etcdkitex.NewEtcdRegistry(endpoints, opts...)
}

// SetFixedAddress sets the fixed address for registering
// setting address to nil to clear the previous address
func SetFixedAddress(r registry.Registry, address net.Addr) {
	etcdkitex.SetFixedAddress(r, address)
}

// NewEtcdRegistryWithRetry creates an etcd based registry with given custom retry configs
func NewEtcdRegistryWithRetry(endpoints []string, retryConfig *retry.Config, opts ...Option) (registry.Registry, error) {
	return etcdkitex.NewEtcdRegistryWithRetry(endpoints, retryConfig, opts...)
}

// NewEtcdRegistryWithAuth creates an etcd based registry with given username and password.
// Deprecated: Use WithAuthOpt instead.
func NewEtcdRegistryWithAuth(endpoints []string, username, password string) (registry.Registry, error) {
	return etcdkitex.NewEtcdRegistryWithAuth(endpoints, username, password)
}
