package s3test

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/aws/smithy-go"
)

func TestObjectStoreMutationProbeRetriesTransientStagesOnSameLane(t *testing.T) {
	for _, stage := range []string{"put", "head"} {
		t.Run(stage, func(t *testing.T) {
			client := newFakeReadinessClient()
			client.failures[stage] = 1
			ops := readinessOperations(client)

			harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
			err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
				ReadyTimeout:    objectStoreClientReadyTimeout,
				PollInterval:    objectStoreHealthPollInterval,
				AttemptTimeout:  objectStoreClientAttemptTimeout,
				CreateNamespace: true,
				ListFirst:       true,
				Bucket:          "ct-suitehash-process1-readiness",
				Key:             ".cartulary-readiness/probe",
			})
			if err != nil {
				t.Fatalf("expected transient %s failure to recover on the same lane: %v", stage, err)
			}
			if got := client.callCount(stage); got < 2 {
				t.Fatalf("expected %s to be retried within readiness, got %d calls", stage, got)
			}
			client.assertNoProbeResidue(t)
		})
	}
}

func TestObjectStoreMutationProbeRejectsCapabilitiesImmediately(t *testing.T) {
	t.Run("deletion_requires_typed_absence", func(t *testing.T) {
		for _, cause := range []error{errors.New("transport unavailable"), &smithy.GenericAPIError{Code: "AccessDenied"}} {
			client := newFakeReadinessClient()
			client.namespaces["probe"] = true
			client.failures["delete_verify"] = 1
			client.failureErrors["delete_verify"] = cause
			failure := runObjectStoreMutationProbe(t.Context(), client, objectStoreReadinessConfig{Bucket: "probe", Key: "key"})
			if failure == nil || failure.Stage != "delete_verify" || !errors.Is(failure.Cause, cause) {
				t.Fatal("non-absence error accepted as deletion", failure)
			}
		}
	})
	client := newFakeReadinessClient()
	client.failureErrors["create_namespace"] = &smithy.GenericAPIError{Code: "AccessDenied"}
	client.failures["create_namespace"] = -1
	ops := readinessOperations(client)

	harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
	err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
		ReadyTimeout:    objectStoreClientReadyTimeout,
		PollInterval:    objectStoreHealthPollInterval,
		AttemptTimeout:  objectStoreClientAttemptTimeout,
		CreateNamespace: true,
		ListFirst:       true,
		Bucket:          "ct-suitehash-process1-readiness",
		Key:             ".cartulary-readiness/probe",
	})
	if err == nil {
		t.Fatal("expected capability rejection")
	}
	var readinessErr *objectStoreReadinessError
	if !errors.As(err, &readinessErr) {
		t.Fatalf("expected typed readiness error, got %T", err)
	}
	if readinessErr.Attempts != 1 || readinessErr.Stage != "create_namespace" {
		t.Fatalf("expected immediate create_namespace rejection, got %#v", readinessErr)
	}
	if client.callCount("delete_namespace") != 1 {
		t.Fatalf("capability rejection did not clean its proven probe namespace: %#v", client.calls)
	}
	if strings.Contains(err.Error(), "AccessDenied") {
		t.Fatalf("readiness diagnostic exposed provider error text: %q", err)
	}
	client.assertNoProbeResidue(t)
}

