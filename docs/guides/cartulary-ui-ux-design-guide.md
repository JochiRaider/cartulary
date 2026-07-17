# Cartulary UI/UX Design Guide

**Revision:** 2.  
**Status:** Derived design-direction specification.  
**Authority:** Subordinate to Cartulary Core 00 through Core 04 for current runtime behavior, subordinate to Core 05 only for claim-bearing timed or fixture-sensitive publication behavior, and subordinate to any later adopted Cartulary NLSpec.  
**Scope:** Base-profile UI/UX direction, with explicit extension-profile and later-scope boundary notes.  
**Non-authority statement:** This guide does not define Base Profile or extension-profile implementation conformance. Current implementation conformance remains owned by Core 00 through Core 04. Core 05 governs claim-bearing publication only. Guide-local normative language governs design-direction interpretation and reviewer discipline only, and MUST NOT be read as widening, narrowing, or replacing owner behavior.[^1][^2][^3]

## 1. Executive Summary

Cartulary’s UX thesis is precise: the product is designed to feel like a serious workbook on the hot path and behave like a disciplined case system underneath. The product preserves the spreadsheet mental model at the view layer while rejecting spreadsheet storage, concurrency, evidence, and audit semantics where those would undermine recoverability or collaboration.[^4][^5][^6]

Revision 2 closes the open design gaps in the previous guide by converting broad direction into observable UI contracts. It defines a closed statement-class grammar, a deterministic shell-composition model, explicit status-strip capacity, bounded visual-language defaults, keyboard and accessibility posture, cross-cutting message patterns, and reviewer-facing acceptance criteria.[^7]

The base guide’s central UI choice is that built-in tabs remain primary, required system views are reachable through an adjacent `System views` switcher, saved views belong under the active surface’s view selector, and coordination surfaces remain workbook-native. The guide rejects all-fourteen-tabs, command-palette-only access to required system views, and separate coordination modules for the base design.[^4][^5]

The guide also tightens behavioral defaults that affect implementation divergence: `Esc` is an interaction-priority ladder, `Syncing` includes paused replay waiting on recovery or re-authentication, pending queues do not survive cross-tab transfer, rollback actions MUST reveal target scope, chip states use a closed visual vocabulary, and ordinary live editing MUST NOT display per-recipient release visibility controls.[^4][^5][^6]

## 2. Reading Guide and Source Method

### 2.1 Authority order

Current product behavior is governed in this order:

1. future adopted Cartulary NLSpecs derived from the core, once adopted;
2. Cartulary Core 00 through Core 04 for current implementation-conformance behavior;
3. non-normative appendices for rationale, examples, illustrations, operating guidance, and future-only context;
4. the exploratory source artifact.[^2]

Core 05 governs claim-bearing timed or fixture-sensitive publication only. It is not Base Profile or extension-profile implementation conformance, and it does not broaden ordinary runtime UI behavior.[^3]

This guide MAY use the development guide, repository bootstrap guide, and progressive implementation/testing guide as implementation-support context. Those guides are subordinate to the authority order above and do not replace it.[^8]

### 2.2 Source classes used by this guide

The table below expands the Core 00 authority order into the source classes this guide consumes. It does not replace the Core 00 precedence order.

| Source class                         | Role in Revision 2             | Permitted effect                                                  |
| ------------------------------------ | ------------------------------ | ----------------------------------------------------------------- |
| Core 00 through Core 04              | Current behavior owner         | MAY be restated as `*Core behavior.*` only.                       |
| Core 05                              | Publication owner only         | MAY be cited for benchmark-publication consequences only.         |
| Appendices A through H               | Non-normative context          | MAY justify design direction, but cannot create runtime behavior. |
| Research reports R01 through R07     | Design rationale               | MAY justify design trade-offs, not conformance behavior.          |
| Development/bootstrap/testing guides | Implementation-support context | MAY shape baseline direction, not product authority.              |

### 2.3 Statement classes in this guide

This guide uses exactly four inline statement-class markers.

| Marker                | Meaning                                                                                       | Normative language rule                                                                                 |
| --------------------- | --------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------- |
| `*Core behavior.*`    | Descriptive restatement of behavior owned by Core 00 through Core 04.                         | Does not issue guide-owned imperatives. Uses “Core X requires...” or “Per REQ-X...” phrasing.           |
| `*Design direction.*` | Guide-owned UI/UX direction derived from Core behavior or project rationale.                  | MAY use MUST, MUST NOT, SHOULD, SHOULD NOT, and MAY as guide-local direction.                           |
| `*Baseline context.*` | Current implementation-baseline direction that shapes design without owning product behavior. | MAY use guide-local normative language only for baseline-facing design constraints.                     |
| `*Later scope.*`      | Extension-profile or future-only territory the base guide MUST leave room for.                | MAY use MAY for future scope and MUST NOT only to prevent misrepresenting later scope as base behavior. |

No synonym markers are allowed. `*Core requirement.*`, `*Baseline.*`, `*Future scope.*`, and similar variants are invalid in this guide.

### 2.4 Marker scope

Statement-class markers apply to every prose paragraph in §§3 through 19. They also apply to standalone normative list items in those sections. For tables, a `Statement class` column is required only when rows have mixed statement classes. Otherwise, the immediately preceding marked paragraph states the class for every row in the table.

Markers are not required inside headings, figure captions, code fences, source footnotes, acceptance criteria, the Revision 2 change log, or the Revision 2 editorial audit.

A paragraph marked `*Core behavior.*` is descriptive. If the guide needs to state a UI consequence, that consequence appears in a separate `*Design direction.*` paragraph. This prevents owner restatements from becoming accidental new requirements.[^2][^8]

### Acceptance criteria

- **R2-AC-001:** §2 states the Core 00 precedence order as an authority hierarchy.
- **R2-AC-002:** §2 states that Core 05 is publication-only and does not broaden runtime conformance.
- **R2-AC-003:** The source-method table is framed as an expansion, not a replacement, of Core 00 §2.
- **R2-AC-004:** No statement in §2 claims that implementation-support guides outrank Core 00 through Core 04.
- **R2-AC-005:** The marker legend exists and defines exactly four markers.
- **R2-AC-006:** Every marked paragraph uses exactly one marker from the closed vocabulary.
- **R2-AC-007:** No `*Core behavior.*` paragraph contains guide-issued MUST, MUST NOT, SHOULD, or SHOULD NOT imperatives.
- **R2-AC-008:** Any normative table either has a `Statement class` column or is introduced by a marked paragraph that applies to every row.
- **R2-AC-009:** Acceptance criteria are not marked with statement-class prefixes and remain binary.

## 3. Product and Interaction Thesis

### 3.1 Thesis contract

*Design direction.* Cartulary MUST be understood as a workbook-native incident workspace whose primary interaction object is the visible row, cell, chip, preview affordance, filter state, grouping state, and status state, not the hidden form model behind it.[^4][^9]

*Design direction.* The base UI MUST preserve a direct-manipulation loop: the user acts on the visible workbook object, the system shows how that action was interpreted, and slower reconciliation work remains visible as semantic state rather than hidden transport state.[^9][^10]

*Core behavior.* Core 03 requires Cartulary to be grid-first and forms-second, with a workbook-like grid supporting inline editing, keyboard navigation, paste, low-friction row creation, saved or system views over projections, and a collapsible inspector for enrichment, relationships, history, and destructive actions.[^4]

*Design direction.* The visible workbook surface MUST remain the default locus for capture, correction, linking, filtering, grouping, sorting, preview, and same-row history access. Routine row work MUST NOT require full-page navigation or a form-first workflow.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Design question                | Required answer                                                                       |
| ------------------------------ | ------------------------------------------------------------------------------------- |
| Main object of thought         | Visible workbook row, cell, chip, count, preview, filter, grouping, and status state. |
| Main editing mode              | Direct manipulation on the grid.                                                      |
| Storage model exposed to users | Structured, relational, auditable state projected through workbook surfaces.          |
| Secondary surfaces             | Adjacent enrichment, review, history, and destructive actions.                        |
| Collaboration target           | Reliable shared case state, not character-by-character coauthoring.                   |
| Error posture                  | Local, same-surface correction before modal or detached recovery.                     |

### 3.2 Design implications

*Design direction.* Cartulary MUST show semantic interpretation at the same level where the user acted. A typed host token remains a token until resolved; an exact-match reuse becomes a chip; an evidence attach becomes count, state, and preview affordance; and a same-field collision becomes a marked cell with a resolver entry point.

*Design direction.* The grid MUST NOT be treated as a temporary intake surface that users graduate away from once structure exists. System views, saved views, coordination surfaces, and optional standardized artifact-backed surfaces MUST retain workbook grammar.[^4][^5]

*Design direction.* The UI MUST hide transport, storage, and synchronization internals unless the user needs a semantic state to choose a next action. Valid semantic states include `Syncing`, `Saved`, `Conflict`, unresolved mention, resolved chip, auto-resolved chip, dismissed mention, queue overflow, replay blocked, and session re-authentication required.

### Acceptance criteria

- **R2-AC-010:** §3 states that the visible workbook object is the primary interaction object.
- **R2-AC-011:** §3 separates Core behavior from design direction.
- **R2-AC-012:** §3 says routine capture, correction, linking, filtering, grouping, sorting, preview, and history access remain same-shell operations.
- **R2-AC-013:** §3 names semantic user-facing states instead of implementation-layer sync internals.

## 4. Workbook-First Rationale

### 4.1 Operating-model rationale

*Design direction.* Cartulary MUST remain workbook-first because it replaces a real incident-response operating model, not an abstract CRUD problem.[^11][^12]

*Design direction.* The UI MUST tolerate incomplete information at capture time. The first useful row MAY contain uncertain time, rough prose, unresolved host or account strings, and a screenshot reference. The user MUST be able to capture that row before canonical structure is complete.[^6][^11]

*Core behavior.* Core 02 permits rough and uncertain input and preserves original rough capture after later normalization or resolution. Core 03 keeps low-friction row creation and progressive structuring on the workbook hot path.[^4][^6]

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Criterion            | Workbook-first Cartulary                    | Forms-first case application                   |
| -------------------- | ------------------------------------------- | ---------------------------------------------- |
| First useful capture | Direct row or cell entry.                   | Create form, choose type, satisfy fields.      |
| Incomplete facts     | Preserved as rough structured state.        | Often blocked or hidden in free-text forms.    |
| Relationship entry   | Text-first, then progressive normalization. | Picker-first or mandatory canonical selection. |
| Evidence attachment  | Same-surface attach and preview.            | Detached upload workflow.                      |
| Review and rollback  | Row-local history and scoped actions.       | Separate administrative audit screen.          |
| Coordination         | Workbook-native surfaces.                   | Separate task or workflow module.              |

### 4.2 Spreadsheet preservation boundary

*Design direction.* Cartulary SHOULD preserve spreadsheet strengths that matter during incident response: direct typing, paste, keyboard navigation, visible rows, flexible filtering, compact tabular scanning, and rapid working-set reshaping.[^9][^11]

*Design direction.* Cartulary MUST reject spreadsheet behaviors that conflict with auditable incident state: row-position identity, silent overwrites, hidden formulas as business logic, evidence paths as authoritative references, unmanaged binary storage, and unversioned relationship semantics.[^4][^5][^6]

### Acceptance criteria

- **R2-AC-014:** §4 states why rough capture MUST be allowed before complete structure exists.
- **R2-AC-015:** §4 differentiates workbook-first design from forms-first case management.
- **R2-AC-016:** §4 lists spreadsheet behaviors to preserve and spreadsheet behaviors to reject.

## 5. Overall Application Shell and Information Architecture

### 5.1 Shell regions

*Design direction.* The application shell MUST keep the workbook surface, active surface identity, saved-view controls, explicit inspector controls, presence, save state, and status state in one continuous workspace.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Region       | Required contents                                                                                       | Boundary                              |
| ------------ | ------------------------------------------------------------------------------------------------------- | ------------------------------------- |
| Top bar      | Incident identity, built-in tabs, active-surface title when not already represented by the selected built-in tab, `System views` switcher, current system-view title, top-bar query controls (`Sort`, `Group: None`, `Filters`, active query chips), and account navigation. | Persistent chrome, not a dashboard.   |
| View bar     | Compact sheet toolbar with saved-view selector, saved-view action menu, inspector opener, and add-row control when allowed. Saved-view mutation controls live inside the action menu, not as persistent rail controls. | Belongs to active surface only.       |
| Grid         | Active workbook surface with `record_id`-bound rows and `field_key`-bound cells.                        | Primary work surface.                 |
| Inspector    | Details, Relationships, Evidence, History, destructive and specialized row actions.                     | Closed by default; adjacent or overlay secondary surface only after explicit open. |
| Status strip | Save state, secondary same-surface message, presence summary or overflow. Save state labels are `Syncing`, `Saved`, or `Conflict`. | Capacity-limited working-state strip. |

*Design direction.* The default Timeline route MUST be grid-first: the first-viewport composition centers the active Timeline grid, compact sheet toolbar, explicit inspector opener, bottom draft row when creation is allowed, and status strip. The inspector MUST be closed by default and MUST open only through explicit controls such as the toolbar inspector control, row action menu, history action, mention action, or equivalent keyboard-accessible command. Incident summary, bootstrap defaults, membership management, incident-metadata patch forms, or other incident administration controls MUST NOT appear as a dominant card stack above the active grid unless the user explicitly opens a secondary incident-control surface or navigates to a distinct administration context.

*Design direction.* Optional incident-create metadata MAY live behind a compact disclosure so create remains lightweight. After creation, an explicitly opened secondary incident-control surface for current `reviewer` or `admin` users MUST expose all Core-owned patchable incident metadata: `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`. The TLP control MAY show readable labels, but it MUST bind to exact canonical machine tokens or `null`; severity and phase controls MAY show suggestions, but MUST preserve otherwise valid bounded text. `incident_key`, `title`, `status`, and `closed_at` are displayed as identity or lifecycle state rather than ordinary patch controls.

*Design direction.* When an explicitly opened secondary incident-control surface displays workbook bootstrap defaults, it MUST preserve the Core-owned `sheet_ref` identity as a structured pointer to a base view schema or saved view. Such displays MAY add readable labels, but MUST NOT treat visible labels, saved-view names, or stringified objects as the authoritative preference value.

*Core behavior.* Core 01 defines closed incidents as visible and readable to current members while blocking authoritative source-state mutations except reopen. Core 03 requires the workbook to render a persistent `Closed, read-only` lifecycle state and to keep local source-mutation drafts rejected by closure non-authoritative.[^4][^5]

*Design direction.* A closed workbook MUST render a persistent banner or shell-level state using the existing banner/read-only primitives and the exact visible label `Closed, read-only`. The state belongs in persistent shell chrome, not in a transient toast and not as a replacement for the `Syncing`, `Saved`, or `Conflict` save-state labels.

*Design direction.* When the incident is closed, source-state write controls MUST be disabled or hidden: add-row, direct cell edit, paste-to-mutate, row delete/restore/rollback/merge/supersede, conflict resolution, mention resolution, incident metadata patch, blob-slot creation, and evidence attachment. Read and derived-output affordances SHOULD remain available when their ordinary authorization allows them, including workbook queries, history, evidence preview/download, saved views, workbook preferences, snapshots, reports, releases, and incident export. Reopen MAY remain available to current incident admins through a secondary incident-control surface.

*Design direction.* The workbook work area between the view bar and status strip MUST own vertical sizing. The active grid and an open inspector MUST fill that same work area for zero, one, three, and many rendered rows, including empty, loading, error, and draft-row states. Grid content and inspector content MUST scroll independently, the status strip MUST remain at the bottom of the shell, and the workbook layout MUST NOT create document-level vertical scrolling. Synthetic filler rows, row-count height calculations, fixed `100vh - Npx` offsets, and surface-specific minimum-height patches are rejected because they make future workbook surfaces brittle.

### 5.2 Surface composition in the shell

*Core behavior.* Core 03 requires five built-in tabs in the base profile: Timeline, Hosts, Identities, Evidence, and Notes. It also requires additional contract-backed system views and keeps structured coordination artifacts workbook-native rather than separate modules.[^4]

