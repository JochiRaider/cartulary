---
title: Cartulary UI/UX Design Guide
revision: 3
document_class: non-normative-design-support-guide
status: current
normative_design_owner: design.md
runtime_behavior_owners:
  - 00_document_set_status_and_precedence.md
  - 01_architecture_storage_and_view_contracts.md
  - 02_domain_model_schema_and_history.md
  - 03_workbook_interaction_collaboration_and_workflows.md
  - 04_security_deployment_and_conformance.md
claim_publication_owner: 05_claim_publication_and_benchmark_reproducibility.md
source_freshness_date: 2026-08-07
supersedes: cartulary-ui-ux-design-guide revision 2
---

# Cartulary UI/UX Design Guide

## 1. Executive Summary

This guide explains Cartulary's user-facing mental model, design rationale,
interaction consequences, rejected alternatives, and implementation
implications. It does not define product behavior, design conformance, runtime
defaults, public interfaces, security behavior, limits, omission semantics, or
implementation acceptance. `design.md` is the sole current design-direction
owner. Core 00 through Core 04 and adopted subsystem or extension NLSpecs
remain the behavioral owners for their assigned scopes. If this guide differs
from a current owner, the owner governs and this guide contains documentation
drift.

No runtime component, generator, test, validator, release gate, conformance
command, or retained-evidence audit derives product behavior by reading,
parsing, hashing, resolving, or otherwise consuming this guide.

Cartulary presents incident-response work as a workbook: a stable shell around
dense, first-class analytical surfaces, with direct editing, progressive
structure, visible collaboration state, and contextual workflows. The guide is
for product designers, frontend implementers, reviewers, and writers who need
to understand why owner-defined behavior forms a coherent experience.

The guide is not a conformance checklist. Core 04 owns implementation
conformance. Core 05 owns publication of claim-bearing timed or
fixture-sensitive evidence. Section 20 records only whether this support guide
is internally complete and correctly sourced.

## 2. Reading Guide and Source Method

### 2.1 Authority and execution boundary

The following matrix is exhaustive for the authority classes used here.

| Topic | Primary current owner | Guide role |
| --- | --- | --- |
| Document status, precedence, profiles, and conformance separation | Core 00 | Explain the authority boundary and point to the owner. |
| Public routes, wire objects, errors, jobs, discovery, and view schemas | Core 01 | Explain visible consequences without reproducing protocol contracts. |
| Records, evidence, parties, mentions, observations, links, and history | Core 02 | Explain the domain distinctions used by the interface. |
| Workbook interaction, collaboration, startup, editing, clipboard, conflict, and saved views | Core 03 | Restate exact interaction outcomes and point to their requirements. |
| Authentication, authorization, access loss, administration, and security | Core 04 | Explain visible consequences without defining permissions. |
| Claim-bearing timed or fixture-sensitive publication | Core 05 | Distinguish ordinary design evidence from publication claims. |
| Tokens, shell composition, responsive transformation, component presentation, focus, and accessibility presentation | `design.md` | Explain the rationale and reference the exact design contract. |
| Extension-specific resources and workspaces | The named adopted extension owner | Explain conditionally applicable surfaces without widening the profile. |
| Grid adapters and frontend verification mechanics | Frontend implementation and testing guides | Record implementation consequences only. |
| Rationale, comparisons, rejected alternatives, and illustrations | This guide and non-normative appendices | Preserve context without creating behavior. |

The owner locator in a restatement is the section or requirement that a reviewer
reads to obtain the complete contract. A summary in this guide is never a
replacement for that contract.

### 2.2 Closed statement classes

Revision 3 uses exactly six statement classes.

| Statement class | Permitted content | Required source information | Excluded content |
| --- | --- | --- | --- |
| Owner restatement | Compact account of behavior already owned elsewhere. | Owner file, locator, applicability, and material omission behavior. | A new default, limit, state, permission, error, interface, or alternative. |
| Design rationale | Explanation of why the adopted design has its current form. | The relevant owner when current behavior is discussed. | Conformance language or a new implementation obligation. |
| Implementation note | Non-binding adapter, component, package, or test consequence. | Implementation-support source when one exists. | Product behavior or a public interface. |
| Illustrative example | A scenario, fragment, wireframe, or sequence that demonstrates an owner-backed concept. | Explicit authority label and source owner. | An unlabeled partial protocol object or a conflicting default. |
| Future consideration | A concept absent from the current profile. | `future-only`, current omission behavior, and the required future owner class. | Present support, reserved compatibility, or conformance language. |
| Guide-maintenance requirement | A rule for this guide's structure, terminology, sourcing, or review. | The affected guide section. | Product or runtime behavior. |

Guide-maintenance requirements:

- This guide **MUST** preserve the six statement classes and the closed
  applicability vocabulary in §2.3.
- This guide **MUST NOT** use guide-authored normative terms to define product
  behavior.
- A future revision **MAY** omit a nonessential illustration; if omitted, its
  owner reference and behavioral summary remain in the relevant main section
  or Appendix A.
- A future revision **MAY** omit a research source that no longer informs any
  text; if omitted, Appendix C records its removal and no current statement
  relies on it.
- Partial protocol examples **MUST** carry the exact label specified in §2.5.
- Every diagram **MUST** declare profile, viewport, density, authority status,
  data classification, and source owner.

### 2.3 Applicability vocabulary

| Applicability | Meaning in this guide | Omission or unavailable behavior |
| --- | --- | --- |
| `base` | Current general behavior or rationale. | The owner's unavailable states apply. |
| `base-required-surface` | A current surface required by the Base Profile. | A conformant Base implementation cannot omit it. |
| `standardized-optional-surface` | A standardized current surface whose implementation is optional. | When unimplemented, the surface and affordances are absent; no placeholder is inferred. |
| `extension-profile` | Conditional behavior owned by an exact current extension profile. | When unclaimed or unavailable, the owner-defined contribution is absent or falls back as the owner defines. |
| `future-only` | A concept outside every current profile. | No current route, state, interface, affordance, compatibility promise, or conformance claim exists. |
| `implementation-support` | Adapter, package, test, or development guidance. | It creates no user-facing behavior. |
| `non-normative-guidance` | Rationale, examples, and explanatory consequences. | It creates no implementation conformance. |

No other applicability token is used. An unclassified proposal belongs in §19
as `future-only` until an owner adopts it.

