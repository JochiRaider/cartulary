# MVP Stand-Up Deployment Package Handoff Tracker

Status: implementation-support live handoff tracker. This document is binding only for sequencing, sprint status, validation recording, and handoff interpretation for the MVP stand-up deployment package work. It is not a Core product specification and does not create, widen, narrow, or replace Base Profile or extension-profile conformance behavior.

Core 00 through Core 04 remain the implementation-conformance authority. Core 05 applies only at claim-bearing publication boundaries. `docs/testing-harness-nlspec.md` owns command invocation, target selection, scheduling, fixture lifecycle, artifact emission, cleanup, and verification gates. `docs/opentelemetry-instrumentation-nlspec.md` owns telemetry configuration containment, privacy, exporter behavior, and log correlation. `docs/domain.md` remains the domain vocabulary and concept-boundary reference.

Default decisions locked for this tracker:

- Package profile: `deployment_profile = "on_prem"`.
- Scope: minimal owner updates only.
- Target shape: one app image plus Postgres and SeaweedFS S3-compatible object store.
- Runtime: embedded frontend assets, no Vite/source-tree dependency.
- Backup: allocate persistent backup root, but do not claim full backup conformance unless backup capture and restore-verification scheduling are added.

Normative terms in this tracker have the following local meanings:

| Term | Meaning inside this tracker |
| --- | --- |
| `MUST` | Required for this handoff tracker to consider the named sprint or work unit complete. |
| `MUST NOT` | Forbidden for work performed under this tracker. |
| `SHOULD` | Required default unless this tracker records a specific exception and validation effect. |
| `MAY` | Optional tracker behavior only when omission behavior is stated in the same row or paragraph. |
| `default` | Required value, interpretation, action, or status when a more specific value is omitted. |

Statement classes:

| Statement class | Meaning | Required handling |
| --- | --- | --- |
| Tracker-owned requirement | Sequencing, status, validation, or handoff interpretation rule owned by this document. | Binding for work performed under this tracker. |
| Core-owner restatement | Compact pointer to behavior owned by Core 00 through Core 04. | MUST cite the owner; MUST NOT be treated as independent product authority. |
| Implementation guidance | Preferred implementation organization that does not change observable product behavior. | Follow unless a stronger local code boundary requires a recorded exception. |
| Validation requirement | Evidence required before a sprint or unit is complete. | Blocking unless recorded as skipped with reason and owner. |
| Research support | Non-normative rationale from research reports R01 through R09. | MAY explain a decision, but MUST NOT override owner documents. |
| Out-of-scope exclusion | Named work this tracker does not authorize. | MUST be handled by a separate tracker, owner-spec change, or product task. |

## Pinned Inputs And Source Snapshot

Default drift behavior: the tracker is bound to the repository snapshot at the time the file is added until this file is explicitly updated. If a source file changes later, the changed source informs future tracker revision but does not silently alter this tracker.

| Source path | Snapshot identity | Authority role | Required use | Drift handling |
| --- | --- | --- | --- | --- |
| `docs/spec/00_document_set_status_and_precedence.md` through `docs/spec/04_security_deployment_and_conformance.md` | Repo snapshot at tracker creation time. | Product-conformance owner corpus. | Resolve topology, runtime roots, deployment config, security, readiness, and acceptance criteria. | If this tracker and Core differ, Core governs and this tracker needs repair. |
| `docs/spec/05_claim_publication_and_benchmark_reproducibility.md` | Repo snapshot at tracker creation time. | Claim-publication owner only. | Use only if package evidence is later published as timed, benchmark, fixture-sensitive, or claim-bearing evidence. | Do not cite this tracker as Core 05 evidence. |
| `docs/testing-harness-nlspec.md` | Repo snapshot at tracker creation time. | Harness mechanics and Make-command owner. | Resolve command invocation, result-root interpretation, generated target handling, cleanup, and finalization. | Update this tracker before relying on changed command mechanics. |
| `docs/opentelemetry-instrumentation-nlspec.md` | Repo snapshot at tracker creation time. | Telemetry subsystem owner. | Preserve telemetry containment, privacy, exporter behavior, and log correlation in runtime/package changes. | Runtime diagnostics must be reconciled with this owner before completion. |
| `docs/domain.md` | Repo snapshot at tracker creation time. | Vocabulary and concept-boundary reference. | Resolve deployment administration, backup/restore, evidence, object blob, and operator-facing wording. | If vocabulary differs from a Core owner, owner section governs and this tracker needs repair. |
| `docs/guides/*.md` | Repo snapshot at tracker creation time. | Implementation-support guidance. | Use for existing build, packaging, and development conventions. | Guides do not override Core or harness owners. |
| `docs/research/R01-*` through `docs/research/R09-*` | Repo snapshot at tracker creation time. | Supporting research only. | Use where materially helpful to justify deployment-package decisions. | Research never creates conformance behavior. |

## Applicability And Omission Semantics

