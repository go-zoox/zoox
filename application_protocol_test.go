package zoox

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/http2"
)

func mustTestCert(t *testing.T) (certPEM, keyPEM []byte) {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	require.NoError(t, err)

	serial := big.NewInt(1)
	template := x509.Certificate{
		SerialNumber: serial,
		Subject:      pkix.Name{CommonName: "localhost"},
		NotBefore:    time.Now().Add(-time.Hour),
		NotAfter:     time.Now().Add(24 * time.Hour),
		KeyUsage:     x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1)},
		DNSNames:     []string{"localhost"},
	}

	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	certPEM = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyBytes, err := x509.MarshalECPrivateKey(key)
	require.NoError(t, err)
	keyPEM = pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return certPEM, keyPEM
}

// Verifies the same stack as serveHTTPS: TLS + http2.ConfigureServer + Application handler negotiates HTTP/2 (ALPN h2).
func TestApplication_HTTP2OverTLS(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	tlsConf := &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
		NextProtos:   []string{"h2", "http/1.1"},
	}

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer ln.Close()

	app := New()
	app.Get("/probe", func(c *Context) { c.String(http.StatusOK, "ok") })

	srv := &http.Server{
		Handler:   app,
		TLSConfig: tlsConf,
	}
	require.NoError(t, http2.ConfigureServer(srv, &http2.Server{}))

	go func() { _ = srv.Serve(tls.NewListener(ln, tlsConf)) }()
	defer srv.Close()

	addr := ln.Addr().String()
	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
				NextProtos:         []string{"h2", "http/1.1"},
			},
			ForceAttemptHTTP2: true,
		},
		Timeout: 5 * time.Second,
	}

	var resp *http.Response
	for i := 0; i < 50; i++ {
		resp, err = client.Get(fmt.Sprintf("https://%s/probe", addr))
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "HTTP/2.0", resp.Proto)
}

func TestApplication_altSvcHeader(t *testing.T) {
	app := New()
	app.Config.EnableHTTP3 = true
	app.Config.HTTPSPort = 8443
	app.Config.HTTP3Port = 9443
	app.Config.HTTP3AltSvcMaxAge = 3600

	require.Equal(t, `h3=":9443"; ma=3600`, app.altSvcHeader())

	app.Config.HTTP3Port = 0
	require.Equal(t, `h3=":8443"; ma=3600`, app.altSvcHeader())

	app.Config.HTTP3AltSvcMaxAge = -1
	require.Equal(t, "", app.altSvcHeader())

	app.Config.HTTP3AltSvcMaxAge = 0
	app.Config.EnableHTTP3 = false
	require.Equal(t, "", app.altSvcHeader())
}

func TestApplication_Handler_HTTP3(t *testing.T) {
	certPEM, keyPEM := mustTestCert(t)
	cert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	tlsConf := &tls.Config{Certificates: []tls.Certificate{cert}}
	h3TLS := http3.ConfigureTLSConfig(tlsConf)

	pc, err := net.ListenPacket("udp", "127.0.0.1:0")
	require.NoError(t, err)
	defer pc.Close()

	app := New()
	app.Get("/h3", func(c *Context) { c.String(http.StatusOK, "ok") })

	srv := &http3.Server{
		TLSConfig: h3TLS,
		Handler:   app,
	}

	go func() { _ = srv.Serve(pc) }()
	defer func() { _ = srv.Close() }()

	port := pc.LocalAddr().(*net.UDPAddr).Port

	tr := &http3.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	defer tr.Close()

	client := &http.Client{
		Transport: tr,
		Timeout:   5 * time.Second,
	}

	var resp *http.Response
	for i := 0; i < 50; i++ {
		resp, err = client.Get(fmt.Sprintf("https://127.0.0.1:%d/h3", port))
		if err == nil {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	require.NoError(t, err)
	defer resp.Body.Close()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	require.Equal(t, "HTTP/3.0", resp.Proto)
}
