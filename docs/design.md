---
version: 0.1.0
name: Cartulary
document_class: design-direction
status: draft
accent_color: "#FACC15"
description: "A dense graphite workbook-native incident workspace with #FACC15 as a scarce warm accent for focus, primary action, and active-shell emphasis. The interface keeps the grid as the primary work surface, uses adjacent inspectors for enrichment and review, and presents conflicts, evidence, presence, and progressive structure as local semantic state rather than detached workflow chrome."

colors:
  accent: "#FACC15"
  accent-hover: "#FDE047"
  accent-active: "#EAB308"
  on-accent: "#111827"
  canvas: "#0B0C0F"
  surface-1: "#111318"
  surface-2: "#161A20"
  surface-3: "#1C222B"
  surface-4: "#252B35"
  hairline: "#2B313C"
  hairline-strong: "#3B4352"
  hairline-focus: "#FACC15"
  ink: "#F8FAFC"
  ink-muted: "#CBD5E1"
  ink-subtle: "#94A3B8"
  ink-tertiary: "#64748B"
  inverse-canvas: "#F8FAFC"
  inverse-ink: "#0F172A"
  semantic-info: "#60A5FA"
  semantic-success: "#34D399"
  semantic-caution: "#FB923C"
  semantic-conflict: "#F87171"
  semantic-destructive: "#F43F5E"
  semantic-presence-self: "#A78BFA"
  semantic-presence-other: "#22D3EE"
  overlay-scrim: "rgba(0, 0, 0, 0.56)"

typography:
  ui:
    fontFamily: "Inter, Geist, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: 0
  surface-title:
    fontFamily: "Inter, Geist, ui-sans-serif, system-ui, sans-serif"
    fontSize: 18px
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: -0.2px
  section-heading:
    fontFamily: "Inter, Geist, ui-sans-serif, system-ui, sans-serif"
    fontSize: 16px
    fontWeight: 600
    lineHeight: 1.30
    letterSpacing: -0.1px
  grid-cell:
    fontFamily: "Inter, Geist, ui-sans-serif, system-ui, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.35
    letterSpacing: 0
  metadata:
    fontFamily: "Inter, Geist, ui-sans-serif, system-ui, sans-serif"
    fontSize: 12px
    fontWeight: 500
    lineHeight: 1.30
    letterSpacing: 0
  button:
    fontFamily: "Inter, Geist, ui-sans-serif, system-ui, sans-serif"
    fontSize: 13px
    fontWeight: 600
    lineHeight: 1.20
    letterSpacing: 0
  mono:
    fontFamily: "JetBrains Mono, Geist Mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.40
    letterSpacing: 0

spacing:
  xxs: 2px
  xs: 4px
  sm: 8px
  md: 12px
  lg: 16px
  xl: 24px
  xxl: 32px
  shell-gap: 8px
  panel-padding: 12px
  inspector-padding: 16px

density:
  compact:
    rowHeight: 28px
    cellPadding: "3px 6px"
  default:
    rowHeight: 36px
    cellPadding: "4px 8px"
  comfortable:
    rowHeight: 44px
    cellPadding: "6px 10px"
  defaultMode: default

rounded:
  xs: 3px
  sm: 5px
  md: 7px
  lg: 10px
  xl: 14px
  pill: 9999px

border:
  hairline: "1px solid {colors.hairline}"
  strong: "1px solid {colors.hairline-strong}"
  focus: "2px solid {colors.hairline-focus}"

elevation:
  none: "none"
  panel: "0 1px 2px rgba(0, 0, 0, 0.32)"
  drawer: "-12px 0 32px rgba(0, 0, 0, 0.36)"
  popover: "0 16px 40px rgba(0, 0, 0, 0.44)"

motion:
  duration-fast: 120ms
  duration-normal: 180ms
  easing-standard: "cubic-bezier(0.2, 0, 0, 1)"

layout:
  baseViewport: "1280x720 CSS px"
  topBarHeight: 48px
  viewBarHeight: 40px
  statusStripHeight: 28px
  inspectorDefaultWidth: 420px
  inspectorMinWidth: 360px
  inspectorMaxWidth: "min(560px, 45vw)"

components:
  button-primary:
    backgroundColor: "{colors.accent}"
    textColor: "{colors.on-accent}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: "7px 12px"
  button-secondary:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.ink}"
    border: "{border.hairline}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: "7px 12px"
  button-danger:
    backgroundColor: "transparent"
    textColor: "{colors.semantic-destructive}"
    typography: "{typography.button}"
    rounded: "{rounded.md}"
    padding: "7px 12px"
  text-input:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    border: "{border.hairline}"
    rounded: "{rounded.sm}"
    padding: "6px 8px"
  chip:
    backgroundColor: "{colors.surface-2}"
    textColor: "{colors.ink-muted}"
    border: "{border.hairline}"
    rounded: "{rounded.pill}"
    padding: "2px 7px"
  grid-cell:
    backgroundColor: "transparent"
    textColor: "{colors.ink}"
    typography: "{typography.grid-cell}"
    padding: "{density.default.cellPadding}"
  inspector:
    backgroundColor: "{colors.surface-1}"
    textColor: "{colors.ink}"
    border: "{border.hairline}"
    rounded: "{rounded.lg}"
    padding: "{spacing.inspector-padding}"
---

# Cartulary Design Direction

## 1. Status, authority, and scope

