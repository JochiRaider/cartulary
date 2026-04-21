package main

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

var (
	processHarnessesErr error
	processPostgres     *pgtest.Harness
	processS3           *s3test.Harness
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	if err := startSharedProcessHarnesses(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "start shared process harnesses: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	if err := s3test.StopShared(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "terminate shared minio testcontainer: %v\n", err)
		code = 1
	}
	if err := pgtest.StopShared(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "terminate shared postgres testcontainer: %v\n", err)
		code = 1
	}

	os.Exit(code)
}

func startSharedProcessHarnesses(ctx context.Context) error {
	processPostgres, processHarnessesErr = pgtest.StartShared(ctx)
	if processHarnessesErr != nil {
		return processHarnessesErr
	}

	processS3, processHarnessesErr = s3test.StartShared(ctx)
	if processHarnessesErr != nil {
		_ = pgtest.StopShared(ctx)
		processPostgres = nil
		return processHarnessesErr
	}

	return nil
}

func sharedProcessHarnesses(t testing.TB) (*pgtest.Harness, *s3test.Harness) {
	t.Helper()

	if processHarnessesErr != nil {
		t.Fatalf("shared process harnesses unavailable: %v", processHarnessesErr)
	}
	if processPostgres == nil || processS3 == nil {
		t.Fatal("shared process harnesses unavailable: package setup did not initialize postgres and minio")
	}

	return processPostgres, processS3
}