*Design direction.* The base shell MUST expose built-in tabs as always-visible primary tabs at the base viewport. Required system views MUST be reachable through an adjacent switcher with accessible name `System views`. Saved views MUST appear under the active surface’s view selector, not as primary tabs by default.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Surface family                                 | Shell exposure                                          | Ordering                                                                             | Saved-view behavior                                                        |
| ---------------------------------------------- | ------------------------------------------------------- | ------------------------------------------------------------------------------------ | -------------------------------------------------------------------------- |
| Built-in tabs                                  | Always-visible primary tabs at base viewport.           | Timeline, Hosts, Identities, Evidence, Notes.                                        | Saved views over a built-in tab appear in that tab’s view selector.        |
| Required system views                          | Adjacent switcher with accessible name `System views`.  | Grouped and ordered by §5.3.                                                         | Saved views over a system view appear in that system view’s view selector. |
| Standardized optional artifact-backed surfaces | Same switcher, shown only when implemented and exposed. | Findings, Investigative Queries, Forensic Keywords.                                  | Saved views over these surfaces appear in that surface’s view selector.    |
| Saved views                                    | Never primary tabs by default.                          | Ordered inside the active surface’s view selector by scope group, then display name. | Do not replace canonical surface identity.                                 |

### 5.3 Required system-view ordering

*Design direction.* The `System views` switcher MUST group required system views in the order below. The implementation MUST NOT alphabetize these required groups differently unless a later guide revision changes this table.

| Group                | Required system views in display order                |
| -------------------- | ----------------------------------------------------- |
| Scope and assessment | Indicators; Assessments; Parties                      |
| Coordination         | Task Requests; Decisions; Communications Log; Handoff |
| Review and learning  | Status Review; Lesson                                 |

*Design direction.* If Findings, Investigative Queries, or Forensic Keywords are exposed, the switcher MUST add a final group named `Optional artifact surfaces` in this order: Findings, Investigative Queries, Forensic Keywords.

### 5.4 Switcher behavior

*Design direction.* At a base viewport of at least `1280x720` CSS pixels, the `System views` switcher trigger MUST remain visible in the top bar adjacent to the primary tab strip.

*Design direction.* The visible trigger label SHOULD be `System views`. An icon-plus-label implementation MAY use a shorter visible label only when the accessible name remains exactly `System views`.

*Design direction.* Selecting a system view MUST open it inside the same workbook shell. Selecting a system view MUST NOT navigate to a separate module, route family, browser tab, or dashboard shell.

*Design direction.* Keyboard operation MUST support Tab focus to the trigger, Enter or Space to open, Arrow keys to move within the open menu, Enter to select, and Esc to close without changing the active surface.

*Design direction.* When the menu closes without selection, focus MUST return to the switcher trigger. When a system view is selected, focus SHOULD move to the first focusable grid element on that surface unless the user invoked a saved keyboard shortcut whose defined behavior preserves prior focus.

### 5.5 Rejected shell-composition alternatives

*Design direction.* Revision 2 rejects three base-profile shell-composition alternatives: all fourteen required surfaces as primary tabs, command-palette-only access to required system views, and separate modules for coordination surfaces.

*Design direction.* All-fourteen-tabs is rejected because it degrades scanning once the visible tab count exceeds seven. Command-palette-only access is rejected because it hides required coordination surfaces behind keyboard discovery. Separate coordination modules are rejected because they fracture the workbook-native incident working model.

### 5.6 Reference wireframe

*Design direction.* The following figure is illustrative but conformance-relevant for composition: it shows the required shell relationships, not pixel-perfect styling.

```text
+--------------------------------------------------------------------------------+
| Incident: INC-2026-0418   [Timeline] [Hosts] [Identities] [Evidence] [Notes]   |
|                                                   [System views ▾] A B +3       |
+--------------------------------------------------------------------------------+
| View: Default ▾   Saved view name   Private   Create  Home  Default  Filters 0 |
| Sort ▾   Group ▾   Filter ▾                         Inspector  Add             |
+--------------------------------------------------------------------------------+
| Grid: active workbook surface                                                  |
| record_id-bound rows and field_key-bound cells                                 |
|                                                                                |
| Bottom draft row appears inside the grid when ordinary creation is allowed.     |
+--------------------------------------------------------------------------------+
| Saved | Queue 0 | Presence 5 | More status ▾                                   |
+--------------------------------------------------------------------------------+
```

When opened explicitly, the inspector is mounted adjacent to the grid at base viewport and exposes Details, Relationships, Evidence, History, and specialized row actions without replacing ordinary grid editing.

*Design direction.* Inspector height MUST follow the shell-owned work area shown in the wireframe, not the number of rows currently rendered in the grid. Long inspector content MUST scroll inside the inspector while the grid scrolls independently.

### 5.7 Status-strip density budget

*Design direction.* The status strip MUST remain a capacity-limited working-state surface. It MUST NOT become a dashboard or management summary.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Slot                           | Required content                                                                                                    | Maximum visible content                            |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------- | -------------------------------------------------- |
| Primary state                  | One save-state label.                                                                                               | Exactly one of `Syncing`, `Saved`, or `Conflict`.  |
| Secondary same-surface message | Queue overflow, replay blocked, session expiry warning, pack degradation, transient confirmation, or keyboard hint. | One message.                                       |
| Presence summary               | Same-surface user presence summary.                                                                                 | Up to five visible avatars or initials, then `+N`. |

*Design direction.* The secondary slot priority MUST be queue overflow, replay blocked on non-retryable failure, same-surface session expiry or re-authentication required, pack degradation affecting the active surface, transient confirmation, then keyboard hint.

*Design direction.* Additional status messages MUST collapse into a same-surface overflow affordance labeled `More status` or an equivalent accessible name.

*Design direction.* Persistent shell chrome and the status strip MUST NOT show incident-level KPIs, time-to-resolution counters, team throughput, external ticket counts, or management dashboard metrics. Those metrics MAY appear inside explicitly opened workbook-native coordination, reporting, or status-review surfaces when otherwise allowed.

### 5.8 Authenticated root landing compositions

*Core behavior.* Core 01 owns authenticated `/` landing selection for every authentication kind. Core 03 owns workbook-startup selection only after Core 01 opens a workbook without an explicit launch `sheet_ref`.

*Design direction.* All rows in the following table inherit `*Design direction.*`; the table describes visible composition only and does not redefine the Core 01 selection algorithm.

| Visible incident state | Landing composition |
| ---------------------- | ------------------- |
| Zero visible incidents | Remain on `/` with an empty visible-incident directory. Present the ordinary create-incident affordance for an active authenticated account using the same placement and tone as the populated directory action. |
| Exactly one visible incident | Transition directly into the incident workbook. During bootstrap, show workbook-shell loading structure rather than a directory interstitial. If visibility disappears before bootstrap completes, return to `/` and render the current directory state. |
| Two or more visible incidents | Remain on `/` with a scan-friendly visible-incident directory. Present stable incident identity, title, status, and updated-time summaries, plus the ordinary create affordance when allowed. |

*Design direction.* Directory sorting, filtering, and presentation memory are directory-display concerns after the directory is shown; they are not incident-selection inputs.

*Design direction.* Incident-directory search, deployment-admin user-list search, and Reference Pack administration search MUST treat the server-side list route as authoritative. Client-side filtering over a partially loaded cursor collection is a local display refinement only and MUST NOT be presented as exhaustive. While a newer search or filter request is pending, the UI MUST keep the prior result set visible, display `Searching`, submit the current search immediately on Enter, and discard stale responses by a monotonically increasing client request sequence.

*Core behavior.* Core 04 owns the `deployment_admin` authorization matrix. Core 01 owns `GET /api/v1/extensions` claim-state discovery and reserved unclaimed extension-family behavior.

*Design direction.* Deployment-administration menu items MUST be rendered from the current session's `deployment_admin` status, the Core 04 matrix, and `GET /api/v1/extensions` claimed-state results. Incident membership controls MUST remain incident-role driven and MUST NOT be exposed solely because the current user holds `deployment_admin`.

*Design direction.* The Reference Packs administration group MUST be hidden when `GET /api/v1/extensions` reports `reference_pack.claimed=false`. When that profile is claimed but the current authenticated session lacks `deployment_admin`, direct navigation to a Reference Packs administration surface MUST render an authorization failure rather than an empty pack list, and the client MUST NOT route-probe `/api/v1/reference-packs/*` to infer pack state for a non-admin.

*Core behavior.* Core 01 §3.3.2.1B owns the `/deployment-administration` browser context, panel availability, and aggregate-administration prohibitions. Core 01 §12.3.6 and §17.5 own imported-incident bootstrap access. Core 01 §17.4 owns Reference Pack list-query search and filters. Core 03 §2.4 owns the `Open imported incident` startup boundary. Core 04 §2 owns direct-navigation denial and capability-loss behavior.

*Design direction.* The global Deployment administration entry MUST appear only as a menu item labeled `Deployment administration` in the upper-right account/application menu. That menu MUST remain reachable from the incident directory and workbook shell, retain visible focus and active states, and collapse lower-priority workbook query controls before it under responsive pressure.

*Design direction.* Deployment administration MUST use a distinct deployment-local administration layout, not the workbook grid shell. The default selected panel MUST be `Users`, and panel navigation MUST expose visible focus state, active selected state, and accessible labels matching the panel names in the following table.

| Panel | Availability | UI contents |
| ----- | ------------ | ----------- |
| Users | Base. | User list, create/update controls, password reset, TOTP reset, session revocation, and safe user details. |
| Administrative audit | Base. | Deployment-scoped administrative audit events with server-owned filters. |
| Reference packs | `reference_pack` claimed. | Pack lifecycle controls and a search field labeled `Search reference packs`. |
| Incident import | `incident_portability` claimed. | Whole-incident bundle jobs, progress/error states, and successful-result action `Open imported incident`. |
| Enterprise authentication bindings | `enterprise_authentication` claimed. | Selected-user binding operations only. |

*Design direction.* Reference Pack search MUST clear or supersede stale pagination state when search or filters change, preserve the prior accepted result set while `Searching`, reject stale responses by client request sequence, and bind accepted pagination only to the accepted query state.

*Design direction.* The Incident import panel MUST show `Open imported incident` only after a successful import result with an imported incident target. Activating it MUST leave Deployment administration and open the imported incident with no explicit workbook `sheet_ref`.

*Design direction.* Deployment administration MUST NOT expose `General settings`, all-incident catalog/search/count/metadata controls, generic cross-incident policy-default editors, provider-definition editors, provider-wide recovery controls, incident membership controls driven only by `deployment_admin`, initial-admin selectors, adoption routes, portable-membership editors, or historical-actor membership mapping controls.

### 5.9 Account settings composition

*Core behavior.* Core 01 §3.3.2.3 owns `/api/v1/account/profile` and `/api/v1/account/preferences`; Core 04 §2 owns current-user-only authorization. This guide describes UI organization only.

*Design direction.* Account settings SHOULD use three top-level areas:

| Area | Current-profile controls |
| ---- | ------------------------ |
| Profile | Read-only email plus display-name edit. |
| Appearance | One density selector with `Use surface default`, `Compact`, `Default`, and `Comfortable`. |
| Security | Entrypoints to existing password-change and TOTP setup or replacement flows. |

*Design direction.* The Profile area MUST NOT present email or login identifier as self-service editable. The Appearance area MUST NOT present theme selection, custom density tokens, custom row height, locale, time-zone, notification, global default incident, or global `home_sheet_ref` controls. The Security area MUST route to existing `/api/v1/auth/*` behavior and MUST NOT imply new account-profile security routes.

### 5.10 Administrative audit views

*Core behavior.* Core 01 §3.3.5.1A owns administrative audit routes, resource shape, filters, pagination, and cursor binding. Core 04 §§2-3 own authorization, redaction, retention, and denial behavior.

*Design direction.* Account and deployment audit views MUST use the server filters declared by Core 01 rather than free-text search. The UI MUST NOT present a client-side search box as exhaustive over a partially loaded cursor collection.

*Design direction.* Redacted audit values MUST be visually distinct from visible JSON `null` values. Redaction presentation MAY use a stable label, icon, or badge, but it MUST preserve the event row, field path, and before/after positions so reviewers can see that a sensitive field changed.

*Design direction.* Deployment audit navigation MUST follow the current session's `deployment_admin` status and the Core 04 matrix. Incident membership audit navigation MUST follow the current incident role `admin` state and MUST NOT be exposed solely because the current user holds `deployment_admin`.

### Acceptance criteria

- **R2-AC-017:** §5.2 defines primary tabs, the `System views` switcher, optional surfaces, and saved-view placement.
- **R2-AC-018:** Built-in tabs are ordered Timeline, Hosts, Identities, Evidence, Notes.
- **R2-AC-019:** Required system views are grouped and ordered exactly as §5.3 states.
- **R2-AC-020:** Saved views appear in the active surface’s view selector, not as primary tabs.
- **R2-AC-021:** The guide rejects all-fourteen-tabs, command-palette-only system-view access, and separate coordination modules for the base design.
- **R2-AC-022:** The wireframe and §5.1 show the tab strip, system-view switcher, compact sheet toolbar, grid, explicit inspector opener, status strip, and grid-first Timeline default with no dominant administration card stack above the active grid.
- **R2-AC-023:** §5.7 defines exactly three status-strip slots.
- **R2-AC-024:** The primary slot permits exactly one save-state label.
- **R2-AC-025:** Overflow uses a same-surface overflow affordance rather than silently dropping messages.
- **R2-AC-026:** KPI and dashboard metrics are banned from persistent shell chrome but not over-banned from opened coordination or reporting surfaces.
- **R2-AC-097:** §5.8 defines zero-, one-, and many-incident landing compositions while deferring selection behavior to Core 01.
- **R2-AC-098:** §5.8 states that incident-directory, deployment-admin user-list, and Reference Pack administration search use authoritative server-side search, preserve prior visible results while pending, show `Searching`, submit immediately on Enter, and discard stale responses.
- **R2-AC-099:** §5.1 allows optional incident-create metadata behind compact disclosure without hiding the post-create reviewer/admin controls for `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`.
- **R2-AC-100:** §5.1 distinguishes TLP canonical machine tokens from labels and severity or phase suggestions from closed vocabularies.
- **R2-AC-101:** §5.1 defines the persistent `Closed, read-only` shell state and separates it from save-state labels.
- **R2-AC-102:** §5.1 disables or hides source-state write affordances while retaining allowed read and reporting actions for closed incidents.
- **R2-AC-105:** §5.8 derives deployment-administration menu items from the Core 04 matrix plus extension claimed state and keeps incident membership controls incident-role driven.
- **R2-AC-106:** §5.9 defines Account settings as Profile, Appearance, and Security while omitting unsupported profile, theme, global-home, custom-density, row-height, and new security-route controls.
- **R2-AC-107:** §5.10 requires administrative audit views to use server filters rather than free-text search, visually distinguish redacted values from visible JSON `null`, and derive deployment versus membership-audit navigation from the Core 04 authorization split.
- **R2-AC-108:** §5.8 places Deployment administration in the upper-right account/application menu, defines allowed panels, states focus/active/responsive behavior, and prohibits aggregate administration controls.
- **R2-AC-109:** §5.8 defines `Search reference packs` stale-response handling and the successful-import `Open imported incident` action without introducing an initial-admin selector or explicit workbook `sheet_ref`.

## 6. Workbook Surface Model

### 6.1 Built-in tabs

*Core behavior.* Core 03 requires the base profile to expose Timeline, Hosts, Identities, Evidence, and Notes as built-in tabs.[^4]

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Built-in tab | `view_schema_id`               | UX role                                                                  |
| ------------ | ------------------------------ | ------------------------------------------------------------------------ |
| Timeline     | `cartulary.view.timeline.v2`   | Primary rough-capture and chronology surface.                            |
| Hosts        | `cartulary.view.hosts.v1`      | Canonical and stub host records.                                         |
| Identities   | `cartulary.view.identities.v1` | Canonical and stub identity records.                                     |
| Evidence     | `cartulary.view.evidence.v1`   | Evidence envelopes, blob attachment state, preview, and custody signals. |
| Notes        | `cartulary.view.notes.v1`      | Artifact-backed note capture in the built-in tab family.                 |

### 6.2 Canonical required system views

*Core behavior.* Core 03 and Core 01 close the current-profile workbook surface registry. The fourteen pack-independent base-profile surface IDs are the five built-in tabs plus nine required system views.[^4][^5]