This document is a design-direction artifact for the Cartulary browser application. It is subordinate to Core 00 through Core 04 for current implementation behavior, subordinate to Core 05 only for claim-bearing timed or fixture-sensitive publication behavior, subordinate to later adopted Cartulary NLSpecs, and aligned with `domain.md` for vocabulary discipline.

`design.md` does not define Base Profile or extension-profile implementation conformance. It must not create routes, schemas, authorization rules, evidence access semantics, lifecycle transitions, benchmark claims, storage boundaries, or extension-profile behavior. If this document conflicts with the normative core, the normative core governs and this document must be repaired.

This document is intentionally not an NLSpec. It defines visual language, layout treatment, interaction presentation, component styling, and implementation-facing design guidance; it does not fully determine software behavior in the way an adopted NLSpec would.

The reference `DESIGN.md` supplied with this project is used only as a structural model for a design artifact with tokens, visual principles, do/don't rules, responsive guidance, and known gaps. Its Linear-specific colors, marketing-page assumptions, typography, and product surface do not apply to Cartulary.

## 2. Product and design goals

Cartulary is a workbook-native incident workspace. Its design goal is to feel like a serious, low-friction workbook on the hot path while behaving like a disciplined case system underneath: source state is typed, relational, versioned, attributed, auditable, and projected through workbook surfaces.

The design must preserve spreadsheet speed where it matters: direct typing, paste, compact scanning, keyboard navigation, visible rows, flexible filtering, and fast reshaping of the current working set. It must reject spreadsheet failure modes where they would damage incident work: row-position identity, silent overwrites, hidden relationship semantics, evidence paths as authority, unmanaged binary storage, and unversioned history.

The design target is not a dashboard, ticket queue, CRM, SIEM, EDR, evidence vault, long-form report editor, or command-center visualization. It is the incident's shared operational workbook: the place where analysts capture facts, link evidence, resolve rough references, coordinate work, inspect history, and maintain a defensible common operating picture.

### 2.1 Product goals

| Goal                              | Design consequence                                                                                                                           |
| --------------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| Preserve low-friction capture     | Default creation and correction happen in the grid or directly adjacent inspector, not in detached forms.                                    |
| Support progressive structure     | Rough text, unresolved mentions, canonical chips, and dismissed mentions have distinct visual states.                                        |
| Keep context during live work     | Evidence preview, conflict resolution, history, and relationship actions keep the grid visible when the base viewport allows it.             |
| Make collaboration legible        | Save state, presence, pending work, and same-field conflicts are visible as local semantic states.                                           |
| Keep coordination workbook-native | Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson surfaces use workbook grammar rather than separate modules. |
| Protect authority boundaries      | UI labels, component names, SQL names, and vendor grid coordinates never become behavior authorities.                                        |

## 3. Target users and primary workflows

Cartulary's primary users are incident responders working under uncertainty, time pressure, and collaboration load. The design must serve front-line analysts, incident leads, reviewers, report writers, and coordination stakeholders without turning ordinary capture into ceremony.

| User                                | Primary needs                                                                                           | Design posture                                                                                              |
| ----------------------------------- | ------------------------------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| Front-line analyst                  | Capture observations quickly, paste data, attach screenshots/files, keep working with incomplete facts. | Grid-first, compact, same-surface feedback, minimal modal interruption.                                     |
| Incident lead or commander          | See scope, queues, blockers, ownership, unresolved items, and status-review material.                   | Persistent shell context, system-view switcher, saved views, readable counts, compact status.               |
| Reviewer or report writer           | Inspect lineage, attribution, evidence, history, rollback eligibility, and stable references.           | Inspector-first review surfaces, row-local history, clear destructive-action boundaries.                    |
| Threat hunter or detection engineer | Pivot through indicators, queries, scoped hosts, identities, and linked evidence.                       | Linkable chips, preserved raw values, same-shell pivots, dense sortable/filterable views.                   |
| Stakeholder coordinator             | Track communications, decisions, handoffs, report timing, and follow-up.                                | Workbook-native coordination surfaces with owner, state, and linkage clarity.                               |
| Deployment administrator            | Manage deployment-local users and recovery concerns outside incident data.                              | Not a design center for the workbook shell; deployment administration must not blur into incident surfaces. |

### 3.1 Primary workflows

| Workflow                                                           | Required design treatment                                                                                                                                             |
| ------------------------------------------------------------------ | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Open incident and orient                                           | Show incident identity, active surface, built-in tabs, `System views`, saved-view selector, presence, grid, inspector affordance, and save state in one shell.        |
| Create rough Timeline row                                          | Permit entry with incomplete time, uncertain prose, unresolved host/account strings, or an attachment-only signal. Do not require canonical selection before capture. |
| Paste or bulk-enter rows                                           | Preserve the user's place, show interpretation and validation locally, and use staged feedback when slower work continues.                                            |
| Resolve host, identity, or indicator references                    | Use inspector or same-surface enrichment; distinguish unresolved text, resolved chip, auto-resolved chip, and dismissed mention.                                      |
| Attach or request evidence                                         | Show pending, requested, received, available, failed, blocked, inconsistent, and quarantined states without implying fake attachment.                                 |
| Preview or download evidence                                       | Keep the workbook shell present; blocked preview remains inline and does not silently fall back to download.                                                          |
| Filter, sort, group, and save views                                | Treat the current view as a workbook working set; saved views remain configurations over one `view_schema_id`.                                                        |
| Resolve same-field conflict                                        | Mark only the affected cell, keep the saved value visible, retain the local draft separately, and open a resolver from that cell.                                     |
| Review history or rollback                                         | Use row-local history and scoped destructive-action presentation in the inspector; do not make routine editing feel like approval workflow.                           |
| Coordinate tasks, decisions, handoffs, status reviews, and lessons | Keep these as workbook-native surfaces, not separate task-management or workflow modules.                                                                             |

