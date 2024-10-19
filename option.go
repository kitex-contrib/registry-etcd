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
	"github.com/cloudwego-contrib/cwgo-pkg/registry/etcd/etcdkitex"
	"time"
)

// Option sets options such as username, tls etc.
type Option = etcdkitex.Option

// WithTLSOpt returns a option that authentication by tls/ssl.
func WithTLSOpt(certFile, keyFile, caFile string) Option {
	return etcdkitex.WithTLSOpt(certFile, keyFile, caFile)
}

// WithAuthOpt returns a option that authentication by usernane and password.
func WithAuthOpt(username, password string) Option {
	return etcdkitex.WithAuthOpt(username, password)
}

// WithDialTimeoutOpt returns a option set dialTimeout
func WithDialTimeoutOpt(dialTimeout time.Duration) Option {
	return etcdkitex.WithDialTimeoutOpt(dialTimeout)
}

// WithEtcdServicePrefix returns an option that sets the Prefix field in the Config struct
func WithEtcdServicePrefix(prefix string) Option {
	return etcdkitex.WithEtcdServicePrefix(prefix)
}

// WithDefaultWeight returns an option that sets the DefaultWeight field in the Config struct
func WithDefaultWeight(defaultWeight int) Option {
	return etcdkitex.WithDefaultWeight(defaultWeight)
}
