package etcd

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"strings"
)

// newCertPool creates x509 certPool with provided CA files.
func newCertPool(CAFiles []string) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()

	for _, CAFile := range CAFiles {
		pemByte, err := os.ReadFile(CAFile)
		if err != nil {
			return nil, err
		}

		for {
			var block *pem.Block
			block, pemByte = pem.Decode(pemByte)
			if block == nil {
				break
			}
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return nil, err
			}

			certPool.AddCert(cert)
		}
	}

	return certPool, nil
}

// newCert generates TLS cert by using the given cert,key and parse function.
func newCert(certfile, keyfile string, parseFunc func([]byte, []byte) (tls.Certificate, error)) (*tls.Certificate, error) {
	cert, err := os.ReadFile(certfile)
	if err != nil {
		return nil, err
	}

	key, err := os.ReadFile(keyfile)
	if err != nil {
		return nil, err
	}

	if parseFunc == nil {
		parseFunc = tls.X509KeyPair
	}

	tlsCert, err := parseFunc(cert, key)
	if err != nil {
		return nil, err
	}
	return &tlsCert, nil
}

type TLSInfo struct {
	// CertFile is the _server_ cert, it will also be used as a _client_ certificate if ClientCertFile is empty
	CertFile string
	// KeyFile is the key for the CertFile
	KeyFile string
	// ClientCertFile is a _client_ cert for initiating connections when ClientCertAuth is defined. If ClientCertAuth
	// is true but this value is empty, the CertFile will be used instead.
	ClientCertFile string
	// ClientKeyFile is the key for the ClientCertFile
	ClientKeyFile string

	TrustedCAFile       string
	ClientCertAuth      bool
	CRLFile             string
	InsecureSkipVerify  bool
	SkipClientSANVerify bool

	// ServerName ensures the cert matches the given host in case of discovery / virtual hosting
	ServerName string

	// HandshakeFailure is optionally called when a connection fails to handshake. The
	// connection will be closed immediately afterwards.
	HandshakeFailure func(*tls.Conn, error)

	// CipherSuites is a list of supported cipher suites.
	// If empty, Go auto-populates it by default.
	// Note that cipher suites are prioritized in the given order.
	CipherSuites []uint16

	selfCert bool

	// parseFunc exists to simplify testing. Typically, parseFunc
	// should be left nil. In that case, tls.X509KeyPair will be used.
	parseFunc func([]byte, []byte) (tls.Certificate, error)

	// AllowedCN is a CN which must be provided by a client.
	AllowedCN string

	// AllowedHostname is an IP address or hostname that must match the TLS
	// certificate provided by a client.
	AllowedHostname string

	// EmptyCN indicates that the cert must have empty CN.
	// If true, ClientConfig() will return an error for a cert with non empty CN.
	EmptyCN bool
}

func (info TLSInfo) String() string {
	return fmt.Sprintf("cert = %s, key = %s, client-cert=%s, client-key=%s, trusted-ca = %s, client-cert-auth = %v, crl-file = %s", info.CertFile, info.KeyFile, info.ClientCertFile, info.ClientKeyFile, info.TrustedCAFile, info.ClientCertAuth, info.CRLFile)
}

func (info TLSInfo) Empty() bool {
	return info.CertFile == "" && info.KeyFile == ""
}

// baseConfig is called on initial TLS handshake start.
//
// Previously,
// 1. Server has non-empty (*tls.Config).Certificates on client hello
// 2. Server calls (*tls.Config).GetCertificate iff:
//    - Server's (*tls.Config).Certificates is not empty, or
//    - Client supplies SNI; non-empty (*tls.ClientHelloInfo).ServerName
//
// When (*tls.Config).Certificates is always populated on initial handshake,
// client is expected to provide a valid matching SNI to pass the TLS
// verification, thus trigger server (*tls.Config).GetCertificate to reload
// TLS assets. However, a cert whose SAN field does not include domain names
// but only IP addresses, has empty (*tls.ClientHelloInfo).ServerName, thus
// it was never able to trigger TLS reload on initial handshake; first
// ceritifcate object was being used, never being updated.
//
// Now, (*tls.Config).Certificates is created empty on initial TLS client
// handshake, in order to trigger (*tls.Config).GetCertificate and populate
// rest of the certificates on every new TLS connection, even when client
// SNI is empty (e.g. cert only includes IPs).
func (info TLSInfo) baseConfig() (*tls.Config, error) {
	if info.KeyFile == "" || info.CertFile == "" {
		return nil, fmt.Errorf("KeyFile and CertFile must both be present[key: %v, cert: %v]", info.KeyFile, info.CertFile)
	}

	_, err := newCert(info.CertFile, info.KeyFile, info.parseFunc)
	if err != nil {
		return nil, err
	}

	// Perform prevalidation of client cert and key if either are provided. This makes sure we crash before accepting any connections.
	if (info.ClientKeyFile == "") != (info.ClientCertFile == "") {
		return nil, fmt.Errorf("ClientKeyFile and ClientCertFile must both be present or both absent: key: %v, cert: %v]", info.ClientKeyFile, info.ClientCertFile)
	}
	if info.ClientCertFile != "" {
		_, err := newCert(info.ClientCertFile, info.ClientKeyFile, info.parseFunc)
		if err != nil {
			return nil, err
		}
	}

	cfg := &tls.Config{
		MinVersion: tls.VersionTLS12,
		ServerName: info.ServerName,
	}

	if len(info.CipherSuites) > 0 {
		cfg.CipherSuites = info.CipherSuites
	}

	// Client certificates may be verified by either an exact match on the CN,
	// or a more general check of the CN and SANs.
	var verifyCertificate func(*x509.Certificate) bool
	if info.AllowedCN != "" {
		if info.AllowedHostname != "" {
			return nil, fmt.Errorf("AllowedCN and AllowedHostname are mutually exclusive (cn=%q, hostname=%q)", info.AllowedCN, info.AllowedHostname)
		}
		verifyCertificate = func(cert *x509.Certificate) bool {
			return info.AllowedCN == cert.Subject.CommonName
		}
	}
	if info.AllowedHostname != "" {
		verifyCertificate = func(cert *x509.Certificate) bool {
			return cert.VerifyHostname(info.AllowedHostname) == nil
		}
	}
	if verifyCertificate != nil {
		cfg.VerifyPeerCertificate = func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			for _, chains := range verifiedChains {
				if len(chains) != 0 {
					if verifyCertificate(chains[0]) {
						return nil
					}
				}
			}
			return errors.New("client certificate authentication failed")
		}
	}

	// this only reloads certs when there's a client request
	// TODO: support server-side refresh (e.g. inotify, SIGHUP), caching
	cfg.GetCertificate = func(clientHello *tls.ClientHelloInfo) (cert *tls.Certificate, err error) {
		cert, err = newCert(info.CertFile, info.KeyFile, info.parseFunc)
		if os.IsNotExist(err) {
			return nil, err
		} else if err != nil {
			return nil, err
		}
		return cert, err
	}
	cfg.GetClientCertificate = func(unused *tls.CertificateRequestInfo) (cert *tls.Certificate, err error) {
		certfile, keyfile := info.CertFile, info.KeyFile
		if info.ClientCertFile != "" {
			certfile, keyfile = info.ClientCertFile, info.ClientKeyFile
		}
		cert, err = newCert(certfile, keyfile, info.parseFunc)
		if os.IsNotExist(err) {
			return nil, err
		} else if err != nil {
			return nil, err
		}
		return cert, err
	}
	return cfg, nil
}

