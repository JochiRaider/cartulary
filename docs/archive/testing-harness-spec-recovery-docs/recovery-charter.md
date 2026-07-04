---
doc_id: THR-S0-CHARTER
title: Testing Harness Recovery Charter
status: active
role: recovery-charter
---

# Testing Harness Recovery Charter

## Charter status

This charter initializes Sprint 0 for the testing harness reverse-specification
recovery. It records the recovery boundary, permitted write surface, prohibited
implementation changes, evidence labels, source-limit log location, and current
repository state.

The operator-provided recovery scope is full repository harness coverage:
all repository surfaces that affect test orchestration or validation behavior
are in scope for recovery inspection and later classification.

## Repository state

| Field | Value |
|---|---|
| Repository root | `/home/askahn/code/cartulary` |
| Canonical module path | `github.com/JochiRaider/cartulary` |
| Branch | `main...origin/main` |
| HEAD revision | `de68f8da3de87e383d37d332e8f17694e1fd1500` |
| Dirty state at S0 start | Clean; `git status --short --branch` printed only `## main...origin/main` |
| Runtime platform | `Linux DeskRip 6.6.87.2-microsoft-standard-WSL2 #1 SMP PREEMPT_DYNAMIC Thu Jun 5 18:30:46 UTC 2025 x86_64 x86_64 x86_64 GNU/Linux` |
| Current date | `2026-05-08` |
| Go runtime observed | `go version go1.26.3 linux/amd64` |
| Go baseline source | Repository procedure and `go.mod`/toolchain baseline: Go `1.26`, toolchain `go1.26.3` |
| Node runtime observed | `v24.15.0` |
| Node baseline source | `package.json` `engines.node`: `24.15.0` |
| pnpm runtime observed | `10.33.0` |
| Package manager source | `package.json` `packageManager`: `pnpm@10.33.0` |

## Authority and non-goals

The main project specification under `docs/spec/00_document_set_status_and_precedence.md`
through Core 04 owns product behavior. This recovery effort may describe how the
testing harness validates product behavior, but it must not redefine product
behavior.

The recovery effort treats implementation, tests, fixtures, CI configuration,
logs, reports, and local policy as evidence. They are not automatic proof of
future normative intent.

This recovery effort must not rewrite the harness. Future remediation belongs in
later roadmap artifacts, not in recovery execution.

## Permitted write locations

Recovery agents may write only under:

- `docs/testing-harness-spec-recovery-docs/**`

Writing outside that tree requires a separate explicit maintainer authorization.

## Prohibited edits

Recovery agents must not modify:

- Harness implementation behavior.
- Product implementation behavior.
- Test logic.
- Fixture, golden, seed, or snapshot contents.
- CI behavior or workflow configuration.
- Cleanup scripts or service lifecycle behavior.
- Generated Go or TypeScript outputs.
- `go.sum`, `pnpm-lock.yaml`, or other tool-managed lockfiles.
- Build, runtime, package-manager, or generated-artifact behavior.

## Evidence labels

Use these labels when recording recovered behavior:

| Label | Meaning |
|---|---|
| `observed` | Directly inspected in repository source, config, docs, tests, fixtures, CI, or committed artifacts. |
| `runtime_observed` | Observed by running a command during recovery. |
| `inferred` | Derived from multiple observed sources. |
| `assumed` | Temporary assumption pending evidence. |
| `intentional` | Behavior confirmed by governing spec, maintainer decision, or reliable project documentation. |
| `compatibility_only` | Behavior that appears accidental but external callers may depend on. |
| `accidental` | Behavior that exists without clear intent and should not become contract by default. |
| `contradiction` | Sources disagree. Do not pick a side without owner decision. |
| `maintainer_decision_required` | Normative closure requires owner authority. |
| `source_limit` | The agent could not inspect enough to decide. |

## Provisional harness boundary candidate list

The following surfaces are candidate in-scope harness surfaces. S1 must classify
them with evidence and may narrow, split, or reclassify them. This list is not a
final ownership decision.

| Boundary ID | Candidate surface | Initial role hypothesis | Evidence status | Notes |
|---|---|---|---|---|
| HB-0001 | `Makefile`, `tools/task_surface.generated.mk`, root `package.json`, workspace package scripts, `scripts/run-*`, `scripts/check-*`, `scripts/lib/**`, `scripts/ci/**` | Entrypoints and orchestration | `observed` | Candidate local, release, service-backed, and validation command surface. |
| HB-0002 | `internal/testutil/**`, `tools/testservices/**`, `apps/web/e2e/**`, `apps/web/src/*test*`, `packages/test-utils/**`, committed goldens, snapshots, fixtures, and seeds | Test support and fixtures | `observed` | Candidate reusable test harness, browser harness, and fixture surface. |
| HB-0003 | `tools/*manifest*.json`, `tools/phase*_test_map.json`, duration baselines, scheduler/resource registries, `docs/testing/*coverage_ledger.md` | Harness manifests and ledgers | `observed` | Candidate scheduler, phase, accounting, and duration-planning surface. |
| HB-0004 | `docker-compose.dev.yml`, `configs/dev/**`, Playwright config, Vitest/Vite/TypeScript/Biome configs | Service and environment inputs | `observed` | Candidate local-service and runner-configuration surface. |
| HB-0005 | `.cartulary/**`, `tmp/**`, `coverage/**`, `test-results/**`, Playwright reports, generated dist assets, local caches | Artifacts and scratch paths | `inferred` | Classify later as generated, temporary, diagnostic, committed, ignored, or external. Do not edit during S0. |
| HB-0006 | `.github/**` | CI workflow surface | `source_limit` | `.github/` was absent during S0 inspection; record as a source limit. |

## Initialized recovery artifacts

| Artifact | Path | Status |
|---|---|---|
| Recovery charter | `docs/testing-harness-spec-recovery-docs/recovery-charter.md` | Initialized by S0. |
| Source-limit log | `docs/testing-harness-spec-recovery-docs/source-limit-log.md` | Initialized by S0. |
| S0 handoff | `docs/testing-harness-spec-recovery-docs/handoffs/2026-05-08-s0-charter-and-setup.md` | Initialized by S0. |

## S1 readiness

S1 may begin from this charter, the source-limit log, and the updated sprint
plan without relying on transcript memory. S1 should treat the boundary list
above as a provisional discovery seed and should record every new gap in
`docs/testing-harness-spec-recovery-docs/source-limit-log.md`.
