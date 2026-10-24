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
	return GenericContainerObserved(ctx, req, nil)
}

// CreateObserver observes final provider-normalized options without modifying
// them. It does not own the request, the returned container, or its cleanup.
type CreateObserver func(client.ContainerCreateOptions)

type observedClient struct {
	client.APIClient
	observe CreateObserver
}

func (c observedClient) ContainerCreate(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
	c.observe(options)
	return c.APIClient.ContainerCreate(ctx, options)
}

// GenericContainerObserved uses the existing provider/client boundary to observe
// publication after Testcontainers has normalized ports. Observation never
// introduces another allocation attempt or changes the provider's result.
func GenericContainerObserved(ctx context.Context, req testcontainers.GenericContainerRequest, observe CreateObserver) (testcontainers.Container, error) {
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
	installedOnly := os.Getenv("CARTULARY_PREPARATION_POLICY") == "installed_only"
	if !installedOnly && observe == nil {
		return testcontainers.GenericContainer(ctx, req)
	}
	if req.Reuse || req.Image == "" || (req.ProviderType != testcontainers.ProviderDefault && req.ProviderType != testcontainers.ProviderDocker) {
		return nil, errors.New("observed or installed-only preparation requires an owned Docker image container")
	}
	// Observation preserves ordinary Testcontainers cleanup policy. Only the
	// installed-only and accounted lifetimes require the suite's exact finalizer
	// instead of the auxiliary reaper; accounting is checked above.
	if installedOnly && os.Getenv("TESTCONTAINERS_RYUK_DISABLED") != "true" {
		return nil, errors.New("installed-only preparation requires suite-managed cleanup")
	}
	providerOptions := []testcontainers.DockerProviderOption{}
	if req.Logger != nil {
		providerOptions = append(providerOptions, testcontainers.WithLogger(req.Logger))
	}
	provider, err := testcontainers.NewDockerProvider(providerOptions...)
	if err != nil {
		return nil, err
	}
	defer provider.Close()
	var api client.APIClient = provider.Client()
	if installedOnly {
		api = installedClient{api}
	}
	if observe != nil {
		api = observedClient{api, observe}
	}
	provider.SetClient(api)
	container, err := provider.CreateContainer(ctx, req.ContainerRequest)
	if err == nil && req.Started && !container.IsRunning() {
		err = container.Start(ctx)
	}
	return container, err
}
