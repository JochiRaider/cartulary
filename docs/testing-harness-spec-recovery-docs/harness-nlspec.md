---
doc_id: THR-S7-HARNESS-NLSPEC
title: Testing Harness Natural Language Specification
status: draft
role: harness-nlspec
---

# Testing Harness Natural Language Specification

## Document Role

This draft specifies the recovered Cartulary testing harness contract. It is
derived from the S0 through S12 recovery package, selected S7 runtime evidence,
and the 2026-05-09 maintainer decisions in
`maintainer-decision-summary-2026-05-09.md`.

This specification governs harness mechanics only: command authority,
orchestration, local verification setup, scheduler behavior, service-backed
test setup, fixture and artifact lifecycle, cleanup bounds, diagnostics,
machine-readable outputs, and provider-neutral CI entrypoints. Product behavior
remains owned by Core 00 through Core 04. `docs/domain.md` remains a vocabulary
reference and does not own harness behavior.

The recovery package is in a reviewable recovered-specification state. S0
through S12 are complete unless a later recovery-process document supersedes
that state.

This document remains a draft until adopted by the repository owner. Within
that draft status, contract-bearing statements are prescriptive for the
recovered harness contract only when they are supported by the evidence and
acceptance rules below. Source-limited and owner-required rows are not final
requirements.

## Normative Language and Evidence

The words in this table are contract-bearing when they appear in this document.

| Term | Meaning |
|---|---|
| `must` | Required harness behavior for the stated scope. |
| `must not` | Prohibited harness behavior for the stated scope. |
| `may` | Permitted behavior that callers cannot require. |
| `default` | Behavior when the caller omits a configurable input. |
| `source-limited` / `source_limit` | Evidence is insufficient for a final normative claim. |
| `owner-required` / `owner_required` | A maintainer or governing owner must decide before the claim can become final. |

Final normative claims must use the evidence split recorded by `MD-S7-0017`
and acceptance row `HAC-0020`.

Hyphenated prose labels and underscore register labels identify the same
contract state:

| Prose label | Register or evidence label | Required treatment |
|---|---|---|
| `source-limited` | `source_limit` | Keep the claim non-final, name the blocked behavior, and name the needed follow-up. |
| `owner-required` | `owner_required` or `maintainer_decision_required` | Keep the claim non-final until a later owner decision supplies the contract. |
| `schema-unknown` | `schema_unknown` | Treat the surface as diagnostic or tool-owned, not as a stable field contract. |

| Evidence status | Contract treatment |
|---|---|
| `selected_runtime_observed` | may support a final harness requirement only for the selected command, inputs, environment, platform/tool profile, exit status, run identity, and artifacts recorded in `runtime-evidence-register.md`. |
| `source_observed` or `observed` | may support source-level declarations, static command contracts, defaults, schema identifiers, and routing rules. It does not prove runtime success. |
| `maintainer_decision` | may close an authority or intent gap when cited by decision ID and not in conflict with Core 00 through Core 04. |
| `source_limit` | must remain non-final. The blocked claim and required follow-up must stay visible. |
| `maintainer_decision_required` or `owner_required` | must remain non-final until a later owner decision closes it. |
| `schema_unknown`, `partial`, or `authority_unknown` | must not be promoted to stable field-level schema contracts. |

A final `must` or `must not` is reviewable only when the same section names one
of these routes: a `HAC-*` acceptance row, a `HAC-GAP-*` blocker, a
`source_limit` row, or a cited maintainer decision. If no route exists, the
sentence is a package defect and must be rewritten as source-limited,
owner-required, or accepted by a new or existing criterion.

## Audit Basis and Contract Delegation

This document adopts the row-level recovery registers below as the exact
current-state interface detail for the recovered harness contract. The rows
remain evidence-backed contract inputs, not independent behavior owners; Core 00
through Core 04 and any later adopted harness NLSpec still control conflicts.

| Recovery input | Row family | Contract detail delegated here |
|---|---|---|
| `entrypoint-command-map.md` | `EP-*` | Declared command shape, caller, inputs, defaults, outputs, side effects, ordering, parallel-safety, and source-observed usage behavior. |
| `observable-interface-map.md` | `OI-*` | Caller-visible stdout/stderr/report/log surfaces, consumers, ordering, stable-local/CI boundaries, and authority class. |
| `structured-output-schema-notes.md` | `SCHEMA-*` | Schema markers, producer/consumer relationships, and whether field-level contracts are complete, partial, tool-owned, or unknown. |
| `service-lifecycle-map.md` | `SVC-*` | Service provisioning, ready conditions, reset behavior, stop/cleanup behavior, and service-specific source limits. |
| `resource-allocation-register.md` | `RES-*` | Allocation, collision behavior, release behavior, reuse, parallel-safety, and concrete-resource non-guarantees. |
| `timeout-retry-register.md` | `TMR-*` | Timeouts, waits, retries, lock waits, watchdogs, polling loops, and timeout-unknown boundaries. |
| `cleanup-lifecycle-matrix.md` | `CLN-*` | Cleanup triggers, scope, order, idempotence, success/failure/timeout/interrupt behavior, and affected external state. |
| `failure-mode-register.md` | `FAIL-*` | Failure trigger, detection, report surface, retryability, ownership, cleanup follow-up, and source-limit routing. |
| `harness-acceptance-matrix.md` | `HAC-*`, `HAC-GAP-*` | Binary acceptance criteria and blocked final claims. |
| `source-limit-log.md` and `ambiguity-register.md` | `SL-*`, `AMB-*` | Preserved evidence gaps, authority gaps, and owner-required decisions. |

An invocation, output, schema, artifact, or service surface that is absent from
the adopted Make task surface and not cited by an `EP-*`, `OI-*`, `SCHEMA-*`, or
`SVC-*` row is not a canonical harness contract surface in this draft. It may be
investigated as source evidence, but callers must not rely on it unless a later
owner decision adopts it.

This delegation rule is accepted by `HAC-0001`, `HAC-0017`, and `HAC-0020`.

## Scope and Non-Goals

This specification owns the current harness contract for these surfaces:

| Surface | In scope | Owner evidence |
|---|---|---|
| Command invocation | Make public targets, aggregate sequences, generated task surface, scheduler entrypoints, package-script boundary. | `AUTH-0002`, `AUTH-0003`, `EP-*`, `MD-S7-0001` |
| Orchestration | Run lifecycle, scheduler fanout/fanin, resource claims, phase slices, browser batches, service-backed wrappers. | `LIFE-*`, `RES-*`, `OI-*` |
| Runtime services | Testservices, Postgres, MinIO, browser owned stack, local-dev Compose, reset route. | `SVC-*`, `ENV-*`, `MD-S7-0004`, `MD-S7-0005` |
| Artifacts and diagnostics | Fixtures, goldens, snapshots, generated artifacts, retained results, summaries, logs, reports, schema status. | `ART-*`, `SCHEMA-*`, `OI-*` |
| Cleanup and destructive safety | Cleanup tiers, stale janitor gates, cache exclusions, signal and reaper limits. | `CLN-*`, `CLEAN-*`, `AUTH-0007`, `MD-S7-0007`, `MD-S7-0008` |
| Provider-neutral CI | `make ci`, CI helper scripts, release gate relationship, stale extended smoke demotion. | `AUTH-0015`, `MD-S7-0015`, `MD-S7-0016` |