## 4. Visual design principles

### 4.1 Dense graphite forensic workspace

The default visual direction is a dark, dense, graphite workspace with warm operational emphasis. It should feel calm, precise, inspectable, and durable. It should not feel cyberpunk, theatrical, militarized, playful SaaS, or generic enterprise gray.

Use subtle surface elevation, crisp hairlines, stable row rhythm, and restrained contrast. The interface should imply that evidence, history, and collaboration state are always inspectable, without turning the workspace into a wall of alerts.

### 4.2 Accent scarcity

`#FACC15` is the Cartulary accent. Use it sparingly for focus rings, selected shell controls, primary affirmative action, active grid handles, and brand emphasis. Do not use it as the default warning color, row background, large panel fill, heatmap, decorative gradient, or arbitrary chart color.

Because `#FACC15` reads as yellow, caution/warning semantics must use a separate semantic role and must include text or shape. A primary button using `#FACC15` must use dark text from `on-accent` and must meet contrast requirements.

### 4.3 Local semantic feedback

State must appear where it matters. Conflicts appear at the affected cell. Evidence state appears in the row, evidence chip, or inspector. Presence appears in the header, row gutter, and cell. Save state appears in the status strip. Auto-resolution disclosure appears near the affected chips or batch result.

### 4.4 Structure without ceremony

The UI must make structure visible without demanding structure before capture. Users may type rough values first; later normalization, resolution, review, and rollback become available through adjacent controls and stateful chips.

### 4.5 Accessibility is part of the visual language

Color is never the only carrier of state. Every state-bearing color must be paired with shape, marker text, accessible name, tooltip, live-region announcement, or equivalent non-color cue. Focus must be visible in dense layouts.

## 5. Color system

### 5.1 Core palette

| Token             | Value     | Use                                                                                                |
| ----------------- | --------- | -------------------------------------------------------------------------------------------------- |
| `accent`          | `#FACC15` | Focus, selected shell control, primary affirmative action, active grid affordance, brand emphasis. |
| `accent-hover`    | `#FDE047` | Pointer hover state for accent-filled controls and active accent affordances.                      |
| `accent-active`   | `#EAB308` | Pointer press or active state for accent-filled controls and active accent affordances.            |
| `on-accent`       | `#111827` | Text/icon on accent-filled controls.                                                               |
| `canvas`          | `#0B0C0F` | Application background and deepest grid surround.                                                  |
| `surface-1`       | `#111318` | Top bar, status strip, base panels.                                                                |
| `surface-2`       | `#161A20` | View bar, cells on hover, menus, secondary controls.                                               |
| `surface-3`       | `#1C222B` | Raised panels, selected non-accent surfaces, inspector section blocks.                             |
| `surface-4`       | `#252B35` | Stronger lifted controls and active menus.                                                         |
| `hairline`        | `#2B313C` | Grid lines, panel separators, control borders.                                                     |
| `hairline-strong` | `#3B4352` | Active row outline, pinned boundaries, inspector edge.                                             |
| `ink`             | `#F8FAFC` | Primary text.                                                                                      |
| `ink-muted`       | `#CBD5E1` | Secondary text.                                                                                    |
| `ink-subtle`      | `#94A3B8` | Metadata, placeholders, secondary counts.                                                          |
| `ink-tertiary`    | `#64748B` | Disabled or low-emphasis text when contrast remains acceptable.                                    |

### 5.2 Semantic colors

| Role           | Token                               | Use                                                            | Non-color requirement                           |
| -------------- | ----------------------------------- | -------------------------------------------------------------- | ----------------------------------------------- |
| Info           | `semantic-info` `#60A5FA`           | Informational state, source metadata, neutral system messages. | Label or info icon.                             |
| Success        | `semantic-success` `#34D399`        | Saved confirmation, successful attach, completed task.         | Check marker or state text.                     |
| Caution        | `semantic-caution` `#FB923C`        | Warning, stale, pending verification, unsupported preview.     | Warning marker and state text.                  |
| Conflict       | `semantic-conflict` `#F87171`       | Same-field conflict, replay blocked, invalid cell.             | Cell-local marker and accessible name.          |
| Destructive    | `semantic-destructive` `#F43F5E`    | Delete, rollback, supersede, destructive confirmation.         | Explicit label and confirmation text.           |
| Presence self  | `semantic-presence-self` `#A78BFA`  | Current user's presence marker when needed.                    | Avatar/initial or self label.                   |
| Presence other | `semantic-presence-other` `#22D3EE` | Other analyst presence.                                        | Avatar/initial, row gutter, or same-cell label. |

Semantic colors must not be substituted for the accent. The accent communicates Cartulary identity, active control, and focus; semantic colors communicate operational state.

### 5.3 Color usage rules

| Do                                                                                         | Do not                                                                   |
| ------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| Use accent for one primary action per local action group.                                  | Use accent for every clickable control.                                  |
| Use accent focus rings around keyboard focus when the ring is visible against the surface. | Hide focus behind color shifts only.                                     |
| Use neutral surfaces and hairlines for density and hierarchy.                              | Use heavy shadows, gradients, or glowing alert panels.                   |
| Use semantic colors only with text, icon, shape, or accessible name.                       | Encode unresolved, resolved, conflict, or evidence state by color alone. |
| Keep warning/caution visually distinct from accent.                                        | Treat `#FACC15` as the warning palette.                                  |

