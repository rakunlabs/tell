package tell

import (
	"context"
	"fmt"
	"path"
	"strings"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// httpURLPath returns the export path of the signal under the collector's base path,
// e.g. "/v1/metrics" or "/otlp/v1/traces".
func (cfg Config) httpURLPath(signal string) string {
	return path.Join("/", cfg.urlPath, "v1", signal)
}

// httpCompression returns the compression of the config, empty value keeps the exporter's default.
func (cfg Config) httpCompression() (string, error) {
	switch c := strings.ToLower(strings.TrimSpace(cfg.Compression)); c {
	case "", "none", "gzip":
		return c, nil
	default:
		return "", fmt.Errorf("unsupported compression [%s], use gzip or none", cfg.Compression)
	}
}

func (cfg Config) metricHTTPOptions() ([]otlpmetrichttp.Option, error) {
	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.Collector),
		otlpmetrichttp.WithURLPath(cfg.httpURLPath("metrics")),
	}

	if cfg.TLS.Enabled {
		tlsConfig, err := cfg.TLS.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate tls config; %w", err)
		}

		if cfg.ServerName != "" {
			tlsConfig.ServerName = cfg.ServerName
		}

		opts = append(opts, otlpmetrichttp.WithTLSClientConfig(tlsConfig))
	} else {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	compression, err := cfg.httpCompression()
	if err != nil {
		return nil, err
	}

	switch compression {
	case "gzip":
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.GzipCompression))
	case "none":
		opts = append(opts, otlpmetrichttp.WithCompression(otlpmetrichttp.NoCompression))
	}

	return opts, nil
}

func (cfg Config) traceHTTPOptions() ([]otlptracehttp.Option, error) {
	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Collector),
		otlptracehttp.WithURLPath(cfg.httpURLPath("traces")),
	}

	if cfg.TLS.Enabled {
		tlsConfig, err := cfg.TLS.Generate()
		if err != nil {
			return nil, fmt.Errorf("failed to generate tls config; %w", err)
		}

		if cfg.ServerName != "" {
			tlsConfig.ServerName = cfg.ServerName
		}

		opts = append(opts, otlptracehttp.WithTLSClientConfig(tlsConfig))
	} else {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	compression, err := cfg.httpCompression()
	if err != nil {
		return nil, err
	}

	switch compression {
	case "gzip":
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.GzipCompression))
	case "none":
		opts = append(opts, otlptracehttp.WithCompression(otlptracehttp.NoCompression))
	}

	return opts, nil
}

// MetricProviderHTTP sets the metric provider with an OTLP http exporter.
func (c *Collector) MetricProviderHTTP(ctx context.Context, cfg MetricProviderSettings, opts ...otlpmetrichttp.Option) error {
	metricExporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create metric exporter: %w", err)
	}

	c.metricProvider(metricExporter, cfg)

	return nil
}

// TraceProviderHTTP sets the trace provider with an OTLP http exporter.
func (c *Collector) TraceProviderHTTP(ctx context.Context, _ TraceProviderSettings, opts ...otlptracehttp.Option) error {
	traceExporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	c.traceProvider(traceExporter)

	return nil
}
