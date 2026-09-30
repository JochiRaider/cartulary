package testcontainersx

import (
	"context"
	"errors"
	"os"

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