## 6. Typography and iconography

### 6.1 Typography

Use a modern, compact UI sans stack: `Inter, Geist, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif`. Use a mono stack only for identifiers, hashes, timestamps where alignment matters, route or contract tokens, code-like snippets, and exact technical values.

| Role            | Token             | Default                                   |
| --------------- | ----------------- | ----------------------------------------- |
| Surface title   | `surface-title`   | `18px`, weight `600`, line-height `1.25`. |
| Section heading | `section-heading` | `16px`, weight `600`, line-height `1.30`. |
| Grid cell       | `grid-cell`       | `14px`, weight `400`, line-height `1.35`. |
| Body/UI         | `ui`              | `14px`, weight `400`, line-height `1.45`. |
| Metadata        | `metadata`        | `12px`, weight `500`, line-height `1.30`. |
| Button          | `button`          | `13px`, weight `600`, line-height `1.20`. |
| Mono            | `mono`            | `12px`, weight `400`, line-height `1.40`. |

Long prose belongs in the inspector, Notes, findings surfaces, or companion findings document. Timeline and coordination grid cells should be compact and scannable.

### 6.2 Iconography

Use exactly one icon family in the workbook shell. The default implementation posture is an outline icon family with `16px` icons for dense inline affordances, `20px` icons for toolbar controls, and consistent stroke weight.

Icon-only controls must have accessible names and visible focus indicators. Icons may support but must not replace labels for destructive actions, conflict resolution, evidence download, evidence preview, rollback, merge initiation, party link/unlink, inspector close, inspector pin, and system-view switching.

## 7. Layout, spacing, density, and responsive posture

### 7.1 Shell regions

The application shell is one continuous workspace with five persistent regions at the base viewport.

| Region       | Contents                                                                                       | Design rule                               |
| ------------ | ---------------------------------------------------------------------------------------------- | ----------------------------------------- |
| Top bar      | Incident identity, built-in tabs, `System views`, current system-view title, presence summary. | Persistent chrome, not dashboard summary. |
| View bar     | Active view selector, saved-view selector, sort, group, filters, active chips.                 | Belongs to the active surface only.       |
| Grid         | Active workbook surface with `record_id`-bound rows and `field_key`-bound cells.               | Primary work surface.                     |
| Inspector    | Details, Relationships, Evidence, History.                                                     | Adjacent secondary surface.               |
| Status strip | Save state, secondary message, presence summary or overflow.                                   | Capacity-limited working-state surface.   |

### 7.2 Surface composition

At the base viewport, built-in tabs are always-visible primary tabs in this order: Timeline, Hosts, Identities, Evidence, Notes. Required system views are available through an adjacent switcher with accessible name `System views`. Saved views appear under the active surface's view selector, not as primary tabs.

| Surface family                          | Shell exposure                              | Ordering                                               |
| --------------------------------------- | ------------------------------------------- | ------------------------------------------------------ |
| Built-in tabs                           | Always-visible primary tabs.                | Timeline, Hosts, Identities, Evidence, Notes.          |
| Scope and assessment system views       | `System views` switcher.                    | Indicators, Compromise Assessments, Parties.           |
| Coordination system views               | `System views` switcher.                    | Task Requests, Decisions, Communications Log, Handoff. |
| Review and learning system views        | `System views` switcher.                    | Status Review, Lesson.                                 |
| Standardized optional artifact surfaces | Same switcher when implemented and exposed. | Findings, Investigative Queries, Forensic Keywords.    |
| Saved views                             | Active surface view selector.               | Scope group, then display name.                        |

For documentation, use `Compromise Assessments` as the canonical surface label for `cartulary.view.assessments.v1`. Constrained UI labels may use `Assessments` as a display shorthand only; no semantic distinction is intended.

Do not expose all fourteen required surfaces as primary tabs. Do not make command-palette-only access the only way to reach required system views. Do not implement coordination surfaces as separate application modules.

### 7.3 Spacing and density

| Density     | Row height | Cell padding | Use                                                      |
| ----------- | ---------- | ------------ | -------------------------------------------------------- |
| Compact     | `28px`     | `3px 6px`    | High-density triage, narrow screens, analyst preference. |
| Default     | `36px`     | `4px 8px`    | Default for all built-in tabs and required system views. |
| Comfortable | `44px`     | `6px 10px`   | Review, training, demos, accessibility preference.       |

All workbook surfaces share the same active density class. Do not create per-surface density models unless a later design revision explicitly scopes the exception. Large incident grids use fixed-height rows by default.

### 7.4 Inspector

| Property         | Default and bounds                                                         |
| ---------------- | -------------------------------------------------------------------------- |
| Default width    | `420px`.                                                                   |
| Minimum width    | `360px`.                                                                   |
| Maximum width    | `min(560px, 45vw)`.                                                        |
| Resize           | User may resize within min/max bounds.                                     |
| Pinning          | Client-local only; not persisted in saved views.                           |
| Section order    | Details, Relationships, Evidence, History.                                 |
| Grid visibility  | At base viewport, grid remains visible whenever inspector is open.         |
| Close affordance | Visible control with accessible name `Close inspector`.                    |
| Pin affordance   | Visible control with accessible name `Pin inspector` or `Unpin inspector`. |

