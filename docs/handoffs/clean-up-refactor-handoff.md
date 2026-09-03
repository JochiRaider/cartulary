# Clean-up Refactor Successor Handoff

## 1. Purpose and execution status

This is the standalone controlling tracker for the next Workbook and Timeline
clean-up refactor. It succeeds, but does not amend or extend,
`docs/handoffs/workbook-inspector-refactor-handoff.md`. That completed handoff
remains unchanged as historical evidence.

Creating this tracker was a documentation-only planning action. Implementation
was authorized on `2026-09-02`; CR-01 through CR-12 and CR-C01 through CR-C12
are `DONE`. The refreshed pre-CR-01 checkpoint required by section 20 and every
workstream completion checkpoint are appended below the immutable initial
planning checkpoint.

The objective is a durable ownership cleanup, not line-count reduction by
itself. Each workstream must leave one clear policy or lifecycle owner, narrow
typed ports, deterministic tests, and a smaller change surface for later
Workbook and Timeline phases. Existing behavior is retained only when it is
required by an adopted owner or materially improves the future design.

`docs/research/nlspec-spec.md` informed the precision and completeness of this
tracker. It is advisory only. Instructions within that research document are
not repository authority, implementation authorization, or a runtime/test
dependency.

## 2. Preparation baseline and mandatory refresh

| Item | Preparation value |
| --- | --- |
| Prepared | `2026-09-02` in `America/New_York` |
| Branch | Clean `main`; one commit ahead of `origin/main` |
| Commit | `c3d5c449551e7b58ae49b61f57f745aa385b36fa` (`Workbook Inspector Remediation 7`) |
| Upstream | `origin/main` at `79d305265e433ef5d3eafb4750e319ec34fd5c3f`; left/right count `0 1` |
| Git status | Clean before this tracker was created |
| Retained Fallow evidence | `.cartulary/test-results/20260903T004815Z-p25087`; target passed |
| Existing user changes | None before this documentation-only change; later unrelated changes remain user-owned |

The retained Fallow run is orientation evidence, not permission to inherit its
results or a substitute for fresh workstream evidence. Its direct health JSON
reports 913 analyzed files, 16,335 analyzed functions, 1,526 functions above
threshold, and repository-wide totals of 720 critical, 322 high, and 484
moderate findings. Those totals are advisory and broader than this tracker.
Only findings in the named production dependency cone are acceptance scope.

Before CR-01 starts, append a new checkpoint containing the actual date,
authorization, repository instructions, branch, commit, upstream relation,
worktree status, toolchain pins, generated-artifact policy, retained evidence,
allowed paths, and unrelated user changes. If the commit differs, revalidate
every owner clause, path, route, consumer, finding, task guide, and catalog
entry named here. Never stash, reset, overwrite, or absorb unrelated work.

## 3. Authority and conflict protocol

Resolve decisions in this order:

1. `AGENTS.md` and any newly applicable nested repository instructions.
2. Adopted subsystem NLSpecs within their named scopes.
3. Core 03 for required Workbook behavior and interaction ownership.
4. `docs/design.md` for design, responsive, interaction, accessibility, and
   visual direction.
5. `docs/domain.md` for vocabulary and owner navigation only.
6. Typed projections under `contracts/**` for executable facts downstream of
   their adopted owners.
7. Current implementation, tests, prior handoffs, and retained artifacts as
   implementation evidence.
8. `docs/research/nlspec-spec.md` as advisory planning input only.

The adopted owners are sufficient for this refactor. No normative Core change
and no `docs/domain.md` change is planned. A mismatch between implementation
and an adopted owner is repaired in implementation or its downstream
projection; existing implementation does not become authoritative merely
because consumers depend on it.

If two adopted owner clauses require incompatible outcomes, keep the current
workstream open as `BLOCKED`, record exactly `BLOCKED: owner contradiction`,
quote or precisely identify both clauses, and stop. `BLOCKED` must not be used
for an ordinary implementation failure, an environmental failure, uncertainty,
or work that is merely difficult.

Runtime code, tests, generators, conformance, and release evidence must not
read, stat, hash, or otherwise depend on this or any other Markdown file.

## 4. Scope, structural rules, and exclusions

The future implementation scope is limited to:

- the active-surface saved-view selector;
- Workbook-owned application shortcuts and Timeline editor keys;
- Workbook grid-query controls and registered overlay/menu focus;
- `WorkbookShell` composition and its private collaborators;
- the shell-lifetime Workbook mutation runtime;
- application-owned Workbook paste execution;
- Timeline replay, loading, mutation, model, and collection-renderer hotspots;
- production structural assertions in that exact dependency cone;
- focused tests and browser evidence needed by those changes;
- authored source ownership, test-family, test-catalog, and execution-topology
  inputs when file or evidence movement requires them;
- Make-generated projections of authorized authored input changes;
- the frontend implementation/testing guide and local source-ownership README
  when the new private ownership boundary must be documented; and
- this tracker as the execution ledger.

The implementation must not introduce a compatibility facade, alias, legacy
path, parallel runtime, wildcard dispatcher, universal state bag,
mega-controller, test-only production export, selector fallback, deferred
focus timer, synthetic clipboard event, or Fallow suppression. Keep feature
state with its semantic owner, keep effects behind exact typed ports, and make
malformed or stale boundary input fail closed.

The following are explicitly out of scope:

- route, wire, server, authorization-policy, persistence, migration,
  dependency, generated public contract, or stored-data changes;
- changes to the Grid Adapter public exports, `SemanticDataGridProps`,
  callback payloads, or `GridHandle`;
- repository-wide Fallow or structural-cast cleanup;
- existing Grid Adapter moderate findings;
- Import Assistant work, dependency work, or unrelated harness cleanup;
- Timeline related-record workflow decomposition;
- Evidence or Mention adapter hotspot decomposition;
- broad Timeline presentation or viewport-continuity redesign; and
- hand-editing any generated root or tool-managed dependency artifact.

## 5. Verified starting hotspots

CR-01 must refresh this inventory. The preparation pass found the following
principal source sizes:

| Source | Lines at baseline |
| --- | ---: |
| `ActiveSurfaceSavedViewSelector.tsx` | 700 |
| `WorkbookGridControls.tsx` | 1,007 |
| `WorkbookShell.tsx` | 1,253 |
| `WorkbookMutationRuntime.ts` | 981 |
| `useTimelinePendingReplayController.ts` | 729 |
| `useTimelineRowsLoader.ts` | 573 |
| `useTimelineMutationCommands.ts` | 572 |
| `useTimelineRowMutationCoordinator.ts` | 684 |
| `workbookTimelineModel.ts` | 704 |
| `useTimelineCollectionRenderer.tsx` | 411 |

Direct inspection of the retained Fallow health JSON identified these scoped
production high or critical findings:

| Severity | Finding |
| --- | --- |
| Critical | `WorkbookShellContent` |
| Critical | `WorkbookGridControls` |
| Critical | `WorkbookMutationRuntime.calculateSnapshot` |
| Critical | `EntityWorkbookSurface.handleEntityPaste` |
| Critical | `useTimelineMutationCommands.queueScalarSave` |
| Critical | `useTimelinePendingReplayController.replayPendingQueue` |
| Critical | `useTimelineRowsLoader.loadTimelineRows` |
| Critical | `useTimelineRowMutationCoordinator.applyAcceptedRowMutation` |
| Critical | `useTimelineRowMutationCoordinator.reconcileDiscardedPendingUnit` |
| High | `ActiveSurfaceSavedViewSelector` |
| High | `WorkbookGridControls.handleMenuKeyboard` |
| High | `WorkbookMutationRuntime.drainManagedPatches` |
| High | `WorkbookMutationRuntime.secondaryCandidates` |
| High | `workbookKeyboard.mapWorkbookKeyboardCommand` |
| High | Timeline clipboard callback, collection renderer, and grid-anchor target resolver |

High or critical findings outside the named cone, including existing test
callbacks, related-record workflows, Evidence/Mention adapters, and unrelated
Workbook operations, are not silently added to this tracker. CR-01 must record
them as exclusions when they appear beside scoped findings.

## 6. Gap G01 — Workbook keyboard ownership

**Remediation:** Delete the global `workbookKeyboard.ts` mapper after its
owners are migrated. Replace it with a private, pure Workbook application-
shortcut policy limited to Quick Link, Evidence preview, History, and
inspector-close decisions. Move Timeline editor save and navigation decisions
to a Timeline-local pure policy. Grid navigation, editor entry, paste, copy,
and fill remain exclusively Grid Adapter decisions.

Every policy decision must state whether its binding prevents the browser
default and stops propagation. A command that is unavailable for the current
mode, role, capability, selection, or surface consumes nothing. Normalize
shortcut keys without changing their observable spelling or modifier
semantics.

Add one private registered-ref overlay/menu navigation primitive for Arrow
keys, Home, End, Escape, opening focus, and trigger restoration. It must use
actual element refs and lifecycle state, not selectors, `requestAnimationFrame`,
or deferred timers. It is a focus primitive, not a new global key policy.

**Areas:** Implementation, pure unit tests, component interaction tests,
accessibility/browser evidence, architecture checks, frontend guide, source
ownership, catalog metadata, and tracker.

**Rationale:** Application shortcuts, Timeline editor behavior, grid semantics,
and menu focus have different owners and admission rules. A single growing
mapper creates hidden precedence and event-consumption coupling.

**Long-term benefit:** Each new shortcut or editor mode has one reviewable
policy owner; Grid Adapter semantics cannot drift through Workbook wrappers;
menus share lifecycle-safe focus without sharing feature decisions.

**Compatibility and migration:** User shortcuts stay stable. Internal mapper
types and imports may break and must migrate atomically without aliases.
Unavailable commands intentionally pass through instead of being consumed.

**Risk if unresolved:** Overlapping handlers can suppress browser or editor
behavior, consume unavailable commands, or duplicate Grid Adapter navigation
and clipboard semantics.

**Validation criteria:** Exhaustive decision tables cover case normalization,
modifiers, mode/capability admission, exact consumption flags, inspector close,
Timeline save/navigation, and unavailable-command pass-through. Interaction
evidence covers menu traversal, Escape, opening focus, trigger restoration,
disabled items, unmount, and surface changes. Closure scans find no global
mapper, application-owned grid/paste/fill decisions, selector focus, or focus
timer.

## 7. Gap G02 — `ActiveSurfaceSavedViewSelector`

**Remediation:** Extract a pure active-surface saved-view model, a surface-
keyed form/action reducer, a stale-result-safe action controller, and focused
selector/action presentation. Represent loading, ready, unavailable, and
invalid-selected-view states explicitly. An unresolved initial list must not
be interpreted as a missing selected view.

At effect dispatch, revalidate the current surface, saved-view identity and
version, ownership, and role. Admit only one action at a time. Ignore late UI
results unless surface, selection, version, action kind, and monotonic action
generation still match. Server authorization remains authoritative.

Replace the form-containing `menu` with an accessible disclosure/dialog
boundary using registered focus. Remove the redundant Rename and Manage
Sharing buttons. The labeled Name and Scope inputs plus one `Update view`
action are the only in-place update path. Retain create, duplicate, reset,
delete, set-home, and set-default actions.

Preserve active-surface scoping, normalized query/layout capture, system-view
immutability, and the current routes and payloads.

**Areas:** Implementation, reducer/model/controller tests, component and
accessibility tests, browser evidence, source ownership/catalog metadata,
frontend guide where applicable, and tracker.

**Rationale:** The current component combines asynchronous resource state,
forms, authorization-sensitive effects, and menu presentation. Explicit
states and operation identity prevent loading ambiguity and stale mutation UI.

**Long-term benefit:** Saved-view operations can grow without multiplying
component-local flags, and surface switches or late network results cannot
corrupt the visible form or status.

**Compatibility and migration:** Routes, payloads, authorization, capture,
and system immutability remain unchanged. Intentional changes are removal of
the two redundant buttons, corrected dialog/disclosure semantics, and explicit
inline unavailable, invalid-selection, fallback, and error presentation.

**Risk if unresolved:** Initial loading can masquerade as a deleted view; late
results can update a different surface or version; form controls can violate
menu semantics; duplicated update paths can diverge.

**Validation criteria:** Cover loading, active-surface filtering, invalid and
removed selections, system immutability, authorization loss, stale completion,
rapid surface/selection changes, one-action admission, every retained action,
inline failures, dialog focus, Escape, trigger restoration, and live-region
announcements. Current route and payload fixtures must remain byte-for-byte
compatible where already asserted.

## 8. Gap G03 — `WorkbookGridControls`

**Remediation:** Extract a pure query-control projection and separate Sort,
Group, Filters, Columns, and Active Chips components over narrow model and
command ports. Replace closure-bearing chips and component-wide transient
state with semantic command descriptors and a surface-keyed reducer.

Implement the existing design contract for ordered sorting: add sort, choose
direction, expose priority, move earlier, move later, remove, enforce the
eight-entry admission limit, and provide complete keyboard operation. Keep
header-click sorting semantically separate from View-bar ordered-sort editing.

Validate controlled select values with exact parsers or type guards rather
than assertions. Invalid filter drafts remain visible and locally actionable
instead of entering the canonical query. Retain canonical chip ordering,
responsive capacity, and all-columns-hidden behavior. Use the registered-ref
menu primitive from G01; do not query the DOM or defer focus.

**Areas:** Implementation, pure projection/reducer tests, component tests,
architecture/design tests, responsive/accessibility/browser evidence, authored
ownership/catalog metadata, generated projections when required, frontend
guide, and tracker.

**Rationale:** Sort, group, filter, column, and chip behavior evolve at
different rates. One component with closure commands obscures query semantics
and has already left the ordered-sort design contract incomplete.

**Long-term benefit:** Query features gain independent test and extension
seams, command identity becomes serializable and inspectable, and new control
types no longer enlarge one shared state machine.

**Compatibility and migration:** Persisted query and layout shapes remain
unchanged. The sort menu intentionally changes to the complete ordered-sort
interaction defined by design. Header sort behavior remains compatible.

**Risk if unresolved:** New controls increase state coupling, select casts can
admit impossible values, keyboard focus remains brittle, and UI sort behavior
continues to disagree with its design owner.

**Validation criteria:** Cover add, direction, priority, reorder boundaries,
remove, limit and recovery, duplicate admission, header sorting, grouping,
filter validation, chip order/overflow/removal, all columns hidden, keyboard
traversal, Escape, restoration, surface changes, and responsive modes. Source
scans find no closure-bearing chip command, assertion-based select parsing,
selector focus, or deferred focus frame.

## 9. Gap G04 — `WorkbookShellContent`

**Remediation:** Replace the monolithic content component with thin
composition around cohesive private units for:

- infrastructure and adapter construction;
- startup, saved-view, and query state;
- active-surface query and collaboration registration;
- extension availability and rendering;
- recovery and conflict focus;
- top-bar surface navigation; and
- active-surface presentation.

Move render-time imperative extension discovery into lifecycle-owned
composition. Extract top-bar navigation and recovery presentation using the
registered-focus primitive. Keep `WorkbookShell` as the route-facing component
and preserve one registry lifetime, lazy extension boundaries, responsive
ordering, surface selection, recovery focus, and facade ownership. Do not
replace one monolith with a universal shell state object or mega-controller.

**Areas:** Implementation, composition/lifecycle tests, owner integration and
browser tests, architecture/design evidence, source ownership/catalog
metadata, local README/frontend guide, and tracker.

**Rationale:** The current component makes session, mutation, extension,
query, collaboration, and presentation changes share one render and review
surface. Render-time discovery is an implicit effect.

**Long-term benefit:** Each shell concern has a stable lifecycle and narrow
port; extension and surface additions can be composed without reopening
unrelated mutation or recovery behavior.

**Compatibility and migration:** `WorkbookShell` remains the route boundary.
Lazy loading, registry identity, navigation, query/collaboration registration,
recovery focus, responsive order, selectors used only as passive test
attributes, and external facades remain compatible.

**Risk if unresolved:** Any new surface or startup concern can perturb session
state, extension discovery, mutation runtime lifetime, focus, or query
registration in the same component.

**Validation criteria:** Prove one adapter/runtime/registry construction per
required lifetime, lifecycle-owned discovery, correct registration cleanup,
lazy extension behavior, base/extension switches, recovery/conflict focus,
responsive ordering, access loss, startup/saved-view states, and
reference-preserving no-ops. Architecture scans reject render-time mutation,
mega state bags, duplicate registries, and new public exports.

## 10. Gap G05 — Workbook mutation runtime

