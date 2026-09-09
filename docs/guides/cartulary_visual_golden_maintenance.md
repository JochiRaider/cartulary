# Cartulary Visual Golden Maintenance Guide

**Status**: Implementation-support guide
**Authority**: Core 00-04 own product behavior. The testing harness NLSpec owns
harness conformance. This guide does not make visual snapshots claim-bearing
evidence or promote them into Base Profile conformance.

## Purpose and Owners

Committed Playwright screenshots are regression inputs for browser-rendered
workbook states. The authoritative visual row inventory is the set of active
`visual` Playwright rows in `tools/test_families/*.json`. The visual fixture
registry declares only the semantic fixtures and design projections it names;
it is not required to enumerate every active screenshot or golden. Do not copy
the registry into a guide, infer ownership from filenames, or derive row success
from test titles.

For pre-MVP browser inspection and design-discovery review, use
`docs/guides/cartulary_browser_design_readiness_workflow.md` before accepting a
refresh. `docs/design.md` supplies design direction only.

## Canonical Commands and Paths

- Validate committed goldens with `make browser-e2e-visual`.
- Refresh goldens only with `make browser-e2e-visual-update`.
- The Playwright source is `apps/web/e2e/workbook.visual.spec.ts`.
- Committed goldens live in
  `apps/web/e2e/workbook.visual.spec.ts-snapshots/`.
- `tools/frontend_visual_renderer_profile.json` pins the executable renderer;
  `tools/frontend_visual_golden_manifest.json` records that profile and every
  committed golden's SHA-256.
- Failed-run actual and diff artifacts live under the retained run root reported
  by the harness.
- Concept art and external bitmaps are inputs, not golden sources.

The update target and ordinary validation both use the harness-owned, pinned
Playwright renderer. The harness verifies the container image digest, platform,
Playwright and Chromium versions, font manifest, locale, device scale factor,
and color scheme before exposing the authenticated loopback endpoint to a visual
worker. The endpoint is private harness state, not a supported user override.
The two targets must otherwise use the same application server, viewport,
fixture lifecycle, and per-row accounting. An update may change snapshot bytes;
it may not bypass functional assertions or start an unowned application server.

## Capture and Fixture Preparation

Every page used by a visual scenario, including auxiliary user pages, installs
its local preference resource before navigating. Start with a null override and
assert the owning surface's resolved default; Timeline inherits compact density.
Use the resource to select each density variant. Account Settings presentation
writes delegate to the same route owner. Real persistence remains covered by the
Account Settings integration row. Never restore visual density through a shared
account request.
Keep account-menu long-label writes in the local profile fixture as well; later
captures must not inherit a presentation scenario's saved display name.

Prepare viewport, zoom, spacing and scenario state; wait for responsive layout
and vendored fonts; normalize dynamic text; settle layout; then establish focus
and the explicit scroll anchor. Center drawer controls in the visible drawer
scrollport and align panel starts, clamping to legal scroll offsets. Observe the
anchor and intended focus across three animation frames. Product focus-continuity
assertions remain observation-only. Preserve intentional workbook-grid framing.
No normalization, arbitrary pixel correction, sleep, retry, or tolerance change
may repair a mismatch after anchoring.

Capture intent v2 records the surface, declared and observed presentation, and
viewport. Registry v6 contains a profile for each exact golden path, including
variant viewport, zoom, and density. Reconciliation v3, browser group result v6,
and browser target result v4 reject old or mismatched current evidence. Null density means an application
shell without workbook density; missing observations and empty strings fail.

The browser stack v7 references a frontend build receipt v1 and its digest. The
preview serves that run's completed private artifact until its consumers exit.
A separate build may publish conventional output without changing active browser
inputs. Investigate resource failures and preparation-stage diagnostics before
interpreting a navigation timeout or a secondary cleanup failure. The restricted
traces and images do not become product conformance evidence.

## Accepted Refresh Triggers

Refresh a golden only when at least one of these is true:

- an adopted UI contract intentionally changed;
- the visual harness intentionally changed viewport, masking, scroll
  normalization, or screenshot scope;