### 7.5 Responsive posture

The design targets desktop browser operation with keyboard and pointer input at a base viewport of at least `1280x720` CSS pixels and `100%` zoom. At that viewport, active surface identity, primary tabs, system-view switcher, grid, save-state label, and status-strip primary state must remain visible.

Below `1280px` width, preserve the grid as the primary work surface. The inspector may collapse into a bottom sheet, modal drawer, or full-height overlay. Primary tabs may condense, but active surface identity, save state, and the system-view switcher must remain reachable by keyboard and pointer.

This document does not claim a mobile-first or touch-first product direction. Touch support may exist, but omission of touch-specific gestures is acceptable when keyboard and pointer paths remain complete.

## 8. Grid interaction guidance

### 8.1 Grid identity and addressing

Rows are identified by stable `record_id`; writable cells are identified by stable `field_key`; optimistic writes use row versioning. The UI may display human labels, but behavior must not depend on visible tab labels, row numbers, column labels, projection table names, storage table names, or React component names.

### 8.2 Cell and row states

| State                       | Visual treatment                                                           | Required non-color cue                                   |
| --------------------------- | -------------------------------------------------------------------------- | -------------------------------------------------------- |
| Active cell                 | Accent focus ring or inset outline.                                        | Visible focus outline.                                   |
| Selected row                | Subtle surface lift or left gutter marker.                                 | Row selection marker and accessible selection state.     |
| Editable cell               | Normal cell with hover affordance; edit mode has stronger focus treatment. | Cursor/focus mode, accessible role.                      |
| Read-only or derived cell   | Muted text or lock/read-only marker on focus.                              | Accessible read-only state.                              |
| Dirty or pending cell       | Small pending dot or local queue marker.                                   | `Syncing` status and accessible pending label.           |
| Saved cell                  | No persistent success decoration by default.                               | Status strip `Saved`.                                    |
| Invalid cell                | Conflict/error semantic border or underline.                               | Error marker and field-local message.                    |
| Conflicted cell             | Cell-local conflict marker; saved server value remains displayed.          | Marker with accessible name `Conflict on <field label>`. |
| Invalidate/refresh required | Subtle stale marker until re-query.                                        | `Syncing` or inline stale label.                         |

### 8.3 Keyboard and pointer behavior

Keyboard navigation is a first-class interaction path. Toolbar, tab, switcher, menu, grid, inspector, preview, resolver, and dialog controls must all be reachable without a pointer. After drawer, dialog, popover, or resolver close, focus returns to the invoking cell, row, or control when it still exists.

Routine editing must not require modal dialogs. Dialogs are reserved for destructive-action confirmation, merge initiation confirmation, rollback confirmation, same-field conflict resolution only when a drawer or same-surface panel cannot preserve context, and release-scope changes when the Snapshot and Reporting Extension Profile is implemented.

### 8.4 Sorting, filtering, grouping, and saved views

Sort, filter, group, and saved-view controls live in the View bar. They apply to the active surface only. Active filters render as compact chips. Grouping must remain visibly subordinate to the workbook surface and must not become a dashboard or report layer.

Saved views must feel like named configurations of a surface, not new canonical surfaces. A `system` saved view is not the same object as a required system view.

## 9. Progressive structure, chips, and semantic states

### 9.1 Chip states

Use the closed visual vocabulary below for entity-mention and relationship chips.

| State           | Border                      | Required marker                      | Accessible name pattern                                              | Notes                                                                     |
| --------------- | --------------------------- | ------------------------------------ | -------------------------------------------------------------------- | ------------------------------------------------------------------------- |
| `unresolved`    | Dashed or dotted.           | Leading `?` or visible `Unresolved`. | `Unresolved <entity type> mention: <raw text>`.                      | Must differ from ordinary text and resolved chips.                        |
| `resolved`      | Solid.                      | No unresolved marker.                | `Resolved <entity type>: <display name>`.                            | Shows canonical target while retaining inspection path to raw mention.    |
| `auto_resolved` | Solid plus auto marker.     | Visible `auto`.                      | `Auto-resolved <entity type>: <display name>; matched <alias text>`. | Remains inspectably marked after transient disclosure fades.              |
| `dismissed`     | Low-emphasis chip or token. | Visible `dismissed`.                 | `Dismissed mention: <raw text>`.                                     | Display only where inspectable; excluded from active relationship values. |

Color alone must not distinguish chip states.

### 9.2 Relationship and collection cells

Relationship cells that mix unresolved mention tokens and canonical chips must preserve those as different object types. Do not coerce unresolved tokens and canonical chips into comma-delimited strings for display, editing, conflict resolution, or copy/paste presentation.

### 9.3 Auto-resolution disclosure

Auto-resolution is helpful only when inspectable. If a value is auto-resolved, the user must be able to tell it was auto-resolved, inspect why, and correct it through the same surface or inspector. Transient disclosure may fade, but the chip state must remain marked.

## 10. Collaboration, save state, and conflict design

### 10.1 Save state

The status strip uses exactly one primary save-state label: `Syncing`, `Saved`, or `Conflict`. Presence does not change this label mapping.

