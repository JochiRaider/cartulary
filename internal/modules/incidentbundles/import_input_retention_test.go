package incidentbundles

import (
	"context"
	"errors"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
)

func testImportInputRetention(t *testing.T) {
	reference, err := ParseBundleStagingRef("incident-bundles/imports/frozen.bundle")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		status  string
		readErr error
		remove  bool
	}{
		{"queued", jobs.StatusQueued, nil, false},
		{"running", jobs.StatusRunning, nil, false},
		{"cancel requested", jobs.StatusCancelRequested, nil, false},
		{"succeeded", jobs.StatusSucceeded, nil, true},
		{"failed", jobs.StatusFailed, nil, true},
		{"canceled", jobs.StatusCanceled, nil, true},
		{"unknown read", jobs.StatusSucceeded, errors.New("lost acknowledgement"), false},
	} {
		t.Run("input retention/"+tc.name, func(t *testing.T) {
			storage := &recordingBundleStorage{staged: []byte("frozen input")}
			manager := &recordingJobOperations{getResource: jobs.Resource{Status: tc.status}, getErr: tc.readErr}
			worker := &incidentBundleWorker{storage: storage, jobManager: manager}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			jobID := uuid.New()
			worker.removeTerminalImportInput(ctx, jobID, reference)
			var wantRemovals []BundleStagingRef
			if tc.remove {
				wantRemovals = []BundleStagingRef{reference}
			}
			if !slices.Equal(manager.getJobIDs, []uuid.UUID{jobID}) || !slices.Equal(storage.stagedRemovals, wantRemovals) {
				t.Fatalf("observed Jobs=%v, removed inputs=%v; want Job=%v, inputs=%v", manager.getJobIDs, storage.stagedRemovals, jobID, wantRemovals)
			}
		})
	}
}