### 2.4 Source classes and freshness

Core 00 through Core 05, `design.md`, and adopted subsystem or extension
NLSpecs are owner sources. Appendices A through I are non-normative context and
traceability. Reports R01 through R09 are research evidence. Implementation
guides describe repository practice. Appendix C records the complete source
register reviewed through 2026-08-07.

### 2.5 Example and identity discipline

Every partial request or response carries this exact label:

> **Illustrative fragment; not a complete valid request or response.**

Illustrative fragment; not a complete valid request or response.

```json
{"sheet_ref":{"kind":"view_schema","id":"timeline"}}
```

Surface identity comes from stable typed identifiers. It does not come from a
visible label, route, component name, CSS selector, storage table, or row
position.

| Surface kind | Stable identity shape | Interpretation source |
| --- | --- | --- |
| Direct view schema | `{kind: "view_schema", id: view_schema_id}` | Core 01 §7.4 and Core 03 §2 |
| Saved view | `{kind: "saved_view", id: saved_view_id}` | Core 01 §3.3.5.2 and Core 03 §2.3 |
| Extension workspace | `{kind: "extension_workspace", extension_profile_id, workspace_key}` | Core 01 discovery contracts and the named extension owner |

## 3. Product and Interaction Thesis

**Design rationale.** A workbook keeps collection, comparison, correction, and
review in one spatial model. Analysts can preserve source language first and
add structure only when the investigation benefits from it. Stable identities
allow labels and layout to change without changing what a surface or record is.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Workbook-first interaction | Core 03 | §1, REQ-03-001 through REQ-03-003 | `base` | Core 03 requires direct workbook interaction while preserving authoritative server state. | Owner-defined loading, stale, unavailable, and authorization states replace fabricated data. | Discuss the workbook as the primary mental model, not as ownership of data semantics. |
| Stable surface identity | Core 01 and Core 03 | Core 01 §7.4; Core 03 §2 | `base` | The owners distinguish typed surface identities from presentation coordinates. | An undiscovered or unavailable identity has no inferred surface. | Refer to IDs, never labels or positions, when explaining continuity. |
| Progressive structure | Core 02 and Core 03 | Core 02 §6-§10; Core 03 §9 and §12-§13 | `base` | Raw capture, mentions, observations, canonical records, and explicit links remain distinct. | Missing resolution leaves the source-backed item unresolved rather than inventing a record. | Explain capture-first analysis without collapsing domain objects. |
| Same-surface recovery | Core 03 and `design.md` | Core 03 §3.3 and §4.4; `design.md` §10.4-§10.5 | `base` | Conflict and transaction recovery remain associated with the active surface and affected work. | Ineligible or inactive-surface recovery does not displace the current surface. | Prefer continuity and explicit recovery over global interruption in rationale. |

## 4. Workbook-First Rationale

**Design rationale.** Incident response repeatedly alternates between scanning,
sorting, comparing, transcribing, and correcting. A table is therefore an
analytical surface, not a temporary staging form. The shell adds relational and
historical context without forcing the analyst into a form-by-form navigation
model.

The adopted design preserves high-value spreadsheet qualities: visible rows,
direct cell targeting, deterministic keyboard traversal, clipboard intake, and
compact state cues. It rejects implicit typing, formula execution, positional
identity, silent coercion, and hidden persistence. Those rejections follow Core
02's typed domain model and Core 03's editing and clipboard contracts.

**Implementation note.** A grid library supplies rendering mechanics. It does
not own record identity, field capability, edit admission, selection semantics,
shortcuts, validation, error classification, or collaboration state.

## 5. Overall Application Shell and Information Architecture

`design.md` §7 defines the shell and its responsive transformations. Query
controls belong in the View bar. The top bar carries application, incident,
account, and global navigation; it is not a second query surface.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Authenticated root | Core 01 and `design.md` | Core 01 §3.3.2.1A; `design.md` §4.6 | `base` | Zero visible incidents produces a successful empty directory; one or more remain in the directory until explicit selection. Selection opens the workbook without a launch `sheet_ref`. | Permission loss clears protected materialization and returns to `/`; no inaccessible incident is retained as visible content. | Treat the root as incident selection, never as a dashboard with inferred priority. |
| Workbook shell regions | `design.md` | §7.1-§7.3 | `base` | The design owner defines top bar, surface navigation, View bar, grid, inspector, and status strip as one composition. | A region absent under a declared responsive state remains reachable through the owner-defined overflow location. | Illustrations preserve region purpose and order without copying layout literals. |
| Deployment administration | Core 01, Core 04, and `design.md` | Core 01 §3.3.2.1B; Core 04 §2 and §9.10; `design.md` §4.5 | `base` | Deployment administration is an application context reached from the account/application menu and contains only owner-declared panels. | Without the application capability, its entry and protected state are absent; it is never a workbook surface or post-login default. | Keep administration visually subordinate to incident work and separate from incident membership. |
| Account settings | Core 01 and `design.md` | Core 01 §3.3.2.3; `design.md` §4.4 | `base` | Profile, density preference, and links to existing security flows form the closed current composition. | Unsupported preferences and self-service identity mutations are absent. | Do not illustrate a generalized settings console. |

The shell works because each region answers one question: where am I, which
surface is active, how is it shaped, what is selected, and what is the current
save/collaboration state. Appendix B provides non-authoritative compositions.

## 6. Workbook Surface Model

Direct surfaces, saved views, and extension workspaces can look table-like but
have different identities and lifecycle owners.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Required built-in surfaces | Core 01 and Core 03 | Core 01 §7.4; Core 03 §2.1-§2.2 | `base-required-surface` | Owner registries define exact surface identities, field membership, order, and discoverability. | Required Base surfaces are not omitted. Owner-defined no-data and unauthorized states remain distinct. | Use registry IDs and owner names; labels are display text only. |
| Standardized optional surfaces | Core 01 and Core 03 | Core 01 §7.4; Core 03 §2.2 | `standardized-optional-surface` | An implemented optional surface uses its standardized identity and contract. | If unimplemented, the surface and affordances are absent. | Do not draw disabled placeholders that imply support. |
| Saved views | Core 01, Core 02, and Core 03 | Core 01 §3.3.5.2; Core 02 §11; Core 03 §2.3 and §14 | `base` | A saved view references an underlying surface and owner-defined query/layout state; dirty comparison is deterministic. | An unavailable underlying surface prevents activation under the owner's error and fallback rules. | Explain saved views as named state, not copied datasets. |
| Network Analysis | Network Flow Activity NLSpec | §1, §7-§10 | `extension-profile` | A claimed `network_flow_activity` profile contributes its typed workspace, tables, and graph adapters. | When unclaimed or unavailable, extension navigation and state are absent; active loss clears extension state and selects the owner-defined Base fallback. | Never present the workspace as a Base surface. |

