//go:build linux

package harnessreview

import (
	"context"
	"errors"
	"maps"
	"net/http"
	"net/url"
	"slices"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

type reviewInput struct {
	Mode          string  `json:"mode"`
	Origin        *string `json:"origin"`
	Profile       *string `json:"profile"`
	Policy        string  `json:"preparation_policy"`
	Presentation  string  `json:"presentation"`
	VerifySamples bool    `json:"verify_samples"`
}
type productionDriver struct {
	mu                     sync.Mutex
	root, runRoot, session string
	input                  reviewInput
	profile                map[string]any
	runtime                *harnesssession.Runtime
	environment            map[string]string
	adapter                *adapter
	admission              *harnessadmission.Lease
	fixtures               *harnessfixtures.Production
	onLost                 func(error)
	isReady                func() bool
	facts                  *adapterFacts
	cleanup                []cleanupStep
}

type cleanupStep struct {
	operation string
	err       error
}

func (d *productionDriver) worker() *adapter { d.mu.Lock(); defer d.mu.Unlock(); return d.adapter }
func (d *productionDriver) Prepare(ctx context.Context) (*int, error) {
	a, err := startAdapter(d.root, d.session, d.runtime, d.environment, d.onLost)
	if err != nil {
		return nil, normalized(err, "tool_configuration")
	}
	d.mu.Lock()
	d.adapter = a
	d.mu.Unlock()
	reply, err := a.call(ctx, "initialize", map[string]any{"input": d.input, "profile": d.profile, "repository": d.root, "run_root": d.runRoot, "runtime_root": d.runtime.Root})
	if err != nil {
		return nil, err
	}
	if reply.Facts == nil {
		return nil, failure("invalid_artifact")
	}
	d.facts = reply.Facts
	if d.input.Mode == "artifacts" {
		return nil, nil
	}
	var parent *string
	if v := d.environment["CARTULARY_HOST_ADMISSION_LEASE"]; v != "" {
		parent = &v
	}
	admissionCtx, cancel := context.WithTimeout(ctx, time.Duration(LimitLock)*time.Millisecond)
	d.admission, err = harnessadmission.Acquire(admissionCtx, "", harnessadmission.Quantities{"browser_stack": 1}, reply.Facts.Capacities, parent, false)
	cancel()
	if err != nil {
		return nil, normalized(err, "capacity_exceeded")
	}
	if err = d.admission.Bind(a.cmd.Process.Pid); err != nil {
		return nil, failure("capacity_exceeded")
	}
	if d.input.Mode == "seeded" {
		privateRoot := d.runRoot
		if d.input.Presentation != "design" {
			privateRoot, err = d.runtime.PrivatePath("lifecycle", d.runtime.RunID)
		}
		if err != nil {
			return nil, err
		}
		if err = publishSource(privateRoot, reply.Facts, *d.input.Profile); err != nil {
			return nil, err
		}
		reply, err = a.call(ctx, "build", map[string]any{"admission_token": d.admission.Token})
		if err != nil {
			return nil, err
		}
		if reply.Facts == nil {
			return nil, failure("invalid_artifact")
		}
		environment := maps.Clone(reply.Facts.Environment)
		environment["CARTULARY_HOST_ADMISSION_LEASE"] = d.admission.Token
		d.fixtures = harnessfixtures.NewProduction(harnessfixtures.ProductionOptions{Root: d.root, RunRoot: privateRoot, Target: "ui-review", Runtime: d.runtime, Environment: environment}, harnessfixtures.Options{})
		lease, err := d.fixtures.Broker.Acquire(ctx, harnessfixtures.Request{Capability: "browser_stack", UnitID: "ui-review", Affinity: "design-review-" + *d.input.Profile, BrowserStage: "webserver-backed", RuntimeProfile: *d.input.Profile})
		if err != nil {
			return nil, preparationError(err, "service_acquisition", "browser_stack")
		}
		maps.Copy(environment, lease.Environment())
		reply, err = a.call(ctx, "seed", map[string]any{"environment": environment})
		if err != nil {
			return nil, err
		}
		if reply.Facts == nil || reply.Facts.Review == nil {
			return nil, failure("invalid_artifact")
		}
		d.facts = reply.Facts
	} else if err = waitOrigin(ctx, *d.input.Origin); err != nil {
		return nil, err
	}
	if d.input.Presentation == "design" {
		go d.health(ctx, a)
		return nil, nil
	}
	reply, err = a.call(ctx, "prepare_browser", map[string]any{"admission_token": d.admission.Token})
	if err != nil {
		return nil, err
	}
	if d.input.Mode == "seeded" {
		go d.health(ctx, a)
	}
	return reply.Outcome.Epoch, nil
}
func preparationError(err error, phase, subject string) error {
	var classified interface{ HarnessFailure() (string, string) }
	if errors.As(err, &classified) {
		kind, reason := classified.HarnessFailure()
		for _, code := range slices.Sorted(maps.Keys(failurePolicy)) {
			p := failurePolicy[code]
			if p.Class == kind && p.Reason == reason {
				return contextual(code, phase, subject, "child_failed", "inspect_failure")
			}
		}
	}
	return contextual("preparation_failed", phase, subject, "unknown", "inspect_failure")
}
func (d *productionDriver) health(ctx context.Context, a *adapter) {
	timer := time.NewTicker(3 * time.Second)
	defer timer.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
			if d.isReady != nil && !d.isReady() {
				continue
			}
			check, done := context.WithTimeout(ctx, 10*time.Second)
			_, err := a.call(check, "check", nil)
			done()
			if err != nil && ctx.Err() == nil {
				d.onLost(failure("session_lost"))
				return
			}
		}
	}
}
func (d *productionDriver) Validate(ctx context.Context, command string, request jsonRequest) error {
	a := d.worker()
	if a == nil {
		return failure("session_mismatch")
	}
	_, err := a.call(ctx, "validate", map[string]any{"command": command, "request": request})
	return err
}
func (d *productionDriver) Execute(ctx context.Context, command string, request jsonRequest, bundle *string, id int) (Operation, error) {
	a := d.worker()
	if a == nil {
		return Operation{}, failure("session_mismatch")
	}
	reply, err := a.call(ctx, "execute", map[string]any{"command": command, "request": request, "bundle_id": bundle, "operation_id": id})
	o := reply.Outcome
	// Preserve all normalized producer failures, including secondary cleanup.
	if len(o.Failures) > 0 {
		err = nil
	}
	return Operation{Epoch: o.Epoch, Counts: o.Counts, Bundle: o.Bundle, Private: o.Private, Failures: o.Failures, Exit: o.Exit, Stop: o.Stop}, err
}
func (d *productionDriver) Interrupt() {
	if a := d.worker(); a != nil {
		a.interrupt()
	}
}
func (d *productionDriver) Close(ctx context.Context) []error {
	failures := []error{}
	attempt := func(operation string, fn func(context.Context) error, maximum time.Duration) {
		step, done := context.WithTimeout(ctx, maximum)
		defer done()
		err := fn(step)
		d.cleanup = append(d.cleanup, cleanupStep{operation, err})
		if err != nil {
			failures = append(failures, normalized(err, "cleanup_failed"))
		}
	}
	if a := d.worker(); a != nil {
		attempt("provider_close", func(ctx context.Context) error { _, err := a.call(ctx, "close_browser", nil); return err }, 30*time.Second)
		attempt("provider_close", func(ctx context.Context) error {
			if err := a.drain(ctx); err != nil {
				a.kill()
				<-a.done
				return err
			}
			_, err := a.call(ctx, "close_store", nil)
			return err
		}, 30*time.Second)
		attempt("provider_close", a.close, 30*time.Second)
	}
	// Every independent process/renderer stop is attempted. A surviving dependent
	// blocks fixture teardown, so its service cannot disappear underneath it.
	dependentErrors := recoverIndependent(ctx, d.root, d.runRoot, d.runtime, d.environment)
	d.cleanup = append(d.cleanup, cleanupStep{"provider_close", errors.Join(dependentErrors...)})
	failures = append(failures, dependentErrors...)
	if len(dependentErrors) == 0 {
		if d.fixtures != nil {
			attempt("services_close", d.fixtures.Close, 125*time.Second)
		}
		recovery := harnessfixtures.ProductionOptions{Root: d.root, RunRoot: d.runRoot, Target: "ui-review", Runtime: d.runtime, Environment: d.environment}
		attempt("services_close", recovery.Recover, 125*time.Second)
	}
	if d.admission != nil {
		attempt("host_release", d.admission.Release, 5*time.Second)
	}
	return failures
}
func waitOrigin(ctx context.Context, origin string) error {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme != "http" || (parsed.Hostname() != "127.0.0.1" && parsed.Hostname() != "::1") {
		return failure("invalid_request")
	}
	client := &http.Client{Timeout: time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	for {
		request, _ := http.NewRequestWithContext(ctx, "GET", origin, nil)
		response, err := client.Do(request)
		if err == nil {
			response.Body.Close()
			if response.StatusCode >= 300 && response.StatusCode < 400 {
				redirect, e := response.Location()
				if e != nil || redirect.Scheme != parsed.Scheme || redirect.Host != parsed.Host {
					return failure("navigation_boundary")
				}
			}
			if response.StatusCode < 500 {
				return nil
			}
		}
		timer := time.NewTimer(250 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return failure("readiness_expired")
		case <-timer.C:
		}
	}
}
func recoverIndependent(ctx context.Context, root, runRoot string, runtime *harnesssession.Runtime, environment map[string]string) []error {
	resources, err := runtime.Resources(4096)
	if err != nil {
		return []error{err}
	}
	failures := []error{}
	for _, kind := range []string{"diagnostic_scope", "preparation_process", "browser_process", "helper_process", "browser_renderer"} {
		for _, resource := range resources {
			if resource.Kind != kind {
				continue
			}
			step, cancel := context.WithTimeout(ctx, 30*time.Second)
			err := runtime.StopResource(step, root, runRoot, environment, resource)
			cancel()
			if err == nil {
				err = runtime.RecordResource(resource.Kind, resource.Target, "released")
			}
			if err != nil {
				failures = append(failures, normalized(err, "cleanup_failed"))
			}
		}
	}
	return failures
}
