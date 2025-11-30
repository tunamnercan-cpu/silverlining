package silverlining

import (
	"crypto/tls"
	"errors"
	"testing"
)

func TestBuildTLSConfigRequiresPair(t *testing.T) {
	_, err := buildTLSConfig(&TLSOptions{CertFile: "server.crt"})
	if !errors.Is(err, ErrTLSCertificatePairRequired) {
		t.Fatalf("expected ErrTLSCertificatePairRequired, got %v", err)
	}
}

func TestBuildTLSConfigCustomConfig(t *testing.T) {
	opts := &TLSOptions{
		Config: &tls.Config{MinVersion: tls.VersionTLS11},
		CustomConfig: func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			if info.ServerName == "modern.local" {
				return &tls.Config{MinVersion: tls.VersionTLS13}, nil
			}
			return nil, nil
		},
	}

	cfg, err := buildTLSConfig(opts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.GetConfigForClient == nil {
		t.Fatal("expected GetConfigForClient to be set")
	}

	customCfg, err := cfg.GetConfigForClient(&tls.ClientHelloInfo{ServerName: "modern.local"})
	if err != nil {
		t.Fatalf("unexpected custom config error: %v", err)
	}

	if customCfg.MinVersion != tls.VersionTLS13 {
		t.Fatalf("expected MinVersion TLS 1.3, got %d", customCfg.MinVersion)
	}

	fallbackCfg, err := cfg.GetConfigForClient(&tls.ClientHelloInfo{ServerName: "legacy.local"})
	if err != nil {
		t.Fatalf("unexpected fallback config error: %v", err)
	}

	if fallbackCfg.MinVersion != tls.VersionTLS11 {
		t.Fatalf("expected fallback MinVersion TLS 1.1, got %d", fallbackCfg.MinVersion)
	}
}
