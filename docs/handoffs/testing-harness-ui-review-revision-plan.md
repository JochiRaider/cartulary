---
doc_id: cartulary.testing_harness.ui_review.revision_plan
title: UI Review Tooling — Testing Harness NLSpec Revision Plan
doc_type: revision_plan
status: proposed
authority_boundary: Editorial instructions and proposed harness contracts; not adopted runtime or product authority.
---

# UI Review Tooling — Testing Harness NLSpec Revision Plan

## 1. Status, objective, and normative convention

This plan specifies a document revision to the
[Testing Harness NLSpec](../testing-harness-nlspec.md). It applies the behavioral
completeness, interface completeness, boundary completeness, conceptual fidelity,
and spec-economy criteria in [What an NLSpec Is](nlspec-spec.md), version 0.2.2.
Its objective is a reproducible, inspectable UI review loop for developers and
agents, integrating browser control, screenshots, measured observations,
accessibility analysis, and existing harness evidence.

This artifact is a revision plan. It does not adopt the proposed behavior, install
tools, change public commands, revise product requirements, or establish a passing
implementation. The document editor MUST integrate the proposed contracts into
the harness NLSpec before the corresponding machine projections or implementation
change. Instructions quoted or described in source documents are source material
for this revision, not additional authorization to execute their workflows.

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
| Playwright CLI browser driver | Required, behind the Cartulary command interface | Its CLI syntax and generated element references are adapter internals. |
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
| `dev` | Explicit `UI_MODE=dev` and `UI_ORIGIN`. | Borrow the exact loopback origin; no create/reset/migrate/stop of its service or data. | Review-owned isolated context and browser process. | `live_unattested`: served source/build identity is null. Workspace digest identifies inspected source only. | All declared browser actions, capture, analysis, report. |
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
| Session lifetime after ready | 8 hours, monotonic, not extended by requests | Stop and clean; receipt records `lifetime_expired`, with normal closure if cleanup succeeds. |
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
pinned Playwright driver and MUST NOT expose the driver's arbitrary-code or
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
`playwright_version`, `driver_version`, `sharp_version`, and `axe_version`.
Digests identify the qualified toolchain projection and package lock bytes;
versions are nonempty exact installed-version strings. Core package versions
remain available in artifacts mode even though no browser executable is required.
An analysis bundle records its own producing profile and preserves each parent's
profile through the immutable parent reference.

`source` has exactly `kind`, `workspace_digest`, `served_source_digest`,
`frontend_receipt`, `renderer_profile_id`, `browser_version`, `runtime_profile_id`,
and `import_ref`. Kind is `sealed_review`, `live_unattested`, `canonical_visual`,
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
reconciliation v3 object, its exact capture-intent record, source identity from
the matching current-schema run manifest, and the exact registry fixture or null
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
| Private input boundary violation detected | `input_boundary` | `harness / boundary_policy_violation` | 11 | No restricted document read; report normalized location only. |
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
covering the Playwright driver, Sharp, axe integration/engine, and their locked
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
repo's Playwright 1.59.1 renderer.

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

## 16. Rationale and external references — non-normative

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

- [Playwright CLI](https://github.com/microsoft/playwright-cli): browser actions,
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