This tracker covers planning and implementation of an MVP stand-up deployment package for one app deployable, Postgres, and a SeaweedFS S3-compatible object store. Work outside that package is out of scope unless required by Core owner conformance or by tests touched by this tracker.

Omission behavior: a file, route, command, package artifact, deployment profile, or verification target not named in this tracker is out of scope unless a listed owner document or touched implementation boundary requires it. Out-of-scope items MUST NOT be added by implication from research reports, guides, or convenience scripts.

The MVP package is an on-prem local stand-up package. It MUST NOT be represented as disconnected-profile conformance unless Core 04 is separately changed to allow packaged local companion services for database and object storage.

## Sprint Governing Table

Sprint status defaults to `not-started`. A sprint can move to `complete` only when the validation status names passing evidence or a recorded skip with owner and impact.

| Sprint | Objectives | Implementation status | Work requirements | Validation status | Validation requirements |
| --- | --- | --- | --- | --- | --- |
| Sprint 0 | Characterize current deployment/build/config/readiness behavior and pin open gaps. | complete | Inspected Core 01/Core 04, config loader, object-store/Postgres bindings, embedded web assets, `/healthz`, `/readyz`, WS Origin, dev compose, and R01-R09 relevance. | complete | Inspection evidence and gap rows recorded in "Sprint 0 Characterization Evidence"; no production movement occurred before characterization. |
| Sprint 1 | Add minimal owner-document closure. | complete | Clarified packaged UI/no Vite runtime in Core 01; documented Postgres service-ref env binding, on-prem stand-up package rules, and readiness acceptance in Core 04; examples/supporting material were not touched. | complete | `make generated-artifact-policy-check` passed at `.cartulary/test-results/20260614T202557Z-p16248`; `make json-shape-check` passed at `.cartulary/test-results/20260614T202601Z-p16432`; `make lint-markdown` passed at `.cartulary/test-results/20260614T202607Z-p16775`. |
| Sprint 2 | Add package scaffolding. | complete | Added `deploy/mvp` package files: app image build, compose, config template, env template, bootstrap example, and package README. | complete | `make deployable-shape` passed at `.cartulary/test-results/20260614T203036Z-p25025`; `make lint-markdown` passed at `.cartulary/test-results/20260614T203040Z-p25210`. Static package-shape Make gate is deferred to Sprint 4 owner `make standup-package-smoke`; impact is that package static assertions remain inspection evidence until the smoke target is added. |
| Sprint 3 | Add runtime/init behavior needed by the package. | complete | Added deployment-local `cartulary-operator object-store init`, structured `/readyz`, ASCII-only service-ref normalization, and redaction-focused backend/process tests while preserving hidden-bucket-creation prohibition. | complete | `make backend-unit` passed at `.cartulary/test-results/20260614T204222Z-p42325`; `make backend-process` passed at `.cartulary/test-results/20260614T204300Z-p49224`; `make format` passed at `.cartulary/test-results/20260614T204216Z-p40935`. |
| Sprint 4 | Add Make-owned package smoke validation. | complete | Added public `standup-package-smoke` through topology owner inputs and regenerated derived task-surface files; smoke builds the image, runs compose, migrates, initializes object storage, starts the app, verifies embedded frontend/assets, no Vite/source-tree runtime, persistent roots, readiness, and WS Origin behavior. | complete | `make standup-package-smoke` passed at `.cartulary/test-results/20260614T205938Z-p74028`; `make deployable-shape` passed at `.cartulary/test-results/20260614T210009Z-p88028`; topology policy/shape checks passed. |
| Sprint 5 | Finalize docs, runbook, and tracker evidence. | complete | Updated package README runbook/troubleshooting, skipped-check notes, backup non-claim boundary, final tracker evidence, and retained-run maintenance artifacts. | complete | `make lint-markdown` passed at `.cartulary/test-results/20260614T210517Z-p52627`; `make check` passed at `.cartulary/test-results/20260614T210202Z-p89361`; `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260614T210202Z-p89361` passed at `.cartulary/test-results/20260614T210401Z-p50213`. |

## Implementation Plans By Sprint

Implementation order for every sprint: characterize current behavior, add or pin tests if needed, extract pure utilities only when needed, extract hooks after state/effect boundaries are clear, extract presentational components after props stabilize, then extract surface-specific components. This deployment-package tracker is backend/runtime focused; frontend hook and component extraction is expected to remain out of scope.

### Sprint 0 - Characterization

Plan order: characterize behavior first; add tests only after gaps are confirmed; no extraction or package edits.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Corpus map | `docs/handoffs/mvp-standup-deployment-package-handoff.md` | Live tracker authority, status rows, assumptions, evidence notes. | Product conformance behavior. | Inputs: Core docs, guides, R01-R09. Output: pinned tracker facts. |
| Runtime fact map | tracker section | Current build/config/service/readiness facts. | New requirements. | Inputs: repo inspection. Output: gap table. |
| Research support | tracker appendix | Material use of R01, R03, R06, R07, R09. | Normative behavior. | Output: concise rationale only. |

