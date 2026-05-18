package metrics

import (
	"testing"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

func TestIntervalFromEnv(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want time.Duration
	}{
		{"unset", "", defaultExportInterval},
		{"valid_seconds", "30s", 30 * time.Second},
		{"valid_minutes", "2m", 2 * time.Minute},
		{"invalid", "not-a-duration", defaultExportInterval},
		{"zero", "0s", defaultExportInterval},
		{"negative", "-5s", defaultExportInterval},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(EnvMetricsInterval, c.raw)
			got := intervalFromEnv()
			if got != c.want {
				t.Errorf("intervalFromEnv() with %q = %v, want %v", c.raw, got, c.want)
			}
		})
	}
}

func TestShutdownNilSafe(t *testing.T) {
	var p *Provider
	if err := p.Shutdown(nil); err != nil {
		t.Errorf("Shutdown on nil receiver: %v", err)
	}
	if err := (&Provider{}).Shutdown(nil); err != nil {
		t.Errorf("Shutdown on empty Provider: %v", err)
	}
}

func TestResourceAttrsKeptAsMetricLabels(t *testing.T) {
	cases := []struct {
		name string
		key  string
		val  string
		want bool
	}{
		{"service.name kept (default filter)", "service.name", "contador", true},
		{"service.namespace kept (default filter)", "service.namespace", "proper", true},
		{"deployment.environment kept", "deployment.environment", "dev", true},
		{"deployment.environment.name kept", "deployment.environment.name", "prod", true},
		{"service.version kept", "service.version", "1.2.3", true},
		{"deployment.environment empty value rejected", "deployment.environment", "", false},
		{"unrelated attribute rejected", "host.name", "instance-abc", false},
		{"cloud.platform rejected", "cloud.platform", "gcp_cloud_run", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			kv := attribute.Key(c.key).String(c.val)
			got := resourceAttrsKeptAsMetricLabels(kv)
			if got != c.want {
				t.Errorf("resourceAttrsKeptAsMetricLabels(%s=%q) = %v, want %v",
					c.key, c.val, got, c.want)
			}
		})
	}
}
