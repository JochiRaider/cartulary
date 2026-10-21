package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/startupproof"
)

type partialSuiteKey struct{}

// Both providers publish into one acquisition receipt. Serialization prevents
// a later callback from replacing the sibling's already acquired authority.
// This object belongs to one start-suite invocation, never to process globals.
type partialSuiteOwnership struct {
	mu       sync.Mutex
	env      map[string]string
	path     string
	postgres postgresService
	store    objectStoreService
	journal  *startupproof.Journal
}

func publishPartialSuite(ctx context.Context, postgres postgresService, store objectStoreService) error {
	pending, ok := ctx.Value(partialSuiteKey{}).(*partialSuiteOwnership)
	if !ok {
		return nil // Standalone owners retain their returned handles.
	}
	pending.mu.Lock()
	defer pending.mu.Unlock()
	if pending.journal != nil {
		if postgres.containerID != "" {
			if err := pending.journal.Acquired(suiteservices.ServicePostgres, postgres.containerID); err != nil {
				return err
			}
		}
		if store.containerID != "" {
			if err := pending.journal.Acquired(suiteservices.ServiceObjectStore, store.containerID); err != nil {
				return err
			}
		}
	}
	if postgres.containerID != "" {
		pending.postgres = postgres
	}
	if store.containerID != "" {
		pending.store = store
	}
	lease, err := writeServiceLease(pending.env, pending.postgres, pending.store)
	if err != nil {
		return err
	}
	if err := copyFile(lease, pending.path, 0o600); err != nil {
		return err
	}
	repository, err := os.Getwd()
	if err != nil {
		return err
	}
	resultRoot, err := suiteservices.ResolveResultsRoot(pending.env)
	if err != nil {
		return err
	}
	runtime, err := harnesssession.BorrowRuntime(repository, filepath.Join(resultRoot, suiteservices.ResolveRunID(pending.env)),
		pending.env[suiteservices.SuiteRuntimeRootEnv], pending.env[suiteservices.SuiteRuntimeRunIDEnv], pending.env[suiteservices.SuiteRuntimeLeaseIDEnv])
	if err != nil {
		return err
	}
	defer runtime.ReleaseHandles()
	return runtime.RecordResource("managed_suite", pending.path, "acquired", pending.postgres.cleanupFiles...)
}

func beginPartialSuite(ctx context.Context, service string, labels map[string]string, files []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	pending, ok := ctx.Value(partialSuiteKey{}).(*partialSuiteOwnership)
	if !ok || pending.journal == nil {
		return nil
	}
	// Before create, the journal records paths but no SQL cleanup credential is
	// required: no container has started. Acquired publication adds the live
	// credential dependencies before Container.Start can expose the service.
	return pending.journal.Begin(service, labels, files)
}
func settlePartialAttempt(ctx context.Context, service string, notSubmitted bool) error {
	pending, ok := ctx.Value(partialSuiteKey{}).(*partialSuiteOwnership)
	if !ok || pending.journal == nil {
		return nil
	}
	if notSubmitted {
		return pending.journal.NotSubmitted(service)
	}
	return pending.journal.Settled(service)
}
func settlePartialSuite(ctx context.Context, service string) error {
	pending, ok := ctx.Value(partialSuiteKey{}).(*partialSuiteOwnership)
	if !ok || pending.journal == nil {
		return nil
	}
	return pending.journal.Settled(service)
}