| Label      | Meaning                                                                                                                                        | Visual treatment                                      |
| ---------- | ---------------------------------------------------------------------------------------------------------------------------------------------- | ----------------------------------------------------- |
| `Syncing`  | One or more workbook mutations are in flight, the local pending queue is non-empty, or replay is paused pending recovery or re-authentication. | Subtle spinner or pending marker plus label.          |
| `Saved`    | No workbook mutation is in flight, the local pending queue is empty, and no unresolved same-field local draft exists.                          | Text label only by default; no celebratory animation. |
| `Conflict` | Unresolved same-field conflict, refused queue overflow, or non-retryable replay failure requiring analyst action.                              | Conflict semantic marker and local entry point.       |

### 10.2 Presence

Show presence at three levels: header avatars for users on the same surface, row-gutter indicators for users focused on a row, and same-cell indicators when another analyst is actively editing the same field and that signal is available.

| Level  | Placement                            | Visible bounds                             |
| ------ | ------------------------------------ | ------------------------------------------ |
| Header | Top bar presence cluster.            | Up to five avatars or initials, then `+N`. |
| Row    | Row gutter.                          | Up to three indicators, then `+N`.         |
| Cell   | Inside or adjacent to affected cell. | One compact indicator or label.            |

Presence must not lock editing, change save state, or imply ownership.

### 10.3 Same-field conflict

A same-field conflict marks only the affected cell. The conflicted cell displays the current saved server value plus a visible conflict marker; the analyst's unsaved local value is retained separately and must not be rendered as saved.

The resolver opens from the conflicted cell, keeps the grid visible when possible, and includes row context, field label, stable `field_key`, saved value with actor and timestamp, local unsaved value, optional merge summary, and direct resolution actions. Initial focus lands on a non-destructive summary or safe control, not on `Use my unsaved value`.

## 11. Evidence design

### 11.1 Evidence state vocabulary

Evidence is not a file path and not a pasted cell attachment. Use the evidence record as the user-facing envelope and object blob state as binary upload metadata. The UI must not imply that a pending blob is attached evidence.

| State                 | Treatment                                                                | Required cue                                                 |
| --------------------- | ------------------------------------------------------------------------ | ------------------------------------------------------------ |
| `requested`           | Evidence row exists without blob.                                        | `Requested` label and optional requested-at metadata.        |
| `pending_receipt`     | Evidence expected but not received.                                      | Pending cue, no preview action.                              |
| `received`            | Evidence metadata received, availability not complete.                   | Received cue; show Attach blob or Awaiting upload affordance as the next expected action depending on whether an upload slot has been issued. |
| `pending` upload slot | Upload slot exists but is not attached.                                  | Local pending marker, no evidence count increment.           |
| `available`           | Evidence available for preview or download according to handle contract. | Evidence count, preview/download affordance as allowed.      |
| `failed`              | Upload or finalization failed.                                           | Failure marker and retry/fresh-slot affordance when allowed. |
| `quarantined`         | Blob/evidence state requires quarantine.                                 | Exact `quarantined` label, no softened paraphrase.           |
| `blocked` preview     | Evidence exists but preview unsupported or blocked.                      | Inline blocked state; no silent download fallback.           |
| `inconsistent`        | Blob/evidence state mismatch or contract mismatch.                       | Inline inconsistent state; preview/download blocked.         |

### 11.2 Evidence preview and download

Preview should open in a side or bottom preview without forcing full-page navigation away from the grid. Preview and download affordances invoke the evidence-access handle contract; returned `href` values are opaque same-origin URLs. The browser must not synthesize object-store URLs, parse handle tokens, or treat object-store locations as evidence identity.

### 11.3 Evidence density

Evidence counts belong in grid cells or chips. Detailed custody, collection source, blob metadata, and preview/download affordances belong in the inspector or Evidence surface. Dense grid rows should show enough state for triage without becoming evidence-detail cards.

## 12. Component and screen-level patterns

### 12.1 Buttons

| Component          | Use                                                        | Treatment                                               |
| ------------------ | ---------------------------------------------------------- | ------------------------------------------------------- |
| Primary button     | One affirmative action in a local action group.            | Accent fill, `on-accent` text, medium radius.           |
| Secondary button   | Ordinary non-primary action.                               | Neutral surface, hairline border.                       |
| Tertiary button    | Low-emphasis toolbar action.                               | Transparent or surface hover only.                      |
| Destructive button | Delete, rollback, supersede, publish-danger, merge-danger. | Destructive semantic text or border; never accent fill. |
| Icon button        | Dense toolbar, chip action, inspector control.             | One icon family, accessible name, visible focus.        |

### 12.2 Inputs and editors

Inputs use neutral surfaces, compact padding, visible focus, and no decorative glow. Validation appears at the field or cell. Placeholder text must not be the only label for controls outside grid edit mode.

### 12.3 Menus and popovers

Menus and popovers use `surface-3`, `hairline`, compact vertical rhythm, keyboard navigation, and focus return to the invoking control. The `System views` menu follows the required grouping order in §7.2.

### 12.4 Inspector sections

| Section       | Purpose                                                            | Default treatment                                    |
| ------------- | ------------------------------------------------------------------ | ---------------------------------------------------- |
| Details       | Main editable or readable fields not comfortable in the grid.      | Stacked compact forms with clear labels.             |
| Relationships | Mentions, resolved entities, indicators, links, parties.           | Chip lists with source state and action affordances. |
| Evidence      | Evidence records, counts, preview/download state, custody summary. | Compact evidence cards or rows.                      |
| History       | Change summaries, actors, timestamps, rollback eligibility.        | Row-local timeline with scoped actions.              |

### 12.5 Empty states

Empty states stay in the active shell. They name the surface, explain the minimum useful create action, and provide an in-place create affordance when creation is permitted. They do not navigate to a setup wizard for ordinary capture.