This specification does not define product behavior, public product APIs,
deployment conformance, provider-specific CI workflows, CI annotations, upload
paths, visual snapshot refresh bounds, complete non-Linux platform support, or
environment-variable precedence. Those claims remain source-limited or
owner-required where named below.

The Phase 6 browser login burst recorded by `MD-S7-0018` is harness-only test
stress margin. It does not change Core 04's human-user concurrent-session cap
or AC-159's sixth-login eviction semantics.

## Authority and Conflict Precedence

| Surface | Owns | Does not own | Conflict behavior | Evidence |
|---|---|---|---|---|
| Core 00 through Core 04 | Product behavior and implementation-conformance semantics. | Harness mechanics unless a core section explicitly says otherwise. | Core owner text governs product behavior. Harness conflicts become authority gaps. | `AUTH-0001` |
| `docs/domain.md` | Project vocabulary and boundary interpretation. | APIs, schemas, routes, tests, or harness mechanics. | Use for terminology only. Do not use it to override owner specs. | `docs/domain.md` |
| This harness NLSpec package | Adopted harness mechanics, validation orchestration, artifact lifecycle, diagnostics, cleanup bounds, and acceptance mapping. | Product semantics. | Later adopted harness text governs harness behavior unless a primary product owner conflicts. | `THR-*` package |
| `make` public surface | Canonical harness command surface. | Product behavior. | Direct scripts remain subordinate unless they re-enter Make wrappers. | `AUTH-0002`, `MD-S7-0001`, `HAC-0001` |
| Direct package scripts | Developer convenience entrypoints. | Canonical harness results, scheduler limits, result-root identity, or cleanup guarantees. | Treat direct outputs as tool-defined unless a script re-enters a Make-owned wrapper. | `AUTH-0003`, `EP-0016`, `OI-0013` |
| Generated task, schedule, Go, and TypeScript artifacts | Fresh downstream execution inputs. | Behavioral ownership. | Regenerate from owners and validate drift; do not hand-edit. | `AUTH-0010`, `MD-S7-0003`, `HAC-0002` |
| Harness implementation, tests, fixtures, and goldens | Evidence of current behavior and validation expectations. | Normative truth when conflicting with owner specs or decisions. | Route conflicts to the owning spec or maintainer decision. | `ART-*`, `PRES-*` |
| Retained local artifacts | Evidence for an explicitly selected run. | Ambient current truth. | Durable claims require explicit run identity. Newest fallback is human investigation only. | `AUTH-0013`, `MD-S7-0012`, `HAC-0014` |
| CI helper scripts | Provider-neutral CI dispatch through Make. | Provider workflow, annotation, upload, dashboard, or hosted-runner behavior. | Keep provider behavior source-limited while `.github/**` is absent. | `AUTH-0015`, `SL-0001`, `HAC-0018` |

## Harness Run Model

A harness run is one caller-visible invocation through the canonical Make
surface. A run starts when the selected Make target begins resolving its
inputs, and it ends when that target returns its final exit status.

### Run Defaults

| Input | Default or omitted behavior | Invalid input behavior | Evidence |
|---|---|---|---|
| Result root | `CARTULARY_TEST_RESULTS_DIR=.cartulary/test-results` for Make-owned runs. | Path validation failures are `configuration_error` or command-specific usage failures. | `EP-0002`, `ART-0013`, `ENV-0001` |
| Run ID | Make creates `CARTULARY_TEST_RUN_ID=<utc>-p<pid>` unless overridden; runner context may use `adhoc` when no env exists. | Missing explicit run identity blocks durable artifact claims. | `ENV-0001`, `AUTH-0013` |
| Output mode | `CARTULARY_OUTPUT_MODE=summary`; `VERBOSE=1` or CI verbosity controls may alter streaming according to Make logic. | Unknown or unsupported output modes fail in the consuming wrapper where source validates them. | `ENV-0002`, `OI-0001` |
| CI mode | `make ci` exports `CI=1`; `scripts/ci/verify.sh` sets `CARTULARY_OUTPUT_MODE=ci` when unset and then execs `make ci`. | Provider workflow behavior remains source-limited. | `EP-0017`, `OI-0014` |
| Generated task surface | Generated Make include and schedule manifests may drive execution only when fresh relative to upstream manifests. | Drift or generated-artifact policy mismatch is `manifest_or_accounting_mismatch`. | `ART-0008`, `ART-0010`, `SCHEMA-0020` |

### Retained Artifact Evidence Claims

Retained local artifacts support durable evidence claims only when each required
identity element is explicit.

| Required element | Contract | If absent |
|---|---|---|
| Result root | The claim names `RESULTS_DIR` or `CARTULARY_TEST_RESULTS_DIR`. | The artifact may be used only for human investigation. |
| Run ID | The claim names the exact `RUN_ID` or `CARTULARY_TEST_RUN_ID`. | Newest-run fallback is diagnostic only. |
| Command or target | The claim names the invoked Make target, scheduler target, or wrapper command. | Runtime behavior cannot be generalized from the artifact. |
| Platform/tool profile | The claim names the observed host/tool profile recorded with the selected run. | Platform support remains source-limited. |
| Exit status | The claim names the final exit status or selected failure state. | Pass/fail behavior cannot be promoted to a final claim. |
| Artifact paths | The claim names the specific summary, log, report, or fixture paths used. | Field-level or cleanup claims remain unsupported. |

### Lifecycle States

| State | Meaning | Caller-visible result | Cleanup treatment |
|---|---|---|---|
| `passed` | All selected child work, required summaries, and required finalizers passed. | Exit `0`; summaries report pass. | Command-owned cleanup completed or was not required. |
| `usage_error` | Caller supplied invalid command shape, flag, mode, or required argument. | Source-observed usage paths exit `2`. | No child cleanup is expected unless partial setup already started. |
| `configuration_error` | Required config, env, manifest, binary, or path is missing or invalid. | Non-zero exit, often before child work. | Cleanup is required only for resources already started. |
| `test_assertion_failure` | Product-under-test assertion failed after harness setup reached test execution. | Child failure propagates through summaries and exit status. | Harness resources used by the test are cleaned according to their owner. |
| `harness_operational_failure` | Harness setup, service, scheduler, report, artifact, fixture, or cleanup behavior failed independently of a product assertion. | Non-zero exit with failure summaries or diagnostics. | Cleanup follows the applicable cleanup tier. |
| `partial_completion` | Setup, child work, reporting, or cleanup completed only partly. | Summary records failed, skipped, or partial units when reporting is reached. | Retained artifacts become diagnostic evidence only. |
| `cancelled_or_interrupted` | Caller, timeout wrapper, CI, signal, or parent process interrupted execution. | Exit status is command-specific and source-limited except selected evidence rows. | Best-effort cleanup unless selected evidence proves a stronger path. |
| `source_limited_unknown` | The current package lacks selected evidence or authority for the claim. | No final contract claim. | Follow-up is required before normativity. |

## Command Contracts

The Make public surface is the sole canonical harness command surface. The
exact public target inventory is owned by the task-surface manifests and
`entrypoint-command-map.md`; this section adopts the command-family contracts
needed by callers and implementers.