Surface switching preserves identity through `sheet_ref`; no guide example
infers an ID from a tab label, route segment, React component, database name,
or current row coordinate.

## 7. Grid Editing Model

**Design rationale.** Direct manipulation is valuable only when commit state,
validation, conflict, and recovery remain visible. A cell editor is a temporary
draft boundary over an owner-defined field capability; it is not the record.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Edit admission and commit | Core 03 | §1, §3.3, §4.1, and field workflows | `base` | Core 03 requires capability-aware editing, exact draft/commit handling, and server-authoritative outcomes. | Read-only, unavailable, invalid, group, and non-record targets do not enter an application edit. | Illustrations distinguish focus, selection, editor, pending, saved, and conflict states. |
| Escape behavior | `design.md` and Core 03 | `design.md` §8.5-§8.6; Core 03 interaction requirements | `base` | The design owner defines a one-layer escape ladder with deterministic focus restoration. | With no eligible application layer, the application takes no action and browser behavior remains unsuppressed. | Avoid the vague instruction “Esc cancels”; name the active layer. |
| Clipboard paste and fill | Core 03 | §11.1 and the bounded fill requirements | `base` | Core 03 owns parsing, shape validation, field capability, atomicity, conflict, and fill bounds. | Invalid or ineligible targets retain committed data and expose owner-defined correction paths; file import is not inferred from paste. | Treat paste as typed workbook input, not arbitrary spreadsheet execution. |
| File import | Core 01 and Core 03 | Core 01 §17.2; Core 03 §11.2 | `extension-profile` | The Import profile owns staged mapping, preview, apply, job, and result behavior. | When the profile is unclaimed, file-import affordances and routes are absent; clipboard paste remains the Base path. | Separate clipboard flow from import-assistant illustrations. |

### 7.1 Shortcut context

Core 03 REQ-03-220 and AC-538 close the application-shortcut boundary.

| Shortcut | Eligible semantic context | Owner-required result | Ineligible result |
| --- | --- | --- | --- |
| `Ctrl/Cmd+K` | Grid navigation on a committed record row and committed cell whose field has an owner-declared link or resolve capability. | Invoke that capability from the grid/work-area boundary. | No application action and no suppression of browser behavior. |
| `Space` | Grid navigation on an eligible committed row with Evidence in its inspector-group registry. | Open Evidence; one previewable item opens directly, otherwise focus enters the evidence list or empty state. | No application action and no suppression of browser behavior. |
| `Alt+H` | Grid navigation on an eligible committed row with History in its inspector-group registry. | Open History and move focus to that group. | No application action and no suppression of browser behavior. |

Unavailable, group-row, draft-row, no-row, editor, menu, and inspector contexts
are ineligible. The semantic context contains navigation mode, committed-row
identity, row kind, field capability, available inspector groups, and
previewable-evidence count; it is not a set of presentation booleans.

## 8. Progressive Structuring and Relational UX

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Mentions and entities | Core 02 and Core 03 | Core 02 §6, §8-§9; Core 03 §9 and §12 | `base` | A mention preserves source text and can remain unresolved, dismissed, auto-resolved, or explicitly resolved under owner rules; an entity is canonical incident state. | Failed or unavailable resolution leaves the mention state explicit and does not fabricate an entity. | Chips show state while preserving source language and correction paths. |
| Observations and indicators | Core 02 and Core 03 | Core 02 §7.4 and §10.2; Core 03 §9.1 and §13 | `base` | An observation is source-bound; an indicator is canonical. Linking is explicit under the owner workflow. | No supported promotion or link action means no canonical indicator is inferred. | Avoid treating every indicator-shaped string as a canonical record. |
| Evidence record and blob | Core 01 and Core 02 | Core 01 §3.3.8 and §16; Core 02 §13 | `base` | Metadata, attachment lifecycle, and stored bytes remain separate dimensions. | A record can exist without available bytes; blocked preview never implies absent metadata. | Present compound states instead of one generic attachment badge. |
| Inspector configuration | Core 01 and Core 03 | Core 01 §6 and §7.4; Core 03 §2.3A | `base` | Discovery declares eligible groups; Core 03 owns opening, row context, invalidation, and workflow activation. | Missing group capability means the group and shortcut are unavailable; row selection alone does not open the inspector. | Use the inspector for context and workflows without turning it into an implicit navigation side effect. |

## 9. Collaboration, Presence, Autosave, and Conflict Resolution

Cartulary uses optimistic collaboration: analysts keep working while ownership,
conflict, and recovery stay explicit. “Saved” never means merely “typed into a
local editor.”

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Save-state label | Core 03 and `design.md` | Core 03 §4.1-§4.2; `design.md` §10.1 | `base` | The owners select one primary label from `Syncing`, `Saved`, and `Conflict`. | Invalid combinations are rejected by the owner algorithm rather than displayed ambiguously. | Keep the primary label singular and stable. |
| Status allocation | `design.md` | §10.1, D-AC-077 | `base` | The strip allocates one primary save label, at most one eligible active-surface secondary message, and deterministic presence with overflow. A separate visible queue-count slot does not exist. | Inactive-surface messages are ineligible; responsive truncation preserves accessible full text and presence overflow. | Do not stack competing secondary banners inside the strip. |
| Secondary priority | `design.md` | §10.1 | `base` | Priority is `client_txn_conflict`, queue overflow, same-field conflict, other terminal replay failure, authentication pause, refresh pause, then queued or in-flight work. | If no active-surface candidate exists, no secondary message is shown. | Conflict entry points route to their same-surface recovery surface. |
| Presence | Core 03 and `design.md` | Core 03 §4.3; `design.md` §10.2-§10.3 | `base` | The owners define stable ordering, deduplication, row/cell attribution, and `+N` overflow. | Absence of current presence data does not imply another user's state. | Presence is awareness, not a lock or authorization signal. |
| Same-field conflict | Core 03 and `design.md` | Core 03 §3.3; `design.md` §10.4 | `base` | Saved and unsaved values remain visible in the cell and same-surface resolver; focus moves only after explicit activation and returns to the cell. | No eligible conflict means no resolver entry point. | Preserve comparison and local draft context. |
| Client transaction conflict | Core 03 and `design.md` | Core 03 §4.4; `design.md` §10.5 | `base` | A non-modal same-surface panel retains the blocked unit and later FIFO work with exact actions `Retry with a new request ID` and `Discard blocked edit`; it does not steal focus. | It never enters the same-field resolver or retries automatically. | Explain recovery as queue repair, not value merging. |
| Queue capacity | Core 03 | §4.4 | `base` | The local queue retains 64 units and refuses the 65th; overflow enters explicit recovery. | It does not evict or automatically retry retained work. | Capacity pressure is visible secondary state, not a silent data-loss policy. |

