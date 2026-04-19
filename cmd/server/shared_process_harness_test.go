package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

var (
	processHarnessesOnce sync.Once
	processHarnessesErr  error
	processPostgres      *pgtest.Harness
	processS3            *s3test.Harness
)

func TestMain(m *testing.M) {
	ctx := context.Background()
	code := m.Run()

	if processS3 != nil {
		if err := processS3.Close(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "terminate shared minio testcontainer: %v\n", err)
			code = 1
		}
	}
	if processPostgres != nil {
		if err := processPostgres.Close(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "terminate shared postgres testcontainer: %v\n", err)
			code = 1
		}
	}

	os.Exit(code)
}

func sharedProcessHarnesses(t testing.TB) (*pgtest.Harness, *s3test.Harness) {
	t.Helper()

	processHarnessesOnce.Do(func() {
		ctx := context.Background()

		processPostgres, processHarnessesErr = pgtest.StartShared(ctx)
		if processHarnessesErr != nil {
			return
		}

		processS3, processHarnessesErr = s3test.StartShared(ctx)
		if processHarnessesErr != nil {
			_ = processPostgres.Close(ctx)
			processPostgres = nil
		}
	})

	if processHarnessesErr != nil {
		t.Fatalf("start shared process harnesses: %v", processHarnessesErr)
	}

	return processPostgres, processS3
}
