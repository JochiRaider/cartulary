package telemetry

import (
	"encoding/binary"

	"go.opentelemetry.io/otel/trace"
)

const (
	SamplerProfileAuto               = "auto"
	SamplerProfileAlwaysOn           = "cartulary.sampler.always_on.v1"
	SamplerProfileAlwaysOff          = "cartulary.sampler.always_off.v1"
	SamplerProfileTraceIDRatioCompat = "cartulary.sampler.traceidratio_compat.v1"
	SamplerProfileReviewAfter        = "2027-01-01"
)

type SamplerCorpusRow struct {
	TraceID string
	Ratio   float64
	Allow   bool
}

func ResolveSamplerProfile(sampleRatio float64, samplerProfile string) string {
	switch samplerProfile {
	case SamplerProfileAlwaysOn, SamplerProfileAlwaysOff, SamplerProfileTraceIDRatioCompat:
		return samplerProfile
	}
	switch {
	case sampleRatio <= 0:
		return SamplerProfileAlwaysOff
	case sampleRatio >= 1:
		return SamplerProfileAlwaysOn
	default:
		return SamplerProfileTraceIDRatioCompat
	}
}

func TraceIDRatioCompatAllows(traceID trace.TraceID, sampleRatio float64) bool {
	if sampleRatio <= 0 || !traceID.IsValid() {
		return false
	}
	if sampleRatio >= 1 {
		return true
	}
	x := binary.BigEndian.Uint64(traceID[8:16]) >> 1
	upperBound := uint64(sampleRatio * (1 << 63))
	return x < upperBound
}

func FixedTraceIDSamplerCorpus() []SamplerCorpusRow {
	return []SamplerCorpusRow{
		{TraceID: "10000000000000000000000000000000", Ratio: 0.0, Allow: false},
		{TraceID: "10000000000000000000000000000000", Ratio: 1.0, Allow: true},
		{TraceID: "10000000000000000000000000000000", Ratio: 0.25, Allow: true},
		{TraceID: "10000000000000004000000000000000", Ratio: 0.25, Allow: false},
		{TraceID: "1000000000000000ffffffffffffffff", Ratio: 0.50, Allow: false},
	}
}