- committed comparison behavior was intentionally enabled or corrected;
- a browser, dependency, platform, or vendored-font pin changed and its rendered
  effect was reviewed;
- the old golden contains a retired fixture identity and the replacement
  semantic identity has already passed the functional row;
- the old golden is otherwise stale relative to already-validated behavior.

Never refresh to conceal an unexplained layout change, failed functional
assertion, unstable text, missing data, runtime mismatch, or incomplete fixture.

## Required Refresh Record

Every refresh must record:

- the accepted trigger;
- every affected semantic owner row ID;
- every affected stable fixture ID;
- the changed golden filenames;
- whether viewport, browser zoom, masks, scroll normalization, or screenshot
  scope changed;
- the validation run root and visual review outcome.

The row list must come from the authored catalog. When a capture claims a
registered fixture, its fixture and golden paths must resolve exactly against the
registry from the same source snapshot. An active capture with no registry entry
is valid when its emitted capture intent, catalog row, scenario, Playwright
project, snapshot template, and committed golden reconcile exactly. Registry
absence alone is neither drift nor an orphan and never permits filename-derived
ownership.

Before moving, refreshing, deleting, or re-accounting a golden, review the
retained `cartulary.frontend_visual_reconciliation.v3` artifact from an ordinary
visual run. It must account for every active capture intent, every committed PNG,
every declared registry fixture, SHA-256, exact catalog/scenario/project identity,
and any declared non-Playwright consumer. Ambiguous mappings and active missing
goldens block mutation. Delete only a reconciled `orphan` with zero consumers;
retain an active nonregistry golden.

## Deterministic Capture Contract

Before capture:

- seed only deterministic fixture data;
- wait for application readiness and `document.fonts.ready`;
- verify that the active vendored Inter and JetBrains Mono faces loaded;
- normalize browser zoom and declared scroll anchors;
- mask only declared dynamic values such as generated record IDs, actor
  connection data, clocks, or cursors;
- assert that the intended fixture and surface state are present;
- use a selector crop only when the fixture registry declares selector scope.

For workbook-grid captures, declare horizontal normalization through
`GridVisualScrollState`: `left` means zero, `right` means the live maximum, and a
number means that exact clamped offset. Avoid historical pixel positions unless
the position itself is under test.

The default Timeline shell fixture is a fixed first-viewport capture. Full-page
captures, captures taller than the viewport, grid-adapter support specimens,
theme specimens, concept images, or demo/admin card stacks cannot satisfy it.

## Review and Acceptance

After `make browser-e2e-visual-update`:

1. Inspect every changed image and, where useful, its old/current/diff sheet.
2. Confirm that each pixel change follows from the accepted trigger.
3. Confirm that unexpected layout, typography, focus, clipping, overflow,
   responsive, loading, error, and component-state changes are absent.
4. Run `make browser-e2e-visual` without update mode.
5. Run fixture-registry, catalog, JSON-shape, generated-policy, and drift checks
   selected by the owner task guide.
6. Commit source, registry changes, goldens, and the refresh record together.

The update writes a complete candidate snapshot directory first. It promotes
that directory and its newly generated golden manifest only after every selected
group, reconciliation check, and schema check passes. If the update fails, the
tracked snapshot directory and manifest must remain byte-identical. After an
accepted renderer or golden refresh, require two fresh ordinary visual runs to
pass against the promoted manifest.

Update mode retains existing goldens that already satisfy the ordinary comparison
contract. Only missing or failing comparisons produce replacement candidates;
incidental byte differences within the existing tolerance do not justify churn.

If the visual validation fails before screenshot comparison, fix the functional
or infrastructure defect first. A successful refresh never substitutes for
per-row terminal evidence.

## Font and Artifact Traceability

A vendored font change is intentional only after reviewing its rendered effect.
The harness retains the font-manifest SHA-256 with screenshot artifacts so font
metric changes can be traced to the exact bundle. Retained actual/diff images are
diagnostic artifacts; committed Playwright outputs remain the only golden inputs.
