//go:build cartulary_harness

package app

import "testing"

func TestHarnessServerProfileEnablesTestRoutesOnlyForExactOne(t *testing.T) {
	for _, tc := range []struct {
		value       string
		wantEnabled bool
	}{
		{value: "", wantEnabled: false},
		{value: "true", wantEnabled: false},
		{value: "1", wantEnabled: true},
	} {
		t.Run(tc.value, func(t *testing.T) {
			runner := newServerRunner(nil, nil)
			runner.lookupEnv = func(key string) (string, bool) {
				if key == "CARTULARY_ENABLE_TEST_ROUTES" {
					return tc.value, tc.value != ""
				}
				return "", false
			}
			options := runner.profile.runtimeOptions(runner.lookupEnv)
			if got := len(options.HTTP.AdditionalRoutes) > 0; got != tc.wantEnabled {
				t.Fatalf("routes enabled got %v want %v", got, tc.wantEnabled)
			}
		})
	}
}
