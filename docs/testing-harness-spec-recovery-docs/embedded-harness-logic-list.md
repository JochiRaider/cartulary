---
doc_id: THR-S1-EMBEDDED-LOGIC
title: Testing Harness Recovery Embedded Harness Logic List
status: active
role: embedded-harness-logic-list
---

# Testing Harness Recovery Embedded Harness Logic List

## Document role

This S1 artifact records examples of harness mechanics embedded in ordinary
tests, app packages, frontend test files, and support modules. It helps later
sprints distinguish product evidence from reusable harness behavior.

This list is representative, not a complete behavioral specification.

## Embedded harness logic examples

| Logic ID | Surface | Embedded harness behavior | Boundary concern | Evidence | Evidence status | Later sprint |
|---|---|---|---|---|---|---|
| EHL-0001 | `internal/testutil/pgtest/pgtest.go` | Starts or attaches Postgres testcontainers, waits for mapped port/client readiness, manages shared harnesses, creates template/package/group/transaction databases, records fixture attribution, and stops shared containers. | Reusable harness mechanics live in Go test support and drive service-backed product tests. | `sed -n '1,260p' internal/testutil/pgtest/pgtest.go` | `observed` | S3/S4/S6 |
| EHL-0002 | `internal/testutil/s3test/s3test.go` | Starts or attaches object-store testcontainers, waits for readiness, creates package buckets, cleans prefixes/buckets, records fixture events, and exposes object-store env. | Object-store lifecycle is harness-owned but used inside product evidence. | `sed -n '1,240p' internal/testutil/s3test/s3test.go` | `observed` | S3/S4 |
| EHL-0003 | `internal/testutil/testcontainersx/testcontainersx.go` | Provides Docker preflight, retry, attempt timeout, retry-backoff, and diagnostic failure shaping for testcontainer startup. | Generic startup policy may become harness contract; exact failure taxonomy belongs later. | `sed -n '1,260p' internal/testutil/testcontainersx/testcontainersx.go` | `observed` | S5/S6 |
| EHL-0004 | `tools/testservices/main.go` | Wraps service-backed child commands, owns suite Postgres/object-store startup, service leases, cleanup/reaper scheduling, web E2E fixture preparation and cleanup, stale fixture janitor, timing buckets, and suite artifacts. | This is a central harness subsystem but its command contract is not yet traced. | `sed -n '1,240p' tools/testservices/main.go`; `make explain-target TARGET=backend-store` | `observed` | S2/S3/S4/S5 |
| EHL-0005 | `internal/testutil/suiteservices/env.go`, `diagnostics.go` | Defines suite env names, result root/run ID resolution, suite artifact directories, event names, service summaries, fixture summaries, cleanup summaries, and browser fixture summaries. | Diagnostic schema and artifact authority should be recovered before NLSpec drafting. | `sed` on `internal/testutil/suiteservices/env.go` and `diagnostics.go` | `observed` | S3/S5 |
| EHL-0006 | `internal/testutil/httptestx/httptestx.go` | Starts in-process app runtimes with temp roots, bootstrap routes, test clock routes, fixture config, and cleanup; provides request/envelope assertions. | Product tests depend on hidden app-runtime setup defaults that may be harness-owned. | `sed -n '1,220p' internal/testutil/httptestx/httptestx.go` | `observed` | S2/S5 |
| EHL-0007 | `internal/testutil/incidentwstest/incidentwstest.go`, `internal/testutil/wstest/wstest.go` | Dials canonical WebSocket routes, sends hello/resume scaffolds, reads presence snapshots, applies timeout contexts, registers close cleanup, and validates rejected dials. | WebSocket helper behavior is harness scaffolding mixed with product protocol assertions. | Files inspected with `sed` | `observed` | S5 |
| EHL-0008 | `internal/testutil/processtest/processtest.go` | Starts `cmd/server` as a subprocess using pre-opened listener fd, waits for `/healthz` and `/readyz`, captures stdout/stderr, validates diagnostics, and terminates the process group. | Process lifecycle and readiness behavior are harness mechanics embedded in backend process tests. | `sed -n '1,260p' internal/testutil/processtest/processtest.go` | `observed` | S2/S4/S5 |
| EHL-0009 | `cmd/server/shared_process_harness_test.go` | Uses `TestMain` to start shared Postgres and object-store harnesses for process tests, then stops them and fails the package if teardown fails. | Package-level shared service lifecycle is test code, but the sharing/teardown policy is harness behavior. | `sed -n '1,220p' cmd/server/shared_process_harness_test.go` | `observed` | S4/S6 |
| EHL-0010 | `internal/testutil/testruntime/reset.go` | Registers a test-only HTTP reset route that truncates mutable tables, preserves migration metadata, restores bootstrap admin, clears object storage, and returns reset counts. | Test-only route is harness-owned and lives outside production app implementation ownership; partial reset semantics remain source-limited. | `internal/testutil/testruntime/reset.go`; `MD-S7-0004` | `observed/maintainer_decision` | S7 |
| EHL-0011 | `apps/web/e2e/global-setup.ts`, `global-teardown.ts` | Clears or preserves worker/suite state based on external-server/shared-state mode, prepares suite admin TOTP state, prepares per-worker admin accounts, uses a shared global setup lock, and cleans worker sessions. | Browser harness global lifecycle is execution-critical and depends on env mode. | Files inspected with `sed`; Playwright shared config references global setup/teardown | `observed` | S4/S5/S6 |
| EHL-0012 | `apps/web/e2e/fixtures.ts` | Extends Playwright fixtures with worker admin context, page/request contexts, session tracking, revocation verification, cleanup, and worker index offset parsing. | Test fixture cleanup is product-test-visible and should be recovered as harness behavior. | `sed -n '1,260p' apps/web/e2e/fixtures.ts` | `observed` | S3/S5 |
| EHL-0013 | `apps/web/e2e/harnessState.ts`, `sessionSupport.ts`, `helpers.ts` | Resolves shared Playwright state directories/files, writes worker admin manifests, cleanup markers, suite admin TOTP state, API origin defaults, measurement sample policy, request holds, and cookie/storage helpers. | Runtime state file ownership and cleanup rules are deferred to artifact/lifecycle sprints. | Files inspected with `sed` | `observed` | S3/S4 |
| EHL-0014 | `scripts/start-web-e2e.sh` | Allocates backend/frontend ports, builds runtime roots, writes stack env/json/log files, sets Playwright state dir, starts server and Vite children, and supports session start/stop modes. | Shell lifecycle controls browser-backed harness but also launches product binaries. | `sed -n '1,220p' scripts/start-web-e2e.sh` | `observed` | S2/S4/S5 |
| EHL-0015 | `apps/web/src/testSetup.ts`, `apps/web/src/testSetup.dom.ts` | Resets Vitest timers after each test, polyfills CSS escape/ResizeObserver/focus/scrollIntoView, and installs layout/dimension fallbacks for jsdom grid tests. | Frontend product unit tests depend on test-only DOM behavior. | `sed` on setup files and Vite config setupFiles | `observed` | S5 |
| EHL-0016 | `scripts/test-*.sh`, `scripts/test-*.mjs` | Shell and Node self-tests create scratch repos/manifests/results dirs, install traps, assert task surface behavior, fake tool output, and verify scheduler/harness smoke behavior. | Harness self-tests are both evidence and consumers of harness behavior; do not treat their fake fixtures as product fixtures. | `find scripts -maxdepth 1 -name 'test-*'`; `task-surface-report --all` logical harness checks | `observed/runtime_observed` | S2/S3/S5 |
| EHL-0017 | `internal/testutil/phase*test/**`, `phase*storetest/**` | Phase harness packages start servers, seed incidents/users/records, build route inventory fixtures, and provide product-domain request builders. | Some content is product assertion scaffolding; later authority pass must avoid making domain fixtures harness-normative by accident. | `find internal/testutil -maxdepth 2 -type f`; `rg` product test imports | `observed` | S5/S8 |
| EHL-0018 | `apps/web/e2e/**/*.spec.ts` | Product browser specs rely on shared worker admin fixtures, session tracking, web E2E stack state, visual snapshots, timing expectations, and browser-stage selection. | Product evidence rows are not themselves harness contract, but their harness dependencies are. | `find apps/web/e2e -maxdepth 2 -type f`; `make explain-phase` browser evidence maps | `observed/runtime_observed` | S2/S5 |

## Boundary notes

- Embedded harness logic should be extracted as behavior only when it describes
  setup, teardown, artifact ownership, diagnostics, resource allocation,
  command contracts, or reusable test infrastructure.
- Product assertions, domain fixture values, acceptance claims, and route
  semantics remain product-test evidence unless a later authority pass assigns
  them to the harness specification.
