package telemetry

import "context"

func RecordIncidentBundleV1Import(ctx context.Context, serviceVersion string) {
	counter, err := Meter(ScopePortability, serviceVersion).Int64Counter(
		IncidentBundleV1ImportMetricName,
	)
	if err != nil {
		return
	}
	counter.Add(ctx, 1)
}
