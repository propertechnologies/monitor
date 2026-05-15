package metrics

import (
	"testing"
	"time"
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