Sprint 0 required characterization rows:

| Area | Characterization requirement | Completion evidence |
| --- | --- | --- |
| App deployable | Confirm `cmd/server` remains the only runtime app unit and embeds frontend assets. | `make deployable-shape` or recorded inspection plus later package smoke. |
| Frontend runtime | Confirm production runtime does not require Vite, `apps/web`, `node_modules`, or a source-tree checkout. | Package image/static assertions in Sprint 4. |
| Config discovery | Confirm default `/etc/cartulary/config.toml` and absolute `CARTULARY_CONFIG_FILE` override behavior. | Config tests and package smoke. |
| Runtime roots | Confirm DB/object/backup/reference-pack/temp/export roots are absolute, persistent, distinct, and not source-tree-relative. | Config tests and compose/package assertions. |
| Object store | Confirm SeaweedFS S3-compatible storage remains the default local/package object-store target. | Owner-doc inspection and package smoke. |
| Readiness | Identify current `/readyz` behavior and gap to Core 04 structured readiness. | Backend/process tests in Sprint 3. |
| Origin validation | Confirm WebSocket cookie-authenticated browser Origin validation uses `application.public_origin`. | Existing WS tests or package smoke. |
| Backup boundary | Confirm persistent backup root exists and full backup conformance is not claimed unless backup jobs are added. | Handoff and README boundary notes. |

### Sprint 1 - Owner Closure

Plan order: pin behavior with doc checks before implementation.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Core 01 clarification | `docs/spec/01_architecture_storage_and_view_contracts.md` | App deployable includes packaged browser assets and no dev frontend runtime. | Config schema or deployment secrets. | Inputs: build/deployable facts. Output: minimal requirement/AC text. |
| Core 04 clarification | `docs/spec/04_security_deployment_and_conformance.md` | Stand-up package acceptance, `CARTULARY_POSTGRES_<REF>_DSN`, readiness/package criteria. | Alternative topology or disconnected-profile relaxation. | Inputs: current config model. Output: explicit on-prem package rules. |
| Supporting examples | appendices/guides only if touched | Example wording. | New config keys or weakened Core rules. | Output: aligned SeaweedFS/on-prem language. |

Sprint 1 closure rules:

- Core 01 text MUST stay limited to application-unit and packaged-resource behavior.
- Core 04 text MUST own operator-facing config, service binding, readiness, trust-boundary, and package acceptance behavior.
- Core 04 MUST document `CARTULARY_POSTGRES_<REF>_DSN` for managed Postgres service refs.
- Core 04 MUST NOT add vendor-specific TOML keys such as `seaweedfs.*`, `s3.*`, `bucket`, `endpoint`, `access_key`, or `secret_key`.
- Appendix or guide examples MAY show concrete Compose and config templates, but MUST NOT widen or relax Core rules.

### Sprint 2 - Package Scaffolding

Plan order: add package files after owner rules are explicit.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| App image | `deploy/mvp/Containerfile` | Final image with `cartulary-server`, `cartulary-migrate`, `cartulary-operator`; non-root runtime; embedded assets. | Node, pnpm, Vite, source tree, generated-file edits. | Inputs: Make build artifacts. Output: runnable app image. |
| Compose package | `deploy/mvp/docker-compose.yml` | `app`, `postgres`, `seaweedfs-s3`, one-shot `migrate`, one-shot object-store init. | Split frontend/API/worker app deployables. | Inputs: env/config/secrets. Output: local package topology. |
| Config templates | `deploy/mvp/config.toml.example`, `.env.example`, bootstrap example | On-prem managed-service bindings and persistent filesystem roots. | Real secrets or hidden runtime defaults. | Output: operator-copyable templates. |

Sprint 2 package rules:

- The final app image MUST contain `cartulary-server`, `cartulary-migrate`, and `cartulary-operator`.
- The final app image MUST run as a non-root user unless a recorded platform constraint blocks it.
- The final app image MUST NOT contain or require Node, pnpm, Vite, `apps/web`, `db/migrations` source files, or a repo checkout.
- Compose MUST mount `/etc/cartulary/config.toml` or set an absolute `CARTULARY_CONFIG_FILE`.
- Compose MUST persist Postgres data, SeaweedFS data, backup storage, reference-pack storage, temp work, and export outputs outside the source tree.
- Compose MAY include one-shot migration and object-store init jobs; those jobs are operational tooling and MUST NOT be described as additional app deployables.

### Sprint 3 - Runtime And Init

