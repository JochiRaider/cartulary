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
| Sprint 0 | Characterize current deployment/build/config/readiness behavior and pin open gaps. | not-started | Inspect Core 01/Core 04, config loader, object-store/Postgres bindings, embedded web assets, `/healthz`, `/readyz`, WS Origin, dev compose, and R01-R09 relevance. | not-started | Record inspection evidence; no production movement before characterization. |
| Sprint 1 | Add minimal owner-document closure. | not-started | Clarify packaged UI/no Vite runtime in Core 01; document Postgres service-ref env binding and stand-up acceptance in Core 04; keep examples/supporting material non-normative. | not-started | `make generated-artifact-policy-check`; `make json-shape-check`; `make lint-markdown`. |
| Sprint 2 | Add package scaffolding. | not-started | Add `deploy/mvp` package files: app image build, compose, config template, env template, bootstrap example, package README. | not-started | Static package-shape tests plus existing build checks. |
| Sprint 3 | Add runtime/init behavior needed by the package. | not-started | Add deployment-local object-store init through `cartulary-operator`; ensure server readiness reflects Postgres/object-store/bucket state; preserve fail-closed config validation. | not-started | Backend/process tests for config, readiness, bucket missing/init, redaction. |
| Sprint 4 | Add Make-owned package smoke validation. | not-started | Add public Make target through owner inputs, not generated files by hand; smoke build image, run compose, migrate, init object store, start app, verify embedded frontend and Origins. | not-started | `make standup-package-smoke`; `make deployable-shape`; targeted service-backed checks. |
| Sprint 5 | Finalize docs, runbook, and tracker evidence. | not-started | Update tracker rows, package README, skipped-check notes, backup non-claim boundary, troubleshooting. | not-started | `make lint-markdown`; `make check` when runtime code changed; `make agent-finalize RESULTS_DIR=<run-root>` after broad verification. |

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

## Live Execution Tracker

This section records implementation-support progress after the tracker is adopted. Rows start as `not-started`.

| Sprint | Status | Blocking work before completion | Evidence and notes |
| --- | --- | --- | --- |
| Sprint 0 | not-started | Characterize deployment/build/config/readiness behavior and record the gap table. | TODO: add inspection evidence and any source-limit notes. |
| Sprint 1 | not-started | Add minimal Core 01/Core 04 closure and validate docs/static checks. | TODO: record owner-document diffs and validation run roots. |
| Sprint 2 | not-started | Add package scaffolding and static package assertions. | TODO: record package files and build evidence. |
| Sprint 3 | not-started | Add operator object-store init, structured readiness, and tests. | TODO: record backend/process validation. |
| Sprint 4 | not-started | Add Make-owned smoke target and package smoke evidence. | TODO: record smoke run root and deployable-shape evidence. |
| Sprint 5 | not-started | Update runbook/tracker evidence and final validation. | TODO: record final commands, skipped checks, and residual risk. |