| Family | Canonical surface | Inputs and defaults | Outputs | Side effects | Failure classes | Acceptance |
|---|---|---|---|---|---|---|
| Help and planning | `make help`, `make help-all`, `make task-guide`, `make target-plan`, `make explain-*` | Target args include `PHASE`, `TARGET`, `DETAIL`, `RESULTS_DIR`, and `RUN_ID` where named by the target. Omitted args are valid only when the target accepts omission in `entrypoint-command-map.md` or source; omission of a required arg is `usage_error`. | Human text or tool-declared JSON; no child tests. | Reads manifests or selected retained artifacts. | `usage_error`, `configuration_error`, `manifest_or_accounting_mismatch`. | `HAC-0001`, `HAC-0014` |
| Phase command targets | Generated `phase_command` Make recipes through `scripts/lib/run-phase.sh`. | Make env, target-specific env, output mode defaults, command args from generated recipes. | Phase summary, bounded terminal lines, logs under result root, and target summary only when the generated recipe declares one. | Target-specific; may validate, build, lint, generate, format, start services, or clean. | Child non-zero, `usage_error`, `configuration_error`, `harness_internal_error`. | `HAC-0001`, `HAC-0020` |
| Aggregate sequences | `make test-fast`, `make test`, `make ci`, `make release-check`. | Sequence name from `tools/task_surface_manifest.json`; `ci` sets `CI=1`. | Run summary, aggregate target summary, child summaries. | Runs child Make targets; may start services, browsers, builds, release evidence, or checks through child targets. | First failing child fails serial sequence; helper/report failures are harness failures. | `HAC-0018`, `HAC-0019` |
| Check scheduler | `make check`. | `CHECK_SCHEDULE_MANIFEST`, host resource envs, service/browser envs; manifest defaults to `tools/check_schedule_manifest.json`. | Check scheduler summary/events, progress log, target/run summaries. | Runs readiness, static checks, nested service-backed work, browser work, builds, and finalizers. | `usage_error`, `configuration_error`, child failure, finalizer failure, cleanup failure. | `HAC-0008`, `HAC-0017` |
| Service-backed scheduler | `test-service-backed`, `test-fast-service-backed`, `check-service-backed`, phase service-backed slices. | `SERVICE_BACKED_SCHEDULE_MANIFEST`, testservices binary, resource envs; manifest defaults to `tools/service_backed_schedule_manifest.json`. | Scheduler summary/events, child summaries, service-scope artifacts. | Starts or attaches to service-backed suite, runs Go/browser work, runs finalizers. | Service start/readiness errors, child failures, artifact failures, cleanup errors. | `HAC-0008`, `HAC-0010` |
| Go target runner | `make backend-unit`, service-backed Go targets, Go phase slices. | Go target name or shard plan; Go cache env defaults to `/tmp/cartulary-go-*`. | Raw Go JSONL, phase summaries, target summaries, timing/accounting artifacts. | Runs Go tests; service-backed variants use Postgres/MinIO through scheduler/testservices. | `test_assertion_failure`, `manifest_or_accounting_mismatch`, service failures. | `HAC-0011`, `HAC-0017` |
| Frontend unit/typecheck | `make frontend-typecheck`, `make frontend-unit`, related lint/build targets. | Node/pnpm runtime from Make; Vitest worker env only for Vitest-backed targets. | Tool output, Vitest runner JSON, watchdog JSON on timeout, summaries through Make. | Typechecks, tests, lints, or builds frontend according to target. | Tool failure, watchdog `timeout`, package-script bypass source limits. | `HAC-0001`, `HAC-0017` |
| Browser E2E | `make browser-e2e-*`. | Browser batch manifest, Playwright env, owned-stack env, worker defaults from Make or Playwright config. | Stack metadata, server/web logs, Playwright diagnostics, reset artifacts, summaries. | Starts backend/frontend stack, prepares DB/bucket fixtures, may reset state, runs Playwright groups. | Service readiness timeout, port conflict, reset failure, Playwright assertion failure, cleanup failure. | `HAC-0005`, `HAC-0006`, `HAC-0016` |
| Testservices | `tools/testservices` via Make and schedulers. | Suite commands, Docker/testcontainers environment, result/run identity. | Service scope, event files, lease, reaper log, fixture summaries. | Starts Postgres/MinIO, creates template DB, wraps child work, schedules or performs cleanup. | `usage_error`, `preflight_error`, `service_start_error`, child exit, cleanup error. | `HAC-0009`, `HAC-0010` |
| Direct package scripts | Root and `apps/web` package scripts. | Tool-native pnpm/Vite/Vitest/Biome/Playwright inputs. | Tool-native stdout/stderr and reports. | Tool-defined build, test, lint, format, or browser-wrapper effects. | Tool-defined failures; Make result-root and scheduler guarantees absent unless wrapper re-entry occurs. | `HAC-0001` |
| Maintenance writers and cleanup | `make generate`, `make format`, ledgers/schedules, baseline refresh, `make clean`, `make distclean`. | Target-specific Make variables; duration refresh uses explicit `RESULTS_DIR`; `agent-finalize` skips duration refresh when `RESULTS_DIR` is unset. | Mutated generated/docs/baseline artifacts, formatted files, or removed ignored paths. | Mutates repo-tracked generated/docs/baselines or deletes ignored runtime/tool artifacts. | Drift, writer failure, unsafe cleanup path, cleanup failure. | `HAC-0002`, `HAC-0011` |

Direct package scripts must not be documented as canonical harness contracts
unless a later owner decision adopts them. Package scripts that call repository
wrappers may rely only on the wrapper behavior they actually invoke.

The command-family rows above incorporate these controlling interface rows by
reference. A later edit that changes a command family's inputs, outputs,
side effects, timing, cleanup, failure handling, or schema status must update
the corresponding recovery row or add a new accepted row instead of relying on
uncited prose.

