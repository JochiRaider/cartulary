package appsupport

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

func TestStartRuntimeAcceptsCanonicalCompositeDeclaration(t *testing.T) {
	restoreRuntimeStarters(t)
	postgres := &pgtest.Harness{}
	objectStore := &s3test.Harness{}
	startPostgresForRuntime = func(testing.TB) *pgtest.Harness { return postgres }
	startS3ForRuntime = func(testing.TB) *s3test.Harness { return objectStore }

	runtime, err := startRuntime(t, map[string]string{
		suiteservices.HarnessServiceDependenciesEnv: "object_store,postgres",
	})
	if err != nil {
		t.Fatalf("start composite runtime: %v", err)
	}
	if runtime.Postgres != postgres || runtime.S3 != objectStore {
		t.Fatal("runtime did not retain both explicitly declared service harnesses")
	}
}

func TestStartRuntimeRejectsOmittedServiceBeforeAcquisition(t *testing.T) {
	restoreRuntimeStarters(t)
	postgresStarts := 0
	objectStoreStarts := 0
	startPostgresForRuntime = func(testing.TB) *pgtest.Harness {
		postgresStarts++
		return &pgtest.Harness{}
	}
	startS3ForRuntime = func(testing.TB) *s3test.Harness {
		objectStoreStarts++
		return &s3test.Harness{}
	}

	runtime, err := startRuntime(t, map[string]string{
		suiteservices.HarnessServiceDependenciesEnv: "postgres",
	})
	if err == nil || runtime != nil {
		t.Fatalf("expected omitted object-store dependency to fail: runtime=%v err=%v", runtime, err)
	}
	if postgresStarts != 0 || objectStoreStarts != 0 {
		t.Fatalf("guard ran after acquisition: postgres=%d object_store=%d", postgresStarts, objectStoreStarts)
	}
	envelope := runtimeDependencyEnvelope(err)
	if envelope.FailureClass != "harness" || envelope.FailureReason != "fixture_error" ||
		envelope.Service != "object_store" || envelope.ReadinessStage != "dependency_guard" {
		t.Fatalf("unexpected dependency envelope: %#v", envelope)
	}
}

func restoreRuntimeStarters(t testing.TB) {
	t.Helper()
	oldPostgres := startPostgresForRuntime
	oldS3 := startS3ForRuntime
	t.Cleanup(func() {
		startPostgresForRuntime = oldPostgres
		startS3ForRuntime = oldS3
	})
}
