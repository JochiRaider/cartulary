package testcontainersx

import (
	"context"
	"testing"

	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestInstalledOnlyRejectsImageAcquisitionWithoutCallingDocker(t *testing.T) {
	// A nil delegate would panic if image acquisition reached the Docker client.
	if _, err := (installedClient{}).ImagePull(context.Background(), "missing", client.ImagePullOptions{}); err == nil {
		t.Fatal("image acquisition was allowed")
	}
	t.Setenv("CARTULARY_PREPARATION_POLICY", "installed_only")
	t.Setenv("TESTCONTAINERS_RYUK_DISABLED", "false")
	if _, err := GenericContainer(context.Background(), testcontainers.GenericContainerRequest{ContainerRequest: testcontainers.ContainerRequest{Image: "missing"}}); err == nil {
		t.Fatal("unmanaged cleanup was allowed")
	}
}