// cafiles returns a list of CA file paths.
func (info TLSInfo) cafiles() []string {
	cs := make([]string, 0)
	if info.TrustedCAFile != "" {
		cs = append(cs, info.TrustedCAFile)
	}
	return cs
}

// ServerConfig generates a tls.Config object for use by an HTTP server.
func (info TLSInfo) ServerConfig() (*tls.Config, error) {
	cfg, err := info.baseConfig()
	if err != nil {
		return nil, err
	}

	cfg.ClientAuth = tls.NoClientCert
	if info.TrustedCAFile != "" || info.ClientCertAuth {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
	}

	cs := info.cafiles()
	if len(cs) > 0 {
		cp, err := newCertPool(cs)
		if err != nil {
			return nil, err
		}
		cfg.ClientCAs = cp
	}

	// "h2" NextProtos is necessary for enabling HTTP2 for go's HTTP server
	cfg.NextProtos = []string{"h2"}

	// go1.13 enables TLS 1.3 by default
	// and in TLS 1.3, cipher suites are not configurable
	// setting Max TLS version to TLS 1.2 for go 1.13
	cfg.MaxVersion = tls.VersionTLS12

	return cfg, nil
}

// ClientConfig generates a tls.Config object for use by an HTTP client.
func (info TLSInfo) ClientConfig() (*tls.Config, error) {
	var cfg *tls.Config
	var err error

	if !info.Empty() {
		cfg, err = info.baseConfig()
		if err != nil {
			return nil, err
		}
	} else {
		cfg = &tls.Config{ServerName: info.ServerName}
	}
	cfg.InsecureSkipVerify = info.InsecureSkipVerify

	cs := info.cafiles()
	if len(cs) > 0 {
		cfg.RootCAs, err = newCertPool(cs)
		if err != nil {
			return nil, err
		}
	}

	if info.selfCert {
		cfg.InsecureSkipVerify = true
	}

	if info.EmptyCN {
		hasNonEmptyCN := false
		cn := ""
		_, err := newCert(info.CertFile, info.KeyFile, func(certPEMBlock []byte, keyPEMBlock []byte) (tls.Certificate, error) {
			var block *pem.Block
			block, _ = pem.Decode(certPEMBlock)
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return tls.Certificate{}, err
			}
			if len(cert.Subject.CommonName) != 0 {
				hasNonEmptyCN = true
				cn = cert.Subject.CommonName
			}
			return tls.X509KeyPair(certPEMBlock, keyPEMBlock)
		})
		if err != nil {
			return nil, err
		}
		if hasNonEmptyCN {
			return nil, fmt.Errorf("cert has non empty Common Name (%s): %s", cn, info.CertFile)
		}
	}

	// go1.13 enables TLS 1.3 by default
	// and in TLS 1.3, cipher suites are not configurable
	// setting Max TLS version to TLS 1.2 for go 1.13
	cfg.MaxVersion = tls.VersionTLS12

	return cfg, nil
}

// IsClosedConnError returns true if the error is from closing listener, cmux.
// copied from golang.org/x/net/http2/http2.go
func IsClosedConnError(err error) bool {
	// 'use of closed network connection' (Go <=1.8)
	// 'use of closed file or network connection' (Go >1.8, internal/poll.ErrClosing)
	// 'mux: listener closed' (cmux.ErrListenerClosed)
	return err != nil && strings.Contains(err.Error(), "closed")
}