**Remediation:** Retain exactly one shell-lifetime runtime while decomposing it
into a status projector, conflict store, surface registry, managed-patch
driver, retry scheduler, transaction ledger, and lifecycle coordinator.
Scheduling and clock effects must be injected for deterministic tests.

Replace untyped queue metadata with a closed `managed_patch | timeline_row`
envelope and exact driver registrations. The runtime owns FIFO drain
serialization, retry and authentication gates, terminal disposal, capacity,
conflicts, save-state projection, refresh debt, and client transaction
tracking. Owner drivers own payload-specific revalidation, transport intent,
and result projection.

Remove `pending<TMeta>()`, `registerDrainer`, shared mutable replay bags, and
the generic metadata cast after Timeline has migrated. A missing or unmounted
owner driver pauses eligible work without loss. Duplicate registration fails
closed and cannot silently replace an active owner.

**Areas:** Runtime implementation, pure state/scheduler/ledger tests,
integration tests, owner adapters, architecture/static checks, source
ownership/catalog metadata, frontend guide, and tracker.

**Rationale:** Queue mechanics and owner-specific replay have become
interleaved through generic metadata and callback registration. The runtime
should coordinate lifecycle; owners should interpret their own units.

**Long-term benefit:** New mutation owners can add one exact envelope member
and driver without copying retry/replay infrastructure. Deterministic clock
tests remove timing flakiness and make failure behavior reviewable.

**Compatibility and migration:** Keep one runtime and preserve public app
behavior, queue ordering, retries, conflicts, save status, refresh debt, and
transaction semantics. Internal runtime types may break atomically. No second
runtime or compatibility facade is permitted.

**Risk if unresolved:** Generic metadata casts can dispatch a unit to the
wrong logic, callback replacement can lose work, timers make lifecycle races
unrepeatable, and each owner may grow its own replay authority.

**Validation criteria:** Cover admission, coalescing, strict order, capacity,
secure-ID failure, retry schedule, authentication pause/recovery, terminal
dispose, explicit mutation state, conflict refresh/resolution, refresh debt,
driver absence/remount, duplicate registration, missing or wrong metadata,
socket transaction settlement, and runtime disposal. No scoped runtime high
or critical finding, suppression, generic metadata path, or parallel runtime
may remain.

## 11. Gap G06 — Application-owned paste execution

**Remediation:** Introduce one private Workbook clipboard-paste transport port
and one exact request/response adapter for the existing route. Delete
`EntityMutationCommandPort.pasteCreate`, `TimelineClipboardPastePort`, and the
Timeline-specific duplicate transport adapter after all consumers migrate.

Grid Adapter responsibility ends at clipboard decoding and semantic target
planning. Immediately before dispatch, each owner-local Workbook controller
revalidates the active surface, interaction mode, writable fields, semantic row
identities, current row versions, grouping, and creation capability. Preserve
Generic scalar editing, Entity entity-origin rectangular creation, and
Timeline mixed existing/create batches.

Separate native editor paste from grid paste. Do not fabricate a React
clipboard event or bridge the two paths with a selector, zero-delay timer, or
polling. Apply accepted rows, grouped conflicts, save state, refresh, focus,
and viewport continuity through explicit owner ports.

**Areas:** Private transport and owner controllers, component/runtime
integration, pure admission/application tests, browser clipboard tests,
architecture/static evidence, source ownership/catalog metadata, frontend
guide, and tracker.

**Rationale:** Clipboard decoding, semantic planning, application mutation,
and owner projection are separate concerns. Surface-specific transport ports
duplicate an invariant route while obscuring last-moment authorization and
version checks.

**Long-term benefit:** New surfaces reuse one exact transport while retaining
owner-specific admission and projection. Native editor behavior remains
independent of grid multi-cell paste.

**Compatibility and migration:** The route, wire body, transaction semantics,
clipboard parsing, exact-match Entity behavior, Timeline change-set behavior,
conflict handling, and Generic scalar behavior remain unchanged. Internal
ports break atomically without a facade.

**Risk if unresolved:** Stale surfaces or rows can be mutated, duplicate
adapters can drift, and synthetic events/timers can cross editor and grid
ownership unpredictably.

**Validation criteria:** Cover comma scalar text, quoted CSV and TSV, native
editor paste, existing/create targets, wrong surface/type, deleted/stale rows,
closed/read-only mode, grouped creation refusal, exact Entity origin,
Timeline mixed batches, grouped conflicts, transport/apply failure, accepted
row application, one refresh, focus, and continuity. Closure scans find no
duplicate ports or adapter, synthetic clipboard event, selector bridge,
polling, or zero-delay timer.

## 12. Gap G07 — Timeline replay, load, mutation, and model hotspots

**Remediation:** Replace `useTimelinePendingReplayController` with a typed
Timeline runtime driver containing pure admission, revalidation, settlement,
conflict, discard, and accepted-result plans. The common runtime performs
scheduling and transport orchestration.

Extract a pure Timeline load state machine for request identity, mutation
epochs, source-version requirements, bounded freshness retry, stale-result
joining, initial/refresh failures, and accepted projection commits. Keep
`flushSync`, if still required, only in a narrow commit adapter.

Split `useTimelineMutationCommands` into mutation-intent construction, queue
admission, and editor/grid adapters. Split
`useTimelineRowMutationCoordinator` into accepted-row projection,
discarded-unit reconciliation, committed-version admission, conflict
projection, socket transaction tracking, and save-state composition.

Divide `workbookTimelineModel` into field registry, row normalization and
materialization, mutation intents, and layout/presentation policy. Replace
collection-renderer downcasts with discriminated relationship and tag
presentation models.

Preserve high-water row versions, exactly one bottom draft row, pending
signatures, conflict drafts, created-row pinning, mentions, auto-resolution
notices, refresh obligations, and viewport continuity. Do not expand this work
into related-record workflow, Evidence/Mention adapter, general presentation,
or continuity redesign beyond narrow port adaptation.

**Areas:** Timeline and runtime implementation, pure state/plan tests, owner
integration and browser tests, architecture/static analysis, source
ownership/catalog metadata, frontend guide, and tracker.

**Rationale:** Replay scheduling, load races, mutation construction, accepted
projection, and presentation typing currently overlap in large hooks and
models. Each has a distinct invariant and failure domain.

**Long-term benefit:** Timeline phases can extend mutation kinds, loading
sources, or collection fields through exact models without reopening replay
scheduling or the entire row coordinator.

**Compatibility and migration:** Observable Timeline mutation, replay,
loading, conflict, mention/notice, draft, pinning, refresh, and viewport
behavior remains compatible. Internal hooks and types may be deleted or
renamed without aliases.

**Risk if unresolved:** Stale loads can overwrite accepted mutations, replay
and runtime scheduling can disagree, casts can misclassify collection fields,
and future Timeline features must enter several shared mutable paths.

**Validation criteria:** Cover superseded/aborted loads, concurrent accepted
mutations, source-version obligations, bounded non-convergence, initial versus
refresh failures, access loss, malformed rows, local draft hydration,
high-water versions, one draft, created-row pinning, replayed create/patch,
scalar/collection saves, keyboard/blur deduplication, conflicts, notices,
mentions, stale selections, discard, driver remount, socket settlement,
refresh, focus, and continuity. Named/replacement paths must have no high or
critical Fallow finding or suppression.

## 13. Gap G08 — Remaining structural assertions

**Remediation:** Inventory every production structural assertion in the named
Workbook/Timeline dependency cone during CR-01, then eliminate it at its
owning boundary through exact construction, type guards, validated decoders,
or discriminated registries.

Targets include `as unknown as`, generated-request assertions, generic
metadata casts, union-member downcasts, synthetic event casts,
assertion-based enum parsing, and unchecked array-index assertions. The known
set includes Timeline grid-anchor, committed-row, replay admission, collection
rendering, field binding, mutation reconciliation, mention decoding, and paste
event assertions, plus shared runtime and paste assertions.

Literal `as const`, `satisfies`, typed CSS literals, and test-only DOM element
assertions are not targets. Generated files remain untouched. Malformed
boundary values must fail closed. This is not repository-wide cast cleanup.

**Areas:** Implementation throughout the scoped cone, negative decoder tests,
compile-time/typecheck evidence, architecture/static scans, ownership metadata,
and tracker.

**Rationale:** Structural assertions conceal missing runtime proof at exactly
the transport, registry, and union boundaries made extensible by G05 through
G07.

**Long-term benefit:** Adding an envelope member, collection kind, field, or
paste target produces an exhaustive compile failure or explicit boundary
rejection instead of an unsafe runtime assumption.

**Compatibility and migration:** Valid inputs remain compatible. Malformed or
impossible internal values intentionally fail closed rather than being forced
through a cast. Generated public contracts are unchanged.

**Risk if unresolved:** New union members may enter incorrect branches,
malformed server or registry values may be treated as trusted shapes, and
runtime refactors may preserve hidden unsoundness.

**Validation criteria:** CR-01 records path, line, assertion class, owner, and
planned closure workstream. CR-11 proves the inventory empty except explicit
non-target categories, adds negative decoder/exact-request tests, passes
typecheck/import boundaries/Biome, and finds no replacement cast or wildcard
dispatch.

## 14. Dependency and workstream ledger

| Workstream | Status | Dependency | Binary exit condition |
| --- | --- | --- | --- |
| CR-01 — Authority, inventory, and characterization | DONE | Separate implementation authorization | Authorities, consumers, routes, metadata, casts, scoped findings, and passing deletion baselines are refreshed without an owner contradiction. |
| CR-02 — Keyboard ownership | DONE | CR-01 `DONE` | Application and Timeline policies have exact consumption semantics; the global mapper and non-Grid grid decisions are gone. |
| CR-03 — Saved-view control | DONE | CR-02 `DONE` | Explicit resource/action state and accessible registered focus replace the monolith and redundant actions. |
| CR-04 — Grid controls | DONE | CR-03 `DONE` | Independent controls and semantic descriptors implement the complete ordered-sort and menu-focus contracts. |
| CR-05 — Shell composition | DONE | CR-04 `DONE` | `WorkbookShell` is thin lifecycle-owned composition with no duplicate registry, render-time discovery, or mega state bag. |
| CR-06 — Mutation-runtime decomposition | DONE | CR-05 `DONE` | One runtime is decomposed into deterministic private responsibilities while existing consumers remain behaviorally stable. |
| CR-07 — Timeline replay convergence | DONE | CR-06 `DONE` | Timeline uses the typed runtime driver and closed envelope; old replay/drainer/generic metadata paths are deleted. |
| CR-08 — Application paste execution | DONE | CR-07 `DONE` | One Workbook paste transport and owner-local admission/application replace duplicate ports and synthetic bridging. |
| CR-09 — Timeline loading | DONE | CR-08 `DONE` | A pure load state machine proves request/mutation/version races and isolates the projection commit effect. |
| CR-10 — Timeline mutation and model decomposition | DONE | CR-09 `DONE` | Mutation, reconciliation, field, row, layout, and collection responsibilities are exact, discriminated, and independently tested. |
| CR-11 — Structural-cast closure | DONE | CR-10 `DONE` | Every scoped target assertion is removed or fail-closed at its boundary with negative and compile-time evidence. |
| CR-12 — Validation and handoff | DONE | CR-11 `DONE` | Catalogs and projections are reconciled; the terminal matrix, Fallow inspection, closure searches, and handoff gates pass. |

Only one workstream may be `IN_PROGRESS`. Work strictly in ledger order. A
workstream may start only after a distinct pre-workstream checkpoint and its
predecessor is `DONE`. Its tracker record must be complete before the successor
checkpoint changes the successor to `IN_PROGRESS`. A required failure leaves
the current workstream `IN_PROGRESS` unless the authority conflict protocol
applies.

## 15. CR-01 through CR-04 — Interaction and control foundations

### CR-01 — Authority, inventory, and characterization

1. Refresh the repository and authority baseline required by section 2.
2. Re-read Core 03, relevant adopted subsystem owners, `docs/design.md`, and
   `docs/domain.md`; record that the owners are sufficient or invoke the exact
   conflict protocol.
3. Inventory all named source consumers, private/public exports, route and
   payload adapters, mutation/runtime registrations, extension boundaries,
   source-ownership entries, test families, catalog rows, task guides,
   generated projections, and current browser fixtures.
4. Inspect the retained Fallow JSON directly, then run fresh owner-only-umask
   Fallow and classify scoped findings separately from repository-wide and
   explicitly excluded findings.
5. Record every scoped production structural assertion by path, line,
   category, owner, replacement technique, and planned closing workstream.
6. Add passing characterization around behavior that will be deleted or
   moved. Record expected-red design gaps as tracker evidence; do not commit
   failing tests or normalize the missing behavior as compatibility.
7. Record baseline source sizes, tests, route fixtures, observable behavior,
   and stable public boundaries.

**Required evidence:** Refreshed task guides and focused characterization for
Workbook keyboard, saved views, grid controls, shell composition, mutation
runtime, all three paste owners, and Timeline replay/load/mutation/model paths;
typecheck and applicable static evidence; fresh Fallow artifact and direct JSON
inventory.

**Exit:** CR-C01 passes. All deletion baselines, assertions, findings,
consumers, authorities, routes, and metadata are inventoried; characterization
is passing; no owner contradiction remains.

**Tracker gate:** Record paths, inventories, decisions, commands/run roots,
failures and classification, compatibility, rollback, residual risk, and next
action; mark CR-01 and CR-C01 `DONE`; append a separate CR-02 checkpoint.

### CR-02 — Keyboard ownership

1. Add pure decision tables for Workbook application shortcuts and Timeline
   editor save/navigation before deleting the global behavior.
2. Add the registered-ref overlay/menu focus primitive and its lifecycle and
   accessibility tests.
3. Migrate every consumer to its correct policy owner. Remove application-side
   grid navigation, paste, fill, or editor-entry branches.
4. Delete `workbookKeyboard.ts`, its obsolete types/tests, and all internal
   imports without an alias.
5. Update authored source ownership and test metadata for moved or new units;
   generate and inspect projections only if those authored inputs require it.
6. Update the frontend implementation/testing guide with the application,
   Timeline editor, Grid Adapter, and registered-focus ownership boundary.

**Required evidence:** Decision-table tests; Workbook and Timeline component
tests; menu/overlay accessibility evidence; focused `web.workbook`,
`web.architecture`, `web.design`, `package.grid_adapter`, `module.workbook`,
and `module.timeline` routes as task guides require; typecheck, frontend unit,
import boundary, Biome, scoped Fallow, and closure searches.

**Exit:** CR-C02 passes. Exact event consumption is proven, unavailable
commands pass through, Grid Adapter is the sole grid-key owner, and no global
mapper, selector focus, deferred focus, alias, or duplicate policy remains.

**Tracker gate:** Complete the full workstream record; mark CR-02 and CR-C02
`DONE`; append a separate CR-03 checkpoint.

### CR-03 — Saved-view control

1. Freeze current valid action requests, system immutability, surface scoping,
   query/layout capture, and route/payload behavior.
2. Add the explicit resource model, keyed reducer, operation identity, and
   stale-result-safe controller with pure tests before migrating effects.
3. Revalidate identity, version, surface, ownership, and role at dispatch; make
   duplicate action admission fail closed.
4. Replace menu-form semantics with the accessible disclosure/dialog and
   registered focus. Add explicit unavailable, invalid-selection, fallback,
   busy, success, and error presentation.
5. Remove Rename and Manage Sharing buttons and their action paths. Keep one
   `Update view` flow and all named retained actions.
6. Delete old state/effect branches only after parity and intentional-change
   evidence is green; reconcile metadata and guide material.

**Required evidence:** Pure state and operation race tests; component and
browser interaction; live-region/accessibility checks; focused Workbook,
architecture, design, collaboration, and applicable service-backed routes;
typecheck, unit, import boundaries, Biome, scoped Fallow, and closure searches.

**Exit:** CR-C03 passes. Loading cannot masquerade as missing data, late
results cannot affect a changed subject, one update path remains, and focus and
feedback satisfy the owner/design contract.

**Tracker gate:** Complete the full workstream record; mark CR-03 and CR-C03
`DONE`; append a separate CR-04 checkpoint.

### CR-04 — Grid controls

1. Freeze persisted query/layout serialization, header-sort semantics, chip
   order, responsive behavior, and surface switching.
2. Add pure query-control projection, semantic command descriptors, and the
   surface-keyed reducer.
3. Extract Sort, Group, Filters, Columns, and Active Chips presentation over
   narrow ports; do not expose private units publicly.
4. Implement ordered-sort addition, direction, priority, reorder, removal,
   eight-entry admission, and complete keyboard behavior.