func TestObjectStoreMutationProbeDeadlineRetainsCleanupEvidence(t *testing.T) {
	t.Run("cancelled_between_probes", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		defer cancel()
		client := newFakeReadinessClient()
		primary := errors.New("transient put failure")
		client.failures["put"], client.failureErrors["put"] = 1, primary
		ops := readinessOperations(client)
		ops.waitPoll = func(ctx context.Context) error { cancel(); return ctx.Err() }
		h := &Harness{ops: ops}
		err := h.waitForObjectStoreReadiness(ctx, objectStoreReadinessConfig{
			ReadyTimeout: objectStoreClientReadyTimeout, AttemptTimeout: objectStoreClientAttemptTimeout,
			CreateNamespace: true, Bucket: "probe", Key: "key",
		})
		diagnostic, ok := ReadinessDiagnosticFromError(err)
		envelope := objectStoreSetupEnvelope("create_bucket", err)
		if !ok || diagnostic.Outcome != "cancelled" || diagnostic.CleanupOutcome != "completed" ||
			!errors.Is(err, context.Canceled) || !errors.Is(err, primary) ||
			envelope.FailureClass != "interrupted" || envelope.FailureReason != "cancelled_or_interrupted" {
			t.Fatalf("poll cancellation lost its cause or classification: diagnostic=%#v envelope=%#v err=%v", diagnostic, envelope, err)
		}
		client.assertNoProbeResidue(t)
	})
	client := newFakeReadinessClient()
	client.failures["put"] = -1
	ops := readinessOperations(client)
	ops.waitPoll = func(context.Context) error { return context.DeadlineExceeded }

	harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
	err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
		ReadyTimeout:    objectStoreClientReadyTimeout,
		PollInterval:    objectStoreHealthPollInterval,
		AttemptTimeout:  objectStoreClientAttemptTimeout,
		CreateNamespace: true,
		ListFirst:       true,
		Bucket:          "ct-suitehash-process1-readiness",
		Key:             ".cartulary-readiness/probe",
	})
	var readinessErr *objectStoreReadinessError
	if !errors.As(err, &readinessErr) || !readinessErr.DeadlineExpired {
		t.Fatalf("expected readiness deadline expiry, got %v", err)
	}
	if readinessErr.Attempts == 0 || readinessErr.CleanupOutcome != "completed" {
		t.Fatalf("expected same-lane polling with cleanup evidence, got %#v", readinessErr)
	}
	client.assertNoProbeResidue(t)
}

func TestObjectStoreMutationProbePreservesPrimaryFailureDuringCleanupFailure(t *testing.T) {
	t.Run("controller_retains_cleanup_failure", func(t *testing.T) {
		for _, test := range []struct {
			name, outcome, cause, class string
			primary, cleanup            error
		}{
			{"transient", "unavailable", "unknown_transient", "harness", errors.New("transient put failure"), errors.New("transient cleanup failure")},
			{"cleanup_deadline", "unavailable", "unknown_transient", "harness", errors.New("transient put failure"), context.DeadlineExceeded},
			{"deadline", "deadline_expired", "operation_timeout", "infra", context.DeadlineExceeded, &smithy.GenericAPIError{Code: "AccessDenied"}},
			{"cancelled", "cancelled", "cancelled", "interrupted", context.Canceled, &smithy.GenericAPIError{Code: "AccessDenied"}},
		} {
			t.Run(test.name, func(t *testing.T) {
				client := newFakeReadinessClient()
				client.failures["put"], client.failureErrors["put"] = 1, test.primary
				client.failures["delete_namespace"], client.failureErrors["delete_namespace"] = 1, test.cleanup
				h := &Harness{ops: readinessOperations(client)}
				err := h.waitForObjectStoreReadiness(t.Context(), objectStoreReadinessConfig{
					ReadyTimeout: objectStoreClientReadyTimeout, AttemptTimeout: objectStoreClientAttemptTimeout,
					CreateNamespace: true, Bucket: "probe", Key: "key",
				})
				diagnostic, ok := ReadinessDiagnosticFromError(err)
				envelope := objectStoreSetupEnvelope("create_bucket", err)
				if !ok || diagnostic.Outcome != test.outcome || diagnostic.CleanupOutcome != "failed" ||
					diagnostic.Stage != "put" || diagnostic.CauseCounts[test.cause] != 1 ||
					envelope.FailureClass != test.class || !errors.Is(err, test.primary) {
					t.Fatalf("cleanup replaced primary failure: diagnostic=%#v envelope=%#v err=%v", diagnostic, envelope, err)
				}
				if !client.namespaces["probe"] || client.callCount("put") != 1 {
					t.Fatal("uncertain cleanup was hidden by another readiness attempt")
				}
				if test.name == "cleanup_deadline" && diagnostic.AttemptTimeoutCount != 0 {
					t.Fatal("cleanup timeout counted as a readiness attempt timeout")
				}
			})
		}
	})
	for _, responseLost := range []bool{false, true} {
		t.Run(fmt.Sprint("response_lost=", responseLost), func(t *testing.T) {
			client := newFakeReadinessClient()
			primary := errors.New("temporary put failure")
			client.failures["put"], client.failureErrors["put"] = 1, primary
			client.failures["delete_namespace"] = 1
			client.namespaceDeleteResponseLost = responseLost
			failure := runObjectStoreMutationProbe(t.Context(), client, objectStoreReadinessConfig{CreateNamespace: true, ListFirst: true, Bucket: "probe", Key: "key"})
			if failure == nil || failure.Cause != primary || failure.Stage != "put" || failure.CleanupOutcome != "failed" || failure.CleanupErr == nil {
				t.Fatalf("lost primary or secondary failure: %#v", failure)
			}
			if client.namespaces["probe"] == responseLost || len(client.objects) != 0 {
				t.Fatal("fake concealed the actual failure state")
			}
		})
	}
	t.Run("nonempty_namespace", func(t *testing.T) {
		client := newFakeReadinessClient()
		client.namespaces["probe"] = true
		client.objects["probe/remaining"] = []byte("retain")
		if err := client.DeleteNamespace(t.Context(), "probe"); err == nil || !client.namespaces["probe"] || string(client.objects["probe/remaining"]) != "retain" {
			t.Fatal("namespace deletion erased remaining objects", err)
		}
	})
}