Plan order: add/pin tests first, then runtime helpers; no frontend hook/component extraction is in scope.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Object-store init | `cmd/operator` / `internal/app/operator.go` | Explicit deployment-local bucket init using configured object-store service ref. | App startup bucket creation or public API routes. | Input: config/env. Output: bucket exists or redacted failure. |
| Readiness provider | `internal/platform/httpapi` and app assembly | Structured readiness state and HTTP status per Core 04. | Secret/endpoint/bucket disclosure. | Inputs: startup dependency state. Output: `/readyz` status. |
| Config tests | `internal/platform/config`, `internal/platform/postgres` | Postgres service-ref env grammar and absolute config path behavior. | Generic DSN fallback. | Output: fail-closed validation evidence. |

Sprint 3 runtime rules:

- `cartulary-operator object-store init` MUST load deployment config, resolve `roots.object_storage`, create or confirm the configured bucket, and fail closed with redacted diagnostics.
- Object-store init MUST be deployment-local tooling only and MUST NOT add public HTTP routes.
- App startup MUST NOT create the production bucket as a hidden side effect.
- `/healthz` MUST remain process liveness and MUST NOT fail solely because the object store becomes unreachable after the process is running.
- `/readyz` MUST emit structured readiness and return HTTP 200 only for `status = "ready"` when object storage participates in the active deployment.
- Readiness/config diagnostics MUST NOT expose raw endpoint hosts, bucket names, credentials, object keys, storage refs, backend URLs, or secret material.

### Sprint 4 - Smoke Harness

Plan order: create Make-owned wrapper and smoke tests after package is runnable.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Public target | task-surface owner inputs | `standup-package-smoke` command surface. | Hand-edited generated task surface. | Output: regenerated generated files through Make only. |
| Smoke script | `scripts/ci/check-standup-package-smoke.sh` | Build image, compose up, migrate, init object store, verify runtime. | Long-lived dev services or source-tree runtime mounts. | Output: retained summary/artifacts. |
| Package assertions | smoke script/tests | No Vite/source tree, persistent roots, health/ready, embedded `/`, WS Origin. | Backup conformance claims. | Output: pass/fail package evidence. |

Sprint 4 smoke assertions:

- The package builds through Make-owned artifacts and does not hand-edit generated files.
- The app starts with `/etc/cartulary/config.toml` or an absolute `CARTULARY_CONFIG_FILE` override.
- The app serves `/` and embedded frontend assets from the server process.
- No Vite server, source-tree mount, `apps/web` runtime path, or `node_modules` runtime path is required.
- `cartulary-migrate` applies migrations from the package.
- `cartulary-operator object-store init` confirms or creates the configured bucket.
- `/healthz` returns HTTP 200 after process start.
- `/readyz` returns HTTP 200 only after dependencies are ready.
- WebSocket Origin validation rejects an untrusted cookie-authenticated browser Origin and accepts the configured `application.public_origin`.
- Backup root persistence is asserted, but backup capture and restore verification are not claimed.

### Sprint 5 - Final Handoff

Plan order: update tracker evidence, then run broad validation if runtime code changed.

| Unit | path | Owns | Must not own | Inputs/outputs |
| --- | --- | --- | --- | --- |
| Operator runbook | `deploy/mvp/README.md` | Build/start/verify/stop/troubleshoot steps. | Normative schema or broad deployment policy. | Output: actionable package instructions. |
| Tracker closure | handoff file | Sprint status, evidence roots, skipped checks, residual risks. | Prose-only completion claims. | Output: live Pursue Goal state. |
| Backup boundary note | handoff + README | Persistent backup root and non-claim boundary. | Full backup/restore conformance unless implemented. | Output: explicit assumption. |

Sprint 5 closure rules:

- Every completed sprint row MUST name validation evidence or a recorded skipped check with reason and impact.
- The final handoff update MUST list changed files, substantive edits, validation commands, results, skipped checks, and residual risks.
- If broad `make check` passes, `make agent-finalize RESULTS_DIR=<successful-check-run-root>` SHOULD be run and recorded.
- If broad `make check` is skipped, the tracker MUST record why narrower validation was sufficient.

## Interfaces And Public Additions

The following additions are planned work, not current behavior claims:

| Interface or artifact | Required behavior | Owner |
| --- | --- | --- |
| `CARTULARY_POSTGRES_<REF>_DSN` | Managed Postgres service-ref binding selected by normalized `roots.database_storage.service_ref`. Generic DSN env fallbacks must not satisfy a managed-service binding. | Core 04 and `internal/platform/postgres`. |
| `cartulary-operator object-store init` | Deployment-local command that resolves configured object storage, creates or confirms the configured bucket, redacts diagnostics, and never exposes public routes. | Core 04 deployment administration and `cmd/operator`. |
| `make standup-package-smoke` | Public Make target that validates the MVP package through owner-generated task-surface files. | `docs/testing-harness-nlspec.md` and task-surface owner inputs. |
| `deploy/mvp/Containerfile` | Builds a non-root app image with server, migrate, and operator binaries and embedded frontend assets. | Supporting deployment package. |
| `deploy/mvp/docker-compose.yml` | Starts app, Postgres, SeaweedFS S3, migration, and object-store init using persistent roots and secrets/env bindings. | Supporting deployment package. |