*Design direction.* §6.2 is the only guide-local canonical enumeration of required system views. Later sections MAY add UX posture but MUST NOT duplicate canonical identity as an independent authority.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Surface label      | `view_schema_id`                  | Primary UX role                                                 |
| ------------------ | --------------------------------- | --------------------------------------------------------------- |
| Indicators         | `cartulary.view.indicators.v1`    | Canonical indicators with pivots to observations and lifecycle. |
| Assessments        | `cartulary.view.assessments.v1`   | Incident-scoped assessment history.                             |
| Task Requests      | `cartulary.view.task_requests.v1` | Queue-oriented owned work.                                      |
| Decisions          | `cartulary.view.decisions.v1`     | Rationale-bearing decision records.                             |
| Parties            | `cartulary.view.parties.v1`       | Stable incident-scoped coordination identities.                 |
| Communications Log | `cartulary.view.comm_log.v1`      | Durable communication memory.                                   |
| Handoff            | `cartulary.view.handoff.v1`       | Shift or phase continuity.                                      |
| Status Review      | `cartulary.view.status_review.v1` | Checkpoint review and rebalancing.                              |
| Lesson             | `cartulary.view.lesson.v1`        | Retrospective follow-through.                                   |

*Design direction.* `Compromise Assessments` is the documentation-canonical surface label for `cartulary.view.assessments.v1`. Constrained UI labels MAY use `Assessments` as display shorthand only; no semantic distinction is intended.

### 6.3 Optional standardized workbook surfaces

*Core behavior.* Core 04 acceptance criteria and Core 01 registry closure allow only Findings, Investigative Queries, and Forensic Keywords as standardized optional artifact-backed workbook surfaces beyond the fourteen required surfaces when implemented.[^5]

*Design direction.* When exposed, the optional standardized surfaces MUST inherit the same shell, row, query, filter, grouping, saved-view, history, and inspector grammar as other workbook surfaces.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Optional surface      | `view_schema_id`                          | Required posture when exposed                                             |
| --------------------- | ----------------------------------------- | ------------------------------------------------------------------------- |
| Findings              | `cartulary.view.findings.v1`              | Workbook-native surface, not a separate hypothesis module.                |
| Investigative Queries | `cartulary.view.investigative_queries.v1` | Workbook-native surface with the same queue, filter, and history grammar. |
| Forensic Keywords     | `cartulary.view.forensic_keywords.v1`     | Workbook-native surface with the same queue, filter, and history grammar. |

### 6.4 Saved views

*Core behavior.* Core 03 defines saved views as incident-bound workbook configurations over exactly one `view_schema_id`. Saved-view scope controls discoverability and mutability of the saved-view object only.[^4]

*Design direction.* Saved views are surface configurations, not storage silos and not canonical workbook-surface identities.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Scope     | Discoverability                 | In-place mutability               | Ordinary create default         |
| --------- | ------------------------------- | --------------------------------- | ------------------------------- |
| `private` | Owner and incident admins only. | Owner and incident admins.        | Default.                        |
| `shared`  | All incident members.           | Owner and incident admins.        | Allowed on ordinary create.     |
| `system`  | All incident members.           | Immutable through ordinary paths. | Not allowed on ordinary create. |

*Core behavior.* Saved views persist portable shared layout and query state. Selection, scroll position, focused cell, local popover state, open inspector state, preview state, and presence remain client-local and are not saved-view state.[^4]

### 6.5 Startup and default surface behavior

*Core behavior.* Core 03 defines the workbook-startup fallback chain: explicit launch `sheet_ref`, caller `home_sheet_ref`, incident `default_sheet_ref`, then Timeline.[^4]

*Design direction.* All rows in the following table inherit `*Core behavior.*` because the table restates owner behavior.

| Order | Startup source                                     | Rule                                           |
| ----- | -------------------------------------------------- | ---------------------------------------------- |
| 1     | Explicit launch `sheet_ref`.                       | Use if present and still valid for the caller. |
| 2     | `user_workbook_preferences.home_sheet_ref`.        | Use if present, valid, and visible.            |
| 3     | `incident_workbook_preferences.default_sheet_ref`. | Use if present, valid, and visible.            |
| 4     | Timeline.                                          | Final fallback.                                |

*Core behavior.* A `sheet_ref` of kind `view_schema` names a canonical required surface. A `sheet_ref` of kind `saved_view` names one distinct saved-view object over a schema. If a pointer is missing, hidden, invalid, or depends on an unavailable optional pack, Core 03 clears it and continues down the fallback chain.[^4]

### Acceptance criteria

- **R2-AC-027:** §6.2 is the only guide-local canonical required-system-view enumeration.
- **R2-AC-028:** The five built-in tabs appear with their `view_schema_id` values.
- **R2-AC-029:** The nine required system views appear with their `view_schema_id` values.
- **R2-AC-030:** §6.3 lists only Findings, Investigative Queries, and Forensic Keywords as optional standardized workbook surfaces.
- **R2-AC-031:** §6.4 states that saved views are configurations, not storage silos or canonical identities.
- **R2-AC-032:** §6.5 states the four-step startup fallback order.

## 7. Grid Editing Model

### 7.1 Direct typing

*Core behavior.* Core 03 requires the primary interaction surface to support inline editing, keyboard navigation, paste, low-friction row creation, and saved or system views over projections.[^4]

*Design direction.* Selecting a cell and typing MUST edit it immediately. The user MUST NOT have to enter a separate form edit mode for the common row-editing path.

*Design direction.* The trailing blank row SHOULD be part of the default grid model. Typing into it SHOULD create a real row as soon as the active surface’s minimum create signal is satisfied after create-time normalization.

*Design direction.* Non-Timeline surfaces MAY require their declared minimum create signal, but they SHOULD preserve same-surface inline creation rather than redirecting to a modal or form page.

*Design direction.* Relationship cells MUST accept raw typing. They MUST NOT require picker-first interaction.

### 7.2 Keyboard contract

*Design direction.* The current keyboard contract is compact and intentionally memorable. All rows in the following table inherit `*Design direction.*`.

| Key         | Required default effect                                                                                                                  | MUST NOT become                                                              |
| ----------- | ---------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Arrow keys  | Move grid selection.                                                                                                                     | A hidden macro language.                                                     |
| Enter       | Commit and move vertically.                                                                                                              | A form-submit detour.                                                        |
| Shift+Enter | Reverse vertical navigation.                                                                                                             | A second editing mode.                                                       |
| Tab         | Commit and move horizontally.                                                                                                            | A shell navigation shortcut.                                                 |
| `Ctrl+V`    | Paste into the visible grid.                                                                                                             | An import-only gesture.                                                      |
| `Ctrl+K`    | Quick link or resolve on the current cell.                                                                                               | A full-screen workflow jump.                                                 |
| `Space`     | Preview linked evidence for the selected row.                                                                                            | A browser scroll side effect.                                                |
| `Alt+H`     | Open history for the selected row.                                                                                                       | A detached review page.                                                      |
| `Esc`       | Apply the focused interaction priority ladder in §7.4: editor-local popup close, unsaved cell-edit discard, inspector close, then no-op. | A hidden autosave, broad destructive discard, or cross-context escape hatch. |

### 7.3 Paste and bulk editing

*Design direction.* Multi-cell paste, fill-down, and multi-row tag assignment are required workbook behaviors in the base design. They MUST NOT rely on hidden macro semantics.

*Design direction.* One primary click on a writable committed scalar cell enters edit mode, focuses the primary control, and places a collapsed caret after the existing text so typing appends by default. Read-only cells remain selectable, while chips and embedded action controls retain their own behavior. The fill handle is an accent-colored, labeled fill affordance rather than an edit affordance; `Ctrl/Cmd+D` provides the keyboard equivalent, and double-click fill-to-end is not supported.

*Design direction.* Clipboard paste is part of the base hot path. It SHOULD accept TSV and CSV copied directly from spreadsheet tools, create additional rows automatically when the pasted range exceeds the existing visible set, and preserve selection and row identity rather than turning paste into an import wizard.[^4][^8]

### 7.4 Autosave, save states, and `Esc`

*Design direction.* The normal grid workflow has no explicit Save button. Autosave MUST occur on Enter, Tab, blur, and paste completion.

*Core behavior.* Core 03 maps save-state presentation to exactly `Syncing`, `Saved`, and `Conflict`. Presence updates alone never change the save-state label.[^4]

*Design direction.* All rows in the following table inherit `*Design direction.*` while restating owner vocabulary.

| Label      | Exact meaning                                                                                                                                                                                         | Required user interpretation                                                       |
| ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `Syncing`  | At least one workbook mutation is in flight, the local pending queue is non-empty, or replay is paused waiting for connectivity recovery, re-authentication, or an HTTP re-query required by Core 03. | The system is preserving work, but authoritative reconciliation is still underway. |
| `Saved`    | No mutation is in flight, the local pending queue is empty, and no unresolved same-field local drafts exist.                                                                                          | Current workbook state is synchronized.                                            |
| `Conflict` | A same-field conflict is unresolved, queue overflow refused admission, or replay halted on a non-retryable failure.                                                                                   | Analyst attention is required.                                                     |

*Core behavior.* The local pending queue durability boundary is defined in §9.5: it does not survive full reload, tab close, cross-tab transfer, browser or application restart, or crash.[^4]

*Design direction.* `Esc` is an interaction-priority ladder, not a global discard. All rows in the following table inherit `*Design direction.*`.

| Priority | Focus and state condition                                                              | Required effect                                                                                                 |
| -------- | -------------------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------- |
| 1        | A focused cell editor has an open editor-local picker, autocomplete, menu, or popover. | `Esc` closes only that editor-local popup.                                                                      |
| 2        | A focused cell editor has uncommitted text and no editor-local popup is open.          | `Esc` discards only that cell editor draft and restores the pre-edit displayed cell value.                      |
| 3        | Focus is outside a cell editor and the inspector is open.                              | `Esc` closes the inspector and returns focus to the prior grid cell when that cell still exists and is visible. |
| 4        | None of the above conditions apply.                                                    | `Esc` is a no-op for workbook state.                                                                            |

*Design direction.* `Esc` is not an autosave trigger. `Esc` MUST NOT cancel already queued replay units, committed authoritative changes, or unresolved same-field conflict objects.

### Acceptance criteria

- **R2-AC-033:** Selecting a cell and typing edits the cell immediately.
- **R2-AC-034:** Relationship cells accept raw typing and do not force picker-first interaction.
- **R2-AC-035:** `Ctrl+V`, fill-down, and multi-row tag assignment work on the visible workbook surface.
- **R2-AC-036:** The visible save-state labels are exactly `Syncing`, `Saved`, and `Conflict`.
- **R2-AC-037:** §7.4 `Syncing` includes replay paused for connectivity recovery, re-authentication, or HTTP re-query.
- **R2-AC-038:** §7.4 states that presence updates alone do not change save-state labels.
- **R2-AC-039:** §7.2 and §7.4 define the same `Esc` behavior and priority order.

## 8. Progressive Structuring and Relational UX

### 8.1 Capture first, structure later

*Core behavior.* Core 02 separates raw capture from canonical entities and preserves original rough capture even after later normalization or resolution.[^6]

*Design direction.* The UI MUST preserve a rough-capture path for timeline rows, notes, relationship cells, and evidence-adjacent text. The user MUST be able to record uncertain facts first and normalize them later.

*Design direction.* The UI SHOULD expose progressive structuring as a visible state transition: raw token, unresolved chip, resolved chip, auto-resolved chip, dismissed mention, or canonical relationship value.

### 8.2 Binding modes and UI consequences

*Core behavior.* Core 02 defines `mention_origin` and `entity_origin` as binding modes. Binding behavior depends on the contract, not on the visible column header.[^6]

*Design direction.* All rows in the following table inherit `*Design direction.*` while restating owner-owned binding consequences.

| UI context                                                   | Binding mode                           | UI consequence                                                                       |
| ------------------------------------------------------------ | -------------------------------------- | ------------------------------------------------------------------------------------ |
| Timeline or Notes cells that reference hosts or identities.  | `mention_origin`.                      | Raw text creates source-bound mentions and MAY show suggestions.                     |
| Clipboard paste into non-entity sheets.                      | `mention_origin`.                      | Preserve one mention per observed value and source position.                         |
| Inspector `Resolve` action.                                  | `mention_origin`.                      | Resolve the selected mention and keep raw text inspectable.                          |
| Inspector `Create host` or `Create identity` from a mention. | `mention_origin` plus explicit create. | Create or reuse a target through explicit action, then resolve the selected mention. |
| Hosts or Identities direct row creation.                     | `entity_origin`.                       | Create or upsert a canonical or stub entity record.                                  |

### 8.3 Chip semantics

*Design direction.* The chip-state vocabulary is closed for Revision 2. §8.3 and §16.1 MUST use the same four states: `unresolved`, `resolved`, `auto_resolved`, and `dismissed`.

*Design direction.* Manual resolution is metadata on the `resolved` state, not a separate chip state. A manually resolved chip MAY expose a non-color marker when provenance or review requires a visible distinction.

*Design direction.* Unresolved chips MUST be visually distinct from resolved chips through a combination of border treatment and an inline state marker. The default unresolved treatment is dashed or dotted border plus a leading `?` marker or visible `Unresolved` label. Color difference alone MUST NOT be the sole distinguishing signal.

*Design direction.* Auto-resolved chips MUST show the `auto` marker defined in §16.1 until the user explicitly changes or reverts the resolution.

*Design direction.* Dismissed mentions MAY appear only where they remain inspectable. When shown, they MUST carry a visible dismissed marker and accessible name.

### 8.4 Inspector contract

*Design direction.* The inspector is an adjacent, secondary surface for enrichment, relationships, evidence, history, and destructive actions. It MUST NOT replace direct grid editing for ordinary capture.

*Design direction.* The inspector MUST be closed by default on workbook surfaces. Selecting a row or cell MUST select or focus only; it MUST NOT open the inspector unless the user activates an explicit inspector, row action, history, mention, evidence, or equivalent command.

*Design direction.* The inspector owns detail editing that would clutter the grid, relationship review, evidence preview and attachment detail, row history, rollback affordances, and destructive confirmations.

*Core behavior.* Core 01 §7.4 defines inspector configuration, feature registry, and route owner metadata as `view_schema_id` metadata; Core 03 §2.3A defines deterministic inspector workflow algorithms; Core 03 requires saved views to inherit that config from immutable `view_schema_id`; and Core 03 requires no-row state to render `no_row_selected`.

*Design direction.* The inspector MUST render panels in the active config order. When every current-profile panel is declared, the order is Details, Relationships, Evidence, History, Workflow. The UI MUST NOT infer behavior from visible labels, component names, route-helper names, CSS selectors, storage tables, or grid-library APIs.

*Design direction.* The `no_row_selected` empty state MUST be compact, action-neutral, and free of stale row details, stale confirmations, previews, merge plans, rollback previews, and local forms.

*Design direction.* When row, row version, authorization, incident lifecycle, deletion, merge, or active surface changes, stale row-bound inspector content MUST clear before replacement content paints. Destructive confirmations SHOULD name the affected record or records, place focus on the confirmation surface when opened, and restore focus to the invoking control or selected grid cell when dismissed.

*Design direction.* The Workflow panel MUST carry heavy same-shell actions that would slow ordinary data entry if placed on the grid: create related task/decision/coordination records, merge planning, supersede forms, rollback previews, evidence-access detail, and party-linking forms. It MUST NOT contain required per-edit checklists, hidden approvals, dashboard metrics, release gates, or external enrichment in the base profile.

*Design direction.* The inspector MUST NOT become a full-page record editor, dashboard, ticketing module, release-control module, or hidden source of saved-view state.

### 8.5 Auto-resolution disclosure and correction

*Design direction.* Auto-resolution is useful only when disclosure and recourse are visible. Auto-resolved chips MUST remain distinguishable from manually resolved chips until the user explicitly changes or reverts them.

*Design direction.* Batch paste MAY produce auto-resolved chips when owner contracts allow deterministic exact-match reuse. The UI MUST expose a same-surface transient confirmation and an inspectable route to review affected cells.

### Acceptance criteria

- **R2-AC-040:** §8.3 and §16.1 use the same four chip states.
- **R2-AC-041:** Unresolved and resolved chips differ by more than color.
- **R2-AC-042:** Auto-resolved chips remain inspectably marked after transient disclosure fades.
- **R2-AC-043:** Dismissed mentions, when displayed, carry a visible dismissed marker and accessible name.
- **R2-AC-044:** The inspector is described as adjacent and secondary, not a replacement for grid editing.