5. Replace assertion parsing with exact validation; keep invalid filter drafts
   local and visible. Adopt registered menu refs and delete selector/frame
   focus.
6. Reconcile ownership/catalog inputs, generated projections, design evidence,
   and the frontend guide.

**Required evidence:** Full control matrix from section 19; responsive,
accessibility, stateful, and visual evidence where changed; focused Workbook,
architecture, design, and affected module routes; typecheck, unit, import
boundaries, Biome, scoped Fallow, and closure searches.

**Exit:** CR-C04 passes. The design-owned ordered-sort behavior is complete,
each control owns one concern, canonical serialization is unchanged, and the
old monolith and brittle focus/parsing paths are absent.

**Tracker gate:** Complete the full workstream record; mark CR-04 and CR-C04
`DONE`; append a separate CR-05 checkpoint.

## 16. CR-05 through CR-08 — Composition, runtime, replay, and paste

### CR-05 — Shell composition

1. Characterize construction/lifetime counts, registration cleanup, lazy
   extension loading, surface changes, recovery focus, and responsive order.
2. Extract infrastructure/adapters, startup/saved-query state,
   active-surface query/collaboration, extension, recovery, top-bar, and
   active-presentation units in cohesive increments.
3. Move extension discovery from render into an explicit lifecycle owner.
4. Retain `WorkbookShell` as the route component and keep facade ownership and
   mutation runtime lifetime stable.
5. Delete superseded composition only after focused route, lifecycle, and
   browser evidence passes. Reject a generic shell state bag or controller.
6. Reconcile source ownership, catalogs, local README, frontend guide, and any
   generated projections required by authored changes.

**Required evidence:** Shell composition and lifecycle tests; all base and
extension surface transitions; startup, access loss, recovery/conflict focus,
responsive and lazy-load evidence; focused Workbook, architecture, design,
Network Flow, and affected module routes; typecheck, unit, import boundaries,
Biome, Fallow, and structural scans.

**Exit:** CR-C05 passes. The route boundary is thin, effects are lifecycle-
owned, registry/runtime identities are stable, and no duplicate controller,
state bag, or render-time imperative discovery remains.

**Tracker gate:** Complete the full workstream record; mark CR-05 and CR-C05
`DONE`; append a separate CR-06 checkpoint.

### CR-06 — Mutation-runtime decomposition

1. Freeze queue, status, conflict, retry/auth, refresh-debt, transaction, and
   lifecycle behavior with deterministic characterization.
2. Introduce injected clock/scheduler ports and split status projector,
   conflict store, surface registry, managed-patch driver, retry scheduler,
   transaction ledger, and lifecycle coordinator behind the one runtime.
3. Preserve current owner consumers during this workstream without creating a
   parallel runtime or compatibility facade. The closed final envelope may be
   staged privately but becomes authoritative only with CR-07 migration.
4. Make driver absence pause eligible units and duplicate registration fail
   closed. Prove terminal disposal and remount behavior.
5. Delete superseded internal branches as their focused evidence turns green;
   reconcile ownership, tests, guide, and generated projections.

**Required evidence:** Deterministic scheduler and runtime state tests; queue,
retry/auth, conflict, status, capacity, refresh, transaction, driver, remount,
and disposal matrix; focused Workbook, architecture, and affected module
routes; typecheck, unit, import boundaries, Biome, scoped Fallow, and scans for
parallel state.

**Exit:** CR-C06 passes. One runtime owns common lifecycle policy through
cohesive private units and retains observable behavior while exposing the
exact driver seam needed by CR-07.

**Tracker gate:** Complete the full workstream record; mark CR-06 and CR-C06
`DONE`; append a separate CR-07 checkpoint.

### CR-07 — Timeline replay convergence

1. Characterize replay admission, versions, conflicts, discard, transport
   errors, result application, remount, transactions, and refresh obligations.
2. Add pure Timeline driver plans and exact tests before moving orchestration.
3. Install the closed `managed_patch | timeline_row` queue envelope and exact
   driver registrations.
4. Migrate Timeline replay to the common runtime's scheduling and transport
   lifecycle while leaving payload revalidation and result projection with
   Timeline.
5. Delete `useTimelinePendingReplayController`, `registerDrainer`,
   `pending<TMeta>()`, shared replay bags, and the generic metadata cast. Do not
   leave adapters or aliases.
6. Reconcile ownership/catalog inputs, generated projections, frontend guide,
   and all affected owner tests.

**Required evidence:** Driver plan tables and integration tests for replayed
create/patch, missing/wrong metadata, driver absence/remount, duplicate
registration, conflicts, discard, transport/apply failure, refresh, save state,
and socket settlement; focused Workbook, Timeline, architecture, and affected
module/service routes; typecheck, unit, import boundaries, Biome, Fallow, and
closure searches.

**Exit:** CR-C07 passes. The common runtime is the only replay scheduler,
Timeline is an exact owner driver, and the old generic/drainer/replay machinery
is absent.

**Tracker gate:** Complete the full workstream record; mark CR-07 and CR-C07
`DONE`; append a separate CR-08 checkpoint.

### CR-08 — Application paste execution

1. Freeze parser and wire fixtures, scalar/editor behavior, Entity exact origin,
   Timeline mixed batches, conflicts, refresh, focus, and continuity.
2. Add the exact Workbook transport port/adapter and owner-local admission and
   result-application plans.
3. Migrate Generic, Entity, and Timeline paths while retaining Grid Adapter's
   exclusive decode/semantic-plan boundary.
4. Revalidate all current mutation facts immediately before transport
   dispatch. Ensure native editor paste never enters the grid path.
5. Delete the Entity and Timeline duplicate ports/adapters, synthetic clipboard
   event, selector/timer bridges, and surface-local transport orchestration.
6. Reconcile ownership/catalog metadata, generated projections, frontend
   guide, and browser fixture ownership.

**Required evidence:** Full paste matrix from section 19; exact request and
response fixtures; owner application, conflict, status, refresh, focus, and
continuity evidence; focused Grid Adapter, Workbook, Timeline, Entities,
architecture/design, and applicable service-backed routes; typecheck, unit,
import boundaries, Biome, browser clipboard evidence, Fallow, and closure
searches.

**Exit:** CR-C08 passes. One private Workbook transport serves exact
owner-local admission/application paths, observable route/wire behavior is
stable, and duplicate/synthetic bridges are absent.

**Tracker gate:** Complete the full workstream record; mark CR-08 and CR-C08
`DONE`; append a separate CR-09 checkpoint.

## 17. CR-09 through CR-12 — Timeline closure and terminal handoff

### CR-09 — Timeline loading

1. Freeze current load/query/mutation race, high-water version, draft, pinning,
   source-version, error, refresh, and continuity behavior.
2. Define the pure load state, events, effects, request identity, mutation
   epoch, convergence bound, and accepted commit plan with exhaustive tests.
3. Move transport/lifecycle effects behind narrow ports and isolate
   `flushSync`, if still necessary, in the projection commit adapter.
4. Migrate the owner hook and delete superseded mutable/request branches only
   after the race matrix passes.
5. Reconcile ownership/catalog metadata, generated projections, frontend
   guide, and Fallow inventory.

**Required evidence:** Superseded/aborted loads, stale joins, concurrent
accepted mutations, high-water rows, bounded non-convergence, source-version
obligations, initial/refresh error, access loss, malformed rows, draft
hydration, pinning, refresh, and continuity; focused Workbook/Timeline and
service-backed routes; typecheck, unit, import boundaries, Biome, Fallow, and
closure scans.

**Exit:** CR-C09 passes. Load transitions and effects are explicit, stale data
cannot overwrite accepted state, one narrow commit adapter owns imperative
projection, and the former load hotspot is absent.

**Tracker gate:** Complete the full workstream record; mark CR-09 and CR-C09
`DONE`; append a separate CR-10 checkpoint.

### CR-10 — Timeline mutation and model decomposition

1. Characterize scalar/collection mutation intent, queue admission,
   editor/grid deduplication, accepted/discarded reconciliation, committed
   versions, conflicts, transactions, save state, field binding, row
   materialization, layout, and collection rendering.
2. Extract exact mutation-intent, queue, editor, and grid units.
3. Extract accepted projection, discarded reconciliation, committed-version,
   conflict, socket transaction, and save-state units.
4. Split the Timeline model into field registry, row normalization/
   materialization, mutation intents, and layout/presentation policy.
5. Install discriminated relationship/tag presentation models and delete
   union-member downcasts.
6. Preserve all invariants listed in G07 and avoid the explicitly excluded
   related-record, adapter, presentation, and continuity expansions.
7. Reconcile ownership/catalog inputs, generated projections, frontend guide,
   and affected tests.

**Required evidence:** Mutation intent and reducer tables; scalar/collection,
keyboard/blur deduplication, accepted/discarded/version/conflict/transaction/
status cases; malformed and additive field/collection cases; reference-
preserving no-ops; focused Workbook/Timeline and affected owner/service routes;
typecheck, unit, import boundaries, Biome, Fallow, and scans.

**Exit:** CR-C10 passes. Each Timeline mutation/model concern has one exact
unit and discriminated boundary; observable invariants are stable; the named
hotspots and parallel branches are absent.

**Tracker gate:** Complete the full workstream record; mark CR-10 and CR-C10
`DONE`; append a separate CR-11 checkpoint.

### CR-11 — Structural-cast closure

1. Reconcile the CR-01 assertion ledger against current production source.
2. Remove every remaining scoped target through exact constructors, guards,
   decoders, discriminated registries, or checked indexing.
3. Add malformed-boundary and negative-member tests and exact request
   construction fixtures. Add compile-time exhaustiveness evidence where it
   improves future extension safety.
4. Inspect every replacement path for equivalent casts, non-null assertions,
   wildcard/default dispatch, or unchecked fallback.
5. Leave generated files and explicit non-target assertion categories
   unchanged; record every residual match and why it is outside scope.

**Required evidence:** Empty scoped assertion ledger except documented
non-targets; negative decoders; exact request tests; typecheck, frontend unit,
import boundaries, Biome, architecture checks, scoped Fallow, and direct
closure searches.

**Exit:** CR-C11 passes. No scoped structural assertion or equivalent escape
hatch remains, malformed values fail closed, and future union expansion is
exhaustive.

**Tracker gate:** Complete the full workstream record; mark CR-11 and CR-C11
`DONE`; append a separate CR-12 checkpoint.

### CR-12 — Validation and handoff

1. Reconcile every authored source-ownership, test-family, test-catalog, and
   execution-topology path with added, moved, renamed, or deleted source and
   evidence. Run Make-owned generation and inspect every generated diff.
2. Refresh task guides for every owner in section 19 and run the focused and
   applicable service-backed slices they declare.
3. Run `make format`, then `make agent-finalize`. Supply `RESULTS_DIR` only for
   a genuine compatible successful retained full warm run; otherwise record
   that retained-run maintenance was skipped because it was unset.
4. Run the terminal matrix in section 19, including two ordinary visual runs
   after any approved visual update.
5. Run fresh owner-only-umask Fallow and inspect its JSON directly. Require no
   high or critical finding or suppression in any named or replacement
   production path.
6. Run all closure searches, Markdown lint, `git diff --check`, generated and
   golden scope inspection, and staged/unstaged diff review.
7. Resolve CR-C01 through CR-C12, update every workstream to `DONE`, and record
   compatibility, rollback, residual risks, exclusions, deferrals, and the next
   extension seam.

**Exit:** CR-C12 passes. All gates and workstreams are `DONE`, required evidence
is successful or explicitly classified, no unintended generated/golden/public
contract change remains, and no authorized work is left.

**Tracker gate:** Append the final handoff checkpoint with paths, decisions,
commands/run roots, failure classifications, compatibility, rollback,
residual risks, deferrals, and next extension seam. Do not start another
successor within this tracker.

## 18. Acceptance-gate ledger

| Gate | Status | Acceptance statement |
| --- | --- | --- |
| CR-C01 — Authority and characterization | DONE | Refreshed owners, baseline, consumers, routes, metadata, cast ledger, scoped Fallow inventory, and passing characterization establish a safe deletion baseline. |
| CR-C02 — Keyboard ownership | DONE | Application shortcuts, Timeline editor keys, Grid Adapter keys, and registered focus have distinct owners and exact event-consumption evidence. |
| CR-C03 — Saved-view lifecycle | DONE | Explicit resource/action identity, stale-result rejection, one update path, authorization revalidation, and accessible dialog focus are proven. |
| CR-C04 — Query controls | DONE | Independent controls satisfy ordered sort, grouping, filter, column, chip, responsive, and keyboard contracts without unsafe parsing or focus. |
| CR-C05 — Shell composition | DONE | The shell is thin lifecycle composition with stable facades, registry/runtime lifetime, surface behavior, recovery focus, and extension loading. |
| CR-C06 — Runtime structure | DONE | One deterministic runtime owns queue lifecycle through cohesive private units and exact driver registration behavior. |
| CR-C07 — Replay convergence | DONE | The closed mutation envelope and Timeline driver replace generic metadata, drainer registration, and parallel replay state. |
| CR-C08 — Paste ownership | DONE | One private transport plus owner-local revalidation/application preserves all three surface behaviors and removes duplicate or synthetic bridges. |
| CR-C09 — Timeline loading | DONE | The pure load machine preserves version, mutation, freshness, error, draft, pinning, refresh, and continuity invariants under races. |
| CR-C10 — Timeline mutation/model | DONE | Exact mutation, reconciliation, registry, row, layout, and discriminated collection units replace the named hotspots without scope expansion. |
| CR-C11 — Structural type safety | DONE | Every scoped production structural assertion is eliminated or an explicit non-target; malformed boundaries fail closed and union growth is exhaustive. |
| CR-C12 — Terminal handoff | DONE | Catalogs/projections, routed and terminal checks, Fallow, closure scans, Git scope, compatibility, rollback, residual risk, and handoff are complete. |

A gate changes to `DONE` only with its owning workstream. Do not use `PASS`,
`PARTIAL`, or prose as a substitute for ledger status. A required failure keeps
both the gate and workstream open.

## 19. Validation and evidence matrix

### 19.1 UI and keyboard coverage

- Saved-view loading, active-surface filtering, invalid selection, system-view
  immutability, authorization, stale async completion, one-action admission,
  create, update, duplicate, reset, delete, home/default, dialog focus,
  fallback/error presentation, and status announcements.
- Ordered sort addition, direction, visible priority, reordering, removal,
  eight-entry admission, grouping, invalid filter drafts, chip ordering and
  overflow, all-columns-hidden, menu traversal, Escape, trigger restoration,
  surface changes, and responsive modes.
- Shortcut case normalization, modifiers, mode and capability admission,
  inspector close, Timeline editor save/navigation, unavailable-command
  pass-through, and proof that Grid Adapter alone owns grid navigation, editor
  entry, paste, copy, and fill.

### 19.2 Mutation and paste coverage

- Queue admission, coalescing, ordering, capacity, secure-ID failure, retries,
  authentication pause/recovery, terminal disposal, refresh debt, remount,
  explicit mutation state, conflict refresh/resolution, and current-version
  revalidation.
- Scalar comma text, quoted CSV and TSV, existing and create targets, wrong
  surface/type, deleted/stale targets, closed/read-only mode, grouped creation
  refusal, exact Entity origin, Timeline mixed targets, grouped conflicts, one
  refresh, row application, native editor paste, grid paste, focus, and
  continuity.
- Runtime-driver absence/remount, duplicate registration, missing/wrong
  metadata, transport failure, apply failure, discard reconciliation, and
  socket transaction settlement.

### 19.3 Timeline race and invariant coverage

- Superseded and aborted loads, stale-result joining, concurrent accepted
  mutations, high-water row versions, bounded non-convergence, source-version
  obligations, initial versus refresh error, access loss, malformed rows,
  local draft hydration, created-row pinning, and exactly one draft row.
- Replayed create/patch settlement, scalar and collection saves,
  keyboard/blur deduplication, conflict draft preservation, notices and
  mentions, stale selection, discarded units, and reference-preserving no-ops.

### 19.4 Required final routes and repository checks

Refresh `make task-guide ROLE=module-author OWNER=<owner-id>` before selecting
rows. The declared final owner set is:

- `web.workbook`, `web.architecture`, `web.design`, and `web.networkflow`;
- `package.grid_adapter`;
- `module.workbook`, `module.timeline`, `module.entities`,
  `module.collaboration`, `module.evidence`, `module.assessments`,
  `module.indicators`, and `module.networkflow`; and
- every applicable service-backed slice declared by refreshed task guides.

The CR-12 terminal matrix includes:

