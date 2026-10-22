package testcontainersx

import (
	"context"
	"errors"
	"os"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessaccount"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

// installedClient enforces the preparation policy at the image acquisition
// boundary, including a race with image removal after the readiness check.
type installedClient struct{ client.APIClient }

func (installedClient) ImagePull(context.Context, string, client.ImagePullOptions) (client.ImagePullResponse, error) {
	return nil, errors.New("installed-only preparation forbids image acquisition; run make test-service-images")
}

// GenericContainer preserves normal test provisioning. Installed-only review
// uses the same provider with image acquisition disabled; the caller retains
// the normal ownership and cleanup of any partially created container.
func GenericContainer(ctx context.Context, req testcontainers.GenericContainerRequest) (testcontainers.Container, error) {
	parent, err := harnessaccount.ContainerParent()
	if err != nil {
		return nil, err
	}
	if parent != "" {
		if req.Reuse || os.Getenv("TESTCONTAINERS_RYUK_DISABLED") != "true" {
			return nil, errors.New("invocation accounting requires owned containers and suite-managed cleanup")
		}
		modify := req.HostConfigModifier
		req.HostConfigModifier = func(config *dockercontainer.HostConfig) {
			if modify != nil {
				modify(config)
			}
			config.CgroupParent = parent
		}
		req.LifecycleHooks = append(req.LifecycleHooks, testcontainers.ContainerLifecycleHooks{PostStarts: []testcontainers.ContainerHook{func(_ context.Context, container testcontainers.Container) error {
			return harnessaccount.VerifyContainer(parent, container.GetContainerID())
		}}})
	}
	if os.Getenv("CARTULARY_PREPARATION_POLICY") != "installed_only" {
		return testcontainers.GenericContainer(ctx, req)
	}
	if req.Reuse || req.Image == "" || os.Getenv("TESTCONTAINERS_RYUK_DISABLED") != "true" {
		return nil, errors.New("installed-only preparation requires an owned image container with suite-managed cleanup")
	}
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return nil, err
	}
	defer provider.Close()
	provider.SetClient(installedClient{provider.Client()})
	container, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err == nil && req.Started && !container.IsRunning() {
		err = container.Start(ctx)
	}
	return container, err
}