## 9. Collaboration, Presence, Autosave, and Conflict Resolution

### 9.1 Collaboration model

*Core behavior.* Core 03 requires field-level optimistic concurrency on top of row versioning. Each visible row is bound to `record_id` and `row_version`, and every grid write includes `record_id`, `base_row_version`, and changed fields only.[^4]

*Design direction.* Collaboration presentation MUST make row identity and conflict scope legible. The UI MUST never imply that row position, sort position, visible label, or display value is the mutation target.

*Core behavior.* Different-field concurrent edits on the same row auto-rebase and accept. Same-field concurrent edits reject with an explicit conflict payload and remain unresolved until analyst action.[^4]

*Design direction.* Same-field conflicts MUST be shown as cell-local state. A conflict on one cell MUST NOT freeze the entire row, sheet, or workbook.

### 9.2 Presence

*Core behavior.* Core 03 defines presence as collaboration state and live row-update behavior. Presence is not save state.[^4]

*Design direction.* Presence SHOULD appear at three levels when available: header-level same-surface presence, row-gutter row focus, and cell-level same-field editing indication.

*Design direction.* Presence indicators MUST NOT lock editing. Presence MAY warn of likely collision but MUST NOT become a hidden edit lock.

### 9.3 Same-field conflict resolution

*Core behavior.* Core 03 requires the resolver to open from the conflicted cell, keep the main grid visible, display row context, field display label, stable `field_key`, saved value with actor and timestamp, analyst local value, and resolution actions. Initial focus does not default to a destructive action.[^4]

*Design direction.* The resolver SHOULD use an adjacent drawer or same-surface panel. A modal resolver is reserved for narrow cases where the drawer cannot preserve the required context.

*Design direction.* Closing the resolver without selecting a resolution MUST leave the cell in conflict state. After an explicit resolution or clear action, focus MUST return to the same cell and scroll position SHOULD be preserved.

### 9.4 Authorship, row history, and rollback

*Core behavior.* Core 01 and Core 03 define row-centric history and rollback target kinds. Available rollback actions draw only from `history_entry`, `change_set`, and `row_restore`.[^4][^5]

*Core behavior.* Core 03 requires inspector Details, Relationships, Evidence, History, and row-local actions to share one active `record_id` subject. When history is open and the active saved row changes, the History section retargets to the new active row rather than continuing to display the prior row's history.

*Core behavior.* The rollback target kinds have different reversal scopes. `history_entry` reverses one individually reversible mutation target. `change_set` reverses every reversible mutation entry in the addressed `change_set` in reverse deterministic order. `row_restore` restores only the authoritative row-backed fields of the selected record revision. Whole-row restore does not implicitly recreate, delete, or repoint non-row-backed associations such as `record_links`, `record_tags`, `entity_mentions`, `indicator_observations`, or evidence associations.[^5][^6]

*Design direction.* When the History section retargets, the inspector SHOULD preserve the drawer and show loading, empty, or public error state in the History section for the active row. It MUST NOT present rollback, delete, restore, or confirmation controls for a stale row.

*Design direction.* The history UI MUST present those three rollback target kinds as different actions, not as aliases for “restore this row.” Whole-row restore affordances MUST label the row-field-only scope before confirmation.

### 9.5 Local pending queue, overflow, and reauthentication

*Core behavior.* Core 03 defines the local pending queue as a same-runtime recovery mechanism for transient transport failure, auth failure on queued write, and `session_revoked` in the same runtime. It does not define reload durability, tab-close survival, cross-tab transfer, restart survival, or crash survival.[^4]

*Core behavior.* Core 03 treats `incident_closed` as terminal for queued or unsent source mutations. Those rejected drafts can remain locally visible and copyable, but they are not authoritative and must not auto-replay while closed or after reopen.[^4]

*Design direction.* All rows in the following table inherit `*Design direction.*` while restating the owner boundary.

| Property           | Base-profile UI requirement                                                                                                                                                                                           |
| ------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Purpose            | Preserve same-runtime user work across transient replay conditions.                                                                                                                                                   |
| Durability         | Survives transient transport failure, auth failure on queued write, and `session_revoked` in the same runtime; does not survive full reload, tab close, cross-tab transfer, browser or application restart, or crash. |
| Overflow           | Refuses admission visibly and moves save state to `Conflict` or an equivalent attention state.                                                                                                                        |
| Re-authentication  | Preserves queued work in the same runtime while the user completes required re-authentication.                                                                                                                        |
| Closed incident    | Stops source-mutation replay, marks affected queued work as rejected local drafts, and requires a fresh user action after reopen before any retained draft can be submitted.                                          |
| Cross-tab behavior | No cross-tab replay or queue-transfer guarantee is implied.                                                                                                                                                           |

*Design direction.* Cross-tab transfer is named explicitly because general-purpose client-state libraries often synchronize or rehydrate state across tabs. Cartulary’s base-profile pending queue MUST NOT be inferred to survive or replay through such cross-tab mechanisms.

*Design direction.* A draft rejected because the incident closed SHOULD stay near its original row or cell context when that context still exists. The UI SHOULD offer copy or discard actions and SHOULD avoid presenting the draft as pending authoritative work. Reopen MUST NOT automatically restart those rejected mutations; the user must make a fresh edit or explicit submit action against reopened current state.

### Acceptance criteria

- **R2-AC-045:** §9.1 states field-level optimistic concurrency over row versioning.
- **R2-AC-046:** Same-field conflict is cell-local and does not freeze the whole workbook.
- **R2-AC-047:** §9.4 distinguishes `history_entry`, `change_set`, and `row_restore` scopes.
- **R2-AC-048:** Whole-row restore is described as row-backed fields only.
- **R2-AC-049:** §7.4 and §9.5 both name cross-tab transfer as a pending-queue non-survival condition.
- **R2-AC-050:** §9.5 states that pending queue durability is same-runtime only.
- **R2-AC-103:** §9.5 states that `incident_closed` turns queued source mutations into rejected local drafts, not retryable pending work.
- **R2-AC-104:** §9.5 states that reopen does not automatically replay drafts rejected by closure.

## 10. Sorting, Filtering, Grouping, and View State

### 10.1 Sorting and filtering

*Core behavior.* Core 01 owns the view-query route, sort entries, filter predicates, canonicalization, cursor binding, and limits. The UI addresses fields by stable `field_key`, not visible labels or storage names.[^5]

*Design direction.* Sorting and filtering controls MUST expose visible labels, but their configured state MUST serialize and replay through stable contract identifiers. Renaming a column label MUST NOT change filter semantics, write-back behavior, or export semantics.

*Design direction.* Workbook query controls are top-bar owned. Filter fields live inside a draft popover opened from the `Filters` control; `Esc` cancels draft changes, and `Apply` commits one canonical query intent. Saved-view actions are menu-owned, and save state is exposed through the shared status strip rather than detached surface badges.

*Design direction.* The UI SHOULD preserve the user’s current working context when sort or filter changes. Pending edits remain bound to `record_id`, not row position.

### 10.2 Grouping boundary

*Core behavior.* Core 01 and Core 03 allow at most one active grouping key in the current profile. Group headers are presentation artifacts and are not writable rows.[^4][^5]

*Design direction.* Grouping MUST remain a view-state operation, not a data-model mutation. Dragging, expanding, or collapsing groups MUST NOT create, delete, or mutate incident records.

*Design direction.* When grouping is rendered through a treegrid pattern, group rows MUST be presented as navigation and summarization affordances, not ordinary incident records. Group rows MUST expose expand/collapse affordances, MUST be keyboard navigable, and MUST NOT expose ordinary writable cell affordances.[^15]

*Design direction.* Copy from a group row MAY expose group-label or summary text when the implementation supports it. Paste, drag fill, editor entry, entity resolution, evidence attach, and destructive record actions MUST NOT be available on group rows.

*Design direction.* Group expansion state is client-local by default. It MAY become persistent only if an owner contract explicitly defines where that state is stored and how it participates in saved views or preferences.

### 10.3 View-state layers

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| State layer                    | Persistence                    | Examples                                                                                 |
| ------------------------------ | ------------------------------ | ---------------------------------------------------------------------------------------- |
| Contract state                 | Owner core.                    | `view_schema_id`, fields, sorting eligibility, filter eligibility, grouping keys.        |
| Saved-view state               | Saved view object.             | `query_json`, portable `layout_json`, display name, scope.                               |
| Client-local state             | Runtime only.                  | Selection, scroll, focused cell, open inspector, local popover, preview state, presence. |
| Account preference state       | Current-account preferences.   | Density override only.                                                                   |
| Incident user preference state | Workbook preferences.          | Home surface pointer.                                                                    |
| Incident default state         | Incident workbook preferences. | Default surface pointer.                                                                 |

*Design direction.* The UI MUST NOT persist client-local state into saved views. Saved views are portable surface configurations, not runtime snapshots.

### Acceptance criteria

- **R2-AC-051:** §10 states that sorting and filtering serialize through stable field identifiers.
- **R2-AC-052:** §10 states that grouping is view state and not a data mutation.
- **R2-AC-053:** §10.3 separates contract, saved-view, client-local, account preference, incident user preference, and incident default state.
- **R2-AC-054:** §10 says client-local state is not saved-view state.
- **R2-RDG-AC-001:** §10.2 states that grouped or treegrid rows are navigation and summarization affordances, not ordinary incident records.
- **R2-RDG-AC-002:** §10.2 forbids paste, drag fill, editor entry, entity resolution, evidence attach, and destructive record actions on group rows.
- **R2-RDG-AC-003:** §10.2 states that group expansion is client-local by default.

## 11. Coordination Surfaces as Workbook-Native UX

### 11.1 Why coordination belongs in the workbook

*Core behavior.* Core 03 requires Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson to remain workbook-native surfaces rather than separate application modules.[^4]

*Design direction.* The design reason is operational continuity: incident coordination is part of the shared case state and MUST stay visible next to the facts, evidence, and history that drive it.[^12][^13]

*Baseline context.* Operating-model guidance recommends using these surfaces for tracker hygiene, handoff quality, status-review cadence, workload redistribution, debrief follow-through, and challenge or escalation practice. That guidance does not define implementation conformance.[^13]

### 11.2 Surface roles

*Design direction.* For the canonical surface-identity list and primary role, see §6.2. This table adds UX posture and must-not-become direction for coordination surfaces.

| Surface reference  | UX posture                                                                                | MUST NOT become                        |
| ------------------ | ----------------------------------------------------------------------------------------- | -------------------------------------- |
| Task Requests      | Queue-oriented grid with owner, status, priority, due, blocked, and follow-through views. | A separate ticketing product.          |
| Decisions          | Review-oriented grid for rationale, owner, status, supersession, and decided-at posture.  | A generalized approval engine.         |
| Communications Log | Durable communication memory with party references and action links.                      | A chat client.                         |
| Handoff            | Continuity record for shift or phase boundaries.                                          | A mandatory ritual for every row edit. |
| Status Review      | Checkpoint and rebalancing surface.                                                       | Persistent dashboard chrome.           |
| Lesson             | Retrospective follow-through surface with linked tasks and evidence.                      | A detached knowledge-base editor.      |

### 11.3 Same-surface creation and linking

*Design direction.* The UI SHOULD allow row-anchored creation or linking of coordination records without losing the active grid context. A selected Timeline row MAY spawn a related Task Request, Decision, Handoff reference, Communication Log reference, or Lesson follow-up when the user has permission and the target contract allows it.

*Design direction.* Coordination creation MUST NOT become a mandatory step in ordinary row capture. Handoff, status review, challenge, escalation, and debrief practices belong at operational boundaries, not on every edit.[^13]

### 11.4 Operating guidance boundary

*Later scope.* Future operating-model guides MAY define recommended cadences or playbook practices for teams, but those practices MUST NOT be represented as Base Profile implementation-conformance requirements unless restated in Core 00 through Core 04.[^2][^13]

### Acceptance criteria

- **R2-AC-055:** §11.2 adds only UX posture and must-not-become direction, not a duplicate canonical surface registry.
- **R2-AC-056:** §11 states that coordination surfaces remain workbook-native.
- **R2-AC-057:** §11 rejects separate ticketing, chat, dashboard, approval-engine, and per-edit ritual interpretations.
- **R2-AC-058:** §11 separates operating-model guidance from implementation conformance.

## 12. Evidence Interaction Design

### 12.1 Evidence is not a cell path

*Core behavior.* Core 01 and Core 04 treat evidence access as application-mediated object access through authoritative incident membership and current evidence/blob state. User-supplied filenames and storage hints are metadata, not authority.[^5][^7]

*Design direction.* Evidence cells MUST NOT be raw object-store URLs, local file paths, or user-editable storage keys. The grid MAY show title, state, count, preview affordance, and linked evidence chips, but the access path remains application-mediated.

### 12.2 Two-step attachment and visible state

*Core behavior.* The evidence flow uses a two-step attachment model so incomplete uploads do not leave fake evidence attached. Attachment state remains visible through record and projection state.[^5][^14]

*Design direction.* Evidence attachment SHOULD feel like a workbook action: paste or drag onto selected row, visible pending or attached state, and same-surface preview when allowed. The UI MUST NOT navigate away from the grid for ordinary screenshot attachment.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| State                  | Required user interpretation                          | UI posture                                                           |
| ---------------------- | ----------------------------------------------------- | -------------------------------------------------------------------- |
| Requested              | Evidence is needed but not yet received.              | Show request state and owner/party context when present.             |
| Pending upload         | Blob slot exists but evidence is not final.           | Show pending state, not attached-as-complete.                        |
| Available              | Evidence is attached and access is requestable.       | Show preview/download affordance when allowed.                       |
| Preview blocked        | Evidence exists but preview is unsafe or unavailable. | Show explicit blocked state; do not silently collapse into download. |
| Failed or inconsistent | Evidence state cannot be trusted for access.          | Show error and require corrective action.                            |

### 12.3 Preview and download without leaving the workbook

*Design direction.* Preview opens in the inspector, an adjacent preview pane, or an equivalent same-shell surface. Download uses an application-mediated handle. The browser MUST NOT construct raw object-store URLs.

*Design direction.* Preview and download failures MUST be explicit. Unsupported, pending, quarantined, missing, failed, or inconsistent evidence state MUST NOT silently become a raw download path.

### 12.4 Reporting and release consequences

*Core behavior.* If the Snapshot and Reporting Extension Profile is implemented, recipient-specific withholding is applied at snapshot, render, and release time rather than by hiding live workspace content from authenticated incident participants.[^7]

*Design direction.* The live workbook MUST NOT present per-recipient visibility affordances such as disclosure-partition chips, recipient-selector controls, or release-scope badges during ordinary editing. Recipient-specific visibility is a snapshot, render, and release-time concern only. Moving those controls into the base live-editing surface would move extension-profile visibility controls into the base workbook.

### Acceptance criteria

- **R2-AC-059:** §12 states that evidence cells are not raw paths or raw object-store URLs.
- **R2-AC-060:** §12 defines visible attachment states for requested, pending, available, preview blocked, and failed or inconsistent evidence.
- **R2-AC-061:** §12 states that preview and download remain application-mediated.
- **R2-AC-062:** §12.4 prohibits per-recipient visibility affordances in ordinary live workbook editing.

## 13. Excel Inspiration: Preserve, Adapt, Reject

### 13.1 Preserve

*Design direction.* Cartulary SHOULD preserve the Excel-like behaviors that directly support incident work: direct typing, keyboard navigation, paste, fill-down, visible rows and columns, quick sorting and filtering, familiar grid scanning, and low ceremony for partial facts.[^9][^11]

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Preserve              | Required Cartulary interpretation                                    |
| --------------------- | -------------------------------------------------------------------- |
| Direct typing         | Grid cells remain ordinary editing targets.                          |
| Paste                 | Clipboard paste is a hot-path workbook action.                       |
| Sort and filter       | Contract-backed operations over stable `field_key` values.           |
| Familiar tabular scan | Few primary tabs, surface-specific saved views, and compact density. |
| Incomplete facts      | Rough capture remains valid and inspectable.                         |

### 13.2 Adapt

