//go:build linux

package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/startupproof"
)

// The inherited pipe is a launch capability, not an endpoint discovered by PID.
// The broker writes admission only after recording the exact helper group.
func openStartup(ctx context.Context, env map[string]string, lease string, services []string) (context.Context, *startupproof.Journal, func(), error) {
	value := env[startupproof.ControlEnv]
	if value == "" {
		return ctx, nil, func() {}, nil
	} // Non-broker standalone owner.
	fd, err := strconv.Atoi(value)
	if err != nil || fd < 3 {
		return nil, nil, nil, errors.New("invalid startup control capability")
	}
	control, err := startupproof.OpenControl(fd)
	if err != nil {
		return nil, nil, nil, err
	}
	fail := func(err error) (context.Context, *startupproof.Journal, func(), error) {
		control.Close()
		return nil, nil, nil, err
	}
	var admitted [1]byte
	if _, err := io.ReadFull(control, admitted[:]); err != nil || admitted[0] != 1 {
		return fail(errors.New("startup owner unavailable"))
	}
	repository, err := os.Getwd()
	if err != nil {
		return fail(err)
	}
	resultRoot, err := suiteservices.ResolveResultsRoot(env)
	if err != nil {
		return fail(err)
	}
	runID := suiteservices.ResolveRunID(env)
	runtime, err := harnesssession.BorrowRuntime(repository, filepath.Join(resultRoot, runID), env[suiteservices.SuiteRuntimeRootEnv], env[suiteservices.SuiteRuntimeRunIDEnv], env[suiteservices.SuiteRuntimeLeaseIDEnv])
	if err != nil {
		return fail(err)
	}
	defer runtime.ReleaseHandles()
	resources, err := runtime.Resources(4096)
	if err != nil {
		return fail(err)
	}
	var dependencies []string
	found := false
	for _, r := range resources {
		path, e := r.Path()
		if e == nil && r.Kind == "managed_suite" && path == lease && r.State == "pending" {
			dependencies = append([]string{}, r.Dependencies...)
			found = true
		}
	}
	if !found {
		return fail(errors.New("startup intent unavailable"))
	}
	journal, err := startupproof.Open(filepath.Join(filepath.Dir(lease), startupproof.Filename), startupproof.Binding{RuntimeRoot: runtime.Root, RuntimeLease: runtime.LeaseID, RunRoot: runtime.RunRoot, RunID: runID, SuiteID: suiteservices.SuiteID(env), Services: services, Dependencies: dependencies})
	if err != nil {
		return fail(err)
	}
	proof, err := harnesssession.CaptureProcess(os.Getpid())
	if err != nil {
		return fail(err)
	}
	data, err := json.Marshal(proof)
	if err != nil {
		return fail(err)
	}
	if err := journal.Register(string(data)); err != nil {
		return fail(err)
	}
	startupCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	go func() { defer close(done); var b [1]byte; _, _ = control.Read(b[:]); cancel() }()
	return startupCtx, journal, func() { control.Close(); cancel(); <-done }, nil
}
