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
	//   - "https://" scheme enables TLS, "http://" scheme is removed.
	Collector string `cfg:"collector" json:"collector"`
	// TLS settings, also gets from OTEL_EXPORTER_OTLP_INSECURE=false, OTEL_EXPORTER_OTLP_CERTIFICATE,
	// OTEL_EXPORTER_OTLP_CLIENT_CERTIFICATE and OTEL_EXPORTER_OTLP_CLIENT_KEY env vars.
	TLS TLSConfig `cfg:"tls" json:"tls"`
	// ServerName sets the grpc.WithAuthority with extract host(server_name) for connection.
	ServerName string `cfg:"server_name" json:"server_name"`
	// Compression for grpc calls, "gzip" or "none".
	//   - Default gets from OTEL_EXPORTER_OTLP_COMPRESSION env var.
	Compression string `cfg:"compression" json:"compression"`

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
	// Disabled to use noop metric provider.
	//   - Also enabled with OTEL_SDK_DISABLED=true or OTEL_METRICS_EXPORTER=none env vars.
	Disabled bool          `cfg:"disabled" json:"disabled"`
	Default  MetricDefault `cfg:"default" json:"default"`
}

type MetricDefault struct {
	GoRuntime bool `cfg:"go_runtime" json:"go_runtime"`
}

type MetricProviderSettings struct {
	// Interval between metric exports.
	//   - Default gets from OTEL_METRIC_EXPORT_INTERVAL env var (milliseconds), otherwise 5s.
	Interval time.Duration `cfg:"interval" json:"interval"`
}

type TraceSettings struct {
	Provider TraceProviderSettings `cfg:"provider" json:"provider"`
	// Disabled to use noop trace provider.
	//   - Also enabled with OTEL_SDK_DISABLED=true or OTEL_TRACES_EXPORTER=none env vars.
	Disabled bool `cfg:"disabled" json:"disabled"`
}

// TraceProviderSettings for trace provider.
//
// Sampler gets from OTEL_TRACES_SAMPLER and OTEL_TRACES_SAMPLER_ARG env vars,
// default is parentbased_always_on.
// Batch span processor gets from OTEL_BSP_* env vars.
type TraceProviderSettings struct{}
