package etcd

import (
	"crypto/tls"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type Option func(cfg *clientv3.Config)

func WithTlsOpt(certFile, keyFile, caFile string) Option {
	return func(cfg *clientv3.Config) {
		tlsCfg, err := newTLSConfig(certFile, keyFile, caFile, "")
		if err != nil {
			tlsCfg = nil
		}
		cfg.TLS = tlsCfg
	}
}

func WithAuthOpt(username, password string) Option {
	return func(cfg *clientv3.Config) {
		cfg.Username = username
		cfg.Password = password
	}
}

func newTLSConfig(certFile, keyFile, caFile, serverName string) (*tls.Config, error) {
	var (
		tlsCfg *tls.Config
		err    error
	)

	if certFile != "" || keyFile != "" || caFile != "" || serverName != "" {
		cfgtls := &TLSInfo{
			CertFile:      certFile,
			KeyFile:       keyFile,
			TrustedCAFile: caFile,
			ServerName:    serverName,
		}
		if tlsCfg, err = cfgtls.ClientConfig(); err != nil {
			return nil, err
		}
	}

	return tlsCfg, nil
}
