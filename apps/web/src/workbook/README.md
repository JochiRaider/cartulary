# workbook/

[Source overview / parent](../README.md)

Workbook shell composition, cross-surface behavior tests, and navigation to workbook implementation owners.

The shell composes incident-scoped lifecycle and presentation owners. Runtime
code consumes package facades for grid, protocol, UI, and view contracts.
Source-owner features retain their own authoring and recovery logic.

Start with [WorkbookShell.tsx](WorkbookShell.tsx), then follow the relevant
subdirectory. Root tests cover interactions spanning workbook owners.

## Subdirectories

| Directory | Responsibility |
| --- | --- |
| [adapters/](adapters/README.md) | Private workbook protocol adapters, captured operation transport, response validation, and semantic outcome conversion. |
| [collaboration/](collaboration/README.md) | Workbook interpretation of decoded collaboration events, presence, reset, and authorization recovery. |
| [composition/](composition/README.md) | Committed retained runtime construction and replacement-safe presentation attachment. |
| [components/](components/README.md) | Shared workbook surface facades, shell chrome, query controls, field editors, and recovery presentation. |
| [continuity/](continuity/README.md) | Semantic grid focus, selection, and viewport continuity through workbook-private adapter bindings. |
| [evidence/](evidence/README.md) | Workbook-wide Evidence access feedback and live-region presentation. |
| [features/](features/README.md) | Workbook feature entry points and owner-specific authoring, inspector, and recovery workflows. |
| [find/](find/README.md) | Runtime-only loaded-cell matching and navigation intent, supplied with source-owned committed text and semantic membership. |
| [history/](history/README.md) | Record History browsing, captured rollback/restore operations, acknowledgement, and recovery ownership. |
| [hooks/](hooks/README.md) | React coordination for shell lifetime, queries, startup, saved views, imports, and shared surface behavior. |
| [inspector/](inspector/README.md) | Canonical inspector subjects, declared capability admission, related-record workflows, and History composition. |
| [layout/](layout/README.md) | Workbook density, responsive layout, column geometry, surface sizing, and shared style slots. |
| [lifecycle/](lifecycle/README.md) | Shared typed invalidation reasons for workbook, query, mutation, and extension lifetimes. |
| [models/](models/README.md) | Pure workbook request, row, query, registry, startup, saved-view, and presentation models. |
| [mutations/](mutations/README.md) | Semantic mutation command assembly, secure action identity, write coordination, and operation outcomes. |
| [policies/](policies/README.md) | Contract-backed surface policy declarations, reference requirements, owner bindings, and application shortcuts. |
| [ports/](ports/README.md) | Semantic workbook read/write capabilities and shared outcomes used across owner boundaries. |
| [preferences/](preferences/README.md) | Current-user home and incident-default workbook preference reads, edits, announcements, and recovery. |
| [query/](query/README.md) | Instance-owned surface queries, committed-record capabilities, latest-request admission, and live-row reconciliation. |
| [runtime/](runtime/README.md) | Workbook mutation lifetime, pending replay, conflict recovery, write coordination, and status projection. |
| [savedviews/](savedviews/README.md) | Saved-view listing and captured create/update/delete operations with retained acknowledgement and recovery. |
| [services/](services/README.md) | Instance-scoped reference-query sharing, cancellation, invalidation, and disposal. |
| [startup/](startup/README.md) | Workbook startup admission and commit planning across base surfaces, saved views, and extension availability. |
| [surfaces/](surfaces/README.md) | Registration-driven composition of concrete workbook surface renderers. |
| [timeline/](timeline/README.md) | Timeline composition, presentation, actions, editing, and semantic runtime integration. |
| [utils/](utils/README.md) | Reusable workbook clipboard, queue, presence, reconciliation, recovery, formatting, and style helpers. |
| [view-state/](view-state/README.md) | Instance-local, schema-keyed workbook query state and defaults. |

## Files

