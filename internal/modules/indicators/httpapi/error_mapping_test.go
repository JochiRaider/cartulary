package httpapi

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
)

func TestIndicatorMutationErrorRoleAndStorageMapping(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   string
	}{
		{name: "source", err: indicators.ErrIndicatorSourceNotFound, wantStatus: http.StatusNotFound, wantCode: "indicator_source_record_not_found"},
		{name: "requested resolved Indicator", err: indicators.ErrResolvedIndicatorNotFound, wantStatus: http.StatusNotFound, wantCode: "resolved_indicator_not_found"},
		{name: "addressed Indicator", err: indicators.ErrIndicatorNotFound, wantStatus: http.StatusNotFound, wantCode: "indicator_not_found"},
		{name: "prior observation dependency", err: indicators.ErrIndicatorObservationNotFound, wantStatus: http.StatusNotFound, wantCode: "indicator_observation_not_found"},
		{name: "support reference", err: indicators.ErrInvalidCreateRequest, wantStatus: http.StatusBadRequest, wantCode: "invalid_mutation_payload"},
		{name: "storage", err: errors.New("secret relation records constraint records_pkey"), wantStatus: http.StatusInternalServerError, wantCode: "internal_error"},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			mapped := mutationError(test.err, "txn-role-mapping")
			if mapped.Status != test.wantStatus || mapped.Code != test.wantCode {
				t.Fatalf("mapped error = %#v, want status=%d code=%q", mapped, test.wantStatus, test.wantCode)
			}
			serialized := mapped.Code + " " + mapped.Message
			for _, secret := range []string{"records", "constraint", "records_pkey"} {
				if strings.Contains(serialized, secret) {
					t.Fatalf("mapped error leaks storage diagnostic %q: %#v", secret, mapped)
				}
			}
		})
	}
}
