---
doc_id: cartulary.testing_harness.ui_review.revision_plan
title: UI Review Harness — Revision and Production Readiness Plan
doc_type: revision_plan
status: TODO
authority_boundary: Human planning and execution tracker; iteration 1 is complete and iteration 2 is planned only. The adopted Testing Harness NLSpec remains runtime authority and product owners retain product authority.
---

# UI Review Harness — Revision and Production Readiness Plan

Subsequent preparation and diagnostic remediation is recorded in the
[remediation handoff](ui-review-preparation-remediation.md). Its coordinated v2
cutover supersedes the historical v1 examples below for current commands. The
adopted Testing Harness NLSpec remains the behavioral authority.

**Current iteration:** [Iteration 2 — production readiness](#18-iteration-2--production-readiness-plan),
planned on 2026-09-28 against `25806426bae985528ee14de9e2595316da93ba8f`.
Implementation has not started. The current request authorizes updating this
document only; implementation requires a later authorized task. Section 18 is the
active plan and tracker. Sections 1–17 preserve iteration 1's adopted decisions,
completed workstreams, and historical validation; their completion claims do not
establish readiness at the new baseline.

## 1. Status, objective, and normative convention

Sections 1–17 record the completed specification and implementation remediation for
the [Testing Harness NLSpec](../testing-harness-nlspec.md). It applies the behavioral
completeness, interface completeness, boundary completeness, conceptual fidelity,
and spec-economy criteria in [What an NLSpec Is](../research/nlspec-spec.md), version 0.2.2.
Its objective is a reproducible, inspectable UI review loop for developers and
agents, integrating browser control, screenshots, measured observations,
accessibility analysis, and existing harness evidence.

This artifact is the sole execution tracker. All six iteration-1 workstreams are complete;
their ordered checkpoints, failures, final acceptance matrix and handoff appear
below. The contracts were adopted as amendment `ui-review-api-1` in the owning
NLSpec before their implementation. The planning contract retained in Sections
2–16 explains the decisions; it does not supersede that adopted owner or create
product authority. Instructions quoted or described in source documents remain
source material, not additional authorization to execute unrelated workflows.

In this plan, MUST and MUST NOT are acceptance-bearing editorial or proposed
behavioral obligations. MAY identifies implementation freedom only where its
alternatives preserve the specified observable contract. There is no advisory
normative keyword. A proposed requirement becomes binding on runtime behavior
only through adoption in the owning NLSpec. `UIR-*` and `UIA-*` identifiers are
document-local planning references; executable registries MUST NOT consume them.

The revision MUST preserve the authority of Core 00–04 over product behavior,
Core 05 over claim publication, and the design owner over presentation direction.
Review findings MUST NOT redefine correct density, focus behavior, workflow,
authorization, colors, or visual-golden acceptance. Documentation MUST remain
outside executable inputs, source digests, generators, and verification evidence.

## 2. Baseline and gap disposition

The inspected baseline was branch `main`, commit `c4fe8d5e0`, with pre-existing
uncommitted product and harness changes. The revision author MUST recheck the
named seams before adopting this plan; these observations are planning evidence,
not a machine completeness registry or a current test result.

| Gap | Baseline evidence | Required disposition | Proposed contract |
| --- | --- | --- | --- |
| G1: Browser interaction lacks a stable Cartulary agent interface | Section 11 has interactive design-review preparation; `tools/harness/browser/design-review.mjs` emits an unversioned scenario index. Raw Playwright commands are conveniences under Sections 1 and 4.4. | Add versioned session, action, capture, result, and cleanup contracts through public Make commands. Reuse the existing lifecycle owner. | UIR-01–UIR-05 |
| G2: Images and structural observations are separate | Visual comparisons, selected geometry attachments, accessibility snapshots, and retained traces already exist in the browser tests. There is no adopted unified review-bundle contract. | Join exact source identities, original images, derived views, geometry, accessibility, and bounded runtime observations in one bundle. | UIR-06–UIR-10 |
| G3: Live development and canonical browser evidence need an explicit boundary | Sections 11.4 and 8 require sealed preview artifacts and pinned rendering for canonical visual evidence. `make dev` serves a changing development environment. | Define separate seeded, dev, and artifacts modes. Dev attachment is explicit and borrowed; its observations never attest a sealed build. | UIR-01, UIR-04, UIR-06 |
| G4: Automated accessibility breadth is incomplete as a tooling interface | `workbook.a11y.spec.ts` has custom keyboard, contrast, and geometry checks; inspected frontend manifests contain no axe integration. | Add an advisory axe channel with explicit incomplete/error semantics. Preserve owner-specific accessibility assertions. | UIR-09 |
| G5: Agent tool resolution depends on shell context | Make selects `tmp/node-runtime/bin`; ambient Node/pnpm were absent. ImageMagick and Pillow were available, while optional analysis tools were absent. | Resolve the core toolchain through pinned Make-owned tooling, independent of ambient PATH. Optional tools are outside core readiness. | UIR-12 |
| G6: Proposed browser data bundles have no adopted retention boundary | Section 15 distinguishes private runtime material, redacted retained evidence, and secret-bearing browser diagnostics. | Define private detail bundles and durable structural receipts separately; prohibit implicit publication or automatic promotion. | UIR-10 |
| G7: Third-party diagnostic formats are not stable interfaces | Section 19 leaves Playwright report/trace/image/geometry schemas outside current conformance. | Adopt only a Cartulary wrapper and supported adapter inputs. Preserve third-party format freedom outside that adapter. | UIR-06, UIR-12 |
| G8: Visual refresh contains an incompatible reconciliation reference | TH-HARNESS-REQ-255 adopts reconciliation v3 and rejects v2 as current evidence, but its refresh paragraph still requires reconciliation v2. | Replace that refresh reference with v3 and check every current producer/consumer reference editorially. No compatibility reader is introduced. | UIR-12 |

Relevant inspected inputs also include `apps/web/package.json`,
`apps/web/playwright.shared.config.ts`, `tools/frontend_visual_renderer_profile.json`,
`tools/frontend_visual_fixture_registry.json`, `tools/task_surface_owner.json`,
`tools/harness_schema_attachments.json`, and the `harness.browser` and
`harness.visual` verification contracts. File paths identify review seams, not
additional behavior authorities.

## 3. Required revision shape

### 3.1 Core and deferred scope

| Capability | Disposition in this revision | Boundary |
| --- | --- | --- |
| Pinned Playwright API browser adapter | Required, behind the Cartulary command interface | Use the repository's Playwright version; element references are Cartulary session-local observations. No separate CLI dependency. |
| Sharp image processing | Required for crops, contact sheets, and overlays | Derived images never replace originals or committed goldens. |
| Existing Playwright expected/actual/diff artifacts | Required read-only adapter | Accept only explicitly selected current-version artifacts; no newest-run inference. |
| Standalone image difference | Required exact RGBA comparison | Pixelmatch is deferred; there is no second configurable tolerance policy. |
| `@axe-core/playwright` | Required advisory analysis | Violations and incomplete checks do not manufacture product-row results. |
| Session-local HTML and JSON review | Required | Same bundle identity for human and agent consumers. |
| Live dev attachment | Required, explicit opt-in mode | Borrow the selected service; own only the review browser and review files. |
| Python `uv`, NumPy, scikit-image, OpenCV | Deferred optional analysis profile | Adoption requires locked dependencies, exact metric parameters, input limits, result schema, and failure/acceptance mappings. |
| Tesseract OCR | Deferred optional image-only analysis profile | Adoption MUST define language data, coordinate transforms, confidence handling, and missing-text behavior; OCR cannot certify secret removal. |
| FFmpeg and video recording | Deferred optional temporal-analysis profile | Adoption MUST define capture ownership, duration/frame/storage limits, timestamps, dropped frames, and private retention. |
| Chrome DevTools MCP or Playwright MCP | Deferred alternate browser adapters | Adoption MUST prove the same session, source, serialization, cleanup, and action-result contracts. |
| External vision API, hosted image review, automatic uploads | Excluded | The local agent can inspect returned private files. No new API dependency or publication path is introduced. |

The deferred rows are explicit scope decisions, not unfinished core behavior.
Neither `doctor` nor `bootstrap` MUST require those tools for the core profile.
The revision MUST NOT introduce another Go browser driver or deployable executable.
Existing Go fixture and service owners retain backend lifecycle responsibilities.

### 3.2 Section-by-section editorial operations

The editor MUST apply this table as one internally consistent amendment. Existing
requirement IDs retain their meaning. New requirements receive unused IDs within
their owning section's existing range; the editor MUST check collisions before
assignment and update `Verified by` references and Section 17.1 together.

| Harness section | Required edit | Plan source |
| --- | --- | --- |
| 1–2: authority and purpose | Add the helper-only UI review boundary; distinguish session operation success, diagnostic findings, product assertions, and publication. Preserve v3 graph/catalog semantics. | Sections 1, 3, 11 |
| 3: terminology | Define review session, mode, observation epoch, bundle, private artifact, structural receipt, binding, and comparison compatibility. | Sections 4, 6, 8 |
| 4.3–4.6: public surface and wrappers | Add the seven commands, command IDs, input/output/artifact/cleanup behavior, runtime dependency profile, and rejected direct-driver equivalence. | Section 5 |
| 5.3–5.5: configuration | Add only the declared inputs and their source, omission, blank, bound, and forwarding rules. | Section 5.2 |
| 6: identity | Define session and operation identities and exact source joins without creating row evidence. | Sections 6, 8 |
| 7: output | Add command-response and retained-receipt projections; transient private paths are absent from retained output. | Sections 5.3, 8.1 |
| 8: schemas | Add the schemas and field contracts below, bundle integrity rules, adapter compatibility, and raster algorithms. Repair the reconciliation reference. | Sections 7–10, 12 |
| 9: failures | Add context-specific mappings to existing classes/reasons; no new generic failure taxonomy. | Section 11 |
| 10: resources | Declare host activity, browser capacity, serial session access, and no evidence-result caching for review commands. | Section 6 |
| 11: lifecycle | Add a numbered UI review subsection next to interactive design review. Reuse its seeded preparation owner; specify borrowed dev and offline modes. | Sections 4, 6 |
| 12: test-only routes | State that the revision adds no product or test HTTP route and grants no authentication bypass. | Sections 4, 7 |
| 13: cleanup | Close interruption, partial acquisition, browser death, stale ownership, idempotent stop, and private-artifact removal. | Section 6 |
| 14: platform/toolchain | Retain Linux/WSL2 support; define pinned core dependencies and explicit host-browser diagnostics. | Section 12 |
| 15: privacy | Add the exact private/retained disposition and escaped local report rules. Preserve canonical browser diagnostics' existing rules. | Section 10 |
| 16: product integration | Keep UI review outside product and release accounting; preserve virtualized identity, presentation owners, and manual accessibility obligations. | Sections 7, 9, 11 |
| 17–17.1: acceptance | Incorporate the acceptance matrix and human requirement-to-acceptance traceability. | Sections 14–15 |
| 18–19: non-normative references and exclusions | Link rationale and tool documentation; narrow the future-schema exclusion only for the newly adopted Cartulary wrapper. List deferred profiles explicitly. | Sections 3.1, 12, 16 |

The editor MUST keep defaults, algorithms, failure consequences, and acceptance
obligations in Sections 1–17 of the revised NLSpec. This plan and third-party
documentation MUST NOT remain necessary to implement the adopted contract.

## 4. Modes, ownership, and source claims

**UIR-01 — Review modes.** A session has exactly one immutable mode:

| Mode | Selection and default | Application resources | Browser | Source claim | Allowed work |
| --- | --- | --- | --- | --- | --- |
| `seeded` | Omitted `UI_MODE` selects this mode. | Fresh owned review suite, database, object namespace, and sealed frontend through the existing review lifecycle. | Review-owned isolated context and browser process. | `sealed_review`: exact source snapshot, frontend receipt, runtime profile, and observed browser identity. | All declared browser actions, capture, analysis, report. |
| `dev` | Explicit `UI_MODE=dev` and `UI_ORIGIN`. | Borrow the exact loopback origin; no create/reset/migrate/stop of its service or data. | Review-owned isolated context and browser process. | `live_unattested`: served source/build identity is absent. Workspace digest identifies inspected source only. | All declared browser actions, capture, analysis, report. |
| `artifacts` | Explicit `UI_MODE=artifacts`. | None. | None; no browser or Docker prerequisite. | Per imported artifact; no new rendered-build claim. | Import via capture request, analysis, report. Browser actions fail before execution. |

Seeded sessions MUST validate the current browser-stack attachment and build
receipt before seeding or navigation. `REVIEW_PROFILE` retains exactly the two
adopted profile values and its existing seeded default. It is invalid in the other
modes. The new commands MUST NOT infer a mode from an existing server, URL,
descriptor, environment variable, Docker state, or installed browser.

Dev origin syntax is exactly `http://127.0.0.1:<port>` or
`http://[::1]:<port>`, with integer port 1–65535 and no userinfo, path other than
optional `/`, query, or fragment. Normalize the optional trailing slash away.
Hostnames, redirects to another origin, and non-loopback origins are rejected.
Failure to reach the selected origin does not permit switching origin or mode.
Dev mode MUST NOT set harness-only server flags or enable test routes.

The default browser context is headless, locale `en-US`, device scale factor 1,
viewport 1440×900 CSS pixels, and browser color scheme `light`. These are review
mechanics, not the product theme. Seeded scenario presentation comes from the
existing product-owned preference fixture; dev captures observe actual presentation.
Review capture MUST NOT switch product theme or persist account settings implicitly.

Browser review uses the host browser installed for the pinned repo Playwright.
It does not attach the private canonical visual-renderer endpoint. A host review
screenshot MUST NOT be labeled canonical or used to refresh a golden. Canonical
comparison and refresh remain exclusive to their existing public targets.

## 5. Commands, inputs, and output projections

### 5.1 Public command registry

**UIR-02 — Commands.** Add exactly these helper-only bindings. The command ID is
`cartulary.harness.command.` followed by the listed suffix. All are excluded from
product, default verification, CI, release, and owner-slice selection. Tests of
their contracts can be selected by existing harness verification routing.

| Make target | Command-ID suffix | Required target inputs | Other target inputs | Result | Side effects |
| --- | --- | --- | --- | --- | --- |
| `ui-review` | `ui_review.v1` | None | `UI_MODE`, `REVIEW_PROFILE`, `UI_ORIGIN` | Interactive session, retained locator and lifecycle receipt | Acquire mode-specific resources; stay alive until stop, signal, failure, or lifetime expiry. |
| `ui-review-status` | `ui_review_status.v1` | `UI_SESSION` | None | One command result | Read exact session state; no acquisition, touch, reset, or liveness extension. |
| `ui-browser` | `ui_browser.v1` | `UI_SESSION`, `UI_REQUEST` | None | One command result and private observation | Execute exactly one declared action. |
| `ui-capture` | `ui_capture.v1` | `UI_SESSION`, `UI_REQUEST` | None | One command result and immutable private bundle | Capture current page or import exact existing artifacts. |
| `ui-analyze` | `ui_analyze.v1` | `UI_SESSION`, `UI_REQUEST` | None | One command result and new immutable private bundle | Derive selected analyses from exact inputs. |
| `ui-review-report` | `ui_review_report.v1` | `UI_SESSION`, `UI_BUNDLE` | None | One command result and private HTML | Render the exact bundle; no server startup or automatic browser launch. |
| `ui-review-stop` | `ui_review_stop.v1` | `UI_SESSION` | None | One command result and terminal receipt | Stop only resources owned by this session. |

The existing `browser-design-review` and smoke command retain their v1 invocation
and output contracts. `ui-review` reuses their preparation and teardown owner; it
MUST NOT copy a second implementation of seeding, service startup, or cleanup.
No raw Playwright, arbitrary JavaScript, shell argument string, alternate CDP URL,
or user-supplied canonical-renderer endpoint is a new public binding.

### 5.2 Closed input contract

For the following target-local inputs, `binding=make_variable`, sources are
`make_command_line` and `internal_default` where a default exists,
`empty_string=invalid`, `invalid_reason=usage_error`, and `child_forwarding=argv`.
Inherited environment variables of these names are ignored. Undeclared command-line
inputs fail before acquisition. Existing global input contracts still apply.
`REVIEW_PROFILE` alone keeps its existing allowed sources and normalization.

| Input | Targets | Type and normalization | Default / omission | Bounds and cross-field rules | Retained summary |
| --- | --- | --- | --- | --- | --- |
| `UI_MODE` | `ui-review` | Enum; trim | `seeded` | `seeded`, `dev`, `artifacts` only | Value |
| `REVIEW_PROFILE` | `ui-review` | Existing adopted enum contract | Existing default in seeded mode; absent otherwise | Explicit use outside seeded mode is invalid | Value when applicable |
| `UI_ORIGIN` | `ui-review` | URL; trim, normalize as Section 4 | Required for dev; absent otherwise | Explicit use outside dev mode is invalid | No URL; mode only |
| `UI_SESSION` | Every finite command | Path token | Required | Exact retained session-locator file; Section 6 path checks | Opaque session ID only |
| `UI_REQUEST` | Browser/capture/analyze | Path token | Required | Regular non-symlink owner-only UTF-8 JSON file, 1–65536 bytes; snapshot bytes once before validation | No path or request content |
| `UI_BUNDLE` | Report | ID; trim | Required | Exact session-local bundle ID defined in Section 8; no `latest` alias | ID only |

Request files MUST be outside documentation roots and Markdown, validated before
opening, and passed without shell evaluation. Their contents can include typed
text; they are private caller material. A command borrows its request file and
MUST NOT delete or rewrite it. The copied in-memory request and any private staging
copy die after the operation. Duplicate JSON object keys, invalid UTF-8, non-finite
numbers, unknown fields, null where not allowed, and unsupported schemas fail
before page mutation or image decoding.

### 5.3 Output classes

`ui-review` uses `interactive_raw` and rejects machine mode before acquisition.
Its ready output identifies the exact retained session-locator path. It MUST NOT
print credentials, a private endpoint, storage state, or request content.

The six finite commands use `summary_with_artifacts` with stable machine stdout
schema `cartulary.ui_review_command_result.v1`. Machine output is one JSON object
plus LF; stderr is empty after the wrapper starts. Human modes obey Section 7's
existing budgets. Browser text, images, raw logs, and full accessibility trees
MUST NOT be streamed into ordinary stdout.

The transient command result and durable receipt are distinct schemas. A
successful content-producing command returns explicit absolute private file paths
for the local agent/viewer. Those paths MUST NOT enter retained logs, receipts,
OTLP, or aggregate output. The public wrapper MUST project the receipt before
retention instead of retaining a byte copy of transient stdout. Raw child output
remains private and redacted according to Section 15.

## 6. Session identity, sequencing, and lifecycle

**UIR-03 — Identity.** A session ID is `uireview-` plus 32 lowercase hexadecimal
digits generated from 128 bits of cryptographic randomness. Collisions are
configuration failures; an existing session is never replaced. Session operations
receive sequential positive integers starting at 1 under the session controller.
An observation epoch starts at 0 and increases once before each browser action
other than a pure snapshot. An action that fails after admission still consumes
its epoch. Capture and analysis do not silently repeat an action.

The adapter validates and resolves a request's target against the old epoch before
advancing it. The response reports the new epoch. It never invalidates a supplied
reference before resolving the action that consumes it. An action deadline begins
after lock acquisition; the lock has its separate five-second budget. After an
action timeout, only snapshot or stop is admitted until a fresh snapshot
reestablishes observations; the action's product effect can be uncertain.

The session locator is `<run_root>/ui-review/session.json`, using existing result
root and run-ID normalization. It contains exactly `schema_id`, `session_id`,
`run_id`, `mode`, `state`, `created_at`, `updated_at`, and `terminal_receipt`.
The schema is `cartulary.ui_review_session.v1`; timestamps use Section 6.4's UTC
format; `terminal_receipt` is null before terminal state and a run-relative
digested reference afterward. Its state changes by secure atomic replacement.
Private controller handles and origin/credential details are absent.

The controller resolves the exact session ID through its owned private lease.
Transport, internal registry layout, and process factoring are implementation
choices. Resolution MUST NOT scan for a newest session or trust a PID alone.
Cross-user, replaced, symlinked, malformed, incompatible, or differently owned
locators fail before connection. A terminal locator remains readable for status
and idempotent stop after private material has been removed.

**UIR-04 — Lifecycle.** The state set and transitions are closed:

| State | Event | Next state and required effect |
| --- | --- | --- |
| `preparing` | All mode-specific readiness predicates hold | `ready`; publish the locator before admitting commands. |
| `preparing` | Preparation fails, signal, or stop | `stopping`; preserve primary cause and release every acquired resource. |
| `ready` | Valid finite operation obtains session lock | `busy`; assign operation ID; execute once. |
| `busy` | Operation succeeds or fails without losing session integrity | `ready`; publish its result/receipt; release lock. |
| `busy` | Browser/controller integrity lost | `stopping`; reject further operations and preserve cause. |
| `ready` or `busy` | Stop, SIGINT, SIGTERM, owner death, or lifetime expiry | `stopping`; reject new work, cancel/reap in-flight children, then clean. |
| `stopping` | All owned cleanup succeeds and no non-cleanup failure exists | `closed`; publish terminal receipt. |
| `stopping` | Primary failure or cleanup failure exists | `failed`; publish primary and secondary normalized failures. |
| `closed` or `failed` | Status or stop | Same state; return existing cleanup outcome without touching other resources. |
| `closed` or `failed` | Browser/capture/analyze/report | Reject as `configuration_error`; never revive the session. |

At most one operation that accesses page state or creates a bundle runs per
session. A competing operation waits at most five seconds for the session lock,
then returns `infra/resource_conflict` without admission. Status does not take
that exclusive lock. Stop has cancellation priority and cannot wait behind a
stalled capture. Concurrent sessions have separate contexts, files, and leases.

| Deadline or bound | Value | Consequence |
| --- | --- | --- |
| Session lifetime after ready | 8 hours, monotonic, not extended by requests | Stop and clean; normal closure if cleanup succeeds. |
| Dev-origin readiness | 30 seconds; probe every 250 ms | `service_readiness_timeout`; do not stop borrowed service. |
| Browser action | 10 seconds; navigation 30 seconds | `timeout_failure`; no replay; reobserve before another mutation. |
| Capture, including settling and axe when requested | 30 seconds | No successful bundle; preserve bounded stage diagnostic. |
| Offline analysis or report | 30 seconds each | Discard unpublished outputs; return `timeout_failure`. |
| Stop | Existing owned-resource teardown deadlines, each applied once | Preserve earlier cause; cleanup-only failure is `cleanup_error`. |

Seeded preparation inherits the exact prerequisite and service deadlines already
owned by Section 11; the new session lifetime starts after preparation completes.
There is no automatic action retry, capture retry, fallback browser, or background
fixture reset. Browser death ends a browser-bearing session; artifact mode does
not acquire browser resources.

Browser-bearing sessions hold declared shared `host_activity` and browser
capacity for their lifetime. Image analysis holds shared `host_activity` while
executing and releases it afterward; a waiting quiet measurement retains existing
fairness. Resource acquisition cannot bypass the shared host arbiter merely
because the command is helper-only. Review outputs and results are never reusable
verification cache entries. Existing sealed frontend build reuse remains governed
by its current receipt/cache contract.

Cleanup order is: stop admission, cancel/reap operation children, stop owned browser
and close handles, release owned application leases, delete private artifacts and
controller material, then publish the structural terminal receipt and scan retained
boundaries. Borrowed dev services and caller input files are never deleted. Reaper
and stale cleanup use the existing exact-ownership/parent-death rules; age alone
never authorizes stopping another process or deleting another session.

## 7. Browser actions and capture preparation

### 7.1 Action request

**UIR-05 — Browser action interface.** The request is a closed object with exactly
`schema_id="cartulary.ui_review_action.v1"`, `expected_epoch`, `action`, and
`parameters`. `expected_epoch` is a nonnegative integer. A mismatch fails with
`configuration_error` before admission; callers MUST take a new snapshot after
another actor changes the session. `parameters` is the closed variant selected
by `action` below. A browser adapter MUST translate these operations to the
pinned Playwright API adapter and MUST NOT expose the driver's arbitrary-code or
arbitrary-endpoint entry points.

| Action | Exact parameters | Observable behavior |
| --- | --- | --- |
| `snapshot` | `{}` | Observe the current page; return epoch and private accessibility/element observations without changing viewport, scroll, or focus. |
| `navigate` | `path: string` | Navigate within the selected origin to a path starting with one `/`; reject `//`, backslash, fragment, credentials, and origin changes. Length 1–2048 UTF-8 bytes. Query text remains private. |
| `click` | `target: Target` | Perform one primary-button click after Playwright actionability checks. No force option. |
| `fill` | `target: Target`, `text: string` | Fill one editable target; allow empty text; maximum 16384 UTF-8 bytes. The text is never retained. |
| `press` | `target: Target` or null, `key: Key` | Send one supported key/chord to the target, or current focused element when null. |
| `select` | `target: Target`, `values: string[]` | Select 1–32 exact option values, each at most 1024 UTF-8 bytes, no duplicates. |
| `scroll` | `target: Target` or null, `x: integer`, `y: integer` | Set absolute CSS-pixel offsets in the target or document, clamped by the browser to legal offsets; return observed offsets. Range 0–1000000. |
| `resize` | `width: integer`, `height: integer` | Set viewport; each dimension 320–3840 inclusive; maximum area 8294400 CSS pixels. |
| `focus` | `target: Target` | Focus the uniquely resolved focusable target and return the observed active-element identity. |
| `authenticate` | `actor: enum` | Seeded mode only; actors `admin`, `editor`, `viewer`, `empty` resolve through the private seed owner. Use ordinary application authentication; no credential output or authorization bypass. |

`Target` is a tagged union: `{kind:"test_id", value:string}`,
`{kind:"role", role:string, name:string}`, or
`{kind:"element_ref", value:string, epoch:integer}`. Test IDs are exact stable
test-ID values through the repo's configured attribute (`data-testid` by default);
roles use Playwright's pinned supported ARIA-role enumeration and exact accessible-name
matching. String fields are nonempty and at most 1024 UTF-8 bytes. Arbitrary CSS,
XPath, regular expressions, DOM source, and JavaScript expressions are unsupported.
Zero or multiple matches fail without selecting the first match. Element references
are private, session-local, and valid only for their exact epoch and current document;
navigation or a new browser action invalidates them.

`Key` is one of `Tab`, `Shift+Tab`, `Enter`, `Escape`, `Space`, `ArrowUp`,
`ArrowDown`, `ArrowLeft`, `ArrowRight`, `Home`, `End`, `PageUp`, `PageDown`,
`Backspace`, `Delete`, `Control+A`, `Control+C`, or `Control+V`. Clipboard content
remains private. Other keys require a later interface revision. Uploads, downloads,
new tabs, popup control, network mocking, browser storage mutation, and direct API
seeding are not v1 actions. An unexpected dialog or new tab fails the action;
the adapter does not accept a dialog, switch tab, or retry implicitly.

Dev authentication occurs through the same explicit UI actions as ordinary use;
the review controller MUST NOT read an existing user browser profile or borrow
ambient cookies. Seeded authentication is fixture preparation only. It advances
the epoch and clears the prior private context before establishing the requested
actor, so roles never share storage accidentally.

### 7.2 Capture request and source variants

**UIR-06 — Capture interface.** A capture request is a closed tagged union with
`schema_id="cartulary.ui_review_capture_request.v1"`, `source`, and the fields
selected below. No field is inherited from a previous request.

| Source | Fields | Defaults and meaning |
| --- | --- | --- |
| `page` | `expected_epoch`, `binding`, `scope`, `targets`, `include_axe` | Browser modes only. `expected_epoch` required; `binding` defaults null; `scope` defaults `{kind:"viewport"}`; `targets` defaults `[]`; `include_axe` defaults true. |
| `canonical_visual` | `run_root`, `capture_id` | Both required; import the exact current reconciliation/capture identity and its declared artifact references. No browser work and no inference from filenames. |
| `image` | `path` | Required local PNG path; class `reference_image`; renderer, scenario, and served-source claims are null. |

`scope` is `{kind:"viewport"}`, `{kind:"element",target:Target}`, or
`{kind:"region",x:integer,y:integer,width:integer,height:integer}`. Region
coordinates are screenshot pixels, with nonnegative origin and positive dimensions,
entirely inside the viewport screenshot. Element scope is the outward-rounded
bounding rectangle of the visible viewport intersection; an empty intersection is
an error. Full-document stitching is excluded because virtualization does not make
offscreen records observable. `targets` contains at most 64 unique Targets.

`binding` is null or the closed object `{owner_id,row_id,scenario_id,capture_id}`
with nonempty exact identifiers from current catalog/capture projections. It
records intended context, not execution of that row. Seeded page binding MUST
resolve to the same served source and runtime profile; dev page binding is invalid.
A null binding is a valid exploratory observation. Declared fixture IDs are joined
from the registry when present; registry absence does not invent or require one.

Image files MUST be local non-symlink regular files. Explicit run roots use Section
6's containment and compatibility rules. The input boundary rejects documentation
and Markdown paths before filesystem access, including image files under `docs/`.
Manual design references MUST first be supplied outside restricted executable-input
roots; the command does not copy from or inspect a restricted root. Input files are
borrowed and never modified. Remote URLs, data URLs, directory globbing, and implicit
artifact discovery are unsupported.

### 7.3 Preparation and observation algorithm

A page capture MUST execute these steps in order under one session lock:

1. Validate request, epoch, source identity, budget, and target resolution before
   capturing. Apply no navigation, authentication, preference write, scroll,
   focus change, masking, or viewport normalization implicitly.
2. Wait for `document.fonts.ready` and record the active expected font faces. A
   bound seeded scenario additionally verifies the readiness and presentation
   declarations of its current owner projections. This imports the existing
   capture preparation contract; it does not duplicate the fixture registry.
3. Observe viewport, page URL privately, document identity, epoch, active element,
   target bounds, selected scroll containers, and presentation values for three
   consecutive animation frames. Rectangles and offsets are compared at 1/64 CSS
   pixel precision. Changes reset the consecutive-frame count within the same
   deadline; no additional capture attempt is created.
4. Collect DOM/ARIA observations and, when selected, axe observations. Capture the
   screenshot. Recheck the document, epoch, viewport, target rectangles, and scroll
   offsets immediately afterward. A mismatch discards the candidate bundle and
   reports unstable capture as `artifact_error`; it does not realign pixels.
5. Record monotonic observation start/end, native pixel dimensions, and the exact
   coordinate transform. Validate every component and atomically publish a bundle.

Dynamic pixel content can remain in an unbound review screenshot. Geometry stability
does not assert pixel stability or application idleness. A bound canonical visual
test keeps its existing stricter preparation and masks; this helper cannot relax
those rules or grant a visual-test pass. Live dev captures record the document's
observed reload generation but do not claim to detect every in-place HMR update.
Their source is always `live_unattested`.

Virtualized targets MUST be explicitly revealed by a browser action before capture.
The observer MUST report only rendered nodes; an absent row is not evidence that a
record does not exist. Pixel coordinates, visual order, and DOM indexes MUST NOT
be converted into product record identity.

## 8. Stable bundle and result interfaces

### 8.1 Schema inventory and common types

**UIR-07 — Wrapper schemas.** Add the following schema attachments. Every JSON
object is closed; unknown fields are rejected. All integer counts are nonnegative
unless a stricter bound is stated. All JSON strings use UTF-8. IDs and digests are
not nullable unless explicitly stated. Arrays that represent sets are unique and
ASCII-sorted; observations retain sequence order. The source NLSpec owns these
rules; schema files are projections.

| Schema ID | Role and lifetime |
| --- | --- |
| `cartulary.ui_review_session.v1` | Retained mutable locator from Section 6. |
| `cartulary.ui_review_action.v1` | Private caller request from Section 7.1. |
| `cartulary.ui_review_capture_request.v1` | Private caller request from Section 7.2. |
| `cartulary.ui_review_analysis_request.v1` | Private caller request from Section 9. |
| `cartulary.ui_review_command_result.v1` | Transient finite-command stdout/result; contains local private references. |
| `cartulary.ui_review_receipt.v1` | Immutable retained structural operation or terminal receipt. |
| `cartulary.ui_review_bundle.v1` | Private immutable observations, analyses, and artifact references. |
| `cartulary.ui_review_observations.v1` | Private DOM, accessibility, axe, console, and network component. |

The schema path for each is `tools/schemas/<schema_id>.schema.json` and the schema
ID is the table's literal value, not a URI inferred from the filename. Add all to
the authored schema-attachment registry with `validation=json-shape-check`.
Classify action/capture/analysis requests as `owner_input`, observation components
as `shared_component`, and session/result/receipt/bundle schemas as
`public_contract`. Classification describes the interface, not permission to
retain or publish its data; private bundles remain private.

`Digest` is exactly 64 lowercase SHA-256 hexadecimal characters. `ArtifactRef`
contains exactly `path`, `sha256`, `bytes`, and `media_type`: a nonempty normalized
POSIX-relative path under its declared containing root, a Digest, a byte count,
and one of `image/png`, `application/json`, `text/html`, `text/plain`,
`application/zip`. Absolute paths, `..`, backslashes, NUL, symlink traversal, and
duplicate paths are invalid. Existing canonical source refs retain their adopted
schemas and are translated only at the adapter boundary.

`Failure` contains exactly `failure_class`, `failure_reason`, and
`diagnostic_code`. Class/reason use Section 9's existing pairings. Diagnostic code
is a closed token from Section 11. There is no arbitrary retained message string.
`PrivateRef` is `{kind, absolute_path}` with kind `bundle`, `image`, `observations`,
or `report`; the path is validated beneath this session's private root.

| Command-result field | Type and rule |
| --- | --- |
| `schema_id` | Literal command-result schema ID. |
| `command_id` | Exact invoked ID from Section 5.1. |
| `session_id` | Session ID, or null only when input validation cannot identify a session. |
| `operation_id` | Positive integer after admission; null for status, idempotent terminal stop, or pre-admission rejection. |
| `state` | Session state; null only when no validated locator is available. |
| `epoch` | Current nonnegative epoch; null in artifact mode or without a live browser observation. |
| `status` | `ok` or `error`; describes command execution only. |
| `failures` | Empty on `ok`; otherwise 1–32 distinct Failures in primary-failure order; repeated identical triples are deduplicated. |
| `receipt` | Digested run-relative ArtifactRef; null for live status, pre-admission rejection, or failed safe receipt publication. |
| `bundle_id` | Exact produced/selected bundle ID or null. |
| `private_refs` | 0–8 PrivateRefs; empty on error or terminal state. |

Status is observational: it returns `status=ok` when it successfully reads a
failed session, with `state=failed`; terminal lifecycle failures are available
through the terminal receipt. Stop reports the terminal lifecycle outcome, so a
previous cleanup failure is not converted into success by a repeated stop.

A receipt contains exactly `schema_id`, `command_id`, `session_id`, `operation_id`,
`mode`, `state`, `status`, `started_at`, `finished_at`, `duration_ms`, `failures`,
`counts`, `bundle_id`, and `cleanup`. Counts are the closed object
`{images,observed_elements,axe_violations,axe_incomplete,console_errors,failed_requests}`.
`cleanup` is `not_terminal`, `complete`, or `failed`. Status, identity, failure,
and null rules match the result. Terminal receipts have command ID `ui_review.v1`
under the full prefix, operation ID null, and counts summed over successful
observations, not inferred from absent files. Receipts contain no page text,
private artifact digests/paths, URLs, account identity, selectors, or credential data.

Validation errors that occur before identity allocation emit the closed error
result with null identity/receipt fields and create no files. Other receipts are
written at `<run_root>/ui-review/operations/<operation_id>/receipt.json`;
terminal receipt is `<run_root>/ui-review/terminal.json`. Status returns the latest
explicit terminal receipt when terminal; while live its receipt is null and it
does not create a synthetic operation. This live-status null is an explicit
additional null case for `receipt`.

Stop's result references the terminal receipt owned by `ui-review`; that receipt's
command ID identifies the session lifecycle rather than the stop wrapper. A
repeated stop returns the same terminal receipt digest. Status reads the exact
locator's terminal reference, never a directory's newest receipt. Failure to
publish a terminal receipt leaves the session non-successful; it does not invent
proof of cleanup. Counts in a terminal receipt are operation totals, so repeated
successful captures are counted as separate observations, not unique UI states.

### 8.2 Bundle contract

A bundle ID is `bundle-` plus the positive decimal operation ID, without leading
zeros. It is unique within its session. A bundle contains exactly:

| Field | Type and semantics |
| --- | --- |
| `schema_id`, `bundle_id`, `session_id` | Exact schema and identities. |
| `classification` | Always `private_diagnostic`; never a conformance or publication class. |
| `tool_profile` | Closed effective tool-version/digest record defined below; describes this bundle's producer, not an imported image's original producer. |
| `parents` | 0–2 exact `{session_id,bundle_id,sha256}` references to input bundle manifests; captures have none. |
| `source` | Tagged source record defined below. |
| `observation` | Null for image-only input; otherwise the closed observation context below. |
| `binding` | Null or exact current catalog/capture binding from Section 7.2, plus resolved `fixture_ids` array, which can be empty. |
| `components` | Closed object with nullable ArtifactRefs `original`, `expected`, `actual`, `diff`, `observations`, `trace`; no file is claimed when null. |
| `derived` | Ordered array of records `{kind,ref,source_refs,rectangle}`; kinds `crop`, `overlay`, `contact_sheet`, `exact_diff`; source_refs is a nonempty ordered ArtifactRef array; rectangle null except for crop. |
| `analysis` | Null or closed analysis result from Section 9. |
| `limitations` | Unique sorted tokens: `live_unattested`, `no_actual`, `no_dom`, `no_axe`, `no_trace`, `reference_only`, `rendered_nodes_only`, `truncated_console`, `truncated_network`, `cross_source_comparison`. |

`tool_profile` contains exactly `pins_sha256`, `lock_sha256`, `node_version`,
`playwright_version`, `sharp_version`, and `axe_version`.
Digests identify the qualified toolchain projection and package lock bytes;
versions are nonempty exact installed-version strings. Core package versions
remain available in artifacts mode even though no browser executable is required.
An analysis bundle records its own producing profile and preserves each parent's
profile through the immutable parent reference.

`source` has exactly `kind`, `workspace_digest`, `served_source_digest`,
`frontend_receipt`, `renderer_profile_id`, `browser_version`, `runtime_profile_id`,
and `import_ref`. Source schemas use a closed variant per kind; non-applicable
fields are absent rather than nullable placeholders. Kind is `sealed_review`, `live_unattested`, `canonical_visual`,
or `reference_image`. Seeded source has non-null workspace/served digests, a
digested frontend receipt, browser version and runtime profile; its canonical
renderer profile is null. Dev source has non-null workspace digest and browser
version; served digest, frontend receipt, renderer profile, runtime profile, and
import ref are null. Canonical source is copied from validated current evidence,
with an import ref to exact reconciliation bytes. Reference-image source has only
kind and import ref non-null. An import ref is exactly
`{input_path,input_sha256,metadata}`. Input path is the validated private input
path, input digest identifies the selected PNG or reconciliation JSON bytes,
and metadata is null for a reference PNG. Canonical metadata is the closed object
`{reconciliation,capture_intent,source_identity,fixture}`: a copied validated
reconciliation v3 object, its exact capture-intent record, the matching current-schema run manifest, and the exact registry fixture or null
for an unregistered capture. These nested values retain their existing adopted
schemas. They are frozen in the private bundle; later analysis does not reread a
mutable source root. Import refs never enter structural receipts.

The observation context has exactly `epoch`, `started_at`, `finished_at`,
`duration_ms`, `viewport`, `image_dimensions`, `device_scale_factor`,
`visual_viewport_scale`, `css_zoom`, `theme`, `density`, `document_generation`,
and `coordinate_transform`. Viewport and image dimensions are `{width,height}`;
all dimensions are positive integers. Browser scale and CSS zoom are finite
positive numbers; unavailable CSS zoom is null, never inferred from image size.
Theme and density are observed owner values or null when unavailable; null makes
an unbound review incomplete for that property but fails a binding that requires
it. Document generation is a session-local nonnegative integer advanced on each
main-frame navigation, not a product revision.

`coordinate_transform` is `{origin_x,origin_y,scale_x,scale_y}` mapping CSS
viewport coordinates to image pixels by `px=(x-origin_x)*scale_x`,
`py=(y-origin_y)*scale_y`. Origins describe the selected crop; scale values are
derived from the actual viewport screenshot and viewport dimensions. They MUST
NOT be guessed from device scale factor alone. CSS zoom is already represented
in measured DOM rectangles and MUST NOT be applied again.

An imported canonical capture lacking actual pixels, DOM observations, or trace
remains a valid bundle with the corresponding components null and limitation
tokens set. An expected-only bundle is labeled `no_actual`; it can be reported
but cannot satisfy an operation requiring a primary image. Expected pixels are
never relabeled as an actual capture.
The adapter MUST NOT fabricate a browser epoch or current DOM snapshot for a
historical capture. For imported evidence, the observation context is null unless
its complete fields can be proven from the exact source; available capture
metadata remains in the validated private import reference.

### 8.3 Private observation component

`cartulary.ui_review_observations.v1` contains exactly `schema_id`, `elements`,
`accessibility_snapshot`, `axe`, `console`, and `network`.

Each element record has exactly `target`, `resolved_ref`, `role`, `name`, `text`,
`text_truncated`, `rect`, `visible_rect`, `focused`, `disabled`, `scroll`, `style`,
and `overflow_candidate`. Role/name/text are strings or null when unavailable;
text is capped at 4096 UTF-8 bytes with an explicit `text_truncated` boolean.
Rectangles are `{x,y,width,height}` finite CSS numbers;
visible rectangle can be null. Scroll is `{left,top,client_width,client_height,
scroll_width,scroll_height}` or null. Style contains exactly `font_family`,
`font_size`, `font_weight`, `line_height`, `color`, `background_color`,
`padding`, `gap`, `overflow_x`, and `overflow_y`; values are computed-style
strings. Computed style is evidence of browser
rendering, not a token-authority registry.

`overflow_candidate` is true when `scrollWidth>clientWidth+1` or
`scrollHeight>clientHeight+1` on the observed element. It is a review hint,
not a defect verdict: intended scrolling can satisfy that predicate. `focused`
compares against the observed active element. `disabled` is the browser-observed
native/ARIA state; absence is false. Bounding-box intersection does not establish
occlusion, clickability, sufficient contrast, or keyboard reachability.

The accessibility snapshot is a nullable string capped at 1048576 UTF-8 bytes.
Exceeding that bound fails required snapshot capture rather than silently dropping
nodes. The element count is bounded by the request's 64 targets; no whole-DOM
dump is implicit. Axe has the exact fields defined in Section 9.2.

Console and network are each `{records,truncated}`. Console records have
`sequence`, `level`, and `text`; level is `error`, `warning`, `info`, or `debug`.
Network records have `sequence`, `method`, `url`, `status`, `outcome`;
status is an HTTP integer or null, outcome is `response` or `failed`.
Capture at most the most recent 200 records per channel since the last main-frame
navigation, retaining sequence order; each text/URL is capped at 4096 UTF-8 bytes.
Truncation is explicit for either count or string truncation. Request/response
bodies, headers, cookies, and browser storage are excluded. Raw private channel
content is never copied into receipts or report URLs.

### 8.4 Atomicity and limits

The private bundle root has an immutable `bundle.json` plus relative component
files. All referenced files MUST exist, match byte count and digest, and validate
before an exclusive atomic publication makes the bundle visible. A failure removes
the unpublished staging tree. An existing bundle is never overwritten. Analyses
create new bundles referencing the originals; report generation does not edit them.

Component and derived refs in a bundle resolve relative to that bundle's directory.
An analysis copies required parent image/component bytes into its own tree before
publication, preserving their digests; a cross-bundle `..` reference is invalid.
The parent-manifest references preserve provenance separately. Logical copied
bytes count against all limits even if the filesystem shares physical storage.
The exact-diff derivation names both copied inputs in left/right order; contact
sheets name every input in display order.

| Resource | Closed limit and consequence |
| --- | --- |
| PNG input | 32 MiB encoded; 16777216 pixels decoded; width/height at most 8192 each. Reject before or during bounded decoding. |
| Image encoding | Single-frame PNG, 8-bit RGB/RGBA, sRGB or no color-profile metadata; reject unsupported color profiles, animation, depth, or corrupt/truncated data. |
| Bundle | At most 64 files and 128 MiB total referenced local bytes. Reject publication when exceeded; no implicit truncation of required components. |
| Session private outputs | At most 512 MiB and 100 published bundles. New work fails `resource_conflict` before exceeding either cap; no automatic eviction. |
| Private JSON component | At most 8 MiB encoded. Required component overflow is `artifact_error`. |
| Report | At most 32 MiB HTML; generated assets count against session storage. Oversize report is `artifact_error`. |

Media encoders' metadata MUST NOT inject filesystem paths or timestamps into
derived images. Pixel content and recorded geometry, rather than incidental PNG
compression bytes, determine image-analysis acceptance. Decoder and encoder
versions are pinned for repeatability.

## 9. Image analysis, accessibility, and human review

### 9.1 Image-analysis request and algorithm

**UIR-08 — Analysis.** The request has exactly
`schema_id="cartulary.ui_review_analysis_request.v1"`, `bundle_id`, `operations`,
`comparison`, and `crops`. Bundle ID is required. Operations defaults to
`["contact_sheet"]`; allowed operations are `contact_sheet`, `crop`, `overlay`,
and `exact_diff`, unique and ASCII-sorted after parsing. An empty operation set
is invalid. Comparison defaults null; crops defaults `[]`.

`comparison` is null or `{bundle_id,kind}` where kind is `matched_capture` or
`reference`. It is required exactly when `exact_diff` is selected and otherwise
invalid. Both bundles MUST belong to the current session. `crops` has 1–16
rectangles when `crop` is selected and MUST be empty otherwise; each rectangle
is `{x,y,width,height}` in original screenshot pixels, entirely within the image.
Duplicate rectangles are invalid. Overlay requires an observation component
with at least one non-null visible element rectangle.

Evaluation order is crop, overlay, exact_diff, then contact_sheet, omitting
unselected operations. The sorted request array describes selection, not execution
order. Derived records use that evaluation order, with crops in request order.

The primary image is `actual` when present, otherwise `original`. Missing both
is `artifact_error`. A request MUST NOT silently select `expected`, a thumbnail,
or a previously derived image as the primary image.

| Operation | Required algorithm | Meaning of result |
| --- | --- | --- |
| `crop` | Copy exact source pixels from the requested rectangles; preserve order and record source image digest and rectangle. | Magnified inspection uses the report viewer; crop creation never resamples. |
| `overlay` | Create a separate RGBA image the same size as the primary image. Map visible rectangles with the recorded transform, round outward, clip to image bounds, and draw a two-pixel opaque magenta inside border. Composite over a copy of the primary. | Shows observed bounds only. It does not prove overlap is a defect. |
| `contact_sheet` | Use available expected, primary, diff, then newly requested crops in that order; omit duplicate image refs. Fit each into a 320×240 cell preserving aspect ratio, never upscale, white opaque cell background, four columns, eight-pixel gutters, row-major order. Sharp's pinned Lanczos3 downsampling is the only resize. | An overview; originals remain available at native resolution. Labels appear in HTML, not rasterized into images. |
| `exact_diff` | Decode both inputs to straight 8-bit RGBA, adding alpha 255 to RGB. No alignment, scaling, color conversion, threshold, blur, or antialias suppression. A pixel differs if any channel differs. Emit transparent black for equal pixels and opaque magenta for differing pixels. | Diagnostic exact difference; never a replacement for Playwright's golden comparator. |

Matched-capture comparison requires equal dimensions, canonical renderer profile,
served source identity class, viewport, zoom, density/theme, scope, scenario,
capture ID, and mask declaration. Source content digests may differ intentionally
and are recorded as the compared revisions. Both inputs MUST carry sufficient
validated canonical metadata; absent metadata is incompatibility. Host captures,
dev captures, and bare reference images cannot satisfy this comparison kind.
Reference comparison requires only equal pixel dimensions; it MUST record
`cross_source_comparison`, and all semantic equivalence claims remain absent.
Unequal dimensions produce `artifact_error`; resizing is never an implicit fix.

Current-schema captures from different source snapshots can be imported for
diagnostic comparison; Section 6's current-evidence compatibility rules still
prevent those artifacts from closing a current product gate. Scope and mask
declarations MUST be present in the exact imported machine evidence or registered
fixture. If either is unavailable, matched_capture is rejected and an explicitly
requested reference comparison is the only pixel comparison available. The
adapter does not reconstruct those declarations from test code or Markdown.

The analysis result is exactly `{operations,comparison,findings}`. Comparison is
null without exact diff; otherwise it contains `kind`, `left_bundle_id`,
`right_bundle_id`, `width`, `height`, `different_pixels`, `total_pixels`, and
`different_fraction`. The fraction is `different_pixels/total_pixels` rounded
to six decimal places with ties upward. A zero difference is an observation,
not an approval. Findings are zero or more closed records
`{code,element_ref,artifact_ref}` where code is `overflow_candidate`,
`different_pixels`, or `accessibility_incomplete`; refs are nullable according
to the finding's available source. The wrapper emits no aesthetic score,
similarity pass threshold, automatic alignment repair, or golden-update request.

### 9.2 Axe normalization

**UIR-09 — Accessibility observations.** Page capture with `include_axe=true`
MUST run the pinned axe engine against the current main document after preparation.
The explicit tag set is `wcag2a`, `wcag2aa`, `wcag21a`, `wcag21aa`, and
`best-practice`. The wrapper accepts no caller rule exclusions or implicit
baseline suppression. Iframe contents and browser chrome are outside v1 scan
scope and MUST be counted as unassessed when present. The adapter disables
axe iframe traversal explicitly. Open shadow roots within the main document
remain in scope; closed shadow-root contents are not observable. Existing keyboard and
focus-continuity checks remain independent obligations.

The `axe` object has exactly `status`, `engine_version`, `scope`, `violations`,
`incomplete`, and `unassessed_frames`. Status is `completed`, `disabled`, or
`unavailable`; scope is `main_document`. Engine version is nonempty for completed
analysis and null otherwise. Disabled means the caller explicitly selected false,
or the operation is `snapshot`, whose contract does not request an axe scan.
Unavailable means an imported source supplied no axe data; live requested analysis
failure fails capture instead of returning unavailable.

`unassessed_frames` is the nonnegative count of main-document iframe elements.
For disabled/unavailable analysis both result arrays are empty; the frame count
is observed when a live main document exists and is null for imported image-only
data. Snapshot otherwise uses the same observation component without an image.

Violations and incomplete arrays contain `{rule_id,impact,node_refs}`. Rule ID is
the exact pinned-engine rule ID. Impact is `minor`, `moderate`, `serious`,
`critical`, or null; node refs are private normalized element references or
engine target strings when no DOM target can be resolved. Retain at most 1000
combined rule/node occurrences; overflow fails required analysis as `artifact_error`.
Do not retain engine HTML snippets or help URLs as machine requirement authority.

| Engine outcome | Wrapper outcome | Command effect |
| --- | --- | --- |
| Violations | Preserve rule ID, impact, affected nodes | Command can succeed; findings are advisory. |
| Incomplete | Preserve separate incomplete records and finding | Command can succeed; no accessibility-complete claim. |
| Passes / inapplicable | No violation record | Does not establish keyboard or screen-reader usability. |
| Unsupported or malformed result | No valid axe component | `artifact_error`; no complete bundle. |
| Engine execution error | No valid axe component | `tool_diagnostic_failure`; no complete bundle. |
| Engine timeout | No valid axe component | `timeout_failure`; no retry. |
| Scan disabled | Empty arrays, status disabled, `no_axe` limitation | Command can succeed. |

The revised spec MUST distinguish this channel from
`cartulary.frontend_accessibility_summary.v4`. V1 review does not change that
summary schema or add product rows. Promoting an axe rule into blocking product
verification requires a separately owner-grounded catalog change; that promotion
is outside this revision.

### 9.3 Report behavior

The report command MUST validate the selected manifest and every referenced digest
before rendering. It MUST generate a standalone local HTML report referencing only
files in its private report tree. The report contains source classification,
capture identity, limitations, originals, derived contact sheet/crops, exact diff
counts, geometry observations, accessibility findings, and bounded console/network
observations when present. Missing optional data is visibly labeled unavailable
or disabled; an empty finding array is not used to imply missing work passed.

Images open at native size; the viewer supports zoom and side-by-side comparison.
A before/after slider is available only for equal-size compared images. Viewer
layout, colors, and implementation library are free choices; displayed values,
labels of evidence class, source identities, and missing-state behavior are fixed.
An imported trace is linked as a local diagnostic file without parsing its
third-party internals or launching a remote viewer.

Repeated rendering of a bundle MUST preserve its manifest digest and numerical
results. Reports are cached only within the private session by bundle digest and
report implementation version; they are never evidence cache hits. No browser
server, remote resource, analytics request, CDN asset, web font, or upload occurs.
The report is inspectable by a human and its underlying JSON by an agent. Model
interpretations belong in the review conversation or a separately authored finding
record; the harness does not present them as measured observations.

## 10. Privacy, retention, and artifact consumption

**UIR-10 — Two data lifetimes.** New UI review commands MUST apply this closed
disposition. This is an intentionally conservative default for the new interface;
it does not retroactively relocate existing canonical visual artifacts.

| Material | Location and permissions | Lifetime | Permitted consumers |
| --- | --- | --- | --- |
| Credentials, browser storage, controller handles, raw child capture | Existing external suite-private runtime, directories 0700/files 0600 | Delete after last consumer or session termination | Exact owning process/adapter only. |
| UI screenshots, crops, overlays, DOM/ARIA text, axe nodes, URLs, console text, HTML | Session-private review tree with the same secure creation policy | Until session stop/expiry/failure; no automatic retained copy | Explicit local human/agent inspection through returned PrivateRefs. |
| Borrowed canonical images/traces or supplied reference files | Original location, validated read-only | Controlled by their original owner | Import exact selected bytes into private review tree; never modify source. |
| Session locator and structural receipts | Normalized retained run root, 0700/0600 | Existing result-root cleanup policy | Harness diagnostics, status, and exact session lookup. |
| Request input file | Caller-owned private location | Caller-owned | Read validated snapshot only; no retention or deletion. |

The agent-visible private file path is a transient local inspection reference,
not an export capability or publication approval. On session closure the report
and image links expire; the structural receipt remains. Durable image export is
not part of v1. A caller needing to inspect the images MUST keep the review session
alive or reimport the original canonical files into a new artifact session.

All private and retained writes MUST establish permissions before writing, reject
symlinks/ownership mismatches, and use the existing private-state publication
primitive where applicable. Every retained receipt is a schema allowlist
projection, followed by existing redaction and retained-boundary scans. It MUST
NOT contain private image hashes, page text, selectors, screenshots encoded as
strings, user identifiers, private paths, or raw engine errors. A scan failure
prevents success but does not skip cleanup.

HTML rendering MUST escape all page-authored strings and attribute values. It
MUST NOT insert observed HTML as markup, execute page-provided scripts, or turn
observed URLs into active links. Its content policy denies network connections,
forms, embedding, and external assets; the renderer supplies only its own local
viewer code and local images. Directories and files returned in PrivateRefs MUST
be validated again at consumption, not trusted because they were once valid.

These commands MUST NOT read Markdown to discover scenarios, rules, thresholds,
selectors, test ownership, or tool versions. Human editorial review links
requirements to machine projections. No parser or conformance test consumes this
plan or the target NLSpec as a requirement registry.

## 11. Failure classification and completion semantics

**UIR-11 — Failure mapping.** The following diagnostic codes are closed for v1.
Use the existing normalized class/reason and public code. The outer GNU Make exit
can differ; consumers use the result/receipt or compact classified output.

| Condition | Diagnostic code | Class / reason | Normalized exit | Required consequence |
| --- | --- | --- | ---: | --- |
| Unknown/blank input, invalid enum, unsupported action, bad JSON request, wrong request variant | `invalid_request` | `config / usage_error` | 2 | No page mutation, service acquisition, or partial output publication. |
| Missing required tool, wrong pin, invalid owned configuration | `tool_configuration` | `config / configuration_error` | 2 | No adapter fallback or network install. |
| Wrong session owner, stale epoch/ref, terminal session, invalid locator | `session_mismatch` | `config / configuration_error` | 2 | No action replay or other-session search. |
| Zero/multiple/unavailable target matches, disabled/non-actionable target | `target_unavailable` | `harness / tool_diagnostic_failure` | 1 | No forced action or first-match selection. |
| Unexpected dialog/tab or rejected navigation origin | `navigation_boundary` | `harness / tool_diagnostic_failure` | 1 | No automatic acceptance, tab switch, or origin fallback. |
| Docker/platform preflight failure in seeded mode | `environment_unavailable` | `infra / preflight_error` | 3 | Attempt partial owned cleanup. |
| Owned browser or service launch fails | `startup_failed` | `infra / service_start_error` | 3 | Stop dependent work; clean acquired resources. |
| Selected dev origin or owned service misses readiness deadline | `readiness_expired` | `infra / service_readiness_timeout` | 3 | No switch to another listener or mode. |
| Lock contention, storage cap, confirmed ENOSPC | `capacity_exceeded` | `infra / resource_conflict` | 4 | Preserve published bundles; remove only unpublished owned output. |
| Wrong/missing artifact, digest mismatch, unsupported imported schema, invalid image, incompatible comparison | `invalid_artifact` | `artifact / artifact_error` | 11 | No substituted/latest artifact; no successful bundle. |
| Document/geometry changes across capture boundary | `unstable_capture` | `artifact / artifact_error` | 11 | Discard candidate; no tolerance change, registration, or recapture. |
| Required snapshot/axe/component size limit exceeded | `observation_limit` | `artifact / artifact_error` | 11 | No silent truncation of required observations. |
| Axe execution fails | `analysis_failed` | `harness / tool_diagnostic_failure` | 1 | No complete bundle. |
| Action/capture/analysis/report watchdog expires | `operation_expired` | `timing / timeout_failure` | 13 | Cancel/reap owned child; do not retry. |
| Browser/controller dies after ready | `session_lost` | `infra / service_start_error` | 3 | End session and clean; no transparent browser restart. |
| Redaction, permissions, secure publication, or retained scan fails | `unsafe_artifact` | `artifact / artifact_error` | 11 | Reject success and attempt remaining cleanup. |
| Private input boundary violation detected | `input_boundary` | `artifact / artifact_error` | 11 | No restricted document read; report normalized location only. |
| Cleanup fails without earlier primary failure | `cleanup_failed` | `harness / cleanup_error` | 12 | Retain failed cleanup outcome and exact ownership proof. |
| Signal/cancellation | `interrupted` | `interrupted / cancelled_or_interrupted` | 130/143/15 | Cancel dependent work; close owned resources. |

Literal malformed path syntax is `invalid_request`; syntactically valid imported
artifacts with unsafe containment or contents are `invalid_artifact`. Borrowed
request-file ownership/mode violations are `tool_configuration`. Missing private
files after a terminal session are `session_mismatch`, not a request to recreate
them. A malformed producer result is `invalid_artifact`.

Primary and secondary failures follow existing Section 9.1 ordering. A finite
operation failure is recorded in that operation's receipt; it does not make the
session's later normal stop fail unless session integrity was lost. Session
lifecycle failures and cleanup failures determine terminal status. Status/stop
MUST NOT erase previously recorded operation failures.

Nonzero pixel difference, an overflow candidate, axe violations, console errors,
or failed application requests are findings, not command failures. A command
succeeds when it faithfully produces its specified observations and cleans its
temporary resources. It does not declare the UI correct. Product test assertions
retain `product/test_assertion_failure` only in their existing canonical test
commands; no new review command emits that class or passing product-row evidence.

## 12. Toolchain, compatibility, and migration

**UIR-12 — Toolchain and cutover.** The editor MUST add a closed core tool profile
covering the Playwright API adapter, Sharp, axe integration/engine, and their locked
transitive dependencies. The existing Node/pnpm/Playwright/browser pins remain
the starting compatibility baseline. The revision MUST NOT adopt a silently
newer browser merely to use a newer driver.

Exact candidate package versions are an adoption input, not a runtime choice.
Before adopting the amendment, the editor MUST complete a compatibility record
with package name, exact version, integrity/lock identity, required Node version,
Playwright/browser compatibility, supported Linux/WSL2 architecture, and the
qualification outcomes in UIA-02. Package releases are selected from official
upstream distributions. A pin that cannot satisfy the interface blocks adoption
of the adapter; it is not permission to weaken the interface or fall back to PATH.
This plan does not assert that the newest CLI release is compatible with the
repo's Playwright 1.59.1 renderer. The execution decision uses its API directly.

After adoption, `tools/toolchain_pins.json` and package-manager lock state project
the qualified exact versions through the existing Make-owned bootstrap and drift
mechanics. `bootstrap` installs the core profile; ordinary review commands do not
download packages or resolve `latest`. `doctor` reports core readiness and version
mismatches without installing anything. Optional Python/OCR/video/MCP packages
remain outside those required checks.

Node/JavaScript wrappers resolve the repository's pinned runtime and package
executables explicitly; ambient global Node, pnpm, Playwright, ImageMagick, or
Pillow MUST NOT alter results. Documentation can describe entering a shell with
the pinned runtime on PATH, but no shell initializer or user-global configuration
is modified automatically. Canonical repository work continues through Make.

| Existing contract | Compatibility decision |
| --- | --- |
| Testing harness document/profile v3 | Add a coherent helper-only amendment; do not create a second harness authority or change graph/catalog identity. Record the adoption revision through the repository's document process. |
| Existing design-review v1 commands | Preserve behavior and output; share private lifecycle implementation with new commands. |
| Canonical visual renderer profile | Preserve its exact pin and ownership. Host review never attests it. |
| Capture intent v2, reconciliation v3, fixture registry v6, browser group v6, browser target v4 | Read current forms only. Repair the stale refresh reference to reconciliation v3. No schema bump without an actual shape/semantic change. |
| Accessibility summary v4 | Preserve canonical row accounting and its existing schema. Axe review observations use a separate wrapper. |
| Existing raw Playwright reports/traces | Third-party internals remain diagnostic and version-specific. Import only declared current artifact references. No generic ZIP/report parser is made public. |
| Legacy/unversioned review-session files | Historical diagnostics only; not accepted as new session locators or bundles. No alias, automatic conversion, or newest-file fallback. |
| HTML/report presentation | Free internal implementation, provided source labels, values, limitations, local-only behavior, and output lifetime remain identical. |

The editor MUST repair TH-HARNESS-REQ-255's `reconciliation v2` refresh reference
to v3 in the same amendment and update its acceptance cross-reference. The plan
does not authorize translating old v2 artifacts. Section 19 MUST continue to
exclude general Playwright format stability while explicitly acknowledging the
new stable Cartulary wrapper schemas.

## 13. Projection and implementation handoff

The document revision and subsequent implementation are separate deliverables.
The latter MUST derive from the adopted revision and machine contracts, not from
this planning file at runtime.

| Boundary | Authored inputs or implementation seam | Required downstream work after adoption |
| --- | --- | --- |
| Commands/configuration | `tools/task_surface_owner.json`, root Make composition | Add exact command/input/output contracts and backing-script inventory; generate task surface through Make. |
| Resource/lifecycle policy | `tools/execution_topology_manifest.json`, current fixture/service broker | Add review resource admission and helper dependencies without selecting review work in default gates. |
| Schemas | `tools/schemas/`, `tools/harness_schema_attachments.json` | Add closed request/result/bundle/receipt schemas and negative fixtures; no Markdown consumption. |
| Dependencies | Root/frontend manifests and `tools/toolchain_pins.json` | Record qualified dev-only packages; update lockfiles only through the package manager under Make-owned tooling. |
| Browser lifecycle | `tools/harness/browser/design-review.mjs` and current lifecycle facade | Reuse acquisition, private credentials, seeded roles, and cleanup; add no competing service owner. |
| Browser operations/capture | `tools/harness/browser/` | Implement one adapter boundary for actions, epochs, snapshots, bounded captures, and source joins. |
| Analysis/report | `tools/harness/browser/` | Implement exact image operations and HTML/JSON projections; library placement is private factoring. |
| Fixtures and presentation | Existing machine fixtures, `packages/test-utils`, visual capture helpers | Consume exact source-owned scenarios and observed presentation; do not establish another fixture or token registry. |
| Verification routing | `contracts/verification/`, `tools/test_catalog_owner.json`, `tools/test_families/` | Assign semantic harness behavior rows to current active owners. `harness.visual` is a verification contract, not automatically a task-guide owner. |
| Documentation | Browser design-readiness and visual-golden guides | Explain the new workflow and private-link lifetime after adoption; link to Make discovery instead of copying its full command inventory. |

Generated artifact membership remains owned by
`tools/generated_artifact_policy.json`. Generated Make/topology outputs and
dependency lockfiles MUST NOT be hand-edited. Changes confined to this revision
plan MUST NOT alter those inputs or outputs.

Implementation order and exits are closed:

| Work package | Depends on | Exit |
| --- | --- | --- |
| Document integration and pin qualification | Baseline recheck | All proposed behavior is owned in target Sections 1–17; pin compatibility record complete; no conflicting requirement or open load-bearing decision. |
| Schemas and command projections | Adopted document | Requests, results, receipts, source variants, inputs, and negative fixtures agree; generation/drift succeeds. |
| Session and browser adapter | Schema/command projections | Ownership, dev borrowing, serialization, epochs, deadlines, private files, and idempotent cleanup satisfy UIA-03–UIA-10. |
| Capture and artifact adapters | Session adapter | Stable capture, exact joins, current-schema imports, bounded images, and atomic bundles satisfy UIA-11–UIA-17. |
| Analysis, axe, and report | Capture adapter | UIA-18–UIA-23 pass without changing canonical comparators or row accounting. |
| Integration and handoff | All earlier exits | Full applicable acceptance matrix passes; guides accurately describe supported workflows and limitations; retained roots are named. |

Before acceptance, an implementation can be removed by deleting only its new
authored command/tool/schema additions and regenerating downstream projections;
existing commands and goldens remain valid. After adoption, retiring a public
command or schema requires an explicit superseding document amendment. Rollback
MUST stop active owned review sessions before removing their tooling and MUST NOT
delete borrowed dev data, prior canonical evidence, or unrelated worktree changes.

## 14. Acceptance matrix for the revised behavior

The editor MUST incorporate these observable obligations into target Section 17
using new unused `TH-HARNESS-AC-*` IDs and update the existing traceability table.
The `UIA-*` labels below are editorial references only. A test count, screenshot
count, clean linter result, or empty axe result does not substitute for a row's
expected behavior. Failure codes below are normalized wrapper codes.

| ID | Contract | Fixture or boundary condition | Required observable result | Failure/cleanup obligation |
| --- | --- | --- | --- | --- |
| UIA-01 | UIR-01, UIR-02, UIR-12 | Inspect default, CI, release, owner-slice, and direct review plans | Seven helper commands have exact IDs/contracts; no review command closes or is selected as product-row work; existing design-review commands retain their interface. | Unexpected row evidence or aggregate selection blocks acceptance. |
| UIA-02 | UIR-12 | Empty ambient Node/pnpm PATH; conflicting global browser/image tools; missing/wrong core pin; absent optional tools | Qualified pinned core runs identically with conflicting ambient tools; missing core fails before acquisition; absent optional tools do not fail core readiness. | Exit 2 for core mismatch; no implicit install, browser upgrade, or fallback. |
| UIA-03 | UIR-01, UIR-04 | Omitted mode; explicit seeded profiles; invalid/blank mode; profile in dev/artifacts | Omission creates a fresh claimed seeded session; valid profile changes only its adopted claim set; invalid combinations fail before startup. | Exit 2; no service/data side effects on rejection. |
| UIA-04 | UIR-01, UIR-04 | Borrowed dev origin; other loopback service; unavailable origin; external redirect | Only explicit selected origin is used; source remains live_unattested; stop leaves borrowed service and data intact; redirect is not followed. | Readiness 3 or navigation 1; no fallback/reset/migrate/borrowed teardown. |
| UIA-05 | UIR-01, UIR-04 | Artifact mode with no Docker/browser installation | Exact imported files can be analyzed/reported; no services/browser are started. Browser operation is rejected. | Unsupported browser work exits 2; borrowed input files unchanged. |
| UIA-06 | UIR-02, UIR-07, UIR-10 | Every finite command in machine mode; start in machine mode; invalid preflight | Finite result is one schema-valid JSON object plus LF, no page data; start rejects machine mode; private refs appear only in transient results. | Exit 2 before child work for unsupported mode; stderr/receipt behavior matches contract. |
| UIA-07 | UIR-03, UIR-04 | Two concurrent commands; two concurrent sessions; competing quiet measurement | One operation per session; independent sessions remain isolated; declared resource fairness applies; contention is bounded. | Lock failure 4; no browser capacity or host-activity bypass. |
| UIA-08 | UIR-03, UIR-05 | Action followed by stale ref/epoch; zero/two matching controls; disabled control; text containing shell metacharacters | Stale/ambiguous target is not acted on; text is passed as data; exact action is executed no more than once. | Stale exit 2, non-actionable exit 1; no forced/first-match action. |
| UIA-09 | UIR-04, UIR-11 | Stop while busy; repeated stop; SIGINT/SIGTERM; owner/browser death; preparation failure | Session reaches closed/failed with exact primary cause; owned resources and private tree are removed where cleanup succeeds; status survives through retained locator. | Signal codes retained; cleanup-only 12; earlier failure stays primary. |
| UIA-10 | UIR-04, UIR-10 | Inject each partial-acquisition, file-open, close, cleanup, and terminal-publication failure | Remaining owned releases are attempted; failed ownership/cleanup remains visible; no unrelated session or borrowed file is removed. | No success after unsafe production or incomplete cleanup. |
| UIA-11 | UIR-06 | Font delay; geometry changing until timeout; geometry changes after screenshot; stable geometry with dynamic text | Stable case publishes one bundle; unstable candidate is discarded; dynamic text is labeled ordinary review without invented pixel stability. | Timeout 13 or unstable artifact 11; no action replay, tolerance adjustment, or recapture. |
| UIA-12 | UIR-05, UIR-06 | Virtualized row absent, then explicitly scrolled into view; focus before/after capture | Absent rendered target is reported; explicit reveal enables observation; capture does not move focus/scroll or infer offscreen record absence. | Unavailable target 1; no hidden data mutation. |
| UIA-13 | UIR-06, UIR-07 | Active registered/unregistered canonical capture; expected-only passing capture; wrong catalog join | Exact current joins import; fixture IDs may be empty; expected-only report shows no_actual; invalid join fails. | Invalid artifact 11; no filename-derived owner or fake actual pixels. |
| UIA-14 | UIR-06, UIR-12 | Old reconciliation/group schema, different renderer, user endpoint, modified imported bytes | Current selected adapter only; old/tampered/incompatible inputs rejected; canonical renderer endpoint is never reused by review. | Exit 11 or input 2; no translation, newest-run fallback, or reclassification. |
| UIA-15 | UIR-06, UIR-07 | Viewport vs element vs region capture; nonunit scale; CSS zoom; partial viewport intersection | Coordinate transform maps observed CSS bounds to correct native pixels once; scopes retain originals and exact offsets. | Invalid/empty region rejected; no double zoom, silent clamp of explicit region, or full-page virtualization claim. |
| UIA-16 | UIR-07, UIR-10 | Missing component, digest mismatch, duplicate path, symlink swap, interrupted publication | No complete bundle becomes visible until every ref validates; previous bundles remain byte-identical. | Artifact 11; only unpublished owned staging removed. |
| UIA-17 | UIR-07 | At/below/above each byte/pixel/count/storage bound; corrupt PNG; unsupported profile/depth/animation | Boundary values accepted exactly; over-limit required data fails; allowed console/network truncation is explicit. | Cap 4 or artifact 11 per table; no unbounded decode or silent required-data loss. |
| UIA-18 | UIR-08 | Equal images; one RGBA-channel change; all pixels changed; unequal dimensions | Exact counts/fraction/diff pixels follow the algorithm; originals and canonical comparator unchanged. | Unequal dimensions 11; nonzero difference alone exits 0. |
| UIA-19 | UIR-08 | Requested crops, out-of-bounds crop, overlay with/without geometry, canonical vs reference comparison | Exact pixels and recorded source rectangle; no hidden alignment; matched_capture requires full compatibility; reference comparison is visibly non-equivalent. | Invalid request 2 or artifact 11; no automatic resize or baseline update. |
| UIA-20 | UIR-09 | Axe violation, incomplete, no finding, disabled scan, missing imported axe data | Each outcome maps to its distinct state/array; no accessibility-complete or product-pass claim. | Findings alone exit 0; existing keyboard/contrast assertions remain independent. |
| UIA-21 | UIR-09, UIR-11 | Axe engine throws, times out, returns malformed/oversize output | No successful partial bundle substitutes for requested analysis. | Tool 1, timeout 13, malformed/limit 11; no suppressed rules or automatic retry. |
| UIA-22 | UIR-08, UIR-10 | HTML-like page text, malicious URL, absent optional artifact, report opened offline | Text escaped; no observed script/URL executes; report is self-contained and correctly labels unavailable channels. | No network/export activity; source bundle digest unchanged. |
| UIA-23 | UIR-07, UIR-10 | Local agent opens returned image and human opens report; session then stops | Both see the same exact bundle; native-resolution originals available while live; links expire after cleanup, structural receipt remains. | No retained raw-detail copy or promise of post-stop image availability. |
| UIA-24 | UIR-10 | Credentials in typed text, page content, console, URL, and binary screenshot; inherited telemetry settings | Detail remains private; retained records contain only closed structural fields; no automatic export or telemetry capture of detail. | Unsafe retention fails 11 and still attempts private cleanup. |
| UIA-25 | UIR-10 | Documentation-path input, disguised Markdown input, symlink escape; neutral policy fixtures | Restricted input is rejected before read/stat/hash; executable validation itself does not consume documents. | Boundary 11; no document content included in diagnostics. |
| UIA-26 | UIR-11 | Each mapped failure alone and paired with cleanup failure; operation failure followed by normal session stop | Correct class/reason/code; secondary cleanup visible; operation and session outcomes remain distinct. | Generic Make failure does not overwrite the normalized cause. |
| UIA-27 | UIR-12 | Visual-update editorial contract and current machine projections | Refresh requires reconciliation v3 everywhere; old v2 cannot qualify; existing ordinary validation and human golden review remain required. | Incompatible reference or acceptance drift blocks adoption/implementation completion. |
| UIA-28 | UIR-01–UIR-12 | Full seeded editor/viewer review and artifact reimport; separate dev capture | Navigate, act, capture, analyze, inspect report, and stop through public Make; source and lifecycle claims agree in all three modes. | No catalog accounting, committed goldens, or unrelated worktree files changed; dev data changes only through explicit requested UI actions, never lifecycle reset/cleanup. |

The numerical image cases MUST include transparent pixels, alpha-only differences,
one-pixel boundaries, and zero/full changed area. Interface fixtures MUST include
omitted vs null vs empty values, unknown fields, duplicate JSON keys, non-finite
numbers, wrong schema IDs, duplicate refs, and each allowed enum boundary. These
are independent obligations even if several are covered by one implementation test.

## 15. Editorial acceptance and verification sequence

### 15.1 Document-revision Definition of Done

The document revision is complete only when all of the following are true:

1. Each G1–G8 disposition has a single normative owner in the revised NLSpec,
   and the Section 3.2 edit matrix is fully applied.
2. Every public input has a type, source precedence, omission rule, empty/null
   rule, bound, invalid-input consequence, and forwarding/retention rule.
3. Every observable operation has a finite state transition, source/ownership
   contract, output schema, deadline, failure mapping, and cleanup consequence.
4. Private observation content, retained structural evidence, optional data,
   incomplete analysis, live dev uncertainty, and canonical evidence are distinct
   in the data model and prose.
5. Each requirement has at least one acceptance obligation; each UIA row maps
   back to an owning requirement. Human traceability stays in the document.
6. Package compatibility is qualified and exact pins are recorded before
   adoption. There is no unresolved tool default, implicit optional dependency,
   pending source-format choice, or load-bearing implementation guess.
7. Current visual/accessibility accounting and the v3-only refresh correction
   agree across the main body, acceptance criteria, and evidence-limit sections.
8. A human walkthrough of all three modes reaches a determined result for
   success, invalid input, timeout, partial failure, interruption, and repeated stop.
9. A second editorial pass applies both the interchangeability and recreatability
   tests from the NLSpec standard. Gaps found by that pass are repaired in their
   owning clauses, not delegated to tools, appendices, or implementation defaults.
10. Documentation lint passes or a pre-existing unrelated failure is identified
    precisely. A linter pass is not represented as semantic spec validation.

Requirement prose MUST use present-tense behavior and exact normative verbs.
Historical observations and third-party recommendations belong in rationale or
revision evidence. The editor MUST NOT copy this plan's narrative baseline into
adopted requirements, introduce executable requirement-ID catalogs, or make a
test parse Markdown to prove coverage.

### 15.2 Verification by deliverable

| Deliverable | Required verification | Explicitly skipped |
| --- | --- | --- |
| This planning document | Human cross-reference/interface/acceptance review; `make lint-markdown` | Product suites, generators, schema drift, release checks, and retained-run maintenance; no executable change and no `RESULTS_DIR`. |
| Actual NLSpec amendment | Editorial Definition of Done above; `make lint-markdown`; completed pin qualification record | No product-readiness claim from documentation-only changes. |
| Implemented projections and tools | Discover active routing with `make help`, `make help-all`, `make task-guide ROLE=module-author OWNER=<active-owner>`, and `make explain-target TARGET=<target> DETAIL=summary`; then narrow schema, harness-contract, and owner checks | No blind broad-suite rerun merely because a new helper exists. |
| End-of-implementation verification | `make agent-finalize` before broader verification; applicable `make json-shape-check`, `make generated-artifact-policy-check`, `make generate`, `make generate-drift`, `make toolchain-drift`, harness checks and selected browser slices | Retained maintenance only when an exact eligible successful full warm root is supplied; otherwise report it skipped. |

The implementation author MUST discover current active owner IDs instead of using
verification-contract IDs as task-guide owners. New smoke/contract scenarios are
routed by semantic behavior through the authored catalogs. Documentation amendment
alone MUST NOT produce or alter schemas, generated artifacts, test rows, or release
evidence. The existing full v3 cutover procedure is not automatically repeated for
an additive helper amendment; broaden verification only for changed ownership or
unresolved risks, and record the reason.

Handoff MUST name the adopted revision, substantive source changes, exact commands,
retained roots where applicable, normalized failures, privacy/lifetime limitations,
and skipped checks with reasons. Completion of this plan means the planning
artifact is written and reviewed; completion of the amendment and implementation
requires their separate exits above.

## 16. Iteration-1 execution tracker — historical

This is the sole execution tracker for the user-authorized remediation. The
implementation baseline is `ce3ef227c`, initially clean, inspected on 2026-09-27
(UTC). Section 2's older baseline is historical evidence. The user selected the
pinned Playwright API in place of the proposed standalone CLI. Implementation
proceeds strictly WS1 → WS2 → WS3 → WS4 → WS5 → WS6, with a recorded completion
checkpoint before the next workstream starts. Documents are human authority and
never executable inputs.

| Workstream | Status | Dependency | Required exit |
| --- | --- | --- | --- |
| WS1 — Specification integration and dependency qualification | DONE | Baseline inspection | Adopted amendment, qualified exact dependencies, editorial review and Markdown lint. |
| WS2 — Schemas, commands, dependencies, and output projections | DONE | WS1 | Closed schemas, helper-only bindings, private output separation, generated/drift checks. |
| WS3 — Session controller and browser adapter | DONE | WS2 | Ownership, mode separation, strict actions, deadlines, cancellation and cleanup verified. |
| WS4 — Capture, artifact imports, and immutable bundles | DONE | WS3 | Exact source joins, stable bounded capture, secure immutable publication verified. |
| WS5 — Image analysis, axe, and local reports | DONE | WS4 | Image algorithms, advisory axe, safe offline reports verified. |
| WS6 — Validation and handoff completion | DONE | WS5 | Applicable UIA-01–UIA-28 complete, guides updated, cleanup and final evidence recorded. |

### Session log

| Date (UTC) | Workstream | Changes and evidence | Validation / next action |
| --- | --- | --- | --- |
| 2026-09-27 | WS1 | Rechecked clean baseline, existing lifecycle, pinned Make runtime and installed Chromium 1217; corrected the research link and driver decision. | Qualify dependencies in disposable storage before adopting the amendment. No implementation pass claimed. |

### Risks recorded at iteration-1 startup — historical

These were startup concerns, subsequently dispositioned by the WS1–WS6
checkpoints. They are not the current iteration's unresolved-work list; use
Section 18.11 for that list.

- Qualification candidates are Playwright 1.59.1, Sharp 0.35.4, axe integration
  4.13.0 and axe engine 4.13.0; package metadata alone is not qualification.
- Existing step output tees stdout into retained logs. New transient results must
  bypass that retention path through an explicit structural projection.
- Current canonical artifact joins must reject ambiguity without broadening
  third-party format compatibility or inventing missing observations.
- No eligible successful full warm `RESULTS_DIR` has been supplied; retained-run
  maintenance is not yet applicable.

## 17. Rationale and external references — non-normative

TypeScript/JavaScript is the initial integration language because the repo already
owns browser fixtures, schema validation, lifecycle orchestration, and frontend
tooling there. Sharp supplies deterministic image operations without adding a
Python runtime to canonical readiness. Optional Python analysis remains useful
when a measured investigation needs structural similarity or computer vision.

Private detail bundles avoid treating an OCR pass, generic redactor, or successful
image decode as proof that screenshots contain no secrets. Structural receipts
preserve lifecycle/debugging evidence after the short-lived review files are gone.
This costs post-session image availability; the lifecycle and UI explicitly expose
that tradeoff. Existing canonical screenshot retention remains separately owned.

Official documentation informs adapter selection only; it does not supply default
behavior missing from the main contract:

- [Playwright API](https://playwright.dev/docs/api/class-playwright): browser actions,
  snapshots, sessions, screenshots, and human observation of a session.
- [Sharp extraction](https://sharp.pixelplumbing.com/api-resize/) and
  [compositing](https://sharp.pixelplumbing.com/api-composite/): deterministic
  crops, derived overview images, and overlays.
- [Playwright accessibility integration](https://playwright.dev/docs/accessibility-testing):
  axe integration and limits of automated accessibility assessment.
- [Playwright MCP](https://github.com/microsoft/playwright-mcp) and
  [Chrome DevTools MCP](https://github.com/ChromeDevTools/chrome-devtools-mcp):
  possible later adapters, subject to the same Cartulary contracts.
- [uv](https://docs.astral.sh/uv/),
  [scikit-image metrics](https://scikit-image.org/docs/stable/api/skimage.metrics.html),
  [OpenCV](https://docs.opencv.org/4.x/d4/d73/tutorial_py_contours_begin.html),
  [Tesseract](https://tesseract-ocr.github.io/tessdoc/Command-Line-Usage.html), and
  [FFmpeg](https://ffmpeg.org/ffmpeg.html): deferred analysis capabilities.

### WS1 qualification evidence

Qualification ran on Linux amd64, WSL2 kernel 6.6.114.1, Node 24.15.0 and
pnpm 10.33.0. The pinned Chromium 1217 reported 147.0.7727.15. No separate
native-Linux kernel host was exercised; Linux userland behavior was exercised
on WSL2. Sharp used libvips 8.18.6 and libpng 1.6.58.

| Package | Exact version | Published integrity |
| --- | --- | --- |
| playwright | 1.59.1 | SHA-512: `C8oWjPR3F81yljW9o5OxcWzfh6avkVwDD2VYdwIGqTkl+OGFISgypqzfu7dOe4QNLL2aqcWBmI3PMtLIK233lw==` |
| playwright-core | 1.59.1 | SHA-512: `HBV/RJg81z5BiiZ9yPzIiClYV/QMsDCKUyogwH9p3MCP6IYjUFu/MActgYAvK0oWyV9NlwM3GLBjADyWgydVyg==` |
| sharp | 0.35.4 | SHA-512: `n++8XWcj+jCOr2IOl7h8LbKnGBDY4aPbmprMONBNFdn0ImXqpGVv5zliDs0V9HbmbCQLpbuo2ej9rAoOQTvMDA==` |
| @axe-core/playwright | 4.13.0 | SHA-512: `6YLx+kxXu5GJceG4ozFg+33a2EMTdjYwWGloJ3sb9Kta5pp+ZNS53uxGVog5JetIY8s++P5UrtX+cri+u0VAVg==` |
| axe-core | 4.13.0 | SHA-512: `UzGt8zg7Ny8djbYMhxl2zuEevVa7r2gJjYY5Lwr1xM7+XU2nd6CkIWFTVcCIbAP63vSz71NaVyyuSk9lHKcy0A==` |

The disposable qualification lock SHA-256 is
`b1252aba996348202700dc4f766217c0b7f63a80f12c14be4c95ee89103e69cf`.
The Make-owned installation succeeded with a disposable Make composition invoking
`frontend-install`. `make -f /tmp/cartulary-ui-qualification/qualification.mk
qualify-ui-review` passed browser launch, strict role/test-ID actions, literal
shell-like input, ARIA snapshots, screenshots, PNG decoding, transparent RGBA
preservation, exact cropping, Lanczos3, and main-document axe with an empty
ambient PATH. This is dependency qualification, not product-row evidence.

Two setup attempts failed before qualification: public `frontend-install` rejected
private stamp overrides as unknown command-line inputs (configuration, exit 2);
the initial scratch Make composition did not propagate its stamp rule to recursive
Make. A probe then failed because Sharp does not export its package.json subpath;
reading installed package metadata by its explicit path repaired the probe. None
changed the canonical browser pin or installed a global tool.

### WS1 completion checkpoint

Adopted amendment: `ui-review-api-1` in `cartulary.testing_harness.v3`.
Changed only this tracker and the Testing Harness NLSpec. Contracts were placed
in their owning Sections 1–16, with 28 acceptance obligations (AC-102–AC-129) and
human traceability in Section 17. The reconciliation refresh reference is v3.
Source variants, observation scope, snapshot references, stop identity and unsafe
terminal publication have explicit behavior. The API decision avoids a second
browser driver. A second editorial review repaired copied section references and
removed the flat nullable source encoding. No executable registry consumes IDs.

`make lint-markdown` passed: retained root
`.cartulary/test-results/20260927T030808Z-p92304`, summary
`adhoc/lint-markdown/tool-run-summary.json`. `git diff --check` passed. Dependency
qualification passed as recorded above. Product/generation/release checks were
skipped for this documentation-only slice. WS1 is complete; WS2 may now begin.

### WS2 completion checkpoint

Added the eight v1 schema attachments, closed source/request variants, semantic
validators and positive/negative contract fixtures. Registered all seven commands
as helper-only with a dedicated `ui_review` Make binding and `ui_review_receipts`
policy. Transient results bypass generic preflight capture, step stdout tee,
command metadata and observability. Raw Make inputs use literal shell transport;
unknown inputs and malformed requests fail before controller acquisition.

Qualified dependencies are now exact root dev dependencies and toolchain pins;
the root lock was updated by Make-owned pnpm installation, never by hand.
Doctor checks installed package and Node pins without installing. A single
Playwright/core version is required. Schema references use the actual run-manifest
v1 identity. Existing artifact-error taxonomy carries restricted-input failures;
no new generic failure reason was invented. Editorial follow-up removed a stale
nullable-source phrase and an unrepresentable lifetime diagnostic token: lifetime
expiry closes normally when cleanup succeeds.

Changed areas: root package/lock, toolchain pins/checker/doctor, schema attachments,
authored task surface and its validation/renderer, eight schemas, the ui-review
contract/input/output/bootstrap modules, and command-surface test registration.
Generated task-surface Make/runtime/JSON and topology render index were regenerated.

Validation (all roots below are under `.cartulary/test-results/`):

| Command | Result and exact retained root |
| --- | --- |
| Make-owned `frontend-install` via scratch Make composition | PASS `20260927T031309Z-p95503` |
| `make generate` after final backing-script changes | PASS `20260927T032919Z-p41376` |
| `make json-shape-check` | PASS `20260927T032543Z-p21664` |
| `make toolchain-drift` | PASS `20260927T032403Z-p10163` |
| `make test-slice OWNER=harness.command_surface` | PASS `20260927T032656Z-p25626` |
| `make generated-artifact-policy-check` | PASS `20260927T032437Z-p15557` |
| `make generate-drift` before subsequent test edits | PASS `20260927T032451Z-p16239` |
| `make test-slice OWNER=harness.generated_artifacts` (includes current drift) | PASS 5/5 `20260927T032937Z-p44375` |
| `make lint-scripts` | PASS `20260927T032452Z-p16501` |
| `make lint-shell` | PASS `20260927T032453Z-p16901` |
| `git diff --check` | PASS |

Related failures were repaired, not waived. Generation initially rejected missing
observability exclusions (`20260927T031837Z-p975`); schema checks rejected missing
top-level identity/closure (`20260927T032403Z-p10137`,
`20260927T032436Z-p15347`). The generated-artifact owner and later direct drift
checks failed on the topology render index after backing-script test edits
(`20260927T032702Z-p26561`, `20260927T032823Z-p33297`,
`20260927T032846Z-p37355`); regeneration repaired the stale hash. The runner
normalized that drift failure as product/test_assertion_failure, not a product
behavior regression. An output-mode command-line attempt was rejected because
that existing selector is environment-owned. No browser/product success is claimed.
Controller execution, resource acquisition and end-to-end privacy remain WS3–WS6
obligations. WS2 is complete; WS3 may begin.

### WS3 start

WS2 checkpoint recorded before starting controller work. Inspection found that
current work-graph resource admission is process-local. Cross-process review and
quiet-measurement exclusion therefore requires a shared host-admission owner used
by both graph execution and review, rather than a review-only lock. This structural
fix is within G1/G3 ownership scope; existing graph-local fairness remains intact.

### WS3 progress — not a completion checkpoint

Shared preparation now lives in `tools/harness/browser/review-preparation.mjs`;
the legacy design-review wrapper delegates to it. Added secure anchored file
operations, host admission, local authenticated IPC, session/terminal receipts,
strict Playwright actions and controller-death recovery. Managed-suite cleanup
now retains its ownership lease after failed termination. A new shell verification
kind and semantic browser-support row route the lifecycle contract tests.

The closed result and receipt now include `exit_code`, adopted in Section 8.5
before projection, to preserve 130/143 through terminal reads. Host admission now
also coordinates independent graph processes; this repairs the process-local
assumption found at WS3 entry. Literal Make transport was corrected to avoid
passing unevaluated authored default expressions as run identity inputs.

A public artifacts session at
`.cartulary/test-results/20260927T034749Z-p57821` started, returned live status,
stopped cleanly, and repeated stop returned the same terminal digest. Initial
four-case lifecycle suite passed via the public owner slice at
`.cartulary/test-results/20260927T035242Z-p68558`. These are partial results only;
controller-death and busy-stop tests are now being exercised. A flock argument
error created an empty task-owned `3` file; the invocation was corrected to use
the inherited descriptor path and that scratch file was removed.

Generation first rejected the new shell row because the browser verification
projection omitted shell evidence (`20260927T035034Z-p59370`); the owner projection
was extended and generation passed at `20260927T035104Z-p62733`. A subsequent
JSON-shape run at `20260927T035204Z-p66588` correctly rejected backing-script hashes
made stale by continuing edits. Regeneration and full WS3 exits remain pending.

### WS3 additional validation and open issues

The six-case lifecycle owner slice passed at
`.cartulary/test-results/20260927T035507Z-p72674`, including exact terminal digest,
controller-death recovery, private-tree removal, cancellation of a busy dev action
and borrowed-origin survival. `make browser-design-review-smoke` passed, including
its sample/role checks, at
`.cartulary/test-results/design-review-1790481365486-bd73433c`; the build emitted an
existing chunk-size advisory and the Playwright installer reported the configured
Ubuntu compatibility build on this host. The UI helper itself never calls that
installer. `make run-harness-smoke-lifecycle` also passed (internal target emitted
no run-root pointer).

Seeded controller starts currently fail with normalized startup_failed and
complete cleanup; required WS3 exit remains open. Exact failed roots include
`20260927T035921Z-p4308`, `20260927T040132Z-p12751`, and
`20260927T040350Z-p17662` under `.cartulary/test-results/`. Shared preparation
succeeds in a disposable Make-owned probe, so the process environment is being
isolated. A fresh run-owned frontend seal is now built even in inspect-only review;
the generic readiness cache refuses installation on a cold prerequisite when that
internal policy is selected. This guard has positive/miss test cases pending their
broader routed run.

The extended smoke run at `.cartulary/test-results/20260927T040210Z-p15160` failed
with harness/unknown_failure. It exposed an outdated retained-run-identity assertion
for the new locator-bound commands, plus run-step/Vitest diagnostics requiring
investigation. The new test assertion is to be repaired around the explicit
ui_review_receipts policy, not by reintroducing generic output capture. Attempts
to invoke individual smoke-check names and a guessed scheduler target were rejected
by Make; these are registry check names, not standalone public commands.


### WS3 completion checkpoint

The seeded startup discrepancy was missing `GO_CACHE_DIR`, `GO_MOD_CACHE_DIR`
and `GO_TMP_DIR` after public-input stripping. The generated review recipe now
forwards the existing machine-state projection. A private disposable diagnostic
identified this cause; all production diagnostic instrumentation was removed.
No private diagnostic was copied into a retained review receipt.

Public seeded review at `.cartulary/test-results/20260927T041657Z-p43491`
started successfully, authenticated editor then viewer through the normal API,
and stopped with cleanup complete. Its terminal receipt SHA-256 is
`3bfe9043aa86db02423fa60c3ebeca84d8ecf636d993155386c94b04480b038e`.
A subsequent public stop during preparation at
`.cartulary/test-results/20260927T042311Z-p90188` also completed cleanup;
terminal SHA-256 `d9a4b21f2b1134bfb716c18f890b311e2aee8c0dc05438905a7159087421c88c`.
Owned browser startup now yields to cancellation. Failed release retains its
ownership record, remaining release owners are attempted, and a failed terminal
publication cannot report success or overwrite existing bytes.

The new redirect fixture found that Playwright routing alone does not intercept
every server redirect hop. The pinned adapter now also guards each document
request through a private Playwright CDP session. No driver details enter the
public interface. Ordinary action timeouts require a new snapshot; the adopted
30-second watchdog closes an undrainable browser operation before another can
start. The controller validates requests again at the IPC boundary. Session
output reservations enforce the total byte budget before snapshot publication.

The shared host arbiter validates closed state, reclaims dead/zombie process
identities, observes browser capacity and quiet-waiter priority, and participates
in scheduled work. Direct browser contract fixtures also acquire that owner.
The smoke runner now forwards the pinned Node directory to fixture descendants;
its previous ambient-node failures are repaired.

Validation (roots under `.cartulary/test-results/` unless stated otherwise):

| Command | Result / evidence |
| --- | --- |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.private_ui_review_session_ownership_and_strict_b_b785350fa4` | PASS nine cases, `20260927T042506Z-p3098` |
| `make harness-contract` | PASS 2/2 units, `20260927T042312Z-p90373` |
| `make run-harness-smoke-extended` | PASS; internal entry emitted no final run-root pointer |
| `make browser-design-review-smoke` after shared asynchronous acquisition change | PASS `design-review-1790483148475-7b906cfd`; owned runtime removed; existing chunk/platform advisories only |
| `make generate` | PASS `20260927T042534Z-p5370` |
| `make json-shape-check` | PASS `20260927T042549Z-p8998` |
| `make lint-scripts` | PASS `20260927T042323Z-p93799` |
| `make lint-shell` | PASS `20260927T042323Z-p93828` |
| `git diff --check` | PASS |

Related failed evidence remains explicit: the redirect test failed at
`20260927T042054Z-p75896` before the guard repair; harness-contract at
`20260927T042055Z-p76379` rejected a render-index hash made stale by an edit and
passed after regeneration. A concurrent lifecycle run at
`20260927T042313Z-p90725` was interrupted after capacity contention exposed a
fixture server left open on failed admission. The fixture now closes on that
path and uses detected capacity; its routed rerun passed. Seeded failures before
the machine-state fix were infrastructure/service_start_error, unrelated to
product assertions. Full capture/import/analysis/report and complete acceptance
closure remain assigned to WS4–WS6. WS3 has no remaining slice blocker and is
complete; WS4 may begin.


### WS4 start

WS3 completion was recorded before capture/import implementation began. The
publication boundary will be the exclusive immutable manifest after every local
component has been validated; unpublished trees are private and inaccessible
through bundle lookup. Canonical imports use exact current target/group/result
references and preserve unavailable channels explicitly.


### WS4 completion checkpoint

Added `png.mjs`, `bundles.mjs`, `source.mjs`, `capture.mjs`, and
`canonical-import.mjs` under the private UI-review owner, plus bundle/import
fixtures and routed tests. The manifest is the exclusive publication commit
point. Every component is validated before publication; existing bundles and
borrowed files are never overwritten. Reads reject links, multiple hard links,
foreign ownership, group/world writes, restricted roots and unsupported formats.
Reservations precede publication, and failed publication releases the reservation.

PNG validation closes chunk integrity, dimensions, decoded pixel count, encoding,
color-profile and animation rules before bounded Sharp decoding. Captures wait for
fonts and three stable animation frames, retain full-precision rectangles for
pixel transforms, compare stability at 1/64 CSS pixel precision, preserve focus
and scrolling, and reject post-screenshot changes. Dynamic pixels do not acquire
a stability claim. The specification and observation v1 projection were amended
together to add the previously unrepresented bounded font-face record. No existing
released schema was changed.

The canonical adapter uses current manifest/target v4/group v6/reconciliation v3,
exact capture-intent v2 and pinned report joins. It checks capture-ID derivation,
renderer, row/scenario/profile, source/receipt digests and registry metadata.
Expected attachments outside the run root are admitted only at the exact declared
golden path with the declared digest. Actual/diff/trace attachments remain within
the selected run and exact test result. Expected-only and unregistered captures
remain valid without fabricated actual pixels, fixture IDs, DOM or epochs.

Validation and evidence:

| Command / workflow | Result / retained root |
| --- | --- |
| `make harness-ui-review-contract` | PASS 16 cases, including transparent RGB channels, PNG limits/corruption, crop transforms, font delay, dynamic text, immutable imports, publication/storage bounds, failed-run imports and exact negative joins; latest routed execution below |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.private_ui_review_session_ownership_and_strict_b_b785350fa4` | PASS `20260927T045014Z-p7489` |
| `make service-backed-test-slice OWNER=module.auth ROWS=module.auth.visual.capture_the_anonymous_auth_gateway_across_initia_755030aa99` | PASS 11/11 units; canonical visual row passed; `20260927T043943Z-p48220` |
| Public artifacts start / exact canonical capture / stop | PASS `20260927T044146Z-p81903`; capture `visual.capture.00320cb9a9b33f98be2b`, expected-only bundle; terminal `15989317918af6cef44af0c40290e1cdbe374cbf4f0edc57e7fc8d61cb071797` |
| Public seeded start / viewport capture with axe explicitly disabled / stop | PASS `20260927T044620Z-p88765`; sealed receipt and image joined; terminal `7e539c15ca1201352a9cbdae5181f05d172aa002640f66d5e66a05171200d392` |
| `make test-slice OWNER=harness.command_surface` | PASS `20260927T044835Z-p98087` |
| `make generate` | PASS `20260927T044941Z-p4200` |
| `make json-shape-check` | PASS `20260927T045042Z-p9638` |
| `git diff --check` | PASS |

All listed run roots are under `.cartulary/test-results/`. The first synthetic
import fixture incorrectly required a registry entry for the unregistered auth
capture (`20260927T044453Z-p83869`); it was corrected to represent the supported
null-fixture variant and rerun successfully. This was a test-fixture failure,
not permission to invent registry membership. The exact public review sessions
closed with cleanup complete and private roots removed. No golden was refreshed.
Axe execution, analysis derivations and reports remain WS5 obligations; the full
end-to-end acceptance audit remains WS6. WS4 is complete before WS5 begins.


### WS5 start

WS4 completion was recorded before beginning image derivations, axe normalization
and offline reports. Numerical algorithms and missing-data semantics remain those
adopted in Sections 8.6–8.8; findings do not become product assertions.


### WS5 completion checkpoint

Implemented `analysis.mjs`, `image-algorithms.mjs`, cancellation-owned
`image-worker.mjs`, `axe.mjs`, and `report.mjs`. New analysis bundles copy required
inputs, preserve parent manifest digests, and record ordered derivation inputs.
Exact RGBA differences include hidden transparent channels; rounding uses integer
arithmetic. Crops never resample; overlays use the recorded transform. Contact
sheets use source expected/primary/diff channels and requested crops, with four
columns, no outer gutter, floor-centered thumbnails and white compositing. These
layout details were adopted in the NLSpec before implementation.

Axe uses the qualified engine and tags with iframe traversal explicitly excluded;
normalized records omit vendor HTML/help URLs. Snapshot is explicitly disabled,
imported missing data remains unavailable, and malformed/failed scans fail the
operation. Reports copy local assets, validate all bundle digests, escape observed
content, block network resources through CSP, support native-resolution zoom and
eligible comparisons, and cache only within the session. Cache tampering fails.

Validation (roots under `.cartulary/test-results/`):

| Command | Result / evidence |
| --- | --- |
| `make harness-ui-review-contract` | PASS 22 cases, `20260927T050551Z-p18664` |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.private_ui_review_session_ownership_and_strict_b_b785350fa4` | PASS `20260927T050618Z-p21399` |
| `make generate` | PASS `20260927T050536Z-p15439` |
| `make json-shape-check` | PASS `20260927T050618Z-p21384` |
| `make lint-scripts` | PASS `20260927T050552Z-p19037` |
| `git diff --check` | PASS |

The six new analysis tests cover exact math, crop/overlay/contact-sheet pixels,
matched/reference compatibility, immutable parents, cancellation, expected-only
reports, report limits and cache tampering, all axe result classes, iframe
exclusion, escaped injection, offline image loading and working zoom. A real
pinned browser opened the generated local report with zero HTTP(S) requests.
No WS5 acceptance blocker remains. Full public-mode workflows, remaining
adversarial acceptance coverage, documentation and final checks belong to WS6.


### WS6 start

WS5 completion was recorded before starting final acceptance validation. Audit
the complete UIA matrix against implementation and current tests, close uncovered
cases, exercise the public commands in all modes, update the existing workflow
guide, then run finalize and final drift/owner checks. Qualification remains
limited to the observed Linux x64 WSL2 environment; no untested host is claimed.


### WS6 acceptance audit and validation record

The final audit added public dev/artifacts workflows with private sentinels,
current-engine/native readiness checks, complete normalized-failure/cleanup
pairing, rendered-row reveal, real 30-second readiness/geometry deadlines,
exact encoded-byte boundaries, serialization and lifetime tests. The existing
workflow guide now describes the seven commands, closed modes, exact selection,
current schemas, advisory findings and private-link expiry. Editorial subsection
numbers were reconciled without adding executable Markdown dependencies.

Broader concurrency exposed a real admission defect: a flock helper waited on its
scheduler parent's event loop while synchronous fixture preparation blocked that
loop. This produced `infra/resource_conflict` even after assertions passed. The
complete bounded transaction now executes and publishes within the helper; its
closed structural input is supplied at spawn, so even pipe EOF delivery cannot
extend lock ownership. A regression freezes the parent for six seconds while an
independent process acquires/releases admission in less than four seconds. The
full browser owner and harness contract subsequently passed concurrently with a
canonical accessibility run. Failed offline releases remain owned for teardown
retry, and primary failures retain their ordering.

Failed evidence is retained and superseded, not hidden:

- `20260927T051442Z-p53481` (`test-slice OWNER=harness.command_surface`) and
  `20260927T051453Z-p64032` (`harness-contract`) failed with
  `infra/resource_conflict`; related to the admission defect above.
- `20260927T051442Z-p53538` (`test-slice OWNER=harness.browser`) was deliberately
  interrupted after the same defect delayed completion; 37/40 units had passed.
  Its final classification is interrupted, exit 130.
- `20260927T051932Z-p17078` (`harness-ui-review-contract`) passed 25/26 cases;
  the new parent-stall regression exposed the remaining stdin-EOF dependency.
  Supplying the closed transaction at spawn repaired it; the isolated Make-owned
  regression and subsequent full owner run passed.
- An `explain-target` invocation mistakenly supplied undeclared `OWNER` and was
  rejected as configuration/usage_error before execution. Guidance was read via
  `explain-test-owner`; this was an invocation error, not a product defect.

All roots below are under `.cartulary/test-results/` unless absolute.

| Final or applicable check | Result / root |
| --- | --- |
| `make test-slice OWNER=harness.command_surface` | PASS `20260927T051933Z-p17520`; request/pin/output and existing command contracts |
| `make test-slice OWNER=harness.generated_artifacts` | PASS 5/5 `20260927T051442Z-p53506` |
| `make test-slice OWNER=harness.browser` | PASS 40/40 `20260927T052326Z-p43748`; includes new review contract tests and existing browser support/integration owners |
| `make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.postgres_cleanup_target_scoped_coordination` | PASS 3/3 `20260927T052525Z-p29964` |
| `make service-backed-test-slice OWNER=module.auth ROWS=module.auth.accessibility.verify_session_mfa_incident_forbidden_loading_an_e8f5d5f3a9` | PASS 11/11 `20260927T052326Z-p43808` |
| `make harness-contract` | PASS 2/2 `20260927T052326Z-p43927` |
| `make generate-drift` | PASS 4/4 `20260927T052457Z-p23328` |
| `make generated-artifact-policy-check` | PASS 3/3 `20260927T052457Z-p23439` |
| `make toolchain-drift` | PASS 2/2 `20260927T052457Z-p23405` |
| `make json-shape-check` | PASS 3/3 `20260927T052526Z-p30169` |
| `make lint-scripts` / `make lint-shell` | PASS `20260927T052457Z-p23919` / `20260927T052457Z-p23987` |
| `make lint-markdown` | PASS `20260927T051117Z-p30482`; final handoff lint follows below |
| `make agent-finalize` before broader verification | PASS `20260927T052301Z-p39423`; no generated changes; retained-run maintenance skipped because `RESULTS_DIR` was unset |

The public seeded editor/viewer workflow used the seven commands through a
Make-owned disposable composition. It authenticated both actors, navigated,
snapshotted, captured with axe, created crops/contact sheets, rendered reports,
reimported the same PNG twice in browser-free artifacts mode, compared exactly,
and stopped. Retained roots and terminal digests:

- `ui-review-final-1790485910454`:
  `cafb61d8e63a23529e777170f9c9eb1982aeab8650d1cf58bb594ff6cf102511`.
- `ui-review-final-1790485910454-artifacts`:
  `ba3de129d3b30372bc8f3699cdaad5217a73a496bb14f9b7e20cc6338a8f80e7`.

Both terminal receipts record complete cleanup. The public workflow fixture
checks every returned private reference disappears, borrowed input bytes remain
unchanged, and retained trees contain only locators/receipts with no page,
console, URL, typed-text, private-path or telemetry sentinel. Dev and artifacts
versions run in the routed review suite. Local report inspection uses the pinned
browser: native image dimensions, zoom, escaped markup, missing channels and zero
HTTP(S) requests are checked. It is interface validation, not a human aesthetic
approval or product publication claim.

| Obligation | Evidence and disposition |
| --- | --- |
| UIA-01 | PASS — authored helper-only commands, task-surface/graph tests, owner parity and existing design-review smoke; no helper product-row closure. |
| UIA-02 | PASS — disposable qualification with empty ambient PATH; missing pins rejected; installed native Sharp and nested engine pins inspected before acquisition; browser-free workflow uses absent browser path and unusable Docker endpoint. |
| UIA-03 | PASS — omitted seeded mode in public editor/viewer workflow, both retained design-review profiles, invalid/blank/cross-mode requests rejected by contract fixtures. |
| UIA-04 | PASS — exact dev origin, live_unattested bundles, external redirect guard, real readiness expiry, borrowed server survives stop. |
| UIA-05 | PASS — browser-free import, reimport, exact analysis and report with absent browser installation; no borrowed mutation. |
| UIA-06 | PASS — all finite public command JSON projections validated, machine start/preflight rejection, isolated transient output and structural retained receipts. |
| UIA-07 | PASS — serialized controller admission, bounded lock wait, separate public sessions, shared/exclusive fairness and parent-stall regression. |
| UIA-08 | PASS — stale epoch/ref, detached, zero/multiple/disabled targets, literal shell-like fill text and no action replay. |
| UIA-09 | PASS — busy stop, repeated terminal digest, signal/preparation interruption, browser/controller death and exact recovery ownership. |
| UIA-10 | PASS — partial acquisition/release paths, retained failed ownership, every normalized failure paired with cleanup, unsafe terminal publication, existing atomic-state and cleanup owner tests. |
| UIA-11 | PASS — delayed fonts, real geometry deadline, post-screenshot mutation rejection and stable-geometry dynamic text. |
| UIA-12 | PASS — missing rendered row then explicit scroll/reveal; focus and scroll preserved during capture. |
| UIA-13 | PASS — exact current synthetic failed evidence and actual canonical auth import; unregistered/null fixture and expected-only reporting preserved. |
| UIA-14 | PASS — unsupported reconciliation/report version, divergent renderer/runtime/registry, tampered bytes and duplicate attachment joins rejected. |
| UIA-15 | PASS — viewport/element/region scope math, outward rounding, asymmetric nonunit scales, clipping and no repeated zoom application. |
| UIA-16 | PASS — secure anchored reads/publication, links/permissions/digests/counts, exclusive manifest commit, existing bundles unchanged after failure. |
| UIA-17 | PASS — request, PNG byte/dimension/pixel, component/file/bundle/storage/count bounds, schema limits and explicit channel truncation; required data never silently truncated. |
| UIA-18 | PASS — zero/full/one-pixel, hidden transparent RGB and alpha-only differences, exact magenta output, unequal dimensions and integer ties-up rounding. |
| UIA-19 | PASS — exact crop pixels/order, two-pixel overlay transform, source-channel contact sheet, complete matched metadata and explicit reference comparison. |
| UIA-20 | PASS — completed violations/incomplete/empty, disabled snapshot/capture and unavailable imports remain distinct; findings do not fail commands. |
| UIA-21 | PASS — engine failure/timeout, wrong version, malformed targets and occurrence/string overflow fail without a partial bundle; live iframe contents excluded. |
| UIA-22 | PASS — real offline report open, escaped script/URL probes, CSP, native dimensions/zoom, missing data and cache tampering. |
| UIA-23 | PASS — exact image bytes/manifest digests and local report inspection agree; public workflow proves all private links expire and structural receipts persist. |
| UIA-24 | PASS — sentinel page/console/URL/fill and inherited telemetry remain private; generic tee/preflight/output metadata are bypassed by the dedicated recipe. |
| UIA-25 | PASS — restricted path checks precede filesystem access; README/Markdown aliases, links and neutral policy fixtures; source-boundary owner checks remain intact. |
| UIA-26 | PASS — complete failure mapping, each primary plus cleanup, repeated stop, and operation failure distinct from normal terminal lifecycle. |
| UIA-27 | PASS — human editorial cross-reference review, current reconciliation-v3 adapter rejection of v2, existing contracts and Markdown lint. |
| UIA-28 | PASS — public seeded editor/viewer, separate dev, browser-free import/reimport/analysis/report/stop; canonical visual and accessibility rows plus design-review compatibility evidence. |

Final serialization/lifetime additions and handoff lint are recorded in the
completion checkpoint after their results, rather than inferred from this matrix.

Qualification archive before disposable storage removal: resolved Linux x64 native packages are `@img/sharp-linux-x64@0.35.4` and `@img/sharp-libvips-linux-x64@1.3.3`; their exact integrities remain in the package-manager-owned `pnpm-lock.yaml`. Sharp requires Node >=20.9.0 and Playwright >=18; the selected Node 24.15.0 satisfies both. Resolved native library versions: `{"aom":"3.15.0","archive":"3.8.9","cairo":"1.18.4","cgif":"0.5.3","exif":"0.6.26","expat":"2.8.3","ffi":"3.8.0","fontconfig":"2.18.3","freetype":"2.14.3","fribidi":"1.0.16","glib":"2.89.4","harfbuzz":"14.3.1","heif":"1.23.2","highway":"1.4.0","imagequant":"2.4.1","lcms":"2.19.1","mozjpeg":"0826579","pango":"1.58.2","pixman":"0.46.4","png":"1.6.58","proxy-libintl":"0.5","rsvg":"2.62.91","sharp":"0.35.4","tiff":"4.7.2","uhdr":"2.0.2","vips":"8.18.6","webp":"1.6.0","xml2":"2.15.3","zlib-ng":"2.3.3"}`. The current repository lock SHA-256 is `e1db2a82e1433557a34ec6509d802cc6adb5ad254993f5fde87a29698c50fb31`.


### WS6 completion checkpoint and final handoff

**Status: DONE.** All six workstreams completed in dependency order; each prior
completion checkpoint was recorded before its successor began. All UIA-01–UIA-28
obligations are satisfied by the evidence above and the final checks below. No
required acceptance blocker or implementation slice remains.

The last coverage additions prove that a second operation cannot overlap the
first, contention fails after the adopted wait, and expiry closes exactly at the
eight-hour boundary using a controlled clock. The final routed suite contains 29
review cases, with the request/command cases additionally routed through the
command-surface owner.

| Final checkpoint check | Result / retained root |
| --- | --- |
| `make generate` after the final test additions | PASS `20260927T052655Z-p48490` |
| `make agent-finalize` before final owner verification | PASS `20260927T052839Z-p51818`; generated structure unchanged; schema, catalog, tier and drift checks passed |
| `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.private_ui_review_session_ownership_and_strict_b_b785350fa4` | PASS `20260927T052924Z-p55854`; all 29 review cases |
| `make test-slice OWNER=harness.generated_artifacts` | PASS 5/5 `20260927T052949Z-p62216` |
| `make test-slice OWNER=harness.command_surface` | PASS `20260927T052950Z-p62435` |
| `make lint-markdown` for the guide and acceptance handoff | PASS `20260927T053151Z-p70912` |
| `git diff --check` | PASS |
| Exact review registry / host admission registry | Zero live records / zero leases after all checks |
| Disposable qualification storage | Removed after archiving structural package/native qualification above |
| Public session terminal receipts | Seeded, dev/artifact test workflows, canonical import and capture sessions closed with cleanup complete; all returned private links removed |

Adopted owner: `cartulary.testing_harness.v3`, amendment `ui-review-api-1`,
requirements and AC-102–AC-129 integrated into their existing owner sections.
Machine projections are the eight `cartulary.ui_review_*.v1` schemas and seven
`cartulary.harness.command.ui_*.v1` public command identities. The commands are
`ui-review`, `ui-review-status`, `ui-browser`, `ui-capture`, `ui-analyze`,
`ui-review-report`, and `ui-review-stop`. Their authored recipes, helper-only
classification, input/output policies and backing inventory live in
`tools/task_surface_owner.json`; generated projections were produced through Make.

Implementation ownership is split among shared review preparation, private
session/controller and Playwright adapter, source/PNG/bundle validation, image and
axe analysis, offline report rendering, secure local publication and shared host
admission. The changed scheduler/provider seams retain existing service owners;
the canonical auth visual and accessibility rows and existing design-review smoke
passed. No product database migration, golden refresh or product-row requirement
change was made. The added harness row verifies helper mechanics only.

Dependencies are exactly the qualified Node/Playwright/Sharp/axe combination above,
with one Playwright/core version and Linux x64 native packages. Qualification is
limited to the observed WSL2 Linux environment; a separate native-Linux host,
Windows-native and macOS are not claimed tested. Existing interactive design-review
commands remain supported through the same preparation/teardown owner. There is
no standalone CLI dependency, legacy schema reader, conversion, compatibility
alias, newest-artifact selection, alternate browser fallback or durable export.

Privacy/lifetime limits remain eight hours, 100 bundles and 512 MiB private
outputs, with the per-request/component/image/bundle/report bounds adopted in the
NLSpec. Original bytes, observations and reports are session-private; only closed
structural receipts/locators persist. Repeated stop preserves the terminal outcome.
Expected-only imports and unavailable accessibility/DOM channels remain visibly
unavailable. Command success and diagnostic findings never imply product,
canonical visual, accessibility-complete, release or Core 05 publication success.

Inspected authorities were the domain vocabulary/owner navigation, NLSpec
standard, adopted Testing Harness NLSpec and this controlling artifact. Substantive
changes include that NLSpec and tracker, the existing browser design-readiness
guide, package/pin/lock inputs, task/schema/test routing inputs and generated
projections, shared review lifecycle/provider/runtime seams, new UI-review modules
and their contract/lifecycle/bundle/analysis/public-workflow tests. No competing
plan was introduced.

Full repository `check`, CI and release gates were not run: owner-scoped checks,
full browser owner, selected canonical visual/accessibility rows and explicit
public workflows cover the changed boundaries without asserting a release.
Retained warm-run maintenance was skipped because no eligible full warm run was
supplied as `RESULTS_DIR`. Optional OCR, Python, video, MCP/hosted analysis and
future product promotion of axe rules remain explicitly outside this completed
scope. Further host qualification is additional platform evidence, not an implied
passing result.

Rollback before public acceptance is to stop owned review sessions, revert the
new authored additions/changes and regenerate downstream outputs through Make;
borrowed development data and canonical artifacts remain untouched. After public
adoption, retiring these versioned interfaces requires a superseding amendment.
At that checkpoint the next action was ordinary code review; no iteration-1
implementation slice remained. Section 18 records the subsequently requested
production-readiness iteration.

## 18. Iteration 2 — production readiness plan

### 18.1 Scope and source posture

**Planning status: DONE. Implementation status: DONE (R2-W1–R2-W7).** This iteration targets
`tools/harness/browser/ui-review/` (`testing-harness-ui-review`), including its
shared preparation seam, public command projections, verification routing, and
human consumption workflow. The inspected branch is `main`, commit
`25806426bae985528ee14de9e2595316da93ba8f`; the working tree was clean before this
document edit. Inspection occurred on 2026-09-28. Findings below are source-review
evidence, not newly reproduced failures or passing runtime evidence.

Production readiness means a dependable **private local development/review tool**:
bounded operations, responsive cancellation, exact resource recovery, trustworthy
observations, maintainable owner boundaries, and repeatable validation on its
declared platforms. It does not mean a production application service, a remotely
accessible browser agent, a release gate, or Core 05 publication approval.

The user authorized implementation of R2-W1–R2-W7 on 2026-09-28. The earlier
planning-only write restriction is historical. The user requests clean structural improvements, so
accidental implementation behavior is not frozen merely because a test currently
asserts it. Proposed public behavior changes still require owner adoption before
implementation. No source, test, schema, dependency, configuration, generated
output, golden, design owner, or domain owner is amended in this planning step.

Source posture and scope:

| Source | Use in this plan | Boundary |
| --- | --- | --- |
| `AGENTS.md`; `cartulary_modular_refactor_planning_framework.md`; Refactor Tracker format | Repository procedure and planning structure. | Framework module catalog is not proof that this harness needs a Go module, domain context, or public package. |
| Testing Harness NLSpec, amendment `ui-review-api-1` | Behavioral owner: §§4.7, 5.6, 6.5, 7.5, 8.4–8.8, 9.2, 10.6, 11.9–11.11, 13.6, 14.1, 15.3, 16.1, 17.0. | Owner-first amendments; no executable reader of Markdown. |
| `docs/domain.md` §§1, 4, 6 | Vocabulary and owner navigation. | Review bundles are diagnostic files, not domain `artifact` records or evidence envelopes. No new domain concept is needed. |
| `docs/design.md` §§1, 16; inspected token/presentation context | Browser-application design authority and exclusions. | Harness findings cannot redefine workbook presentation. An offline diagnostic report does not inherit a requirement to recreate the product shell/theme. |
| Current source, typed schemas, task and verification manifests | Actual implementation, contracts, callers, and routing. | Code and tests can expose defects; they cannot amend their adopted owner. |
| Sections 1–17, existing guide and review skill | Prior decisions, historical results, current operator guidance. | Historical passes are not this baseline's validation; source-document instructions do not authorize execution. |

Keep the three modes, pinned API adapter, useful interactive design-review
preparation, exact immutable imports, advisory axe, and offline reports. Continue
to exclude legacy readers, compatibility aliases, auto-discovery, driver fallback,
durable private exports, hosted analysis, and automatic golden updates. OCR,
Python, video, alternative drivers, and a plugin framework remain `DEFERRED`;
there is no demonstrated need for them in this iteration.

### 18.2 Current-state repository inventory

All **22 files** in the target directory were opened. Paths in the next table are
relative to `tools/harness/browser/ui-review/`. Every listed export is an internal
module interface unless the row identifies the Make/JSON boundary. Test keys:
**C** = `test-ui-review-contract.mjs`, **L** = `test-ui-review-lifecycle.mjs`,
**B** = `test-ui-review-bundles.mjs`, **A** = `test-ui-review-analysis.mjs`,
**W** = `test-ui-review-workflow.mjs` and `ui-review-public-workflow.mjs`,
**I** = `ui-review-import-fixture.mjs`, all under
`tools/harness/browser/tests/`. These are inspected tests, not current passes.
**V1** means the relevant `cartulary.ui_review_*.v1` projections; **M** means
authored Make/task metadata and its generated projections. No target source file
is a generated output.

| Path | Current responsibility | Exported/public symbols or surface | Inbound callers | Outbound dependencies | Tests touching it | Contracts/generated surface | Owner candidate | Risk | Notes |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `launch.sh` | Pinned Node startup and pre-Node error output | Public Make launcher | Generated `ui_review` recipes | Node, CLI | C/L/W | M; result V1 | Command entry | High | Preserve transient-output bypass. |
| `cli.mjs` | Full readiness, command dispatch, fallback JSON | Seven command entry points | Launcher | Toolchain, inputs, controller, failure | C/L/W | M; result V1 | Command entry | High | Full profile checked twice, before every command. |
| `inputs.mjs` | Source-aware Make input resolution | `resolveInputs` | CLI | Authored task owner, output resolver | C | M; input contracts | Command admission | Medium | Keep explicit UI assignments and mode exclusions. |
| `output.mjs` | Transient result presentation | `emitResult` | Controller | Contract validator | C/L/W | Result V1 | Output projection | High | Never move into generic stdout tee. |
| `contract.mjs` | IDs, limits, defaults, cross-field validation | `validate`, `parseRequest`, `result`, `emptyCounts` | Most target modules/tests | Shared schema/JSON validation, failures | C/L/B/A | All V1 | Review contracts | High | Separate stable vocabulary from resource/runtime owners. |
| `failure.mjs` | Closed diagnostics and mapping | `ReviewFailure`, mappings/record | Contract, CLI, toolchain | None | C/L | Result/receipt V1 | Review contracts | High | Preserve primary ordering and signal exits. |
| `toolchain.mjs` | Pin/native resolution and root location | `toolProfile`, `repoRoot`, doctor mode | CLI, doctor, most modules | Package/lock/pin files, Sharp | C/W | Tool profile V1; pins | Readiness adapter | High | Lightweight control must not import native analysis. |
| `controller.mjs` | Client IPC, parent supervision, controller server | `execute`; private `send/start/serve` | CLI and self-fork | Sockets, process fork, session, recovery | L/W | Result/session V1 | Session transport/composition | High | Framing, parent acknowledgement, cancellation share one file. |
| `session.mjs` | Modes, locks, operations, storage totals, shutdown | `ReviewSession` | Controller; tests | Broker preparation, browser, runtime, admission, all operations | L/B/A/W | Most V1 | Session coordinator | High | Broad mutable object passed to operation owners. |
| `session-files.mjs` | Path boundary, registry, locator, receipts, hash helpers | Input/read/publish/register/resolve/terminal helpers | Controller/session/recovery, source, bundles/tests | Secure files, process identity, restricted inputs | L/B/W | Session/receipt V1 | Split registry identity from artifact IO | High | Live-only resolver blocks dead-controller attachment. |
| `recovery.mjs` | Parent-triggered controller-death cleanup | `recoverSession` | Controller parent | Runtime, preparation recovery, receipts | L | Session/receipt V1 | Lifecycle recovery | High | Distinct terminal builder; depends on original input. |
| `browser.mjs` | Browser contexts, targeting, action epochs, observations | `ReviewBrowser`, `browserReady`, `unavailableAxe` | Session/capture; tests | Playwright, seed TOTP, contracts | L/B/A | Action/observations V1 | Browser adapter | High | DOM sampling and actor login currently share adapter. |
| `capture.mjs` | Stable page capture, imports, image crop, bundle construction | `execute`, `scopeRectangle` | Session; tests | Browser internals, Sharp, PNG, source/import/store | B/A/W | Capture/bundle/observations V1 | Capture coordinator | High | Accesses page, generation, seeded details directly. |
| `source.mjs` | Canonical catalog/profile binding and source reads | `pageBinding`, `catalogRow`, `readJSON`, containment | Capture/import | Current catalog, fixture/golden manifests | B | Current producer schemas; bundle V1 | Source adapter | High | Exact identity joins stay here, not in report code. |
| `canonical-import.mjs` | Full run/result/attachment association | `importCanonical`, `selectCaptureResult` | Capture; tests | Source adapter, pinned report format, bundles | B/I | Reconciliation v3, target v4, group v6, bundle V1 | Canonical adapter | High | Dense join procedure; uses current workspace source refs. |
| `bundles.mjs` | Build, validate, publish and load immutable bundles | Bundle builders/store functions | Capture, analysis, report | Secure files, PNG, contracts, session budget/maps | B/A/W | Bundle V1 | Private artifact store | High | Publication and rollback mutate session-owned maps. |
| `png.mjs` | PNG framing/CRC/profile checks and native decode/encode | `pngHeader`, `decodePNG`, `encodePNG`, `crc32` | Capture, store, algorithms, report | Sharp, limits | B/A | Encoding/limit contract | Raster adapter | High | Native decode runs outside image worker on several paths. |
| `image-worker.mjs` | Analysis subprocess and cancellation | `runImageWork`; self-fork | Analysis | Child process, algorithms | A | Private job protocol | Bounded work executor | High | Current isolation covers compute, not all raster work. |
| `image-algorithms.mjs` | RGBA difference, crop, overlay, contact sheet | Pure raster functions, `computeImages` | Worker; tests | PNG/Sharp | A | Image algorithm contract | Image analysis | Medium | Preserve exact math; avoid new tolerance policies. |
| `analysis.mjs` | Parent provenance, compatibility, admission, result assembly | `execute`, `compatible`, `primary` | Session; report imports `primary` | Store, worker, host admission | A/W | Analysis/bundle V1 | Analysis coordinator | High | Resource ownership and presentation helper are mixed. |
| `axe.mjs` | Advisory engine invocation and normalization | `observeAxe`, `normalizeAxe`, tags | Capture; tests | Axe, contracts | A | Observations V1 | Accessibility adapter | Medium | Keep explicit incomplete/unavailable/failure outcomes. |
| `report.mjs` | HTML/CSS/JS, assets, cache, publication | `renderReport`, `execute`, escaping | Session; tests | Store, analysis helper, PNG, secure files | A/W | Report/private refs; bundle V1 | Report presenter + store consumer | Medium | Rendering and persistence currently share implementation. |

Adjacent seams were inspected to establish callers and ownership, not to open
unbounded refactors:

| Seam | Evidence inspected | Scope/disposition |
| --- | --- | --- |
| Shared review preparation | Entire `review-preparation.mjs` and `design-review.mjs` | Reuse the broker/service owners. Recovery currently constructs `design-review-<profile>-allocation-001`; replace the naming dependency with returned exact ownership data. |
| Runtime security/admission | `secure-local-files.mjs`, `host-admission.mjs`; close, borrow, private-path and stale-cleanup sections of `suite-runtime.mjs` | Existing authority for secure IO, process proof, admission and suite-private lifetime. Extend only the narrow missing primitive; do not duplicate these systems. Uninspected provider internals remain out of scope until a changed port requires them. |
| Commands and generation | UI-review recipes/targets and contract-test recipes in `task_surface_owner.json`; UI recipe in `task-surface/make-renderer.mjs`; generated policy search | Authored projections change before Make regeneration. Existing generated task/topology outputs are not editing targets. |
| Contracts/readiness | Eight schema attachments; selected action/capture/result/bundle fields; `contract.mjs`; doctor UI-profile invocation | Eight public identities remain tracked. Full future schema audit belongs to R2-W1; no claim that sampling proves every constraint. |
| Verification | All seven review test/helper files named above, `harness.browser.json` review row, command-surface routing, public owner guidance | Most execution tests enter one static/standard shell row. Public test calls `publicWorkflow()` with `seeded=false`; seeded helper branch exists but is not invoked by that wrapper. |
| Human consumers | Existing browser design-readiness guide and repository `cartulary-ui-review/SKILL.md` | Maintain progressive disclosure. Current capture returns a bundle manifest ref; image consumers must resolve its exact components. |
| Product and optional code | Product UI, Go feature modules, goldens, optional OCR/video/MCP integrations | Explicitly excluded. No product behavior/design refactor is justified by this harness iteration. |

### 18.3 Module boundary diagnosis

The target is a local orchestration subsystem with browser, filesystem, native
image and producer-format adapters. It is not a domain bounded context and does
not need a separately published package. Preserve its repo-local placement while
making the decisions below private behind small, concrete interfaces.

| Responsibility found | Current location | Correct owner candidate | Keep / move / split / defer | Evidence | Notes |
| --- | --- | --- | --- | --- | --- |
| Public command admission and transient result | CLI/inputs/output/launcher | One command composition boundary | keep | Dedicated Make recipe bypasses retained capture | Dispatch control commands before optional engine readiness only after adoption. |
| Session state, operation ordering and cancellation | Session plus controller timers | Session coordinator; IPC transport is separate | split | `handle/operation/finish`, `serve/end` | One state-transition owner; no owner mutates another owner's state flags. |
| Mode resources and exact recovery proof | Session/preparation/recovery | Mode preparation adapters using existing resource owners | split | Inline seeded/dev branches; reconstructed allocation name | Return exact handles/proofs at acquisition; account for late acquisition during stop. |
| Private bytes, immutable publication, rollback and lookup | Session/bundles/report/session-files | Session-scoped artifact store over secure runtime IO | move | Operations share mutable maps/counters and duplicate rollback | Store owns reservation/commit/release; callers receive immutable refs. |
| Expensive raster execution and budgets | PNG, capture, store, image worker, analysis | One bounded work boundary over pinned native tools | split | Decoding occurs both in controller and subprocess | Retain algorithms; make every expensive path cancelable and admitted. |
| Browser targeting and observation | Browser/capture/axe | Browser adapter plus observation/capture coordinator | split | Capture reads adapter internals and seeded runtime details | Narrow observation handle hides Playwright/CDP; no new generic driver framework. |
| Canonical producer-format interpretation | Import/source; comparison reads raw metadata | Canonical adapter with normalized private source view | keep | Exact join logic, `compatible` | Preserve frozen public provenance; consumers should not redo joins. |
| Offline presentation | Report plus `analysis.primary` | Report renderer consuming store/read model | split | Rendering imports execution coordinator | Share primary-image selection with bundle model, not analysis execution. |
| Future tool plugins or exported SDK | None required | No current owner | defer | No second production adapter/use case | Growth means local changes at real boundaries, not speculative abstraction. |

Directional rule: command composition selects session/mode capabilities; session
admits operations and publishes structural outcomes; operations use a narrow
artifact-store and browser/work capability. Adapters do not import the session
coordinator. Presentation does not import execution/resource acquisition. Private
helper exports with only one caller need no compatibility shim. Consolidate small
helpers when they do not hide an independent decision; file count is not a goal.

### 18.4 Public contract and behavior freeze map

Freeze adopted behavior, not internal object layouts or historical test setup.
The following surfaces have an owner and a characterization obligation. Proposed
departures are explicitly called out; all implementation remains a later task.

| Contract | Current owner | Evidence | Existing tests | Required characterization tests | Refactor risk | Notes |
| --- | --- | --- | --- | --- | --- | --- |
| Seven Make commands, input source rules, helper-only selection | Harness §§4.7, 5.6, 7.5; task owner | Dedicated recipe and closed input resolver | C | All finite commands via Make, invalid input before side effects, exact one-object output, exclusion from product/default/CI/release selections | High | Keep names/IDs and no aliases. Proposed readiness exception below needs adoption. |
| Mode and service ownership | REQ-415/416, shared preparation | Fresh seeded resources, explicit borrowed dev, browser-free artifacts | L/W; historical seeded smoke | Both seeded profiles and actors through a routed public workflow; partial/late acquisition | High | Preserve existing interactive design-review interface. |
| Epochs, target handles, action effects | REQ-159/417 | Snapshot refs, exact matches, consumed epoch, uncertain timeout | L/B | Sequential/concurrent public calls, cancellation before/after admission, detached targets and navigation | High | No replay/force/eval; internal handles are not public compatibility. |
| Terminal status, stop and cleanup | REQ-416/502/616 | Stop cache, terminal refs, registry and recovery | L | Parent+controller loss, missing receipt, damaged engine install, hung release, cleanup failure plus private-data purge | Critical | Correct existing owner mismatches; new recovery availability is an owner amendment. |
| Private/retained boundary | REQ-081/285/616 | Private runtime plus closed receipts and special launcher | C/L/W | Sentinel across argv metadata/stdout/stderr/telemetry/receipts; injected publication and cleanup failures | Critical | No durable detail export or raw-error retention. |
| Bundle/source identity and exact imports | REQ-284/285/552 | Eight V1 schemas, exact run/capture joins | B/A | Valid producer fixtures plus missing/duplicate joins, supported failed runs, drifted source refs, illegal source/component combinations | High | Current forms only; no archive compatibility layer. |
| Capture, image math and axe | REQ-284/286/287 | Three-frame/post-image check; exact RGBA; main-frame advisory scan | B/A | Native-work cancellation, boundary budgets, navigation/disposal during observation, all axe states | High | No automatic masks/alignment/tolerance, missing data stays explicit. |
| Offline report and local consumption | Harness §8.8, REQ-616/679 | Escaped CSP viewer, digest-checked cache | A/W | Images resolved from bundle refs, keyboard-operated zoom/slider, no network, cache corruption, expiry | Medium | Diagnostic UI; no product redesign or accessibility-complete claim. |
| Product HTTP/WS/records/views/revisions/selectors | Core/product owners; unchanged here | Authentication calls ordinary login; targets observe existing DOM | Existing product rows, only if shared seam changes | Select current canonical browser/accessibility/visual rows by affected owner | High if widened | No new route, bypass, stored field, DB migration, view schema, selector policy, or golden change is proposed. |

R2-W1 must adopt the explicit control-command dependency profile and any new
bounded-work/admission rules before code changes. Keep current external schemas
where their meaning is unchanged. If a required semantic change cannot be
expressed faithfully by V1, adopt one successor and a deliberate cutover: stop old
sessions, preserve their structural evidence, and reimport explicitly selected
sources. Do not add dual readers, implicit conversion, aliases, or version-shaped
internal branches. Retiring an adopted interface needs a superseding amendment.

### 18.5 Coupling and boundary findings

Classification is `must_fix`, `should_fix`, `defer`, or `intentional/no_action`.
Priority P0 precedes P1; P2 follows the risk-reducing foundations. A source-level
finding is not a claim that its failure was reproduced in this planning session.

| Finding | Evidence | Risk | Classification | Proposed owner | Required planning action |
| --- | --- | --- | --- | --- | --- |
| R2-F01 — Distributed lifecycle decisions and broad mutable operation context | `session.handle/operation/finish`; `controller.serve`; operations accept the full session and mutate its maps/leases | Growth adds temporal coupling; late acquisition/publication can outlive cancellation | must_fix, P0 | Session coordinator and mode owners | R2-W1/W2: explicit transition/admission boundary, narrow capabilities, cancel/drain/terminal ordering with fault injection. |
| R2-F02 — Recovery and private-data purge share failure-sensitive teardown | `finish` skips `runtime.close()` after any earlier cleanup failure; resolver rejects dead process; recovery is called by surviving parent; preparation reconstructs allocation name | Private content can remain after failed release; loss of both supervisors lacks proven exact recovery | must_fix, P0 | Lifecycle recovery and suite-runtime owners | R2-W2: separate removable detail from minimum recovery proof; bounded best-effort release; persist exact owner handles; prove double-death handling. |
| R2-F03 — Control requires analysis dependencies; terminal evidence has exceptional-path gaps | CLI calls full `toolProfile()` before every command; `terminal()` defaults status to success with null receipt; `finish` unregisters on no cleanup failures even if publication failed | Dependency damage obstructs stop/status; missing proof can be misread or recovery records dropped | must_fix, P0 | Command readiness and terminal publication | R2-W1/W2: adopt minimal control profile; enforce REQ-502 null-receipt artifact failure and preserve required proof. No ambient fallback. |
| R2-F04 — Artifact operations own storage mutations and failure rollback | Bundle/report `execute` manage maps and reservations; both release reservation before recursive rollback; rollback error can replace the original | Multiple extensions would duplicate commit logic, mis-account surviving staging files, or obscure primary failure | must_fix, P0 | Private artifact store | R2-W3: one transaction/reservation owner; account for failed cleanup, preserve both errors, close source/component invariants. |
| R2-F05 — Expensive work is not uniformly isolated or budgeted | `decodePNG` called from capture/store; report renders in controller; analysis worker receives copied buffers; offline operation admission acquired only inside analysis | Large valid inputs can delay control, multiply memory use, or contend with quiet measurements on paths not covered by the current analysis-only admission rule | must_fix, P1 | Work executor and existing host admission | R2-W1/W4: close execution/admission policy; bound aggregate in-flight work, isolate native work, keep controller responsive. This is a risk needing stress proof, not a measured OOM. |
| R2-F06 — Capture and import knowledge crosses adapter boundaries | Capture reads browser page/generation and seeded internals; comparison reads nested reconciliation/fixture data; importer performs all joins inline | New source channels require edits across coordinators; ownership and provenance become harder to review | should_fix, P1 | Capture and canonical adapters | R2-W5: immutable normalized source/observation views; retain exact public provenance and current-only input checks. |
| R2-F07 — Repeated semantic facts and permissive cross-field cases | Engine version repeated in toolchain/axe/contract/import; bundle schema permits nullable components broadly; tests construct expected-only reference-image bundles | Pin/semantic drift can validate impossible evidence combinations or require synchronized manual edits | should_fix, P1 | Review contracts and bundle model | R2-W1/W3/W5: audit owner-required invariants; one typed source for shared facts and explicit variant validators. Do not infer required combinations without owner review. |
| R2-F08 — Repeatable acceptance is coarser than the previous handoff | One static/standard row launches L/B/A/W; public wrapper never selects seeded helper branch; recovery test kills controller only | Broad passing row obscures which boundary/platform actually ran; manual seeded evidence is hard to reproduce | must_fix, P1 | Harness browser/command-surface verification owners | Add semantic routed tests with each slice; final R2-W7 matrix must show automated versus manual evidence and exact omissions. |
| R2-F09 — Presentation and agent consumption depend on incidental structure | Report imports `analysis.primary` and embeds rendering/publication; skill says open returned image paths, while capture returns a bundle ref | Presentation growth couples to execution; agents may capture without locating/inspecting pixels | should_fix, P2 | Report presenter and human workflow docs | R2-W6: immutable view model/store consumer; explicit component-resolution recipe and real local-consumption trial. |
| R2-F10 — Existing owner boundaries are useful | Secure anchored IO, host arbiter, shared preparation, current-only adapter, three modes | Replacing these with another runtime/plugin/driver adds compatibility burden | intentional/no_action | Existing owners | Keep and strengthen their narrow ports; no parallel infrastructure or driver/tool upgrade for convenience. |

The following dispositions make the long-term tradeoffs explicit:

| Finding | Remediation and areas | Rationale and long-term benefit | Compatibility/migration impact | Risk if unresolved | Validation for completion |
| --- | --- | --- | --- | --- | --- |
| R2-F01 | Implementation + tests; owner clarification only for newly specified transitions | One state owner makes cancellation, ordering, and additional operations locally understandable | Internal API moves; preserve command/epoch semantics | New operations inherit hidden timing dependencies | Transition/race tests and public concurrent commands prove one admission, one effect, one terminal outcome. |
| R2-F02 | Specification clarification + runtime/preparation/implementation + tests + recovery guide | Separate privacy deletion from resource-recovery proof; exact handles survive process factoring | Stop existing sessions before private-layout cutover; no legacy registry reader | Stale resources or sensitive details require manual cleanup | Every release attempted within its owner deadline; detail absent when safely deletable, residual proof minimal; borrowed resources survive; double-death exact recovery verified. |
| R2-F03 | Specification + CLI/readiness/terminal implementation + command tests + docs | Recovery remains available when an unrelated engine is damaged; evidence semantics remain honest | Intentional change: validated status/stop no longer require Sharp/axe/Playwright readiness. Still require pinned Node, contract validation, and applicable recovery owners | Operator may be unable to stop; absent receipts mistaken for proof | Missing/wrong analysis package tests leave control usable and data work rejected; null receipt is artifact failure; repeated terminal receipt stays immutable. |
| R2-F04 | Implementation + semantic validators/schemas where justified + security tests | A transactional store centralizes bytes, provenance, publication and rollback | No public shape change intended; reject owner-invalid combinations rather than preserve test artifacts | Partial storage/accounting corruption spreads to every producer | Fail after each publication step, fail rollback, tamper inputs, fill all limits, and assert no published partial bundle or hidden secondary failure. |
| R2-F05 | Specification + executor/resource policy + implementation + stress tests | One bounded execution path supports future operations without degrading control responsiveness | Any new externally observable cap or admission rule is adopted first; current image algorithms stay exact | Timeout/cancellation claims weaken under large valid workloads | Upper-bound inputs, native stall, abort during decode/render/publication, and quiet contention prove bounded completion, reaping and accounting. |
| R2-F06 | Implementation + adapter fixtures/tests; spec only if actual meaning changes | Vendor/current-producer knowledge stays local; operations consume Cartulary observations | Keep current-only sources and exact joins; no cross-checkout/archive reader added | New formats require edits throughout analysis/report/lifecycle | Producer-shaped fixtures exercise successful/failed/expected-only and ambiguous captures; parent hashes and comparison identity unchanged. |
| R2-F07 | Specification review + machine projections + semantic tests | Single typed facts and valid variants reduce synchronized maintenance and impossible states | Preserve valid V1 inputs; owner correction may tighten invalid acceptance; deliberate successor only if meaning changes | Pin drift and fabricated channel/source combinations erode trust | Exhaustive source/channel truth table and pin mismatch tests; no executable Markdown dependency. |
| R2-F08 | Tests + authored routing + generated projections + evidence docs | Pure contracts, browser behavior and service-backed ownership can be run and assessed separately | Internal row selection may change with regenerated current mappings; seven helper commands stay outside product selections | Historical manual checks silently substitute for repeatable readiness | Every required obligation has a routed test or explicitly named manual trial; seeded editor/viewer and both profiles have reproducible Make-owned execution. |
| R2-F09 | Report implementation/tests + guide/skill documentation | Presentation depends on a stable read model; agents actually inspect the selected evidence | Keep private offline report and expiry semantics; no export feature or new product design | Reports become costly to extend and capture success is confused with review | Native-size image inspection, keyboard controls, expected-only limitations, escaping/no-network, digest cache and post-stop expiry all verified. |

### 18.6 Refactor workstreams

Each workstream is exactly one implementation slice with its own checkpoint.
Execute in order **R2-W1 → R2-W2 → R2-W3 → R2-W4 → R2-W5 → R2-W6 → R2-W7**.
Sequential execution limits simultaneous ownership changes; it is not a runtime
module or test-row naming scheme. Do not advance past a required blocked exit.

| Workflow ID | Name | Class | Required previous | Required subsequent | Goal | Files likely involved | Validation | Handoff checkpoint |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| R2-W1 | Owner decisions and contract characterization | root | Later implementation authorization | R2-W2 | Close new behavior and map current acceptance before movement | NLSpec, schemas/contract, task/verification owners, focused tests | C; shape/command checks; editorial review | Adopted clause/AC map, source/channel matrix, baseline results and version decision |
| R2-W2 | Session control, ownership and recovery | chain | R2-W1 | R2-W3 | One lifecycle owner with usable, bounded control and exact recovery | CLI/toolchain/controller/session/session-files/recovery; shared preparation/runtime seams | L/C, service cleanup, public stop/status, design-review smoke | Resource acquisition/release proof, death matrix, terminal/privacy evidence |
| R2-W3 | Transactional private artifact store | chain | R2-W2 | R2-W4 | One reservation/publication/load/rollback owner | Bundles, report persistence, session accounting, secure IO, semantic validators | B/A, publication faults and retained-boundary checks | Ownership and byte accounting for success and every failure stage |
| R2-W4 | Bounded raster and report work | chain | R2-W3 | R2-W5 | Responsive control while costly operations run under explicit budgets | Image worker/algorithms/PNG, capture/store/report execution, admission | A/B/L, cancellation/limit/quiet-contention fixtures | Boundaries, timings, child reaping and admission records |
| R2-W5 | Browser and canonical source adapters | chain | R2-W4 | R2-W6 | Hide source/tool internals behind immutable operation inputs | Browser/capture/axe/source/canonical-import/analysis, producer fixtures | L/B/A plus affected canonical rows | Exact-source association and capture invariants; no compatibility branches |
| R2-W6 | Offline presentation and review consumption | chain | R2-W5 | R2-W7 | Clear local evidence consumption independent of execution internals | Report/bundle read model, guide, repository review skill | Offline viewer tests and agent/human trial | Inspected pixels/report, privacy/expiry proof, progressive-disclosure guide |
| R2-W7 | Validation and handoff completion | chain | All prior exits | None | Requalify this iteration and retire superseded internal paths | Routed tests, authored manifests, generated outputs, guide and this tracker | Full applicable matrix and public Make workflows | Exact evidence, platform support, cleanup, limitations and final readiness disposition |

### 18.7 Proposed refactor slice plan

All slices **are authorized by the 2026-09-28 implementation request**. `Preserve` below means
adopted observable behavior; `correct` repairs an owner mismatch; `change` requires
R2-W1 adoption. A test that preserves a defect must be replaced with owner-backed
characterization rather than treated as a compatibility veto.

| Slice ID | Depends on | Intended change | Files/packages likely involved | Contract risks | Tests to add/preserve | Validation command | Rollback note | Completion criterion |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| R2-S1 / R2-W1 | Authorization | Change: define minimal control readiness, expensive-work admission and recoverable proof disposition. Preserve all useful public behavior. Audit cross-field requirements and actual coverage | Owning NLSpec first, then typed inputs and C/B fixtures | Accidentally relaxing recovery identity or redefining retained data | Current V1 positives/negatives; source/channel matrix; dependency failure contracts | `make test-slice OWNER=harness.command_surface`; `make json-shape-check`; Markdown review | Revert unadopted proposals; no runtime movement until adopted | Every F01–F09 has an owner, adoption decision, acceptance test and route; no unresolved owner contradiction. |
| R2-S2 / R2-W2 | R2-S1 | Preserve: explicit coordinator/mode ports. Correct: receipt/proof failure paths. Change: light control and supported exact recovery after both supervisors are gone | Lifecycle files and existing runtime/preparation owners | Hung releases, PID reuse, stop during prepare, changed internal registry layout | Fault each acquire/release/publication; stalled child; missing dependencies; double death; borrowed survival | Browser row below; command-surface slice; `make service-backed-test-slice OWNER=harness.browser`; `make browser-design-review-smoke` | Stop owned sessions before reverting private layout; never delete proof to permit rollback | Stop is responsive; every owned release attempted once; bounded failure retains only necessary proof; valid terminal repetition is byte-stable. |
| R2-S3 / R2-W3 | R2-S2 | Preserve: immutable bundles. Correct: error/accounting closure through one private transaction store | Bundle/session/report storage, secure file port, schema semantics | Releasing budget while failed staging remains; aliasing source and derived bytes | Existing digest/limit fixtures; failure before/after commit; rollback error paired with primary; source variants | Browser row; `make json-shape-check`; `make lint-scripts` | Keep prior published bytes read-only; remove only proven owned staging | Callers cannot mutate store maps/counters; no partial ref returned; budget matches owned surviving files; primary and cleanup errors survive. |
| R2-S4 / R2-W4 | R2-S3 | Preserve exact algorithms; change only adopted execution budgets. Isolate native decode/encode and heavy report work; reserve before allocation and publication | Worker/PNG/algorithms and their capture/store/report callers; host admission port | CPU/memory amplification, uncancelable codec, quiet-measurement interference | Maximal valid inputs, one-over limits, injected stalled work, cancellation at each stage, event-loop/control responsiveness | Browser row; `make service-backed-test-slice OWNER=harness.browser` where scheduler/resource owners require it | Remove new execution path as one owner; no parallel legacy worker remains | One shared operation deadline is enforced across stages; child death is drained; measured bounds and quiet admission meet R2-W1 contract. |
| R2-S5 / R2-W5 | R2-S4 | Preserve: strict targeting, geometry, provenance, current-only imports. Split browser observation and source normalization from coordinators | Browser/capture/axe/source/import/analysis; narrow test fixtures | Changing coordinates, accessible observations, attachment association or renderer claims | Stable/moving geometry; iframe/axe outcomes; duplicate results; multiple captures; expected-only; two source revisions; pin drift | Browser row; command slice if input semantics change; selected canonical owner rows discovered with task guide | Move one adapter at a time; remove redundant forwarding exports after callers switch | Analysis/report no longer know Playwright report layout; capture does not operate through unrestricted session internals; expected bytes/metadata stay immutable. |
| R2-S6 / R2-W6 | R2-S5 | Preserve report meaning; separate rendering/read model from store. Correct guide/skill component resolution and recovery instructions | Report, bundle model; existing guide and skill only where needed | Lost escaping, inaccessible controls, inferred observations, stale links | No-network/HTML injection, keyboard zoom/slider, correct original dimensions, report cache/digest, agent reads manifest then image | Browser row; `make lint-markdown`; manual local report/image trial | Retain report semantics, not incidental markup; no product golden refresh | Human and agent inspect the same source bundle while live; inaccessible/absent evidence explicitly reported; all private links expire. |
| R2-S7 / R2-W7 | R2-S1–S6 | Complete routing, supported-platform qualification, removal of obsolete internals and handoff | Test families/task owner/generated outputs; guide; tracker | Calling unit coverage end-to-end proof; accepting platform assumptions | Complete matrix in §18.8, seeded/dev/artifacts, public failure/interrupt and cleanup workflows | `make agent-finalize`, then affected owner/drift/contract checks and public workflows below | Revert authored changes and regenerate only after owned sessions stop; adopted public retirement needs amendment | Every required acceptance row has current evidence; zero unresolved required gaps; no private residue/unowned process claim; final handoff complete. |

Per-workstream risks and exits in these tables are gates, not optional aspirations.
After each completion and **before starting the next workstream**, update §18.9
and append §18.10 evidence: changed/inspected files, owner/contract decisions,
exact commands and run roots, normalized failures with relatedness, skipped
checks and reasons, resource/retention disposition, residual risks and next step.
Use only `TODO`, `IN_PROGRESS`, `BLOCKED`, `DONE`, `DEFERRED`, and `DROPPED`.

### 18.8 Validation plan

Initial discovery ran through public Make help, task-guide and explain-target
commands; that planning step ran no runtime baseline. R2-W7 retired the coarse
browser row and adopted separate contract, lifecycle, artifact, execution,
presentation and workflow rows. Current examples are:

```bash
make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.ui_review_contract,harness.browser.boundary_support.ui_review_artifacts
make service-backed-test-slice OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_default,harness.browser.integration.ui_review_seeded_network_flow_claimed
```

Use `make explain-test-owner OWNER=harness.browser` for the complete current
semantic row inventory. The old row has no compatibility alias. Historical
checkpoints retain the commands actually executed. Harness mechanics remain separate from review
commands: running a helper never becomes product-row evidence.

| Validation layer | Discovered command | Scope | Required before implementation? | Notes |
| --- | --- | --- | --- | --- |
| Navigation | `make help-all`; `make task-guide ROLE=module-author OWNER=harness.browser`; corresponding command-surface/generated-artifacts guides | Current public commands and owners | Yes; done for planning | Recheck if baseline changes. |
| Unit/contract | `make test-slice OWNER=harness.command_surface`; narrow browser row above | Inputs, JSON variants, output, algorithms, ownership | Capture baseline in R2-W1 | Existing browser row contains real-browser tests; it is not browser-free unit coverage. |
| Integration/resource | `make service-backed-test-slice OWNER=harness.browser` | Shared fixture and cleanup seams | Baseline before shared-owner change | Broaden only when those owners change or failure remains unresolved. |
| Public compatibility | `make browser-design-review-smoke`; `make browser-design-review-smoke REVIEW_PROFILE=default` | Both interactive preparation profiles | Before and after lifecycle change | Existing helper-only interface remains useful. |
| Public workflow | `make ui-review`, finite review commands and explicit dev/artifacts starts | Seeded editor/viewer, both profiles; borrowed dev; exact imports/reimports; report and stop | After relevant slices | R2-W7 routes W and both seeded integration rows through authored Make commands; no raw-script workaround. |
| Affected canonical rows | `make test-slice OWNER=<discovered-owner> ROWS=<selected-row-ids>` or service-backed equivalent | Current visual/a11y/browser rows touched by shared lifecycle/import changes | Select before those edits | Use owner guidance; no broad golden refresh, no carry-forward of historical row hashes. |
| Shape and generated contracts | `make json-shape-check`; `make generated-artifact-policy-check`; `make generate`; `make generate-drift`; `make toolchain-drift` | Authored schema/task/routing changes and downstream outputs | Establish relevant baseline in R2-W1 | Generate only in implementation; never hand-edit generated outputs/lockfiles. |
| Static | `make lint-scripts`; `make lint-shell` for shell edits | Changed executable sources | After affected edits | Existing owner mechanisms, no new phase-name scanner. |
| Documentation | `make lint-markdown`; `git diff --check`; human link/table review | Configured Markdown plus this tracker manually | This planning step | Current lint globs omit this tracker; a passing lint run is not direct tracker coverage. |
| Final owner closure | `make agent-finalize`, then browser/command-surface/generated-artifacts owner slices and `make harness-contract` | End-of-implementation integration | R2-W7 | Supply `RESULTS_DIR` only for an eligible full warm run; otherwise record retained-run maintenance skipped. |
| Full repository checks | `make check`, CI/release targets as separately warranted | Only changed ownership or unresolved risk justifies broader scope | No | This iteration does not assert product release readiness from helper tests. |

Required acceptance supplements below are human planning IDs, not machine inputs.
Map them to owner clauses and semantic test claims in R2-W1 alongside all still
applicable AC-102–AC-129; do not replace the adopted matrix with this table.

| ID | Required evidence | Completion rule |
| --- | --- | --- |
| R2-A01 | Control with damaged/missing analysis dependencies; malformed and cross-user locators | Valid exact control works under the adopted minimal profile; unsafe identities fail before side effects; work commands still fail readiness. |
| R2-A02 | Acquire/prepare/action/stop interleavings, two sessions, lock contention, action timeout | No duplicated effects, no late admission or successful publication after terminal cancellation, correct epochs, borrowed isolation. |
| R2-A03 | Controller death, parent death, both lost, PID reuse, hung release | Existing resource owners recover only proven owned resources; all releases bounded, failed proof retained, no unrelated process/service touched. |
| R2-A04 | Cleanup and terminal publication fail independently and together | Primary/secondary diagnostics survive; no false successful terminal evidence; safely removable detail is purged; receipt digest immutable. |
| R2-A05 | Artifact transaction faults, symlink/ownership races, changed bytes, each storage bound | No partial publication, overwritten parent/source, budget leakage, unsafe access or hidden rollback failure. |
| R2-A06 | Native-worker stall, maximal inputs, abort during decode/encode/report, quiet waiter | Control remains responsive; shared deadline/admission respected; children reaped; measured peak resources satisfy adopted limits. |
| R2-A07 | Every source/channel combination and current producer join | Impossible claims rejected; failed/expected-only valid cases stay diagnostic; multiple captures cannot cross-associate. |
| R2-A08 | Browser observation, geometry, fonts, iframe/axe states, exact raster fixtures | Existing algorithms and limitations preserved; missing coverage is never labeled successful assessment. |
| R2-A09 | Retained stdout/argv/errors/telemetry/private-path sentinels and transient outputs | Only structural receipts persist; no private detail retained on success, interruption, or injected failure. |
| R2-A10 | Public seeded editor/viewer + both profiles, dev service survival, browser-free artifacts | Reproducible Make-owned tests pass and cleanup is proved; product/golden/release selections remain unchanged. |
| R2-A11 | Real image/report consumption using guide/skill; malicious observed content | Exact manifest components viewed, offline controls usable, no network/active observed URL, expiry observed, no retained private copy. |
| R2-A12 | Platform qualification and final repository closure | WSL2 Linux x64 and separately supported native Linux amd64 have exact dependency/workflow evidence, or support is explicitly narrowed by an adopted amendment before claiming readiness. |

### 18.9 Top-level work tracker

| ID | Work item | Workstream | Status | Depends on | Evidence or artifact | Exit condition |
| --- | --- | --- | --- | --- | --- | --- |
| R2-T00 | Inventory, owner review, findings and next-iteration plan | Planning | DONE | Current document request | §18.1–18.8; source/command inspection in §18.10 | Only tracker changed; current evidence distinguished from history. |
| R2-T01 | Adopt decisions and characterize contracts | R2-W1 | DONE | Authorized implementation request | Amendment ui-review-api-2, typed work profiles, acceptance map and W1 checkpoint | Specification and valid-variant/version decisions closed. |
| R2-T02 | Session control and recovery | R2-W2 | DONE | R2-T01 | W2 checkpoint; lifecycle/death matrix, shared contracts, both smoke profiles and direct seeded trial | Exact, bounded recovery and honest terminal outcome. |
| R2-T03 | Private artifact store | R2-W3 | DONE | R2-T02 | W3 checkpoint; transaction faults, residual accounting, immutable provenance and source matrix | Transaction, reservation, provenance and rollback closure. |
| R2-T04 | Bounded work execution | R2-W4 | DONE | R2-T03 | W4 checkpoint; weighted tree admission, worker cancellation, measured WSL2 envelopes | Stress/cancellation/admission criteria satisfied. |
| R2-T05 | Observation and source adapters | R2-W5 | DONE | R2-T04 | W5 checkpoint; scoped browser observation, normalized canonical comparison and current producer fixtures | Narrow adapters preserve exact capture/source meaning. |
| R2-T06 | Report and agent/human consumption | R2-W6 | DONE | R2-T05 | W6 checkpoint; user accepted regenerated review, exact image inspection, keyboard/offline tests and verified expiry | Local inspection succeeds before verified expiry. |
| R2-T07 | Validation and handoff completion | R2-W7 | DONE | R2-T01–R2-T06 | W7 completion checkpoint; all semantic rows, both profiles, owner suites, harness contract and full check 977/977; verified cleanup | Required WSL2 evidence complete; native qualification user-deferred. |
| R2-T08 | Optional engines, hosted/export workflows and plugin framework | Outside iteration | DEFERRED | Separate demonstrated need and adoption | §18.1 exclusions | Not required for this iteration's completion. |

### 18.10 Session handoff log

The original records describe the 2026-09-28 planning session. Implementation
checkpoints appended below distinguish current execution from that history.

#### R2-W1 execution start — 2026-09-28

- User authorized the complete seven-workstream plan; R2-W1 is IN_PROGRESS.
- Baseline: `main`, `25806426bae985528ee14de9e2595316da93ba8f`; only this tracker
  had pre-existing edits (414 additions, 10 deletions), preserved in place.
- Host: WSL2 Linux x86_64, kernel `6.6.114.1-microsoft-standard-WSL2`.
- User subsequently directed: **skip native Linux amd64 qualification in this
  iteration**. WSL2 qualification remains mandatory. Native qualification is
  DEFERRED by that explicit scope correction; existing declared platform support
  is not rewritten and this iteration makes no new native-Linux readiness claim.
  R2-A12 and W4/W7 exits are assessed against this authorized WSL2-only scope.
- Adopt bounded parallel work through the existing host arbiter: raster jobs
  reserve 1 CPU/1 process/1024 MiB; report jobs 1 CPU/1 process/512 MiB. These
  are admission envelopes pending measured qualification, not measured peaks.
- Preserve public V1 review envelopes, current-only imports, private lifetimes,
  and existing input bounds. Retire invalid test fixtures and private internals.
- Additional F02 evidence: the shared suite-runtime stale janitor currently
  removes aged runtime trees without process-liveness/recovery checks. W2 must
  protect live ownership and unresolved proof before shared layout changes.
- Commands: `git status --short`, `git log -1 --format='%H %s'`, `uname -a`,
  targeted source reads. No runtime validation yet; no resources acquired.
- Next: amend owner contracts and typed policy, establish fresh baselines, then
  record the W1 exit before starting W2.


#### Scope and authority

#### R2-W1 adoption and validation map — current

Adopted amendment: `ui-review-api-2`. Public review V1 wire shapes and identities
remain current; invalid semantic variants are rejected rather than translated.
Machine work envelopes are authored topology profiles, read through the private
review policy module; no generator or runtime consumes this document. W2 owns
minimal-control/cleanup implementation, W3 owns semantic storage enforcement,
and W4 owns worker/admission enforcement. Limits are unchanged apart from the
explicit newly adopted aggregate worker reservations.

| Acceptance / findings | Implementation owner | Current baseline and required extension |
| --- | --- | --- |
| AC-102–107, 125–128; A01/A09; F03/F07 | Command surface, private control and contracts | C baseline; add engine-damage control and semantic source cases; retain output/input/privacy rejection coverage. |
| AC-108–111, 127; A02–04; F01–03 | Session coordinator, preparation, runtime recovery | L baseline; add prepare/stop races, both-supervisor death, publication/cleanup faults, live aged runtime and unresolved proof. |
| AC-114–118; A05/A07; F04/F07 | Private store and source adapter | B baseline; add transactional fault stages, failed-rollback accounting and all source/channel variants. |
| AC-108/111/118–120; A06; F05 | Worker executor and existing host arbiter | A/L baseline; add weighted parallel mixed jobs, shared deadline, large input, native stall and process-tree peak qualification. |
| AC-109/112–116/119–122; A07–08; F06/F07 | Browser/canonical/axe adapters | B/A/L baseline; add producer associations and preserve exact raster/observation semantics. |
| AC-123–126; A11; F09 | Presenter and private evidence consumers | A/W baseline; add actual component resolution, keyboard use and real local inspection before expiry. |
| AC-129; A10/A12; F08 | Routed public workflows and final handoff | W currently covers dev/artifacts; add both seeded profiles with editor/viewer. WSL2 required; native host qualification explicitly deferred by user. |

Fresh pre-change baselines passed: browser slice run
`.cartulary/test-results/20260928T230116Z-p2153` (1/1 units, 60703 ms),
command-surface slice `.cartulary/test-results/20260928T230116Z-p2166`
(1/1 units, 28487 ms). These are baseline evidence, not proof of new behavior.

#### R2-W1 completion checkpoint — 2026-09-28

- Owner amendment and updated AC-103/108/111 adopted; source/component truth
  table, control-only readiness, recovery/detail boundaries and parallel resource
  policy specified. No owner contradiction found. Review schemas/commands remain V1.
- Changed: NLSpec; authored topology resource profiles; catalog profile roster;
  family/work-graph-owner schema enums; private `policy.mjs`; contract tests;
  generated topology render index via Make; this tracker. No product or golden edit.
- Profiles also declare 1 IO token, as required by the existing executable-profile
  contract. CPU/process/memory claims remain the user-approved values.
- `make generate` PASS: `.cartulary/test-results/20260928T230631Z-p19114`.
  Earlier attempts `20260928T230442Z-p12339` and `20260928T230528Z-p15817`
  failed at topology generation due to ordering and the closed profile roster;
  both were related authored-projection omissions, corrected before the passing run.
- `make test-slice OWNER=harness.command_surface` PASS:
  `.cartulary/test-results/20260928T230648Z-p22324`.
- `make json-shape-check` PASS: `20260928T230648Z-p22293` (3/3 units).
- `make generated-artifact-policy-check` PASS: `20260928T230835Z-p28351` (3/3).
- `make lint-markdown` PASS: `20260928T230649Z-p22556`,
  `adhoc/lint-markdown/tool-run-summary.json`; tracker itself reviewed directly
  because configured lint globs exclude it. `git diff --check` PASS.
- New lifecycle/store/worker behavior is specified, not yet implemented or
  claimed passing. Full repository/platform qualification and retained-run
  maintenance remain for W7; no eligible warm RESULTS_DIR supplied.
- No session or service acquired by this workstream beyond baseline test-owned
  disposable resources. No private artifact handed off. Next: R2-W2.

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-09-28 | Planning-only repository session | Iteration 1 historical; iteration 2 planned | AGENTS, framework, tracker format, relevant NLSpec/domain/design sections; only this tracker edited | `git status --short`; `git log -1 --format='%H %s'`; `git branch --show-current`; targeted `rg`, `sed`, `cat`, `wc` | Clean baseline on main; 22 target files inspected; current scope documented | No owner-to-owner contradiction identified | Start R2-W1 only under a later implementation request. |

#### Backend and resource boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-09-28 | Source inspection | No product backend edit | Session/controller/recovery/registry; shared preparation and runtime/admission/secure-IO seams | Targeted source reads and inbound import searches | Resource owners exist; session context and private cleanup/proof need separation | RB-002 | Characterize failure paths without inventing a new service owner. |

#### Frontend and presentation boundary

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-09-28 | Source inspection | Product UI unchanged | Browser/capture/axe/report/analysis; design scope; guide and skill entry point | Targeted reads | Offline presenter and evidence-consumer seam identified; no design rewrite proposed | None for planning | R2-W5/W6 preserve observations and inspect actual pixels. |

#### Contract and code generation

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-09-28 | Contract inspection | No projection/generation changes | Contract/failure modules, selected schema fields, eight attachments, task recipes, Make renderer, generated policy search | Targeted reads; `make explain-target TARGET=harness-ui-review-contract DETAIL=summary` | Seven public helpers and V1 identities retained; new readiness policy requires adoption | RB-001/RB-002 | R2-W1 resolves owner clauses before typed changes; generate only in implementation. |

#### Tests and harness

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-09-28 | Verification discovery | No runtime suites run | C/L/B/A/W and import fixture; browser/command routing; Markdown config | `make help-all`; task guides for `harness.browser`, `harness.command_surface`, `harness.generated_artifacts`; explanations for `harness-ui-review-contract`, `browser-design-review-smoke`, `lint-markdown` | All discovery commands passed; seeded automated route gap and coarse row identified | RB-003 | Establish fresh baseline and semantic routing in R2-W1; old roots remain historical only. |
| 2026-09-28 | Document validation | PASS within stated coverage | This tracker manually reviewed; configured Markdown lint selection | `make lint-markdown`; `git diff --check`; direct structure, table and link review | Both commands passed. Lint run `.cartulary/test-results/20260928T223217Z-p89243`, summary `adhoc/lint-markdown/tool-run-summary.json`. Lint configuration omits this tracker; its new section, navigation, tables and historical/current boundaries were reviewed directly. | No runtime verification implied | Planning complete; establish runtime evidence in R2-W1. |

#### Security and authorization

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-09-28 | Boundary inspection | No browser/service acquired | Secure IO, retained projections, stop/recovery, owner privacy clauses | Source reads only | No new credentials/private artifacts produced; failed-cleanup detail retention is a current structural risk | RB-002 | Plan proof/detail separation and tests; do not delete runtime evidence during planning. |

#### Open risks and next session

| Time | Agent/session | Current state | Files inspected or touched | Commands run | Result | Blockers | Next action |
| --- | --- | --- | --- | --- | --- | --- | --- |
| 2026-09-28 | Planning handoff | R2-T00 DONE; R2-T01–T07 TODO | Sole controlling tracker updated | Runtime/shape/generation/finalization suites intentionally not run: documentation-only scope; no eligible full warm `RESULTS_DIR` supplied | Document-only scope preserved; retained-run maintenance skipped; no new production-readiness claim | §18.11 adoption/evidence gates | Recheck HEAD/dirty state and owner guidance, mark R2-W1 IN_PROGRESS after implementation authorization; record each completed slice before the next. |

#### R2-W2 completion checkpoint — 2026-09-28

- Control loads pinned Node and contracts without image/browser engines. The
  controller loads lifecycle code only for a new session. Preparation uses an
  owned subprocess; exact provider handles and boot/start process identities
  are recorded at acquisition. Cancellation drains late acquisitions, reaps
  detached groups even after leader death, and prevents successful late actions.
- One terminal publisher now serves shutdown and recovery. Missing/invalid
  receipts are artifact failures. Registry deletion follows cleanup, receipt
  publication and the retained scan. Age cannot remove live or unresolved roots.
- Owner clarification in REQ-502: explicit stop may retry exact unresolved
  resources after a failed terminal controller dies; it preserves the original
  failed receipt byte-for-byte. This adds no command, schema, alias or reader.
- Private detail disposal is separate from ownership proof, including credentials,
  preparation history and frontend copies. Borrowed runtimes use the same secret
  registry implementation. Existing testservices cleanup now accepts its exact
  lease independently of discarded diagnostic history, without fabricating events.
- Changed: review CLI/toolchain/controller/session/registry/browser/recovery;
  new ownership, preparation and terminal modules; shared preparation, provider,
  secure IO and suite-runtime ports; testservices cleanup and its existing test;
  lifecycle/contract/shared-runtime tests; owner text and generated topology index.
- Final focused lifecycle plus service-owner rows PASS (2/2):
  `.cartulary/test-results/20260928T234601Z-p92805`.
  Tests include hung release, PID reuse, late acquisition, parent/controller/both
  deaths, real browser reaping, borrowed-origin survival, engine-blocked exact
  status/stop, publication faults and immutable terminal repetition.
- Command-surface slice PASS: `20260928T233056Z-p47321`.
  Service-backed browser slice PASS (15/15): `20260928T233236Z-p80446`.
  Shared `make harness-contract-tests` PASS: `20260928T234643Z-p2608`,
  `harness-contract-tests/harness-contract-tests/step-summary.json`.
- Both `make browser-design-review-smoke` profiles PASS, closed with private
  roots removed: `design-review-1790638292506-64a31823` (default) and
  `design-review-1790638370674-4713f5ff` (network_flow_claimed).
- Direct public seeded default trial `r2-w2-seeded-default-2` reached ready and
  explicit stop closed successfully with cleanup complete. Earlier
  `r2-w2-seeded-default` failed: borrowed runtime lacked secret-registration
  capabilities, then recovery incorrectly depended on deleted lifecycle history.
  Both defects were repaired. Exact stop subsequently removed that trial's
  resources, private root and registry while preserving its failed receipt.
  Accidental auxiliary recovery diagnostics in that trial's retained `_shared`
  directory were removed by exact named-file deletion; only UI receipts remain.
  Temporary private debugging instrumentation was removed, and its private file
  disappeared with the recovered root. No unresolved trial resource remains.
- Related failures: shared contracts `20260928T233236Z-p80951` detected stale
  generated topology inputs (fixed by Make regeneration); service-owner row
  `20260928T234209Z-p71577` found a new test incorrectly rejecting ENOENT for
  intentionally absent history (assertion corrected; rerun
  `20260928T234303Z-p80284` passed). Invalid owner discovery for
  `harness.scheduler` was corrected to existing harness targets.
- Final `make generate` PASS: `20260928T234559Z-p92081`; shape PASS:
  `20260928T234644Z-p3105`; scripts lint PASS: `20260928T234633Z-p1243`;
  shell lint PASS: `20260928T234210Z-p71977`; Markdown lint PASS:
  `20260928T234631Z-p99766`. `git diff --check` and direct tracker review PASS.
- Compatibility: restart private sessions after cutover; old unresolved proof
  is preserved, never interpreted as deletion authority. Public V1 outcomes remain
  current. No database, product UI, canonical golden or borrowed input changed.
  Remaining work is artifact transactions, worker/resource qualification,
  adapters, presentation and final routed acceptance. Native qualification is
  user-deferred. Finalize/full closure and warm retained-run maintenance remain W7.
- R2-W2 DONE. R2-W3 is now IN_PROGRESS; no W3 implementation preceded this exit.

#### R2-W3 completion checkpoint — 2026-09-29

- Added the session-scoped `artifact-store.mjs`; migrated capture, analysis,
  reports and operation observations to its reservations, immutable publication,
  verified lookup, cache and disposal. Removed producer/session storage maps.
  Secure IO now preserves initiating failures alongside rollback failures and
  returns the surviving quarantine path so residual bytes remain charged.
- Added `bundle-semantics.mjs` and enforced the adopted source/channel truth
  table, exact immutable parent provenance and observation/image consistency.
  Updated bundle, analysis and command-contract fixtures to use real canonical
  expected-only evidence. No wire version, dependency, product or golden change.
- Fault injection covers writes before/after materialization and commit, failed
  rollback, cancellation, shared report/observation accounting, tampering and
  byte/count limits. No partial reference escapes; failed deletion keeps its
  reservation until later verified disposal. Test-owned private roots are removed.
- PASS `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.private_ui_review_session_ownership_and_strict_b_b785350fa4`:
  `.cartulary/test-results/20260929T000307Z-p40910`.
- PASS `make test-slice OWNER=harness.command_surface`:
  `.cartulary/test-results/20260929T000602Z-p50661`.
- PASS `make generate`: `20260929T000415Z-p47641`; `make json-shape-check`:
  `20260929T000603Z-p50868`; `make lint-scripts`: `20260929T000306Z-p40624`;
  `make harness-contract-tests`: `20260929T000605Z-p51364` (all under
  `.cartulary/test-results/`). `git diff --check` and direct tracker review pass.
- Related failures: command run `20260929T000019Z-p24692` rejected fabricated
  canonical metadata; replaced it with the current producer fixture. Run
  `20260929T000220Z-p34381` passed those cases but concurrent generation changed
  its source fingerprint; generation was completed and the final run passed with
  edits frozen. Neither failure is waived.
- Deferred to ordered successors: worker isolation/resource qualification (W4),
  narrow observation adapters (W5), consumption (W6), final routing/drift/full
  closure (W7). Retained-run maintenance remains skipped: no eligible full warm
  `RESULTS_DIR`. Native platform qualification is user-deferred for this iteration.
- R2-W3 DONE. R2-W4 is now IN_PROGRESS; no W4 implementation preceded this exit.

#### R2-W4 completion checkpoint — 2026-09-29

- Replaced the analysis-only buffer IPC worker with `executor.mjs`: one admitted,
  cancellable process per data operation, one 30-second deadline, exact process
  registration before work, reaping before input disposal and store adoption only
  after successful exit and a final cancellation check. Raster validation,
  import, derivation, encoding, report construction and cache validation run there.
  Browser observation spools its screenshot for that worker; IPC carries metadata
  and private paths rather than image collections.
- Extended the existing host arbiter with CPU/process/memory/IO claims using the
  existing capability resolver. Scheduled host-activity participants propagate
  their exact parent token. Subtrees consume the maximum of inherited claims and
  summed children, preventing duplicate credit. Existing activity can drain while
  a quiet waiter blocks new activity; nested quiet work excludes other roots.
  Exact browser, preparation and worker proofs preserve claims after controller
  death. Corrupt proof is rejected rather than treated as dead ownership.
- Added store work reservations and immutable file payloads. Secure IO streams
  parent/report copies and rechecks source digests; loading materializes only
  selected files. Failed workers retain their full staging reservation until
  confirmed deletion. Derivation stops at the existing cumulative bundle limit.
  Shared pure fraction arithmetic also repairs the W3 semantic check to apply
  the specified rounding. No image algorithm, public command or V1 envelope changed.
- Changed areas: executor, store, payload/secure IO, capture, analysis, report,
  image math/algorithms, session/preparation/browser process guards, host admission,
  scheduler/runner integration, execution fixtures, authored task backing inputs
  and Make-generated projections. Removed `image-worker.mjs` and its callers.
- PASS final browser slice (the exact W3 command):
  `.cartulary/test-results/20260929T003319Z-p77910` (63,783 ms). Includes maximum
  32-MiB PNG/16,777,216-pixel inputs, sixteen full-size crops, mixed parallel
  raster/report work, stalled workers, cancellation, resource exhaustion,
  inherited sibling claims, quiet order, corrupt proof and dead-parent/live-worker
  retention. Final measured process-tree peaks: raster **474,406,912 bytes** of
  1,073,741,824; report **416,043,008 bytes** of 536,870,912. Node 24.15.0,
  Linux x64, WSL2 kernel 6.6.114.1. Envelopes were not raised.
- PASS `make harness-contract-tests`: `20260929T003319Z-p78482`;
  `make service-backed-test-slice OWNER=harness.browser`: 15/15,
  `20260929T002543Z-p10393`; `make lint-scripts`: `20260929T003319Z-p78113`;
  `make json-shape-check`: `20260929T002829Z-p49469`; final `make generate`:
  `20260929T003258Z-p74749` (roots under `.cartulary/test-results/`).
  Shared contracts were rerun after final admission changes. `git diff --check`
  and direct tracker review pass.
- Related generation failure `20260929T001725Z-p61929` referenced the removed
  worker in authored backing inputs; replacing that registration and adding the
  new owners repaired it. No runtime failure was waived. Intermediate passing
  runs remain historical; the final peak values above belong to the final code.
- Cutover inspection found the host registry empty and no active review process
  before changing its private shape. No legacy lease was translated or deleted.
  Final inspection again shows zero leases and no review worker/controller.
  Borrowed sources survive; task-owned staging and test roots were disposed.
- Remaining work is W5–W7. Native Linux qualification is explicitly user-deferred;
  no new native readiness claim. Full closure/drift and retained-run maintenance
  remain W7; no eligible full warm `RESULTS_DIR` is currently available.
- R2-W4 DONE. R2-W5 is now IN_PROGRESS; no W5 implementation preceded this exit.

#### R2-W5 completion checkpoint — 2026-09-29

- Browser capture now goes through its observation adapter with scoped driver
  access, immutable observations and an explicit source-attestation capability.
  Capture/analysis/report operations receive identity, store and staged-input
  ports; they no longer receive a mutable session or unrestricted browser page.
  Preserved exact targeting/epochs, focus/scroll, font readiness, three settled
  frames, immediate post-screenshot geometry checks, native scale and RGBA math.
- Canonical producer joins and comparison identities now belong to the canonical
  adapter. The store exposes an immutable normalized comparison identity; analysis
  compares that identity without interpreting producer report/fixture formats.
  Missing scope remains unavailable for matched comparison while valid expected-only
  evidence remains importable and reportable. Source bytes and revisions stay exact.
- Centralized pin-derived engine facts; work-policy loading is separate so control
  does not depend on worker configuration. Added shared immutable and bundle-model
  helpers. Removed the old analysis primary/comparison exports, capture forwarding,
  duplicate repository-root ownership and unused public axe-state helper export.
- Changed: browser and new browser-observation adapter; source/canonical-import;
  capture, analysis, report, store and bundle model; axe/contract/semantic facts;
  policy/work-policy, immutable helper; affected narrow imports, producer fixtures,
  authored backing scripts and generated projections. No product, schema version,
  image algorithm, canonical driver or golden changes.
- PASS `make test-slice OWNER=harness.browser ROWS=harness.browser.boundary_support.private_ui_review_session_ownership_and_strict_b_b785350fa4`:
  `.cartulary/test-results/20260929T004426Z-p14409` (63,361 ms). Extended cases
  cover multiple real fixture captures in one failed producer result, duplicate
  result joins, expected-only sources, changed source revisions, incompatible
  comparison identities and immutable normalized output. Existing capture/axe,
  exact raster, privacy, cancellation and resource-envelope regressions also pass.
- PASS selected canonical visual-support rows via
  `make service-backed-test-slice OWNER=harness.browser` with `ROWS` equal to
  `harness.browser.boundary_support.visual_anchor_geometry`,
  `harness.browser.boundary_support.visual_asset_failure`,
  `harness.browser.boundary_support.visual_preference_isolation`,
  `harness.browser.boundary_support.visual_preference_lifetime`, and
  `harness.browser.boundary_support.visual_presentation_profiles`:
  `.cartulary/test-results/20260929T004426Z-p14437` (11/11 units).
- PASS command surface `20260929T004013Z-p94621`; final generation
  `20260929T004354Z-p11204`; script lint `20260929T004426Z-p14612` (roots under
  `.cartulary/test-results/`). Owner guidance rediscovered through Make.
  `git diff --check`, removed-interface search and direct tracker review pass.
- No runtime acceptance failure or skipped required W5 check. Test-owned processes,
  leases and private roots were cleaned by their owners; no private evidence copied
  into this tracker. W6 owns actual image/report consumption and W7 owns remaining
  routing and full closure. Native qualification remains user-deferred; retained-run
  maintenance remains skipped without an eligible full warm `RESULTS_DIR`.
- R2-W5 DONE. R2-W6 is now IN_PROGRESS; no W6 implementation preceded this exit.

#### R2-W6 implementation and validation — 2026-09-29

- Presentation now consumes an immutable store read model with selected,
  digest-checked component capabilities. It receives no mutable store map or
  storage root. Primary image selection belongs to the shared bundle model.
- Fixed a report resolution defect: the copied manifest previously named
  components that the report renamed or omitted. Reports now preserve exact
  component filenames and bytes, including observations/provenance, and reuse
  the same assets across views. Cache version is 2. Local inspection links have
  no download workflow; escaping, CSP, native resolution and missing-channel
  labels remain in place.
- Changed store/report, analysis/report tests, the existing browser-readiness
  guide, and the review skill plus its artifact/live/recovery references. The
  documentation explains exact manifest lookup, digest/size verification,
  inspection before stop, minimal control readiness, dead-owner recovery and
  retained proof. No product UI, canonical golden or public V1 schema changed.
- PASS generation `20260929T005030Z-p56713`; browser slice
  `20260929T005051Z-p60166`; script lint `20260929T005051Z-p60335`; Markdown
  lint `20260929T005051Z-p60378`. Roots are under `.cartulary/test-results/`.
  The configured Markdown glob omits this tracker; direct table/link/diff review
  is separate. Tests cover manifest resolution, forged read-model references,
  escaping, no HTTP(S) requests, keyboard zoom/reveal, cache integrity and expiry.
- Corrected the W4 stress fixture during this audit: sixteen identical crop
  rectangles bypassed the public unique-items constraint. The fixture now uses
  sixteen distinct large crops and passes through the public request parser.
  The W4 measurements remain historical worker stress evidence, superseded for
  public-request qualification by the fresh valid-input run below.
- PASS corrected browser slice `20260929T005752Z-p79466` (63,114 ms), generation
  `20260929T005737Z-p76214`, and script lint `20260929T005753Z-p79746`.
  Measured WSL2 process-tree peaks: raster **571,531,264 bytes** of 1,073,741,824;
  report **351,223,808 bytes** of 536,870,912. Neither envelope was raised.
  A subsequent keyboard test also requires reaching controls through Tab rather
  than direct programmatic focus. Final generation PASS:
  `20260929T005954Z-p87962`; browser slice PASS: `20260929T010013Z-p91275`
  (63,076 ms); script lint PASS: `20260929T010015Z-p91548`. Private successful
  test logs were disposed by their owner; the earlier recorded peak trial uses
  the same executor and valid stress request. No additional runtime failure.
- Live public trial: `CARTULARY_TEST_RUN_ID=r2-w6-consumption make ui-review
  REVIEW_PROFILE=default`; exact status, editor authentication, page capture and
  report commands succeeded through Make. Agent inspected the manifest-selected
  1440×900 original in bundle-2, showing the editor incident directory. Source
  and report image bytes match SHA-256
  `62cd3e8082838c18b292fc84a520ba1f88f73025d0e40866e67ef4373e539c89`.
  The app returned queued for the local report tab; this is not proof of human
  inspection. An explicit human trial question is pending. The session remains
  live for that inspection; private paths and image content are not copied here.
- W6 is IN_PROGRESS. Human inspection and subsequent exact stop/expiry/structural
  retention checks remain required. W7 has not begun. Native Linux qualification
  remains user-deferred; final closure and warm retained-run maintenance remain
  W7, with no eligible full warm `RESULTS_DIR` supplied.

#### R2-W6 report regeneration — 2026-09-29

- User requested a regenerated Private UI review. Exact Make stop closed the
  original `r2-w6-consumption` session with cleanup complete; the foreground
  process exited zero and its private root, source image and report disappeared.
  Repeated stop returned the identical terminal SHA-256
  `a7a90dce2b63a7af76f61fb5a8367b3d4d94da9983f8382bd165487be668ad27`.
  The retained run contains only the locator and structural operation/terminal
  receipts. Human inspection of that original report was not confirmed.
- Started `CARTULARY_TEST_RUN_ID=r2-w6-consumption-refresh make ui-review
  REVIEW_PROFILE=default`, authenticated editor, captured the page with axe and
  rendered bundle-2 using the exact public Make commands. All succeeded. This
  creates a fresh sealed capture and report without overwriting an immutable
  bundle or circumventing the report cache.
- Agent inspected the fresh 1440×900 original. Its 53,946 bytes and SHA-256
  `ae7d0d0b41cb8c3fc675211dd86b0ea871db9e71f95c3907d94cb38eae8fd96f`
  match the manifest and report image. The app queued the new local HTML tab.
  The fresh session stays live for human consumption; caller-owned request
  scratch remains in use and will be removed after the final stop.
- W6 remains IN_PROGRESS pending human inspection and final session cleanup;
  W7 remains TODO. No implementation or dependency changed during regeneration.
  No new runtime failure; `git diff --check` passes. The prior W6 validation
  remains applicable, and native qualification remains explicitly deferred.

#### R2-W6 completion checkpoint — 2026-09-29

- User accepted the regenerated live review with “Great close WS 6”. This
  resolves the pending human-consumption gate for the same manifest-selected
  source image inspected by the agent. Automated Tab/arrow/Home/End, escaping,
  offline, cache and source-resolution evidence is recorded above; no additional
  manual keyboard result is invented from that acceptance.
- Exact public Make stop closed `r2-w6-consumption-refresh`; the foreground
  process exited zero, cleanup is complete, and repeated stop returned identical
  terminal SHA-256
  `ef9321e187921871af0e335ec9c16565941f28f2d3b244e5df2802e5be89e44a`.
  The complete private root and exact registry record are absent. Image/report
  links have expired. Only the locator and structural receipts remain under
  `.cartulary/test-results/r2-w6-consumption-refresh`.
- Removed the two exact caller-owned request files and their now-empty scratch
  directory. No borrowed resource was acquired or removed by this seeded trial.
  No cleanup failure or unresolved proof remains from either consumption session.
- Changed files, specification decisions, compatibility and validation results
  are in the preceding W6 records. No product/golden/dependency change; no durable
  private export. Native Linux remains user-deferred. Final owner suites, routing,
  full repository closure and handoff belong to W7. Retained-run maintenance has
  no eligible full warm `RESULTS_DIR` yet.
- R2-W6 DONE. R2-W7 is now IN_PROGRESS; no W7 implementation preceded this exit.

#### R2-W7 routing qualification — 2026-09-29

- Added six semantic shell rows: `harness.browser.boundary_support.ui_review_`
  followed by `contract`, `lifecycle`, `artifacts`, `execution`, `presentation`,
  or `workflow`. Separate verification identities own these responsibilities.
  Retired the old coarse row without an alias; its existing internal contract
  target now runs only the semantic request/source contract tests.
- Added `harness.browser.integration.ui_review_seeded_default` and
  `harness.browser.integration.ui_review_seeded_network_flow_claimed`, each with
  explicit profile, service dependencies and resource claims. Their public Make
  workflow owns preparation through the existing review coordinator, so the
  row does not acquire a redundant fixture lease. Both authenticate editor and
  viewer, then navigate, snapshot, capture with axe, derive and report.
- Extended the public artifacts workflow to import and reimport exact current
  producer-shaped failed/expected-only canonical evidence and report it without
  inventing actual pixels. Borrowed goldens are checked unchanged. Every workflow
  checks immutable parents, repeated terminal results, private-link expiry,
  borrowed dev survival and structural-only retained output.
- Changed authored task surface, browser test family, verification projection,
  public workflow fixture, seeded wrapper and pin-derived producer fixture;
  downstream task/topology outputs were generated through Make.
- PASS `make generate`: `20260929T012231Z-p13886`; six-row `make test-slice
  OWNER=harness.browser ROWS=<the six semantic rows above>`: 6/6,
  `20260929T012254Z-p17220`; two-profile `make service-backed-test-slice
  OWNER=harness.browser ROWS=<the two integration rows above>`: 4/4,
  `20260929T012254Z-p17257`; `make lint-scripts`: `20260929T012254Z-p17431`.
  All roots are under `.cartulary/test-results/`. No runtime failure; all workflow
  cleanup assertions pass. Owner guidance and full-check plan were rediscovered.
- R2-W7 remains IN_PROGRESS. Finalize runs before broader owner/contract/check
  validation. No eligible full warm `RESULTS_DIR` is supplied. Native Linux
  qualification remains user-deferred; no platform result is inferred.

#### R2-W7 broad-validation correction — 2026-09-29

- First finalize PASS: `20260929T012438Z-p62394`; command-surface PASS:
  `20260929T012553Z-p66981`; generated-artifact owner PASS: 5/5,
  `20260929T012553Z-p67021`; service-backed browser owner PASS: 17/17,
  `20260929T012553Z-p66951`. Retained-run maintenance was skipped because
  `RESULTS_DIR` was unset. These are fresh evidence, not iteration-1 carry-over.
- Broader browser run `20260929T012553Z-p66931` exposed a lifecycle assertion
  failure and a worker cancellation test failure under concurrent load. The
  latter left its interval alive after an assertion, preventing termination.
  Interrupted this exact owned run after observing the failure: 41 passed,
  1 failed, 5 cancelled. Its outer retained-boundary check then failed on an
  unpublished Playwright trace resource. This failed run is not waived or counted
  as passing. Its temporary `.playwright-artifacts-0` files were disposed after
  the runner stopped; structural failed/cancelled evidence remains intact.
- Corrected test ownership/routing: browser tests now inherit their scheduled
  host lease, rather than acquiring a competing root lease. Added authored
  qualification profiles reserving a browser plus raster envelope (1 CPU,
  1 process, 1536 MiB, 1 browser stack), and two parallel raster jobs (2 CPU,
  2 process, 2048 MiB). These describe test composition; production worker
  envelopes and five-second admission/30-second operation limits are unchanged.
- The stalled-worker test now always clears its interval, aborts/reaps pending
  work and disposes its store, including when an assertion fails. Its injected
  deadline allows five seconds for startup and stall detection; responsiveness
  requires a timer tick while work is pending rather than an arbitrary tick rate
  sensitive to other host activity. Remaining fixture engine identities now
  derive from the typed pins, with explicit invalid-version probes retained.
- Changed execution/browser fixture tests, authored resource/routing projections
  and their exact schema/catalog rosters. PASS regeneration:
  `20260929T013110Z-p99022`. Fresh narrow reruns and a real canonical auth capture
  are in progress before finalization and broader closure are repeated.
- No unresolved review registry record remained after the interrupted run;
  a pre-existing empty recovery-lock file was left untouched. No unrelated
  service or process was terminated. W7 remains IN_PROGRESS.

#### R2-W7 acceptance audit — current implementation

The following aliases identify exact current semantic rows under
`harness.browser.boundary_support`: C=`ui_review_contract`,
L=`ui_review_lifecycle`, B=`ui_review_artifacts`, E=`ui_review_execution`,
P=`ui_review_presentation`, W=`ui_review_workflow`. Their files are respectively
`test-ui-review-contract.mjs`, `test-ui-review-lifecycle.mjs`,
`test-ui-review-bundles.mjs`, `test-ui-review-execution.mjs`,
`test-ui-review-analysis.mjs` and `test-ui-review-workflow.mjs`, all under
`tools/harness/browser/tests/`. S denotes the two exact integration rows in the
routing checkpoint. No acceptance assertion is derived from reading Markdown.

Fresh corrected C/L/B/E/P/W evidence: 6/6 in
`.cartulary/test-results/20260929T013154Z-p2752`; script lint passed in
`20260929T013154Z-p2932`. Finalization after those corrections passed in
`20260929T013311Z-p44874`, including generated drift, shapes and catalog/tier
coverage. Broader final closure is recorded separately when it completes.

| Adopted acceptance | Exact verification responsibility and evidence |
| --- | --- |
| AC-102 | C public Make rejection/input cases; command-surface owner; authored helper-only commands remain outside product evidence. Both design-review smoke profiles passed in W2. |
| AC-103 | C core readiness and mode cases; L parent/both-supervisor tests reject engine loading while exact status/stop recover. |
| AC-104 | C mode/omission/cross-mode cases; S explicit default and claimed seeded workflows with both actors. |
| AC-105 | L exact actions, unavailable origin, redirects, busy stop; W explicitly borrowed origin survives teardown and stays live_unattested. |
| AC-106 | W/S artifact startup has absent browser path and unusable Docker endpoint; exact PNG/canonical import, analysis and report still pass. |
| AC-107 | C transient JSON/output/preflight; W/S validate every finite public result and receipt without retaining private refs. |
| AC-108 | L serialization and admission ordering; E inherited sibling claims, quiet waiters, exact worker proof, capacity limits and parallel raster/report work. |
| AC-109 | L exact handles/epochs, duplicate and disabled controls, stale/detached refs and literal shell-like input; no action replay. |
| AC-110 | L busy stop, signals, controller/parent/both loss, repeated receipts and exact lifetime expiry; W/S complete owned cleanup. |
| AC-111 | L late preparation, bounded/hung release, PID reuse, stale-runtime proof and terminal faults; B transactional rollback; E cancel/reap before input disposal. |
| AC-112 | B page capture case covers delayed fonts, real geometry deadline, post-screenshot mutation, dynamic text and stable frames. |
| AC-113 | L explicit rendered-row reveal; B focus/scroll preservation during native capture. |
| AC-114 | B exact canonical joins, unregistered/expected-only and multiple captures; W/S failed expected-only reimports. Fresh real passing canonical trial is recorded below. |
| AC-115 | B old/wrong report and reconciliation versions, source/digest/renderer divergence, duplicate associations and frozen provenance. |
| AC-116 | B native scope transforms, nonunit scales, outward rounding, clipping and no repeated CSS zoom. |
| AC-117 | B immutable publication, missing/mismatched components, every write/commit/rollback fault, retained residual accounting and secure path checks. |
| AC-118 | C request/semantic bounds; B PNG byte/pixel/dimension and bundle/file/count/storage limits; E maximal valid public request; P observation overflow. |
| AC-119 | P exact RGBA math, transparent RGB, integer rounding, unequal dimensions and immutable source bytes. |
| AC-120 | P crop/overlay/contact-sheet pixels and comparison identity; B changed source revisions and producer associations; no golden changes. |
| AC-121 | P axe findings, incomplete/empty, disabled and unavailable states; live iframe exclusion; C source/channel truth table. |
| AC-122 | P engine throw/timeout/malformed/oversize failures, with no successful partial result. |
| AC-123 | P real offline HTML execution, escaped injection/URLs, zero HTTP(S) requests, keyboard zoom/reveal, cache tampering and exact manifest resolution. |
| AC-124 | W6 user acceptance of the regenerated report and agent inspection of its exact source bytes; both public trial roots removed after stop. |
| AC-125 | W/S private text/URL/console/telemetry sentinels and retained-tree inspection; C closed output fields; L failure/interrupt cleanup. |
| AC-126 | C/L restricted input and secure path probes; existing source-boundary/command checks; typed contracts remain upstream of executable behavior. |
| AC-127 | L every normalized primary paired with secondary cleanup failure and immutable terminal repetition; C JSON result consistency. |
| AC-128 | Human owner/guide review; B rejects obsolete producer versions; pin-derived current fixtures and unchanged canonical goldens. |
| AC-129 | W/S all seven public commands, editor/viewer and both profiles, dev survival, PNG/canonical reimports and verified expiry. W6 supplies actual report consumption. |

| Supplement | Current disposition |
| --- | --- |
| R2-A01 | PASS — C/L exact minimal control, engine-damage and identity rejection coverage. |
| R2-A02 | PASS — L serialized transitions/actions and late-acquisition drain; E cancellation and admission. |
| R2-A03 | PASS — L controller/parent/both loss, PID reuse, bounded release and borrowed survival. |
| R2-A04 | PASS — L/B terminal and cleanup faults preserve primary/secondary outcomes and unresolved proof. |
| R2-A05 | PASS — B store faults, immutable provenance, secure paths and every storage boundary; P cache integrity. |
| R2-A06 | PASS — E schema-valid maximal workload, mixed workers, inherited claims, quiet ordering, stalls and reaping; W6 records measured WSL2 peaks within unchanged worker envelopes. |
| R2-A07 | PASS — C source/channel truth table; B producer-shaped multiple/duplicate/failed/expected-only joins; W/S public canonical reimports. |
| R2-A08 | PASS — B/P/L observation, geometry/fonts, exact raster and distinct advisory axe states. |
| R2-A09 | PASS — C/L/W/S privacy, transient output, failed/interrupted cleanup and structural-only receipts. |
| R2-A10 | PASS — both routed S profiles and actors, W borrowed dev and browser-free artifacts. |
| R2-A11 | PASS — W6 user/agent trial, source/report hash equality, keyboard/offline/injection checks and physical expiry. |
| R2-A12 | PASS — WSL2 runtime/dependency qualification, final full check 977/977 and cleanup inspection complete. Native Linux is DEFERRED by explicit user instruction; no native qualification is claimed. |

Fresh real canonical trial: `make service-backed-test-slice OWNER=module.auth
ROWS=module.auth.visual.capture_the_anonymous_auth_gateway_across_initia_755030aa99`
passed 11/11 in `.cartulary/test-results/20260929T013154Z-p2762`. Its exact
`auth-initial` capture `visual.capture.f9d688041051343ee0c9` was imported through
public Make into artifacts run `r2-w7-canonical`. Manifest, expected image and
report resolved correctly with `no_actual`, `no_axe`, `no_dom`, `no_trace`;
the agent inspected the digest-checked expected image. Exact stop and foreground
exit succeeded, cleanup is complete, private root/registry and caller scratch
are absent. Terminal SHA-256:
`d36a3d4aaf40d88bb40cee8ad38f009a2c83e7102212bf65a6bf7a409c748a1a`.
The borrowed canonical run and golden remain unchanged. No golden refresh.

#### R2-W7 cancellation race found during closure — 2026-09-29

- Full browser rerun `20260929T013350Z-p49394` completed 46/47; lifecycle and
  both seeded profiles passed. Execution failed only the cancellation case.
  Its retained unit diagnostic identifies `host admission unavailable` where
  `interrupted` was required: abort killed the worker during ownership binding,
  and the failed bind replaced the cancellation cause. The maximal workload
  itself passed, with raster 589,144,064 bytes and report 343,560,192 bytes within
  the unchanged envelopes. The run as a whole remains failed evidence.
- Fixed the executor to check cancellation before binding and give cancellation
  precedence when that binding fails after abort. A non-cancellation binding
  failure is normalized to resource conflict. Reaping and reservation disposal
  remain in the existing finally path. The regression now holds resource
  registration until an explicit abort, making this boundary deterministic.
- This is a production ordering correction, not a waived timing failure. The
  prior fixture/routing fixes remain useful: they removed leaked test intervals,
  repeated pin facts and duplicate browser admission. Fresh execution validation,
  finalization and broader closure follow this last implementation change.
- Additional completed checks before this correction: `make harness-contract`
  PASS `20260929T013350Z-p49701`; generated policy PASS
  `20260929T013350Z-p49318`; toolchain drift PASS `20260929T013350Z-p49412`;
  shell lint PASS `20260929T013350Z-p49905`; Markdown lint PASS
  `20260929T013350Z-p49950`. All roots are under `.cartulary/test-results/`.
  W7 remains IN_PROGRESS; no failed check is represented as a pass.

#### R2-W7 handoff facts

- Adopted owner remains `cartulary.testing_harness.v3`, amendment
  `ui-review-api-2`. The seven public Make commands and their existing V1 command
  IDs remain current: start, status, browser actions, capture, analyze, report
  and stop. The eight review V1 schemas remain current; impossible owner-invalid
  variants fail validation, while valid expected-only imports stay supported.
- Responsibility is now cohesive: the session coordinator owns admission,
  transitions, cancellation and terminal outcomes; exact ownership/recovery
  proof uses existing runtime/service owners; the artifact store owns private
  transactions and accounting; one bounded worker owns each admitted data
  operation; browser/canonical adapters normalize source observations and
  comparison identities; presentation consumes an immutable store read model.
  Secure IO, fixture preparation, host arbitration and pinned Playwright remain
  the shared owners. No alternate driver, scheduler, store or legacy reader.
- Production raster/report reservations remain respectively 1 CPU/1 process/
  1024 MiB and 1 CPU/1 process/512 MiB, each with 1 IO unit. Parallel work uses
  the same weighted host tree as scheduled activity and respects quiet waiters.
  Admission is bounded by five seconds and the remaining operation deadline;
  the single operation budget is 30 seconds. Qualification-only test composition
  profiles are separate authored claims, not changes to worker envelopes.
- Existing bounds remain: 64-KiB requests, 32-MiB PNGs, 8192 maximum dimension,
  16,777,216 pixels, 8-MiB structured components, 128-MiB/64-file bundles,
  100 bundles, 512-MiB private session storage, 32-MiB report HTML and eight-hour
  session lifetime. Observations remain advisory/private; cleanup expires images
  and reports, retaining only structural receipts plus genuinely unresolved
  minimum ownership proof when recovery still needs it.
- Fresh qualification is WSL2 Linux x86_64, kernel
  `6.6.114.1-microsoft-standard-WSL2`, Node 24.15.0, Playwright/core 1.59.1,
  Chromium 147.0.7727.15, Sharp 0.35.4, axe/core integration 4.13.0;
  installed native packages are `@img/sharp-linux-x64@0.35.4` and
  `@img/sharp-libvips-linux-x64@1.3.3`. Final file inspection and runtime suites
  establish these facts. Pin SHA-256:
  `0d113cb824430fab98503a36606045493c8c37adb61a020af3d9e39f11e9f23b`;
  lock SHA-256:
  `e1db2a82e1433557a34ec6509d802cc6adb5ad254993f5fde87a29698c50fb31`.
  Native Linux amd64 is explicitly user-deferred this iteration; no new native,
  Windows-native or macOS qualification is claimed.
- Compatibility: drain owned sessions before private registry/layout cutover and
  restart after deployment. Existing unresolved ownership is preserved rather
  than translated or erased. Public commands, epochs and immutable source meaning
  remain supported. The coarse internal verification row and obsolete worker/
  forwarding interfaces are retired without aliases. Rollback stops owned
  sessions, reverts authored changes and regenerates projections; it never
  deletes unresolved proof or alters borrowed inputs.
- Product behavior, database/schema migrations, workbook design and canonical
  goldens were not changed. Dependency manifests, pins and package-manager lock
  were not changed. The working tree remains uncommitted on the baseline revision
  named in §18.1, including the preserved original tracker edits and iteration-1
  history. Optional engines, hosted viewers, durable private exports and plugin
  frameworks remain outside this iteration.

#### R2-W7 final owner closure — 2026-09-29

- After the cancellation fix, generation PASS `20260929T013952Z-p22818`;
  exact E row PASS `20260929T014025Z-p26325`; script lint PASS
  `20260929T014025Z-p26442`; final `make agent-finalize` PASS
  `20260929T014113Z-p27873`. No retained-run maintenance was attempted without
  an eligible successful full warm `RESULTS_DIR`.
- Subsequent `make test-slice OWNER=harness.browser` PASS **47/47**:
  `20260929T014155Z-p32037`. This includes all semantic review rows, both seeded
  editor/viewer profiles, ordinary browser support and affected service owners.
  It supersedes the two failed/interrupted broader attempts without erasing them.
- `make harness-contract` PASS **2/2** `20260929T014156Z-p32281`;
  `make test-slice OWNER=harness.command_surface` PASS `20260929T014155Z-p32114`;
  `make test-slice OWNER=harness.generated_artifacts` PASS **5/5**
  `20260929T014155Z-p32159`. All roots are under `.cartulary/test-results/`.
- All C/L/B/E/P/W/S acceptance mappings above now have passing full-owner evidence
  after the last implementation change. `make check` is in progress as the
  required broader closure for shared lifecycle/scheduler changes. W7 and A12
  remain IN_PROGRESS until that result and final resource/retention inspection.

#### R2-W7 full-check startup failure and exact rerun — 2026-09-29

- First `make check` finished **976/977**, with only
  `harness.browser.integration.ui_review_seeded_default` failing during startup:
  `.cartulary/test-results/20260929T014432Z-p9140`. Its public terminal diagnostic
  is `startup_failed`; this did not publish a successful review. The claimed
  profile and every other unit passed. Private preparation detail was disposed
  of by the normal failure cleanup; the retained diagnostic does not establish
  a more specific cause, so none is inferred or waived.
- An unchanged-source exact rerun through `make service-backed-test-slice
  OWNER=harness.browser ROWS=harness.browser.integration.ui_review_seeded_default`
  passed **3/3** in `20260929T015704Z-p22122`, including both actors, artifact
  imports/reporting, terminal repetition and complete private cleanup. No review
  registry record remained afterwards. No timeout or assertion was weakened.
- The full check is repeated because its prior required exit failed. This is
  validation of the same finalized source; no new implementation change requires
  regeneration or another finalization. W7 and A12 remain IN_PROGRESS.

#### R2-W7 completion checkpoint — 2026-09-29

- `make check` PASS **977/977**, zero failed/skipped/cancelled units, 358,322 ms:
  `.cartulary/test-results/20260929T015839Z-p49175`. Every C/L/B/E/P/W row and
  both seeded editor/viewer integration profiles passed in this full run.
  The retained-secret scan passed across 4,841 files. This completes broader
  closure after the recorded finalization, owner suites and harness-contract
  passes; the earlier failed run remains failed historical evidence.
- Final inspection confirms zero UI-review registry records, zero sockets and
  zero host-admission leases. No review controller, preparation process or
  executor remains. Both W6 consumption roots, the W7 canonical trial, the
  failed full-check root, the exact rerun and the final full-check runtime roots
  are absent. Caller-owned W6/W7 request scratch is absent. Manual trial terminal
  receipts still match the hashes recorded above, report cleanup complete and
  retain only locators and operation/terminal receipts. The pre-existing empty
  recovery coordination lock was left untouched; it is not unresolved resource
  proof. Unrelated long-running processes and borrowed resources were untouched.
- Final authored change groups are the owner amendment, UI-review coordinator/
  recovery/store/executor/adapters/presentation, the existing shared secure-IO/
  preparation/admission/runtime seams, focused regression and workflow fixtures,
  authored verification/task/topology inputs and their Make-generated outputs,
  and the existing guide/skill/tracker. Exact files, commands and compatibility
  decisions are recorded in their owning workstream checkpoints. The obsolete
  image worker, coarse verification row and forwarding interfaces are removed.
  No temporary qualification fixture or private image/report was added to the
  authored tree. Source was unchanged between the exact rerun and final check.
- `git diff --check` and direct tracker table/link/checkpoint review pass. Prior
  script/shell/Markdown lint, generated policy/drift/shape and toolchain checks
  are recorded above; finalization and the full check cover the finalized source.
  The configured Markdown target does not select this tracker, so its final
  checkpoint was reviewed directly. No broad check was repeated merely to obtain
  retained-run maintenance evidence.
- Skipped scope: native Linux amd64 is explicitly deferred by the user for this
  iteration; Windows-native/macOS and CI/release publication are not claimed.
  Retained-run maintenance was skipped at finalization because no eligible
  successful full warm `RESULTS_DIR` was supplied. No golden update, dependency
  update, database migration, hosted viewer or durable private export occurred.
- Remaining limitations are the declared private lifetime/offline behavior,
  current-only pinned producer/engine support, and the deferred native platform.
  One earlier seeded startup failure could not be assigned a more specific cause
  from its structural-only receipt; unchanged-source exact and full reruns pass.
  It remains recorded as a diagnostic limitation, not a waived test or claimed
  fix. There is no unresolved recovery residue or required acceptance blocker.
- **R2-W7 DONE. R2-W1–R2-W7 are all DONE for the authorized WSL2-only iteration.**
  AC-102–AC-129 and R2-A01–A12 have the exact mapped passing evidence above;
  native-platform evidence is excluded only by the explicit user scope change.
  The adopted revision, dependency/resource policy, interfaces, cutover/rollback,
  compatibility and support limits in the handoff facts are the final handoff.
  Iteration-1 history and the original uncommitted tracker edits remain preserved.

### 18.11 Open questions and blockers

No unresolved owner question or required acceptance blocker remains. Final
validation and handoff completed in R2-W7; native qualification remains explicitly
user-deferred for this iteration.
Use `BLOCKED: owner contradiction` if subsequent inspection finds conflicting
adopted owners, and stop only dependent work until the owners are reconciled.

| ID | Question or blocker | Why it matters | Needed authority or evidence | Current status |
| --- | --- | --- | --- | --- |
| RB-001 | Adopt the control-only readiness profile | Control needs fewer dependencies than data production | R2-W1 amendment specifying pinned Node/contract validation and exact recovery tools, while engine-dependent commands remain fail-closed | DONE — ui-review-api-2 adopted in W1; engine-damage control trials pass in W2. |
| RB-002 | Close bounded-work and failed-cleanup proof rules | Adopted envelopes, exact proof retention, bounded cleanup and weighted execution | W1/W2/W4 checkpoints and measured WSL2 trials | DONE — declared bounds qualified on the required host; native trial user-deferred. |
| RB-003 | Establish repeatable seeded and platform evidence | Initial wrapper exercised dev/artifacts; old qualification only named WSL2 | Make-owned seeded routing and fresh WSL2 execution roots; native Linux explicitly deferred by the user for this iteration | DONE — both routed editor/viewer/profile workflows pass in the W7 routing checkpoint. |
| RB-004 | Recheck generated membership and shared-provider ports before moving them | Shared changes must follow their existing owners | Current generated policy and exact provider source at the changed seam during R2-W1/W2 | DONE — authored providers, runtime and testservices cleanup seams inspected; topology index regenerated through Make. |

No database migration or canonical golden migration is proposed. For internal
layout changes, stop owned sessions and re-create fresh ones after the cutover;
borrowed inputs/data stay untouched. If rollback is needed, stop owned resources,
revert authored changes for the slice, and regenerate downstream projections
through Make. Do not preserve obsolete internal exports or registry formats for
ephemeral sessions. Required public version changes are adopted deliberately,
with structural historical evidence retained under its existing lifetime policy.

### 18.12 Binary completion criteria

Planning is complete only when all target files are inventoried, findings have
owners and test posture, every slice has dependency/risk/rollback/exit criteria,
behavior changes are distinguished from structural moves, commands are discovered
or explicitly unresolved, history is preserved, and only this tracker is changed.
The framework's domain-module templates are intentionally adapted to a local
harness; no phase-shaped runtime module is introduced.

Implementation completion requires all of the following:

- R2-W1–R2-W7 are `DONE`, with a recorded checkpoint after each workstream and
  before its successor; every required acceptance blocker is resolved.
- Adopted amendments and typed projections agree; relevant existing AC-102–AC-129
  and R2-A01–A12 have exact current test/run or manual-inspection evidence.
- The seven commands and three modes work through public Make, including seeded
  editor/viewer and both profiles, explicit dev, browser-free artifact imports,
  failures, interrupts, and terminal repetition; existing design review survives.
- Recovery does not depend on unrelated analysis packages; no unbounded owned
  operation prevents stop; resource releases and proof preservation obey owners.
- Private originals and borrowed files are unchanged, publication is immutable,
  safe detail cleanup is demonstrated, and no raw diagnostic detail enters durable
  outputs. Any retained recovery residue is explained and prevents a clean-readiness
  claim until the required recovery is complete.
- Tests route by semantic ownership rather than implementation phase or class
  internals. No runtime/generator/test consumes Markdown as executable authority.
- Obsolete internal implementations, temporary fixtures and redundant exports
  are removed; no legacy reader, second driver, speculative plugin layer, or
  automatic conversion remains.
- The final handoff names the adopted revision, exact supported dependency and
  platform qualification, public interfaces and compatibility decisions,
  privacy/lifetime limits, complete acceptance disposition, commands and run
  roots, cleanup evidence, skipped checks, and optional work still deferred.

R2-W7 **Validation and handoff completion** is the final required slice. Unit
passes, historical handoffs, or a smaller diff cannot substitute for that exit.