| Family | Controlling rows | Acceptance or blocker |
|---|---|---|
| Help and planning | `EP-0001`, `EP-0014`, `OI-0011`, `OI-0016`, `SCHEMA-0017`, `FAIL-0001`, `FAIL-0003`, `FAIL-0027`. | `HAC-0001`, `HAC-0014` |
| Phase command targets | `EP-0002`, `OI-0001`, `SCHEMA-0001`, `SCHEMA-0002`, `FAIL-0001`, `FAIL-0003`, `FAIL-0011`, `FAIL-0013`, `FAIL-0017`. | `HAC-0001`, `HAC-0020` |
| Aggregate sequences | `EP-0008`, `OI-0003`, `SCHEMA-0004`, `SCHEMA-0006`, `FAIL-0011`, `FAIL-0013`, `FAIL-0017`, `FAIL-0026`. | `HAC-0018`, `HAC-0019`, `HAC-GAP-0007` |
| Check scheduler | `EP-0007`, `OI-0004`, `SCHEMA-0005`, `SCHEMA-0006`, `RES-0001` through `RES-0010`, `TMR-0020`, `TMR-0021`, `FAIL-0004`, `FAIL-0013`, `FAIL-0014`, `FAIL-0025`. | `HAC-0008`, `HAC-0017`, `HAC-0020` |
| Service-backed scheduler | `EP-0006`, `EP-0012`, `OI-0005`, `SCHEMA-0005`, `SCHEMA-0006`, `SVC-0001`, `RES-0003`, `RES-0005` through `RES-0010`, `TMR-0020`, `TMR-0021`, `CLN-0011`, `CLN-0012`, `FAIL-0005`, `FAIL-0006`, `FAIL-0014`. | `HAC-0008`, `HAC-0010`, `HAC-0020` |
| Go target runner | `EP-0004`, `EP-0005`, `OI-0006`, `SCHEMA-0007`, `SCHEMA-0011`, `SVC-0004`, `SVC-0006`, `RES-0014` through `RES-0016`, `RES-0025`, `TMR-0018`, `TMR-0019`, `TMR-0022`, `TMR-0023`, `TMR-0026`, `FAIL-0011`, `FAIL-0013`. | `HAC-0011`, `HAC-0017`, `HAC-0020` |
| Frontend unit/typecheck | `EP-0002`, `EP-0019`, `OI-0007`, `OI-0013`, `SCHEMA-0008`, `SCHEMA-0012`, `FAIL-0011`, `FAIL-0012`, `FAIL-0013`. | `HAC-0001`, `HAC-0017`, `HAC-0020` |
| Browser E2E | `EP-0009`, `EP-0010`, `EP-0018`, `OI-0008`, `OI-0009`, `SCHEMA-0009`, `SCHEMA-0010`, `SCHEMA-0013`, `SVC-0009` through `SVC-0014`, `RES-0017`, `RES-0021` through `RES-0023`, `RES-0026`, `TMR-0008` through `TMR-0012`, `TMR-0016`, `TMR-0024`, `TMR-0025`, `TMR-0029`, `TMR-0030`, `CLN-0013` through `CLN-0019`, `FAIL-0008` through `FAIL-0011`, `FAIL-0014`, `FAIL-0021`, `FAIL-0022`. | `HAC-0005`, `HAC-0006`, `HAC-0016`, `HAC-GAP-0006` |
| Testservices | `EP-0011`, `OI-0010`, `SCHEMA-0014`, `SCHEMA-0015`, `SCHEMA-0016`, `SVC-0001` through `SVC-0007`, `SVC-0011`, `RES-0011` through `RES-0016`, `TMR-0001` through `TMR-0007`, `TMR-0017`, `TMR-0028`, `CLN-0011` through `CLN-0013`, `FAIL-0005` through `FAIL-0007`, `FAIL-0014`, `FAIL-0015`. | `HAC-0009`, `HAC-0010` |
| Direct package scripts | `EP-0016`, `EP-0018`, `EP-0019`, `OI-0013`, `SCHEMA-0018`, `TMR-0032`, `CLN-0020`, `FAIL-0018`. | `HAC-0001` |
| Maintenance writers and cleanup | `EP-0020`, `OI-0015`, `SCHEMA-0020`, `CLN-0001`, `CLN-0002`, `TMR-0027`, `TMR-0031`, `FAIL-0013`, `FAIL-0016`, `FAIL-0026`. | `HAC-0002`, `HAC-0011` |

For package scripts that invoke repository wrappers, only the wrapper behavior
identified by the relevant `EP-*` and `OI-*` rows is adopted. Package-manager
sequencing, tool-native reports, direct Playwright/Vitest output, and artifacts
outside the Make result-root policy remain tool-defined or authority-ambiguous.

### Command Modes and Side Effects

| Command mode | Representative surfaces | Allowed side effects | Required result boundary | Acceptance or blocker |
|---|---|---|---|---|
| Inspection and planning | Help, task-guide, target-plan, explain, fixture-report, task-surface reports. | Reads manifests, source metadata, or explicitly selected retained artifacts. It does not start child tests or backing services. | Human output or tool-declared machine output only; retained-artifact evidence requires the identity fields above. | `HAC-0001`, `HAC-0014` |
| Pure validation | Unit, lint, typecheck, static drift, and build-check targets that do not require shared service setup. | May write run summaries, logs, coverage, tool caches, and build outputs declared by the target. A target that requires the service-backed stack belongs in service-backed validation. | Child failure becomes target failure; tool-owned reports remain tool-owned unless schema-adopted. | `HAC-0001`, `HAC-0017`, `HAC-0020` |
| Service-backed validation | Service-backed schedulers, service-backed phase slices, store/integration/process targets with Postgres or MinIO. | Starts or attaches to testservices, creates DBs/buckets/templates, writes service scope artifacts, and schedules or runs owned cleanup. | Service setup, child work, finalizers, and cleanup failures are harness failures unless classified as product test assertions. | `HAC-0008`, `HAC-0009`, `HAC-0010` |
| Browser validation | Browser E2E targets and webserver-backed batches. | Starts owned backend/frontend stacks, prepares browser fixtures, writes Playwright and stack diagnostics, and may call the reset route only through test wiring. | Browser reports and traces are diagnostic unless schema-adopted; reset-route behavior remains harness-owned. | `HAC-0005`, `HAC-0006`, `HAC-0016`, `HAC-GAP-0006` |
| Maintenance writers | `generate`, `format`, phase ledger/schedule generation, duration baseline refresh, release evidence writers. | May mutate only the repo-tracked generated, formatted, baseline, or release-evidence paths declared for that target. Generated outputs remain downstream inputs, not owners. | Drift or writer failure is `manifest_or_accounting_mismatch` or `harness_internal_error` according to the failing target. | `HAC-0002`, `HAC-0011` |
| Cleanup | `clean`, `distclean`, target cleanup helpers, stale janitors, reapers. | Deletes only the scoped paths or external resources authorized by cleanup bounds and proof gates. | Cleanup failures are `cleanup_error`; stronger cleanup guarantees require selected evidence. | `HAC-0009`, `HAC-0010`, `HAC-0011` |
| Direct package scripts | Root or `apps/web` package scripts invoked outside Make. | Tool-defined side effects only, except when the script re-enters a repository wrapper. | They do not inherit Make result-root, scheduler, cleanup, or evidence guarantees beyond the wrapper behavior actually invoked. | `HAC-0001` |

## Scheduler and Resource Model

Scheduler lanes and claims are logical constraints on harness work selection and
ordering. They are not physical capacity guarantees for host CPUs, Docker,
Postgres, MinIO, browser processes, ports, or network resources.

| Resource class | Meaning | Default or override | Boundary | Evidence |
|---|---|---|---|---|
| Check host lanes | `host_cpu` and `host_io` bound check-scheduler work. | Auto policies with `CHECK_HOST_CPU_JOBS` and `CHECK_HOST_IO_JOBS`; CLI `--resource-limit` may override. | No OS CPU or I/O reservation is implied. | `RES-0001`, `RES-0002` |
| Suite service stack | Serializes check-scheduler ownership of the shared testservices stack. | Registry default limit `1`. | Protects harness stack setup, not every concrete DB or bucket operation. | `RES-0003` |
| Service-backed DB/object lanes | `postgres`, `minio`, and `postgres_reset` account for service-backed work. | Registry defaults and schedule claims; reset-heavy work requires the reset lane. | Logical only; actual service capacity remains runtime/platform-dependent. | `RES-0005`, `RES-0006`, `RES-0010` |
| Go lanes | `go_cpu` and `go_io` bound service-backed Go shards. | Auto policies with `CARTULARY_SERVICE_BACKED_GO_CPU_LIMIT` and `CARTULARY_SERVICE_BACKED_GO_IO_LIMIT`. | Auto estimates are scheduler inputs, not host guarantees. | `RES-0007` |
| Browser lanes | `browser_stack`, `browser_stage_<stage>`, and `process` bound browser stages and process-heavy work. | `CARTULARY_SERVICE_BACKED_BROWSER_STACK_LIMIT` may override stack limit; manifests declare stage claims. | Does not allocate ports or browser profiles by itself. | `RES-0008`, `RES-0009` |
| Ports | Browser stacks prefer dynamic ports; dev/Compose use fixed defaults unless env overrides. | Browser dynamic selection checks candidates when `ss` exists; dev defaults `8080` and `5173`; Compose maps `5432`, `9000`, `9001`. | `ss` absence makes port checks platform-limited. | `RES-0017`, `RES-0018`, `RES-0019` |
| External Go caches | Go build/module caches under `/tmp/cartulary-go-*`. | Make and runner defaults. | Outside default `clean` and `distclean`; Go tooling owns cache consistency. | `RES-0025`, `MD-S7-0009`, `HAC-0011` |

