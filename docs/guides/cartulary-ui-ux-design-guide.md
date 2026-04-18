
# Cartulary UI/UX Design Guide

**Status:** Derived design-direction specification.  
**Authority:** Subordinate to Cartulary Core 00 through Core 04 for current runtime behavior, and subordinate to Core 05 for claim-bearing timed or fixture-sensitive publication behavior only.  
**Scope:** Base-profile UI/UX direction, with explicit extension-profile and later-scope boundary notes.  
**Interpretation:** In this guide, **MUST** and **MUST NOT** identify design-direction requirements that are necessary to implement or review the current Cartulary product correctly. **SHOULD** and **SHOULD NOT** identify strong defaults. **MAY** identifies optional behavior with explicit omission semantics. These terms govern this guide’s interpretation of Cartulary’s UX direction; they MUST NOT be read as widening or replacing the owner behavior defined by the normative core.[^1][^20]

## 1. Executive Summary

Cartulary’s UX thesis is precise: the product MUST feel like a serious workbook on the hot path and MUST behave like a disciplined case system underneath.[^2][^3][^6]

That thesis has eight immediate consequences. The primary work surface MUST be a virtualized workbook grid. Direct typing, keyboard navigation, clipboard paste, fill-down, sort, filter, grouping, and low-friction row creation MUST remain first-class behaviors. The inspector MUST be adjacent and secondary, not the default editor. Relationship cells MUST stay text-first and support progressive normalization rather than picker-first creation. Evidence MUST be an attached object model with application-mediated preview and download, not path text or raw object-store URLs. Collaboration MUST use field-level optimistic concurrency over row-versioned records, with automatic reconciliation for different-field edits and explicit same-field resolution when analysts collide. Coordination surfaces such as Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson MUST remain workbook-native surfaces rather than separate modules. The current implementation baseline MUST be treated as concrete: one application unit, PostgreSQL, S3-compatible object storage, a Go backend, and a React plus `react-data-grid` plus Vite browser client. File-based structured import beyond clipboard paste MUST remain separate extension-profile territory and MUST NOT define the base workbook grammar.[^2][^3][^4][^5][^20]

The product therefore does not aim to replace the spreadsheet mental model. It aims to preserve the spreadsheet mental model at the view layer while rejecting spreadsheet storage, concurrency, automation, and evidence semantics where those would undermine auditability, recoverability, or collaboration. Cartulary is workbook-first and forms-second by design, not by nostalgia.[^2][^3][^6][^9]

The rest of this guide explains the resulting interface model, information architecture, state model, and design rationale in implementation-ready terms. It is written to stand on its own, so a reader can understand the product without reopening the corpus.[^1][^6]

## 2. Reading Guide and Source Method

This guide is a derived design-direction artifact, not a new owner document. The authority order for current product behavior in the provided corpus is: the Cartulary normative core for current runtime behavior, then appendices and research reports for explanation and rationale, then implementation guides for current realization direction. No adopted Cartulary NLSpec appears in the provided product corpus as a higher current product authority than Core 00 through Core 04.[^1][^20]

The table below defines how this guide uses the corpus.

| Source class | Role in this guide | What it is allowed to decide |
| --- | --- | --- |
| Core 00 through Core 04 | Runtime owner | Current base-profile behavior, boundaries, identifiers, defaults, and UI-affecting mechanics |
| Core 05 | Publication owner only | Claim-bearing timed or fixture-sensitive publication behavior, not ordinary runtime UX |
| Appendices A through H | Explanatory context | Rationale, illustrations, historical context, operating guidance, and future-only backlog |
| Research reports R01 through R07 | Project-specific rationale | Why Cartulary should preserve, adapt, or reject specific interaction patterns |
| Development, bootstrap, and testing guides | Implementation-direction context | Current stack baseline, module boundaries, and verification hooks |

This guide distinguishes four kinds of statements. **Current core behavior** means behavior required by the current base profile. **Design-direction consequence** means the UX implication of that core behavior. **Implementation-baseline context** means the current concrete realization direction that should shape design decisions without overriding the owner sections. **Extension or later scope** means behavior the current guide should leave room for without presenting it as part of the current base workbook contract.[^1][^3][^20]

External UI/UX sources are used selectively and only when they sharpen a Cartulary-specific decision. Direct-manipulation research is used to justify why the grid must remain the visible object of work. Table and spreadsheet-comprehension research is used to justify why context recovery, adjacent detail, and local structure matter. Latency and progressive-interface research is used to justify immediate acknowledgement, visible provisional state, and same-surface correction rather than opaque background normalization. Progressive-disclosure guidance is used only to justify the inspector as a secondary adjacent surface rather than a form-first replacement for the grid.[^12][^13][^14][^15][^16][^17][^18][^19]

## 3. Cartulary’s Product and Interaction Thesis

### Design contract

Cartulary MUST be understood as a workbook-native incident workspace whose primary interaction object is the visible row and cell, not the hidden form model behind it.[^2][^3]

The product’s core thesis can be stated in one sentence: **Cartulary MUST feel as fast and direct as a serious spreadsheet during live incident work, while making structure, evidence, attribution, revision, and concurrent editing far more explicit than a spreadsheet ever can**.[^6][^9]

That thesis resolves several design questions up front.

| Design question | Required answer |
| --- | --- |
| What is the main object of thought? | The visible workbook row, cell, chip, count, and filter state |
| What is the main editing mode? | Direct manipulation on the grid |
| What is the system actually storing? | Structured, relational, auditable state underneath workbook projections |
| What is the role of secondary surfaces? | Adjacent enrichment, review, history, and destructive actions |
| What is the collaboration target? | Reliable shared case state, not character-by-character coauthoring |

This is not a decorative application of spreadsheet visual styling. Shneiderman’s direct-manipulation model is relevant because it describes exactly the interaction quality Cartulary needs: users work on a visible representation of the task, effects are incremental and legible, and the interface stays cognitively close to the underlying work object. For Cartulary, that means the system’s interpretation of a change MUST remain visible at the workbook surface: a typed host token remains a token until resolved, an exact-match reuse becomes a chip, an evidence attach becomes a count and preview affordance, and a same-field collision becomes a marked cell rather than a hidden transport failure.[^12][^13]

The grid is therefore not a temporary intake layer before the “real application” starts. Table research shows why that would be the wrong move. Analysts continue to use tabular views because those views keep them close to the data, make reorganization cheap, and preserve human-readable detail. Cartulary extends that professional strength; it does not treat the grid as an unfortunate legacy surface to be escaped as soon as more structure exists.[^12][^16]

### Acceptance criteria

- A user who can describe their work as “that row,” “this host token,” or “the blocked requests due today” can complete the common path without leaving the workbook shell.[^2][^7]
- The grid remains the default editor for capture, correction, paste, sorting, filtering, grouping, and quick relationship actions.[^2]
- The system never requires a user to understand storage tables, object-store locations, sync queues, or internal route families in order to stay oriented while editing.[^3][^12]

## 4. Why Cartulary Is Workbook-First

### Design contract

Cartulary MUST be workbook-first because it is replacing a real incident-response operating model, not an abstract CRUD problem.[^6][^9][^10][^11]

The Spreadsheet of Doom persists for concrete reasons. It tolerates incomplete facts, supports direct typing, rewards paste, allows quick reshaping of working sets, and gives an incident team one shared tabular memory. Research into spreadsheet-centric DFIR practice, Aurora, and Kanvas all show the same underlying truth: responders keep returning to workbook-shaped coordination because it is operationally honest about how incident facts actually arrive.[^9][^10][^11]