- `make generate` and inspection of all generated diffs;
- `make format`, followed by `make agent-finalize`;
- generation drift, generated-artifact policy, and JSON-shape checks;
- frontend typecheck, unit, import-boundary, and Biome checks;
- accessibility, measurement, stateful, support, webserver-backed, and visual
  browser evidence;
- `make test-fast`;
- fresh owner-only-umask Fallow with direct JSON inspection; and
- Markdown lint, `git diff --check`, and staged/unstaged scope review.

Run the narrowest relevant route within each workstream, then broaden only as
its ownership and risk require. Record the exact command, result, retained run
root, failing row/summary when applicable, and whether each failure is
introduced, pre-existing, environmental, or unrelated. A retry does not erase
the first failure record.

### 19.5 Closure searches

CR-12 must search production source, authored metadata, and relevant tests for:

- the deleted global Workbook mapper and copied keyboard decisions;
- selector/timer/frame-driven menu or overlay focus;
- redundant Rename and Manage Sharing saved-view actions;
- the old runtime generic pending metadata, drainer, and replay-bag paths;
- duplicate Entity/Timeline paste ports or transports;
- synthetic clipboard events and editor/grid paste bridges;
- the old Timeline replay controller and replacement parallel replay paths;
- every CR-01 scoped structural assertion and equivalent escape hatch;
- compatibility aliases, facades, legacy imports, wildcard dispatch, and
  test-only production exports;
- new TODO/FIXME markers in the changed dependency cone;
- misplaced Grid Adapter vendor imports; and
- Markdown runtime or test dependencies.

Require no high or critical Fallow finding or suppression in any named or
replacement production path. Do not claim repository-wide Fallow cleanup.

## 20. Tracker protocol and checkpoint schema

Before every workstream:

1. Confirm the predecessor and its acceptance gate are `DONE`.
2. Append a distinct pre-workstream checkpoint; do not rewrite old evidence.
3. Refresh relevant owners, repository status, task guides, and changed-path
   metadata.
4. Change only that workstream to `IN_PROGRESS`; leave its gate `PLANNED` until
   the exit condition passes.
5. Characterize current behavior before deletion or replacement.

At every workstream gate, append a checkpoint containing all of:

- timestamp, branch, commit, upstream relation, authorization, and worktree
  scope;
- workstream and acceptance-gate status;
- changed, added, moved, generated, and deleted paths;
- owner/design decisions and intentional behavior changes;
- commands, selected rows, run roots, summaries, and direct artifact review;
- every failure and its introduced/pre-existing/environmental/unrelated
  classification;
- compatibility and migration effect;
- a source-only workstream-granular rollback unit;
- residual risk, exclusions, and deferrals; and
- the exact next action.

Keep the workstream `IN_PROGRESS` while any required check fails. After all
exit conditions pass, mark its acceptance gate and workstream `DONE`. Append a
separate successor checkpoint before marking the successor `IN_PROGRESS`.

Use this append-only form:

| Timestamp | Status transition | Paths and decisions | Evidence, compatibility, rollback, risk, and next action |
| --- | --- | --- | --- |
| _Append only_ | _Example: CR-01 `DONE`; CR-02 remains `PLANNED`_ | _Exact paths and decisions_ | _Commands/run roots, failures, compatibility, rollback, residual risk, next action_ |

## 21. Catalog, generation, and documentation discipline

When an implementation workstream adds, moves, renames, or deletes source or
tests, update the authored owner inputs under `tools/**` in that same
workstream. Run the Make-owned generator and inspect generated projections
before closing the workstream. Never hand-edit generated roots or generated
task/topology outputs.

Update `docs/guides/cartulary_frontend_implementation_testing_guide.md` as the
new keyboard, registered-focus, query-control, runtime-driver, and paste
ownership boundaries become real. Update `apps/web/src/README.md` only when its
source navigation becomes stale. These documents describe implemented
boundaries; they do not define runtime requirements.

Do not commit failing characterization tests. An owner/design behavior that is
currently missing is an expected-red tracker item until its implementation
workstream lands the corrected assertion. Do not preserve a known gap through
a compatibility shim.

## 22. Compatibility and migration policy

No route, wire, server, authorization rule, persistence, dependency, generated
public contract, stored data, Grid Adapter public API, or `GridHandle` change
is planned. Internal TypeScript contracts may break and must migrate atomically
without aliases or legacy import paths.

The intentional observable changes are:

- complete ordered-sort View-bar behavior;
- explicit saved-view unavailable, stale, invalid, and error feedback;
- corrected saved-view disclosure/dialog and menu focus semantics; and
- removal of redundant Rename and Manage Sharing actions.

Unavailable application commands consuming no event is a correctness rule,
not a compatibility burden. Valid keyboard shortcuts, header sorting, saved-
view routes/payloads, query/layout persistence, mutation/replay semantics,
paste route/wire/transactions, Timeline invariants, and passive consumer test
selectors remain compatible.

## 23. Rollback and visual policy

Rollback is source-only and workstream-granular. Each workstream's rollback
unit includes its production source, tests, authored ownership/catalog inputs,
Make-generated projections, guide/README updates, and tracker records. No data
or external-system rollback is required.

No visual golden change is expected. If an intentional visual delta is
unavoidable, use only the repository visual-update Make target, manually
review every image and manifest change, record the owner/design reason, and
require two subsequent ordinary visual runs. Revert the approved golden and
manifest with the originating CR-03, CR-04, CR-05, CR-08, or Timeline
workstream, not in an unrelated cleanup.

## 24. Residual risk and future extension seam

This plan deliberately leaves repo-wide static debt and adjacent large
workflows untouched. Its success is measured by removing parallel ownership
inside the named cone, not by reducing the repository-wide Fallow totals.
Future changes must not reopen a completed workstream merely to absorb an
unrelated finding.

After CR-12, the safe extension seam is:

1. one pure policy or discriminated model owned by the feature;
2. one narrow effect or runtime-driver port;
3. the unchanged public Grid Adapter contract when semantic grid behavior is
   sufficient;
4. one exact mutation-envelope member only when a genuinely new owner kind is
   required;
5. owner-local admission, lifecycle, projection, focus, and continuity; and
6. focused owner, negative-boundary, accessibility, browser, and static
   evidence.

Carry a feature forward only when it improves that seam. Remove behavior that
would require a second policy, runtime, transport, or compatibility path.

## 25. Initial planning checkpoint