Deployment TOML MUST remain backend-neutral. It MUST NOT include vendor-specific object-store or database connection keys. Service-specific endpoint, credential, bucket, and DSN values belong in service-ref environment bindings or secret mechanisms, not in the config schema.

## Definition Of Done

| Acceptance row | Required evidence | Blocking |
| --- | --- | --- |
| App deployable embeds frontend | `make deployable-shape` and package smoke prove embedded `/` and `/assets/*`. | yes |
| No dev frontend runtime | Package smoke proves no Vite/source-tree/`apps/web` runtime dependency. | yes |
| Config discovery | Tests and smoke prove `/etc/cartulary/config.toml`, absolute override, and relative override rejection. | yes |
| Persistent roots | Static package checks prove DB/object/backup/reference-pack/temp/export roots are persistent and non-source-tree-relative. | yes |
| Postgres binding documented | Core 04 and tests cover `CARTULARY_POSTGRES_<REF>_DSN`. | yes |
| Object-store init | Operator command creates/confirms bucket with redacted errors and no public route. | yes |
| Readiness | `/healthz` and `/readyz` match Core 04 behavior. | yes |
| Origin validation | Smoke or process tests prove cookie-authenticated WS Origin exact-match against `application.public_origin`. | yes |
| Backup boundary | README and tracker state persistent backup root only; no backup conformance claim unless backup jobs are added. | yes |

## Commands

Planning and inspection:

```sh
make help
make help-all
make task-guide ROLE=feature-dev PHASE=phase0
make explain-target TARGET=deployable-shape DETAIL=summary
```

Build and package:

```sh
make build-server
make build-migrate
make build-operator
make deployable-shape
```

Docs and static validation:

```sh
make generated-artifact-policy-check
make json-shape-check
make lint-markdown
```

Runtime and service validation:

```sh
make backend-unit
make backend-process
make seaweedfs-compatibility
```

New package gate:

```sh
make standup-package-smoke
```

Broader closeout when runtime code changes:

```sh
make check
make agent-finalize RESULTS_DIR=<successful-check-run-root>
```

## Research Support

Research reports are supporting inputs only. They do not create owner requirements.

| Report | Material use for this tracker |
| --- | --- |
| R01 Aurora incident response report | Supports avoiding an Electron/file-local or renderer-trusted runtime shape and avoiding remote/CDN runtime assumptions. |
| R02 CRM/TEM DFIR report | Not material to deployment package mechanics; may support handoff discipline only. |
| R03 Kanvas technical report | Supports the need for low-friction portability while avoiding remote asset dependencies in deployable runtime. |
| R04 responsive browser spreadsheet UI memo | Not material to package mechanics beyond preserving browser-app behavior through service-backed runtime. |
| R05 responsive interface design report | Not material to package mechanics beyond preserving UI responsiveness and avoiding storage-internal coupling. |
| R06 spreadsheet-of-doom DFIR report | Supports structured Postgres plus object-store architecture, workbook projection model, and deliberate export/portability boundaries. |
| R07 spreadsheet-of-doom report | Supports preserving low-friction stand-up operations while improving audit/search/storage discipline. |
| R08 Handsontable React report | Not material to deployment package mechanics. |
| R09 React Data Grid report | Supports keeping Vite/demo/build tooling distinct from runtime dependencies. |

## Assumptions And Dependencies

- The MVP package is an on-prem local stand-up package, not disconnected-profile conformance.
- The object-store default remains SeaweedFS S3-compatible storage.
- The server serves same-origin upload targets through the app; S3 direct browser CORS must still avoid wildcard and `null` origins where exposed.
- The app image must not contain or require Vite, pnpm, Node runtime, `apps/web`, `db/migrations` source files, or a repo checkout.
- Backup storage is persistent in the MVP package, but backup capture, retention, and restore verification are separate unless explicitly added under this tracker.
- Full backup conformance requires Core 01/Core 04 backup-set, retention, encryption, and restore-verification evidence; this tracker does not claim that behavior by allocating a backup root.
- Generated files such as `tools/task_surface.generated.mk` MUST NOT be hand-edited. Update owner inputs and run the Make-owned generator when public target surfaces change.

## Sprint 0 Characterization Evidence

Sprint 0 status: complete. Source-limit note: characterization used local repository inspection only; no production package, runtime init behavior, or generated task-surface edits were made during Sprint 0.