Duration baselines are planning weights. They may inform scheduler ordering and
drift checks, but they do not prove throughput, runtime capacity, or service
health.

Timing and resource-sensitive behavior uses the following source-observed
defaults and boundaries. These rows define what callers may rely on without
turning source-limited runtime behavior into guaranteed cleanup, throughput, or
platform support.

| Surface | Default or bound | Exhaustion or invalid behavior | Preserved limit | Rows |
|---|---|---|---|---|
| Scheduler progress and resource waits | Progress tick default `60000ms`; no generic wall-clock timeout for pending work waiting on dependencies or resource claims. | Failed dependencies, failed children, invalid manifests, or invalid resource claims fail or skip work through scheduler rules; ordinary resource waiting is not itself a timeout failure. | No physical capacity or throughput guarantee. | `TMR-0020`, `TMR-0021`, `RES-0001` through `RES-0010` |
| Suite service startup | Preflight `3s`; Postgres wrapper `2m`; MinIO wrapper `5m`; template setup `2m`; stale fixture janitor `10s`. | Startup, readiness, template, or janitor failure is a harness setup or fixture failure before or around child work as recorded by service summaries. | Live Docker/testcontainers timing and retry exhaustion remain source-limited outside selected evidence. | `TMR-0001` through `TMR-0007`, `SVC-0001` through `SVC-0007` |
| Browser backend/frontend readiness | Backend and frontend readiness loops use 180 attempts at 1 second; Playwright webServer timeout is `180000ms`. | Early process exit or readiness exhaustion fails browser startup and retains recent logs/stack artifacts where the wrapper reaches reporting. | Browser runtime success, failure bundles, and interruption cleanup remain source-limited outside selected evidence. | `TMR-0008`, `TMR-0012`, `SVC-0009`, `SVC-0010`, `FAIL-0009`, `FAIL-0021` |
| Process groups and port release | Stop sends TERM, waits 50 x 0.2 seconds, then sends KILL and waits again; port release checks also use 50 x 0.2 seconds when `ss` exists. | Cleanup or release failures become cleanup diagnostics or cleanup failure only where the wrapper observes them. | Signal behavior, parent death, and `ss`-absent behavior remain source-limited. | `TMR-0009`, `TMR-0010`, `RES-0017`, `RES-0020`, `CLN-0015` |
| Lock waits | Playwright global setup lock deadline is `120s`; Go dependency warm and shared report locks default to `300s` with 100ms polling. | Lock timeout fails the setup or report-capture path with the source-observed failure surface. | Abrupt-exit stale lock recovery remains source-limited. | `TMR-0011`, `TMR-0018`, `TMR-0019`, `RES-0022` |
| Reset route call and route internals | No route-specific timeout is recovered; reset script and handler use curl/request context behavior. | Non-200, invalid JSON, unexpected schema/counts, or handler error fails the reset wrapper or returns `500 test_runtime_reset_failed`. | Partial reset failure semantics and route-specific timeout remain source-limited. | `TMR-0016`, `TMR-0030`, `SVC-0014`, `FAIL-0010` |
| Ordinary DB and object operations | Ordinary fixture DB create/drop/reset and MinIO bucket create/cleanup use caller context; no fixed per-operation timeout is recovered. | SQL or object-store operation failures fail the owning setup, cleanup, or fixture path. | Active DB cleanup and bucket operation timeout guarantees remain source-limited. | `TMR-0022`, `TMR-0023`, `SVC-0004`, `SVC-0006`, `CLN-0005` through `CLN-0010` |
| Retained artifact readers | Durable claims require explicit `RESULTS_DIR`, run ID, command/target, platform/tool profile, exit status, and artifact paths. | Missing or ambiguous selected artifacts fail the reader or remain human-investigation-only according to the tool. | Newest-run fallback never supports durable normative claims. | `TMR-0027`, `OI-0011`, `HAC-0014` |

## Environment and Platform

The harness documents only source-observed environment variables, defaults, and
validation behavior. Cross-layer precedence remains
`TODO: precedence_unknown` until a future owner decision supplies a precedence
matrix.

| Environment group | Default or valid values | Invalid or omitted behavior | Precedence status | Evidence |
|---|---|---|---|---|
| Result identity | `CARTULARY_TEST_RESULTS_DIR=.cartulary/test-results`; run ID from Make UTC timestamp plus PID. | Durable claims without explicit result dir and run ID are invalid. | Source-observed; direct package behavior differs. | `ENV-0001`, `HAC-0014` |
| Output mode | `CARTULARY_OUTPUT_MODE=summary`; `VERBOSE`/CI verbosity may change streaming. | Wrapper-specific usage/config failure when invalid. | Source-observed. | `ENV-0002` |
| Go cache/toolchain | `GO`, `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`, `GOCACHE`, `GOMODCACHE`; default `/tmp/cartulary-go-*`. | Go tool failures propagate through target summaries. | Source-observed. | `ENV-0003`, `HAC-0011` |
| Node/pnpm | Repo-local `tmp/node-runtime`, pnpm under node runtime, Make-provided `PATH`. | Missing tools fail doctor/bootstrap/toolchain or child command. | Direct package scripts may bypass Make setup. | `ENV-0004` |
| Scheduler capacity | `CHECK_HOST_*`, `CARTULARY_SERVICE_BACKED_*` envs plus CLI resource overrides. | Invalid positive-limit validation fails scheduler setup before child work. | Source-observed; nested forwarding remains source-limited. | `ENV-0006`, `HAC-0008` |
| Service attach | Valid attach configuration requires complete Postgres and MinIO attach envs; the DSN template contains `{database}`; MinIO secure defaults false when blank. | Missing or malformed values fail attach/start setup. | Source-observed. | `ENV-0008`, `ENV-0010` |
| Local-dev Compose | Compose file default `docker-compose.dev.yml`; Postgres wait `180s`; MinIO wait `120s`; bucket `cartulary`. | Compose, readiness, or reset failures fail local commands. | Source-observed; local-only. | `ENV-0012`, `HAC-0007` |
| Runtime config | `CONFIG_FILE` and `CARTULARY_CONFIG_FILE` default to `configs/dev/config.toml` through Make/dev/browser paths. | Missing/unreadable config fails startup. | Precedence with product config is source-limited. | `ENV-0013` |
| Browser stack | Runtime root under target artifact dir or `/tmp`; Playwright public origin default `http://127.0.0.1:4173`; workers default differs by caller; scheduled browser groups require explicit worker-admin slot env. | Port conflicts, missing required webserver-backed env, missing scheduled worker slot env, or startup failure fail browser target. | Source-observed; direct Playwright differs. | `ENV-0014`, `ENV-0016`, `ENV-0017` |
| Test routes | `CARTULARY_ENABLE_TEST_ROUTES=1` enables test-route wiring in owned browser backend. | Omitted/default runtime leaves reset route disabled. | Source-observed and maintainer-decided. | `ENV-0019`, `MD-S7-0004`, `HAC-0005` |
| Platform tools | WSL/Linux current host observed with Go, Node, pnpm, Docker, Compose, Playwright, `bash`, `curl`, `ss`, `setsid`, `realpath`, `timeout`. | Missing-tool behavior is specified only where source validates it. | Full platform matrix is source-limited. | `ENV-0024`, `MD-S7-0011`, `HAC-0013` |