## 10. Sorting, Filtering, Grouping, and View State

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Query state | Core 01 and Core 03 | Core 01 workbook query contracts; Core 03 §14 | `base` | Typed sort, filter, search, group, and cursor state is bound to the active surface and request generation. | Unsupported fields or operators are unavailable and stale responses do not replace current generations. | Present query controls in the View bar and explain their active-surface scope. |
| Grouping | Core 03 | §14.3 | `base` | Core 03 defines the allowed grouping keys and group-row interaction boundary. | A group row is not a committed record row and is ineligible for record actions and shortcuts. | Style group rows as structural summaries, never as records. |
| Saved layout | Core 01, Core 02, and Core 03 | Core 01 §3.3.5.2; Core 02 §11; Core 03 §14.4-§14.7 | `base` | Owner-defined normalized state determines saved-view identity and dirty comparison. | Unsaved local state remains distinguishable; absent optional fields follow owner defaults. | Avoid label-based or serialized-object-order comparisons. |
| Loading generation | `design.md` | §10.6 and D-AC-078 | `base` | A stable generation identity binds delayed loading to the same active-surface request generation. | Generation change or terminal state cancels the prior delayed state. | Loading copy is never carried across surface generations. |

## 11. Coordination Surfaces as Workbook-Native UX

Coordination records remain typed workbook material rather than a generalized
workflow engine. Parties belong with coordination because people and
organizations are explicit incident-scoped reference points.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Tasks, decisions, and communications | Core 01, Core 02, and Core 03 | Core 01 §7.4; Core 02 §10.4; Core 03 §16.4 and §20 | `base-required-surface` | Owners define typed surfaces, lifecycle fields, and same-surface create/link workflows. | Unauthorized or unavailable actions are absent; records remain readable only under current authorization. | Keep work visible in sortable rows with contextual actions. |
| Parties | Core 01, Core 02, and Core 03 | Core 01 §18; Core 02 §8.2-§8.3; Core 03 §16.2 | `base-required-surface` | Party identity and relationship workflows use stable record identifiers. | Unresolved mentions remain distinct; no party is inferred from display text alone. | Place Parties with coordination and relationship context. |
| Handoff, status review, and lessons | Core 01, Core 02, and Core 03 | Core 01 §7.4 and §19; Core 02 §10.4.4A; Core 03 §2.3A and §20 | `base-required-surface` | Owner-specific feature groups provide review and acknowledgement workflows. | A missing declared feature group removes its entry point rather than substituting a generic approval flow. | Preserve each owner's vocabulary and workflow boundary. |

## 12. Evidence Interaction Design

Evidence is a lifecycle and custody model, not a file-shaped cell value. The
interface distinguishes request, metadata creation, upload, verification,
availability, preview authorization, download authorization, quarantine, and
failure.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Attachment lifecycle | Core 01, Core 02, and Core 03 | Core 01 §3.3.8; Core 02 §13; Core 03 §8.1-§8.3 | `base` | The two-step workflow preserves the evidence record and blob lifecycle boundary. | Missing or failed bytes remain explicit and do not erase authorized metadata. | Display compound lifecycle state and recovery at the record context. |
| Preview and download | Core 01, Core 04, and `design.md` | Core 01 §16; Core 04 §4.3 and §9.10; `design.md` §11.2-§11.3 | `base` | Access is reauthorized and preview/download remain distinct actions. | Blocked preview retains authorized metadata, exposes no bytes, and offers download only when separately authorized. | Keep blocked state inside the inspector and preserve focus. |
| Evidence shortcut | Core 03 | REQ-03-220 and AC-538 | `base` | `Space` opens Evidence only from the eligible committed-row grid context; one previewable item opens directly, otherwise focus enters the list or empty state. | Every ineligible context produces no application action and leaves browser behavior unsuppressed. | Explain the shortcut by semantic context, not key interception alone. |

## 13. Excel Inspiration: Preserve, Adapt, Reject

This comparison is design rationale, not a compatibility contract.

| Preserve | Adapt | Reject |
| --- | --- | --- |
| Dense scanning and comparison | Cell editing through typed field capabilities | Formula execution and macro behavior |
| Keyboard traversal | Clipboard intake through validated atomic operations | Positional row identity |
| Visible sort and filter state | Saved views as normalized owner-defined state | Silent coercion and inferred schemas |
| Direct manipulation | Structured mentions, evidence, and relationships | Hidden persistence or local-only “saved” claims |
| Familiar rows and columns | Inspector context beside the active grid | Arbitrary cross-surface copy semantics |

**Illustrative example.** Pasting three rows into writable Timeline fields can
feel spreadsheet-native while Core 03 still validates shape, field capability,
atomicity, and conflict behavior. The familiar gesture does not import Excel's
type coercion or formula model.

## 14. Design Implications of the Current Implementation Baseline

This section has applicability `implementation-support`. It creates no
user-facing behavior.

The shared grid adapter translates typed surface metadata and semantic
interaction context into the selected grid vendor. Application code remains
responsible for owner-driven data state, stable identity, authorization loss,
error families, request generations, and workflows. Generated protocol and UI
contracts carry runtime-consumed values; Markdown does not.

Implementation and verification procedures live in:

- `docs/guides/cartulary_frontend_implementation_guide.md`;
- `docs/guides/cartulary_frontend_testing_strategy.md`;
- `docs/guides/cartulary_visual_golden_maintenance.md`;
- `docs/testing-harness-nlspec.md`.

Dependency versions, component ownership, test commands, and grid-vendor
mechanics are deliberately absent here. Their omission prevents this guide
from becoming a stale package or harness registry.

