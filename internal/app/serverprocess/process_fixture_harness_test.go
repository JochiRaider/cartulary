package serverprocess

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

var (
	processPostgres *pgtest.Harness
	processS3       *s3test.Harness
)

type processHarnessStarters struct {
	startPostgres    func(context.Context) (*pgtest.Harness, error)
	startObjectStore func(context.Context) (*s3test.Harness, error)
}

func TestMain(m *testing.M) {
	lifecycleCtx, cancelLifecycle := context.WithTimeout(context.Background(), 30*time.Minute)
	postgresHarness, objectStoreHarness, err := startProcessHarnesses(lifecycleCtx, processHarnessStarters{
		startPostgres:    pgtest.StartShared,
		startObjectStore: s3test.StartShared,
	})
	if err != nil {
		cancelLifecycle()
		fmt.Fprintf(os.Stderr, "start shared process harnesses: %v\n", err)
		os.Exit(1)
	}
	processPostgres = postgresHarness
	processS3 = objectStoreHarness

	code := m.Run()

	cleanupCtx, cancelCleanup := context.WithTimeout(context.Background(), 2*time.Minute)
	if err := processS3.Close(cleanupCtx); err != nil {
		fmt.Fprintf(os.Stderr, "release borrowed object-store transport: %v\n", err)
		code = 1
	}
	cancelCleanup()
	cancelLifecycle()

	os.Exit(code)
}

func startProcessHarnesses(ctx context.Context, starters processHarnessStarters) (*pgtest.Harness, *s3test.Harness, error) {
	postgresHarness, err := starters.startPostgres(ctx)
	if err != nil {
		return nil, nil, err
	}

	objectStoreHarness, err := starters.startObjectStore(ctx)
	if err != nil {
		return nil, nil, err
	}

	return postgresHarness, objectStoreHarness, nil
}

func sharedProcessHarnesses(t testing.TB) (*pgtest.Harness, *s3test.Harness) {
	t.Helper()

	if processPostgres == nil || processS3 == nil {
		t.Fatal("shared process harnesses unavailable: package setup did not initialize postgres and object store")
	}

	return processPostgres, processS3
}