*Design direction.* Spreadsheet metaphors that would otherwise be ambiguous MUST be adapted to Cartulary’s auditable source model.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Spreadsheet pattern     | Cartulary adaptation                                                           |
| ----------------------- | ------------------------------------------------------------------------------ |
| Cell identity           | Bound to `record_id` plus `field_key`, never position alone.                   |
| Sheet tabs              | Built-in tabs plus system-view switcher, not unlimited primary tabs.           |
| Formula-like derivation | Contract-backed projections and derived fields.                                |
| Sheet sharing           | Incident membership, role model, WebSocket presence, and row-versioned writes. |
| Undo                    | Row history and scoped rollback, not opaque local undo stack.                  |

### 13.3 Reject

*Design direction.* Cartulary MUST reject spreadsheet behaviors that undermine case integrity.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Reject                                 | Reason                                                                    |
| -------------------------------------- | ------------------------------------------------------------------------- |
| Row-position mutation targeting        | Sorting and filtering would corrupt writes.                               |
| Silent same-cell overwrite             | Collaboration MUST preserve attribution and explicit conflict resolution. |
| Evidence paths in cells                | Evidence access MUST be authorized and application-mediated.              |
| Hidden formulas as business logic      | Behavior MUST follow explicit contracts.                                  |
| Per-user private truth                 | Shared incident facts MUST be authoritative and auditable.                |
| Dashboard chrome in the workbook shell | Persistent chrome MUST preserve hot-path work, not management reporting.  |

### Acceptance criteria

- **R2-AC-063:** §13 contains Preserve, Adapt, and Reject subsections.
- **R2-AC-064:** §13 rejects row-position mutation targeting, silent same-cell overwrite, raw evidence paths, hidden formulas as business logic, per-user private truth, and dashboard chrome.

## 14. Design Implications of the Current Implementation Baseline

### 14.1 Baseline scope

*Baseline context.* The current implementation baseline is one application unit, PostgreSQL, S3-compatible object storage, a Go backend, and a React plus `react-data-grid` plus Vite browser client.[^5][^8]

*Baseline context.* This section contains implementation-baseline consequences only. It does not restate canonical surface identity, saved-view scope, grouping keys, or Core behavior except by reference to owner sections.

*Baseline context.* The RDG-specific baseline row below is an adapter-consequence statement only. It does not make vendor package internals design authority over Cartulary workbook behavior.[^15]

*Baseline context.* All rows in the following table inherit `*Baseline context.*`.

| Baseline choice                                    | UX consequence                                                                                                                                                                                                                                                                                                       |
| -------------------------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| React browser client                               | The shell, grid, inspector, popovers, and status strip SHOULD be implemented as one browser workspace.                                                                                                                                                                                                               |
| `react-data-grid` through `/packages/grid-adapter` | Grid-first editing, keyboard, paste, virtualization, custom renderers, explicit editor adapters, grouping/treegrid behavior, frozen-column behavior, and CSS integration MUST remain compatible with the adapter contract. The UI guide MUST NOT assume vendor behavior that the adapter has not exposed and tested. |
| Vite                                               | The browser bundle SHOULD remain self-contained for deployed runtime assets.                                                                                                                                                                                                                                         |
| Go application unit                                | Evidence preview, download handles, API, WebSocket, and background jobs stay same-origin from the browser perspective.                                                                                                                                                                                               |
| PostgreSQL                                         | Authoritative structured state remains server-side, relational, and auditable.                                                                                                                                                                                                                                       |
| S3-compatible object storage                       | Binary evidence remains outside PostgreSQL and is accessed through application-mediated handles.                                                                                                                                                                                                                     |
| Permissive-license envelope                        | Design MUST avoid assuming commercial-grid features.                                                                                                                                                                                                                                                                 |
| No AG Grid Enterprise baseline                     | Required UX MUST be implementable without enterprise-only grid behavior.                                                                                                                                                                                                                                             |

### 14.2 Optional-pack degradation

*Baseline context.* Optional reference-pack activation MAY enrich labels, type registries, and overlays, but the base workbook MUST remain usable when optional packs are absent.[^5][^7]

*Design direction.* Pack degradation SHOULD appear as same-surface informational state. It MUST NOT block ordinary rough capture, grid editing, or base workbook navigation.

### 14.3 Clipboard hot path versus file-based import

*Core behavior.* Clipboard paste is part of the base workbook surface. File-based structured import beyond clipboard paste belongs to the Import Extension Profile and dedicated internal imports module.[^5][^8]

*Design direction.* The UI MUST NOT route ordinary clipboard paste through a file-import wizard. A file-import assistant MAY exist only under the Import Extension Profile and MUST NOT redefine the base workbook grammar.

### Acceptance criteria

- **R2-AC-065:** §14 contains only baseline-consequence content and cross-references non-baseline sections.
- **R2-AC-066:** §14 names React, `react-data-grid`, Vite, Go, PostgreSQL, S3-compatible object storage, permissive licensing, and absence of AG Grid Enterprise assumptions.
- **R2-AC-067:** §14 states that optional-pack degradation does not block rough capture or base workbook navigation.
- **R2-AC-068:** §14 states that clipboard paste is not a file-import wizard.
- **R2-RDG-AC-004:** §14.1 names `/packages/grid-adapter` as the `react-data-grid` integration boundary.
- **R2-RDG-AC-005:** §14.1 states that UI assumptions about vendor behavior must be adapter-exposed and tested.

## 15. Risks, Non-Goals, and Positive Patterns

### 15.1 Primary risk

*Design direction.* The primary UX risk is false equivalence: either building a spreadsheet clone without stronger guarantees or building a structured case tool that loses the speed and directness of the workbook model.[^11][^12]

*Design direction.* Reviewers SHOULD treat design proposals as suspect when they hide grid work behind forms, convert coordination surfaces into modules, convert status strip into a dashboard, or treat unresolved text as invalid input.

### 15.2 Concrete failure modes and guardrails

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Failure mode               | Guardrail                                                  |
| -------------------------- | ---------------------------------------------------------- |
| Forms-first drift          | Common capture and correction remain on the grid.          |
| Picker-first capture loss  | Relationship cells accept text before canonical selection. |
| Sheet-wide conflict freeze | Same-field conflict stays cell-local.                      |
| Hidden sync failure        | Save-state labels remain visible on the hot path.          |
| Dashboard sprawl           | Persistent chrome and status strip stay capacity-limited.  |
| Module fragmentation       | Coordination surfaces remain workbook-native.              |
| Release-control leakage    | Per-recipient controls stay out of ordinary live editing.  |
| Client-state overreach     | Pending queue remains same-runtime only.                   |

### 15.3 Non-goals

*Later scope.* The base UI guide does not claim a mobile-first or touch-first product profile, a dedicated design-token package, a component-library artifact, a Figma source of truth, or extension-profile-specific UX documents.

*Later scope.* The base UI guide does not define field-level ACL UX, generalized approval workflows, cross-incident analytics, workflow-engine UX, generalized ticketing UX, or pack-dependent framework overlays as workbook-native base surfaces.

*Later scope.* The base UI guide does not define self-service email or login-identifier changes, locale, time-zone, notification settings, theme selection, global default incident, global `home_sheet_ref`, custom density values, or custom row-height controls.

### 15.4 Positive patterns

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Positive pattern                                    | Required reviewer question                                                                 | Protected failure mode            |
| --------------------------------------------------- | ------------------------------------------------------------------------------------------ | --------------------------------- |
| Grid-primary with inspector-secondary               | Does the common path remain on the grid?                                                   | Forms-first drift.                |
| Text-first relationship cells with chip progression | Can rough text be captured before canonical selection?                                     | Picker-first capture loss.        |
| Same-surface evidence preview                       | Can the user preview without leaving the workbook shell?                                   | Detached media workflow.          |
| Visible save state on the hot path                  | Does the user know whether work is syncing, saved, or conflicted?                          | Hidden sync failure.              |
| Cell-local same-field conflict                      | Is only the affected cell in conflict state?                                               | Sheet-wide freeze.                |
| Workbook-native coordination                        | Do tasks, decisions, handoffs, and lessons remain surfaces inside the shell?               | Module fragmentation.             |
| Narrow contract-backed grouping                     | Is grouping governed by declared keys and one active key?                                  | Dashboard/report drift.           |
| Row-anchored coordination action                    | Can a user create or link coordination records from a selected row without losing context? | Context loss during coordination. |

### Acceptance criteria

- **R2-AC-069:** §15 names false equivalence as the primary UX risk.
- **R2-AC-070:** §15.2 includes failure modes and guardrails for forms-first drift, picker-first capture loss, sheet-wide conflict freeze, hidden sync failure, dashboard sprawl, module fragmentation, release-control leakage, and client-state overreach.
- **R2-AC-071:** §15.4 exists with at least the eight required positive patterns.
- **R2-AC-072:** Each positive-pattern row names both a reviewer question and the failure mode it protects against.

## 16. Visual Language Direction

### 16.1 Chip states

*Design direction.* The chip-state vocabulary is closed for Revision 2. All rows in the following table inherit `*Design direction.*`.

| State           | Border                                                      | Required visible marker                           | Accessible name pattern                                              | Notes                                                                     |
| --------------- | ----------------------------------------------------------- | ------------------------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `unresolved`    | Dashed or dotted border.                                    | Leading `?` marker or visible `Unresolved` label. | `Unresolved <entity type> mention: <raw text>`.                      | MUST differ from ordinary text cells and resolved chips.                  |
| `resolved`      | Solid border.                                               | No unresolved marker.                             | `Resolved <entity type>: <display name>`.                            | Shows canonical target while preserving inspection path to raw mention.   |
| `auto_resolved` | Solid border plus auto marker.                              | Visible `auto` marker.                            | `Auto-resolved <entity type>: <display name>; matched <alias text>`. | MUST remain inspectably marked after transient disclosure fades.          |
| `dismissed`     | Low-emphasis chip or token treatment plus dismissed marker. | Visible `dismissed` marker.                       | `Dismissed mention: <raw text>`.                                     | Display only where inspectable; excluded from active relationship values. |

*Design direction.* Color difference alone MUST NOT distinguish any pair of chip states.

### 16.2 Save, conflict, and presence presentation

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Element                    | Required placement or marker                                      | Bounds                                                                              |
| -------------------------- | ----------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| Save-state label           | Primary status-strip slot.                                        | Exactly one visible label.                                                          |
| Same-field conflict marker | Cell-local marker at the affected cell.                           | MUST include visible shape marker and accessible name; color alone is insufficient. |
| Conflict resolver entry    | From conflicted cell.                                             | Grid remains visible; affected row remains in view where possible.                  |
| Queue overflow banner      | Same-surface non-modal banner or secondary status message.        | MUST NOT block grid editing.                                                        |
| Header presence            | Top bar.                                                          | Show up to five visible avatars or initials, then `+N`.                             |
| Row presence               | Row gutter.                                                       | Show up to three visible indicators, then `+N`.                                     |
| Cell presence              | Cell-level indicator when same-field editing signal is available. | MUST NOT lock editing.                                                              |

*Design direction.* The default cell conflict marker is a corner triangle or equivalent shape marker plus accessible name `Conflict on <field label>`. If the implementation uses a different shape, it MUST still be visible without relying on color.

### 16.3 Inspector drawer

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Property         | Default and bounds                                                             |
| ---------------- | ------------------------------------------------------------------------------ |
| Default width    | `420px`.                                                                       |
| Minimum width    | `360px`.                                                                       |
| Maximum width    | `min(560px, 45vw)`.                                                            |
| Resize           | User MAY resize within min/max bounds.                                         |
| Pinning          | Client-local only; not persisted in saved views.                               |
| Section order    | Details, Relationships, Evidence, History, Workflow when all are configured.   |
| Grid visibility  | At base viewport, the grid remains visible whenever the inspector is open.     |
| Close affordance | Visible close control with accessible name `Close inspector`.                  |
| Pin affordance   | Visible pin control with accessible name `Pin inspector` or `Unpin inspector`. |

*Design direction.* If the viewport cannot preserve both the minimum grid width and the minimum inspector width, §17 narrow-viewport behavior applies.

### 16.4 Typography and density

*Design direction.* All rows in the following type table inherit `*Design direction.*`.

| Token role      | Default minimum       |
| --------------- | --------------------- |
| Metadata text   | `12px` CSS font size. |
| Grid cell text  | `14px` CSS font size. |
| Section heading | `16px` CSS font size. |
| Surface title   | `18px` CSS font size. |

*Design direction.* All rows in the following density table inherit `*Design direction.*`.

| Density     | Row height | Cell padding                       |
| ----------- | ---------- | ---------------------------------- |
| Compact     | `28px`.    | `3px` vertical, `6px` horizontal.  |
| Default     | `36px`.    | `4px` vertical, `8px` horizontal.  |
| Comfortable | `44px`.    | `6px` vertical, `10px` horizontal. |

*Core behavior.* Core 01 §3.3.2.3 owns the persisted account density preference. `density_mode=null` means no user override; effective density is `Compact` for Timeline and `Default` for every other workbook surface. Non-null persisted values are exactly `Compact`, `Default`, or `Comfortable` as UI labels for `compact`, `default`, and `comfortable`.

*Design direction.* The default workbook density is `Default`. The default Timeline grid uses `Compact` density from the shared density table to keep the first incident-response viewport grid-first. Hosts, Identities, Evidence, Notes, required system views, and the account density preference MUST use the same shared density classes and MUST NOT invent separate row-height or padding models.

*Design direction.* These density row heights are fixed-height defaults. Large incident grids MUST use fixed-height density rows unless a later design revision explicitly approves variable-height behavior for a bounded surface.[^15]

*Design direction.* Variable row height MAY be used only for a bounded specialized surface or diagnostic fixture. It MUST NOT become the default Timeline, Hosts, Identities, Evidence, Notes, or required system-view behavior without focused large-grid, virtualization, keyboard, frozen-column, and visual-regression evidence.

### 16.5 Color and contrast

*Design direction.* The base guide uses only semantic color roles for state: `info`, `success`, `caution`, `conflict`, `presence-self`, and `presence-other`.

*Design direction.* Every state represented by color MUST also be represented by shape, text, or accessible name. State-bearing text and controls MUST meet WCAG 2.2 AA contrast against the surrounding surface.

### 16.6 Icon and affordance conventions

*Design direction.* Mention chip actions, evidence preview, evidence download, history, rollback, merge initiation, party link, party unlink, inspector close, inspector pin, and system-view switcher SHOULD use one icon family.

*Design direction.* The guide does not select the icon family. The implementation MUST use exactly one icon family for those affordances within the base workbook shell. Icon-only controls MUST have accessible names and visible focus indicators.

### 16.7 Grid CSS integration

*Baseline context.* The RDG stylesheet and layout mechanics are implementation baseline inputs, not design-authority sources. Design direction MAY define semantic states, density, contrast, iconography, and accessible markers. It MUST NOT require styling hooks that depend on generated vendor class-name internals.[^15]

*Design direction.* Semantic Cartulary states such as conflict, presence, unresolved mention, selected row, active cell, evidence preview, and grouped row MUST be expressible through stable Cartulary wrapper classes, CSS variables, accessible names, shape, or text. Color-only and generated-class-only state encoding is non-conformant for this guide.

### 16.8 Reference figures

*Design direction.* The figures below are schematic and define relative semantics, not colors or exact pixel layout.

```text
Chip state examples
[? Unresolved host: WS-023?]   [Resolved host: WS-023]   [auto WS-023]   [dismissed WS-023?]
 dashed + ? marker              solid resolved chip        auto marker    dismissed marker
```

```text
Save/conflict/status placement
Top bar:       [System views ▾] [Presence A B +3]
Grid cell:     timeline.activity_synopsis_text  "VPN logon..."  ◢ Conflict on Activity Synopsis
Status strip:  Conflict | Queue 0 | More status ▾
```

```text
Inspector drawer
+-----------------------------+
| Inspector        [pin] [x]  |
| Details                     |
| Relationships               |
| Evidence                    |
| History                     |
+-----------------------------+
```

```text
Presence levels
Header: Presence A B C +2
Row gutter: row rec_123  A B +1
Cell: timeline.activity_synopsis_text  B editing
```

### Acceptance criteria