| Area | Current evidence | Gap or follow-up |
| --- | --- | --- |
| App deployable | Core 01 REQ-01-002 and REQ-01-003 require one web application deployable. `cmd/server/main.go` is the runtime app unit; `cmd/migrate/main.go` and `cmd/operator/main.go` are operational entrypoints. `scripts/ci/check-deployable-shape.sh` already checks exactly those three entrypoints and verifies embedded frontend strings in the built server. | Sprint 1 must clarify that packaged browser assets live inside the app deployable and that production/package runtime must not require Vite or the source tree. Sprint 4 must run package smoke plus `make deployable-shape`. |
| Frontend runtime | `internal/platform/httpapi/httpapi.go` serves `/` and `/assets/` from `internal/platform/httpapi/webassets`; `scripts/ci/check-deployable-shape.sh` verifies `internal/platform/httpapi/webassets/dist/index.html` and a hashed asset are embedded in `./server`. | No MVP package image exists yet to prove absence of Node, pnpm, Vite, `apps/web`, `node_modules`, or source-tree runtime paths. |
| Config discovery | `internal/platform/config/config.go` sets `DefaultConfigPath = "/etc/cartulary/config.toml"` and accepts only an absolute `CARTULARY_CONFIG_FILE` override. `internal/platform/config/config_phase0_test.go` covers default selection, absolute override, relative override rejection, and overlay behavior. | Package compose/templates still need to mount `/etc/cartulary/config.toml` or use an absolute selector and prove this in smoke. |
| Runtime roots | Core 04 §12 requires explicit database, object, backup, reference-pack, temporary-work, and export roots. Config validation rejects missing roots, relative paths, shell-expanded paths, overlaps, and incompatible bindings. `configs/dev/config.toml` uses managed service refs for database/object storage and filesystem roots for backup/reference-pack/temp/export. | Package scaffolding must persist every root outside the source tree. Static package assertions and smoke are still missing. |
| Postgres binding | `internal/platform/postgres/postgres.go` resolves filesystem-root DSNs from `postgres.dsn` and managed service refs through `CARTULARY_POSTGRES_<REF>_DSN`; generic DSN fallback is not used for managed-service refs. | Core 04 currently lacks a parallel explicit Postgres service-ref binding paragraph, so Sprint 1 must add it. |
| Object store | Core 01 §4.2 names SeaweedFS S3 as the default official local/service-backed/disconnected example target. `docker-compose.dev.yml` starts `seaweedfs-s3`. `internal/platform/objectstore/objectstore.go` resolves object-storage managed-service refs through `CARTULARY_S3_<REF>_*` service-binding env vars and fails startup when the configured bucket is missing. | The package needs SeaweedFS S3 compose wiring and an explicit operator bucket-init command. App startup must continue not creating buckets as a hidden side effect. |
| Readiness | `/healthz` in `internal/platform/httpapi/httpapi.go` is process liveness and returns `200 ok`. `/readyz` currently returns static text `200 ready`. Core 04 REQ-04-078 requires structured readiness and `503` unless active dependencies are ready. | Sprint 3 must replace static `/readyz` with structured dependency readiness while keeping `/healthz` independent from post-start object-store reachability. |
| Origin validation | Core 01 REQ-01-277 and Core 04 REQ-04-110 require cookie-authenticated WebSocket Origin validation against `application.public_origin`. `internal/platform/ws/ws.go` and `internal/modules/collaboration/routes.go` enforce this before joining the incident stream; `internal/modules/collaboration/phase6_integration_test.go` rejects an untrusted Origin. | Sprint 4 smoke must prove the packaged runtime rejects an untrusted cookie-authenticated Origin and accepts the configured public origin. |
| Dev compose | `docker-compose.dev.yml` provides Postgres and SeaweedFS S3 only for local development, with SeaweedFS data persisted in a Compose volume and CORS defaulting to the dev public origin. | No `deploy/mvp` package exists yet; Sprint 2 must add package-local compose, config/env templates, bootstrap example, and README. |
| Backup boundary | Core 01 §12 and Core 04 §12.3.3 require real backup-set capture, retention, and restore-verification evidence for backup conformance. This tracker and Core/domain vocabulary keep backup/restore deployment-local. | The MVP package may allocate a persistent backup root only. Sprint 5 docs must avoid disconnected-profile and full-backup conformance claims. |
| Research support | R01 supports avoiding renderer-trusted or file-local runtime assumptions; R03 supports low-friction portability without remote asset dependencies; R06/R07 support structured Postgres plus object-store architecture; R09 supports keeping Vite/demo/build tooling out of runtime. | Research remains non-normative and did not alter Core owner behavior. |

## Live Execution Tracker

This section records implementation-support progress after the tracker is adopted. Rows start as `not-started`.

