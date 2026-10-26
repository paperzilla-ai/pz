package api

import (
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

// CVE-2025-61729: formatting a mismatched certificate must not expand every
// peer-controlled SAN. The IP case was also quadratic before the Go fix.
func TestTLSHostnameErrorFormattingIsBounded(t *testing.T) {
	cert := &x509.Certificate{}
	for i := 0; i < 5000; i++ {
		cert.IPAddresses = append(cert.IPAddresses, net.IPv4(192, 0, 2, 10))
		cert.DNSNames = append(cert.DNSNames, fmt.Sprintf("host%d.example.invalid", i))
	}
	for _, tc := range []struct{ name, mismatch, valid string }{
		{"IP SAN", "192.0.2.11", "192.0.2.10"},
		{"DNS SAN", "missing.example.invalid", "host1.example.invalid"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			err := cert.VerifyHostname(tc.mismatch)
			var hostnameErr x509.HostnameError
			if !errors.As(err, &hostnameErr) {
				t.Fatalf("mismatch error type = %T, want x509.HostnameError", err)
			}
			if n := len(err.Error()); n > 1024 {
				t.Errorf("certificate error expands SAN list to %d bytes", n)
			}
			if err := cert.VerifyHostname(tc.valid); err != nil {
				t.Fatalf("valid hostname rejected: %v", err)
			}
		})
	}
}

func TestFetchProjectsOverTLS(t *testing.T) {
	for _, http2 := range []bool{false, true} {
		t.Run(fmt.Sprintf("http2=%t", http2), func(t *testing.T) {
			wantProtocol := 1
			if http2 {
				wantProtocol = 2
			}
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.TLS == nil || r.ProtoMajor != wantProtocol {
					t.Errorf("connection = %s, TLS = %t; want HTTPS/%d", r.Proto, r.TLS != nil, wantProtocol)
				}
				if r.Method != http.MethodGet || r.URL.Path != "/api/projects" {
					t.Errorf("request = %s %s", r.Method, r.URL.Path)
				}
				if r.Header.Get("Authorization") != "Bearer test-token" {
					t.Error("missing bearer token")
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`[{"id":"tls-project","name":"TLS project"}]`))
			}))
			server.EnableHTTP2 = http2
			server.StartTLS()
			t.Cleanup(server.Close)

			originalClient := http.DefaultClient
			http.DefaultClient = server.Client()
			t.Cleanup(func() {
				http.DefaultClient.CloseIdleConnections()
				http.DefaultClient = originalClient
			})
			t.Setenv("PZ_API_URL", server.URL)

			projects, err := FetchProjects("test-token")
			if err != nil {
				t.Fatalf("FetchProjects over TLS: %v", err)
			}
			if len(projects) != 1 || projects[0].ID != "tls-project" {
				t.Fatalf("projects = %+v", projects)
			}
		})
	}
}
