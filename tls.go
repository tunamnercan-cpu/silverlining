package silverlining

import (
	"crypto/tls"
	"errors"
	"net"
)

var (
	// ErrTLSConfigRequired indicates that a nil *tls.Config was supplied to a TLS server.
	ErrTLSConfigRequired = errors.New("silverlining: tls config is required")
	// ErrTLSCertificatePairRequired indicates that only one of certificate or key files was provided.
	ErrTLSCertificatePairRequired = errors.New("silverlining: both certificate and key files are required")
)

// TLSOptions describes the knobs that can be used when starting a TLS listener.
// The zero value is valid if Config already provides certificates or GetCertificate hooks.
type TLSOptions struct {
	// CertFile and KeyFile are optional helpers for loading a certificate pair from disk.
	CertFile string
	KeyFile  string

	// Config holds a user supplied *tls.Config. When nil, a new empty config is created.
	Config *tls.Config

	// CustomConfig lets callers pick a configuration per ClientHello (for example, for SNI support).
	// Returning nil falls back to the default configuration built from CertFile/KeyFile or Config.
	CustomConfig func(info *tls.ClientHelloInfo) (*tls.Config, error)
}

// ListenAndServeTLS starts a TLS server on addr with the provided handler and options.
func ListenAndServeTLS(addr string, handler Handler, opts *TLSOptions) error {
	cfg, err := buildTLSConfig(opts)
	if err != nil {
		return err
	}

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	srv := &Server{
		Handler: handler,
	}

	return srv.ServeTLS(ln, cfg)
}

// ServeTLS wraps the provided listener with TLS using cfg and serves incoming connections.
func (s *Server) ServeTLS(l net.Listener, cfg *tls.Config) error {
	if cfg == nil {
		return ErrTLSConfigRequired
	}

	tlsListener := tls.NewListener(l, cfg)
	return s.Serve(tlsListener)
}

func buildTLSConfig(opts *TLSOptions) (*tls.Config, error) {
	var cfg *tls.Config
	if opts != nil && opts.Config != nil {
		cfg = opts.Config.Clone()
	} else {
		cfg = &tls.Config{}
	}

	if opts != nil && (opts.CertFile != "" || opts.KeyFile != "") {
		if opts.CertFile == "" || opts.KeyFile == "" {
			return nil, ErrTLSCertificatePairRequired
		}

		cert, err := tls.LoadX509KeyPair(opts.CertFile, opts.KeyFile)
		if err != nil {
			return nil, err
		}

		cfg.Certificates = append(cfg.Certificates, cert)
	}

	if opts != nil && opts.CustomConfig != nil {
		defaultCfg := cfg.Clone()
		defaultCfg.GetConfigForClient = nil

		customFn := opts.CustomConfig
		cfg.GetConfigForClient = func(info *tls.ClientHelloInfo) (*tls.Config, error) {
			customCfg, err := customFn(info)
			if err != nil {
				return nil, err
			}

			if customCfg == nil {
				return defaultCfg, nil
			}

			return customCfg, nil
		}
	}

	return cfg, nil
}
