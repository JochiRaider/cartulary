package telemetry

import (
	"testing"

	"go.opentelemetry.io/otel/trace"
)

func TestResolveSamplerProfile(t *testing.T) {
	for _, tc := range []struct {
		name    string
		ratio   float64
		profile string
		want    string
	}{
		{name: "auto off", ratio: 0, profile: SamplerProfileAuto, want: SamplerProfileAlwaysOff},
		{name: "auto on", ratio: 1, profile: SamplerProfileAuto, want: SamplerProfileAlwaysOn},
		{name: "auto fractional", ratio: 0.10, profile: SamplerProfileAuto, want: SamplerProfileTraceIDRatioCompat},
		{name: "explicit off", ratio: 0.50, profile: SamplerProfileAlwaysOff, want: SamplerProfileAlwaysOff},
		{name: "explicit on", ratio: 0.50, profile: SamplerProfileAlwaysOn, want: SamplerProfileAlwaysOn},
		{name: "explicit traceidratio", ratio: 1, profile: SamplerProfileTraceIDRatioCompat, want: SamplerProfileTraceIDRatioCompat},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := ResolveSamplerProfile(tc.ratio, tc.profile); got != tc.want {
				t.Fatalf("ResolveSamplerProfile(%f, %q) = %q want %q", tc.ratio, tc.profile, got, tc.want)
			}
		})
	}
}

func TestFixedTraceIDRatioCompatCorpus(t *testing.T) {
	if SamplerProfileReviewAfter != "2027-01-01" {
		t.Fatalf("sampler review metadata drifted: %q", SamplerProfileReviewAfter)
	}
	if SamplerProfileTraceIDRatioCompat != "cartulary.sampler.traceidratio_compat.v1" {
		t.Fatalf("sampler fractional profile drifted: %q", SamplerProfileTraceIDRatioCompat)
	}
	for _, row := range FixedTraceIDSamplerCorpus() {
		traceID, err := trace.TraceIDFromHex(row.TraceID)
		if err != nil {
			t.Fatalf("parse fixed trace ID %q: %v", row.TraceID, err)
		}
		if got := TraceIDRatioCompatAllows(traceID, row.Ratio); got != row.Allow {
			t.Fatalf("TraceIDRatioCompatAllows(%q, %f) = %t want %t", row.TraceID, row.Ratio, got, row.Allow)
		}
	}
	zeroTraceID, _ := trace.TraceIDFromHex("00000000000000000000000000000000")
	if TraceIDRatioCompatAllows(zeroTraceID, 1.0) {
		t.Fatal("invalid zero trace ID must not be sampled")
	}
}