WSL/Linux is the primary observed environment. This draft makes no final
portability claim beyond source-observed Linux behavior and the explicit
source-limited platform rows. Non-Linux hosts, missing tools, browser
installations, network restrictions, and provider CI environments remain
outside the complete support matrix.

## Services and Local Verification

Local-dev Compose, `make db-up`, `make services-up`, `make db-reset`, and
`make dev` are local verification behavior, not deployment conformance.

| Service surface | Provision and ready condition | Default bounds | Reset/stop/cleanup | Boundary |
|---|---|---|---|---|
| Testservices suite | Starts Postgres and MinIO, creates migrated template DB, writes lease, runs stale fixture janitor, then starts child work. | Preflight `3s`; Postgres wrapper startup `2m`; MinIO wrapper startup `5m`; template `2m`; stale fixture janitor `10s`. | Normal `run` defers owned cleanup; detached `start-suite` requires `terminate-suite --lease`; reaper completion remains source-limited. | `SVC-0001`, `RES-0012` |
| Suite Postgres | Testcontainers Postgres 16; ready after port wait and admin DSN ping. | Port mapping `15s`; readiness `15s` with `250ms` poll and `5s` attempt timeout; startup attempts `3` with `500ms` retry backoff. | Container cleanup or reaper; active DB cleanup remains source-limited. | `SVC-0002`, `SVC-0004` |
| Suite MinIO | Testcontainers MinIO; ready after port wait and authenticated `ListBuckets`. | Port mapping `30s`; readiness `60s` with `500ms` poll and `5s` attempt timeout; wrapper overall `5m`; attempts `2`. | Container cleanup or reaper; bucket cleanup uses caller context where no fixed timeout is declared. | `SVC-0005`, `SVC-0006` |
| Local Compose | Starts Postgres and MinIO from `docker-compose.dev.yml`. | Postgres wait `180s`; MinIO wait `120s`. | No `docker compose down` target is recovered; `db-reset` resets only the local Postgres DB and not MinIO. | `SVC-0008`, `MD-S7-0005` |
| Browser owned backend/frontend | Allocates runtime root and ports, starts backend and Vite, waits for `/readyz` and frontend origin. | Backend/frontend readiness loops 180 attempts at 1 second; Playwright webServer timeout `180000ms`. | Exit trap stops process groups; runtime root removed unless retained under target artifacts; completion under interruption is source-limited. | `SVC-0009`, `SVC-0010` |
| Browser active fixture | `prepare-web-e2e` clones a DB and creates a bucket under active testservices. | Uses template startup timeout `2m`. | `cleanup-web-e2e` records retirement; suite cleanup later leak-checks and reclaims. | `SVC-0011` |
| Browser standalone fixture | Starts local Compose, creates `cartulary_web_e2e_<pid>`, runs migrations. | Compose and backend/frontend readiness bounds above. | Drops standalone E2E DB; Compose services and volumes persist. | `SVC-0012` |
| Local `make dev` | Starts backend/frontend after caller starts backing services. | Default ports `8080` and `5173`; ready timeout `180s`; config `configs/dev/config.toml`. | Exit trap stops processes; no automatic DB/object reset. | `SVC-0015`, `HAC-0007` |

Service rows that name `TODO: readiness_unknown`,
`TODO: db_operation_timeout_unknown`, `TODO: bucket_operation_timeout_unknown`,
or related source limits must not be upgraded by inference.

## Test Runtime Reset Route

The test runtime reset route is harness-owned. Its implementation must live
outside production application ownership, remain disabled by default, and be
registered only by explicit test-route wiring. The recovered source-observed
wiring uses `CARTULARY_ENABLE_TEST_ROUTES=1`.

| Property | Contract |
|---|---|
| Route status | Test hook only; not a product API. |
| Default behavior | Disabled by default; default runtime returns 404 for `/api/v1/test/runtime/reset`. |
| Enabled behavior | Test-route-enabled runtime exposes reset behavior for browser reset boundaries. |
| Schema | Response data schema marker is `cartulary.test.runtime_reset.v1`. |
| Side effects | Resets app test state through Postgres and object-store dependencies, reruns bootstrap admin, and lets reset script clear configured Playwright state only after valid response. |
| Failure behavior | Non-200, invalid JSON, or unexpected response causes reset wrapper failure. Partial reset failure semantics remain source-limited. |
| Acceptance | `HAC-0004`, `HAC-0005`, `HAC-0006`. |

## Artifacts, Fixtures, and Schemas

### Artifact Authority

| Artifact class | Contract treatment | Update or cleanup rule | Evidence |
|---|---|---|---|
| Canonical fixtures and goldens | Committed validation inputs and expected outputs. They may encode test truth, not product authority by themselves. | Controlled owner-reviewed edits only unless a later supported refresh command is adopted. | `ART-0001`, `ART-0002`, `ART-0005`, `HAC-0015` |
| Visual snapshots | Committed validation baselines. | `make browser-e2e-visual` may validate; no update command, OS, browser, or version is adopted. | `ART-0003`, `MD-S7-0013`, `HAC-0016`, `HAC-GAP-0002` |
| Generated Go/TS/task/schedule artifacts | Downstream execution inputs when fresh. | Use generation targets and drift/policy checks; do not hand-edit. | `ART-0006` through `ART-0010`, `HAC-0002` |
| Duration baselines | Committed planning inputs derived from successful selected runs. | Refresh only with explicit `RESULTS_DIR=<successful run dir>`; no refresh was run in S7 evidence. | `ART-0011`, `runtime-evidence-register.md` |
| Retained result artifacts | Durable evidence only with selected run identity. | `make clean` removes default result root; stale artifacts are investigation-only unless selected. | `ART-0013` through `ART-0018`, `HAC-0014` |
| Tool/runtime caches | Disposable or reused local tool state. | `clean` and `distclean` follow Make cleanup bounds; external Go caches remain outside both. | `ART-0022` through `ART-0024`, `HAC-0011` |
| External service state | Databases, buckets, containers, processes, browser profiles, ports. | Cleanup follows owning service/helper tier and source limits. | `ART-0025` through `ART-0029`, `CLN-*` |
| Direct package-script artifacts | Tool-defined unless package script re-enters harness wrapper. | Cleanup coverage is mixed and not canonical Make policy. | `ART-0030`, `OI-0013` |

Fixture, golden, and snapshot updates require explicit intent, targeted file
list, source evidence or test reason, targeted verification command, and review
note explaining why the change is expected. The S7 missing fixture/golden/
snapshot review found no missing item in the audited source set; visual snapshot
refresh authority remains owner-required.

### Schema Status

| Schema status | Stable contract allowed | Required boundary | Acceptance or blocker |
|---|---|---|---|
| Stable source schema ID | Field names, types, required/default/conditional fields, ordering, null/omission behavior, and compatibility rules may be specified where source declares them or acceptance names them. | Do not infer fields from examples or retained artifacts alone. | `HAC-0017`, `HAC-0020` |
| `source_observed` | Source-declared shape and routing may be referenced. | Runtime success, failure coverage, and unextracted fields remain source-limited. | `HAC-0017` |
| `partial` | Only inspected fields and cases may be relied on. | Unknown fields, variants, ordering, and absent/failure cases remain source-limited. | `HAC-0017` |
| `schema_unknown` | No field-level harness contract is allowed. | Treat the output as diagnostic or tool-owned. | `HAC-GAP-0006` where browser/Playwright-owned. |
| `authority_unknown` | No first-class harness schema is allowed. | Requires a later owner decision before adoption. | `HAC-0017`, `HAC-0020` |