A forms-first application fails precisely where incident work is weakest. The first useful version of a fact is often “maybe jdoe on WS-023 via VPN; screenshot attached; time uncertain.” Core Cartulary behavior explicitly allows that roughness: `occurred_at` may be null, summary may be absent if another field or attachment exists, host and identity text may remain unresolved, and detail may remain unstructured. The design consequence is that the UI MUST let that row exist before the system has perfect structure.[^4][^6]

The table below captures the design difference.

| Criterion | Workbook-first Cartulary | Forms-first case application |
| --- | --- | --- |
| First useful capture | Direct row or cell entry | Create form, choose type, satisfy fields |
| Relationship handling | Text-first, then explicit resolve/create | Picker-first or schema-first |
| Evidence attach | Same-surface attach to current row | Detached media workflow |
| Iteration speed | Immediate local capture, later normalization | Up-front structure before persistence |
| Cognitive model | One visible working surface | Many record detail screens |

Workbook-first also does **not** mean “just a data table.” The workbook shell is a complete interaction model: tabs and system views over projections, saved views, filters, grouping, inspector-based enrichment, ambient presence, history, evidence preview, and coordination surfaces that stay inside the same shell. The point is not to copy spreadsheet internals. The point is to preserve the workbook as the object of work.[^2][^3][^7]

### Acceptance criteria

- A user can create a timeline row that contains unresolved host or identity text, an uncertain time, and attached evidence without being forced into a record form first.[^2][^4][^7]
- The design never assumes the user must choose a record subtype, workflow state, or relationship target before entering the first usable version of a fact.[^2][^6]
- The grid remains the primary work surface after structure accumulates; richer structure does not cause the product to “graduate” the user into form-first operation.[^12][^16]

## 5. Overall Application Shell and Information Architecture

### Design contract

The Cartulary shell MUST present one coherent workbook environment rather than a dashboard shell plus embedded mini-applications.[^2][^3]

The shell has five regions.

| Region | Required contents | Purpose | Must not become |
| --- | --- | --- | --- |
| Top bar | Incident identity, workbook surfaces, presence avatars, global search entry | Stable incident and surface context | A KPI dashboard or module switchboard |
| View bar | Saved-view selector, sort, group, filter controls | View-state control for the active surface | A second data model |
| Grid | Rows and cells for the active surface | Primary capture and editing plane | A read-only report |
| Inspector | Non-blocking adjacent drawer | Enrichment, detail, review, and destructive actions | The default editor |
| Status strip | Save state, queue/conflict telemetry, keyboard hints | Operational state and confidence | A management dashboard |

The shell SHOULD include a compact bottom status strip or equivalent same-surface telemetry region. This is not a contradiction of the “not a dashboard” rule. Cartulary MUST reject KPI-console drift, but it SHOULD expose dense, domain-relevant status telemetry such as `Syncing`, `Saved`, `Conflict`, queue-overflow messages, and row-local keyboard hints because those signals directly protect analyst work on the hot path.[^2][^7]

The shell MUST keep the workbook surface identity legible at all times. The active surface, active saved view if any, grouping state, filter chips, and presence state SHOULD be visible without opening a second application area. When the inspector opens, it MUST be obvious that the user is still on the same workbook surface, acting on the same selected record.[^2][^7]

The shell MUST also preserve same-origin continuity. Evidence preview, history, conflict resolution, row review, and workbook-native coordination actions MUST occur inside the same application surface. They MUST NOT feel like external tools launched from the grid.[^3][^5]

### Acceptance criteria

- A user can move from Timeline to Task Requests, open a selected row in the inspector, preview evidence, open row history, and return focus to the same grid cell without full-page navigation.[^2][^7]
- The status strip or equivalent telemetry area exposes save/conflict state and urgent queue conditions without becoming a general dashboard surface.[^2][^7]
- No coordination surface feels like a detached workflow module hosted inside the workbook shell.[^2][^8]

## 6. Workbook Surface Model

### 6.1 Built-in tabs

The current base profile has **fourteen required workbook surfaces**: five built-in tabs and nine required system views. The five built-in tabs are Timeline, Hosts, Identities, Evidence, and Notes.[^2][^3]

The built-in tabs are required workbook peers. They are not separate storage silos and they are not optional shell customizations.

| Surface label | `view_schema_id` | Surface kind | Primary role |
| --- | --- | --- | --- |
| Timeline | `cartulary.view.timeline.v1` | Built-in tab | Primary capture surface |
| Hosts | `cartulary.view.hosts.v1` | Built-in tab | Canonical host and stub entity surface |
| Identities | `cartulary.view.identities.v1` | Built-in tab | Canonical identity and stub entity surface |
| Evidence | `cartulary.view.evidence.v1` | Built-in tab | Evidence records and linked object state |
| Notes | `cartulary.view.notes.v1` | Built-in tab | Artifact-backed notes surface |

### 6.2 System views

The required system views are equally authoritative workbook surfaces. This guide uses **Assessments** as the default reader-facing label for `cartulary.view.assessments.v1`. The owner corpus also uses **Compromise Assessments** for the same surface. No semantic distinction is intended.[^2][^3]

| Surface label | `view_schema_id` | Surface kind | Primary role |
| --- | --- | --- | --- |
| Indicators | `cartulary.view.indicators.v1` | Required system view | Canonical indicators with pivots to observations and lifecycle |
| Assessments | `cartulary.view.assessments.v1` | Required system view | Incident-scoped assessment history |
| Task Requests | `cartulary.view.task_requests.v1` | Required system view | Queue-oriented owned work |
| Decisions | `cartulary.view.decisions.v1` | Required system view | Rationale-bearing decision records |
| Parties | `cartulary.view.parties.v1` | Required system view | Stable incident-scoped coordination identities |
| Communications Log | `cartulary.view.comm_log.v1` | Required system view | Durable communication memory |
| Handoff | `cartulary.view.handoff.v1` | Required system view | Shift or phase continuity |
| Status Review | `cartulary.view.status_review.v1` | Required system view | Checkpoint review and rebalancing |
| Lesson | `cartulary.view.lesson.v1` | Required system view | Retrospective follow-through |

These fourteen surfaces are the authoritative required workbook surfaces of the base profile. A saved view over one of these surfaces is additive and non-canonical. Pack-dependent overlays such as ATT&CK, D3FEND, and VERIS are **not** current workbook-native `view_schema` surfaces in the base profile or the current Reference Pack Extension Profile.[^2][^3][^5]

### 6.3 Optional standardized workbook surfaces

Findings, Investigative Queries, and Forensic Keywords are the only additional current-profile standardized workbook surfaces beyond the fourteen required surfaces, and only when the implementation exposes them. When present, they MUST inherit the same workbook grammar: the same shell, the same row/query/filter/group/history model, and the same refusal to become separate modules.[^3][^20]

| Optional surface | `view_schema_id` | Required posture when exposed |
| --- | --- | --- |
| Findings | `cartulary.view.findings.v1` | Workbook-native surface, not a separate hypothesis module |
| Investigative Queries | `cartulary.view.investigative_queries.v1` | Workbook-native surface with the same queue/filter/history grammar |
| Forensic Keywords | `cartulary.view.forensic_keywords.v1` | Workbook-native surface with the same queue/filter/history grammar |

### 6.4 Saved views

Saved views are incident-bound workbook configurations over exactly one `view_schema_id`. They are **surface configurations**, not storage silos.[^2][^3]