| Sprint | Status | Blocking work before completion | Evidence and notes |
| --- | --- | --- | --- |
| Sprint 0 | complete | None. | Characterization complete. Files changed: this handoff tracker only. Validation: repository inspection of Core 01/Core 04, config loader, Postgres/object-store bindings, embedded web asset serving, `/healthz`, `/readyz`, WS Origin, dev compose, and R01-R09 materiality. Skipped checks: no Make target required for Sprint 0; `make deployable-shape` is deferred to Sprint 4 package smoke to avoid production movement before characterization. Residual risks: package image, structured readiness, object-store init, and smoke target remain open for Sprints 1-4. |
| Sprint 1 | complete | None. | Files changed: `docs/spec/01_architecture_storage_and_view_contracts.md`, `docs/spec/04_security_deployment_and_conformance.md`, and this handoff tracker. Edits: Core 01 now states packaged deployments serve browser assets from the single application deployable without Vite, frontend source-tree, `apps/web`, `node_modules`, or a separate browser-UI deployable; Core 04 now states on-prem stand-up package boundaries without disconnected-profile claims, documents `CARTULARY_POSTGRES_<REF>_DSN`, and aligns readiness acceptance with active dependency readiness and redaction. Validation passed: `make generated-artifact-policy-check` (`.cartulary/test-results/20260614T202557Z-p16248`), `make json-shape-check` (`.cartulary/test-results/20260614T202601Z-p16432`), and `make lint-markdown` (`.cartulary/test-results/20260614T202607Z-p16775`). Skipped checks: none for Sprint 1. Residual risks: implementation and package smoke remain open for Sprints 2-4. Source-limit note: no examples, guides, generated files, or runtime code were edited in Sprint 1. |
| Sprint 2 | complete | None. | Files changed: added `deploy/mvp/Containerfile`, `deploy/mvp/docker-compose.yml`, `deploy/mvp/config.toml.example`, `deploy/mvp/.env.example`, `deploy/mvp/bootstrap-admin.json.example`, `deploy/mvp/README.md`, and this handoff tracker. Edits: app image copies Make-built `server`, `migrate`, and `operator` binaries as `cartulary-server`, `cartulary-migrate`, and `cartulary-operator` into a non-root distroless runtime; compose defines one app image plus Postgres plus SeaweedFS S3, one-shot migration, one-shot object-store init, absolute `/etc/cartulary/config.toml`, and Docker-managed persistent volumes for Postgres, SeaweedFS, backup, reference-pack, temp, and export roots; templates use on-prem managed service refs and placeholder secrets only. Validation passed: `make deployable-shape` (`.cartulary/test-results/20260614T203036Z-p25025`) and `make lint-markdown` (`.cartulary/test-results/20260614T203040Z-p25210`). Skipped/deferred check: executable static package-shape Make gate is owned by Sprint 4 `standup-package-smoke`, reason `cartulary-operator object-store init` and package smoke are not implemented until Sprints 3-4, impact static no-Vite/source-tree/root assertions are inspection evidence only until Sprint 4. Residual risks: compose `object-store-init` command will not pass until Sprint 3 runtime work; package image content and WS Origin behavior still need Sprint 4 smoke evidence. |
| Sprint 3 | complete | None. | Files changed: `internal/app/operator.go`, `internal/app/runtime.go`, `internal/app/operator_test.go`, `cmd/operator/operator_phase10_test.go`, `internal/platform/httpapi/httpapi.go`, `internal/platform/httpapi/readiness.go`, `internal/platform/httpapi/httpapi_test.go`, `internal/platform/objectstore/contract.go`, `internal/platform/objectstore/objectstore.go`, `internal/platform/objectstore/objectstore_phase0_support_test.go`, `internal/platform/postgres/postgres.go`, `internal/platform/postgres/postgres_phase0_settings_test.go`, and this handoff tracker. Edits: `cartulary-operator object-store init [-config <path>]` loads deployment config and creates/confirms the configured object-store bucket with redacted failure reason codes; app startup still rejects a missing managed-service bucket and does not create it; no public HTTP route was added; `/healthz` remains liveness; `/readyz` emits `cartulary.readiness.v1` JSON and returns `200` only for `status="ready"`; readiness reason data uses closed reason codes without raw endpoints, buckets, credentials, object keys, storage refs, backend URLs, or DSNs; Postgres and object-store service-ref normalization is ASCII-only and generic env fallbacks remain rejected. Validation passed: `make format` (`.cartulary/test-results/20260614T204216Z-p40935`), `make backend-unit` (`.cartulary/test-results/20260614T204222Z-p42325`), and `make backend-process` (`.cartulary/test-results/20260614T204300Z-p49224`). Validation note: initial `make backend-unit` failed at `.cartulary/test-results/20260614T204118Z-p33270` because empty test handlers passed a typed nil `*pgxpool.Pool` into readiness; `internal/platform/httpapi/readiness.go` now normalizes typed nil dependencies before probing. Skipped checks: none for Sprint 3. Residual risks: Make-owned package smoke, executable image-shape assertions, and WS Origin package evidence remain open for Sprint 4. Source-limit note: no generated files, frontend hooks/components, public object-store routes, or app-start bucket creation were added. |
| Sprint 4 | complete | None. | Files changed: `scripts/ci/check-standup-package-smoke.sh`, `deploy/mvp/Containerfile`, `deploy/mvp/runtime-roots/*/.keep`, `docs/testing-harness-nlspec.md`, `tools/execution_topology_manifest.json`, generated task-surface outputs (`tools/task_surface_manifest.json`, `tools/task_surface.generated.mk`, `tools/execution_topology_render_index.json`), and this handoff tracker. Edits: added a Make-owned package smoke target from topology owner inputs; the smoke builds the MVP image, proves the image contains `cartulary-server`, `cartulary-migrate`, and `cartulary-operator` without Node, pnpm, Vite, `apps/web`, `node_modules`, `db/migrations`, or a source checkout; runs the compose topology with Postgres and SeaweedFS S3; verifies migration and deployment-local object-store init; asserts `/healthz`, structured `/readyz`, embedded `/` and `/assets/*`, Docker-volume persistent roots, no app source-tree bind mount, and WebSocket Origin rejection/acceptance. Package image mount targets now include non-root-owned runtime-root placeholders so Docker named volumes are writable while the runtime remains non-root. Validation passed: `make generated-artifact-policy-check` (`.cartulary/test-results/20260614T205934Z-p73524`), `make json-shape-check` (`.cartulary/test-results/20260614T205934Z-p73548`), `make lint-shell` (exit 0, no run root emitted), `make standup-package-smoke` (`.cartulary/test-results/20260614T205938Z-p74028`), and `make deployable-shape` (`.cartulary/test-results/20260614T210009Z-p88028`). Validation note: initial `make standup-package-smoke` failed at `.cartulary/test-results/20260614T205438Z-p64330` because the non-root app container could not write the Docker-volume-backed backup root; the Containerfile/runtime-root fix resolved it. Skipped check: no separate `make service-backed-slice` run, owner `docs/testing-harness-nlspec.md` plus this handoff, reason the new `standup-package-smoke` is the sprint-scoped service-backed package validation and already owns the Compose lifecycle and dependency probes, impact no unrelated phase-slice evidence is added to this package handoff. Residual risks: final runbook wording, skipped-check notes, backup non-claim boundary, broad `make check`, and `make agent-finalize` remain open for Sprint 5. Source-limit note: generated files were updated only by `node scripts/render-execution-topology-artifacts.mjs`; no product routes, frontend hooks/components, split app deployables, disconnected-profile claims, or backup/restore conformance claims were added. |
| Sprint 5 | complete | None. | Files changed: `deploy/mvp/README.md`, this handoff tracker, and `make agent-finalize` retained-run artifacts (`tools/browser_e2e_duration_baselines.json`, `tools/go_test_duration_baselines.json`, `tools/harness_smoke_duration_baselines.json`, `tools/scheduler_manifest.json`, `tools/service_backed_make_target_duration_baselines.json`). Edits: README now documents the normal Compose `app` startup path, object-store init rerun, `/healthz` vs `/readyz`, persistent non-root runtime roots, `make standup-package-smoke`, no disconnected-profile claim, no backup/restore conformance claim, and troubleshooting for root permissions, object store, readiness, migration, and WebSocket Origin. Validation passed: `make lint-markdown` (`.cartulary/test-results/20260614T210517Z-p52627`), `make check` (`.cartulary/test-results/20260614T210202Z-p89361`), and `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260614T210202Z-p89361` (`.cartulary/test-results/20260614T210401Z-p50213`). Skipped checks: none for Sprint 5; retained-run maintenance was performed with the successful `make check` run root. Blockers: none. Residual risks: backup capture, retention, and restore verification remain out of scope; disconnected-profile and full backup conformance remain unclaimed. Source-limit note: finalizer-generated files were updated by `make agent-finalize`; no hand edits were made to generated artifacts. |