Schema adoption is not inferred from the presence of JSON, retained examples, or
tool reports. A stable source schema ID adopts only the producing/consuming
surface and the source-declared field contract recorded by
`structured-output-schema-notes.md` and the referenced producer source.
`partial` rows adopt only the inspected fields and cases named by their row.
Tool-defined Go, Vitest, Playwright, pnpm, Vite, Biome, shell log, and provider
CI outputs remain tool-owned unless a later owner decision adopts a harness
schema for them.

Stable schema IDs include phase, target, run, scheduler, tool-run, web E2E
stack, selected reset/session, and generated manifest schemas recorded in
`SCHEMA-0001`, `SCHEMA-0003`, `SCHEMA-0004`, `SCHEMA-0005`, `SCHEMA-0006`,
`SCHEMA-0009`, `SCHEMA-0010`, and `SCHEMA-0020`. Tool-defined Go, Vitest,
Playwright, shell log, package-script, fixture-report, and provider-CI outputs
remain limited exactly as recorded in `structured-output-schema-notes.md`.

## Cleanup and Destructive Safety

Cleanup claims use these tiers. A claim must not exceed the tier supported by
selected evidence or source.

| Cleanup tier | Allowed claim strength |
|---|---|
| `observed_successful_cleanup` | Selected runtime evidence shows the specific cleanup path completed successfully. |
| `observed_cleanup_scheduling` | Source or selected evidence shows cleanup was scheduled; completion is not proven. |
| `delayed_after_state_evidence` | Later observation saw expected after-state; synchronous cleanup completion is not proven. |
| `best_effort_cleanup` | Source attempts cleanup; failures, interrupts, parent death, or detached completion are not guaranteed. |
| `source_limited_cleanup` | Source exists or routing is known, but selected runtime evidence or owner decision is missing. |

Stale janitors may delete databases, buckets, or containers only when every
proof gate in this table passes.

| Proof gate | Required condition | If absent |
|---|---|---|
| Generated resource name | Candidate matches Cartulary-generated DB, bucket, or container naming. | Deletion prohibited. |
| Ownership evidence | Harness metadata, service lease, label, or equivalent ownership proof exists. | Deletion prohibited. |
| Age or completed-summary evidence | Candidate is conservatively stale or tied to a completed cleanup summary. | Deletion prohibited. |
| Scope-limited resource type | Candidate is a DB, bucket, or container type the janitor is authorized to handle. | Deletion prohibited. |

Cleanup boundaries:

| Boundary | Contract |
|---|---|
| `make clean` | Removes configured repo-local build, report, coverage, result, release, Vite cache, and selected tmp scratch paths with guarded removal. |
| `make distclean` | Includes `clean` scope plus repo-local tool/runtime caches and dependency install state such as node runtime, toolbin, frontend install/cache, `.cache`, root/workspace `node_modules`, and `.pnpm-store`. |
| External Go caches | `/tmp/cartulary-go-*` remains outside default `clean` and `distclean`. |
| Timeout and interrupt | Selected S7 evidence supports only selected scenarios and delayed after-state where recorded. Parent-death cleanup remains source-limited. |
| Detached reaper | Scheduling and delayed Docker after-state are evidence; hard completion guarantee remains source-limited. |
| Active DB connections | Unit-level leak detection exists; live active-connection cleanup guarantee remains source-limited. |

Cleanup trigger mapping:

| Trigger | Required cleanup treatment | Failure behavior | Evidence route |
|---|---|---|---|
| Normal successful Make-owned run | Run command-owned cleanup to the tier supported by source or selected evidence. | Cleanup failure after successful child work changes the harness result to `cleanup_error` where the wrapper observes it. | `HAC-0010` |
| Child failure after setup | Attempt the cleanup owned by the service, browser stack, fixture, or target helper already started. | Child failure remains the primary failure unless cleanup also fails and the wrapper records cleanup failure. | `HAC-0010`, `HAC-0020` |
| Timeout or interrupt | Attempt only best-effort cleanup unless selected evidence names a stronger path. | Exit status and cleanup completion remain command-specific and source-limited outside selected evidence. | `HAC-GAP-0003`, `SL-0014` |
| Detached reaper scheduling | Scheduling may be claimed when source or selected evidence proves scheduling. | Hard completion cannot be claimed without later evidence. | `SL-0014`, `cleanup-signal-evidence-register.md` |
| Stale janitor deletion | Delete only when every proof gate passes. | Missing proof prohibits deletion; unsafe attempted deletion is `cleanup_error` or `harness_internal_error`. | `HAC-0009` |
| `make clean` and `make distclean` | Delete only the scoped cleanup-boundary paths above. | Guarded-removal failures are cleanup failures. | `HAC-0011` |
| External Go caches | Do not delete `/tmp/cartulary-go-*` through default `clean` or `distclean`. | Cache staleness is tool-managed and not a cleanup failure. | `HAC-0011` |

## Failure Taxonomy

Every harness failure must be classified with one of these classes or routed as
source-limited.

