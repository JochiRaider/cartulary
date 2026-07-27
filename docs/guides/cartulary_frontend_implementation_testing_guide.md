---
title: Cartulary Frontend Implementation and Testing Guide
class: implementation-support guide
status: current
---

# Cartulary Frontend Implementation and Testing Guide

## 1. Authority and scope

This guide explains how frontend implementation evidence is represented and run. It
does not define product behavior, Core conformance, design-token membership, or
claim-bearing publication. Core 00 through Core 04, adopted subsystem NLSpecs, and
their reviewed machine verification contracts own those behaviors.

`docs/testing-harness-nlspec.md` owns catalog selection, runner adaptation,
scheduling, artifacts, evidence accounting, cleanup, and command behavior.
`docs/design.md` supplies design direction only. Core 05 owns any active
claim-publication boundary.

Retired delivery tables, generated coverage ledgers, guide digests, cumulative
joins, and frontend-only row accounting are not current execution or ownership
inputs.

## 2. Frontend ownership model

Frontend tests use the same owner catalog as backend, shell, and other browser
evidence:

- `tools/test_catalog_owner.json` registers active owners and their manifest paths.
- `tools/test_families/*.json` contains owner-qualified semantic families and rows.
- `contracts/verification/registry.json` and its owner contracts identify the
  adopted product or support postconditions each row verifies.
- `tools/execution_topology_manifest.json` owns runtime, resource, and fixture
  profiles. Rows reference profile IDs and do not embed topology or environment.

The primary owner is the owner of the verified postcondition. For cross-module
behavior it is the owner of the externally visible postcondition or primary durable
mutation. Platform and harness mechanism evidence belongs to the mechanism owner.
Other participating owners are collaborators. File location, runner, UI surface,
and maintainer identity are not ownership tie-breakers.

Every active frontend row declares:

- one immutable `owner_id`, `family_id`, and `row_id`;
- nonempty active `verification_ids`;
- one exact Vitest, Playwright, or registered shell selector;
- one evidence class and claim posture;
- runtime, resource, and fixture profiles;
- whether default `make check` selects the row.

## 3. Runner contracts

### 3.1 Vitest

A Vitest selector contains one repository-relative test file and one nonempty sorted
array of exact full test titles. Globs, regular expressions, title prefixes, and
file-pattern ownership are invalid. A title resolves to exactly one active row.

Unit and component tests should favor stable semantic assertions over visible-copy
or implementation-location assertions. Boundary-policy tests use support verification
contracts and must not claim product behavior by themselves.

### 3.2 Playwright

A Playwright selector contains one repository-relative file, project ID, stage,
stable scenario IDs, and matching diagnostic titles. The closed stages are:

- `webserver_backed`
- `stateful`
- `support`
- `visual`
- `accessibility`
- `measurement`

Scenario identity is semantic and must not contain a delivery milestone. Runtime
profile, fixture isolation, service ownership, reset policy, and resource locks come
from the referenced profiles.

### 3.3 Shell/static evidence

Shell rows name a stable command ID registered by the task surface. They never embed
raw shell, argv, executable paths, or row-local environment variables. Static and
security verification contracts name the exact public target that closes them.

## 4. Canonical commands

Use public Make targets from the repository root:

```bash
make task-guide ROLE=module-author OWNER=web.workbook
make explain-test-owner OWNER=web.workbook
make test-slice OWNER=web.workbook
make test-slice OWNER=web.workbook ROWS=web.workbook.interaction.some_row
make service-backed-test-slice OWNER=web.workbook
```

Omitted `ROWS` selects every active row owned by the requested owner. On the
service-backed command it selects every owned row whose runtime profile requires
managed services. `default_check` does not narrow either selection. Invalid,
duplicate, blank, unknown, cross-owner, and zero-row selections fail before setup.

Useful broad frontend gates remain:

```bash
make frontend-typecheck
make frontend-unit
make frontend-import-boundary-check
make lint-biome
make browser-e2e-webserver-backed
make browser-e2e-stateful
make browser-e2e-a11y
make browser-e2e-visual
make browser-e2e-measurement
```

These Make-owned targets resolve rows and runtime profiles first, create or
attach to exact isolated test-service sessions, and validate immutable v4 stack
evidence before Playwright workers start. They do not use the local development
database, bucket, Compose project, or object-store proxy. Package aliases are
valid only when they delegate to these Make targets without recursion. Raw
Playwright, Playwright UI mode, and IDE Playwright launchers do not establish
the attachment contract and are unsupported as executable evidence surfaces.
Use `make help` and `make help-all` for the current target inventory rather than
copying target lists into new guidance.

Run the narrow owner slice first, then the evidence-class gates generated for the
owner. Do not treat a passing broad target as proof that an explicit owner selection
ran unless retained row accounting records the exact selected inventory.

## 5. Accessibility, visual, and measurement evidence

Accessibility, visual, and measurement remain distinct evidence classes:

- accessibility rows require `browser-e2e-a11y`;
- visual rows require golden-digest validation and `browser-e2e-visual`;
- measurement rows require `browser-e2e-measurement` and remain informative unless
  an active Core 05 claim profile authorizes publication.

Visual fixture records use stable owner, fixture, and scenario IDs. They do not carry
guide paths as behavior. A path-only golden rename must preserve bytes and prove
before/after SHA-256 equality; never run a visual-update target merely to rename a
fixture. Pixel changes follow the visual golden maintenance guide as separate work.