## Final Closure Evidence

- Planning summary: Sprints 0-5 executed in order using this handoff as the live tracker; Core 00-04, the testing harness spec, telemetry spec, and domain vocabulary remained the authority boundaries.
- Files changed: Core 01/Core 04 owner text, `deploy/mvp` package files, runtime/object-store/readiness/Postgres/operator code and tests, testing-harness command registry, topology owner input and generated task-surface outputs, finalizer-owned duration/scheduler artifacts, and this handoff tracker.
- Substantive edits: added the MVP package for one app image plus Postgres plus SeaweedFS S3, embedded frontend runtime evidence, deployment-local `cartulary-operator object-store init`, structured dependency readiness, ASCII-only service-ref binding normalization, package smoke validation, and operator runbook/troubleshooting.
- Validation: Sprint evidence rows above record all run roots. Final validation passed with `make lint-markdown`, `make check`, and `make agent-finalize RESULTS_DIR=.cartulary/test-results/20260614T210202Z-p89361`.
- Skipped checks: Sprint 4 records the only skipped check, a separate `make service-backed-slice`, with owner, reason, and impact; Sprint 5 skipped none.
- Blockers: none remaining.
- Residual risks and non-claims: the package allocates persistent storage roots only; backup capture, retention, restore verification, disconnected-profile conformance, and full backup conformance remain out of scope and unclaimed.