## 15. Risks, Non-Goals, and Positive Patterns

### 15.1 Risks and counter-patterns

| Risk | Owner-backed counter-pattern |
| --- | --- |
| A dense grid hides semantic differences. | Use typed capabilities, semantic states, accessible names, and inspector context from Core 01/Core 03/`design.md`. |
| Optimistic work appears saved before authority accepts it. | Use Core 03 save states and the design-owned status allocation. |
| Labels become accidental identifiers. | Carry typed record, field, surface, and extension identities end to end. |
| Global banners detach failures from work. | Use the design-owned locus and active-surface message priority. |
| Presence is mistaken for authorization or locking. | Treat presence only as Core 03 awareness state. |
| Extension UI leaks into Base behavior. | Admit contributions only through adopted discovery and omission contracts. |

### 15.2 Current non-goals

Core 00 and the named owners exclude mobile/touch conformance, arbitrary
spreadsheet compatibility, formula execution, a generalized workflow builder,
an all-incident administration catalog, theme selection beyond the required
design theme, and behavior inferred from UI labels. Their observable omission
is absence: no route, affordance, compatibility promise, or Base conformance
claim is implied.

### 15.3 Positive patterns

Stable identity, explicit capability, deterministic fallback, local recovery,
retained authorized context, non-color state cues, owner-generated contracts,
and narrow adapter boundaries reinforce one another. They allow density without
making state or authority implicit.

## 16. Visual Language Direction

`design.md` §§3, 5, 6, 9, 10, 11, and 12 own the semantic token registry,
resolution algorithm, required theme, density modes, component variants,
compound-state precedence, chip states, status presentation, and evidence
states. The generated design-token and presentation projections are the
runtime inputs.

This guide intentionally contains no duplicate pixel, color, spacing,
typography, density, motion, z-index, or duration registry. Illustrations use
semantic names such as `dark_graphite`, owner-defined density identifiers, and
state names. An implementation resolves exact values through the generated
contracts and adapter layers.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Theme and tokens | `design.md` | §3 and §6 | `base` | `dark_graphite` and the semantic token resolver define current visual values. | Optional themes remain absent under the design owner's omission rules. | Refer to tokens, not copied color or spacing literals. |
| Density | `design.md` | §3.9 and §4.4 | `base` | Owner-defined density modes determine shell and grid presentation; account preference uses the exact public contract. | A cleared preference returns to the owner-defined default; custom row heights are absent. | Describe information density semantically. |
| Component states | `design.md` | §12 | `base` | Closed variants and compound precedence define visual, focus, action, and accessibility results. | Unsupported compound states are rejected or resolve through declared precedence. | Do not invent one-off component variants in examples. |
| Error presentation | `design.md` | §10.8 | `base` | Typed error and operation families select locus, retention, actions, focus, and live behavior. | Unknown errors use typed operation and authorization context, never human text. | Preserve public-error sanitization and use the generated mapping. |

## 17. Input, Viewport, and Accessibility Posture

Base support remains keyboard-and-pointer desktop. `design.md` §7.4-§7.5 owns
the exact responsive shell and inspector selection for each supported viewport
state. A mobile/touch profile is not current.

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Responsive shell | `design.md` | §7.4-§7.5 | `base` | Inline and block-size algorithms select one shell chrome and inspector presentation with deterministic overflow. | Below-minimum layouts retain safe navigation but do not claim design conformance. | Do not offer guide-local drawer, sheet, modal, or overlay alternatives. |
| Keyboard and focus | Core 03 and `design.md` | Core 03 §1 and REQ-03-220; `design.md` §8.4-§8.6 and §14 | `base` | Owners define mode-aware key behavior, focus destinations, restoration, and non-interception. | Ineligible contexts take no application action and do not suppress browser behavior. | Review focus as part of every visible state. |
| Accessible state | `design.md` | §14.1-§14.3 | `base` | Required states have non-color cues, accessible names, declared live behavior, and token-pair contrast coverage. | No state can rely on color, animation, or raw record identity alone. | Illustrations name live and focus behavior in metadata or captions. |
| Reduced motion | `design.md` | §6.3 and §14 | `base` | Motion is not required to perceive or operate state; the owner defines reduced-motion outcomes. | When motion is reduced, the same state remains perceptible without animation. | Avoid animation-dependent explanations. |

## 18. Cross-Cutting UX Patterns

### 18.1 Loading and transient state

| Topic | Owner | Owner locator | Applicability | Current summary | Omission/unavailable behavior | Guide consequence |
| --- | --- | --- | --- | --- | --- | --- |
| Delayed initial loading | `design.md` | §10.6, D-AC-078, D-VFIX-013 | `base` | `Still loading this surface` appears after exactly 2,000 ms of monotonic time in the same active-surface generation and is announced politely once. | It is absent through 1,999 ms; generation change or terminal state cancels it; appearance alone creates no retry action. | Tie delayed copy to generation identity, not component mount duration. |
| Transient confirmation | `design.md` | §10.7 and D-AC-079 | `base` | An actionless confirmation dismisses after 5,000 ms of cumulative visible, unpaused time. Pointer hover, keyboard focus within, and `document.hidden` pause; resume uses the remainder. | A still-valid action, error, progress state, or unresolved critical state remains persistent. | Use one shared controller so pause and resume semantics do not diverge. |

### 18.2 Error-presentation projection

The following compact projection is navigation to `design.md` §10.8, not an
independent contract. Error classification uses typed error and operation
families and preserves the public-error sanitization boundary. Summary and
detail text are never classification inputs.