func TestAttachedObjectStoreUsesMutationProbeContract(t *testing.T) {
	client := newFakeReadinessClient()
	client.namespaces["ct-suite-readiness-probe"] = true
	ops := readinessOperations(client)

	harness := &Harness{ops: ops,
		suiteHash:   "suitehash",
		processHash: "process1",
		attached:    true,
		probeBucket: "ct-suite-readiness-probe",
	}
	if err := verifyAttachedHarness(context.Background(), harness); err != nil {
		t.Fatalf("verify attached object store: %v", err)
	}
	for _, stage := range []string{"list", "put", "head", "delete", "delete_verify"} {
		if client.callCount(stage) == 0 {
			t.Fatalf("attached readiness did not execute %s", stage)
		}
	}
	if client.callCount("create_namespace") != 0 || client.callCount("delete_namespace") != 0 {
		t.Fatalf("attached readiness changed its broker-owned namespace: %#v", client.calls)
	}
	if !client.namespaces["ct-suite-readiness-probe"] || len(client.objects) != 0 {
		t.Fatalf("attached readiness did not preserve a clean broker namespace: namespaces=%#v objects=%#v", client.namespaces, client.objects)
	}
}

func TestPackageBucketAdmissionProbesWithoutRemovingTheBucket(t *testing.T) {
	client := newFakeReadinessClient()
	client.namespaces["package-bucket"] = true
	ops := readinessOperations(client)

	harness := &Harness{ops: ops, suiteHash: "suitehash", processHash: "process1"}
	if err := harness.waitForObjectStoreReadiness(context.Background(), objectStoreReadinessConfig{
		ReadyTimeout:    objectStoreClientReadyTimeout,
		PollInterval:    objectStoreHealthPollInterval,
		AttemptTimeout:  objectStoreClientAttemptTimeout,
		CreateNamespace: false,
		ListFirst:       false,
		Bucket:          "package-bucket",
		Key:             ".cartulary-readiness/package-probe",
	}); err != nil {
		t.Fatalf("admit package bucket: %v", err)
	}
	if !client.namespaces["package-bucket"] {
		t.Fatal("package admission removed its borrowed package bucket")
	}
	if client.callCount("create_namespace") != 0 || client.callCount("delete_namespace") != 0 {
		t.Fatalf("package admission changed namespace ownership: %#v", client.calls)
	}
	if len(client.objects) != 0 {
		t.Fatalf("package admission left probe objects: %#v", client.objects)
	}
	if got, want := client.calls, []string{"put", "head", "delete", "delete_verify"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("package admission mutation order mismatch: got=%#v want=%#v", got, want)
	}
}

