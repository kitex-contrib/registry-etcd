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
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	"github.com/cloudwego/kitex/pkg/klog"
	clientv3 "go.etcd.io/etcd/client/v3"
)

// Option sets options such as username, tls etc.
type Option func(cfg *clientv3.Config)

// WithTLSOpt returns a option that authentication by tls/ssl.
func WithTLSOpt(certFile, keyFile, caFile string) Option {
	return func(cfg *clientv3.Config) {
		tlsCfg, err := newTLSConfig(certFile, keyFile, caFile, "")
		if err != nil {
			klog.Errorf("tls failed with err: %v , skipping tls.", err)
		}
		cfg.TLS = tlsCfg
	}
}

// WithAuthOpt returns a option that authentication by usernane and password.
func WithAuthOpt(username, password string) Option {
	return func(cfg *clientv3.Config) {
		cfg.Username = username
		cfg.Password = password
	}
}

// WithDialTimeoutOpt returns a option set dialTimeout
func WithDialTimeoutOpt(dialTimeout time.Duration) Option {
	return func(cfg *clientv3.Config) {
		cfg.DialTimeout = dialTimeout
	}
}

func newTLSConfig(certFile, keyFile, caFile, serverName string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	caCert, err := ioutil.ReadFile(caFile)
	if err != nil {
		return nil, err
	}
	caCertPool := x509.NewCertPool()
	successful := caCertPool.AppendCertsFromPEM(caCert)
	if !successful {
		return nil, errors.New("failed to parse ca certificate as PEM encoded content")
	}
	cfg := &tls.Config{
		Certificates: []tls.Certificate{cert},
		RootCAs:      caCertPool,
	}
	return cfg, nil
}