The scope model is closed and explicit.

| Scope | Discoverability | In-place mutability | Ordinary create default |
| --- | --- | --- | --- |
| `private` | Owner and incident admins only | Owner and incident admins | **Default** |
| `shared` | All incident members | Owner and incident admins | Allowed on ordinary create |
| `system` | All incident members | Immutable through ordinary paths | Not allowed on ordinary create |

The ordinary saved-view create path MUST default `scope` to `private`. The ordinary create path MUST reject `scope='system'`. `private` and `shared` saved views record one owner. `system` saved views are implementation-owned or admin-seeded saved-view objects and remain distinct from the canonical underlying system view itself.[^2][^3]

Saved views persist only portable shared layout and query state. Selection, scroll position, focused cell, local popovers, open inspector state, preview state, and presence remain client-local and MUST NOT be persisted as part of the saved view.[^2][^3]

### 6.5 Startup and default surface behavior

Workbook startup is deterministic. The starting surface is chosen in this exact order.[^2][^3]

| Order | Startup source | Rule |
| --- | --- | --- |
| 1 | Explicit launch `sheet_ref` | Use it if present and still valid for the caller |
| 2 | `user_workbook_preferences.home_sheet_ref` | Use it if valid and visible |
| 3 | `incident_workbook_preferences.default_sheet_ref` | Use it if valid and visible |
| 4 | Timeline | Fallback default |

A `sheet_ref` of kind `view_schema` names the canonical required surface. A `sheet_ref` of kind `saved_view` names one distinct saved-view object over a schema. These are not interchangeable identities. If a pointer is missing, hidden, invalid, or depends on an unavailable optional pack, the implementation MUST clear the invalid pointer and continue down the fallback chain rather than failing workbook open.[^2][^3]

### Acceptance criteria

- The shell exposes fourteen required workbook surfaces: exactly five built-in tabs and exactly nine required system views.[^2][^3]
- Opening the workbook with no explicit launch target, no valid home surface, and no valid incident default always lands on Timeline.[^2]
- Ordinary saved-view create defaults to `private`, and ordinary create never exposes `system` as a user-creatable scope.[^2]
- Saved views persist view configuration only. Opening the same saved view on another device does not inherit prior selection, scroll position, open inspector state, preview state, or presence.[^2][^3]
- If Findings, Investigative Queries, or Forensic Keywords are implemented, they behave like workbook surfaces rather than module escapes.[^3][^20]

## 7. Grid Editing Model

### 7.1 Direct typing

Selecting a cell and typing MUST edit it immediately. The user does not enter a separate “form edit mode.” The grid is already the editor.[^2][^7]

The trailing blank row is part of that contract. Typing into it SHOULD create a real row as soon as the active surface’s minimum create signal is satisfied after create-time normalization. Timeline is the low-friction exemplar: one non-empty usable value in the row is enough to create a real record. Non-Timeline surfaces MAY require their declared minimum create signal, but they still MUST preserve same-surface inline creation rather than redirecting to a modal or form page.[^2][^3][^7]

Relationship cells MUST accept raw typing. They MUST NOT require picker-first interaction. If the user types raw host or identity text into a relationship cell, the system should capture that text under the active field contract first and offer explicit enrichment or resolution second.[^2][^4][^7]

### 7.2 Keyboard contract

The current keyboard contract is compact and intentionally memorable.[^2][^7][^20]

| Key | Required default effect | Must not become |
| --- | --- | --- |
| Arrow keys | Move grid selection | A hidden macro language |
| Enter | Commit and move vertically | A form-submit detour |
| Shift+Enter | Reverse vertical navigation | A second editing mode |
| Tab | Commit and move horizontally | A shell navigation shortcut |
| `Ctrl+V` | Paste into the visible grid | An import-only gesture |
| `Ctrl+K` | Quick link or resolve on the current cell | A full-screen workflow jump |
| `Space` | Preview linked evidence for the selected row | A browser scroll side effect |
| `Alt+H` | Open history for the selected row | A detached review page |
| `Esc` | Close the inspector and return focus to the prior cell | A destructive discard shortcut |

The key point is semantic consistency. `Ctrl+K` is the lightweight relationship and normalization affordance. `Space` keeps evidence preview in flow. `Alt+H` keeps history row-local. `Esc` reinforces that the inspector is an adjacent state, not another place.[^2][^7]

### 7.3 Paste and bulk editing

Multi-cell paste, fill-down, and multi-row tag assignment are required workbook behaviors. Bulk edits are mutation batches. They MUST NOT rely on hidden macro semantics.[^2][^7]

Clipboard paste is part of the base hot path. It SHOULD accept TSV and CSV copied directly from Excel or other spreadsheet tools, create additional rows automatically when the pasted range exceeds the existing visible set, and preserve selection and row identity rather than turning paste into an import wizard.[^2][^3][^7][^20]

### 7.4 Autosave and visible save states

The normal grid workflow has no explicit Save button. Autosave MUST occur on Enter, Tab, blur, and paste completion.[^2]

The UI MUST expose exactly three save-state labels.

| Label | Exact meaning | Required user interpretation |
| --- | --- | --- |
| `Syncing` | At least one mutation is in flight or the local pending queue is non-empty | The system is preserving work, but authoritative reconciliation is still underway |
| `Saved` | No mutation is in flight, the local pending queue is empty, and no unresolved same-field local drafts exist | Current workbook state is synchronized |
| `Conflict` | A same-field conflict is unresolved, queue overflow refused admission, or replay halted on a non-retryable failure | Analyst attention is required |

Latency research matters here. Interactive delay changes user strategy, not just satisfaction. Cartulary’s design implication is that acknowledgement of typing, selection changes, paste completion, and save-state transitions MUST feel immediate and predictable. Local acknowledgement should happen first; authoritative truth and exception states should then be rendered clearly without stalling the user’s thought process.[^12][^15]

### Acceptance criteria

- Selecting a cell and typing edits the cell immediately. No separate edit form or record detail page is required for the common path.[^2]
- Typing into the trailing blank row creates a real row when the active surface’s minimum create signal is satisfied.[^2][^7]
- Relationship cells accept raw typing and do not force picker-first interaction.[^2]
- `Ctrl+V`, fill-down, and multi-row tag assignment work on the visible workbook surface and are implemented as explicit mutation batches rather than hidden automation.[^2]
- The visible save-state labels are exactly `Syncing`, `Saved`, and `Conflict`.[^2]

## 8. Progressive Structuring and Relational UX

### 8.1 What “capture first, structure later” means

“Capture first, structure later” means the UI MUST accept rough, incomplete, and uncertain input as valid first-class state, then support later normalization without erasing the original user-entered material.[^4][^6]

That principle governs both entity references and evidence. A timeline row can exist with unresolved host and identity text, uncertain time, and linked evidence. The row becomes relationally richer later, but the rough capture remains preserved as source-bound lineage.[^4][^7]

### 8.2 Binding modes and their UI consequences

Cartulary distinguishes `mention_origin` from `entity_origin`. The UI MUST respect that distinction because the two modes answer different questions.[^4]

