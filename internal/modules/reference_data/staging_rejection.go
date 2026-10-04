package reference_data

import (
	"context"
	"errors"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

func importStagingError(ctx context.Context, err error) error {
	if cause := ctx.Err(); cause != nil {
		return cause
	}
	// Single-cause wrapping preserves classification; joined operational
	// failures deliberately do not expose a publishable content rejection.
	for cause := err; cause != nil; cause = errors.Unwrap(cause) {
		bound, ok := cause.(*StagingLimitError)
		if !ok {
			continue
		}
		if bound.Maximum < 1 || bound.Observed != uint64(bound.Maximum)+1 {
			return errors.New("reference pack: invalid staging limit measurement")
		}
		id, expected, actual := "max_container_bytes", strconv.FormatInt(bound.Maximum, 10), strconv.FormatUint(bound.Observed, 10)
		summary, buildErr := packformat.CheckSummary(ctx, "container_bytes", nil, func(_ context.Context, emit packformat.FindingSink) error {
			return emit(packformat.Finding{Path: "$", Details: packformat.SafeDetails{LimitID: &id, ExpectedToken: &expected, ActualToken: &actual}})
		})
		if buildErr != nil {
			return buildErr
		}
		return &ContentRejection{Code: "container_bytes_exceeded", CheckID: "container_bytes", Summary: summary}
	}
	return err
}

// ImportContentRejection recognizes a sole typed rejection for transport
// adapters, without hiding any joined operational failure.
func ImportContentRejection(err error) (*ContentRejection, bool) {
	for cause := err; cause != nil; cause = errors.Unwrap(cause) {
		if rejection, ok := cause.(*ContentRejection); ok {
			return rejection, true
		}
	}
	return nil, false
}