| File | Responsibility |
| --- | --- |
| [WorkbookShell.tsx](WorkbookShell.tsx) | Route-facing Workbook composition over incident-scoped lifecycle and presentation owners. |

## Tests

| File | Responsibility |
| --- | --- |
| [workbookInspectorOwnershipArchitecture.test.ts](workbookInspectorOwnershipArchitecture.test.ts) | Tests inspector construction and composition remain in their declared feature owners. |
| [workbookSaveStatus.test.tsx](workbookSaveStatus.test.tsx) | Tests workbook save status across surfaces and overlapping generic operations. |
| [WorkbookShell.actionSequencing.test.tsx](WorkbookShell.actionSequencing.test.tsx) | Tests Timeline actions use the current row version after earlier workbook mutations. |
| [WorkbookShell.assessments.test.tsx](WorkbookShell.assessments.test.tsx) | Tests Assessment creation and confidence-band payload behavior on the workbook surface. |
| [WorkbookShell.autosave.test.tsx](WorkbookShell.autosave.test.tsx) | Timeline mutation autosave and pending-save characterization. |
| [WorkbookShell.collaboration.test.tsx](WorkbookShell.collaboration.test.tsx) | Tests keyed presence and collaboration state coexist with pending workbook mutations. |
| [WorkbookShell.evidence.test.tsx](WorkbookShell.evidence.test.tsx) | Tests Evidence attachment counts and lifecycle feedback across workbook surfaces. |
| [WorkbookShell.grid.test.tsx](WorkbookShell.grid.test.tsx) | Timeline mutation grid/create behavior characterization. |
| [WorkbookShell.gridProvenance.test.tsx](WorkbookShell.gridProvenance.test.tsx) | Tests Hosts, Identities, and Notes grids retain contract-backed row provenance. |
| [WorkbookShell.history.test.tsx](WorkbookShell.history.test.tsx) | Tests row-centric History access and server-owned event metadata in the inspector. |
| [WorkbookShell.inspector.test.tsx](WorkbookShell.inspector.test.tsx) | Workbook interaction inspector and row-local action tests. |
| [WorkbookShell.mentionChips.test.ts](WorkbookShell.mentionChips.test.ts) | Mention chip model tests. |
| [WorkbookShell.payload.test.tsx](WorkbookShell.payload.test.tsx) | Timeline mutation request payload characterization. |
| [WorkbookShell.query.test.tsx](WorkbookShell.query.test.tsx) | Tests stable schema identities in saved-view, sort, filter, and grouping controls. |
| [WorkbookShell.saveState.test.tsx](WorkbookShell.saveState.test.tsx) | Tests derived save-state detail in the workbook status strip. |
| [WorkbookShell.sentinel.test.tsx](WorkbookShell.sentinel.test.tsx) | Tests keyboard grid anchors and focus continuity use stable record and field identities. |
| [WorkbookShell.support.test.tsx](WorkbookShell.support.test.tsx) | Tests Timeline collection helpers preserve manual/auto-resolution state and nullable confidence. |
| [WorkbookShell.surfaces.test.tsx](WorkbookShell.surfaces.test.tsx) | Multi-surface workbook shell tests. |
| [WorkbookShell.timelineQuery.test.tsx](WorkbookShell.timelineQuery.test.tsx) | Tests Timeline query integration preserves owner row identity. |

Table paste and Timeline bulk actions use [retained batch ownership](runtime/README.md).
Controllers prepare source-specific intent; Workbook retains immutable requests,
complete outcomes, conflicts and read obligations through presentation changes.

`workbookRecoveryNavigation.test.tsx` covers coordinated authoring and authority
withdrawal. `WorkbookShell.surfaces.test.tsx` also exercises simultaneous retained
Note authoring, acknowledged read recovery, uncertain batch work and Indicator
review through production shell controls. Shared presentation is documented in
[shared](../shared/README.md); the shell stores no feature execution payloads.
