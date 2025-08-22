package tell

import (
	"crypto/tls"
	"time"

	"github.com/twmb/tlscfg"
)

type Config struct {
	// Collector to show URL of grpc otel collector.
	//   - If empty disable for metric and trace. It is add a noop metric/trace and your code works without change.
	//   - Default gets from OTEL_EXPORTER_OTLP_ENDPOINT env var.
	Collector string    `cfg:"collector" json:"collector"`
	TLS       TLSConfig `cfg:"tls" json:"tls"`
	// ServerName sets the grpc.WithAuthority with extract host(server_name) for connection.
	ServerName string `cfg:"server_name" json:"server_name"`

	Metric MetricSettings `cfg:"metric" json:"metric"`
	Trace  TraceSettings  `cfg:"trace" json:"trace"`

	Logger Logger `cfg:"-" json:"-"`
}

type TLSConfig struct {
	Enabled            bool `cfg:"enabled" json:"enabled"`
	InsecureSkipVerify bool `cfg:"insecure_skip_verify" json:"insecure_skip_verify"`
	// CertFile is the path to the client's TLS certificate.
	// Should be use with KeyFile.
	CertFile string `cfg:"cert_file" json:"cert_file"`
	// KeyFile is the path to the client's TLS key.
	// Should be use with CertFile.
	KeyFile string `cfg:"key_file" json:"key_file"`
	// CAFile is the path to the CA certificate.
	// If empty, the server's root CA set will be used.
	CAFile string `cfg:"ca_file" json:"ca_file"`
}

// Generate returns a tls.Config based on the TLSConfig.
//
// If the TLSConfig is not enabled, nil is returned.
func (t TLSConfig) Generate() (*tls.Config, error) {
	if !t.Enabled {
		return nil, nil
	}

	opts := []tlscfg.Opt{}

	// load client cert
	if t.CertFile != "" && t.KeyFile != "" {
		opts = append(opts, tlscfg.WithDiskKeyPair(t.CertFile, t.KeyFile))
	}

	// load CA cert
	opts = append(opts, tlscfg.WithSystemCertPool())
	if t.CAFile != "" {
		opts = append(opts, tlscfg.WithDiskCA(t.CAFile, tlscfg.ForClient))
	}

	cfg, err := tlscfg.New(opts...)
	if err != nil {
		return nil, err
	}

	if t.InsecureSkipVerify {
		cfg.InsecureSkipVerify = true
	}

	return cfg, nil
}

type MetricSettings struct {
	Provider MetricProviderSettings `cfg:"provider" json:"provider"`
	Disabled bool                   `cfg:"disabled" json:"disabled"`
	Default  MetricDefault          `cfg:"default" json:"default"`
}

type MetricDefault struct {
	GoRuntime bool `cfg:"go_runtime" json:"go_runtime"`
}

type MetricProviderSettings struct {
	Interval time.Duration `cfg:"interval" json:"interval"`
}

type TraceSettings struct {
	Provider TraceProviderSettings `cfg:"provider" json:"provider"`
	Disabled bool                  `cfg:"disabled" json:"disabled"`
}

type TraceProviderSettings struct{}
