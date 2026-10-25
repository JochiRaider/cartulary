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
		api := &observedClient{APIClient: delegate, observe: func(got client.ContainerCreateOptions) {
			order = append(order, "observation")
			if !reflect.DeepEqual(got, options) {
				t.Fatal("observation changed request")
			}
		}}
		if err := api.beginObservation(ctx, testcontainers.ContainerRequest{}); err != nil {
			t.Fatal(err)
		}
		result, err := api.ContainerCreate(ctx, options)
		if result.ID != "partial-owned-container" || err != failure || !reflect.DeepEqual(order, []string{"observation", "delegate"}) {
			t.Fatal(result, err, order)
		}
	})
	t.Run("auxiliary_creates_do_not_own_target_observation", func(t *testing.T) {
		ctx := t.Context()
		target := client.ContainerCreateOptions{Config: &dockercontainer.Config{Image: "fixture"}, HostConfig: &dockercontainer.HostConfig{PublishAllPorts: false}}
		// Equal values are not request identity. An auxiliary create must remain
		// outside the observation even if its configuration happens to match.
		auxiliaryConfig := *target.Config
		auxiliary := target
		auxiliary.Config = &auxiliaryConfig
		delegated, observed := 0, 0
		api := &observedClient{APIClient: createTestClient{create: func(gotContext context.Context, got client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
			if gotContext != ctx || (got.Config != target.Config && got.Config != auxiliary.Config) || got.HostConfig != target.HostConfig {
				t.Fatal("delegated request changed")
			}
			delegated++
			return client.ContainerCreateResult{ID: "provider-result"}, nil
		}}, observe: func(got client.ContainerCreateOptions) {
			if got.Config != target.Config {
				t.Fatal("auxiliary request reached target observer")
			}
			observed++
		}}
		create := func(options client.ContainerCreateOptions) {
			t.Helper()
			result, err := api.ContainerCreate(ctx, options)
			if err != nil || result.ID != "provider-result" {
				t.Fatal("provider result changed", result, err)
			}
		}
		create(auxiliary)
		if err := api.beginObservation(ctx, testcontainers.ContainerRequest{}); err != nil {
			t.Fatal(err)
		}
		create(target)
		create(auxiliary)
		create(target)
		if delegated != 4 || observed != 2 {
			t.Fatalf("delegated=%d observed=%d; target repeats must remain visible", delegated, observed)
		}
		api.endObservation()
		create(target)
		if delegated != 5 || observed != 2 || api.targetConfig != nil || api.observe != nil {
			t.Fatal("observation lifetime retained target state")
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
