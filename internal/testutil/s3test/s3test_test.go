package s3test

import (
	"bytes"
	"context"
	"testing"
)

func TestHarnessStartsMinIOAndRoundTripsObjects(t *testing.T) {
	harness := Start(t)

	bucket, err := harness.BootstrapBucket(context.Background(), "bootstrap")
	if err != nil {
		t.Fatalf("bootstrap bucket: %v", err)
	}
	defer func() {
		if err := harness.CleanupBucket(context.Background(), bucket); err != nil {
			t.Fatalf("cleanup bucket: %v", err)
		}
	}()

	payload := []byte("cartulary bootstrap object")
	got, err := harness.RoundTrip(context.Background(), bucket, "proof.txt", payload)
	if err != nil {
		t.Fatalf("round trip object: %v", err)
	}

	if !bytes.Equal(got, payload) {
		t.Fatalf("unexpected object payload: %q", got)
	}
}