| Context | Binding mode | Default UI result | Forbidden shortcut |
| --- | --- | --- | --- |
| Timeline, Notes, and other non-entity record cells that refer to hosts or identities | `mention_origin` | Preserve raw token as an `entity_mention`; show unresolved or resolved chip state later | Auto-create canonical entities silently |
| Interactive clipboard paste into non-entity sheets | `mention_origin` | Create one mention per observed cell value and source row | Coalesce repeated mentions into one entity |
| Hosts and Identities direct row creation | `entity_origin` | Create or exact-match reuse a host or identity record | Convert the row into a mere mention |
| Inspector “Resolve to existing” | Explicit mention action | Preserve raw mention; add canonical link | Replace raw token with canonical text only |
| Inspector “Create host” or “Create identity” from mention | Explicit mention action | Create one stub by default and resolve the selected mention only | Resolve sibling mentions automatically |

This distinction is one of the places where Cartulary most clearly diverges from spreadsheet sloppiness without sacrificing spreadsheet speed. Rough tokens on non-entity surfaces stay source-bound until explicitly resolved. Direct entity surfaces are where entity rows are created or upserted directly.[^4][^7]

### 8.3 Chips, unresolved state, and exact-match boundaries

Relationship state SHOULD be rendered as chips or equivalent compact inline state, not as indefinitely raw delimited strings. The grid needs to show relational status without turning relationship editing into a form-only experience.[^2][^7]

At minimum, the UI needs three visible states.

| Cell state | Visible representation | Meaning |
| --- | --- | --- |
| Unresolved | Dotted, outlined, or otherwise marked raw token chip | A preserved source-bound mention awaiting action |
| Resolved | Plain canonical chip | A current canonical link exists |
| Auto-resolved | Canonical chip with inspectable auto-resolution mark | The system applied the bounded exact-match auto-resolution rule |

Exact-match suggestion boundaries MUST remain conservative. Auto-resolution is allowed only in the current narrow scope: interactive mention capture on Timeline host and identity relationship cells during inline commit or interactive clipboard paste. It depends on exact alias equality under the declared comparison substrate and is blocked by uncertainty suppressors such as `?`, `~`, `maybe`, `prob`, `probably`, `approx`, and `approximately`. Fuzzy matching, token deletion, transliteration, stemming, punctuation rewriting, and background cleanup workflows are out of scope for auto-resolution in the current profile.[^2][^4]

Manual linking is scoreless in the current profile. The workbook MUST NOT require analysts to enter numeric confidence for routine manual links. Unsupported attempts to supply manual link confidence are validation errors rather than silent downgrades.[^2]

### 8.4 Inspector contract: what it owns, and what it must not

The inspector is the secondary enrichment surface. It MUST remain a non-blocking drawer, MUST keep the main grid visible while open, and MUST NOT become the default primary capture surface.[^2]

Its ownership boundary is explicit.

| Owned by the grid | Owned by the inspector | Not allowed |
| --- | --- | --- |
| Direct typing, row creation, paste, sort, filter, grouping, quick row-level navigation | Detail, relationships, evidence inspection, row history, rollback, destructive actions, explicit mention flows, indicator observation flows, party create/link/clear flows, merge initiation | Replacing routine row creation or routine field editing as the default path |
| Immediate cell editing | Explicit enrichment and review | Blocking the grid behind a modal workflow |
| Text-first relationship entry | Explicit resolve/create/dismiss/restore | Requiring a full-page record form |
| Working-set navigation | Same-surface detail and review | Becoming a second application module |

The inspector MUST support all of the following current-profile capabilities when the selected surface and row make them relevant: mention resolve, create from mention, dismiss, restore to unresolved, indicator observation link/create/dismiss flows, indicator lifecycle inspection or editing, party create-from-text, link-existing, clear-link, clear-text, clear-both, evidence inspection, rollback access, and explicit merge initiation for authorized users.[^2][^4]

The inspector MUST also preserve key distinctions that spreadsheets usually blur: raw mention lineage versus current canonical link, raw indicator observation lineage versus current canonical indicator and lifecycle state, visible source-preserving party text versus optional hidden linked `party_id`, and workbook context versus destructive reviewer actions.[^2][^4]

This is progressive disclosure used narrowly and deliberately: deep detail, relationship review, history, and destructive actions are one move away, but the visible workbook surface remains the primary object of thought. The inspector MUST therefore add depth without stealing the center of gravity from the grid.[^12][^19]

### 8.5 Auto-resolution disclosure, undo, and later correction

Cartulary’s bounded auto-resolution behavior is only safe if it is visible and reversible. The UI MUST NOT silently auto-resolve.[^2]

When auto-resolution occurs, the same sheet MUST show an immediate non-modal disclosure that includes the raw token, canonical target, matched alias text, direct `Undo`, and direct `Review`. `Review` SHOULD open the current row or chip in the same-surface relationship review context without losing grid visibility or current row anchoring. For batch paste, the disclosure MUST also include the number of tokens auto-resolved in the same visible change set.[^2]

The resolved chip or cell MUST remain inspectably marked as auto-resolved after the transient disclosure fades. Row history MUST preserve the raw token, matched alias text, confidence, and mutation source. `Undo` from the immediate disclosure MUST restore the raw unresolved token, remove the auto-created link, preserve focus, and preserve scroll position. After the disclosure expires, the user MUST still be able to choose `Revert to unresolved` from the chip context or row history in no more than two actions, and that later correction MUST create a new attributed revision.[^2][^4]

This pattern directly expresses Cartulary’s “never silently normalize” stance. Optimistic and progressive-interface research supports staged feedback when provisionality is legible and correction remains cheap. Cartulary’s implication is concrete: exact auto-resolution can help flow, but only if the user can immediately see what happened and undo it without losing place.[^12][^18]

### Acceptance criteria

- Rough capture remains valid: a row can exist with unresolved host or identity text, uncertain time, and evidence linkage.[^4][^6]
- `mention_origin` and `entity_origin` produce visibly different default behavior on the grid.[^4]
- The inspector keeps the grid visible and remains secondary; the common path of timeline typing does not require opening it.[^2]
- The inspector exposes explicit mention, indicator, party, evidence, history, rollback, and merge flows without becoming the routine editor.[^2][^4]
- Auto-resolution is never silent. The sheet shows disclosure, `Undo`, `Review`, batch counts when relevant, and later `Revert to unresolved` within two actions.[^2]

## 9. Collaboration, Presence, Autosave, and Conflict Resolution

### 9.1 Collaboration model

Cartulary’s collaboration model is not character-by-character coauthoring. It is field-level optimistic concurrency on top of row versioning.[^2][^3]

The server-side behavior is intentionally simple.

| Situation | Required server behavior | Required UI consequence |
| --- | --- | --- |
| Current `base_row_version` still matches | Apply patch normally | Standard save-state progression |
| Another user changed a different field | Auto-rebase and accept | No user-visible conflict |
| Another user changed the same field | Reject with conflict payload | Only the affected cell enters conflict state |

The client MUST address rows by `record_id` and `base_row_version`, never by visible row number, sort position, grouped position, or displayed values. That identity discipline is what allows the UI to stay stable while live updates, sorting, filtering, and grouping continue around the active row.[^2][^3]

### 9.2 Presence

Presence is ambient, not blocking. The UI MUST provide workbook-header avatars for users on the same surface, row-gutter indicators when another analyst is focused on the same row, and same-cell indicators when another analyst is actively editing the same field and the signal is available.[^2]

Presence MAY warn; it MUST NOT lock. It is there to reduce surprise before a collision becomes a same-field conflict, not to turn ordinary editing into a lock protocol.[^2][^7]

### 9.3 Same-field conflict resolution

A same-field conflict is a cell-local problem, not a sheet-wide freeze. When one occurs, only the affected cell MUST enter conflict state. The cell continues to show the saved server value plus a visible conflict marker. The client retains the analyst’s unsaved local value separately and MUST NOT render that local value as if it were already saved.[^2]