| Family | Locus and retained state | Actions and focus | Live behavior |
| --- | --- | --- | --- |
| Local validation | Affected editor or cell; committed value and exact local draft remain. | Correct or cancel; no automatic focus move. | Assertive. |
| Same-field conflict | Cell marker and same-surface resolver; saved and unsaved values remain. | Owner-declared resolver actions; focus moves only after activation and returns to the cell. | Assertive. |
| `client_txn_conflict` | Same-surface non-modal recovery panel; blocker and later FIFO units remain. | `Retry with a new request ID` and `Discard blocked edit`; no focus theft. | Assertive once. |
| Queue overflow | Status secondary and same-surface overflow notice; 64 units remain and the 65th is refused. | Enter recovery; never evict or retry automatically. | Assertive. |
| Stale refresh | Non-blocking grid and status state; previously authorized materialization remains. | `Retry`; preserve selection and focus. | Assertive. |
| Initial-load failure | Blocking grid state with no fabricated rows. | `Retry` only after failure; no automatic focus. | Assertive. |
| Authentication required | Same-surface session-recovery notice; unsent work and previously authorized materialization remain. | Reauthenticate; no focus theft. | Assertive. |
| Permission or incident-access loss | Protected materialization clears and navigation returns to `/`. | No invented retry; focus the authenticated-root heading. | Assertive. |
| Extension unavailable | Known-unavailable navigation is omitted; active extension state clears and the ordinary Base fallback runs. | No automatic retry; focus the selected Base surface heading. | Assertive. |
| Evidence preview blocked | Inspector-local state; authorized metadata remains and preview bytes do not. | Offer download only when separately authorized; preserve focus. | Polite. |
| Unknown future error | Typed operation family and authorized-materialization state select blocking, stale, or operation-local locus; human text is never inspected. | No destructive action; retain only verified authorized data and clear when authorization is unknown. | Assertive. |

### 18.3 Cross-cutting sequences

These sequences summarize owner interactions without creating a new
orchestrator.

1. A surface activation establishes typed identity and a new request
   generation; stale responses cannot replace it.
2. Initial loading remains blocking; the delayed sentence appears only at the
   design-owned boundary for the same generation.
3. A typed terminal outcome selects an owner-defined data state and the
   generated presentation row.
4. Focus and retained data follow authorization, active-surface, and error-locus
   rules.
5. Any actionless confirmation uses the shared cumulative visible-time
   controller.

## 19. Deferred Scope

| Concept | Applicability | Required future owner | Current omission behavior |
| --- | --- | --- | --- |
| Mobile and touch-first workbook profile | `future-only` | A future adopted interaction and design owner coordinated with Core 00 | No touch-specific layout, gesture, viewport, or conformance promise exists. |
| Additional themes | `future-only` | A future `design.md` revision and generated token projection | No selector or compatibility promise exists beyond current owner-defined themes. |
| Generalized workflow builder | `future-only` | A future domain, public-contract, interaction, security, and design owner set | No generic workflow schema, route, surface, or approval engine exists. |
| Arbitrary spreadsheet files | `future-only` | A future import extension revision | No formula, macro, workbook-format, or round-trip compatibility exists. |
| Undeclared optional workbook surface | `future-only` | Core 00 profile adoption plus Core 01/Core 03/`design.md` ownership | No placeholder, discovery entry, route, or saved-view compatibility is reserved. |
| Adopted extension contribution | `extension-profile` | Its named adopted extension NLSpec and imported Core owners | When unclaimed or unavailable, its contribution is omitted or falls back exactly as its owner defines. |
| Standardized optional Base surface | `standardized-optional-surface` | Core 01/Core 03 and `design.md` | When unimplemented, its surface and affordances are absent. |

Future consideration text remains here or in Appendix C. It does not appear as
a present-tense capability elsewhere in this guide.

## 20. Editorial Acceptance Criteria

These criteria test this document, not the product. A result is accepted only
through the human review record in Appendix E.

| ID | Review method | Pass condition | Failure condition |
| --- | --- | --- | --- |
| `UXG-RP-AC-001` | Metadata review | Revision, class, status, owners, freshness, and supersession fields match §1. | A field is missing or differs. |
| `UXG-RP-AC-002` | Authority review | The opening states the non-normative boundary and owner precedence. | The guide claims behavioral or design authority. |
| `UXG-RP-AC-003` | Execution-boundary review | The guide prohibits every listed executable consumer from deriving behavior from it. | An executable consumer is allowed or omitted. |
| `UXG-RP-AC-004` | Structure review | Main sections 1 through 20 and Appendices A through E exist once and in order. | A section is missing, duplicated, or reordered. |
| `UXG-RP-AC-005` | Statement-class review | Every guide statement fits one of the six classes in §2.2. | A seventh or ambiguous class remains. |
| `UXG-RP-AC-006` | Normative-word review | Guide-authored normative terms govern guide maintenance only. | Product behavior is stated as a guide requirement. |
| `UXG-RP-AC-007` | Wording review | No advisory normative terms or unresolved open-choice phrases remain. | Advisory or undecidable wording remains. |
| `UXG-RP-AC-008` | Optionality review | Every guide-maintenance `MAY` states omission behavior. | Maintenance optionality lacks an omission result. |
| `UXG-RP-AC-009` | Applicability review | Every applicability value belongs to the closed §2.3 vocabulary. | An unknown token appears. |
| `UXG-RP-AC-010` | Owner-table review | Every owner-restatement table uses all seven required columns. | A required column is absent. |
| `UXG-RP-AC-011` | Locator review | Every current behavioral summary has an exact owner locator. | A summary is unowned or points only to this guide. |
| `UXG-RP-AC-012` | Identity review | Surface and record explanations use stable typed identifiers. | A label, route, component, selector, storage name, or position is treated as identity. |
| `UXG-RP-AC-013` | Root-composition review | §5 restates the directory at `/` for zero, one, and multiple incidents, with explicit workbook selection. | Count-dependent automatic selection or an omitted count case appears. |
| `UXG-RP-AC-014` | Shell review | §5 assigns query controls to the View bar and keeps administration outside the workbook. | Shell ownership or placement differs from `design.md`. |
| `UXG-RP-AC-015` | Surface review | §6 distinguishes direct, saved, optional, and extension identities with omission behavior. | Surface classes collapse or omission is ambiguous. |
| `UXG-RP-AC-016` | Editing review | §7 traces editing, clipboard, fill, import, and shortcuts to owners. | A guide-local editing algorithm remains. |
| `UXG-RP-AC-017` | Shortcut review | Every eligible and ineligible context in AC-538 is represented. | A shortcut context or result is missing. |
| `UXG-RP-AC-018` | Domain review | §8 preserves mention/entity, observation/indicator, and evidence/blob distinctions. | Distinct owner objects are collapsed. |
| `UXG-RP-AC-019` | Status review | §9 states the closed allocation, active-surface eligibility, priority, and no queue-count slot. | Allocation or priority is incomplete. |
| `UXG-RP-AC-020` | Collaboration review | §9 distinguishes same-field and transaction conflict recovery. | The two recovery classes share an invented resolver. |
| `UXG-RP-AC-021` | Evidence review | §12 traces lifecycle, preview, download, authorization, focus, and retained metadata. | A compound evidence state is omitted or widened. |
| `UXG-RP-AC-022` | Token review | §16 references owner tokens and projections without copying registries. | Pixel, color, density, or timing registries are duplicated. |
| `UXG-RP-AC-023` | Responsive review | §17 points to one owner-defined inspector and overflow outcome per viewport state. | A guide-local presentation alternative remains. |
| `UXG-RP-AC-024` | Timing review | §18 restates exact 1,999/2,000 and 5,000 ms boundaries, cancellation, and pauses. | Any boundary or state condition is ambiguous. |
| `UXG-RP-AC-025` | Error review | §18 contains every error family and each locus, retention, action/focus, and live result. | A row or dimension is missing. |
| `UXG-RP-AC-026` | Classification review | Error text is excluded as a classification input and sanitization remains intact. | Summary or detail text can select presentation. |
| `UXG-RP-AC-027` | Example review | Every partial protocol example has the exact warning label. | A partial protocol fragment is unlabeled. |
| `UXG-RP-AC-028` | Diagram review | Every diagram has all six metadata fields. | Any metadata field is missing. |
| `UXG-RP-AC-029` | Retirement review | Revision 2 acceptance identifiers have no active occurrence or compatibility alias. | An active identifier or evidence route remains. |
| `UXG-RP-AC-030` | Independent review | Two reviewers record identity, revision digest, results, and unresolved defects in Appendix E. | Fewer than two complete independent records exist. |