### 12.6 Banners, toasts, and inline messages

| Pattern        | Purpose                                                                                                                              | Duration                                                        |
| -------------- | ------------------------------------------------------------------------------------------------------------------------------------ | --------------------------------------------------------------- |
| Banner         | Persistent same-surface state such as session expiry, queue overflow, collaboration degradation, or active-surface pack degradation. | Until state clears.                                             |
| Toast          | Transient confirmation such as completed long queue or batch disclosure.                                                             | Auto-dismiss after `5 seconds`; pause while hovered or focused. |
| Inline message | Cell-local, row-local, or action-local issue.                                                                                        | Until corrected or replaced.                                    |
| Dialog         | Confirmation where confirmation is the point.                                                                                        | Until explicit action or safe cancel.                           |

### 12.7 Screen-level patterns

| Screen or surface    | Design pattern                                                                                    |
| -------------------- | ------------------------------------------------------------------------------------------------- |
| Incident shell       | Dense top bar, view bar, virtualized grid, inspector, status strip.                               |
| Timeline             | Fastest capture surface; text-first, unresolved references, evidence count, capture/review state. |
| Hosts and Identities | Canonical or stub entities; sortable/filterable scope fields; relationship pivots.                |
| Evidence             | Evidence lifecycle, collector/source text and party references, upload/preview/download state.    |
| Notes                | Artifact-backed text surface; not a replacement for task, decision, handoff, or evidence state.   |
| Indicators           | Canonical indicators plus pivots to source observations and lifecycle history.                    |
| Compromise Assessments | Append-style assessment history, confidence band, rationale, support links.                     |
| Parties              | Incident-scoped coordination identities; linkable from coordination surfaces, task requests, decisions, and communications log entries. |
| Task Requests        | Queue-oriented work with owner, status, priority, due, blockers, and linked records.              |
| Decisions            | Rationale-bearing coordination choices with owner, status, review class, and support references.  |
| Communications Log   | Durable communication memory with audience, channel, summary, decisions, and action follow-up.    |
| Handoff              | Current state, open work, open decisions, risks, and next checks.                                 |
| Status Review        | Blocked work, pending evidence, open decisions, risk summary, next report timing.                 |
| Lesson               | Retrospective/improvement record linked to follow-up work and evidence.                           |

## 13. Accessibility expectations

The design floor is WCAG 2.2 AA for state-bearing text and controls. The UI must support keyboard operation for hot-path grid work, the system-view switcher, inspector actions, evidence preview/download, conflict resolution, menus, dialogs, and saved-view controls.

| Area          | Expectation                                                                                                                                                                                                         |
| ------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Focus         | Visible focus ring, default accent focus where contrast passes.                                                                                                                                                     |
| Keyboard      | No pointer-only hot-path operation.                                                                                                                                                                                 |
| Screen reader | Save-state changes, conflict state, queue overflow, replay blocked, chip state, auto-resolution, presence, and inspector open/close are exposed through roles, accessible names, or live regions where appropriate. |
| Color         | No color-only state. Every state has text, shape, marker, icon, accessible name, or equivalent.                                                                                                                     |
| Row identity  | Accessibility-tree identity binds to stable `record_id`; spoken labels should be human-readable and not only raw IDs.                                                                                               |
| Motion        | Motion is short, functional, and non-essential. Reduced-motion-specific design is deferred, but motion-only state is not allowed.                                                                                   |
| Touch         | No mobile/touch conformance claim. Touch may be additive only.                                                                                                                                                      |

## 14. Design, domain, implementation, and mockup boundaries

### 14.1 What this document owns

`design.md` owns the visual design direction, token vocabulary, color use, typography direction, density values, component visual patterns, screen-level design patterns, accessibility expectations, and coding-agent design guardrails for the browser application.

### 14.2 What this document does not own

| Boundary                                                                                                                 | Owner                                                                                       |
| ------------------------------------------------------------------------------------------------------------------------ | ------------------------------------------------------------------------------------------- |
| Product behavior, conformance, profile scope, and authority order                                                        | Core 00 through Core 04.                                                                    |
| Claim-bearing timed or fixture-sensitive publication                                                                     | Core 05.                                                                                    |
| Domain vocabulary and forbidden substitutions                                                                            | `domain.md`, with owner sections governing behavior.                                        |
| Public routes, wire shapes, view schemas, field keys, projection contracts, job contracts, storage boundaries            | Core 01.                                                                                    |
| Record model, history, mentions, indicators, evidence, parties, task requests, decisions, artifacts, closed vocabularies | Core 02.                                                                                    |
| Workbook interactions, collaboration, save state, presence, conflict resolution, evidence workflow, workflows            | Core 03.                                                                                    |
| Authentication, authorization, trust boundaries, deployment, evidence security, export release boundaries                | Core 04.                                                                                    |
| Repo-local package boundaries, grid adapter, generated contracts, developer workflow                                     | Development, bootstrap, and testing guides as subordinate implementation-support artifacts. |
| Future screenshots, Figma files, or mockups                                                                              | Illustrative examples only unless later adopted as governing design artifacts.              |

### 14.3 Future mockups and screenshots

Future mockups may show examples of this design language, but they must not create behavior by implication. A mockup that conflicts with core behavior, domain vocabulary, or this design direction must be labeled illustrative or revised before implementation.

## 15. Implementation guidance for developers and coding agents

