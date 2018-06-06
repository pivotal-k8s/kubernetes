/*
Copyright 2016 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package vclib_test

import (
	"context"
	"crypto/sha1"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"k8s.io/kubernetes/pkg/cloudprovider/providers/vsphere/vclib"
	"k8s.io/kubernetes/pkg/cloudprovider/providers/vsphere/vclib/fixtures"
)

func createTestServer(t *testing.T, caCertPath, serverCertPath, serverKeyPath string, handler http.HandlerFunc) (*httptest.Server, string) {
	caCertPEM, err := ioutil.ReadFile(caCertPath)
	if err != nil {
		t.Fatalf("Could not read ca cert from file")
	}

	serverCert, err := tls.LoadX509KeyPair(serverCertPath, serverKeyPath)
	if err != nil {
		t.Fatalf("Could not load server cert and server key from files: %#v", err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(caCertPEM); !ok {
		t.Fatalf("Cannot add CA to CAPool")
	}

	server := httptest.NewUnstartedServer(http.HandlerFunc(handler))
	server.TLS = &tls.Config{
		Certificates: []tls.Certificate{
			serverCert,
		},
		RootCAs: certPool,
	}

	// calculate the leaf certificate's fingerprint
	x509LeafCert := server.TLS.Certificates[0].Certificate[0]
	tpBytes := sha1.Sum(x509LeafCert)
	tpString := fmt.Sprintf("%x", tpBytes)

	return server, tpString
}

func TestWithValidCaCert(t *testing.T) {
	handler, verify := getRequestVerifier(t)

	server, _ := createTestServer(t, fixtures.CaCertPath, fixtures.ServerCertPath, fixtures.ServerKeyPath, handler)
	server.StartTLS()
	u := mustParseUrl(t, server.URL)

	connection := &vclib.VSphereConnection{
		Hostname: u.Hostname(),
		Port:     u.Port(),
		CACert:   fixtures.CaCertPath,
	}

	// Ignoring error here, because we only care about the TLS connection
	connection.NewClient(context.Background())

	verify()
}

func TestWithValidThumbprint(t *testing.T) {
	handler, verify := getRequestVerifier(t)

	server, serverThumbprint := createTestServer(t, fixtures.CaCertPath, fixtures.ServerCertPath, fixtures.ServerKeyPath, handler)
	server.StartTLS()
	u := mustParseUrl(t, server.URL)

	connection := &vclib.VSphereConnection{
		Hostname:   u.Hostname(),
		Port:       u.Port(),
		Thumbprint: serverThumbprint,
	}

	// Ignoring error here, because we only care about the TLS connection
	connection.NewClient(context.Background())

	verify()
}

func TestWithInvalidCaCertPath(t *testing.T) {
	connection := &vclib.VSphereConnection{
		Hostname: "should-not-matter",
		Port:     "should-not-matter",
		CACert:   "invalid-path",
	}

	_, err := connection.NewClient(context.Background())

	if err != vclib.ErrCaCertNotReadable {
		t.Fatalf("should have occurred")
	}
}

func TestInvalidCaCert(t *testing.T) {
	connection := &vclib.VSphereConnection{
		Hostname: "should-not-matter",
		Port:     "should-not-matter",
		CACert:   fixtures.InvalidCertPath,
	}

	_, err := connection.NewClient(context.Background())

	if err != vclib.ErrCaCertInvalid {
		t.Fatalf("should have occurred")
	}
}

func TestUnsupportedTransport(t *testing.T) {
	notHttpTransport := new(fakeTransport)

	connection := &vclib.VSphereConnection{
		Hostname: "should-not-matter",
		Port:     "should-not-matter",
		CACert:   fixtures.CaCertPath,
	}

	err := connection.ConfigureTransportWithCA(notHttpTransport)
	if err != vclib.ErrUnsupportedTransport {
		t.Fatalf("should have occurred")
	}
}

type fakeTransport struct{}

func (ft fakeTransport) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, nil
}

func getRequestVerifier(t *testing.T) (http.HandlerFunc, func()) {
	gotRequest := false

	handler := func(w http.ResponseWriter, r *http.Request) {
		gotRequest = true
	}

	checker := func() {
		if !gotRequest {
			t.Fatalf("Never saw a request, maybe TLS connection could not be established?")
		}
	}

	return handler, checker
}

func mustParseUrl(t *testing.T, i string) *url.URL {
	u, err := url.Parse(i)
	if err != nil {
		t.Fatalf("Cannot parse URL: %v", err)
	}
	return u
}