The resolver MUST open from the conflicted cell, keep the main grid visible, and preserve row context. Closing the resolver without selecting a resolution leaves the cell in conflict state. The analyst MUST still be able to continue editing other rows or cells while the conflict remains unresolved elsewhere.[^2]

The resolver contents are also closed and explicit. It MUST show row context, field display label and stable `field_key`, the saved value plus actor and timestamp, the analyst’s unsaved local value, diff support when the field’s conflict class supports it, and direct resolution actions in plain language. Initial focus MUST land on a non-destructive summary element. Pressing Enter on first open MUST NOT resolve the conflict accidentally.[^2]

Conflict handling depends on the declared `conflict_resolution_class`.

| Resolution class | Typical field family | Required resolver behavior |
| --- | --- | --- |
| `atomic_replace` | Scalar fields such as enums, timestamps, numbers, single-value identifiers | Explicit `Keep saved value` and `Use my unsaved value` actions |
| `text_compare_merge` | Analyst-authored free text such as summary, details, note body, description | Side-by-side comparison with change highlighting and optional `Edit merged value` path |
| `collection_review` | Multi-value chip or set fields such as tags, Timeline Hosts, Timeline Identities | Base, saved, and local deltas plus final preview before commit |

If a field’s class is unknown or omitted, the resolver defaults to `atomic_replace` behavior.[^2]

### 9.4 Authorship, row history, and rollback

Authorship MUST be low friction. At minimum, the row surface exposes last editor and relative update time, and history is reachable in one click or one shortcut such as `Alt+H`.[^2]

The row history surface is newest-first and row-centric. It MUST show, at minimum, the actor, committed time, displayable operation label, diff summary, `change_set_id`, current reversibility state, and available rollback actions. The rollback action family is closed: `history_entry`, `change_set`, and `row_restore`. When whole-row restore is legal for a given displayed logical history item, the entry also includes `revision_no`. When a logical history item maps to exactly one mutation target eligible for single-entry addressing, it includes a stable opaque `history_entry_ref`.[^3][^4]

The design implication is that row history is not just an audit trail. It is a working review surface. The user should be able to inspect what changed, who changed it, and what can still be reversed without leaving workbook context. Rollback and restore belong to the inspector or history surface, not to ordinary cell editing.[^2][^3][^4]

Delete, restore, rollback, supersede, and merge append history. They MUST NOT narrow prior retained history for an extant incident record. Current rollback eligibility may narrow over time, but history visibility must not.[^3][^4]

### 9.5 Local pending queue, overflow, and reauthentication

The base-profile client-local pending queue is part of the UX contract, not a transport implementation detail. It exists so transient network interruptions do not lose typed work.[^2]

Its user-visible properties are explicit.

| Property | Base-profile requirement |
| --- | --- |
| Scope | Incident-scoped and client-instance-scoped |
| Storage | Memory-local in the current browser runtime |
| Durability | Survives transient transport failure, auth failure on queued write, and `session_revoked` in the same runtime; does **not** survive full reload, tab close, restart, or crash |
| Capacity | Exactly 64 replay units per `(incident_id, client_instance_id)` |
| Order | FIFO by original enqueue order |
| Coalescing | Allowed only for one still-uncommitted local row create or one contiguous same-record patch run |
| Overflow behavior | Preserve existing queue, refuse new admission, preserve current visible edit as unsaved local work, set save state to `Conflict`, show same-surface non-modal overflow message |

Overflow is one of the most important trust moments in the product. The client MUST NOT silently evict older or newer queued units to make room. It MUST preserve everything already queued, refuse the new replay unit, keep the current edit visibly unsaved, and tell the user what happened on the same workbook surface.[^2]

A same-surface message SHOULD use copy equivalent to: “Pending local edits are full. Keep this tab open, resolve conflicts, or wait for sync before adding more queued edits.” This is a design recommendation, not a new owner-level string constant.[^7]

Queued unsent writes and unresolved same-field local drafts MUST survive reauthentication within the same browser runtime. Replay resumes only after the client re-establishes an authenticated session when required, re-derives current incident authorization, and completes any required HTTP re-query. If replay stops at a same-field conflict, the blocking unit leaves the pending queue and enters the same-field conflict queue; later queued units remain queued behind it rather than being applied out of order. If replay stops at another terminal failure, later queued units remain queued, save state remains `Conflict`, and the blocking failure is surfaced on the same sheet.[^2]

### Acceptance criteria

- Different-field concurrent edits reconcile automatically. Same-field concurrent edits do not silently overwrite; only the affected cell enters conflict state.[^2]
- The resolver opens from the conflicted cell, keeps the grid visible, and does not default focus to a destructive action.[^2]
- `Alt+H` opens a row-local newest-first history surface with actor, time, diff summary, rollback availability, and stable item identity where applicable.[^2][^3][^4]
- The pending queue is memory-local, FIFO, capacity 64, survives transient transport and same-runtime reauthentication, and does not promise survival across reload.[^2]
- Queue overflow preserves the existing queue, refuses only the new unit, keeps the current edit unsaved, sets `Conflict`, and shows a same-surface non-modal message.[^2][^7]

## 10. Sorting, Filtering, Grouping, and View State

### 10.1 Sorting and filtering

Sorting and filtering are workbook behaviors, not separate analytical modules. Column-header sorting and inline filter chips MUST apply without leaving the active sheet, and they MUST follow the capabilities declared by the active `view_schema` rather than inferred visible labels.[^2][^3]

The UI SHOULD make the current sort, filters, and grouping state visible in the view bar. Clearing a user sort override returns the surface to schema default sort only. Clearing grouping persists omission of `group_by`; “Group: None” is not represented as JSON `null`.[^2][^3]

### 10.2 Grouping boundary

Grouping is a presentation-only transform over the current filtered result set. It MUST NOT create, delete, or mutate source records, projection rows, links, or tags.[^2]

The current base profile allows exactly one active grouping key or `Group: None`. For Timeline, the allowed grouping keys are closed and contract-backed.

| Timeline grouping key | Required meaning |
| --- | --- |
| `timeline.occurred_day` | Group by occurred-at calendar day |
| `timeline.recorded_day` | Group by recorded-at calendar day |
| `timeline.capture_state` | Group by capture lifecycle state |
| `timeline.has_evidence` | Group by whether evidence is linked |
| `timeline.has_unresolved_mentions` | Group by unresolved-mention presence |

This narrow whitelist is a design strength. It keeps grouping useful for triage, queueing, and overview while preventing grouping from drifting into a second mutation model or a free-form spreadsheet gimmick.[^2][^7]

### 10.3 View state layers

Cartulary has three distinct layers of workbook state. The UI MUST keep them conceptually separate.

| Layer | What it is | Persistence rule |
| --- | --- | --- |
| Canonical surface identity | The base workbook surface identified by `sheet_ref.kind='view_schema'` | Stable and sharable |
| Saved-view configuration | One incident-bound configuration over exactly one `view_schema_id` | Persisted as saved view when explicitly saved |
| Local session state | Selection, scroll, focused cell, open inspector state, preview state, local popovers, presence | Client-local only; not persisted in saved views |

This separation matters because users experience each layer differently. Surface identity tells them where they are. Saved views tell them how this surface is shaped. Local session state tells them what they are doing right now. Mixing them would make startup, sharing, and multi-device use harder to explain.[^2][^3]