func TestObjectStoreMutationProbeRejectsSizeMismatch(t *testing.T) {
	client := newFakeReadinessClient()
	client.namespaces["package-bucket"] = true
	client.headSizeDelta = 1

	failure := runObjectStoreMutationProbe(context.Background(), client, objectStoreReadinessConfig{
		CreateNamespace: false,
		ListFirst:       false,
		Bucket:          "package-bucket",
		Key:             ".cartulary-readiness/package-probe",
	})
	if failure == nil || failure.Stage != "head" || failure.CleanupOutcome != "completed" {
		t.Fatalf("expected size mismatch with completed cleanup, got %#v", failure)
	}
	if len(client.objects) != 0 {
		t.Fatalf("size mismatch left probe objects: %#v", client.objects)
	}
}

func readinessOperations(client objectStoreReadinessClient) fixtureOperations {
	return fixtureOperations{waitPoll: func(ctx context.Context) error { return ctx.Err() }, readinessClient: func(context.Context, *Harness) (objectStoreReadinessClient, error) { return client, nil }}
}

type fakeReadinessClient struct {
	namespaceDeleteResponseLost bool
	mu                          sync.Mutex
	namespaces                  map[string]bool
	objects                     map[string][]byte
	failures                    map[string]int
	failureErrors               map[string]error
	headSizeDelta               int64
	calls                       []string
}

func newFakeReadinessClient() *fakeReadinessClient {
	return &fakeReadinessClient{
		namespaces:    make(map[string]bool),
		objects:       make(map[string][]byte),
		failures:      make(map[string]int),
		failureErrors: make(map[string]error),
	}
}

func (c *fakeReadinessClient) ListBuckets(context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.recordAndFail("list")
}

func (c *fakeReadinessClient) CreateNamespace(_ context.Context, bucket string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.recordAndFail("create_namespace"); err != nil {
		return err
	}
	c.namespaces[bucket] = true
	return nil
}

func (c *fakeReadinessClient) Put(_ context.Context, bucket string, key string, payload []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.recordAndFail("put"); err != nil {
		return err
	}
	if !c.namespaces[bucket] {
		return &smithy.GenericAPIError{Code: "NoSuchBucket"}
	}
	c.objects[bucket+"/"+key] = append([]byte(nil), payload...)
	return nil
}

func (c *fakeReadinessClient) HeadSize(_ context.Context, bucket string, key string) (int64, error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	objectKey := bucket + "/" + key
	payload, exists := c.objects[objectKey]
	stage := "head"
	if !exists {
		stage = "delete_verify"
	}
	if err := c.recordAndFail(stage); err != nil {
		return 0, err
	}
	if !exists {
		return 0, &smithy.GenericAPIError{Code: "NoSuchKey"}
	}
	return int64(len(payload)) + c.headSizeDelta, nil
}

func (c *fakeReadinessClient) Delete(_ context.Context, bucket string, key string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if err := c.recordAndFail("delete"); err != nil {
		return err
	}
	delete(c.objects, bucket+"/"+key)
	return nil
}

func (c *fakeReadinessClient) DeleteNamespace(_ context.Context, bucket string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	err := c.recordAndFail("delete_namespace")
	if err != nil && !c.namespaceDeleteResponseLost {
		return err
	}
	for key := range c.objects {
		if strings.HasPrefix(key, bucket+"/") {
			return &smithy.GenericAPIError{Code: "BucketNotEmpty"}
		}
	}
	delete(c.namespaces, bucket)
	return err
}

func (c *fakeReadinessClient) recordAndFail(stage string) error {
	c.calls = append(c.calls, stage)
	remaining := c.failures[stage]
	if remaining == 0 {
		return nil
	}
	if remaining > 0 {
		c.failures[stage] = remaining - 1
	}
	if err := c.failureErrors[stage]; err != nil {
		return err
	}
	return errors.New("temporary object-store unavailability")
}

func (c *fakeReadinessClient) callCount(stage string) int {
	c.mu.Lock()
	defer c.mu.Unlock()
	count := 0
	for _, call := range c.calls {
		if call == stage {
			count++
		}
	}
	return count
}

func (c *fakeReadinessClient) assertNoProbeResidue(t testing.TB) {
	t.Helper()
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.namespaces) != 0 || len(c.objects) != 0 {
		t.Fatalf("object-store probe residue: namespaces=%#v objects=%#v", c.namespaces, c.objects)
	}
}
