//go:build cartulary_harness

package server

import (
	"context"
	"net/http"

	"github.com/JochiRaider/cartulary/internal/modules/auth"
	networkflowharnesscontrol "github.com/JochiRaider/cartulary/internal/modules/networkflow/harnesscontrol"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/harnessruntime"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/httpruntime"
)

type harnessServerProfile struct{}

func newServerProfile() serverProfile {
	return harnessServerProfile{}
}

func (harnessServerProfile) validateEnvironment(func(string) (string, bool)) error {
	return nil
}

func (harnessServerProfile) runtimeOptions(lookup func(string) (string, bool)) Options {
	enabled, _ := lookup("CARTULARY_ENABLE_TEST_ROUTES")
	if enabled != "1" {
		return Options{}
	}

	testClock := httpapi.NewTestClock()
	harnessControls := harnessruntime.NewControls()
	networkFlowControls := networkflowharnesscontrol.NewControls()
	return Options{
		Now: testClock.Now,
		HTTP: httpapi.Options{
			Dependencies: httpapi.DependencySet{PublicErrorFaults: harnessControls.PublicErrorFaults},
			AdditionalRoutes: append(
				harnessruntime.RegisterRoutes(harnessControls, testClock, networkFlowControls.Contribution()),
				auth.RegisterTestRoutes(),
				savedviews.RegisterTestRoutes(),
				timeline.RegisterTestRoutes(),
			),
		},
	}
}

func (harnessServerProfile) inheritedListenerFD(lookup func(string) (string, bool)) string {
	listenerFD, _ := lookup("CARTULARY_HTTP_LISTEN_FD")
	return listenerFD
}

func (harnessServerProfile) serve(ctx context.Context, handler http.Handler, options httpruntime.Options) error {
	return httpruntime.Serve(ctx, handler, options)
}
