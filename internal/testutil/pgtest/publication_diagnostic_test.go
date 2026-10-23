package pgtest

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"github.com/containerd/errdefs"
	dockercontainer "github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
)

func testPublicationDiagnostic(t *testing.T) {
	t.Run("SDK_removal_status", func(t *testing.T) {
		for _, status := range []int{http.StatusNotFound, http.StatusConflict, http.StatusInternalServerError} {
			t.Run(fmt.Sprint(status), func(t *testing.T) {
				server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					if r.URL.Path == "/_ping" {
						w.WriteHeader(http.StatusOK)
						return
					}
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(status)
					_, _ = w.Write([]byte(`{"message":"fixture status"}`))
				}))
				defer server.Close()
				api, err := client.New(client.WithHost(server.URL))
				if err != nil {
					t.Fatal(err)
				}
				defer api.Close()
				_, err = api.ContainerInspect(context.Background(), "exact-owned-container", client.ContainerInspectOptions{})
				if errdefs.IsNotFound(err) != (status == http.StatusNotFound) {
					t.Fatalf("status %d classified as %T", status, err)
				}
			})
		}
	})
	port := network.MustParsePort("5432/tcp")
	request := func() client.ContainerCreateOptions {
		return client.ContainerCreateOptions{Config: &dockercontainer.Config{ExposedPorts: network.PortSet{port: {}}}, HostConfig: &dockercontainer.HostConfig{PortBindings: network.PortMap{port: {{HostIP: netip.MustParseAddr("127.0.0.1"), HostPort: "45678"}, {HostIP: netip.MustParseAddr("::")}}}}}
	}
	t.Run("exact_bounded_redacted_request", func(t *testing.T) {
		d := newStartupDiagnostic(t.Context())
		if d.Publication.Status != "not_observed" || d.Publication.PublishAllPorts != nil {
			t.Fatal(d.Publication)
		}
		options := request()
		before, _ := json.Marshal(options)
		observe := d.publicationObserver([]string{"5432/tcp", "5432/tcp"})
		observe(options)
		after, _ := json.Marshal(options)
		o := d.Publication
		if string(before) != string(after) || o.Status != "observed" || o.PublishAllPorts == nil || *o.PublishAllPorts || !reflect.DeepEqual(o.RequestedPorts, []string{"5432/tcp"}) || !reflect.DeepEqual(o.EffectivePorts, o.RequestedPorts) || len(o.Bindings) != 2 || o.Bindings[0].AddressScope != "loopback" || o.Bindings[0].PortKind != "fixed" || o.Bindings[1].PortKind != "ephemeral" {
			t.Fatal(o)
		}
		d.finish("container_start", nil)
		d.CleanupOutcome = "not_needed"
		payload, _ := json.Marshal(d.PostgresReadinessDiagnostic)
		for _, secret := range []string{"127.0.0.1", "45678", "::"} {
			if strings.Contains(string(payload), secret) {
				t.Fatal("endpoint retained", secret)
			}
		}
		value, err := harnessschema.Decode(payload)
		if err != nil {
			t.Fatal(err)
		}
		if err := harnessschema.Validate(d.SchemaID, value); err != nil {
			t.Fatal(err)
		}
		observe(options)
		if d.Publication.Status != "unavailable" || len(d.Publication.Bindings) != 0 || d.Publication.PublishAllPorts != nil {
			t.Fatal("repeated observation concealed", d.Publication)
		}
	})
	for _, name := range []string{"missing_config", "missing_host", "unknown_port", "too_many_ports", "too_many_bindings", "invalid_host_port", "invalid_authored_port"} {
		t.Run(name, func(t *testing.T) {
			d := newStartupDiagnostic(t.Context())
			options, authored := request(), []string{"5432/tcp"}
			switch name {
			case "missing_config":
				options.Config = nil
			case "missing_host":
				options.HostConfig = nil
			case "unknown_port":
				options.Config.ExposedPorts[network.Port{}] = struct{}{}
			case "too_many_ports":
				for i := uint16(1); i <= 17; i++ {
					p, _ := network.PortFrom(i, network.TCP)
					options.Config.ExposedPorts[p] = struct{}{}
				}
			case "too_many_bindings":
				options.HostConfig.PortBindings[port] = make([]network.PortBinding, 33)
			case "invalid_host_port":
				options.HostConfig.PortBindings[port][0].HostPort = "secret"
			case "invalid_authored_port":
				authored[0] = "sensitive:5432"
			}
			d.publicationObserver(authored)(options)
			if !reflect.DeepEqual(d.Publication, emptyPublication("unavailable")) {
				t.Fatal(d.Publication)
			}
		})
	}
}