### Acceptance criteria

- Column-header sorts and inline filter chips operate on the active workbook surface without navigating away.[^2]
- Grouping never mutates source state, and there is never more than one active grouping key in the current base profile.[^2]
- Timeline exposes only the five allowed grouping keys plus `Group: None`.[^2]
- Saved views preserve surface configuration but not selection, scroll position, focused cell, open inspector state, preview state, or presence.[^2][^3]

## 11. Coordination Surfaces as Workbook-Native UX

### 11.1 Why they belong in the workbook

Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson MUST remain workbook-native surfaces because they are part of the same incident-working grammar as Timeline, entities, and evidence. Analysts move between facts, actions, decisions, communication, and follow-through continuously. Requiring a module change would fracture the common operating picture the workbook is supposed to preserve.[^2][^8][^21]

The core rule is simple: coordination belongs **inside** the workbook shell but **off** the routine timeline hot path. Ordinary row capture MUST NOT require task, decision, owner, approver, challenge, or checklist fields on the Timeline sheet itself. Communications logs, handoffs, status reviews, and lessons remain explicit workbook-native actions, but the system MUST NOT interrupt ordinary row editing to solicit them.[^2][^8]

### 11.2 Surface roles

| Surface | Primary purpose | Required UX posture | Must not become |
| --- | --- | --- | --- |
| Task Requests | Owned units of work with queue, due, blocked, and requester semantics | Queue-oriented workbook surface with sort, filter, group, and link-back behavior | A detached ticketing subsystem |
| Decisions | Rationale-bearing incident decisions | Review-oriented workbook surface with clear status and supersession | A generic approval engine |
| Communications Log | Durable communication memory | Same-shell communication history with linked decisions and tasks | A chat replacement |
| Handoff | Shift or phase continuity artifact | Explicit continuity artifact linked to open work and risks | A mandatory per-edit ritual |
| Status Review | Checkpoint review and workload rebalance | Cadence artifact for blockers, risks, and next report timing | A dashboard KPI page |
| Lesson | Follow-through on retrospective findings | Workbook-native learning artifact linked to tasks and evidence | A free-text graveyard |

Task Requests and Decisions are required base surfaces. Communications Log, Handoff, Status Review, and Lesson are artifact-backed coordination surfaces with workbook-native identities. All six use the same shell, the same query/filter/group/history logic, and the same expectation that the selected row remains the user’s anchor of work.[^2][^3][^8]

### 11.3 Same-surface creation and linking

From a selected Timeline, Host, Identity, Evidence, Notes, Task Requests, or Decisions row, the analyst MUST be able to create or link a task request, decision, coordination artifact, or party reference without leaving workbook flow. Preseeded linked-record or party context is editable context only; it does not satisfy minimum create signal by itself.[^2]

Task Requests and Evidence remain text-first on the grid when they expose requester, collector, or source semantics. Party linking is same-surface enrichment from the inspector or Parties view. It MUST NOT block row creation, row commit, or later text editing.[^2][^4]

Exact-match reuse for parties is deliberately narrower than for ordinary text suggestions. Create-from-text or link-existing MAY reuse the same incident-scoped party only on a unique exact match of normalized `primary_email` or `external_ref`. Raw requester or source text remains preserved alongside any linked `party_id`. Display name, organization, role, and phone-like text are suggestion inputs, not auto-link keys.[^2][^4][^20]

### 11.4 Operating-model guidance versus product behavior

The workbook-native coordination surfaces are the durable product contract. Tracker hygiene, handoff quality, status-review cadence, workload redistribution, and challenge or escalation practice remain operating-model guidance rather than runtime workflow requirements. Cartulary should support those practices with explicit surfaces and saved views, but it MUST NOT turn routine row capture into a mandatory ceremony.[^1][^8][^21]

### Acceptance criteria

- Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson appear and behave as workbook surfaces rather than separate application modules.[^2][^20]
- Timeline row capture does not require coordination-only fields or ritual checkpoints.[^2]
- From a selected workbook row, a user can create or link a task, decision, coordination artifact, or party reference without losing row context or leaving the workbook shell.[^2]
- Party create, link, unlink, and clear actions preserve visible text-plus-link independence and keep the originating row context, focus, and scroll position.[^2]

## 12. Evidence Interaction Design

### 12.1 Evidence is not a cell path

Evidence in Cartulary is a user-facing record family with attached object state behind it. It is not a file path in a cell, and it is not raw object-store URL UX.[^3][^5]

This separation produces visible design rules. Evidence counts reflect committed evidence state, not upload intent. Preview and download are redeemed through same-origin opaque handles issued by the application. The browser MUST treat those handles as opaque and MUST NOT synthesize or store raw object-store URLs as supported evidence-access state.[^3][^5]

### 12.2 Two-step attachment and visible state

The attach flow is explicitly two-step: create a blob slot, upload bytes, then attach the blob to an evidence record. Pending or failed upload slots MUST NOT look like committed evidence. Evidence counts, `has evidence` flags, and preview affordances update only after successful finalization.[^2][^3][^7]

The base user-visible states are:

| Evidence or blob condition | Grid meaning | Preview or download consequence |
| --- | --- | --- |
| Requested or pending receipt, no blob yet | Valid evidence row exists without bytes | No preview; row remains first-class evidence |
| Pending unattached blob slot | Upload in progress or awaiting finalization | Does not increment evidence count; not yet attached |
| Available attached blob | Committed evidence is present | Preview or download as allowed by current preview policy |
| Unsupported or blocked preview | Evidence exists but inline preview is not currently allowed | Surface the blocked state inline; do not silently fall back |
| Failed, inconsistent, quarantined, or missing blob state | Evidence or blob state needs repair or retry | Preview and download fail closed until state is repaired |

Cartulary MUST also allow evidence rows to exist before bytes do. Requested or pending-receipt evidence is still real evidence. That matters for incident work, because teams often know that a collection has been requested, staged, or promised before they actually have the file.[^4][^7]

### 12.3 Preview and download without leaving the workbook

Evidence preview MUST open without forcing full-page navigation away from the grid. A side or bottom preview is acceptable. When preview is unavailable or blocked, the grid or inspector MUST remain in place and surface the state inline rather than silently falling back to download or navigating away.[^2][^5]

The `Space` shortcut is important because it makes evidence preview an in-flow workbook action rather than a page change. Dragging or pasting a screenshot onto the selected row SHOULD feel like a workbook gesture, not like entering a media-management module.[^2][^7]

### 12.4 Reporting and release consequences

Evidence UX in the live workbook and evidence behavior in rendered outputs are intentionally different concerns. Live workbook visibility is governed by incident membership. Recipient-specific withholding belongs to snapshot-derived rendering and release time, not to hiding live workbook content from incident participants. This distinction is essential to keeping the workbook usable while still allowing disciplined release outputs later.[^3][^5]

### Acceptance criteria

- Evidence counts and `has evidence` indicators update only after successful blob finalization and attachment.[^2][^3]
- Preview and download use application-mediated same-origin handles, not raw object-store URLs.[^3][^5]
- Preview opens without full-page navigation, and blocked or unsupported states are surfaced inline.[^2]
- Evidence rows may exist before binary bytes are present, and that state remains first-class rather than “not yet evidence.”[^4][^7]

## 13. Excel Inspiration: Preserve / Adapt / Reject

### 13.1 Preserve

The spreadsheet behaviors below MUST be preserved because they are the behaviors that make the workbook operationally valuable in incident work.[^6][^9][^12]

