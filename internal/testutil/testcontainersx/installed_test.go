package testcontainersx

import (
	"context"
	"errors"
	"reflect"
	"testing"

	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

func TestInstalledOnlyRejectsImageAcquisitionWithoutCallingDocker(t *testing.T) {
	t.Run("final_create_observation_preserves_provider_request_and_result", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		options := client.ContainerCreateOptions{Config: &dockercontainer.Config{Image: "fixture"}, HostConfig: &dockercontainer.HostConfig{PublishAllPorts: true, CgroupParent: "owned-parent"}}
		failure := errors.New("provider failure")
		order := []string{}
		delegate := createTestClient{create: func(gotContext context.Context, got client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
			order = append(order, "delegate")
			if gotContext != ctx || !reflect.DeepEqual(got, options) {
				t.Fatal("request changed")
			}
			return client.ContainerCreateResult{ID: "partial-owned-container"}, failure
		}}
		api := observedClient{APIClient: delegate, observe: func(got client.ContainerCreateOptions) {
			order = append(order, "observation")
			if !reflect.DeepEqual(got, options) {
				t.Fatal("observation changed request")
			}
		}}
		result, err := api.ContainerCreate(ctx, options)
		if result.ID != "partial-owned-container" || err != failure || !reflect.DeepEqual(order, []string{"observation", "delegate"}) {
			t.Fatal(result, err, order)
		}
	})
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

type createTestClient struct {
	client.APIClient
	create func(context.Context, client.ContainerCreateOptions) (client.ContainerCreateResult, error)
}

func (c createTestClient) ContainerCreate(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
	return c.create(ctx, options)
}
