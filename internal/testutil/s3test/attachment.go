package s3test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"strings"
	"sync"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

// One process binding, including its terminal admission result. Tests use their
// own instances. A changed binding requires a different process, not replacement.
var processAdmission sharedAdmission

type sharedAdmission struct {
	mu      sync.Mutex
	binding fixtureBinding
	done    chan struct{}
	harness *Harness
	err     error
	verify  func(context.Context, *Harness) error
}

type fixtureBinding struct{ inputs map[string]string }

var fixtureInputKeys = [...]string{
	suiteservices.S3EndpointEnv, suiteservices.S3AccessKeyEnv, suiteservices.S3SecretKeyEnv,
	suiteservices.S3SecureEnv, suiteservices.S3ProbeBucketEnv, "SSL_CERT_FILE",
	suiteservices.HarnessServiceDependenciesEnv, suiteservices.SuiteIDEnv, suiteservices.TargetEnv,
	suiteservices.SuiteRuntimeRootEnv, suiteservices.SuiteRuntimeLeaseIDEnv, suiteservices.SuiteRuntimeRunIDEnv,
	suiteservices.CallModeEnv, suiteservices.LifecycleModeEnv, suiteservices.PersistentBorrowerEnv,
	"CARTULARY_TEST_RESULTS_DIR", "CARTULARY_TEST_RUN_ID",
}

func processFixtureInputs() map[string]string {
	inputs := make(map[string]string)
	for _, key := range fixtureInputKeys {
		if value, ok := os.LookupEnv(key); ok {
			inputs[key] = value
		}
	}
	return inputs
}

func bindFixture(inputs map[string]string) fixtureBinding {
	copied := make(map[string]string)
	for _, key := range fixtureInputKeys {
		if value, ok := inputs[key]; ok {
			copied[key] = value
		}
	}
	return fixtureBinding{inputs: copied}
}

func (b fixtureBinding) environment() map[string]string {
	if b.inputs == nil {
		return map[string]string{}
	}
	return b.inputs
}

func StartShared(ctx context.Context) (*Harness, error) { return processAdmission.start(ctx) }

func (a *sharedAdmission) start(ctx context.Context) (*Harness, error) {
	return a.startWithInputs(ctx, processFixtureInputs())
}

func (a *sharedAdmission) startWithInputs(ctx context.Context, inputs map[string]string) (*Harness, error) {
	if ctx == nil {
		return nil, errors.New("attach object-store harness: context is required")
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	binding := bindFixture(inputs)
	a.mu.Lock()
	if a.done != nil {
		if !maps.Equal(binding.inputs, a.binding.inputs) {
			a.mu.Unlock()
			return nil, errors.New("attach object-store harness: admitted binding changed")
		}
		done := a.done
		a.mu.Unlock()
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-done:
		}
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		return a.harness, a.err
	}
	a.binding, a.done = binding, make(chan struct{})
	a.mu.Unlock()
	h, err := a.admit(ctx, binding)
	a.mu.Lock()
	a.harness, a.err = h, err
	close(a.done)
	a.mu.Unlock()
	return h, err
}

func (a *sharedAdmission) admit(ctx context.Context, binding fixtureBinding) (*Harness, error) {
	if err := suiteservices.CheckServiceDependencies(binding.environment(), "object_store"); err != nil {
		return nil, err
	}
	h, err := harnessFromBinding(binding)
	if err != nil {
		return nil, err
	}
	if h.probeBucket == "" {
		return nil, errors.New("attach object-store harness: broker probe namespace is required")
	}
	verify := a.verify
	if verify == nil {
		verify = verifyAttachedHarness
	}
	err = verify(ctx, h)
	err = errors.Join(err, ctx.Err())
	if err == nil {
		err = h.recordEvent(suiteservices.Event{Type: suiteservices.EventS3Attach})
	}
	err = errors.Join(err, ctx.Err())
	if err != nil {
		return nil, errors.Join(fmt.Errorf("attach object-store harness: admission: %w", err), h.Close(context.Background()))
	}
	return h, nil
}

func harnessFromBinding(binding fixtureBinding) (*Harness, error) {
	env := binding.environment()
	get := func(key string) string { return strings.TrimSpace(env[key]) }
	if get(suiteservices.S3EndpointEnv) == "" || get(suiteservices.S3AccessKeyEnv) == "" || get(suiteservices.S3SecretKeyEnv) == "" {
		return nil, errors.New("object-store requires a complete explicit service binding")
	}
	secure, err := suiteservices.ParseBool(get(suiteservices.S3SecureEnv))
	if err != nil || !secure {
		return nil, errors.New("object-store binding requires verified TLS")
	}
	if get("SSL_CERT_FILE") == "" {
		return nil, errors.New("object-store binding requires an explicit fixture CA")
	}
	return &Harness{
		binding:             binding,
		Endpoint:            get(suiteservices.S3EndpointEnv),
		AccessKey:           get(suiteservices.S3AccessKeyEnv),
		SecretKey:           get(suiteservices.S3SecretKeyEnv),
		Secure:              true,
		RootCertificateFile: get("SSL_CERT_FILE"),
		suiteHash:           suiteservices.SuiteHash(env),
		processHash:         suiteservices.ProcessHash(),
		attached:            true,
		probeBucket:         get(suiteservices.S3ProbeBucketEnv),
	}, nil
}

// CleanupBucketFromBinding uses saved recovery inputs without admission or
// provisioning. The caller must first prove ownership of the exact bucket.
func CleanupBucketFromBinding(ctx context.Context, inputs map[string]string, bucket string) error {
	h, err := harnessFromBinding(bindFixture(inputs))
	if err != nil {
		return err
	}
	err = h.CleanupBucket(ctx, bucket)
	return errors.Join(err, h.Close(ctx))
}