| Behavior | Required treatment | Product problem solved |
| --- | --- | --- |
| Direct typing into visible cells | Keep as the default editing mode | Low-ceremony capture under time pressure |
| Keyboard navigation | Keep as first-class shell behavior | Fast movement through the working set |
| Clipboard paste | Keep on the base hot path | Fast ingestion from existing tools |
| Fill-down and bulk editing | Keep as explicit workbook operations | Repeated incident work without repetition tax |
| Fast local sorting and filtering | Keep in-surface | Rapid triage and working-set reshaping |
| Visible tabular working surface | Keep central | Shared operational memory |
| Quick row creation with minimal ceremony | Keep central | Capture facts before they are lost |

### 13.2 Adapt

The spreadsheet behaviors below SHOULD be adapted rather than copied literally because Cartulary needs stronger semantics underneath the same visible grammar.[^2][^3][^4][^7]

| Behavior | Required adaptation | Product problem solved |
| --- | --- | --- |
| Workbook tabs | Built-in tabs plus saved and system views over shared state | Preserve workbook feel without storage silos |
| Relationship cells | Chips, unresolved tokens, and explicit resolution actions | Make relational state visible and correctable |
| Evidence | Attached object records with preview and same-origin handles | Replace path-in-cell fragility |
| Saved views and grouping | Contract-backed workbook behavior | Make view state portable and deterministic |
| Presence, autosave, and conflict state | Explicit same-surface collaboration telemetry | Make multi-analyst editing legible |
| Coordination surfaces | Workbook-native surfaces | Keep facts, work, and decisions in one shell |
| Progressive normalization | Exact-match reuse, explicit create/resolve flows, preserved raw text | Add structure without blocking capture |

### 13.3 Reject

The spreadsheet behaviors below MUST be rejected because preserving them would directly undermine Cartulary’s purpose.[^2][^3][^5][^6]

| Behavior | Required rejection | Product problem avoided |
| --- | --- | --- |
| Formulas, macros, merged-cell semantics, workbook automation | Not part of the product model | Hidden logic, security risk, unrecoverable semantics |
| Hidden macro semantics for bulk actions | Not allowed | Invisible mutation behavior |
| Raw object-store URLs as evidence UX | Not supported | Storage leakage and broken authorization boundaries |
| Forms-first CRUD as default interaction | Not allowed | Structure before facts are ready |
| Generic workflow-engine logic on the hot path | Not allowed | Ceremony replacing flow |
| Character-by-character collaborative cell editing as primary target | Not the design center | Complexity spent in the wrong place |
| Import/storage/module boundaries that mimic Excel internals | Not allowed | Spreadsheet internals leaking below the view layer |

### Acceptance criteria

- The product still feels like a workbook for direct typing, paste, keyboard navigation, sorting, filtering, and row creation.[^2][^6]
- The product does **not** inherit formulas, macros, merged-cell semantics, hidden automation, or raw object-store URLs as supported behavior.[^3][^5]
- Relationship cells, saved views, grouping, evidence interaction, collaboration, and coordination surfaces all preserve the workbook mental model while making underlying state explicit and auditable.[^2][^3][^4]

## 14. Design Implications of the Current Implementation Baseline

### 14.1 Baseline and UX consequences

The current implementation direction is concrete enough to shape design. The browser-facing baseline is React plus `react-data-grid` plus Vite. The backend baseline is Go. The deployment baseline is one application unit plus PostgreSQL and S3-compatible object storage. The browser runtime uses native `fetch`, `WebSocket`, Clipboard, and keyboard APIs. Exact package versions remain owned by repo-control files, but these stack families are the current design baseline.[^20]

The table below states the main UX consequences.

| Boundary | Current baseline | UX consequence | The design must not depend on |
| --- | --- | --- | --- |
| Deployable shape | One application unit plus PostgreSQL plus S3-compatible object storage | Same-origin integrated shell, no service seam in the user’s mental model | Multi-service choreography exposed in the UI |
| Browser grid | React plus `react-data-grid` | Virtualization, cell renderers, keyboard editing, paste, grouping, `record_id` anchoring | Enterprise-grid-only affordances |
| Browser APIs | Native `fetch`, `WebSocket`, Clipboard, Keyboard | Direct paste, same-origin collaboration stream, explicit focus and selection management | Wrapper-library semantics as product requirements |
| Row identity | Projection-backed rows with stable `record_id` and `row_version` | Focus, selection, and edits stay anchored to the record, not row index | Index-based identity |
| Evidence access | Application-mediated opaque handles | Explicit preview/download states and same-origin redemption | Raw object-store URL construction |
| Collaboration state | Explicit conflict, presence, and pending-queue semantics in application code | User-visible trust signals stay domain-specific | Generic client-state libraries owning product semantics |
| License envelope | Permissive-licensed OSS baseline; no AG Grid Enterprise | Design assumes OSS grid capabilities rather than commercial magic | Features that require an enterprise-only grid |

The permissive-license constraint matters. The UX should assume an open-source grid capability envelope and SHOULD NOT promise enterprise-grid features the current baseline intentionally excludes. That constraint is not a weakness if the design keeps its complexity budget on row identity, mutation semantics, evidence handles, history, and conflict resolution, which is where Cartulary actually differentiates.[^20]

### 14.2 Optional-pack degradation

Optional reference packs are overlays, not preconditions for the core workbook. If optional packs are absent, disabled, failed, or missing, Cartulary MUST continue to support timeline capture, entity resolution, evidence attachment, and core editing. Only affected overlay labels, enrichment semantics, non-canonical analytical widgets, or snapshot/report derivations may degrade.[^3][^5]

The UX implication is straightforward. The base workbook MUST never require ATT&CK, D3FEND, VERIS, or other framework overlays to stay usable. A smallest supported disconnected deployment is still usable without those packs. Missing packs may remove overlay labels or enrichment affordances, but they MUST NOT create or remove current-profile workbook surfaces and MUST NOT break capture, linking, or evidence workflows.[^3][^5]

### 14.3 Snapshot/reporting boundary

Snapshot and reporting behavior is extension-profile scope, but the current workbook UX MUST leave room for it without pretending it already exists in the base hot path.[^1][^3]

When snapshot/reporting is present, it operates on immutable snapshots and a canonical export model rather than live workbook state. `release_scope` is a narrow artifact-scoped concept with a default of `internal_draft`. It does **not** belong on ordinary row editing surfaces. The live workbook therefore MUST distinguish clearly between source evidence, derived analytic material, curated narrative, and working material, even if only the reporting extension turns that distinction into rendered output rules later.[^3][^5]

The workbook-side consequences are concrete: generated outputs must be self-contained, narrative released externally must carry support references, direct-source coordination text defaults to working material unless explicitly curated, and recipient-specific withholding happens at snapshot/render/release time rather than by hiding live workbook content from incident members.[^3][^5]

### 14.4 Clipboard hot path versus file-based import

Clipboard paste is base-profile behavior. File-based structured import beyond clipboard paste belongs to a dedicated `imports` module and the Import Extension Profile. The base workbook shell MUST therefore stay isolated from import-assistant choices.[^3][^20]

The mapping engine may be shared between clipboard paste and file import. The base shell grammar must not be. The default workbook surface MUST NOT grow import wizards, mapping confirmation dialogs, workbook-inspection chrome, or import status bars as permanent shell furniture. Those belong to import workflows when the Import Extension Profile is claimed, not to the base workbook experience.[^3][^20]