1. Use the tokens in this file as the first source for color, spacing, density, radius, and elevation decisions.
2. Use canonical product terms from `domain.md` and stable identifiers from owner contracts when naming surfaces, state, tests, CSS variables, and component props.
3. Style `react-data-grid` through `/packages/grid-adapter`, stable Cartulary wrapper classes, CSS variables, accessible names, and state attributes. Do not depend on generated vendor class-name internals.
4. Keep grid work on the grid. Use the inspector for enrichment, relationships, evidence, history, and scoped destructive actions.
5. Do not create new tabs, route families, record types, lifecycle states, semantic statuses, or approval workflows from design need alone.
6. Do not infer behavior from visible labels, row order, column labels, SQL names, projection names, component names, or style classes.
7. When a design choice would alter product behavior, mark it `TODO: owner unresolved` and route it to the correct owner section instead of implementing a local guess.
8. Use visual regression fixtures for representative states: rough Timeline row, unresolved mention, resolved chip, auto-resolved chip, dismissed mention, pending evidence, blocked preview, same-field conflict, presence at header/row/cell, queue overflow, and inspector history.

## 16. Do and don't rules

### Do

- Do make the grid the protagonist of the workspace.
- Do keep `#FACC15` scarce, intentional, and accessible.
- Do encode state with text or shape in addition to color.
- Do keep conflict, evidence, and resolution feedback local to the affected row, cell, chip, or inspector section.
- Do use compact hairlines and neutral surfaces to preserve density.
- Do keep system views reachable without leaving the workbook shell.
- Do show the user's current save/conflict state in the status strip.
- Do keep raw capture inspectable after resolution or normalization.

### Don't

- Don't use yellow accent as the warning system.
- Don't turn routine row creation into a modal, wizard, approval, challenge ritual, or full-page flow.
- Don't make all fourteen required surfaces primary tabs.
- Don't hide required system views behind command-palette-only discovery.
- Don't create dashboard modules for coordination surfaces.
- Don't use color-only chips, color-only conflict markers, or color-only evidence state.
- Don't use raw object-store URLs, SQL table names, projection names, or React component names as user-facing design concepts.
- Don't add neon threat maps, cinematic gradients, heavy glow, or decorative risk heatmaps to the live workbook shell.

## 17. Known gaps and TODOs

| Item                            | Status                                     | Required next input                                                                                             |
| ------------------------------- | ------------------------------------------ | --------------------------------------------------------------------------------------------------------------- |
| Exact icon family               | `TODO:` not selected.                      | Pick one outline icon family compatible with the frontend license policy.                                       |
| Light theme                     | Out of scope for this draft.               | Decide whether dual theme is required before implementing light-mode tokens.                                    |
| High-contrast theme             | Deferred.                                  | Define a dedicated high-contrast theme only if adopted by a later design revision or accessibility requirement. |
| Current rendered UI screenshots | Not available in this uploaded source set. | Add screenshots or visual fixtures once the application shell exists.                                           |
| External visual reference board | Not source-of-truth.                       | Collect screenshots only as inspiration; do not treat them as behavior authority.                               |
| Report/export visual design     | Out of scope for this draft.               | Create a separate extension-profile design artifact if Snapshot and Reporting UI is implemented.                |
| Mobile/touch-specific design    | Out of scope for this draft.               | Define later only if a mobile/touch profile is adopted.                                                         |

## 18. Acceptance criteria

A reviewer may treat this `design.md` as ready to guide implementation only when every criterion below passes.

| ID       | Criterion                                                                                                                                                                        |
| -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| D-AC-001 | The document states that it is subordinate to Core 00 through Core 04 and does not define implementation conformance.                                                            |
| D-AC-002 | The document states that Core 05 is publication-only for claim-bearing timed or fixture-sensitive criteria.                                                                      |
| D-AC-003 | The document defines `#FACC15` as the accent and restricts its usage so it is not conflated with warning/caution.                                                                |
| D-AC-004 | The document includes color, typography, spacing, density, radius, elevation, layout, and component tokens.                                                                      |
| D-AC-005 | The document defines the target visual direction as dense, serious, workbook-native, graphite, and restrained.                                                                   |
| D-AC-006 | The document describes target users and primary workflows without inventing new runtime behavior.                                                                                |
| D-AC-007 | The document preserves the built-in tab model and required system-view switcher model.                                                                                           |
| D-AC-008 | The document defines visual treatment for grid cell state, chip state, save state, presence, conflict, evidence, inspector, and status strip.                                    |
| D-AC-009 | The document defines accessibility expectations for keyboard reachability, visible focus, non-color-only state, accessible names, live-state exposure, and WCAG 2.2 AA contrast. |
| D-AC-010 | The document defines component-level patterns for buttons, inputs, chips, menus, inspector sections, empty states, banners, toasts, inline messages, and dialogs.                |
| D-AC-011 | The document defines screen-level patterns for all five built-in tabs and required base-profile system views.                                                                    |
| D-AC-012 | The document forbids behavior inference from labels, row order, SQL names, projection names, vendor grid coordinates, component names, and styling classes.                      |
| D-AC-013 | The document states the boundary between design guidance, domain vocabulary, implementation specs, implementation-support guides, and future mockups.                            |
| D-AC-014 | The document marks unresolved design inputs as `TODO:` or out of scope instead of presenting guesses as settled facts.                                                           |
| D-AC-015 | The document is patch-ready Markdown and can be committed as `design.md` without requiring screenshots, external references, or proprietary font assets.                         |