- **R2-AC-073:** §16 exists with chip states, save/conflict/presence, inspector, typography/density, color/contrast, icon, and grid CSS integration subsections.
- **R2-AC-074:** §16 defines inspector width defaults and min/max bounds.
- **R2-AC-075:** §16 defines row-height and cell-padding density values.
- **R2-AC-076:** §16 defines maximum visible presence indicators before `+N` overflow.
- **R2-AC-077:** §16 defines semantic color roles and the non-color-only rule.
- **R2-AC-078:** §16 requires one icon family for named workbook affordances.
- **R2-AC-079:** §16 includes four schematic figures.
- **R2-AC-080:** §5.9, §10.3, and §16.4 keep account density preference separate from incident workbook preferences and omit unsupported account-profile controls.
- **R2-RDG-AC-006:** §16.4 says fixed density row heights are the default.
- **R2-RDG-AC-007:** §16.4 prohibits variable row height as the default for Timeline, Hosts, Identities, Evidence, Notes, or required system views.
- **R2-RDG-AC-008:** §16.4 requires focused evidence before variable row height becomes ordinary behavior.
- **R2-RDG-AC-009:** §16.7 forbids generated-class-only state encoding.
- **R2-RDG-AC-010:** §16.7 requires semantic states to be expressible through stable Cartulary wrappers, CSS variables, accessible names, shape, or text.

## 17. Input, Viewport, and Accessibility Posture

### 17.1 Primary target

*Design direction.* The base guide targets desktop browser operation with keyboard and pointer input.

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Property            | Value                                                              |
| ------------------- | ------------------------------------------------------------------ |
| Width               | At least `1280` CSS pixels.                                        |
| Height              | At least `720` CSS pixels.                                         |
| Browser zoom        | `100%`.                                                            |
| Device scale factor | Not specified by the guide unless a benchmark profile supplies it. |

*Design direction.* At the base viewport, the active surface identity, primary tabs, system-view switcher, grid, save-state label, and status-strip primary state MUST remain visible.

### 17.2 Narrow viewport

*Design direction.* Below `1280` CSS pixels wide, the shell SHOULD preserve the grid as the primary work surface.

*Design direction.* In narrow viewport behavior, the inspector MAY collapse to a bottom sheet, modal drawer, or full-height overlay. Primary tabs MAY condense, but the active surface identity MUST remain visible. The save-state label MUST remain reachable without opening settings or a different module. The system-view switcher MUST remain reachable by pointer and keyboard.

*Later scope.* The base guide does not claim a responsive mobile experience. Touch-input support MAY exist. Omission of touch-specific gestures is conformant. If touch support exists, it MUST NOT replace keyboard or pointer paths for any base hot-path operation.

### 17.3 Keyboard-first operation

*Design direction.* Every hot-path operation named in §7.2 MUST be reachable by keyboard. Every destructive or enrichment action exposed in the inspector MUST be reachable by keyboard. Focus return MUST be specified for every dialog, drawer, popover, or resolver introduced by this guide.

### 17.4 Screen-reader direction

*Design direction.* The UI MUST expose save-state label changes, same-field conflict state, queue overflow, replay blocked, presence state, chip state, auto-resolution disclosure, and inspector open and close through accessible names, roles, or live-region announcements where appropriate.

*Design direction.* Row accessibility MUST bind DOM or accessibility-tree identity to stable `record_id`. The spoken row label MAY use the surface title and primary display fields. It SHOULD NOT expose raw `record_id` as the only human-facing row name.

### 17.5 Color and contrast

*Design direction.* Every state-bearing visual element MUST meet WCAG 2.2 AA contrast against the surrounding surface. Color MUST NOT be the sole carrier of state.

### 17.6 Deferred accessibility and locale work

*Later scope.* Revision 2 defers internationalization beyond locale-independent core behavior, right-to-left layout, a dedicated high-contrast theme, and reduced-motion preference-specific design.

*Design direction.* The deferral does not permit color-only state, motion-only state, inaccessible focus, or contrast below the §17.5 floor.

### Acceptance criteria

- **R2-AC-080:** §17 defines the base viewport as `1280x720` CSS pixels at `100%` zoom.
- **R2-AC-081:** §17 states that the base profile targets desktop keyboard and pointer input.
- **R2-AC-082:** §17 defines omission semantics for touch support.
- **R2-AC-083:** §17 requires keyboard reachability for all §7.2 hot-path operations.
- **R2-AC-084:** §17 requires accessible names or announcements for save, conflict, queue, presence, and chip states.
- **R2-AC-085:** §17 separates stable `record_id` binding from human-readable spoken row labels.
- **R2-AC-086:** §17 sets WCAG 2.2 AA as the contrast floor.

## 18. Cross-Cutting UX Patterns

### 18.1 Loading states

*Design direction.* Workbook first paint MUST show, at minimum, primary tab strip, system-view switcher, active surface identity, view bar shell, and in-place grid skeleton.

*Design direction.* If the first useful viewport is not available after `2 seconds`, the grid skeleton MUST show an explicit delayed-state message. The message MUST NOT navigate away from the workbook shell.

*Design direction.* Inspector first paint on row selection MUST show a metadata shell before binary preview bytes or full blob content load.

### 18.2 Empty states

*Design direction.* Every required surface MUST define an empty state. The empty state MUST name the minimum create signal for that surface and offer an in-place create affordance when the caller has create permission.

*Design direction.* Empty states MUST NOT suggest actions that leave the workbook shell.

### 18.3 Error presentation

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Error origin                       | Default placement                                                | Blocking rule                                         |
| ---------------------------------- | ---------------------------------------------------------------- | ----------------------------------------------------- |
| Grid mutation, field-local         | Cell-level inline error.                                         | MUST NOT block the entire grid.                       |
| Grid mutation, row-level           | Row-level inline error or row banner.                            | MUST NOT navigate away.                               |
| Destructive inspector action       | In-drawer error near the action.                                 | MUST NOT silently retry.                              |
| Collaboration stream interruption  | Status strip secondary slot.                                     | MAY add banner only if persistent.                    |
| Queue overflow                     | Status strip secondary slot plus same-surface non-modal message. | MUST NOT silently evict.                              |
| Session expiry or re-auth required | Status strip secondary slot; MAY use banner.                     | MUST preserve local pending work in the same runtime. |

*Design direction.* If multiple error classes are active, the UI MUST use the status-strip priority in §5.7.

### 18.4 Banner, toast, and inline messages

*Design direction.* All rows in the following table inherit `*Design direction.*`.

| Pattern        | Purpose                                       | Default duration                   | Dismissal                                               |
| -------------- | --------------------------------------------- | ---------------------------------- | ------------------------------------------------------- |
| Banner         | Persistent same-surface state.                | Until the underlying state clears. | Not dismissible while the state remains true.           |
| Toast          | Transient confirmation.                       | Auto-dismiss after `5 seconds`.    | Pauses while hovered, focused, or containing an action. |
| Inline message | Cell-local, row-local, or action-local state. | Until corrected or replaced.       | Dismissible only when informational.                    |
| Modal dialog   | Confirmation where confirmation is the point. | Until explicit action or cancel.   | `Esc` cancels only when cancellation is safe.           |

*Design direction.* Banners are reserved for overflow, session expiry, persistent collaboration degradation, and pack degradation affecting the active surface. Toasts are reserved for transient confirmation such as auto-resolution disclosure after batch paste or save success after a long queue. Inline messages carry cell-local or row-local state.

### 18.5 Dialog versus drawer

*Design direction.* The inspector is the default drawer. Routine enrichment MUST NOT use modal dialogs.

*Design direction.* Dialogs are reserved for destructive-action confirmation, merge initiation confirmation, rollback confirmation, same-field conflict resolution only when a drawer or same-surface panel cannot preserve required context, and release-scope-change acknowledgment when the Snapshot and Reporting Extension Profile is implemented.

*Design direction.* Dialog initial focus MUST land on a non-destructive summary or the safest action. `Esc` cancels only when cancellation is safe and leaves state unchanged. After close, focus MUST return to the invoking cell, row, or control when still present.

### 18.6 Truncation and overflow

*Design direction.* Chip labels, cell values, saved-view names, system-view labels, and filter chips MUST truncate with end ellipsis when they exceed their container. The full value MUST be available on hover and keyboard focus through a tooltip, popover, or accessible description. Visible active-query chips MUST remain pointer- and keyboard-reachable; overflow MUST stay within the query rail and MUST NOT cover adjacent controls.

*Design direction.* Tab-strip overflow beyond the visible strip SHOULD use end-anchored overflow, not horizontal scroll, at the base viewport. Long system-view lists use the ordering defined in §5.3, not ad hoc sorting.

### Acceptance criteria

- **R2-AC-087:** §18 exists with loading, empty, error, message, dialog, and truncation subsections.
- **R2-AC-088:** §18.1 defines first-paint minimum elements and the `2 seconds` delayed-state threshold.
- **R2-AC-089:** §18.2 requires empty states with minimum create signal and in-place create affordance.
- **R2-AC-090:** §18.3 defines placement for grid, inspector, collaboration, queue, and session errors.
- **R2-AC-091:** §18.4 defines banner, toast, inline, and modal defaults.
- **R2-AC-092:** §18.5 defines dialog focus and `Esc` behavior.
- **R2-AC-093:** §18.6 defines truncation, full-value disclosure, and overflow behavior.

## 19. Deferred Scope

*Later scope.* Revision 2 defers the component library and design-token artifact, Figma source, claimed mobile or touch-first profile, internationalization and right-to-left layout, dedicated high-contrast theme, reduced-motion preference-specific design, repo-specific design-system linkage, and extension-profile UX documents for Import, Snapshot and Reporting, Incident Portability, Reference Pack, and Enterprise Authentication.

*Later scope.* Deferred items MAY be specified in later guides or NLSpecs. They MUST NOT be represented as completed Revision 2 design requirements or Base Profile runtime behavior.

### Acceptance criteria

- **R2-AC-094:** §19 names all deferred scope items.
- **R2-AC-095:** Deferred scope is marked as `*Later scope.*`.
- **R2-AC-096:** Deferred items do not appear as completed change-log items.

## 20. Revision 2 Verification Matrix

| Area                   | Required pass condition                                                                                                       |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Authority              | §2 directly states Core 00 precedence and subordinate implementation-support context.                                         |
| Statement classes      | §§3 through 19 use the four closed markers.                                                                                   |
| Core restatements      | `*Core behavior.*` paragraphs are descriptive and do not issue guide-owned imperatives.                                       |
| Shell composition      | §5.2 defines tabs, switcher, ordering, optional surfaces, saved-view placement, keyboard behavior, and rejected alternatives; §5.8 defines authenticated root landing compositions, deployment-administration menu placement, allowed panels, Reference Pack search handling, and import-completion action; §5.10 defines administrative audit view presentation. |
| Status strip           | §5.7 defines three slots, priority, overflow, and KPI exclusion boundary.                                                     |
| `Esc` behavior         | §7.2 and §7.4 define the same priority ladder.                                                                                |
| Save state             | §7.4 includes replay-paused `Syncing` and presence non-effect.                                                                |
| Queue durability       | §7.4 and §9.5 both name cross-tab transfer as non-survival.                                                                   |
| Chip states            | §8.3 and §16.1 use the same state table.                                                                                      |
| Rollback               | §9.4 distinguishes `history_entry`, `change_set`, and row-backed-only `row_restore`.                                          |
| Snapshot boundary      | §12.4 prohibits per-recipient ordinary live-editing controls.                                                                 |
| Coordination coherence | §6.2 owns surface enumeration; §11.2 owns posture only.                                                                       |
| Baseline coherence     | §14 contains only baseline-consequence content and treats `react-data-grid` behavior through `/packages/grid-adapter`.        |
| Positive patterns      | §15.4 exists with required positive-pattern rows.                                                                             |
| Visual language        | §16 exists with bounds, state treatment, fixed-height density defaults, grid CSS boundary, figures, and icon-family rule.     |
| Accessibility          | §17 defines viewport, input, keyboard, screen-reader, contrast, touch omission, and deferred scope.                           |
| Cross-cutting patterns | §18 defines loading, empty, error, message, dialog, and truncation defaults.                                                  |
| Acceptance criteria    | All acceptance criteria blocks use the heading `### Acceptance criteria` and remain binary.                                   |
| Citations              | Citations support claims actually present in the revised section.                                                             |
| Change log             | Revision 2 change log is deterministic and complete.                                                                          |
| Editorial audit        | Temporary audit lists guide-issued MUST and MUST NOT statements only.                                                         |

## Revision 2 change log

| Item ID | Priority | Target section             | Statement class affected                                       | Change kind        | Summary                                                                                                          | Verification hook                |
| ------- | -------- | -------------------------- | -------------------------------------------------------------- | ------------------ | ---------------------------------------------------------------------------------------------------------------- | -------------------------------- |
| P0-1    | P0       | §2                         | Core behavior                                                  | normative-accuracy | Repaired authority order and Core 05 runtime boundary.                                                           | R2-AC-001..R2-AC-004             |
| P0-2    | P0       | §7.4, §9.5                 | Core behavior, Design direction                                | normative-accuracy | Added cross-tab transfer as pending-queue non-survival condition.                                                | R2-AC-049..R2-AC-050             |
| P0-3    | P0       | §7.4                       | Core behavior, Design direction                                | normative-accuracy | Added replay-paused cases to `Syncing` and stated presence non-effect.                                           | R2-AC-037..R2-AC-038             |
| P0-4    | P0       | §7.2, §7.4                 | Design direction                                               | normative-accuracy | Replaced broad `Esc` behavior with a priority ladder.                                                            | R2-AC-039                        |
| P0-5    | P0       | §9.4                       | Core behavior, Design direction                                | normative-accuracy | Clarified rollback target scope and row-backed-only restore.                                                     | R2-AC-047..R2-AC-048             |
| P0-6    | P0       | §12.4                      | Design direction                                               | normative-accuracy | Prohibited per-recipient visibility controls in ordinary live editing.                                           | R2-AC-062                        |
| P1-1    | P1       | §2, §§3-19                 | Core behavior, Design direction, Baseline context, Later scope | design-direction   | Added closed statement-class grammar and marker scope.                                                           | R2-AC-005..R2-AC-009             |
| P1-2    | P1       | §5                         | Design direction                                               | design-direction   | Closed shell composition, switcher ordering, keyboard behavior, saved-view placement, and rejected alternatives. | R2-AC-017..R2-AC-022             |
| P1-3    | P1       | §16                        | Design direction                                               | design-direction   | Added visual-language bounds for chips, conflict, presence, inspector, typography, density, color, and icons.    | R2-AC-073..R2-AC-079             |
| P1-4    | P1       | §5.8                       | Core behavior, Design direction                                | design-direction   | Added authenticated root landing compositions while deferring selection behavior to Core 01.                     | R2-AC-097..R2-AC-098             |
| P2-1    | P2       | §6                         | Design direction                                               | coherence          | Made §6.2 the only guide-local canonical surface enumeration.                                                    | R2-AC-027..R2-AC-032             |
| P2-2    | P2       | §11                        | Design direction                                               | coherence          | Reduced coordination section to UX posture and must-not-become guidance.                                         | R2-AC-055..R2-AC-058             |
| P2-3    | P2       | §14                        | Baseline context                                               | baseline-context   | Reshaped implementation baseline into baseline consequences only.                                                | R2-AC-065..R2-AC-068             |
| P3-1    | P3       | §8.3, §16.1                | Design direction                                               | design-direction   | Added one shared chip-state table.                                                                               | R2-AC-040..R2-AC-043             |
| P3-2    | P3       | §17                        | Design direction, Later scope                                  | design-direction   | Added viewport, keyboard, screen-reader, touch, and contrast posture.                                            | R2-AC-080..R2-AC-086             |
| P3-3    | P3       | §18                        | Design direction                                               | design-direction   | Added loading, empty, error, banner, toast, dialog, and truncation defaults.                                     | R2-AC-087..R2-AC-093             |
| P3-4    | P3       | §15.4                      | Design direction                                               | design-direction   | Added positive-pattern reviewer table.                                                                           | R2-AC-071..R2-AC-072             |
| P3-5    | P3       | §5.7                       | Design direction                                               | design-direction   | Added numeric status-strip capacity and KPI exclusion boundary.                                                  | R2-AC-023..R2-AC-026             |
| P4-1    | P4       | Revision 2 editorial audit | N/A                                                            | editorial          | Added temporary normative-voice audit.                                                                           | Verification matrix, audit table |
| P4-2    | P4       | All acceptance criteria    | N/A                                                            | editorial          | Normalized acceptance-criteria heading and binary style.                                                         | R2-AC-001..R2-AC-109             |
| P4-3    | P4       | §5.1                       | Design direction                                               | incident-metadata  | Added compact create metadata and complete post-create incident-control guidance.                                | R2-AC-099..R2-AC-100             |
| P4-3    | P4       | Sources                    | N/A                                                            | editorial          | Cleaned citations to load-bearing source groups.                                                                 | Sources block                    |
| P4-4    | P4       | §§10.2, 14.1, 16.4, 16.7   | Design direction, Baseline context                             | baseline-context   | Added RDG adapter, treegrid/group-row, fixed-height density, and generated-class styling boundaries.             | R2-RDG-AC-001..R2-RDG-AC-010     |
| P4-5    | P4       | §5.1, §9.5                 | Core behavior, Design direction                                | incident-lifecycle | Added closed incident read-only shell guidance and rejected-draft non-replay behavior.                           | R2-AC-101..R2-AC-104             |
| P4-6    | P4       | §5.8                       | Core behavior, Design direction                                | deployment-admin   | Added deployment-administration menu derivation from the Core 04 matrix and extension claimed state.             | R2-AC-105                        |
| P4-7    | P4       | §5.10                      | Core behavior, Design direction                                | administrative-audit | Added administrative audit view guidance for server filters, redaction display, and split navigation.          | R2-AC-107                        |
| P4-8    | P4       | §5.8                       | Core behavior, Design direction                                | deployment-admin   | Added Deployment administration placement, panel, Reference Pack search, import-completion, and prohibited-control guidance. | R2-AC-108..R2-AC-109             |