### Acceptance criteria

- The design assumes the current concrete baseline of React plus `react-data-grid` plus Vite on the browser side and one Go application unit plus PostgreSQL plus S3-compatible object storage on the backend side.[^20]
- Focus, selection, pending edits, and live updates remain anchored by `record_id` and `row_version`, not by row index or current sort position.[^2][^3][^20]
- Missing optional reference packs degrade overlays and enrichment only; the core workbook remains usable and no current-profile workbook surfaces appear or disappear because of missing framework packs.[^3][^5]
- Snapshot/reporting rules remain artifact-scoped and extension-scoped; they do not colonize live workbook editing.[^1][^3][^5]
- File-based import beyond clipboard paste remains separate module and extension territory; the base workbook shell does not become an import wizard.[^3][^20]

## 15. Risks, Non-Goals, and UX Failure Modes

### 15.1 Primary risk: false equivalence

The dominant UX failure mode is **false equivalence**: a design that looks spreadsheet-like while quietly moving the real work elsewhere.[^6]

The pattern usually looks like this: the grid remains visible, but the real editor is the inspector; relationship cells appear text-first, but the practical path is picker-first; the workbook shell remains present, but tasking, decisions, or evidence preview actually happen in detached modules; or low-friction capture exists in demos, but implementation shortcuts gradually turn forms and background automation into the real product. That is the failure mode this guide is designed to prevent.[^2][^6]

### 15.2 Concrete failure modes and guardrails

| Failure mode | What it looks like | Why it fails | Required guardrail |
| --- | --- | --- | --- |
| Inspector-first drift | Users must open the inspector to do ordinary row creation or routine field editing | The grid stops being the primary work surface | Keep the inspector secondary and adjacent |
| Silent normalization | Tokens become entities without clear disclosure or cheap reversal | Analyst trust collapses | Use explicit exact-match boundaries, disclosure, `Undo`, and later `Revert to unresolved` |
| Queue invisibility | Work is queued or blocked but the user cannot tell | Users do not know whether the system protected their edits | Keep save-state and overflow states visible on the same surface |
| Grouping as data model | Dragging or regrouping implies mutation | View state and source state become confused | Keep grouping presentation-only |
| Evidence storage leakage | The UI treats object-store paths or raw handles as evidence identity | Authorization and preview semantics become fragile | Keep evidence application-mediated |
| Coordination module drift | Tasks and decisions feel like a second application | The common operating picture fractures | Keep coordination workbook-native |
| Dashboard drift | The shell fills with KPIs, charts, and management chrome | The product stops being a working surface | Limit shell telemetry to state needed for the current incident surface |
| Import colonization | The default shell grows wizard chrome and mapping UI | Base workbook flow is slowed by non-base workflows | Keep file import separate from clipboard hot-path grammar |

### 15.3 Non-goals

Cartulary is not trying to be any of the following in the current profile.[^1][^5][^6]

- It is not a dashboard product.
- It is not a generic ticketing system.
- It is not an Excel clone with formulas, macros, merged cells, and workbook automation.
- It is not a raw telemetry repository.
- It is not a generic approval engine for ordinary edits.
- It is not a multi-master local-first collaboration platform in the current profile.
- It is not primarily optimizing for several users typing into the same cell character by character.
- It is not using file-based structured import to define the base workbook interaction model.
- It is not using live recipient-specific withholding to narrow what incident members can see in the workbook.

### Acceptance criteria

- Reviewers can identify whether a design change preserves the grid as the primary capture surface and the inspector as a secondary enrichment surface.[^2][^6]
- Reviewers can reject any proposal that silently normalizes rough tokens, hides queue or conflict state, or turns grouping into mutation semantics.[^2][^6]
- Reviewers can distinguish legitimate shell telemetry from KPI-dashboard drift and legitimate coordination surfaces from detached workflow modules.[^2][^8]

## 16. Conclusion

Cartulary should be built and reviewed as a workbook-native caseworking system, not as a form-heavy case application that happens to render a grid.[^2][^6]

That conclusion is not rhetorical. It drives concrete design choices: a virtualized grid as the primary interaction surface; text-first relationship entry with progressive normalization; a non-blocking inspector with a sharply bounded ownership model; same-surface evidence preview through application-mediated handles; field-level optimistic concurrency with cell-local same-field conflict resolution; workbook-native coordination surfaces; contract-backed sorting, filtering, grouping, saved views, and startup rules; and a concrete implementation baseline that favors an integrated browser experience over distributed-service spectacle.[^2][^3][^4][^5][^20]

If Cartulary holds those lines, it can replace the Spreadsheet of Doom without forgetting why the workbook was valuable in the first place. If it does not, it will either become a spreadsheet clone without stronger guarantees or a structured case tool that quietly abandoned the speed and directness incident work actually needs.[^6][^9][^10][^11]

## Sources

[^1]: `00_document_set_status_and_precedence.md`; `05_claim_publication_and_benchmark_reproducibility.md`.
[^2]: `03_workbook_interaction_collaboration_and_workflows.md`.
[^3]: `01_architecture_storage_and_view_contracts.md`.
[^4]: `02_domain_model_schema_and_history.md`.
[^5]: `04_security_deployment_and_conformance.md`.
[^6]: `A_problem_framing_rationale_tradeoffs_and_sanity_check.md`; `G_source_archive_exploratory_design_artifact.md`.
[^7]: `D_workflow_and_ui_illustrations_source_extract.md`.
[^8]: `H_operating_model_supporting_guidance.md`.
[^9]: `R06-spreadsheet_of_doom_dfir_research_report.md`; `R07-spreadsheet-of-doom-sod-report.cr.md`.
[^10]: `R01-aurora_incident_response_report.md`.
[^11]: `R03-Kanvas_technical_research_report.md`.
[^12]: `R04-responsive_browser_spreadsheet_ui_research_memo.md`; `R05-responsive-interface-design-report.cr.md`.
[^13]: Ben Shneiderman, “Direct Manipulation: A Step Beyond Programming Languages,” *IEEE Computer* 16(8), 1983.
[^14]: Sruti Srinivasa Ragavan, Advait Sarkar, and Andrew D. Gordon, “Spreadsheet Comprehension: Guesswork, Giving Up and Going Back to the Author,” CHI 2021.
[^15]: Zhicheng Liu and Jeffrey Heer, “The Effects of Interactive Latency on Exploratory Visual Analysis,” *IEEE Transactions on Visualization and Computer Graphics* 20(12), 2014.
[^16]: Lyn Bartram, Michael Correll, and Melanie Tory, “Untidy Data: The Unreasonable Effectiveness of Tables,” arXiv:2106.15005.
[^17]: George Chalhoub and Advait Sarkar, “It’s Freedom to Put Things Where My Mind Wants: Understanding and Improving the User Experience of Structuring Data in Spreadsheets,” CHI 2022.
[^18]: Danyel Fisher et al., “Trust Me, I’m Partially Right: Incremental Visualization Lets Analysts Explore Large Datasets Faster,” CHI 2012; Dominik Moritz et al., “Trust, but Verify: Optimistic Visualizations of Approximate Queries for Exploring Big Data,” CHI 2017.
[^19]: Jakob Nielsen, “Progressive Disclosure,” Nielsen Norman Group, 2006.
[^20]: `cartulary-dev-guide.md`; `cartulary_repository_bootstrap_guide.md`; `cartulary_implementation_testing_guide.md`.
[^21]: `R02-cartulary_crm_tem_dfir_research_report.md`.
