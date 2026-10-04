package reference_data

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
)

func testStagingRejectionClassification(t *testing.T) {
	for _, maximum := range []int64{128, math.MaxInt64} {
		bound := &StagingLimitError{Maximum: maximum, Observed: uint64(maximum) + 1}
		err := importStagingError(context.Background(), fmt.Errorf("storage: %w", bound))
		rejection, ok := ImportContentRejection(err)
		if !ok || rejection.Code != "container_bytes_exceeded" || rejection.Summary == nil {
			t.Fatal(err)
		}
		api := coordinatorAPIError(err)
		if api.Status != 409 || api.Code != "reference_pack_verification_failed" {
			t.Fatal(api)
		}
		expected, observed := "128", "129"
		if maximum == math.MaxInt64 {
			expected, observed = "9223372036854775807", "9223372036854775808"
		}
		details := rejection.Summary.Issues[0].Details
		if details.ExpectedToken == nil || *details.ExpectedToken != expected || details.ActualToken == nil || *details.ActualToken != observed {
			t.Fatal("unsafe numeric diagnostics", details)
		}
		operational := errors.New("storage unavailable")
		for _, failure := range []error{operational, errors.Join(bound, operational), errors.Join(err, operational)} {
			classified := importStagingError(context.Background(), failure)
			if _, ok := ImportContentRejection(classified); ok || coordinatorAPIError(classified).Status != 500 {
				t.Fatal("operational fault became rejection", classified)
			}
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := importStagingError(ctx, bound); !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	}
}