## Appendix A — Owner Cross-Reference Matrix

This appendix is navigation only.

| Behavior family | Owner and locator | Applicability | Guide section | Omission or unavailable behavior |
| --- | --- | --- | --- | --- |
| Authority and precedence | Core 00 §§1-§5.1 | `base` | §1-§2 | Non-owner material cannot override an owner. |
| Authenticated root | Core 01 §3.3.2.1A; `design.md` §4.6 | `base` | §5 | Access loss clears protected state and returns to `/`. |
| Shell and responsive composition | `design.md` §7 | `base` | §5, §17 | Controls move only through declared overflow states. |
| Surface identity and registry | Core 01 §7.4; Core 03 §2 | `base` | §2, §6 | Unknown or undiscovered surfaces are absent. |
| Editing and clipboard | Core 03 §1, §3.3, §4, §11 | `base` | §7 | Ineligible operations take no application action or produce the owner error. |
| Application shortcuts | Core 03 REQ-03-220; AC-538 | `base` | §7, §12, §17 | Ineligible contexts do not suppress browser behavior. |
| Progressive structure | Core 02 §6-§10; Core 03 §9, §12-§13 | `base` | §8 | No canonical record is inferred from source text. |
| Collaboration and queue | Core 03 §3-§4; `design.md` §10 | `base` | §9 | Overflow refuses new work without eviction. |
| Query and saved state | Core 01 workbook query contracts; Core 02 §11; Core 03 §14 | `base` | §10 | Stale generations do not replace current state. |
| Coordination | Core 01 §7.4 and §18-§19; Core 02 §8 and §10; Core 03 §16 and §20 | `base-required-surface` | §11 | Missing workflow capability creates no generic substitute. |
| Evidence | Core 01 §3.3.8 and §16; Core 02 §13; Core 03 §8; Core 04 §4 | `base` | §12 | Unauthorized bytes remain unavailable and protected state is cleared as required. |
| Tokens and component states | `design.md` §3, §6, §9, §12 | `base` | §16 | Unregistered visual values and states are not inferred. |
| Accessibility | `design.md` §14; Core 03 interaction requirements | `base` | §17 | Color-only, focusless, or unnamed controls are invalid under the owner. |
| Loading, transient, and errors | `design.md` §10.6-§10.8 | `base` | §18 | Cancellation, persistence, and unknown-error fallback follow exact owner rows. |
| Claim publication | Core 05 | `non-normative-guidance` | §1 | Design or test evidence is not a public benchmark claim without Core 05 publication. |

## Appendix B — Illustrative Wireframes and Sequences

Every figure in this appendix is illustrative and non-authoritative.

### B.1 Workbook shell

| Metadata | Value |
| --- | --- |
| Profile | `base` |
| Viewport | `1440x900 CSS px` |
| Density | `compact` |
| Authority status | Illustrative; non-normative |
| Data classification | Synthetic incident names, actor names, timestamps, and IDs |
| Source owner | `design.md` §7 and §10 |

```text
+ Top bar: app | incident | account ---------------------------------------+
| Surface navigation                                                    |
+ View bar: saved state | search | filter | sort | group ----------------+
| Grid                                               | Inspector (closed) |
|                                                    | explicit opener   |
+ Status: primary save | one secondary | presence ------------------------+
```

### B.2 Inspector and conflict

| Metadata | Value |
| --- | --- |
| Profile | `base` |
| Viewport | `1280x720 CSS px` |
| Density | `compact` |
| Authority status | Illustrative; non-normative |
| Data classification | Synthetic saved value, draft, actor, and record ID |
| Source owner | Core 03 §3.3 and `design.md` §7.3, §10.4 |

```text
Grid cell [conflict marker] || Inspector: History / Evidence / Workflow
                             || Resolver opened by explicit activation
                             || Saved value | Local draft | owner actions
                             || Close -> focus returns to originating cell
```

### B.3 Evidence sequence

| Metadata | Value |
| --- | --- |
| Profile | `base` |
| Viewport | `1280x720 CSS px` |
| Density | `compact` |
| Authority status | Illustrative; non-normative |
| Data classification | Synthetic filename, hash, status, and record ID |
| Source owner | Core 01 §3.3.8 and §16; Core 02 §13; Core 03 §8; `design.md` §11 |

```text
request -> evidence metadata -> upload -> verification -> available
                                      \-> failed/quarantined
available -> authorize preview -> preview bytes or inspector-local blocked state
          -> authorize download -> download only when separately authorized
```

### B.4 Deployment administration

| Metadata | Value |
| --- | --- |
| Profile | `base` |
| Viewport | `1280x720 CSS px` |
| Density | `compact` |
| Authority status | Illustrative; non-normative |
| Data classification | Synthetic user, provider, job, and audit data |
| Source owner | Core 01 §3.3.2.1B; Core 04 §2 and §9.10; `design.md` §4.5 |