| Failure class | Trigger phase | Caller-visible exit/report behavior | Cleanup behavior | Retryability | Owner |
|---|---|---|---|---|---|
| `usage_error` | Argument, flag, path, mode, or unsupported combination is invalid before child work. | Source-observed CLI usage paths use exit `2`; otherwise non-zero usage exit is source-limited. No child summary is required unless the wrapper has already entered reporting. | No child cleanup unless partial setup already started. | No until caller fixes invocation. | Harness command owner. |
| `configuration_error` | Required env, config, binary, manifest, path, or toolchain is missing, invalid, or contradictory. | Non-zero before child work or before dependent child work; summary exists only when reporting is reached. | Cleanup only for resources already started. | No until configuration changes. | Harness/platform owner. |
| `preflight_error` | Docker, testcontainers, platform, secret, or prerequisite check fails before managed service startup. | Non-zero setup failure with preflight diagnostics where the wrapper reports them. | No service cleanup unless partial startup began. | Only after the environment or prerequisite is fixed. | Harness/platform owner. |
| `service_start_error` | Required service cannot be provisioned or started. | Non-zero harness setup failure; scheduler or service summary records failure when reporting is reached. | Cleanup follows started-service owner and supported tier. | Only after the transient service or environment cause clears. | Harness service owner. |
| `service_readiness_timeout` | Service starts but does not satisfy its ready condition before the declared timeout. | Non-zero harness setup failure with readiness diagnostics when the owning wrapper reports them. | Cleanup follows service owner and timeout source limits. | Only after the readiness cause clears or timeout/configuration changes. | Harness service owner unless readiness endpoint is product-under-test. |
| `fixture_error` | Fixture, reset, DB, bucket, template, janitor, or fixture shape operation fails. | Non-zero harness or target failure before or during child work, depending on fixture phase; unplanned/unsafe fixture use is included. | Cleanup follows fixture/service owner; destructive retry requires proof gates. | Only after the fixture cause is corrected or evidence classifies it as transient. | Harness fixture/service owner. |
| `resource_conflict` | Port, DB, bucket, lock, cache, worker, lane, process, or profile conflict blocks setup or scheduling. | Non-zero setup/scheduler failure or source-limited platform behavior for unvalidated conflicts. | Cleanup is limited to owned resources already allocated. | Only after the conflict clears. | Harness resource owner. |
| `test_assertion_failure` | Product-under-test assertion fails after harness setup reaches test execution. | Child failure propagates through target and aggregate summaries. | Harness resources are cleaned by their owner; cleanup failure is additional harness failure. | No by default unless a separate flake/rerun policy exists. | Product/test owner. |
| `harness_internal_error` | Harness code fails independently of product assertions. | Non-zero harness failure with logs or summaries when reporting is reached. | Cleanup follows the relevant owner and source limits. | Only after the harness defect is fixed. | Harness owner. |
| `timeout` | Configured deadline elapses. | Non-zero timeout result or wrapper-specific timeout reporting; exact exit is source-limited unless selected evidence records it. | Best-effort cleanup unless selected evidence proves stronger. | Only after the timeout cause, resource bound, or configuration is fixed. | Harness owner unless product behavior is explicitly timeout-based. |
| `cancelled_or_interrupted` | Caller, CI, signal, parent death, or timeout wrapper interrupts execution. | Exit status is command-specific and source-limited outside selected evidence. | Best-effort cleanup; parent-death completion remains source-limited. | Unknown. | Harness/external controller. |
| `cleanup_error` | Cleanup failed, leak check failed, reaper scheduling failed, or cleanup completion is unverifiable where completion is required. | Non-zero harness failure or retained diagnostic state when the wrapper observes the cleanup issue. | No stronger retry or deletion authority is implied. | Only after the cleanup cause is fixed; destructive retries require authority. | Harness cleanup owner. |
| `duration_baseline_drift` | Explicit duration-baseline or warm scheduler timing drift validation fails. | Non-zero timing-maintenance command or finalizer substep with retained timing evidence. | No cleanup contract beyond the source run owner. | Only after a successful uncontaminated retained run refreshes or validates timing evidence. | Harness timing-maintenance owner. |
| `unsupported_platform` | Current host/tool profile cannot satisfy declared harness contract. | Non-zero preflight/configuration failure only where source validates it; otherwise platform support is source-limited. | Cleanup only for resources already started. | No until environment or support profile changes. | Harness/platform owner. |
| `missing_secret` | Required credential or secret is unavailable. | Non-zero preflight/configuration failure where source validates the secret. | No cleanup unless partial setup started. | No until credential exists. | Harness/platform owner. |
| `manifest_or_accounting_mismatch` | Manifests, generated task surfaces, ledgers, schemas, or accounting reports disagree. | Non-zero drift, policy, schedule, or accounting failure. | Cleanup is target-specific and does not repair drift. | No until owner/generated artifacts are fixed. | Harness manifest/generated owner. |
| `authority_required` | Behavior exists but cannot be made normative without owner decision. | Not a runtime retry class; it blocks final claim wording. | No cleanup contract. | Not runtime-retryable. | Named owner/maintainer. |
| `unknown_failure` | Failure was observed but not classified with available evidence. | Diagnostic only; it cannot remain in a final accepted claim. | Source-limited until classified. | Unknown. | Source-limited until classified. |

`unknown_failure` must not remain in a final accepted claim.

## CI, Release, and Stale Smoke

CI remains provider-neutral. Repository documentation must not invent `.github`
workflow behavior, provider annotations, upload paths, dashboards, or hosted
runner semantics while provider workflow sources are absent.

| Surface | Contract | Acceptance |
|---|---|---|
| `make ci` | Provider-neutral CI enforcement entrypoint. It composes the canonical Make task surface through `check`. | `HAC-0018` |
| CI helper scripts | `scripts/ci/verify.sh` delegates to `make ci`; deployable-shape helper validates deployable-shape files. | `HAC-0018` |
| `make release-check` | Release verification gate composed from `check`, license report, SBOM, and build work. Release readiness is claimed only when evidence supports it. | `HAC-GAP-0007` when not passing or not claimed |
| `run-harness-smoke-extended` | Demoted stale diagnostic target. It must not block phase advancement, `ci`, or `release-check`. | `HAC-0019` |
| `check-harness-smoke` and fast smoke | Developer-gate smoke surface when passing through current task-surface rules. | `HAC-0019` |

## Preserved Source Limits

These items remain blocked until selected evidence or owner decision closes
them.

| Blocked claim | Current treatment | Blocker |
|---|---|---|
| Environment-variable precedence | `TODO: precedence_unknown`; source-observed vars/defaults only. | `HAC-GAP-0001`, `SL-0015`, `MD-S7-0010` |
| Visual snapshot refresh OS/browser/version/command | Validation-only; no update command or platform bound adopted. | `HAC-GAP-0002`, `AUTH-0014`, `MD-S7-0013` |
| Parent-death cleanup | No guaranteed abrupt-exit cleanup claim. | `HAC-GAP-0003`, `SL-0014` |
| Active DB cleanup | No guaranteed live active-connection cleanup claim. | `HAC-GAP-0004`, `SL-0014` |
| Detached reaper hard completion | Scheduling and delayed after-state are not a hard guarantee. | `SL-0014`, `cleanup-signal-evidence-register.md` |
| CI provider annotations/uploads | Provider-specific behavior unavailable while `.github/**` is absent. | `HAC-GAP-0005`, `SL-0001` |
| Playwright report/trace/video/screenshot internals | Tool-owned or `schema_unknown`. | `HAC-GAP-0006`, `SCHEMA-0013` |
| Release readiness beyond recorded evidence | Separate from stale-smoke demotion. | `HAC-GAP-0007` |
| Phase 7/Phase 8 active coverage | Planned future work only, not current active coverage. | `HAC-0003`, `MD-S7-0002` |

## Acceptance Map

Acceptance criteria live in `harness-acceptance-matrix.md`. This document is
reviewable only when each final `must` or `must not` maps to one row below or
to a named source limit in the previous section.

| Requirement area | Acceptance or blocker |
|---|---|
| Make is canonical and package scripts are subordinate. | `HAC-0001` |
| Generated artifacts are downstream inputs only. | `HAC-0002` |
| Phase 7 and Phase 8 are planned future work. | `HAC-0003` |
| Reset route is harness-owned, disabled by default, enabled only through test wiring, and schema-preserving. | `HAC-0004`, `HAC-0005`, `HAC-0006` |
| Local-dev Compose and `make dev` are local verification behavior, not deployment conformance. | `HAC-0007` |
| Scheduler lanes are logical constraints only. | `HAC-0008` |
| Stale janitors require destructive deletion proof gates. | `HAC-0009` |
| Cleanup strength and external Go cache boundaries remain bounded. | `HAC-0010`, `HAC-0011` |
| Environment/default documentation and platform claims remain source-bounded. | `HAC-0012`, `HAC-0013`, `HAC-GAP-0001` |
| Durable retained-artifact claims require explicit run identity. | `HAC-0014` |
| Fixture/golden refresh workflow and visual snapshot validation-only status are explicit. | `HAC-0015`, `HAC-0016`, `HAC-GAP-0002` |
| Unknown schemas remain unknown. | `HAC-0017`, `HAC-GAP-0006` |
| Failure taxonomy and source-limited failure routing are evidence-gated. | `HAC-0020` |
| CI remains provider-neutral and stale extended smoke is non-blocking. | `HAC-0018`, `HAC-0019`, `HAC-GAP-0005`, `HAC-GAP-0007` |
| Final normative language is evidence-gated. | `HAC-0020` |
