package s3test

import (
	"context"

	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type fixtureOperations struct {
	waitPoll        func(context.Context) error
	readinessClient func(context.Context, *Harness) (objectStoreReadinessClient, error)
	createPackage   func(context.Context, *Harness, string, fixtureAttribution) error
	cleanupPackage  func(context.Context, *Harness, string, fixtureAttribution) error
	readyPackage    func(context.Context, *Harness, string) error
}

func (h *Harness) operations() fixtureOperations {
	ops := h.ops
	if ops.readinessClient == nil {
		ops.readinessClient = func(ctx context.Context, h *Harness) (objectStoreReadinessClient, error) {
			client, err := h.Client(ctx)
			if err != nil {
				return nil, err
			}
			return s3ReadinessClient{client: client}, nil
		}
	}
	if ops.createPackage == nil {
		ops.createPackage = func(ctx context.Context, h *Harness, name string, attribution fixtureAttribution) error {
			return h.createOwnedBucket(ctx, name, suiteservices.FixtureReuseObjectStorePackage, attribution)
		}
	}
	if ops.cleanupPackage == nil {
		ops.cleanupPackage = func(ctx context.Context, h *Harness, bucket string, attribution fixtureAttribution) error {
			return h.cleanupPrefixWithDetails(ctx, bucket, "", suiteservices.FixtureReusePrefix, attribution)
		}
	}
	if ops.readyPackage == nil {
		ops.readyPackage = func(ctx context.Context, h *Harness, bucket string) error {
			return h.waitForObjectStoreReadiness(ctx, objectStoreReadinessConfig{
				Phase: "package_admission", ReadyTimeout: objectStoreClientReadyTimeout,
				PollInterval: objectStoreHealthPollInterval, AttemptTimeout: objectStoreClientAttemptTimeout,
				Bucket: bucket, Key: h.nextProbeKey(),
			})
		}
	}
	return ops
}