## Revision 2 editorial audit

The audit is temporary reviewer scaffolding, not product direction. It SHOULD be removed in Revision 3 or later after voice discipline is established. It lists guide-issued MUST and MUST NOT statements and excludes descriptive `*Core behavior.*` restatements.

| Section | Statement excerpt                                                                                                                                                                                                                                                                   | Statement class  | Reason retained                                        | Owner-boundary note                                       |
| ------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ---------------- | ------------------------------------------------------ | --------------------------------------------------------- |
| §3.1    | The base UI MUST preserve a direct-manipulation loop: the user acts on the visible workbook object, the system shows how that action was interpreted, and slower reconciliation work remains visible as semantic state rather than hidden transport state.                          | Design direction | Defines UI posture.                                    | Does not alter Core interaction mechanics.                |
| §3.1    | The visible workbook surface MUST remain the default locus for capture, correction, linking, filtering, grouping, sorting, preview, and same-row history access.                                                                                                                    | Design direction | Prevents form-first drift.                             | Implements UI consequence of Core 03.                     |
| §3.1    | Routine row work MUST NOT require full-page navigation or a form-first workflow.                                                                                                                                                                                                    | Design direction | Keeps hot path workbook-native.                        | Does not add API behavior.                                |
| §4.1    | Cartulary MUST remain workbook-first because it replaces a real incident-response operating model, not an abstract CRUD problem.                                                                                                                                                    | Design direction | Fixes design center.                                   | Does not define conformance.                              |
| §4.2    | Cartulary MUST reject spreadsheet behaviors that conflict with auditable incident state: row-position identity, silent overwrites, hidden formulas as business logic, evidence paths as authoritative references, unmanaged binary storage, and unversioned relationship semantics. | Design direction | Defines rejection boundary.                            | Restates UI implications of owner contracts.              |
| §5.1    | After creation, an explicitly opened secondary incident-control surface for current `reviewer` or `admin` users MUST expose all Core-owned patchable incident metadata: `description`, `severity`, `tlp`, `current_phase`, and `primary_external_case_ref`. | Design direction | Prevents optional create fields from disappearing after creation. | Restates UI consequence of Core 01 incident PATCH. |
| §5.1    | The TLP control MAY show readable labels, but it MUST bind to exact canonical machine tokens or `null`; severity and phase controls MAY show suggestions, but MUST preserve otherwise valid bounded text. | Design direction | Keeps labels and suggestions separate from stored metadata. | Core 01/Core 02 own validation and token membership. |
| §5.1    | A closed workbook MUST render a persistent banner or shell-level state using the existing banner/read-only primitives and the exact visible label `Closed, read-only`. | Design direction | Defines the persistent closed-state affordance. | Restates UI consequence of Core 01/Core 03 lifecycle behavior. |
| §5.1    | When the incident is closed, source-state write controls MUST be disabled or hidden: add-row, direct cell edit, paste-to-mutate, row delete/restore/rollback/merge/supersede, conflict resolution, mention resolution, incident metadata patch, blob-slot creation, and evidence attachment. | Design direction | Prevents closed incidents from looking writable. | Core 01 owns the operation matrix. |
| §5.2    | The base shell MUST expose built-in tabs as always-visible primary tabs at the base viewport.                                                                                                                                                                                       | Design direction | Closes IA decision.                                    | Does not change required surface inventory.               |
| §5.2    | Required system views MUST be reachable through an adjacent switcher with accessible name `System views`.                                                                                                                                                                           | Design direction | Closes IA decision.                                    | Does not change surface identity.                         |
| §5.2    | Saved views MUST appear under the active surface’s view selector, not as primary tabs by default.                                                                                                                                                                                   | Design direction | Prevents saved-view identity drift.                    | Uses Core saved-view model.                               |
| §5.3    | The `System views` switcher MUST group required system views in the order below.                                                                                                                                                                                                    | Design direction | Defines deterministic ordering.                        | UI ordering only.                                         |
| §5.3    | The implementation MUST NOT alphabetize these required groups differently unless a later guide revision changes this table.                                                                                                                                                         | Design direction | Prevents divergent shells.                             | UI ordering only.                                         |
| §5.3    | If Findings, Investigative Queries, or Forensic Keywords are exposed, the switcher MUST add a final group named `Optional artifact surfaces` in this order: Findings, Investigative Queries, Forensic Keywords.                                                                     | Design direction | Defines optional-surface placement.                    | Conditional on owner-allowed surfaces.                    |
| §5.4    | At a base viewport of at least `1280x720` CSS pixels, the `System views` switcher trigger MUST remain visible in the top bar adjacent to the primary tab strip.                                                                                                                     | Design direction | Defines viewport-visible IA.                           | UI presentation only.                                     |
| §5.4    | Selecting a system view MUST open it inside the same workbook shell.                                                                                                                                                                                                                | Design direction | Prevents module fragmentation.                         | Restates Core workbook-native consequence.                |
| §5.4    | Selecting a system view MUST NOT navigate to a separate module, route family, browser tab, or dashboard shell.                                                                                                                                                                      | Design direction | Prevents shell drift.                                  | UI route presentation only.                               |
| §5.4    | Keyboard operation MUST support Tab focus to the trigger, Enter or Space to open, Arrow keys to move within the open menu, Enter to select, and Esc to close without changing the active surface.                                                                                   | Design direction | Defines accessible switcher behavior.                  | Does not alter Core keyboard contracts.                   |
| §5.4    | When the menu closes without selection, focus MUST return to the switcher trigger.                                                                                                                                                                                                  | Design direction | Defines focus return.                                  | UI-only requirement.                                      |
| §5.7    | The status strip MUST remain a capacity-limited working-state surface.                                                                                                                                                                                                              | Design direction | Prevents dashboard sprawl.                             | UI chrome only.                                           |
| §5.7    | It MUST NOT become a dashboard or management summary.                                                                                                                                                                                                                               | Design direction | Prevents dashboard sprawl.                             | UI chrome only.                                           |
| §5.7    | Additional status messages MUST collapse into a same-surface overflow affordance labeled `More status` or an equivalent accessible name.                                                                                                                                            | Design direction | Defines overflow behavior.                             | UI-only requirement.                                      |
| §5.7    | Persistent shell chrome and the status strip MUST NOT show incident-level KPIs, time-to-resolution counters, team throughput, external ticket counts, or management dashboard metrics.                                                                                              | Design direction | Prevents dashboard sprawl.                             | Does not ban metrics inside opened surfaces.              |
| §5.8    | Incident-directory search and deployment-admin user-list search MUST treat the server-side list route as authoritative.                                                                                                                                                             | Design direction | Prevents partial-cursor filtering from being presented as exhaustive search. | Restates UI consequence of Core 01 list search.           |
| §5.8    | While a newer search or filter request is pending, the UI MUST keep the prior result set visible, display `Searching`, submit the current search immediately on Enter, and discard stale responses by a monotonically increasing client request sequence.                            | Design direction | Defines responsive list-search state.                  | UI request handling only.                                 |
| §5.8    | Deployment-administration menu items MUST be rendered from the current session's `deployment_admin` status, the Core 04 matrix, and `GET /api/v1/extensions` claimed-state results.                                                                                                 | Design direction | Prevents guide-local deployment-role drift.            | Presentation of owner-owned authorization only.           |
| §5.8    | The Reference Packs administration group MUST be hidden when unclaimed and MUST show authorization failure rather than a false empty state when claimed but forbidden.                                                                                                              | Design direction | Prevents extension claim and authorization ambiguity.   | Presentation of Core 01 and Core 04 owner contracts only. |
| §5.8    | Incident membership controls MUST remain incident-role driven and MUST NOT be exposed solely because the current user holds `deployment_admin`.                                                                                                                                      | Design direction | Keeps deployment and incident administration separate.  | UI consequence of Core 04 authorization.                  |
| §5.8    | The global Deployment administration entry MUST appear only as a menu item labeled `Deployment administration` in the upper-right account/application menu.                                                                                                                         | Design direction | Fixes global navigation placement.                     | Presentation of Core 01 route only.                       |
| §5.8    | Deployment administration MUST use a distinct deployment-local administration layout, not the workbook grid shell.                                                                                                                                                                  | Design direction | Keeps administration out of workbook surfaces.          | UI consequence of Core 01 and Core 03 boundaries.         |
| §5.8    | Reference Pack search MUST clear or supersede stale pagination state when search or filters change, preserve the prior accepted result set while `Searching`, reject stale responses by client request sequence, and bind accepted pagination only to the accepted query state.        | Design direction | Prevents stale search and cursor confusion.             | UI request handling for Core 01 list query.               |
| §5.8    | The Incident import panel MUST show `Open imported incident` only after a successful import result with an imported incident target.                                                                                                                                                | Design direction | Provides explicit import completion action.             | UI entrypoint for Core 03 startup behavior.               |
| §5.8    | Deployment administration MUST NOT expose `General settings`, all-incident catalog/search/count/metadata controls, generic cross-incident policy-default editors, provider-definition editors, provider-wide recovery controls, incident membership controls driven only by `deployment_admin`, initial-admin selectors, adoption routes, portable-membership editors, or historical-actor membership mapping controls. | Design direction | Prevents aggregate administration drift.                | Omission consequence of Core 01/Core 04 boundaries.       |
| §5.10   | Account and deployment audit views MUST use the server filters declared by Core 01 rather than free-text search.                                                                                                                                                                   | Design direction | Prevents incomplete client-side audit inspection.       | Core 01 owns filters and search rejection.                |
| §5.10   | The UI MUST NOT present a client-side search box as exhaustive over a partially loaded cursor collection.                                                                                                                                                                           | Design direction | Prevents misleading partial-cursor search.              | UI presentation only.                                     |
| §5.10   | Redacted audit values MUST be visually distinct from visible JSON `null` values.                                                                                                                                                                                                    | Design direction | Keeps redaction meaning visible to reviewers.           | Core 04 owns redaction requirements.                      |
| §5.10   | Redaction presentation MAY use a stable label, icon, or badge, but it MUST preserve the event row, field path, and before/after positions so reviewers can see that a sensitive field changed.                                                                                      | Design direction | Preserves review context without exposing secrets.      | Presentation of Core 01/Core 04 contracts only.           |
| §5.10   | Deployment audit navigation MUST follow the current session's `deployment_admin` status and the Core 04 matrix.                                                                                                                                                                    | Design direction | Prevents local navigation-role drift.                   | Core 04 owns authorization.                               |
| §5.10   | Incident membership audit navigation MUST follow the current incident role `admin` state and MUST NOT be exposed solely because the current user holds `deployment_admin`.                                                                                                          | Design direction | Preserves deployment-versus-incident split.             | UI consequence of Core 04 authorization.                  |
| §6.3    | When exposed, the optional standardized surfaces MUST inherit the same shell, row, query, filter, grouping, saved-view, history, and inspector grammar as other workbook surfaces.                                                                                                  | Design direction | Keeps optional surfaces workbook-native.               | Conditional on owner-allowed optional surfaces.           |
| §7.1    | Selecting a cell and typing MUST edit it immediately.                                                                                                                                                                                                                               | Design direction | Defines direct edit posture.                           | UI behavior only.                                         |
| §7.1    | The user MUST NOT have to enter a separate form edit mode for the common row-editing path.                                                                                                                                                                                          | Design direction | Prevents form-first drift.                             | UI behavior only.                                         |
| §7.1    | Relationship cells MUST accept raw typing.                                                                                                                                                                                                                                          | Design direction | Preserves text-first capture.                          | UI consequence of binding contracts.                      |
| §7.1    | They MUST NOT require picker-first interaction.                                                                                                                                                                                                                                     | Design direction | Prevents picker-first capture loss.                    | UI consequence only.                                      |
| §7.3    | Multi-cell paste, fill-down, and multi-row tag assignment are required workbook behaviors in the base design. They MUST NOT rely on hidden macro semantics.                                                                                                                         | Design direction | Defines bulk-edit posture.                             | Does not define mutation wire beyond owner contracts.     |
| §7.4    | Autosave MUST occur on Enter, Tab, blur, and paste completion.                                                                                                                                                                                                                      | Design direction | Defines save UX trigger posture.                       | UI trigger posture, not storage owner.                    |
| §7.4    | `Esc` MUST NOT cancel already queued replay units, committed authoritative changes, or unresolved same-field conflict objects.                                                                                                                                                      | Design direction | Closes destructive shortcut ambiguity.                 | Does not alter owner replay or history.                   |
| §8.3    | §8.3 and §16.1 MUST use the same four states: `unresolved`, `resolved`, `auto_resolved`, and `dismissed`.                                                                                                                                                                           | Design direction | Prevents vocabulary drift.                             | Visual vocabulary only.                                   |
| §8.3    | Unresolved chips MUST be visually distinct from resolved chips through a combination of border treatment and an inline state marker.                                                                                                                                                | Design direction | Ensures observable state.                              | Visual presentation only.                                 |
| §8.3    | Color difference alone MUST NOT be the sole distinguishing signal.                                                                                                                                                                                                                  | Design direction | Accessibility and state clarity.                       | Visual presentation only.                                 |
| §8.3    | Auto-resolved chips MUST show the `auto` marker defined in §16.1 until the user explicitly changes or reverts the resolution.                                                                                                                                                       | Design direction | Preserves auto-resolution disclosure.                  | Visual presentation only.                                 |
| §8.3    | When shown, they MUST carry a visible dismissed marker and accessible name.                                                                                                                                                                                                         | Design direction | Preserves dismissed-state inspection.                  | Visual presentation only.                                 |
| §8.4    | It MUST NOT replace direct grid editing for ordinary capture.                                                                                                                                                                                                                       | Design direction | Prevents inspector overreach.                          | UI posture only.                                          |
| §8.4    | The inspector MUST NOT become a full-page record editor, dashboard, ticketing module, release-control module, or hidden source of saved-view state.                                                                                                                                 | Design direction | Prevents shell drift.                                  | UI posture only.                                          |
| §8.5    | Auto-resolved chips MUST remain distinguishable from manually resolved chips until the user explicitly changes or reverts them.                                                                                                                                                     | Design direction | Preserves disclosure.                                  | UI presentation only.                                     |
| §8.5    | The UI MUST expose a same-surface transient confirmation and an inspectable route to review affected cells.                                                                                                                                                                         | Design direction | Defines auto-resolution recourse.                      | UI behavior only.                                         |
| §9.1    | The UI MUST never imply that row position, sort position, visible label, or display value is the mutation target.                                                                                                                                                                   | Design direction | Prevents identity drift.                               | UI presentation of owner identity rule.                   |
| §9.1    | Same-field conflicts MUST be shown as cell-local state.                                                                                                                                                                                                                             | Design direction | Prevents sheet-wide freeze.                            | UI consequence of Core 03.                                |
| §9.1    | A conflict on one cell MUST NOT freeze the entire row, sheet, or workbook.                                                                                                                                                                                                          | Design direction | Prevents overblocking.                                 | UI consequence only.                                      |
| §9.2    | Presence indicators MUST NOT lock editing.                                                                                                                                                                                                                                          | Design direction | Keeps presence advisory.                               | Does not alter concurrency owner behavior.                |
| §9.3    | Closing the resolver without selecting a resolution MUST leave the cell in conflict state.                                                                                                                                                                                          | Design direction | Defines resolver behavior.                             | Restates UI consequence of Core 03.                       |
| §9.3    | After an explicit resolution or clear action, focus MUST return to the same cell and scroll position SHOULD be preserved.                                                                                                                                                           | Design direction | Defines focus recovery.                                | UI behavior only.                                         |
| §9.4    | The history UI MUST present those three rollback target kinds as different actions, not as aliases for “restore this row.”                                                                                                                                                          | Design direction | Prevents rollback ambiguity.                           | Presents Core 01 scope.                                   |
| §9.4    | Whole-row restore affordances MUST label the row-field-only scope before confirmation.                                                                                                                                                                                              | Design direction | Prevents destructive misunderstanding.                 | UI consequence only.                                      |
| §9.5    | Cartulary’s base-profile pending queue MUST NOT be inferred to survive or replay through such cross-tab mechanisms.                                                                                                                                                                 | Design direction | Closes client-state ambiguity.                         | Does not alter Core 03 queue contract.                    |
| §9.5    | Reopen MUST NOT automatically restart those rejected mutations; the user must make a fresh edit or explicit submit action against reopened current state.                                                                                                                           | Design direction | Prevents stale local drafts from becoming authoritative after reopen. | Core 03 owns rejected-draft behavior.                    |
| §10.2   | Grouping MUST remain a view-state operation, not a data-model mutation.                                                                                                                                                                                                             | Design direction | Prevents grouping/write confusion.                     | UI consequence of Core 01.                                |
| §10.2   | Dragging, expanding, or collapsing groups MUST NOT create, delete, or mutate incident records.                                                                                                                                                                                      | Design direction | Prevents accidental mutation semantics.                | UI consequence only.                                      |
| §10.2   | When grouping is rendered through a treegrid pattern, group rows MUST be presented as navigation and summarization affordances, not ordinary incident records.                                                                                                                      | Design direction | Prevents writable-record ambiguity.                    | UI consequence of adapter and grouping contract.          |
| §10.2   | Group rows MUST expose expand/collapse affordances, MUST be keyboard navigable, and MUST NOT expose ordinary writable cell affordances.                                                                                                                                             | Design direction | Defines group-row affordance boundary.                 | UI-only consequence of treegrid presentation.             |
| §10.2   | Paste, drag fill, editor entry, entity resolution, evidence attach, and destructive record actions MUST NOT be available on group rows.                                                                                                                                             | Design direction | Prevents synthetic-row mutation.                       | Does not change record mutation owner.                    |
| §10.3   | The UI MUST NOT persist client-local state into saved views.                                                                                                                                                                                                                        | Design direction | Prevents saved-view drift.                             | Restates Core saved-view state boundary.                  |
| §11.3   | Coordination creation MUST NOT become a mandatory step in ordinary row capture.                                                                                                                                                                                                     | Design direction | Prevents per-edit ritual.                              | Maintains Core hot path.                                  |
| §11.4   | Future operating-model guides MAY define recommended cadences or playbook practices for teams, but those practices MUST NOT be represented as Base Profile implementation-conformance requirements unless restated in Core 00 through Core 04.                                      | Later scope      | Preserves owner boundary.                              | Matches Core 00 supporting-guidance boundary.             |
| §12.1   | Evidence cells MUST NOT be raw object-store URLs, local file paths, or user-editable storage keys.                                                                                                                                                                                  | Design direction | Prevents unsafe evidence semantics.                    | UI consequence of Core evidence contracts.                |
| §12.2   | The UI MUST NOT navigate away from the grid for ordinary screenshot attachment.                                                                                                                                                                                                     | Design direction | Preserves evidence hot path.                           | UI behavior only.                                         |
| §12.4   | The live workbook MUST NOT present per-recipient visibility affordances such as disclosure-partition chips, recipient-selector controls, or release-scope badges during ordinary editing.                                                                                           | Design direction | Prevents snapshot boundary drift.                      | Restates Core 04 release-time boundary.                   |
| §13.3   | Cartulary MUST reject spreadsheet behaviors that undermine case integrity.                                                                                                                                                                                                          | Design direction | Defines rejection boundary.                            | UI interpretation only.                                   |
| §14.2   | It MUST NOT block ordinary rough capture, grid editing, or base workbook navigation.                                                                                                                                                                                                | Design direction | Defines pack-degradation UX.                           | Does not alter pack owner behavior.                       |
| §14.1   | Grid-first editing, keyboard, paste, virtualization, custom renderers, explicit editor adapters, grouping/treegrid behavior, frozen-column behavior, and CSS integration MUST remain compatible with the adapter contract.                                                          | Baseline context | Keeps vendor behavior mediated by adapter.             | Does not make RDG the behavior owner.                     |
| §14.1   | The UI guide MUST NOT assume vendor behavior that the adapter has not exposed and tested.                                                                                                                                                                                           | Baseline context | Prevents design drift toward unowned vendor internals. | Adapter contract remains implementation-support boundary. |
| §14.3   | The UI MUST NOT route ordinary clipboard paste through a file-import wizard.                                                                                                                                                                                                        | Design direction | Keeps paste hot path.                                  | UI behavior only.                                         |
| §15.3   | They MUST NOT be represented as completed Revision 2 design requirements or Base Profile runtime behavior.                                                                                                                                                                          | Later scope      | Maintains deferred-scope boundary.                     | Does not define runtime behavior.                         |
| §16.1   | Color difference alone MUST NOT distinguish any pair of chip states.                                                                                                                                                                                                                | Design direction | Accessibility and state clarity.                       | Visual presentation only.                                 |
| §16.2   | If the implementation uses a different shape, it MUST still be visible without relying on color.                                                                                                                                                                                    | Design direction | Accessibility and state clarity.                       | Visual presentation only.                                 |
| §16.4   | The default Timeline grid uses `Compact` density from the shared density table while the default workbook density remains `Default`.                                                                                                                                                 | Design direction | Keeps Timeline grid-first without changing Core 01.    | UI presentation only.                                     |
| §16.4   | Hosts, Identities, Evidence, Notes, required system views, and user-selected density preferences MUST use shared density classes and MUST NOT invent separate density models.                                                                                                        | Design direction | Prevents density drift.                                | UI presentation only.                                     |
| §16.4   | Large incident grids MUST use fixed-height density rows unless a later design revision explicitly approves variable-height behavior for a bounded surface.                                                                                                                          | Design direction | Protects large-grid performance and visual stability.  | UI presentation only.                                     |
| §16.4   | Variable row height MUST NOT become the default Timeline, Hosts, Identities, Evidence, Notes, or required system-view behavior without focused large-grid, virtualization, keyboard, frozen-column, and visual-regression evidence.                                                 | Design direction | Prevents default variable-height drift.                | UI presentation only.                                     |
| §16.5   | Every state represented by color MUST also be represented by shape, text, or accessible name.                                                                                                                                                                                       | Design direction | Accessibility and state clarity.                       | Visual presentation only.                                 |
| §16.5   | State-bearing text and controls MUST meet WCAG 2.2 AA contrast against the surrounding surface.                                                                                                                                                                                     | Design direction | Sets accessibility floor.                              | UI presentation only.                                     |
| §16.6   | The implementation MUST use exactly one icon family for those affordances within the base workbook shell.                                                                                                                                                                           | Design direction | Prevents affordance drift.                             | UI presentation only.                                     |
| §16.6   | Icon-only controls MUST have accessible names and visible focus indicators.                                                                                                                                                                                                         | Design direction | Accessibility requirement.                             | UI presentation only.                                     |
| §16.7   | Design direction MUST NOT require styling hooks that depend on generated vendor class-name internals.                                                                                                                                                                               | Baseline context | Prevents brittle styling hooks.                        | RDG styling remains implementation detail.                |
| §16.7   | Semantic Cartulary states such as conflict, presence, unresolved mention, selected row, active cell, evidence preview, and grouped row MUST be expressible through stable Cartulary wrapper classes, CSS variables, accessible names, shape, or text.                               | Design direction | Defines stable semantic styling path.                  | UI presentation only.                                     |
| §17.1   | At the base viewport, the active surface identity, primary tabs, system-view switcher, grid, save-state label, and status-strip primary state MUST remain visible.                                                                                                                  | Design direction | Defines base viewport behavior.                        | UI presentation only.                                     |
| §17.2   | Primary tabs MAY condense, but the active surface identity MUST remain visible.                                                                                                                                                                                                     | Design direction | Defines narrow behavior.                               | UI presentation only.                                     |
| §17.2   | The save-state label MUST remain reachable without opening settings or a different module.                                                                                                                                                                                          | Design direction | Defines narrow behavior.                               | UI presentation only.                                     |
| §17.2   | The system-view switcher MUST remain reachable by pointer and keyboard.                                                                                                                                                                                                             | Design direction | Defines narrow behavior.                               | UI presentation only.                                     |
| §17.2   | If touch support exists, it MUST NOT replace keyboard or pointer paths for any base hot-path operation.                                                                                                                                                                             | Later scope      | Defines touch omission semantics.                      | Does not claim touch profile.                             |
| §17.3   | Every hot-path operation named in §7.2 MUST be reachable by keyboard.                                                                                                                                                                                                               | Design direction | Accessibility and input contract.                      | UI behavior only.                                         |
| §17.3   | Every destructive or enrichment action exposed in the inspector MUST be reachable by keyboard.                                                                                                                                                                                      | Design direction | Accessibility and input contract.                      | UI behavior only.                                         |
| §17.3   | Focus return MUST be specified for every dialog, drawer, popover, or resolver introduced by this guide.                                                                                                                                                                             | Design direction | Accessibility and input contract.                      | UI behavior only.                                         |
| §17.4   | The UI MUST expose save-state label changes, same-field conflict state, queue overflow, replay blocked, presence state, chip state, auto-resolution disclosure, and inspector open and close through accessible names, roles, or live-region announcements where appropriate.       | Design direction | Accessibility and state clarity.                       | UI presentation only.                                     |
| §17.4   | Row accessibility MUST bind DOM or accessibility-tree identity to stable `record_id`.                                                                                                                                                                                               | Design direction | Stable accessibility association.                      | UI presentation of owner identity rule.                   |
| §17.5   | Every state-bearing visual element MUST meet WCAG 2.2 AA contrast against the surrounding surface.                                                                                                                                                                                  | Design direction | Sets accessibility floor.                              | UI presentation only.                                     |
| §17.5   | Color MUST NOT be the sole carrier of state.                                                                                                                                                                                                                                        | Design direction | Accessibility and state clarity.                       | UI presentation only.                                     |
| §18.1   | Workbook first paint MUST show, at minimum, primary tab strip, system-view switcher, active surface identity, view bar shell, and in-place grid skeleton.                                                                                                                           | Design direction | Defines loading baseline.                              | UI behavior only.                                         |
| §18.1   | If the first useful viewport is not available after `2 seconds`, the grid skeleton MUST show an explicit delayed-state message.                                                                                                                                                     | Design direction | Defines delayed loading behavior.                      | UI behavior only.                                         |
| §18.1   | The message MUST NOT navigate away from the workbook shell.                                                                                                                                                                                                                         | Design direction | Prevents loading detour.                               | UI behavior only.                                         |
| §18.1   | Inspector first paint on row selection MUST show a metadata shell before binary preview bytes or full blob content load.                                                                                                                                                            | Design direction | Defines inspector loading behavior.                    | UI behavior only.                                         |
| §18.2   | Every required surface MUST define an empty state.                                                                                                                                                                                                                                  | Design direction | Defines empty-state completeness.                      | UI behavior only.                                         |
| §18.2   | The empty state MUST name the minimum create signal for that surface and offer an in-place create affordance when the caller has create permission.                                                                                                                                 | Design direction | Defines empty-state action.                            | UI behavior only.                                         |
| §18.2   | Empty states MUST NOT suggest actions that leave the workbook shell.                                                                                                                                                                                                                | Design direction | Prevents module detour.                                | UI behavior only.                                         |
| §18.3   | If multiple error classes are active, the UI MUST use the status-strip priority in §5.7.                                                                                                                                                                                            | Design direction | Defines error precedence.                              | UI behavior only.                                         |
| §18.5   | Routine enrichment MUST NOT use modal dialogs.                                                                                                                                                                                                                                      | Design direction | Prevents modal drift.                                  | UI behavior only.                                         |
| §18.5   | Dialog initial focus MUST land on a non-destructive summary or the safest action.                                                                                                                                                                                                   | Design direction | Defines modal safety.                                  | UI behavior only.                                         |
| §18.5   | After close, focus MUST return to the invoking cell, row, or control when still present.                                                                                                                                                                                            | Design direction | Defines focus return.                                  | UI behavior only.                                         |
| §18.6   | Chip labels, cell values, saved-view names, system-view labels, and filter chips MUST truncate with end ellipsis when they exceed their container.                                                                                                                                  | Design direction | Defines overflow behavior.                             | UI presentation only.                                     |
| §18.6   | The full value MUST be available on hover and keyboard focus through a tooltip, popover, or accessible description.                                                                                                                                                                 | Design direction | Defines accessible disclosure.                         | UI presentation only.                                     |