Design and accessibility evidence is implementation/readiness evidence unless a
normative owner explicitly promotes a narrower requirement. A screenshot, design
token, or accessibility report does not independently create product behavior.

## 6. Evidence accounting and audit

Every selected row produces exactly one terminal record: `passed`, `failed`,
`infrastructure_failed`, `skipped_dependency`, `cancelled`, or
`skipped_authorized`. A passing owner invocation requires every selected row to pass
or have a valid unexpired authorization.

Owner artifacts preserve source snapshot, catalog, verification, selector, runtime,
resource, and fixture identity. Audits require explicit compatible retained roots:

```bash
make test-evidence-audit \
  OWNER=web.workbook \
  CHECK_RESULTS_DIR=<check-root> \
  BROWSER_SUPPORT_RESULTS_DIR=<support-root> \
  BROWSER_VISUAL_RESULTS_DIR=<visual-root> \
  BROWSER_A11Y_RESULTS_DIR=<a11y-root> \
  BROWSER_MEASUREMENT_RESULTS_DIR=<measurement-root>
```

Only roots required by the selected owner inventory are mandatory. Supplied valid
unnecessary roots are reported as unused. The audit never selects a newest run and
never mixes source, catalog, verification, or profile digests.

## 7. Implementation boundaries

- `apps/web` owns application composition and web behavior.
- `packages/grid-adapter` owns the shared grid integration boundary; application
  code does not import the underlying grid library directly.
- `packages/protocol-ts` and `packages/ui-contracts` generated roots are downstream
  artifacts and are never hand-edited.
- Runtime UI code consumes machine-owned design tokens and contract packages rather
  than parsing documentation.
- Frontend tests and validators may retain inert documentation references for human
  traceability, but must not open, stat, resolve, or hash those paths.

Grid evidence uses the following ownership split:

| Postcondition | Required implementation/evidence boundary |
| --- | --- |
| Stable identities, visible semantic positions, state precedence, callback payloads, and target rejection | Deterministic grid-adapter policy tests; the lightweight `./test-support` fake may share these policies and must retain consumer-contract parity. |
| Workbook mutation submission, reconciliation, messaging, and decoded clipboard dispatch | Application tests that explicitly mock `@cartulary/grid-adapter` with `@cartulary/grid-adapter/test-support`. |
| Non-virtualized DOM rendering needed to isolate a package-local unit postcondition | The immutable package-private DOM-unit binding, imported only by `packages/grid-adapter` test files. It is support evidence, not a production-path claim. |
| RDG callbacks and lifecycle, virtualization, frozen columns, scrolling, offscreen semantic focus/targeting, and vendor choreography | The production root binding in package-local integration tests or live browser rows. The fake and DOM-unit binding cannot verify these postconditions. |
| Accessibility and visual behavior | Their dedicated live browser evidence classes and owner rows. |
| DOM bounds or timing measurements | The measurement class only; results remain informative unless Core 05 separately authorizes publication. |

Production virtualization is an immutable root-component binding. Runtime code,
application tests, shared setup files, the root facade, and `./test-support` must not
import or expose the package-private DOM-unit binding. There is no mutable diagnostic
setter, environment switch, storage switch, or consumer prop.

Extension-gated frontend code consumes the explicit serving-epoch availability
projection returned by production APIs. It does not infer claims from route probes,
generated registry defaults, raw configuration, or mutable globals. Optional profile
assets load only after their exact contribution is available. Tests must cover
claimed and Base fallback behavior, claim loss, disposal of stale async responses,
and preservation of Base state. The Import Assistant remains lazy, while its
production-path browser evidence covers unit discovery, preview, mapping,
select/skip, warnings, apply progress, cancellation, partial outcomes, and result
navigation. Network Flow is a collaborator on its regression rows, not the owner of
a second import worker.

## 8. Change workflow

1. Identify the normative postcondition and current semantic owner.
2. Update or add the reviewed machine verification contract if needed.
3. Update the owner family manifest with an exact selector and profiles.
4. Run catalog/schema validation and the narrow owner slice.
5. Run the applicable frontend and browser evidence-class gates.
6. Regenerate owner-derived topology only through `make generate` and verify with
   `make generate-drift`.
7. Audit retained owner evidence when the change affects release closure.

A browser helper that prepares a virtualized grid row may scroll before the
mutating action. After the action begins, a focus or viewport postcondition
helper is an observer: it waits for the response's stable source record and
row-version floor, then reads focus, visibility, and scroll state without
focusing, scrolling, clicking, or dispatching input. Keep setup helpers and
postcondition observers separate so the test cannot manufacture continuity
that the application failed to restore.

A new runner requires an adopted NLSpec and runner-registry revision, closed selector
and result schemas, an allowlisted checked-in adapter, and positive/negative contract
fixtures. Dynamic runner plugins and executable loading are unsupported.

## 9. Acceptance criteria

Frontend implementation evidence is complete only when:

- every retained test has one semantic owner row and active verification reference;
- every selector resolves exactly once and overlaps no other active row;
- no frontend-only ownership namespace, guide digest, delivery-phase join, or
  unowned catch-all participates in execution;
- applicable type, unit, boundary, browser, accessibility, visual, measurement,
  security, and release gates are passing or exactly `not_applicable_zero_rows`;
- generated artifacts are reproducible and drift-clean;
- retained evidence uses current owner schemas and one compatible semantic identity;
- no executable frontend validator reads documentation as behavior.