| Timestamp | Status transition | Paths and decisions | Evidence, compatibility, rollback, risk, and next action |
| --- | --- | --- | --- |
| `2026-09-02` | Tracker created; CR-01 through CR-12 and CR-C01 through CR-C12 remain `PLANNED` | Added only `docs/handoffs/clean-up-refactor-handoff.md`. Preserved the completed Workbook Inspector handoff, Core owners, `docs/domain.md`, product source, tests, catalogs, generated artifacts, and goldens unchanged. Recorded Core 03/design authority, domain vocabulary/navigation scope, advisory research status, the clean baseline commit/upstream relation, retained Fallow root, eight gaps, strict workstream order, acceptance gates, validation matrix, compatibility, rollback, exclusions, and future seam. | PASS: `make lint-markdown` at `.cartulary/test-results/20260903T021248Z-p49249`; `git diff --check`; a no-index check of the new file produced no whitespace diagnostics; staged/unstaged scope inspection found only this untracked tracker and no change to the historical handoff. Formatting, generation, product tests, browser tests, and Fallow reruns were intentionally skipped because tracker creation changes no implementation, authored machine input, generated artifact, or golden. Compatibility impact is none. Rollback is deletion of this new tracker. Residual risk is baseline drift before execution, addressed by the mandatory refreshed pre-CR-01 checkpoint. Next action: obtain separate implementation authorization, then append that checkpoint and mark only CR-01 `IN_PROGRESS`. |
| `2026-09-02T23:13:30-04:00` | Implementation authorized; CR-01 `IN_PROGRESS`; CR-C01 and CR-02 remain `PLANNED` | Refreshed `AGENTS.md`, Core 03, design, domain vocabulary/navigation, branch `main`, commit `c3d5c449551e7b58ae49b61f57f745aa385b36fa`, upstream `origin/main` at `79d305265e433ef5d3eafb4750e319ec34fd5c3f` (`0 1`), and worktree scope. The only pre-existing change is this staged tracker. Toolchain pin digest is `d530190d057f26e08128967c3265d9b94aa75936dab9eec299d106a2de8541e9`; generated-artifact-policy digest is `ddc06647b89945dec1354cc4761d9ffbc9ddb79e14beb7e3e97b25ea83b374d5`. Allowed paths and exclusions remain sections 4 and 24. No owner contradiction was found during the refreshed authority read. | Authorization is the user's explicit request to implement the controlling plan. Retained orientation evidence remains `.cartulary/test-results/20260903T004815Z-p25087`; it is not accepted as fresh CR-01 evidence. Refreshed task guides for `web.workbook`, `web.architecture`, `web.design`, `package.grid_adapter`, `module.workbook`, `module.timeline`, and `module.entities` share catalog digest `cb4c8fef...` and routing digest `85b5a5e3...`; their focused and service-backed routes will be run as applicable. Compatibility impact is none. Rollback is removal of this appended execution checkpoint and restoration of CR-01 to `PLANNED` before product changes. Residual risk is characterization or Fallow drift. Next action: inventory consumers/assertions/findings and run passing CR-01 characterization. |
| `2026-09-02T23:26:35-04:00` | CR-01 and CR-C01 `DONE`; CR-02 remains `PLANNED` | Confirmed Core 03, design, and domain vocabulary owners are sufficient and conflict-free. Refreshed source sizes match section 5. Inventoried the global keyboard mapper; saved-view, grid-control, shell, runtime, Entity/Timeline paste, Timeline replay/load/mutation/model/renderer consumers; private exports; route adapters; source ownership; and test-family coverage. The scoped assertion ledger contains: boolean-select parsing in `WorkbookGridControls`; generic runtime metadata in `WorkbookMutationRuntime`; generated-request assertions in `createWorkbookMutationCommandPorts` and `workbookConflictResolutionAdapter`; replay admission in `useTimelinePendingReplayController`; synthetic paste event conversion in `useTimelineClipboardPasteController`; grid-anchor conversion in `useTimelineGridAnchorController`; committed-row indexing in `useTimelineCommittedRows`; field-binding, collection decoding, and collection-renderer downcasts in Timeline models/components; mutation reconciliation decoding; and mention decoding. `as const`, `satisfies`, typed CSS literals, generated source, and test-only DOM assertions remain explicit non-targets. | PASS: `web.workbook` `.cartulary/test-results/20260903T031411Z-p66379`; `web.architecture` `.cartulary/test-results/20260903T031421Z-p67152`; `web.design` `.cartulary/test-results/20260903T031421Z-p67172`; `package.grid_adapter` `.cartulary/test-results/20260903T031421Z-p67120`; `module.workbook` `.cartulary/test-results/20260903T031421Z-p67100`; `module.timeline` `.cartulary/test-results/20260903T031421Z-p67104`; `module.entities` `.cartulary/test-results/20260903T031421Z-p67230`; typecheck `.cartulary/test-results/20260903T031911Z-p24070`; service-backed Workbook `.cartulary/test-results/20260903T032002Z-p25502`, Timeline `.cartulary/test-results/20260903T032002Z-p25504`, and Entities `.cartulary/test-results/20260903T032444Z-p39413`. Fresh owner-only-umask Fallow passed at `.cartulary/test-results/20260903T031925Z-p24597`; direct JSON inspection reproduced the named critical/high findings and separately classified adjacent adapter, workflow, collaboration, query, and test findings as out of scope. No failure occurred. Compatibility impact is none. Rollback is the CR-01 tracker-only source unit. Residual risk is deliberate repository-wide static debt outside the cone. Next action: append the CR-02 checkpoint and establish the new keyboard/focus ownership. |
| `2026-09-02T23:26:35-04:00` | CR-02 `IN_PROGRESS`; CR-C02 remains `PLANNED` | Predecessor CR-01/CR-C01 confirmed `DONE`. Branch, commit, upstream relation, staged tracker, owner decisions, task-guide digests, allowed paths, and exclusions are unchanged. CR-02 is limited to Workbook application shortcuts, Timeline editor policy, registered overlay/menu focus, direct consumers, focused evidence, authored metadata, guide/source navigation, and this tracker. | Characterization baselines are the CR-01 routes above. Compatibility requires stable valid shortcuts while unavailable commands become non-consuming. Rollback is the complete CR-02 source/test/metadata/guide/tracker unit. Residual risks are copied grid semantics and focus regressions; closure scans and owner routes are mandatory. Next action: add pure policies and the registered-ref primitive, migrate consumers, then delete the global mapper. |
| `2026-09-03T00:11:43-04:00` | CR-02 and CR-C02 `DONE`; CR-03 remains `PLANNED` | Added `apps/web/src/workbook/policies/workbookApplicationShortcuts.ts` and its test plus `apps/web/src/workbook/focus/useRegisteredOverlayNavigation.ts` and its lifecycle test. Deleted `apps/web/src/workbook/utils/workbookKeyboard.ts` and its obsolete test. Migrated `timeline/models/timelineKeyboardIntentModel.ts`, `timeline/hooks/useTimelineKeyboardController.ts`, `timeline/hooks/useTimelineInspectorSelection.ts`, `timeline/components/TimelineRowActions.tsx`, `timeline/presentation/useTimelineWorkbookPresentation.tsx`, `components/SystemViewSwitcher.tsx`, `components/WorkbookGridControls.tsx`, `hooks/useIncidentControlsDrawer.ts`, `layout/WorkbookSurfaceLayout.tsx`, and `WorkbookShell.tsx`; updated their direct tests, `apps/web/src/README.md`, `tools/frontend_source_ownership.json`, the three affected test-family manifests, and `tools/test_catalog_row_migrations.json`. Make-owned generation changed only `tools/execution_topology_render_index.json` input hashes/digest; no public contract, generated product source, route fixture, persistence artifact, or golden changed. Workbook application shortcuts now cover only Quick Link, Evidence preview, History, and inspector close; Timeline owns editor save/navigation; Grid Adapter remains the only grid navigation/paste/copy/fill owner. Registered refs own menu/overlay opening, traversal, disabled-item skipping, Escape, surface invalidation, and trigger/fallback restoration without selectors, animation frames, or timers. | PASS focused rows: Workbook `.cartulary/test-results/20260903T035143Z-p22455`, Timeline `.cartulary/test-results/20260903T035143Z-p22463`, and combined web interaction `.cartulary/test-results/20260903T035143Z-p22477`. PASS full owners: `web.workbook` `.cartulary/test-results/20260903T035158Z-p24341`, `module.workbook` `.cartulary/test-results/20260903T035158Z-p24342`, `module.timeline` `.cartulary/test-results/20260903T035158Z-p24346`, `web.architecture` `.cartulary/test-results/20260903T035731Z-p61762`, `package.grid_adapter` `.cartulary/test-results/20260903T035731Z-p61765`, and `web.design` `.cartulary/test-results/20260903T035829Z-p8208`. PASS service-backed Workbook `.cartulary/test-results/20260903T035939Z-p55282`, Timeline `.cartulary/test-results/20260903T035939Z-p55293`, and design `.cartulary/test-results/20260903T035939Z-p55309`; format `.cartulary/test-results/20260903T035131Z-p18168`; typecheck `.cartulary/test-results/20260903T035013Z-p16505`; frontend unit `.cartulary/test-results/20260903T040816Z-p17582`; import boundary `.cartulary/test-results/20260903T040816Z-p17599`; Biome `.cartulary/test-results/20260903T040816Z-p17626`; Markdown `.cartulary/test-results/20260903T040950Z-p36723`; generate `.cartulary/test-results/20260903T041051Z-p38369`; drift `.cartulary/test-results/20260903T041116Z-p41505`; JSON shape `.cartulary/test-results/20260903T041116Z-p41527`; generated policy `.cartulary/test-results/20260903T041116Z-p41531`. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T040909Z-p35691` was inspected directly: 915 files, 16,342 functions, 720 critical, 320 high, and 486 moderate; no new scoped high/critical finding, with only a moderate decision-table complexity note. Introduced failures were retained and resolved: initial type errors `.cartulary/test-results/20260903T033628Z-p93459`, missing test cleanup `.cartulary/test-results/20260903T034040Z-p98361`, focus-fallback assertions `.cartulary/test-results/20260903T034426Z-p3495`, `.cartulary/test-results/20260903T034455Z-p4398`, and `.cartulary/test-results/20260903T034558Z-p5261`, misplaced hook/dependency lint findings `.cartulary/test-results/20260903T034858Z-p7400`, and expected stale topology input `.cartulary/test-results/20260903T040950Z-p36581`. Closure scans found no mapper/import/alias, application grid/paste/fill policy, or selector/timer/frame focus in the migrated menu/overlay paths; `git diff --check` passed. Compatibility preserves shortcut spellings/modifiers and all Grid/public APIs; unavailable commands intentionally consume nothing, and fallback focus is corrected. Rollback is the complete listed CR-02 source/test/metadata/generated-index/guide unit. Residual scoped risk is the non-blocking moderate policy finding; unrelated future-workstream timer/focus paths remain assigned to their planned owners. Next action: append the CR-03 pre-workstream checkpoint and characterize saved-view resources/actions before replacement. |
| `2026-09-03T00:11:43-04:00` | CR-03 `IN_PROGRESS`; CR-C03 remains `PLANNED` | Predecessor CR-02/CR-C02 confirmed `DONE`. Refreshed owner/task-guide routing, branch/upstream scope, changed paths, generated-artifact policy, Core 03 saved-view behavior, design direction, and the CR-01 saved-view inventory remain controlling. CR-03 is limited to the active-surface saved-view resource/reducer/action identity, disclosure/dialog presentation, retained actions, removed Rename/Manage Sharing paths and selectors, direct tests, metadata/guide updates, and this tracker. | Characterization starts from the passing CR-02 owner runs above. Compatibility preserves routes, payloads, normalized query/layout capture, system immutability, passive retained selectors, and server authorization; Rename and Manage Sharing are intentionally removed in favor of one Update view flow. Rollback is the complete CR-03 source/test/metadata/guide/tracker unit. Residual risks are stale cross-surface completion and loading/deletion ambiguity. Next action: freeze request fixtures and inventory every saved-view state/action/selector consumer before extracting the resource and action machine. |
| `2026-09-03T00:56:37-04:00` | CR-03 and CR-C03 `DONE`; CR-04 remains `PLANNED` | Core 03 REQ-03-012 through REQ-03-026 and design section 8.3 remain sufficient and conflict-free. Added `models/workbookSavedViewControl.ts` with explicit loading/ready/unavailable/invalid-selection resources, active-surface projection, exact scope parsing, a surface-keyed reducer, and complete subject identity; added `hooks/useActiveSurfaceSavedViewActions.ts` for one-action admission, last-moment surface/selection/version/owner/role validation, and stale-completion rejection; added focused pure/component tests. Split `ActiveSurfaceSavedViewSelector.tsx`, added `SavedViewActionPanel.tsx`, and migrated `useWorkbookSavedViewController.ts`, `useWorkbookShellRuntime.ts`, and `WorkbookShell.tsx`. The action surface is now an anchored labelled dialog with registered opening/Escape/restoration focus; create, update, duplicate, reset, delete, home, and default remain while Rename and Manage Sharing plus their UI-contract selectors were deleted. Updated direct tests, `docs/guides/cartulary-dev-guide.md`, source ownership, `web.workbook` test-family rows, selector tests, and the Make-generated topology render-index hashes. Routes, request bodies, normalized query/layout capture, authorization, system immutability, passive selectors, public contracts, persistence, and goldens are unchanged. During browser validation, corrected the CR-02 registered-focus primitive so a programmatically focused overlay root remains inside its own focus boundary and trigger restoration occurs after close; added exact lifecycle evidence and `tabIndex=-1` to registered overlay roots. | PASS focused saved-view/model/controller/surface/selector/focus rows: `.cartulary/test-results/20260903T043236Z-p13555`, `.cartulary/test-results/20260903T043257Z-p14649`, `.cartulary/test-results/20260903T043337Z-p19517`, `.cartulary/test-results/20260903T043250Z-p14113`, `.cartulary/test-results/20260903T043447Z-p25031`, `.cartulary/test-results/20260903T044310Z-p66950`, and post-decomposition `.cartulary/test-results/20260903T045533Z-p92748`. PASS full owners: `web.workbook` `.cartulary/test-results/20260903T043511Z-p25734`, `web.architecture` `.cartulary/test-results/20260903T043511Z-p25736`, `web.design` `.cartulary/test-results/20260903T043511Z-p25743`, `package.ui` `.cartulary/test-results/20260903T043511Z-p25764`, `module.workbook` `.cartulary/test-results/20260903T043650Z-p91915`, and browser-inclusive `module.savedviews` 25/25 `.cartulary/test-results/20260903T044319Z-p67494`. PASS service-backed Saved Views `.cartulary/test-results/20260903T044435Z-p19059`, Workbook `.cartulary/test-results/20260903T044545Z-p68850`, and design `.cartulary/test-results/20260903T044755Z-p25051`; generate `.cartulary/test-results/20260903T044900Z-p71907`; generation drift, generated policy, JSON shape, import boundary, Biome, and Markdown `.cartulary/test-results/20260903T044914Z-p74974`, `p74995`, `p75012`, `p75469`, `p75556`, and `p75631`; final format `.cartulary/test-results/20260903T045509Z-p88095`; final typecheck/Biome `.cartulary/test-results/20260903T045541Z-p93670` and `.cartulary/test-results/20260903T045551Z-p94181`. Fresh owner-only-umask Fallow passed at `.cartulary/test-results/20260903T045556Z-p94673`; direct JSON inspection found 920 files and no CR-03 high/critical finding after responsibility extraction, with only moderate CRAP/coverage notes. Introduced failures were retained and resolved: formatter dependency lint `.cartulary/test-results/20260903T042419Z-p48846`; type errors `.cartulary/test-results/20260903T042533Z-p57645` and `.cartulary/test-results/20260903T044914Z-p75438`; source-ownership failures `.cartulary/test-results/20260903T042625Z-p58889` and `.cartulary/test-results/20260903T042847Z-p2689`; test cleanup/matcher failure `.cartulary/test-results/20260903T043153Z-p8372`; stale topology `.cartulary/test-results/20260903T043409Z-p21063`; browser focus failures `.cartulary/test-results/20260903T043650Z-p91917` and `.cartulary/test-results/20260903T044008Z-p6245`; incomplete focus test harness `.cartulary/test-results/20260903T044242Z-p62123`; and operator-error row selections `.cartulary/test-results/20260903T043920Z-p1042` and `.cartulary/test-results/20260903T045525Z-p92013`. Closure scans found no production Rename/Manage Sharing selector/action or selector/timer/frame focus in the saved-view boundary; `git diff --check` passed. Compatibility impact is limited to intentional redundant-action removal and explicit unavailable/invalid/error feedback. Rollback is the listed CR-03 source/test/metadata/generated-index/guide/tracker unit plus the registered-focus corrective unit; no data rollback is needed. Residual risk is limited to non-blocking moderate static findings and repo-wide excluded debt. Next action: append the CR-04 checkpoint and characterize grid-control serialization, interaction, focus, and responsive behavior. |
| `2026-09-03T00:56:37-04:00` | CR-04 `IN_PROGRESS`; CR-C04 remains `PLANNED` | Predecessor CR-03/CR-C03 confirmed `DONE`. Refreshed Core 03 query-control clauses, design direction, task-guide routing, worktree scope, generated-artifact policy, and the CR-01 grid-control inventory remain controlling. CR-04 is limited to pure query projection and surface-keyed transient state, semantic command descriptors, private Sort/Group/Filters/Columns/Chips units, complete ordered sorting, exact controlled-value parsing, registered focus, direct tests/evidence, authored metadata/guide updates, and this tracker. | Characterization starts from the passing CR-03 owner and browser routes above. Compatibility preserves persisted query/layout shapes, canonical chip order, header sorting, all-columns-hidden behavior, Grid Adapter APIs, routes, and public contracts; the View-bar gains the owner-required full ordered-sort lifecycle. Rollback is the complete CR-04 source/test/metadata/generated-index/guide/tracker unit. Residual risks are accidental serialization drift, duplicate sort admission, invalid draft submission, responsive overflow, and focus restoration. Next action: inventory the current monolith and existing model/contract tests, freeze canonical serialization, then introduce the pure projection/reducer/descriptor boundary before splitting presentation. |
| `2026-09-03T01:42:44-04:00` | CR-04 and CR-C04 `DONE`; CR-05 remains `PLANNED` | Core 03 ordered-query requirements and design sections 7.5/8.3 remain sufficient and conflict-free. Replaced the 1,060-line control monolith with `models/workbookGridQueryControls.ts`, a pure active-contract/query/layout/responsive projection, closed semantic command descriptors, exact field/group/boolean parsers, a surface-keyed transient reducer, and explicit ordered-sort add/direction/move/remove policy with duplicate, boundary, and eight-entry fail-closed no-ops. Added separate `WorkbookSortControl.tsx`, `WorkbookGroupControl.tsx`, `WorkbookFiltersControl.tsx`, `WorkbookColumnsControl.tsx`, and `WorkbookActiveQueryChips.tsx` units plus shared local styles and focused model/component tests. `WorkbookGridControls.tsx` is now the active-surface composition and one command dispatcher. Invalid filter drafts remain visible with local feedback and cannot invoke apply; group/filter select values require exact declared keys. Canonical chips remain group, applied sorts, normalized filters with unchanged responsive capacities. Header sorting and all Grid Adapter APIs remain separate and unchanged. Columns remain operable when every data column is hidden. Updated `apps/web/src/README.md`, the frontend guide, source ownership, `web.workbook` test-family metadata, and the Make-generated topology render-index digest/hashes only. No route, wire, saved-query/layout serialization, public contract, persistence, dependency, or golden changed. The CR-02 focus primitive gained unmount-time fallback restoration after full-owner testing exposed the Timeline context-menu lifecycle; its key target calculation was then extracted into pure decisions to remove a fresh high static finding. | PASS focused grid projection/component/existing-query rows `.cartulary/test-results/20260903T051325Z-p33466`; focused registered-focus and Timeline unmount restoration `.cartulary/test-results/20260903T051616Z-p56012`; full `web.workbook` 154/154 `.cartulary/test-results/20260903T051642Z-p56751`; `web.architecture` 12/12 `.cartulary/test-results/20260903T051831Z-p82023`; browser-inclusive `web.design` 15/15 `.cartulary/test-results/20260903T051836Z-p82591`; `package.grid_adapter` 43/43 `.cartulary/test-results/20260903T051942Z-p31432`; `module.workbook` 68/68 `.cartulary/test-results/20260903T052243Z-p68468`. PASS service-backed design `.cartulary/test-results/20260903T052041Z-p76300`, Grid Adapter `.cartulary/test-results/20260903T052143Z-p23731`, and Workbook `.cartulary/test-results/20260903T052456Z-p27438`. Explicit browser PASS: accessibility 12/12 `.cartulary/test-results/20260903T052957Z-p6759`, responsive measurement 22/22 `.cartulary/test-results/20260903T053122Z-p51931`, stateful 34/34 `.cartulary/test-results/20260903T053549Z-p10324`, visual 12/12 `.cartulary/test-results/20260903T053807Z-p60438`. PASS frontend unit 413/413 `.cartulary/test-results/20260903T052719Z-p83973`; Markdown `.cartulary/test-results/20260903T052816Z-p4418`; final format/generate/drift/generated-policy/JSON/typecheck/import/Biome `.cartulary/test-results/20260903T054005Z-p5939`, `20260903T054009Z-p10043`, `20260903T054018Z-p12917`, `20260903T054027Z-p15949`, `20260903T054028Z-p16360`, `20260903T054031Z-p16834`, `20260903T054041Z-p17336`, and `20260903T054044Z-p17757`. Final fresh owner-only-umask Fallow `.cartulary/test-results/20260903T054218Z-p25066` was inspected directly: 929 files, 16,540 functions, 719 critical, 318 high, 490 moderate repository-wide; the CR-04/registered-focus cone has no high/critical finding and two non-blocking moderate coverage notes. Introduced failures were retained and resolved: readonly test typing `.cartulary/test-results/20260903T050630Z-p4294`; invalid/unsorted authored row IDs rejected by Make preflight without a run root; expected stale-topology JSON `.cartulary/test-results/20260903T050832Z-p10813`; enabled-item End-focus assertion `.cartulary/test-results/20260903T050856Z-p11979`; missing component-test cleanup `.cartulary/test-results/20260903T051204Z-p22976`; Timeline unmount restoration `.cartulary/test-results/20260903T051331Z-p33908` and focused reproduction `.cartulary/test-results/20260903T051429Z-p50489`; source ownership ordering `.cartulary/test-results/20260903T051801Z-p80170`; and the passing `.cartulary/test-results/20260903T054059Z-p18296` Fallow run whose scoped high focus-handler finding required and received structural extraction before final acceptance. Closure scans found no closure-bearing chip model, select assertion, selector/timer/frame focus, copied header-sort policy, new TODO/FIXME, or misplaced Grid Adapter vendor import; `git diff --check` passed and Git scope showed no generated product/public contract/golden change. Compatibility preserves query/layout persistence, header semantics, active-chip ordering/overflow, all-columns-hidden recovery, routes, callbacks, and public APIs; the intended addition is the complete View-bar ordered-sort lifecycle and explicit invalid-draft feedback. Rollback is the complete listed CR-04 source/test/metadata/render-index/guide/README/tracker unit plus the focus corrective unit. Residual risk is limited to the two moderate static notes and excluded repo-wide debt. Next action: append the CR-05 checkpoint and characterize shell construction/lifetimes before extracting composition units. |
| `2026-09-03T01:42:44-04:00` | CR-05 `IN_PROGRESS`; CR-C05 remains `PLANNED` | Predecessor CR-04/CR-C04 confirmed `DONE`. Refreshed repository scope, owners, task-guide routes, generated-artifact policy, shell source/consumer inventory, and CR-01 construction/lifecycle baselines remain controlling. CR-05 is limited to private shell infrastructure/runtime construction, startup/saved-view/layout/query state, active-surface registration, extension availability/renderer lifecycle, recovery/conflict focus, top-bar navigation, active-surface presentation, direct tests, metadata/guides, and this tracker. | Characterization starts from the passing CR-04 owner, service-backed, browser, and static evidence above. Compatibility preserves `WorkbookShell` props/route boundary, registry and runtime identities, lazy imports, navigation, responsive ordering, recovery focus, passive selectors, and public exports. Rollback is the complete CR-05 source/test/metadata/generated-index/guide/README/tracker unit. Residual risks are duplicate construction, cleanup/remount drift, render-time discovery mutation, stale extension availability, and a replacement mega-controller. Next action: run refreshed shell owner guides, inventory construction/effect sites and existing lifecycle tests, then move render-time discovery into an exact lifecycle owner before incremental composition extraction. |
| `2026-09-03T02:22:25-04:00` | CR-05 and CR-C05 `DONE`; CR-06 remains `PLANNED` | Reduced `WorkbookShell.tsx` from 1,273 to 427 lines of route-facing composition. Added exact owners for adapter/runtime construction and reference-broker disposal (`useWorkbookShellInfrastructure`), startup/saved/query state (the retained `useWorkbookShellRuntime`), surface query loading/projection (`useWorkbookSurfaceQueries`), collaboration invalidation/registration (`useWorkbookCollaborationLifecycle`), authorization recovery, extension controller/discovery/fallback lifecycle, recovery focus, top-bar navigation, view-bar controls, incident-controls lazy presentation, active built-in/extension renderer selection, and the recovery frame. `ExtensionAvailabilityController.setDiscovery` now reports whether discovery changed, and only `useWorkbookExtensionAvailability` invokes it from an effect; a direct lifecycle test proves no render-time invocation, stable controller identity for one subject, no revision for equal discovery, and revision on claim loss. Updated source ownership, `web.workbook` catalog input, local README, frontend guide, and Make-generated topology input hashes. No public barrel, route, wire, persistence, authorization, dependency, Grid Adapter contract, selector, generated product/public contract, or golden changed. | PASS final focused extension lifecycle `.cartulary/test-results/20260903T061821Z-p40613`; full `web.workbook` 155/155 `.cartulary/test-results/20260903T061857Z-p43974`; `web.architecture` 12/12 `.cartulary/test-results/20260903T061821Z-p40622`; design 15/15 `.cartulary/test-results/20260903T061857Z-p43982`; Network Flow 38/38 `.cartulary/test-results/20260903T060401Z-p83760`; module Workbook 68/68 and service-backed 39/39 `.cartulary/test-results/20260903T060558Z-p54717` and `p54775`; service-backed design 15/15 `p54749`; frontend unit 414/414 `p54947`; stateful browser 34/34 `.cartulary/test-results/20260903T060901Z-p31792`; webserver-backed 60/60 `p31820`; final accessibility 12/12 `.cartulary/test-results/20260903T061857Z-p44421`; final visual 12/12 `p44278`. PASS final format/typecheck `.cartulary/test-results/20260903T061801Z-p35881` and `20260903T061805Z-p40031`; generate `.cartulary/test-results/20260903T060320Z-p75916`; final drift/generated-policy/JSON/import/Biome/Markdown `.cartulary/test-results/20260903T062145Z-p925`, `20260903T062154Z-p4044`, `20260903T062155Z-p4455`, `20260903T062158Z-p4916`, `20260903T062202Z-p5313`, and `20260903T062203Z-p5758`. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T061829Z-p43049` was inspected directly: 942 files, 16,572 functions, 718 critical, 318 high, and 491 moderate repository-wide; the extracted CR-05 production cone has no critical/high finding and only moderate top-bar/recovery coverage-complexity notes. Introduced failures were retained and resolved: missing source ownership `.cartulary/test-results/20260903T055615Z-p52610`, expected stale topology JSON `.cartulary/test-results/20260903T060214Z-p73901`, and the test-only implicit-`this` type error `.cartulary/test-results/20260903T061701Z-p30786`. Closure scans prove one controller/coordinator/runtime/registry construction site, effect-only discovery, no lazy renderer or facade selection in the route component, no new public/barrel export, no selector/timer/frame focus, no TODO/FIXME, no generated/golden delta, and clean Git whitespace. Compatibility is unchanged. Rollback is the complete listed CR-05 source/test/metadata/render-index/README/guide/tracker unit; no data rollback is needed. Residual risk is limited to the two moderate static notes and excluded repository debt. Next action: append the CR-06 checkpoint and freeze runtime queue/status/conflict/retry/auth/refresh/transaction behavior before structural extraction. |
| `2026-09-03T02:22:25-04:00` | CR-06 `IN_PROGRESS`; CR-C06 remains `PLANNED` | Predecessor CR-05/CR-C05 confirmed `DONE`. Refreshed runtime and registry construction sites, task-guide routing, current owner evidence, generated-artifact policy, CR-01 assertion ledger, and changed-path scope remain controlling. CR-06 is limited to the one shell-lifetime `WorkbookMutationRuntime`, injected clock/scheduler ports, private status/conflict/surface/managed-patch/retry/transaction/lifecycle responsibilities, deterministic tests, direct consumers that must remain stable, metadata/guides, and this tracker. The closed envelope/driver seam may be staged but generic Timeline replay removal remains CR-07. | Compatibility requires unchanged FIFO admission, coalescing/capacity, retry/auth gates, conflict/save-state/refresh-debt/transaction behavior, exact shell/runtime lifetime, and no parallel runtime or facade. Rollback is the complete CR-06 source/test/metadata/generated-index/guide/tracker unit. Residual risks are timer nondeterminism, registration replacement, lost paused work, and state split across old/new authorities. Next action: refresh Workbook/architecture task guides, characterize the runtime public methods and tests, then introduce injected time/scheduling behind the existing facade before extracting one responsibility at a time. |
| `2026-09-03T02:50:36-04:00` | CR-06 and CR-C06 `DONE`; CR-07 remains `PLANNED` | Branch `main` remains at `c3d5c449551e7b58ae49b61f57f745aa385b36fa`, `0 1` relative to `origin/main`; implementation authorization and the cumulative unstaged/staged worktree remain unchanged. Reduced `WorkbookMutationRuntime.ts` from 981 to 500 lines and added private `WorkbookClientTransactionLedger`, `WorkbookConflictStore`, `WorkbookManagedPatchDriver`, `WorkbookMutationDriverRegistry`, `WorkbookRetryScheduler`, `WorkbookRuntimeLifecycle`, `WorkbookSurfaceRegistry`, pure status projector, and clock/scheduler ports. The one existing runtime now owns FIFO coordination while each responsibility owns one state boundary. The closed envelope admits only `managed_patch` and `timeline_row`; exact registration returns structured duplicate rejection without replacement, and a claimed unit remains queued when its driver is absent. Managed patches use injected time/retry scheduling and release owner claims only on terminal removal. Simplified the status and collection-action helpers after direct static inspection. Added deterministic responsibility/driver tests and updated source ownership, `web.workbook` catalog input, local README, frontend guide, and only the Make-generated execution-topology hashes. No production generated/public contract, route, wire, persistence, authorization, dependency, Grid Adapter API, or golden changed. | PASS focused runtime rows `.cartulary/test-results/20260903T064836Z-p8561`; final `web.workbook` 156/156 `.cartulary/test-results/20260903T064944Z-p11254`; `web.architecture` 12/12 `.cartulary/test-results/20260903T064203Z-p44019`; module Workbook 68/68 and service-backed 39/39 `.cartulary/test-results/20260903T064203Z-p44040` and `p44069`; frontend unit 415/415 `.cartulary/test-results/20260903T064511Z-p79662`; final format/typecheck/Biome `.cartulary/test-results/20260903T064808Z-p4095`, `20260903T064836Z-p8665`, and `p8693`; generate `.cartulary/test-results/20260903T064136Z-p40846`; drift/JSON/import/generated-policy/Markdown `.cartulary/test-results/20260903T064511Z-p79402`, `p79448`, `p79643`, `20260903T064944Z-p11196`, and `p11414`. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T064849Z-p10106` was inspected directly: 952 files, 16,669 functions, 717 critical, 315 high, and 493 moderate repository-wide; the runtime cone has no high/critical finding and only moderate managed-drain and primary-status coverage/complexity notes. Introduced failures/findings were retained and resolved: unused facade field typecheck `.cartulary/test-results/20260903T063104Z-p15369` and the initial passing Fallow run `.cartulary/test-results/20260903T064608Z-p2706` whose new status-projector high plus pre-existing collection-helper high prompted structural simplification. Closure scans prove one production runtime factory, one registry, injected browser time only at the composition port, no parallel runtime/state, no new TODO/FIXME, clean Git whitespace, and no generated product/golden delta. Generic `pending<TMeta>()`, `registerDrainer`, and Timeline replay timers remain explicitly deferred to CR-07 and are not compatibility commitments. Compatibility is unchanged. Rollback is the complete CR-06 source/test/metadata/render-index/README/guide/tracker unit; no data rollback is needed. Residual risk is limited to the two moderate findings and the temporary CR-07 migration seam. Next action: append the CR-07 checkpoint, install the exact Timeline driver, migrate replay settlement/scheduling, and delete the generic/drainer paths atomically. |
| `2026-09-03T02:50:36-04:00` | CR-07 `IN_PROGRESS`; CR-C07 remains `PLANNED` | Predecessor CR-06/CR-C06 confirmed `DONE`. Refreshed the closed driver/envelope seam, Timeline pending-save/replay consumers, Workbook/Timeline task guides, current worktree scope, generated-artifact policy, and CR-01 generic-metadata ledger. CR-07 is limited to one Timeline owner driver with pure admission/revalidation/settlement/conflict/discard/accepted-result plans, common-runtime scheduling and transport orchestration, direct consumers/tests, deletion of the old replay controller and generic/drainer/replay-bag paths, metadata/guides, and this tracker. | Compatibility requires high-water versions, pending signatures, conflicts/drafts, created-row pinning, mentions/notices, refresh debt, focus/viewport continuity, queue order/auth/retry/capacity, and socket settlement. Rollback is the complete CR-07 source/test/metadata/generated-index/guide/tracker unit. Residual risks are wrong-owner dispatch, stale row revalidation, duplicated schedulers, apply failures, and remount gaps. Next action: characterize every old controller branch and binding lifecycle, then move admission metadata into the Timeline driver before switching the common runtime to sole dispatch scheduling. |
| `2026-09-03T04:32:10-04:00` | CR-07 and CR-C07 `DONE`; CR-08 remains `PLANNED` | Branch `main` remains at `c3d5c449551e7b58ae49b61f57f745aa385b36fa`, `0 1` relative to `origin/main`; implementation authorization and the cumulative staged/unstaged worktree remain unchanged. Replaced and deleted `timeline/hooks/useTimelinePendingReplayController.ts` with the exact `useTimelineMutationDriver.ts`; added pure `timelineMutationDriverPlans.ts` and tests plus shell-runtime-keyed Timeline-owned state in `timelinePendingSaves.ts`. The common runtime is now the sole FIFO scheduler, retry owner, pending transport caller, and socket transaction ledger; Timeline owns exact admission, current-row/version revalidation, accepted/rejected/conflict/discard plans, row projection, and replay context. The runtime exposes no generic `pending<TMeta>()`, replaceable drainer, owner cast, shared replay bag, or replay timer. Exact driver registration rejects duplicates without replacement, missing/unmounted drivers pause work, and Timeline replay context/order/signatures survive authorization-driven surface unmount and remount. Updated runtime/composition/hooks/tests, source ownership, Workbook/Timeline test-family inputs, local README, frontend guide, and the Make-generated execution-topology render-index hashes only. Strengthened grouped-edit evidence to require its owner-mandated post-mutation query refresh and made blank-row duplicate admission tests scheduler-independent. Corrected committed-row projection to restore both normalized values and raw scalar cells at a known committed version. No route, wire body, persistence, authorization policy, dependency, generated public contract, Grid Adapter API, golden, or valid user-visible behavior changed. | PASS final focused driver plans `.cartulary/test-results/20260903T083106Z-p89250`; `web.architecture` 12/12 `.cartulary/test-results/20260903T082333Z-p51882`; `web.workbook` 157/157 `.cartulary/test-results/20260903T082338Z-p53379`; Timeline 54/54 `.cartulary/test-results/20260903T080129Z-p94642`; Workbook 68/68 `.cartulary/test-results/20260903T081347Z-p68828`; Collaboration 31/31 `.cartulary/test-results/20260903T082422Z-p70203`; service-backed Timeline 30/30 `.cartulary/test-results/20260903T081601Z-p27968`, Workbook 39/39 `.cartulary/test-results/20260903T082036Z-p86860`, and Collaboration 22/22 `.cartulary/test-results/20260903T082558Z-p21867`; revoked-session remount replay 11/11 `.cartulary/test-results/20260903T081240Z-p24733`. PASS generate `.cartulary/test-results/20260903T082257Z-p43585`; final format/typecheck/Biome/import/drift/JSON `.cartulary/test-results/20260903T083052Z-p84665`, `20260903T083056Z-p88833`, `20260903T083115Z-p89864`, `20260903T083117Z-p90315`, `20260903T083120Z-p90645`, and `20260903T083128Z-p93667`; generated policy `.cartulary/test-results/20260903T082315Z-p49555`; Markdown `.cartulary/test-results/20260903T082752Z-p77054`. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T083132Z-p94137` was inspected directly: 955 files, 16,684 functions, 716 critical, 315 high, and 495 moderate repository-wide; the new runtime/driver cone has no high/critical finding and four moderate complexity/coverage notes. The two pre-existing critical reconciliation/projection findings in `useTimelineRowMutationCoordinator.ts` remain explicitly assigned to CR-10. Introduced implementation/test failures were retained and resolved: early type/format runs `20260903T065810Z-p37022`, `20260903T065903Z-p37917`, and `20260903T070726Z-p67140`; stale-row/callback diagnostic runs `20260903T070019Z-p43542`, `20260903T070139Z-p60657`, `20260903T070242Z-p61513`, `20260903T070357Z-p63053`, `20260903T071138Z-p78670`, `20260903T071328Z-p80253`, `20260903T071409Z-p81077`, `20260903T071434Z-p81786`, `20260903T071458Z-p82496`, `20260903T071549Z-p87495`, `20260903T071729Z-p93280`, `20260903T072032Z-p99569`, `20260903T072144Z-p5679`, `20260903T072222Z-p6421`, `20260903T072320Z-p7460`, and `20260903T072345Z-p8159`; unused-type/stale-catalog/full-suite runs `20260903T072022Z-p99099`, `20260903T073055Z-p26601`, `20260903T073058Z-p27014`, and `20260903T073210Z-p32726`; source-ownership ordering `20260903T074455Z-p99927`; Timeline owner timing assertions `20260903T074714Z-p62078`, `20260903T075228Z-p24709`, `20260903T075259Z-p26016`, and `20260903T075606Z-p32473`; the real remount defect `20260903T074501Z-p1911` and `20260903T080612Z-p56846`; and final mechanical rename type failures `20260903T081201Z-p23160` and `20260903T083021Z-p83959`. The wrong-umask Fallow invocation `20260903T073617Z-p60916` was procedural; subsequent static runs `20260903T073644Z-p62142`, `20260903T074053Z-p69489`, and `20260903T074219Z-p75999` exposed introduced replay complexity that was structurally reduced before final acceptance. Closure scans find no old controller, generic pending API, drainer, shared replay bag, replay timer, direct Timeline transport, duplicate runtime, new TODO/FIXME, misplaced Grid Adapter vendor import, generated product/golden delta, or Git whitespace error. Compatibility remains source-only and behavior-preserving. Rollback is the complete CR-07 runtime/Timeline source, tests, ownership/catalog, generated-index, README/guide, and tracker unit; no data rollback is required. Residual risk is limited to moderate driver notes and the explicitly deferred CR-10 hotspots. Next action: append the CR-08 checkpoint, refresh all three paste-owner paths and exact wire fixtures, then install one private Workbook paste transport before deleting the duplicate Entity/Timeline ports and synthetic bridge. |
| `2026-09-03T04:34:39-04:00` | CR-08 `IN_PROGRESS`; CR-C08 remains `PLANNED` | Predecessor CR-07/CR-C07 confirmed `DONE`. Refreshed Core 03 REQ-03-145 through REQ-03-152 and REQ-03-308, the closed/read-only interaction rule, design keyboard/paste direction, domain owner navigation, current Generic/Entity/Timeline controllers, Grid Adapter semantic paste boundary, route adapters, exact fixtures, source ownership, task-guide routes, generated policy, Git scope, and the CR-01 assertion ledger. The adopted owners remain sufficient and conflict-free. CR-08 is limited to one private Workbook clipboard-paste port/adapter, owner-local last-moment surface/mode/field/record/version/group/create revalidation and projection, native-editor/grid separation, exact tests, deletion of duplicate ports/transports/synthetic bridges, metadata/guides, and this tracker. | Characterization begins from the complete green CR-07 evidence above. Compatibility preserves `/api/v1/incidents/{incident_id}/workbook/clipboard-paste`, generated request/response vocabulary, transaction semantics, parser behavior, Entity `entity_origin`, Timeline mixed create/record batches, Generic scalar behavior, conflicts, focus, viewport continuity, and one post-paste refresh. No public Grid Adapter or generated contract change is allowed. Rollback is the complete CR-08 source/test/metadata/generated-index/README/guide/tracker unit. Residual risks are stale/wrong-surface target dispatch, duplicate transport drift, editor/grid event bridging, partial application, and focus/viewport discontinuity. Next action: freeze exact request/result fixtures and the three owner admission matrices, then introduce the common private transport before migrating Generic, Entity, and Timeline in that order. |
| `2026-09-03T06:16:46-04:00` | CR-08 and CR-C08 `DONE`; CR-09 remains `PLANNED` | Added the exact private adapter-layer `WorkbookClipboardPastePort`, sole `createWorkbookClipboardPasteAdapter`, bounded generated-vocabulary constructors, exact request/response and malformed-boundary tests, pure Timeline and Entity admission/target plans, an Entity execution controller, and split Timeline admission/request/conflict/execution helpers. Shell infrastructure constructs and injects one shared adapter. Generic scalar paste remains on its existing patch path; Entity scalar paste uses the existing edit command while rectangular paste intentionally sends all-create targets after proving the source records current, preserving exact `entity_origin` reuse; Timeline revalidates the active Timeline surface, writable bindings, authority, grouping/create capability, stable targets, and committed versions before dispatch, then locally decodes rows/conflicts, tracks the socket transaction, applies once, refreshes once, and restores semantic focus/viewport continuity. Native Timeline editor paste updates the controlled draft from the browser selection and invokes the scalar save command directly. Deleted `TimelineClipboardPastePort`, `createTimelineClipboardPasteAdapter`, `EntityMutationCommandPort.pasteCreate`, their construction/mocks/tests, and the synthetic editor-to-grid bridge. Updated shell/surface/Timeline composition, fixture and browser helpers, source ownership, Workbook/Timeline/Entities catalog inputs, local README, frontend guide, and only Make-generated topology input hashes. The pre-CR-08 checkpoint's `/workbook/clipboard-paste` route text was an error: the unchanged canonical route is `/api/v1/incidents/{incident_id}/views/{view_schema_id}/clipboard-paste`, proven by the exact adapter fixture and adopted OpenAPI operation. No owner contradiction, route, wire body, persistence, authorization policy, dependency, generated public contract, Grid Adapter API, golden, or valid user-visible behavior changed. | PASS exact adapter/model and owner plans `.cartulary/test-results/20260903T091357Z-p528`, `20260903T094343Z-p38514`, and `20260903T094348Z-p39052`; frontend unit 420/420 `.cartulary/test-results/20260903T094558Z-p53043`; final Timeline 55/55 `.cartulary/test-results/20260903T094738Z-p92659`; Workbook 68/68 `.cartulary/test-results/20260903T095210Z-p53742`; `web.workbook` 159/159 `.cartulary/test-results/20260903T100353Z-p71723`; architecture 12/12 `.cartulary/test-results/20260903T095940Z-p82173`; design 15/15 `.cartulary/test-results/20260903T100356Z-p72136`; explicit service-backed Workbook 39/39 `.cartulary/test-results/20260903T100617Z-p68091`, Timeline 30/30 `.cartulary/test-results/20260903T100822Z-p24804`, Entities 33/33 `.cartulary/test-results/20260903T101257Z-p85177`, Grid Adapter 13/13 `.cartulary/test-results/20260903T100255Z-p26645`, and design 15/15 `.cartulary/test-results/20260903T100458Z-p20061`. The first complete Entities route `.cartulary/test-results/20260903T095415Z-p10692` passed 42/43 and failed only a four-row unit at object-store readiness with harness classification `infra/service_readiness_timeout`; the exact failed service-backed rows then passed 3/3 at `.cartulary/test-results/20260903T095848Z-p64248`, and the final full service-backed route passed. The Grid Adapter focused route `.cartulary/test-results/20260903T095945Z-p82636` passed 41/43 and likewise failed only pre-browser object-store readiness; its service-backed rerun passed completely. PASS generate/drift/generated-policy/JSON/import/Biome `.cartulary/test-results/20260903T094525Z-p45306`, `20260903T094534Z-p48194`, `20260903T094543Z-p51278`, `20260903T094544Z-p51689`, `20260903T094547Z-p52166`, and `20260903T094550Z-p52563`; final format/typecheck/Markdown `.cartulary/test-results/20260903T101504Z-p37786`, `20260903T101508Z-p41953`, and `20260903T101518Z-p42475`; Git whitespace is clean. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T094355Z-p39644` was inspected directly: 963 files, 16,748 functions, 715 critical, 314 high, and 494 moderate repository-wide; the named/replacement paste cone has no high/critical finding or suppression, after the earlier `.cartulary/test-results/20260903T093312Z-p25096` exposed and prompted structural removal of critical Entity/Timeline paste closures. Remaining scoped results are four moderate coverage/complexity notes: Entity cell presentation, pure Entity batch planning, and two Timeline editor helpers. Introduced implementation/test failures were retained and resolved: format/typecheck/catalog and focused runs `20260903T085152Z-p4611`, `20260903T085226Z-p13471`, `20260903T090014Z-p27326`, `20260903T090146Z-p32621`, `20260903T090250Z-p38353`, `20260903T090535Z-p82919`, `20260903T090620Z-p84846`, `20260903T091102Z-p92057`, `20260903T091138Z-p93143`, `20260903T091708Z-p48108`, `20260903T092332Z-p57728`, `20260903T092833Z-p20493`, and final planner-fixture typecheck `20260903T094242Z-p33023`; all were corrected in source, tests, metadata, or layering and followed by green evidence. Closure scans find no deleted port/adapter/API, production synthetic clipboard event/cast, duplicate paste operation executor, selector/timer/frame focus in the paste cone, old replay controller, new TODO/FIXME, misplaced vendor import, generated product/public-contract/golden delta, or compatibility alias. Browser/test-support `ClipboardEvent` construction is retained as legitimate native-event evidence, not production bridging. Compatibility is source-only and behavior-preserving; rollback is the complete CR-08 adapter/model/controller/composition/editor/fixture/test/browser/ownership/catalog/generated-index/README/guide/tracker unit with no data rollback. Residual risk is limited to the four moderate static notes and broader Timeline loading/mutation/type-safety debt assigned to CR-09 through CR-11. Next action: append the CR-09 checkpoint, refresh loader identities/effects and race fixtures, then install the pure load machine before removing superseded mutable request branches. |
| `2026-09-03T06:19:16-04:00` | CR-09 `IN_PROGRESS`; CR-C09 remains `PLANNED` | Predecessor CR-08/CR-C08 confirmed `DONE`. Refreshed Core 03 query-state, high-water-version, mutation-refresh, continuity, collaboration-refresh, access-loss, and closed/read-only requirements; the current `useTimelineRowsLoader`, committed-row/query-admission epoch and generation ports, query identity construction, editor-draft registry, created-row pin, source-record continuity obligation, latest-request abort runtime, mutation/collaboration callers, focused task-guide routes, generated policy, Fallow artifact, Git branch/scope, and CR-01 assertion ledger. Owners remain sufficient and conflict-free. Current characterization preserves a bounded retry depth of two, initial versus refresh error presentation, stale-query join against already committed rows, mutation-epoch rejection, source-version obligations, high-water row selection, local draft hydration, exactly one draft, created-row pinning, protected-row clearing on access loss, semantic continuity completion, and abort-on-unmount. Fresh Fallow identifies the 34-cyclomatic/43-cognitive `loadTimelineRows` function as the CR-09 critical hotspot; the two row-mutation coordinator critical findings remain assigned to CR-10, and committed projection is a moderate CR-10 concern. | CR-09 is limited to a pure incident/surface/query/generation/mutation-epoch/source-obligation load machine, explicit start/success/stale/failure/accepted-mutation/access-loss/subject-change/retry-exhaustion events and request/retry/commit/status-error/protected-clear effects, one projection-commit adapter containing any evidence-required `flushSync`, loader migration/deletion of superseded mutable branches, deterministic plan/reducer and integration/browser evidence, metadata/guides, and this tracker. Compatibility requires unchanged query route/wire, retry bound, rows/drafts/pinning/high-water versions, refresh obligations, errors, access-loss callback, focus, and viewport continuity. Rollback is the complete CR-09 model/controller/test/metadata/generated-index/README/guide/tracker unit with no data rollback. Residual risks are late-response row regression, mutation/query race loss, retry non-convergence, stale subject commits, duplicated loading presentation, and premature continuity completion. Next action: add exhaustive pure transition tests for every event/effect and subject mismatch, then integrate the machine around the existing exact query and row-reconciliation boundaries before deleting the legacy recursive decision branches. |
| `2026-09-03T06:54:14-04:00` | CR-09 and CR-C09 `DONE`; CR-10 remains `PLANNED` | Added pure `timelineLoadMachine.ts` and exhaustive cataloged tests with an explicit two-retry freshness limit, complete incident/surface/query/request-generation/mutation-epoch/source-version identity, all required events, and explicit request/retry/commit/status/error/protected-row effects. Rebuilt `useTimelineRowsLoader.ts` as the effect interpreter around the unchanged query/reconciliation boundaries; stable scalar query and sheet identities replace object-identity coupling, complete-subject admission rejects stale commits/retries/exhaustion, satisfied superseded obligations may only join already-current rows, and accepted mutations advance the authoritative epoch. Added the sole `timelineProjectionCommitAdapter.ts` containing evidence-required `flushSync`; removed direct loader/coordinator imports and the superseded `committedRowsChangedSince` and legacy recursive decision branches. Preserved high-water row filtering, source-version obligations, local draft hydration, exactly one bottom draft, created-row pinning, mention/notice pruning, access-loss clearing, initial/refresh error distinction, semantic focus, and viewport continuity. Updated composition/coordinator tests, source ownership, Timeline test-family metadata, local README, frontend guide, the Make-generated execution-topology digest, and this tracker. No route, wire, persistence, authorization policy, dependency, generated public contract, Grid Adapter API, golden, or valid user-visible behavior changed. | PASS pure machine `.cartulary/test-results/20260903T102833Z-p63231` and final exact seven-case row `.cartulary/test-results/20260903T105147Z-p86390`; focused race/draft/refresh evidence `.cartulary/test-results/20260903T102937Z-p69128`, `20260903T103018Z-p70138`, and final stale-response rows `.cartulary/test-results/20260903T105354Z-p95540`; frontend unit 421/421 `.cartulary/test-results/20260903T103038Z-p71043`; Timeline owner 56/56 and service-backed 30/30 `.cartulary/test-results/20260903T103552Z-p24774` and `20260903T104030Z-p87538`; Workbook owner 68/68 and service-backed 39/39 `.cartulary/test-results/20260903T104512Z-p47153` and `20260903T104726Z-p6589`; `web.workbook` 159/159 `.cartulary/test-results/20260903T104934Z-p62687`; architecture 12/12 `.cartulary/test-results/20260903T105019Z-p79854`. PASS final format/typecheck/import/Biome `.cartulary/test-results/20260903T105143Z-p82255`, `20260903T105151Z-p86934`, `20260903T105400Z-p96193`, and `20260903T105403Z-p96589`; generate/drift/generated-policy/JSON `.cartulary/test-results/20260903T105201Z-p87355`, `20260903T105210Z-p90219`, `20260903T105218Z-p93296`, and `20260903T105219Z-p93700`; Git whitespace is clean. The initial integration run `.cartulary/test-results/20260903T102855Z-p63922` exposed unstable sheet-ref object identity and passed after canonical `sheetRefKey` correction; the later format attempt `.cartulary/test-results/20260903T105132Z-p81958` correctly rejected non-ASCII-sorted authored selector metadata, which was reordered before the complete green chain. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T105229Z-p94229` was inspected directly: 966 files, 16,785 functions, 714 critical, 314 high, and 496 moderate repository-wide; the CR-09 machine/loader/commit-adapter cone has no high/critical finding and only two moderate notes (`applyLifecycleEffects` and `loadTimelineRows`). Closure scans find direct `flushSync` only in the one commit adapter and no old epoch predicate, legacy retry constant, old replay controller, generic pending/drainer API, new TODO/FIXME, product-generated/public-contract/golden delta, or Git whitespace error. Compatibility is source-only and behavior-preserving. Rollback is the complete CR-09 model/adapter/loader/composition/coordinator/test/ownership/catalog/generated-index/README/guide/tracker unit; no data rollback is required. Residual risk is limited to the two moderate loader notes and the row-mutation/model/type-safety debt assigned to CR-10 and CR-11. Next action: append the CR-10 checkpoint, refresh the mutation/coordinator/model/collection dependency cone and exact invariants, then extract one owner per concern without widening into related-record, Mention, viewport, or broad presentation work. |
| `2026-09-03T06:56:49-04:00` | CR-10 `IN_PROGRESS`; CR-C10 remains `PLANNED` | Predecessor CR-09/CR-C09 confirmed `DONE`. Refreshed Core 03 field-level concurrency, committed-version high-water, keyboard/blur deduplication, conflict/draft retention, same-surface continuity, Timeline mutation and collection requirements; design collection typing, focus, save-state, conflict, and editor requirements; current mutation-command, row-coordinator, field/row model, collection-renderer dependencies; source ownership, catalog routes, generated policy, CR-01 assertion ledger, Git scope, and Timeline/Workbook/web/architecture task guides. Branch remains `main` at `c3d5c449551e7b58ae49b61f57f745aa385b36fa`, `0 1` relative to `origin/main`; the cumulative 133-path staged/unstaged worktree remains authorized and must be preserved. Owners are sufficient and conflict-free. The green CR-09 Timeline 56/56, Timeline service 30/30, Workbook 68/68, Workbook service 39/39, web Workbook 159/159, architecture 12/12, and frontend 421/421 runs are the deletion baseline. Direct final Fallow inspection identifies scoped critical `queueScalarSave`, `applyAcceptedRowMutation`, and `reconcileDiscardedPendingUnit` findings, high collection-renderer branching, and moderate `queueCollectionSave`; structural relationship/tag and registry/decoder casts remain ledgered for CR-10/CR-11. | CR-10 is limited to exact mutation-intent/queue/editor/grid units; accepted projection, discarded reconciliation, committed-version, conflict, socket-transaction, and save-state units; field registry, row normalization/materialization, mutation intent, layout/presentation policy; discriminated relationship/tag presentation; focused tests, ownership/catalog/generated index, guides, and this tracker. It must preserve scalar/collection/create/action behavior, queue order/signatures, keyboard/blur single-commit admission, latest committed versions, accepted/discarded/conflict projections, socket settlement, save labels, one bottom draft, created-row pinning, mentions/notices, focus, viewport continuity, and reference-preserving no-ops. Related-record, Mention adapter, broad presentation, and viewport-continuity redesign remain excluded. Rollback is the complete CR-10 source/test/metadata/generated-index/README/guide/tracker unit with no data rollback. Residual risks are duplicate mutation admission, lost drafts, version regression, incorrect conflict/status precedence, transaction leaks, union-member misrendering, and lifecycle/focus discontinuity. Next action: freeze pure intent/admission/reconciliation/presentation tables, then extract them behind the current composition ports before deleting each superseded branch. |
| `2026-09-03T07:52:33-04:00` | CR-10 and CR-C10 `DONE`; CR-11 remains `PLANNED` | Deleted `timeline/models/workbookTimelineModel.ts`; renamed its test to `timelineModelBoundaries.test.ts`; migrated every production consumer to exact new `timelineFieldRegistry.ts`, `timelineRowModel.ts`, `timelineMutationIntents.ts`, `timelineLayoutPolicy.ts`, and `timelineConflictState.ts` owners without an alias or barrel facade. Added pure `timelineMutationQueueAdmission.ts`, `timelineAcceptedProjection.ts`, `timelineAcceptedMutationEffects.ts`, `timelineDiscardedReconciliation.ts`, `timelineCommittedVersionLedger.ts`, and discriminated `timelineCollectionPresentation.ts`; added exact scalar-grid and socket-transaction adapters plus `TimelineCollectionCell.tsx`. Rebuilt `useTimelineMutationCommands.ts`, `useTimelineCommittedRows.ts`, `useTimelineRowMutationCoordinator.ts`, and `useTimelineCollectionRenderer.tsx` as narrow interpreters/composition around those units while retaining the existing exact editor, conflict, and save-state adapters. Relationship and tag presentation is now exhaustive rather than cast-selected, committed versions are monotonic and reference-preserving, keyboard/blur admission is deterministic, accepted/discarded projection is pure, socket settlement is isolated, and collection focus no longer uses animation-frame polling; the editor adapter delegates semantic reveal to one Grid Adapter scroll command. Added `timelineMutationModels.test.ts` and its catalog row, updated direct tests/imports, architecture fixtures, source ownership, Timeline/Workbook catalog inputs, local README, frontend guide, and only the Make-generated execution-topology digest. No related-record, Mention, viewport, broad presentation, route, wire, persistence, authorization-policy, dependency, generated public contract, Grid Adapter API, or golden change was introduced. | PASS exact mutation-model row `.cartulary/test-results/20260903T112101Z-p60593`; exact coordinator row `.cartulary/test-results/20260903T112125Z-p61686`; frontend unit 422/422 `.cartulary/test-results/20260903T112503Z-p69510`; Timeline 57/57 and service-backed 30/30 `.cartulary/test-results/20260903T112718Z-p15640` and `20260903T113159Z-p76772`; Workbook 68/68 and service-backed 39/39 `.cartulary/test-results/20260903T113645Z-p36990` and `20260903T113854Z-p93573`; `web.workbook` 159/159 `.cartulary/test-results/20260903T113638Z-p36547`; architecture 12/12 `.cartulary/test-results/20260903T112712Z-p15201`; final typecheck/import/Biome/format `.cartulary/test-results/20260903T112431Z-p63482`, `20260903T112441Z-p63939`, `20260903T112501Z-p69084`, and `20260903T112457Z-p64895`; generate/drift/generated-policy/JSON `.cartulary/test-results/20260903T114111Z-p50538`, `20260903T114120Z-p53404`, `20260903T114128Z-p56475`, and `20260903T114129Z-p56879`; measurement 22/22 `.cartulary/test-results/20260903T114305Z-p58010`; support 19/19 `.cartulary/test-results/20260903T114741Z-p16982`; and visual 12/12 against unchanged goldens `.cartulary/test-results/20260903T114907Z-p61879`. Introduced failures were retained and corrected: strict nullable/optional collection props `.cartulary/test-results/20260903T110729Z-p17621`; readonly coordinator port mismatch `.cartulary/test-results/20260903T111302Z-p33296`; a pure ledger fixture that incorrectly assumed the first accepted projection would retain an absent raw cell `.cartulary/test-results/20260903T112030Z-p55718`; and Biome import ordering `.cartulary/test-results/20260903T112444Z-p64330`. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T115101Z-p7827` was inspected directly: 980 files, 16,826 functions, 711 critical, 313 high, and 497 moderate repository-wide; the CR-10 replacement cone has no high/critical finding and three moderate coverage/complexity notes (`committedTimelineProjection`, ledger `latest`, and `queueCollectionSave`). Closure scans find no production old model/reference, relationship/tag downcast, old replay controller, selector/timer/frame focus in the replaced collection/editor paths, structural cast in the new exact production units, new TODO/FIXME, product-generated/golden delta, or Git whitespace error. The scoped public-contract diff remains only the intentional CR-03 removal of obsolete saved-view selectors. Compatibility is source-only and preserves all listed Timeline behavior and public boundaries. Rollback is the complete CR-10 model/adapter/hook/coordinator/component/import/test/ownership/catalog/generated-index/README/guide/tracker unit; no data rollback is required. Residual risk is limited to those three moderate static notes and the broader scoped boundary assertions assigned to CR-11. Next action: append the distinct CR-11 checkpoint, reconcile every CR-01 assertion-ledger entry against current production, then replace each remaining target with an exact constructor, guard, decoder, checked lookup, or exhaustive dispatch before adding negative and compile-time closure evidence. |
| `2026-09-03T07:54:11-04:00` | CR-11 `IN_PROGRESS`; CR-C11 remains `PLANNED` | Predecessor CR-10/CR-C10 confirmed `DONE`. Refreshed the Core 03 fail-closed, request, mutation, focus, and Timeline requirements; current generated request vocabulary; Workbook runtime/queue, command adapters, conflict decoder, grid-continuity/anchor, Timeline mention/row/model boundaries; source ownership and test-family inputs; task guides for `web.workbook`, `web.architecture`, `module.workbook`, `module.timeline`, and `module.entities`; branch `main`, HEAD `c3d5c449551e7b58ae49b61f57f745aa385b36fa`, upstream relation `0 1`, and the cumulative 197-path authorized worktree. Toolchain and generated-policy digests remain `d530190d...` and `ddc06647...`. Reconciled CR-01 entries already closed by CR-02 through CR-10: boolean query-control parsing, generic runtime-envelope/replay admission, synthetic paste conversion, committed-row indexing, field registry and relationship/tag downcasts, and mutation-reconciliation decoding. Remaining scoped targets are generated-request assertions in Workbook command/pending/query/conflict adapters, generic queue JSON/conflict casts, grid viewport/anchor assertions and unchecked layout indexing, Timeline mention decoding, conflict decoding, and any equivalent replacement assertion or wildcard dispatch discovered during closure. Adjacent Evidence/Import Assistant/related-record/presentation assertions and test-only DOM assertions remain recorded exclusions. | CR-11 is limited to exact request constructors, guards/decoders, discriminated registries, checked indexing, exhaustive `never` handling, negative/runtime boundary tests, compile-time exhaustive evidence, metadata/guides when ownership changes, generated projection reconciliation, and this tracker. Valid route/wire/public behavior remains compatible; malformed values intentionally fail closed. Rollback is the complete CR-11 production/test/metadata/generated-index/guide/tracker unit with no data rollback. Residual risks are malformed generated input being trusted, new union members entering fallback branches, wrong grid anchors, unsafe conflict replay, and assertion-shaped replacement helpers. Next action: inspect each remaining target with its generated type and caller, add exact constructors/decoders first, then delete the assertions category by category and close with negative tests plus a zero-target scoped scan. |
| `2026-09-03T08:56:47-04:00` | CR-11 and CR-C11 `DONE`; CR-12 remains `PLANNED` | Added private `workbook/adapters/workbookProtocolTypes.ts` and exact `models/workbookRequestDecoders.ts` plus negative/exact tests. Migrated Workbook query, create, patch, pending, conflict, Generic, Entity, Assessment, Timeline anchor, continuity, mention, presence, saved-view/startup, Evidence lifecycle, inspector, and queue boundaries from generated-request, JSON-record, union-member, grid-anchor, enum/select, synthetic-event, and unchecked-index assertions to exact construction, guards, checked lookup, discriminated registries, or fail-closed decoders. Removed the query failure compatibility re-export and made callers use the canonical port helper. Generated collection-action and inspector-owner registries use exhaustive `satisfies Record<union, ...>` coverage; unknown action kinds, conflict classes, mention kinds, direct scalar contracts, enum values, timestamps, non-clearable nulls, ambiguous changes, malformed collection members, and future request keys now fail closed. Updated source ownership, Workbook/Entities test-family rows, local README, frontend guide, and the Make-generated execution-topology digest. No production file was deleted or moved in CR-11; no route, wire, persistence, authorization, Grid Adapter public API, generated public contract, dependency, or golden changed. The cumulative authorized worktree contains 244 paths and remains `main` at `c3d5c449551e7b58ae49b61f57f745aa385b36fa`, `0 1` relative to `origin/main`. | PASS exact request boundary `.cartulary/test-results/20260903T122127Z-p36350`, mention boundary `.cartulary/test-results/20260903T122140Z-p36950`, conflict boundary `.cartulary/test-results/20260903T122158Z-p37601`, continuity/inspector correction `.cartulary/test-results/20260903T123327Z-p94884` and `.cartulary/test-results/20260903T123327Z-p94924`; final frontend unit 423/423 `.cartulary/test-results/20260903T124928Z-p92246`; `web.workbook` 159/159 `.cartulary/test-results/20260903T123342Z-p98106`; Workbook 69/69 and service-backed 39/39 `.cartulary/test-results/20260903T124928Z-p92122` and `.cartulary/test-results/20260903T123342Z-p98144`; Timeline 57/57 and service-backed 30/30 `.cartulary/test-results/20260903T123740Z-p53131` and `.cartulary/test-results/20260903T123740Z-p53157`; Entities 43/43 and service-backed 33/33 `.cartulary/test-results/20260903T124928Z-p92101` and `.cartulary/test-results/20260903T123740Z-p53196`; architecture 12/12 `.cartulary/test-results/20260903T124928Z-p92109`; final format/typecheck/import/Biome `.cartulary/test-results/20260903T125625Z-p51851`, `.cartulary/test-results/20260903T125625Z-p51787`, `.cartulary/test-results/20260903T125625Z-p51823`, and `.cartulary/test-results/20260903T125625Z-p51831`; generate/drift/generated-policy/JSON `.cartulary/test-results/20260903T124240Z-p75773`, `.cartulary/test-results/20260903T124310Z-p79571`, `.cartulary/test-results/20260903T124310Z-p79588`, and `.cartulary/test-results/20260903T124310Z-p79594`. Introduced failures were retained and corrected: type narrowing at `.cartulary/test-results/20260903T120505Z-p16377`, `.cartulary/test-results/20260903T120817Z-p18268`, `.cartulary/test-results/20260903T121313Z-p21339`, `.cartulary/test-results/20260903T121358Z-p22103`, and `.cartulary/test-results/20260903T122035Z-p35109`; protocol-boundary/source-order failures `.cartulary/test-results/20260903T122531Z-p44975` and `.cartulary/test-results/20260903T122531Z-p44851`; stale discriminator/token test expectations `.cartulary/test-results/20260903T122531Z-p44952`; missing refactor binding `.cartulary/test-results/20260903T124833Z-p86553`; and transient unrelated debug-harness lazy-load timeouts under four-way resource contention `.cartulary/test-results/20260903T123342Z-p98247`, which passed in the isolated full rerun. Fresh owner-only-umask Fallow `.cartulary/test-results/20260903T125258Z-p43754` was inspected directly: 983 files, 16,894 functions, 711 critical, 308 high, and 507 moderate repository-wide; every named structural replacement and the formerly high grid-anchor resolver are absent from high/critical findings, with no suppression. Remaining high/critical entries are explicit test, unrelated Workbook-operation, startup, related-record, Evidence/Mention-adapter, or other section-4 exclusions. Closure scans leave no scoped production assertion, non-null escape, generated-request cast, wildcard union dispatcher, query compatibility re-export, old mapper/replay/paste symbol, new TODO/FIXME, Markdown dependency, generated-product/golden delta, or Git whitespace error; residual matches are TypeScript import renames/UI prose, allowed `as const`/`satisfies`/typed CSS, test-only DOM assertions, and the two recorded Import Assistant casts. Compatibility is source-only for valid inputs; malformed and future values intentionally fail closed. Rollback is the complete CR-11 source/test/ownership/catalog/generated-index/README/guide/tracker unit with no data rollback. Residual risk is only the explicitly excluded repository-wide static debt. Next action: append the separate CR-12 checkpoint, reconcile every final owner/task guide and generated projection, then execute the terminal validation and handoff matrix. |
| `2026-09-03T08:58:28-04:00` | CR-12 `IN_PROGRESS`; CR-C12 remains `PLANNED` | Predecessor CR-11/CR-C11 confirmed `DONE`; the post-workstream Markdown lint passed at `.cartulary/test-results/20260903T125755Z-p57487`. Refreshed Core 03, design, domain vocabulary/navigation, verification/catalog/topology ownership, all 13 final module-author task guides, branch `main`, HEAD `c3d5c449551e7b58ae49b61f57f745aa385b36fa`, upstream relation `0 1`, and the cumulative 244-path authorized worktree. Toolchain and generated-policy digests remain `d530190d...` and `ddc06647...`; no owner contradiction or authority drift was found. Every guide resolves to the expected focused route, and applicable owners expose service-backed routes plus the browser classes recorded in section 19. The latest generated reconciliation changes only `tools/execution_topology_render_index.json` for authored test-family hashes; generated product roots and goldens remain unchanged. | CR-12 is limited to resolving any validation-discovered defect within the authorized dependency cone, final owner/catalog/topology reconciliation, focused and service-backed owner routes, the complete terminal Make matrix, direct artifact and generated/golden/public-contract inspection, closure scans, compatibility/rollback/residual-risk/deferral records, and final tracker completion. Rollback is the cumulative CR-01 through CR-12 source/test/metadata/generated-index/guide/tracker unit, workstream-granular where dependencies permit, with no data migration. Residual risks are terminal browser/resource contention, stale generated metadata, global closure-search false positives, and explicitly excluded repository-wide Fallow debt. Next action: run every final focused and applicable service-backed owner route, then execute formatting, agent-finalize, terminal frontend/browser/test/static checks and close only after direct scope review. |
| `2026-09-03T09:44:39-04:00` | CR-12 and CR-C12 `DONE`; all workstreams and gates `DONE` | Reconciled the final owner, test-family, catalog, and execution-topology inputs and refreshed all 13 module-author task guides. Focused routes passed for `web.workbook` 159/159 `.cartulary/test-results/20260903T125937Z-p61946`, `web.architecture` 12/12 `.cartulary/test-results/20260903T125937Z-p61960`, `web.design` 15/15 `.cartulary/test-results/20260903T125937Z-p61980`, `web.networkflow` 38/38 `.cartulary/test-results/20260903T125937Z-p62004`, `package.grid_adapter` 43/43 `.cartulary/test-results/20260903T130132Z-p34378`, `module.workbook` 69/69 `.cartulary/test-results/20260903T130132Z-p34371`, `module.timeline` 57/57 `.cartulary/test-results/20260903T130132Z-p34392`, `module.entities` 43/43 `.cartulary/test-results/20260903T130633Z-p52425`, `module.collaboration` 31/31 `.cartulary/test-results/20260903T130829Z-p5359`, `module.evidence` 36/36 `.cartulary/test-results/20260903T130829Z-p5337`, `module.assessments` 28/28 `.cartulary/test-results/20260903T130829Z-p5362`, `module.indicators` 20/20 `.cartulary/test-results/20260903T130829Z-p5338`, and `module.networkflow` 34/34 `.cartulary/test-results/20260903T131025Z-p71151`. Applicable service-backed routes passed for Grid Adapter 13/13 `.cartulary/test-results/20260903T131240Z-p31912`, design 15/15 `.cartulary/test-results/20260903T131240Z-p31922`, Workbook 39/39 `.cartulary/test-results/20260903T131453Z-p86633`, Timeline 30/30 `.cartulary/test-results/20260903T131453Z-p86645`, Entities 33/33 `.cartulary/test-results/20260903T131453Z-p86631`, Collaboration 22/22 `.cartulary/test-results/20260903T131938Z-p54469`, Evidence 25/25 `.cartulary/test-results/20260903T131938Z-p54476`, Assessments 19/19 `.cartulary/test-results/20260903T131240Z-p31916`, Indicators 8/8 `.cartulary/test-results/20260903T131240Z-p31938`, and Network Flow 28/28 `.cartulary/test-results/20260903T131938Z-p54482`. The first concurrently loaded Entities focused run reached 41/43 at `.cartulary/test-results/20260903T130132Z-p34408` because one `mentions.resolve` browser focus-continuity row timed out under concurrent browser load; its immediate isolated 43/43 pass classifies that failure as resource contention rather than a product regression. | PASS generation `.cartulary/test-results/20260903T132201Z-p15821` with no generated diff; drift `.cartulary/test-results/20260903T132220Z-p18907`; generated policy `.cartulary/test-results/20260903T132220Z-p18926`; JSON shape `.cartulary/test-results/20260903T132220Z-p18947`; format `.cartulary/test-results/20260903T132231Z-p22839`; `agent-finalize` `.cartulary/test-results/20260903T132238Z-p26997` with `RESULTS_DIR` intentionally unset; typecheck `.cartulary/test-results/20260903T132257Z-p30260`; frontend unit 423/423 `.cartulary/test-results/20260903T132257Z-p30281`; import boundary `.cartulary/test-results/20260903T132257Z-p30301`; Biome `.cartulary/test-results/20260903T132257Z-p30346`; accessibility 12/12 `.cartulary/test-results/20260903T132335Z-p42272`; measurement 22/22 `.cartulary/test-results/20260903T132505Z-p87970`; stateful 34/34 `.cartulary/test-results/20260903T132936Z-p46734`; support 19/19 `.cartulary/test-results/20260903T133158Z-p97286`; webserver-backed 60/60 `.cartulary/test-results/20260903T133323Z-p42698`; visual 12/12 `.cartulary/test-results/20260903T133915Z-p1544`; `test-fast` 475/475 `.cartulary/test-results/20260903T134117Z-p47140`; and post-tracker Markdown lint `.cartulary/test-results/20260903T134551Z-p55710`. Fresh owner-only-umask Fallow passed at `.cartulary/test-results/20260903T134207Z-p53493`; direct JSON inspection reports 983 files, 16,894 functions, 711 critical, 308 high, and 507 moderate findings repository-wide, while every named or replacement production path remains free of high/critical findings and suppression. The only high replacement-file match is the explicitly excluded pre-existing Generic create builder. Closure searches found no deleted mapper, copied Grid policy, menu/overlay selector/timer/frame focus, redundant saved-view production action, generic metadata/drainer/replay bag, duplicate paste transport, synthetic clipboard bridge, old Timeline replay controller, scoped structural escape, compatibility facade or legacy import, wildcard owner dispatcher, test-only production export, new TODO/FIXME, misplaced Grid Adapter vendor import, or Markdown runtime dependency. Residual matches are the negative removed-action test, TypeScript import renames, fail-closed/presentation defaults, existing typed queue vocabulary, allowed literals, and the two recorded Import Assistant casts. Git staged/unstaged review and whitespace checks are clean; Make generation added no diff; generated product roots and visual goldens are unchanged; the only public-package change is the intentional CR-03 deletion of obsolete Rename/Manage Sharing selector exports. Valid routes, wire bodies, persistence, authorization, Grid Adapter APIs, remaining selectors, and user behavior stay compatible; malformed/future values fail closed and Rename/Manage Sharing are intentionally removed. No data migration or rollout is required. Rollback is source-only and workstream-granular, including each workstream's production, tests, authored metadata, generated topology digest, guide/README, and append-only tracker record. Residual risk is limited to the documented moderate notes and explicitly excluded repository-wide Fallow, Import Assistant, Evidence/Mention, related-record, broad viewport/presentation, and unrelated harness debt. The extension seam is one owner-local pure policy/model, one narrow effect or exact driver port, the unchanged Grid Adapter contract where sufficient, and a new closed runtime-envelope member only for a genuinely new mutation owner. No authorized work or deferred in-scope action remains; handoff is complete. |

Do not alter this planning row when implementation begins. Append the refreshed
pre-CR-01 checkpoint below it and continue the log append-only.