## Sources

[^1]: `cartulary-ui-ux-design-guide.md`, previous revision, especially §§1-16.
[^2]: `00_document_set_status_and_precedence.md`, especially §§1-5.1 and REQ-00-051.
[^3]: `05_claim_publication_and_benchmark_reproducibility.md`, especially §1.
[^4]: `03_workbook_interaction_collaboration_and_workflows.md`, especially §§1-4, §§7-16, and requirements for built-in tabs, system views, saved views, collaboration, save states, pending queue, and same-field conflicts.
[^5]: `01_architecture_storage_and_view_contracts.md`, especially §§1-3.3.5, §7.4, and workbook/view/rollback contracts.
[^6]: `02_domain_model_schema_and_history.md`, especially §§1-7 and history, mention, entity, party, and provenance requirements.
[^7]: `04_security_deployment_and_conformance.md`, especially §§2, 4.2, 9.0, and 9.1.
[^8]: `cartulary-dev-guide.md`; `cartulary_repository_bootstrap_guide.md`; `cartulary_implementation_testing_guide.md`.
[^9]: `R04-responsive_browser_spreadsheet_ui_research_memo.md`; `R05-responsive-interface-design-report.cr.md`.
[^10]: Ben Shneiderman, “Direct Manipulation: A Step Beyond Programming Languages,” *IEEE Computer* 16(8), 1983; Zhicheng Liu and Jeffrey Heer, “The Effects of Interactive Latency on Exploratory Visual Analysis,” *IEEE Transactions on Visualization and Computer Graphics* 20(12), 2014; Danyel Fisher et al., “Trust Me, I’m Partially Right: Incremental Visualization Lets Analysts Explore Large Datasets Faster,” CHI 2012; Dominik Moritz et al., “Trust, but Verify: Optimistic Visualizations of Approximate Queries for Exploring Big Data,” CHI 2017.
[^11]: `A_problem_framing_rationale_tradeoffs_and_sanity_check.md`; `G_source_archive_exploratory_design_artifact.md`.
[^12]: `R06-spreadsheet_of_doom_dfir_research_report.md`; `R07-spreadsheet-of-doom-sod-report.cr.md`; `R01-aurora_incident_response_report.md`; `R03-Kanvas_technical_research_report.md`.
[^13]: `H_operating_model_supporting_guidance.md`; `R02-cartulary_crm_tem_dfir_research_report.md`.
[^14]: `D_workflow_and_ui_illustrations_source_extract.md`; `B_architecture_diagrams_and_explanatory_source_extract.md`; `C_schema_reference_and_ddl_source_extract.md`; `E_roadmap_open_questions_and_decision_backlog.md`; `F_source_traceability_matrix.md`.
[^15]: `R09-react-data-grid-research-report.md`, especially §§3, 12–15, 17–21, and the evidence ledger for adapter-relevant RDG behavior, controlled state, grouping/treegrid behavior, CSS, performance, and fragile grid combinations.
