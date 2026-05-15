package tracing

import (
	"testing"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestSamplerFromEnv(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want sdktrace.Sampler
	}{
		{"unset", "", sdktrace.AlwaysSample()},
		{"invalid", "not-a-float", sdktrace.AlwaysSample()},
		{"out_of_range_high", "1.5", sdktrace.AlwaysSample()},
		{"out_of_range_low", "-0.1", sdktrace.AlwaysSample()},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Setenv(EnvSampleRatio, c.raw)
			got := samplerFromEnv()
			if got.Description() != c.want.Description() {
				t.Errorf("samplerFromEnv() with %q = %q, want %q",
					c.raw, got.Description(), c.want.Description())
			}
		})
	}
}

func TestSamplerFromEnvValidRatio(t *testing.T) {
	t.Setenv(EnvSampleRatio, "0.25")
	got := samplerFromEnv()
	// ParentBased(TraceIDRatioBased(x)) has a non-AlwaysOn description; the
	// concrete string isn't part of the contract, just confirm we got a different
	// sampler than the AlwaysSample default.
	if got.Description() == sdktrace.AlwaysSample().Description() {
		t.Errorf("expected non-AlwaysSample sampler for 0.25; got %q", got.Description())
	}
}

func TestTracerShutdownNilSafe(t *testing.T) {
	var tr *Tracer
	if err := tr.Shutdown(nil); err != nil {
		t.Errorf("Shutdown on nil Tracer: %v", err)
	}
	if err := (&Tracer{}).Shutdown(nil); err != nil {
		t.Errorf("Shutdown on empty Tracer: %v", err)
	}
}
