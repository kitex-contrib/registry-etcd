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
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"io/ioutil"
	"math/big"
	"net"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

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

func TestNewRegistryWithTLS(t *testing.T) {
	caFile, certFile, keyFile, err := setupCertificate()
	require.Nil(t, err)
	defer os.Remove(caFile)
	defer os.Remove(certFile)
	defer os.Remove(keyFile)

	s, endpoint := setupEmbedEtcdWithTLS(t, caFile, certFile, keyFile)

	opts := []Option{
		WithTLSOpt(certFile, keyFile, caFile),
	}

	rg, err := NewEtcdRegistry([]string{endpoint}, opts...)
	require.Nil(t, err)
	rs, err := NewEtcdResolver([]string{endpoint}, opts...)
	require.Nil(t, err)

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

func setupEmbedEtcdWithTLS(t *testing.T, caFile, certFile, keyFile string) (*embed.Etcd, string) {
	endpoint := fmt.Sprintf("unixs://localhost:%06d", os.Getpid())
	u, err := url.Parse(endpoint)
	require.Nil(t, err)
	dir, err := ioutil.TempDir("", "etcd_resolver_test")
	require.Nil(t, err)

	cfg := embed.NewConfig()

	cfg.ClientTLSInfo.CertFile = certFile
	cfg.ClientTLSInfo.KeyFile = keyFile
	cfg.ClientTLSInfo.TrustedCAFile = caFile

	require.Nil(t, err)
	cfg.LCUrls = []url.URL{*u}
	// disable etcd log
	cfg.LogLevel = "panic"
	cfg.Dir = dir

	s, err := embed.StartEtcd(cfg)
	require.Nil(t, err)

	<-s.Server.ReadyNotify()
	return s, endpoint
}

func setupCertificate() (caFile, certFile, keyFile string, err error) {
	tempDir := os.TempDir()
	caFile = filepath.Join(tempDir, "ca.pem")
	certFile = filepath.Join(tempDir, "server.pem")
	keyFile = filepath.Join(tempDir, "server-key.pem")

	// set up our CA certificate
	ca := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"Company, INC."},
			Country:      []string{"Beijing"},
			Province:     []string{"Beijing"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(10, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// create our private and public key
	caPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	// create the CA
	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	// pem encode
	caPEM := new(bytes.Buffer)
	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})

	// set up our server certificate
	cert := &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixNano()),
		Subject: pkix.Name{
			Organization: []string{"Company, INC."},
			Country:      []string{"Beijing"},
			Province:     []string{"Beijing"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		DNSNames:     []string{"localhost"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	certPrivKey, err := rsa.GenerateKey(rand.Reader, 4096)
	if err != nil {
		return
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, ca, &certPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return
	}

	certPEM := new(bytes.Buffer)
	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)
	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivKey),
	})

	// write to file
	if err = os.WriteFile(caFile, caPEM.Bytes(), 0o644); err != nil {
		return
	}
	if err = os.WriteFile(certFile, certPEM.Bytes(), 0o644); err != nil {
		return
	}
	if err = os.WriteFile(keyFile, certPrivKeyPEM.Bytes(), 0o644); err != nil {
		return
	}

	return
}

func teardownEmbedEtcd(s *embed.Etcd) {
	s.Close()
	_ = os.RemoveAll(s.Config().Dir)
}