```text
Account/application menu -> Deployment administration
                            [owner-declared panel tabs]
                            [panel content and local state]
No workbook tab | no all-incident catalog | no generalized settings console
```

### B.5 Responsive movement

| Metadata | Value |
| --- | --- |
| Profile | `base` |
| Viewport | `1024x720 CSS px` |
| Density | `compact` |
| Authority status | Illustrative; non-normative |
| Data classification | Synthetic labels and presence names |
| Source owner | `design.md` §7.4-§7.5 |

```text
Wide:    [surface tabs] [View bar controls] [grid || adjacent inspector]
Narrow:  [Surfaces]     [owner overflow]     [grid || selected inspector mode]
Status:  [primary] [truncated secondary with full accessible text] [presence +N]
```

### B.6 Transaction recovery sequence

| Metadata | Value |
| --- | --- |
| Profile | `base` |
| Viewport | `1280x720 CSS px` |
| Density | `compact` |
| Authority status | Illustrative; non-normative |
| Data classification | Synthetic transaction and record IDs |
| Source owner | Core 03 §4.4 and `design.md` §10.5 |

```text
FIFO replay -> client transaction conflict -> halt at blocked unit
             -> retain blocked and later units
             -> same-surface panel, no focus theft
             -> Retry with a new request ID | Discard blocked edit
```

## Appendix C — Research and Source Register

All rows are informative unless the authority column names an owner document.

| Source | Class and authority | Guide use | Freshness |
| --- | --- | --- | --- |
| Core 00 | Normative owner | Status, precedence, profiles, conformance boundary | Reviewed 2026-08-07 |
| Core 01 | Normative owner | Public contracts, discovery, views, evidence, administration | Reviewed 2026-08-07 |
| Core 02 | Normative owner | Domain distinctions, lifecycle, history, persistence vocabulary | Reviewed 2026-08-07 |
| Core 03 | Normative owner | Workbook interaction, collaboration, recovery, workflows | Reviewed 2026-08-07 |
| Core 04 | Normative owner | Security, authorization, access loss, conformance | Reviewed 2026-08-07 |
| Core 05 | Normative publication owner | Claim-publication boundary only | Reviewed 2026-08-07 |
| `docs/design.md` | Sole normative design-direction owner | Tokens, composition, component states, accessibility presentation | Reviewed 2026-08-07 |
| Appendix A | Non-normative rationale | Problem framing, trade-offs, sanity check | Reviewed 2026-08-07 |
| Appendix B | Non-normative illustration | Architecture context | Reviewed 2026-08-07 |
| Appendix C | Non-normative illustration | Schema context | Reviewed 2026-08-07 |
| Appendix D | Non-normative illustration | Workflow and UI source extracts | Reviewed 2026-08-07 |
| Appendix E | Non-normative history and future backlog | Deferred scope archaeology | Reviewed 2026-08-07 |
| Appendix F | Non-normative traceability | Source-to-owner navigation | Reviewed 2026-08-07 |
| Appendix G | Non-normative archive | Original exploratory artifact | Reviewed 2026-08-07 |
| Appendix H | Non-normative operating guidance | Human operating context | Reviewed 2026-08-07 |
| Appendix I | Non-normative evidence | Projection authority and boundary context | Reviewed 2026-08-07 |
| R01 | Research report | Incident-response product comparison; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |
| R02 | Research report | CRM/TEM/DFIR comparison; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |
| R03 | Research report | Kanvas technical research; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |
| R04 | Research memo | Responsive browser spreadsheet UI; direct evidence and author inference remain distinguished in the memo | Reviewed 2026-08-07 |
| R05 | Research report | Responsive interface design; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |
| R06 | Research report | Spreadsheet-of-doom DFIR analysis; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |
| R07 | Research report | Spreadsheet-of-doom review; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |
| R08 | Research report | Handsontable/React investigation; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |
| R09 | Research report | React Data Grid investigation; direct evidence and author inference remain distinguished in the report | Reviewed 2026-08-07 |

Research supports rationale. It does not independently create current behavior.

## Appendix D — Revision History and Migration Crosswalk

Revision 3 supersedes Revision 2. Revision 2's guide-local normative design
direction is withdrawn. Its acceptance identifiers are retired from active
design, product verification, test selection, evidence accounting, and
compatibility routing. Repository history preserves their exact text; this
guide does not reproduce compatibility aliases.

| Revision 2 material | Revision 3 disposition |
| --- | --- |
| Status and four-marker reading model | Replaced by §1-§2 authority boundary, six statement classes, and closed applicability vocabulary. |
| Product thesis and workbook rationale | Preserved as sourced rationale in §3-§4. |
| Shell, root, and status behavior | Moved to `design.md`; restated in §5 and §9. |
| Surface identities and startup | Returned to Core 01/Core 03; restated in §6. |
| Editing, keyboard, clipboard, and autosave | Returned to Core 03 and `design.md`; restated in §7. |
| Progressive structure and inspector behavior | Returned to Core 01/Core 02/Core 03; restated in §8. |
| Collaboration, presence, and conflicts | Returned to Core 03 and `design.md`; restated in §9. |
| Query, grouping, and saved-view algorithms | Returned to Core 01/Core 02/Core 03; restated in §10. |
| Coordination and evidence | Returned to their Core owners; restated in §11-§12. |
| Visual literals and responsive alternatives | Replaced by design-token, projection, and exact owner references in §16-§17. |
| Loading, transient, and error behavior | Moved to `design.md` and its machine projection; restated in §18. |
| Product acceptance matrix | Replaced by guide-only editorial criteria in §20. |

## Appendix E — Editorial Completion Record

Revision 3 is editorially accepted only after two independent reviewers record
identity, the same revision digest or commit, all criterion results, and every
unresolved defect. A pending row is not acceptance evidence.

| Review | Reviewer identity | Revision digest or commit | Criteria result | Unresolved defects | Status |
| --- | --- | --- | --- | --- | --- |
| Independent review 1 | Pending | Pending | `UXG-RP-AC-001` through `UXG-RP-AC-030`: pending | Pending | Not accepted |
| Independent review 2 | Pending | Pending | `UXG-RP-AC-001` through `UXG-RP-AC-030`: pending | Pending | Not accepted |

The completion record is corpus-maintenance evidence. It is not runtime,
conformance, release, or Core 05 claim-publication evidence.
