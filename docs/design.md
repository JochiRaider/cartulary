---
version: 0.2.0
name: Cartulary
document_class: design-direction-contract
status: draft/closed-design-contract
design_contract_schema_id: cartulary.design_direction.v1
token_registry_schema_id: cartulary.design_tokens.v1
default_theme_id: dark_graphite
description: "A dense graphite workbook-native incident workspace with {colors.accent} as the scarce warm accent for focus, primary action, and active-shell emphasis. The interface keeps the grid as the primary work surface, uses adjacent inspectors for enrichment and review, and presents conflicts, evidence, presence, and progressive structure as local semantic state rather than detached workflow chrome."

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
  hairline-focus: "{colors.accent}"
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
  default-mode: default

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
  duration-reduced-max: 80ms
  easing-standard: "cubic-bezier(0.2, 0, 0, 1)"

layout:
  baseViewport: "1280x720 CSS px"
  zoomDefault: "100%"
  baseMinWidth: 1280px
  baseMinHeight: 720px
  narrowMinWidth: 1024px
  compactMinWidth: 768px
  compactMinHeight: 640px
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

Design contract. This document is a design-direction contract for the Cartulary browser application.

Design contract. This document governs only observable UI design behavior in the following families: visual language, token interpretation, theme scope, shell composition, responsive presentation, component visual states, local semantic UI states, accessibility presentation, visual fixture expectations, and coding-agent design guardrails.

Design contract. This document MUST NOT define Base Profile or extension-profile implementation conformance. It MUST NOT create routes, schemas, authorization rules, evidence access semantics, lifecycle transitions, benchmark claims, storage boundaries, extension-profile behavior, record types, field registries, or public wire shapes.

Design contract. Core 00 through Core 04 govern current-profile implementation behavior. Core 05 governs claim-bearing timed or fixture-sensitive publication only. If this document and an owner section differ on product behavior, the owner section governs and this document MUST be repaired.

Design contract. This document is NLSpec-grade only inside its design-direction scope. It fully specifies the design behavior it owns, but it is not the Base Profile implementation-conformance corpus.

Design contract. `domain.md` governs repository vocabulary interpretation. This document uses domain terms for UI design only and MUST NOT redefine `party`, `artifact`, `view schema`, `saved view`, `system view`, `entity mention`, `object blob`, `workbook surface`, or any other domain concept.

## 2. Normative language and statement classes

Design contract. The key words **MUST**, **MUST NOT**, and **MAY** are normative inside this design-direction contract. **MUST** and **MUST NOT** define design-direction requirements. **MAY** defines optional design behavior only when omission behavior is specified. `default` defines the required behavior when no explicit caller, user, deployment, or owner-controlled override applies.

Design contract. This revision uses no advisory normative terms. A later revision is valid only if it defines the effect of any newly introduced advisory normative term in this section.

Design contract. Accepted normative sections MUST NOT contain unresolved placeholder text. Future work MUST be classified as a non-goal, future profile candidate, or external owner dependency with explicit omission behavior.

| Statement class | Meaning | Normative-language rule |
| --- | --- | --- |
| `Design contract.` | Binding design-direction requirement owned by this document. | Uses **MUST**, **MUST NOT**, **MAY**, and `default` only as defined in §2. |
| `Core restatement.` | Non-authoritative summary of Core-owned behavior. | MUST cite or name the owner and MUST NOT issue independent design-owned product behavior. |
| `Rationale.` | Explanation only. | MUST NOT contain normative keywords. |
| `Non-goal.` | Explicit omission boundary. | MUST state observable omission behavior. |
| `External dependency.` | Decision owned outside this document. | MUST name the owner or mark the revision blocked. |

Design contract. A paragraph without one of the statement-class prefixes above is non-normative only when it appears in a heading, code fence, table caption, acceptance criterion, or source note. Normative prose in all other locations MUST use one statement-class prefix.

## 3. Token registry, token resolution, and theme contract

### 3.1 Canonical token registry

Design contract. The YAML front matter is the single canonical token registry for this document. Body text MUST reference token IDs rather than restating canonical literal token values, except when the body defines a grammar, an example explicitly marked as non-normative, or an acceptance criterion about token validation.

Design contract. The token registry schema ID is `cartulary.design_tokens.v1`. A conforming token registry MUST contain only the top-level namespaces in the table below.

| Namespace | Required value shape | Notes |
| --- | --- | --- |
| `colors` | Hex color, `rgba()` string, or token reference. | Utility tokens such as `inverse-canvas` and `inverse-ink` are not alternate themes. |
| `typography` | Object with `fontFamily`, `fontSize`, `fontWeight`, `lineHeight`, and `letterSpacing`. | `lineHeight` can be unitless. `letterSpacing` can be `0` or a CSS length. |
| `spacing` | CSS pixel length. | Negative spacing is invalid. |
| `density` | Object with `rowHeight` and `cellPadding`, plus `default-mode`. | Valid modes are `compact`, `default`, and `comfortable`. |
| `rounded` | CSS pixel length or `9999px` for pill radius. | Negative radius is invalid. |
| `border` | CSS border shorthand with token references allowed. | Border color MUST resolve to a `colors` token. |
| `elevation` | `none` or CSS box-shadow string. | Elevation MUST NOT encode semantic state. |
| `motion` | CSS duration or cubic-bezier string. | Durations MUST be non-negative. |
| `layout` | CSS pixel length, CSS expression, percent string, viewport descriptor, or breakpoint descriptor where declared. | Breakpoint values are design-owned presentation bounds. |
| `components` | Object composed from primitive tokens and literal component values. | Component tokens MUST NOT redefine primitive token semantics. |

### 3.2 Token ID grammar

Design contract. Token references MUST use this grammar:

```text
token_ref = "{" namespace "." token_key "}"
namespace = one of the closed top-level namespaces in §3.1
token_key = declared key path inside that namespace, preserving the exact registry spelling and using "." between nested segments
```

Design contract. The canonical token reference for a top-level primitive token is `{namespace.token-key}`. The canonical token reference for a nested token is `{namespace.parent-key.child-key}`.

### 3.3 Token resolution algorithm

Design contract. Token resolution MUST be deterministic and MUST use this algorithm:

```pseudocode
resolve_token(value, stack):
  if value is an exact token_ref:
    target = lookup(value.namespace, value.token_key)
    if target does not exist: return invalid_token_reference
    if value is already in stack: return token_reference_cycle
    return resolve_token(target, stack + value)

  if value is a string containing one or more embedded token_ref substrings:
    for each token_ref in lexical order of appearance:
      replacement = resolve_token(token_ref, stack)
      if replacement is an error: return replacement
      replace the token_ref substring with replacement
    return the replaced string

  return value
```

Design contract. Token validation MUST fail before design-conformance evidence is accepted when any invalid state in the table below occurs.

| Invalid state | Required result |
| --- | --- |
| Unknown top-level namespace | Token validation fails. |
| Unknown token key | Token validation fails. |
| Token-reference cycle | Token validation fails. |
| Unsupported unit for the namespace value shape | Token validation fails. |
| Duplicate token after exact registry-path normalization | Token validation fails. |
| Component token references undeclared primitive token | Token validation fails. |
| `density.default-mode` does not equal one declared density mode | Token validation fails. |
| A body section restates a canonical token literal outside a marked non-normative example | Design-document validation fails. |

### 3.4 CSS variable interface

Design contract. If an implementation exposes design tokens as CSS custom properties, it MUST use this mapping:

```text
--ct-<namespace>-<token-key>
```

Design contract. For nested component tokens, the CSS custom property mapping MUST be:

```text
--ct-component-<component-key>-<property-key>
```

Design contract. If an implementation does not expose public CSS variables, design conformance MUST be verified through rendered visual and accessibility fixture evidence rather than variable-name inspection.

### 3.5 Theme registry

Design contract. The current theme registry is closed to the rows in this table.

| Theme ID | Required | Exposure | Omission behavior |
| --- | ---: | --- | --- |
| `dark_graphite` | Yes | Default and only required theme. | Not applicable. |
| `light` | No | MUST NOT be exposed as a supported theme in this revision. | Omission is conformant. |
| `high_contrast` | No | MUST NOT be claimed as supported unless a later design revision defines complete tokens, contrast targets, and fixtures. | Omission is conformant, but `dark_graphite` still MUST satisfy §14 accessibility criteria. |

Design contract. A theme switcher MUST NOT be exposed unless every selectable theme has a complete token registry, accessibility matrix, and visual fixture coverage.

Design contract. Browser or operating-system forced-color modes can alter rendering outside this document's dedicated theme registry. This revision does not claim a dedicated forced-color conformance profile; omission of a dedicated forced-color theme is conformant when required keyboard, non-color state, and contrast criteria remain satisfied for `dark_graphite`.

### 3.6 Density registry

Design contract. Density selection is closed to `compact`, `default`, and `comfortable`. The default density is `{density.default-mode}`.

Design contract. All workbook surfaces MUST use the same active density mode. Per-surface density overrides are invalid in this revision. User-selected density is valid as client or user preference only when it applies uniformly to all workbook surfaces; if no such preference is persisted, `{density.default-mode}` applies.

Design contract. Large incident grids use fixed-height rows by default. Variable-height rows are valid only in inspector sections, preview areas, or non-grid detail regions. Omission of variable-height grid rows is conformant.

### 3.7 Iconography profile

Design contract. A conforming implementation MUST use exactly one icon family in the workbook shell.

Design contract. This revision uses a bounded substitution profile instead of naming a package. The selected family MUST satisfy every constraint below.

| Constraint | Requirement |
| --- | --- |
| Style | Outline-only, single-stroke icon family. |
| Stroke weight | One default stroke weight across the shell, except optical corrections inside the selected family. |
| Size grid | `16px` for dense inline affordances; `20px` for toolbar controls; no other shell icon size unless a component table declares it. |
| Filled icons | Forbidden unless a later revision enumerates a specific semantic exception. |
| Family mixing | Forbidden inside one rendered shell. |
| Semantic IDs | Icons MUST be addressed by stable semantic icon IDs in application code or design fixtures, not by visible labels. |
| Missing icon | Render the text label or a fallback glyph with the same accessible name. Silent omission is invalid. |
| Icon-only control | MUST have an accessible name and visible focus indicator. |

Design contract. Icons can support labels, but icons MUST NOT replace labels for destructive actions, conflict resolution, evidence preview, evidence download, rollback, merge initiation, party link/unlink, inspector close, inspector pin, or system-view switching.

## 4. Product and interaction contract

### 4.1 Product thesis

Design contract. Cartulary is a workbook-native incident workspace. The design MUST feel like a serious low-friction workbook on the hot path while behaving like a disciplined case system underneath.

Design contract. The primary interaction object is the visible workbook row, cell, chip, count, preview affordance, filter state, grouping state, and status state. The storage model, route handlers, projection tables, and grid vendor coordinates MUST NOT be exposed as user-facing design concepts.

Core restatement. Core and domain materials define Cartulary as preserving the spreadsheet mental model at the view layer while keeping source state typed, relational, versioned, attributed, auditable, and recoverable underneath.

Design contract. The UI MUST preserve spreadsheet speed where it matters: direct typing, paste, compact scanning, keyboard navigation, visible rows, flexible filtering, and fast reshaping of the current working set.

Design contract. The UI MUST reject spreadsheet failure modes that damage incident work: row-position identity, silent overwrites, hidden relationship semantics, evidence paths as authority, unmanaged binary storage, and unversioned history.

Non-goal. Cartulary is not a dashboard, ticket queue, CRM, SIEM, EDR, evidence vault, long-form report editor, or command-center visualization. The shell MUST NOT adopt those products' information architecture as the default workbook model.

### 4.2 Product goals and design consequences

Design contract. All rows in the table below are binding design consequences.

| Goal | Design consequence |
| --- | --- |
| Preserve low-friction capture | Default creation and correction happen in the grid or directly adjacent inspector, not in detached forms. |
| Support progressive structure | Rough text, unresolved mentions, canonical chips, auto-resolved chips, and dismissed mentions have distinct visual states. |
| Keep context during live work | Evidence preview, conflict resolution, history, and relationship actions keep the grid visible in the base viewport. |
| Make collaboration legible | Save state, presence, pending work, and same-field conflicts appear as local semantic states. |
| Keep coordination workbook-native | Task Requests, Decisions, Communications Log, Handoff, Status Review, and Lesson surfaces use workbook grammar rather than separate modules. |
| Protect authority boundaries | UI labels, component names, SQL names, route helper names, and vendor grid coordinates never become behavior authorities. |

### 4.3 Target users and workflows

Design contract. Cartulary MUST serve front-line analysts, incident leads, reviewers, report writers, threat hunters, detection engineers, stakeholder coordinators, and deployment administrators without turning ordinary capture into ceremony.

| User | Primary needs | Design posture |
| --- | --- | --- |
| Front-line analyst | Capture observations quickly, paste data, attach screenshots/files, keep working with incomplete facts. | Grid-first, compact, same-surface feedback, minimal modal interruption. |
| Incident lead or commander | See scope, queues, blockers, ownership, unresolved items, and status-review material. | Persistent shell context, system-view switcher, saved views, readable counts, compact status. |
| Reviewer or report writer | Inspect lineage, attribution, evidence, history, rollback eligibility, and stable references. | Inspector-first review surfaces, row-local history, clear destructive-action boundaries. |
| Threat hunter or detection engineer | Pivot through indicators, queries, scoped hosts, identities, and linked evidence. | Linkable chips, preserved raw values, same-shell pivots, dense sortable and filterable views. |
| Stakeholder coordinator | Track communications, decisions, handoffs, report timing, and follow-up. | Workbook-native coordination surfaces with owner, state, and linkage clarity. |
| Deployment administrator | Manage deployment-local users and recovery concerns outside incident data. | Not a design center for the workbook shell; deployment administration MUST NOT blur into incident surfaces. |

Design contract. Primary workflows MUST use the design treatments in the table below.

| Workflow | Required design treatment |
| --- | --- |
| Open incident and orient | Show incident identity, active surface, built-in tabs, `System views`, saved-view selector, presence, grid, inspector affordance, and save state in one shell. |
| Create rough Timeline row | Permit entry with incomplete time, uncertain prose, unresolved host or account strings, or an attachment-only signal. Do not require canonical selection before capture. |
| Paste or bulk-enter rows | Preserve the user's place, show interpretation and validation locally, and use staged feedback when slower work continues. |
| Resolve host, identity, or indicator references | Use inspector or same-surface enrichment; distinguish unresolved text, resolved chip, auto-resolved chip, and dismissed mention. |
| Attach or request evidence | Show requested, pending receipt, received, pending upload slot, available, failed, quarantined, blocked preview, and inconsistent states without implying fake attachment. |
| Preview or download evidence | Keep the workbook shell present; blocked preview remains inline and does not silently fall back to download. |
| Filter, sort, group, and save views | Treat the current view as a workbook working set; saved views remain configurations over one `view_schema_id`. |
| Resolve same-field conflict | Mark only the affected cell, keep the saved value visible, retain the local draft separately, and open a resolver from that cell. |
| Review history or rollback | Use row-local history and scoped destructive-action presentation in the inspector; routine editing MUST NOT feel like approval workflow. |
| Coordinate tasks, decisions, handoffs, status reviews, and lessons | Keep these as workbook-native surfaces, not separate task-management or workflow modules. |

## 5. Visual design principles

### 5.1 Dense graphite forensic workspace

Design contract. The default visual direction is a dark, dense, graphite workspace with warm operational emphasis. It MUST feel calm, precise, inspectable, and durable. It MUST NOT feel cyberpunk, theatrical, militarized, playful SaaS, or generic enterprise gray.

Design contract. The UI MUST use subtle surface elevation, crisp hairlines, stable row rhythm, and restrained contrast. Evidence, history, and collaboration state MUST remain inspectable without turning the workspace into a wall of alerts.

### 5.2 Accent scarcity

Design contract. `{colors.accent}` is the Cartulary accent. It MUST be used sparingly for focus rings, selected shell controls, primary affirmative action, active grid handles, and brand emphasis.

Design contract. `{colors.accent}` MUST NOT be used as the default warning color, row background, large panel fill, heatmap color, decorative gradient, or arbitrary chart color.

Design contract. Caution and warning semantics MUST use `{colors.semantic-caution}` or an equivalent semantic state token and MUST include text, shape, marker, or accessible name. Primary accent-filled buttons MUST use `{colors.on-accent}` for text or icons.

### 5.3 Local semantic feedback

Design contract. State MUST appear at the locus where it changes user action.

| State family | Required primary locus |
| --- | --- |
| Same-field conflict | Affected cell. |
| Evidence state | Row, evidence chip, Evidence surface, or inspector Evidence section. |
| Presence | Header, row gutter, or cell according to §10.2. |
| Save state | Status strip. |
| Auto-resolution disclosure | Affected chip, batch result, or inspector relationship section. |
| Queue overflow | Status strip plus same-surface non-modal message. |

### 5.4 Structure without ceremony

Design contract. The UI MUST make structure visible without demanding structure before capture. Users can type rough values first. Later normalization, resolution, review, and rollback become available through adjacent controls and stateful chips.

### 5.5 Accessibility as visual language

Design contract. Color MUST NOT be the only carrier of state. Every state-bearing color MUST be paired with at least one non-color cue: shape, marker text, accessible name, icon with accessible name, tooltip, live-region announcement, or equivalent.

Design contract. Focus MUST be visible in dense layouts and MUST NOT be represented only by color shift.

## 6. Color, typography, motion, and base component values

### 6.1 Color usage

Design contract. The token registry defines the color values. The body of this document governs usage only.

| Token role | Required use | Forbidden use |
| --- | --- | --- |
| `{colors.accent}` | Focus, selected shell control, primary affirmative action, active grid affordance, brand emphasis. | Warning palette, row heatmap, decorative fill, alert background. |
| `{colors.canvas}` | Application background and deepest grid surround. | State marker by itself. |
| `{colors.surface-1}` through `{colors.surface-4}` | Shell chrome, view bar, cells, menus, raised panels, selected neutral surfaces. | Semantic success, warning, conflict, or destructive state without an additional semantic marker. |
| `{colors.hairline}` and `{colors.hairline-strong}` | Grid lines, panel separators, control borders, pinned boundaries, inspector edge. | Replacement for visible focus. |
| `{colors.semantic-info}` | Informational state, source metadata, neutral system messages. | Relationship identity by itself. |
| `{colors.semantic-success}` | Completed state, successful attach, completed task, saved confirmation when a success marker is needed. | Persistent celebratory decoration. |
| `{colors.semantic-caution}` | Warning, stale, pending verification, unsupported preview. | Accent or brand emphasis. |
| `{colors.semantic-conflict}` | Same-field conflict, replay blocked, invalid cell. | Destructive action confirmation by itself. |
| `{colors.semantic-destructive}` | Delete, rollback, supersede, destructive confirmation. | Primary affirmative action. |

### 6.2 Typography

Design contract. Typography tokens are defined once in the token registry. UI surfaces MUST use the role tokens in the table below.

| UI role | Required token |
| --- | --- |
| Surface title | `{typography.surface-title}` |
| Section heading | `{typography.section-heading}` |
| Grid cell | `{typography.grid-cell}` |
| Body/UI | `{typography.ui}` |
| Metadata | `{typography.metadata}` |
| Button | `{typography.button}` |
| Code-like identifiers, hashes, timestamps, route tokens, exact values | `{typography.mono}` |

Design contract. Long prose belongs in the inspector, Notes, findings surfaces, or companion findings document. Timeline and coordination grid cells MUST remain compact and scannable.

### 6.3 Motion

Design contract. Motion MUST be short, functional, and non-essential. Motion MUST NOT be the only signal for a state change.

Design contract. When `prefers-reduced-motion: reduce` is active, non-essential transitions MUST be disabled or capped at `{motion.duration-reduced-max}`. Essential transitions are valid at `{motion.duration-reduced-max}` only when they preserve orientation and no motion-free substitute is available.

## 7. Shell, surface composition, and responsive behavior

### 7.1 Shell regions

Design contract. At the base viewport `{layout.baseViewport}` and `{layout.zoomDefault}`, the application shell MUST render one continuous workspace with the five persistent regions in the table below.

| Region | Required contents | Boundary |
| --- | --- | --- |
| Top bar | Incident identity, built-in tabs, `System views` switcher, current system-view title, presence summary. | Persistent chrome, not dashboard summary. |
| View bar | Active surface selector, saved-view selector, sort, group, filters, active chips. | Belongs to the active surface only. |
| Grid | Active workbook surface with `record_id`-bound rows and `field_key`-bound cells. | Primary work surface. |
| Inspector | Details, Relationships, Evidence, History. | Adjacent secondary surface at base viewport. |
| Status strip | Primary save-state label, secondary message, presence summary or overflow. | Capacity-limited working-state surface. |

Design contract. The shell MUST expose active surface identity, primary save-state label, and a keyboard-reachable route back to the active grid in every supported viewport band.

### 7.2 Shell-exposure surface registry

Core restatement. Core 01 defines fourteen required pack-independent `view_schema` entries and three standardized optional workbook surfaces. This document does not redefine their source record types, field registries, write-back rules, or route behavior.

Design contract. The shell-exposure registry below is exhaustive for this design revision. `Display label` is a display hint. `view_schema_id` is the canonical behavior key.

| Display label | `view_schema_id` | Required status | Shell exposure | Group label | Group order | Surface order | Saved-view placement | Optional exposure rule |
| --- | --- | --- | --- | --- | ---: | ---: | --- | --- |
| Timeline | `cartulary.view.timeline.v1` | Required built-in sheet | Primary tab at base viewport | Built-in | 1 | 1 | Active surface view selector | Not optional. |
| Hosts | `cartulary.view.hosts.v1` | Required built-in sheet | Primary tab at base viewport | Built-in | 1 | 2 | Active surface view selector | Not optional. |
| Identities | `cartulary.view.identities.v1` | Required built-in sheet | Primary tab at base viewport | Built-in | 1 | 3 | Active surface view selector | Not optional. |
| Evidence | `cartulary.view.evidence.v1` | Required built-in sheet | Primary tab at base viewport | Built-in | 1 | 4 | Active surface view selector | Not optional. |
| Notes | `cartulary.view.notes.v1` | Required built-in sheet | Primary tab at base viewport | Built-in | 1 | 5 | Active surface view selector | Not optional. |
| Indicators | `cartulary.view.indicators.v1` | Required system view | `System views` | Scope and assessment | 2 | 1 | Active surface view selector | Not optional. |
| Compromise Assessments | `cartulary.view.assessments.v1` | Required system view | `System views` | Scope and assessment | 2 | 2 | Active surface view selector | Not optional. |
| Parties | `cartulary.view.parties.v1` | Required system view | `System views` | Scope and assessment | 2 | 3 | Active surface view selector | Not optional. |
| Task Requests | `cartulary.view.task_requests.v1` | Required system view | `System views` | Coordination | 3 | 1 | Active surface view selector | Not optional. |
| Decisions | `cartulary.view.decisions.v1` | Required system view | `System views` | Coordination | 3 | 2 | Active surface view selector | Not optional. |
| Communications Log | `cartulary.view.comm_log.v1` | Required system view | `System views` | Coordination | 3 | 3 | Active surface view selector | Not optional. |
| Handoff | `cartulary.view.handoff.v1` | Required system view | `System views` | Coordination | 3 | 4 | Active surface view selector | Not optional. |
| Status Review | `cartulary.view.status_review.v1` | Required system view | `System views` | Review and learning | 4 | 1 | Active surface view selector | Not optional. |
| Lesson | `cartulary.view.lesson.v1` | Required system view | `System views` | Review and learning | 4 | 2 | Active surface view selector | Not optional. |
| Findings | `cartulary.view.findings.v1` | Standardized optional workbook surface | `System views` only when implemented and exposed | Optional artifact surfaces | 5 | 1 | Active surface view selector | If not implemented, MUST NOT appear. |
| Investigative Queries | `cartulary.view.investigative_queries.v1` | Standardized optional workbook surface | `System views` only when implemented and exposed | Optional artifact surfaces | 5 | 2 | Active surface view selector | If not implemented, MUST NOT appear. |
| Forensic Keywords | `cartulary.view.forensic_keywords.v1` | Standardized optional workbook surface | `System views` only when implemented and exposed | Optional artifact surfaces | 5 | 3 | Active surface view selector | If not implemented, MUST NOT appear. |

Design contract. Required system views MUST NOT be command-palette-only. All required system views MUST be reachable by keyboard and pointer from the shell.

Design contract. If a standardized optional surface is not implemented, it MUST NOT appear in the switcher. If a required surface is unavailable because of implementation defect or load failure, the UI MUST show an error state for that surface instead of silently removing it from the shell.

Design contract. Saved views MUST appear under the active surface's view selector. A saved view MUST NOT replace canonical system-view identity or create a new primary tab by default.

Design contract. Constrained UI labels can use `Assessments` as a display shorthand for `Compromise Assessments` only when the accessible name or surrounding context exposes the full label. No semantic distinction is intended.

### 7.3 Inspector contract

Design contract. The inspector MUST use the following default and bounds.

| Property | Required behavior |
| --- | --- |
| Default width | `{layout.inspectorDefaultWidth}`. |
| Minimum width | `{layout.inspectorMinWidth}`. |
| Maximum width | `{layout.inspectorMaxWidth}`. |
| Resize | User can resize within min/max bounds. Omission of resize persistence is conformant. |
| Pinning | Client-local only; MUST NOT persist in saved views. |
| Section order | Details, Relationships, Evidence, History. |
| Grid visibility | At base viewport, grid remains visible whenever inspector is open. |
| Close affordance | Visible control with accessible name `Close inspector`. |
| Pin affordance | Visible control with accessible name `Pin inspector` or `Unpin inspector`. |

Design contract. If the inspector opens as an overlay in a narrower viewport band, the grid behind the overlay MUST be inert to pointer and keyboard until the overlay closes.

### 7.4 Responsive breakpoint matrix

Design contract. Viewport-band selection MUST use this deterministic algorithm:

```pseudocode
select_viewport_band(width_css_px, height_css_px):
  if width_css_px >= layout.baseMinWidth and height_css_px >= layout.baseMinHeight:
    return base
  if width_css_px >= layout.narrowMinWidth and height_css_px >= layout.baseMinHeight:
    return narrow_desktop
  if width_css_px >= layout.compactMinWidth and height_css_px >= layout.compactMinHeight:
    return compact_desktop
  return below_supported_minimum
```

Design contract. Each viewport band MUST render according to this table.

| Viewport band | Width and height condition | Top bar | View bar | Grid | Inspector | Status strip | Unsupported behavior |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `base` | Selected by algorithm above. | Incident identity, built-in primary tabs, `System views`, current surface title, presence summary. | Full active-surface controls. | Primary, visible while inspector is open. | Adjacent right panel. | Visible primary save label and secondary message. | None. |
| `narrow_desktop` | Selected by algorithm above. | Incident identity, `Surfaces` control, `System views`, current surface title, presence summary. | Full active-surface controls; chips can wrap into one additional row; if not wrapped, all controls remain keyboard reachable. | Primary; remains visible when overlays are closed. | Full-height right overlay; grid inert behind overlay. | Visible primary save label; secondary message can truncate when accessible full text remains available. | Primary built-in tabs collapse into `Surfaces` control. |
| `compact_desktop` | Selected by algorithm above. | Incident identity, `Surfaces`, `System views`, current surface title. Presence summary can move to status strip; if it stays in the top bar, status strip presence summary is omitted. | Active-surface controls remain keyboard reachable; filter chips can move into a `Filters` popover; if they remain inline, the popover is omitted. | Primary when overlays are closed. | Full-screen or full-height overlay; grid inert behind overlay. | Visible primary save label. Secondary message can be accessible-only when space is exhausted; if no secondary message exists, nothing is announced. | No all-tabs layout. No command-palette-only system view access. |
| `below_supported_minimum` | Selected by algorithm above. | Safe navigation and session controls remain reachable. | Not required. | Can degrade, horizontally scroll, or show supported-viewport message. | Can be unavailable. | Primary save label MUST remain visible when unsaved work exists. | The implementation MUST NOT claim design conformance for this viewport. |

Design contract. Below the supported minimum, keyboard session logout and safe navigation MUST remain available. Omission of mobile/touch-specific gestures is conformant.

## 8. Grid, view bar, and workbook interaction

### 8.1 Grid identity and addressing

Core restatement. Workbook behavior is keyed by stable identifiers such as `view_schema_id`, `record_id`, `row_version`, `field_key`, and `client_txn_id`, not by visible tab labels, row numbers, column labels, projection table names, storage table names, or React component names.

Design contract. The UI can display human labels, but interaction state, mutation affordances, focus anchoring, visual fixture selectors, and accessibility relationships MUST bind to stable identifiers when those identifiers exist.

Design contract. Presentation-only rows such as group headers, empty states, loading rows, and measurement rows MUST NOT emit mutation events and MUST NOT be presented as authoritative records.

### 8.2 Cell and row state precedence

Design contract. Cell and row visual state selection MUST use the precedence table below. Priority `1` is highest.

| Priority | State ID | Applies when | Co-display rule | Required non-color cue |
| ---: | --- | --- | --- | --- |
| 1 | `conflicted` | Same-field conflict exists for the cell. | May co-display active focus. Overrides invalid and pending visual treatment. | Marker with accessible name `Conflict on <field label>`. |
| 2 | `invalid` | Current cell value fails field or route validation. | May co-display active focus. Overrides dirty and pending. | Error marker and field-local message. |
| 3 | `dirty_or_pending` | Local edit or queued mutation exists and no higher-priority conflict or invalid state applies. | May co-display active focus. | Pending marker and accessible pending label. |
| 4 | `active_cell` | Cell has keyboard focus or edit focus. | Co-displays with higher-priority state. | Visible focus outline. |
| 5 | `selected_row` | Row is selected and no cell-local primary marker replaces it. | Row-level only; does not override cell-local markers. | Row selection marker and accessible selected state. |
| 6 | `read_only_or_derived` | Cell is read-only, hidden-derived, or non-writable. | Applies when no higher-priority error, conflict, pending, or active edit state applies. | Accessible read-only state or visible read-only marker on focus. |
| 7 | `saved` | No higher-priority state applies. | Default steady state. | Status strip `Saved`; no persistent success decoration required. |

Design contract. `invalidate_or_refresh_required` is row-block state, not a cell-state priority. When it applies, the row or block MUST show a stale marker and MUST trigger refresh through the owner route behavior rather than inventing a design-local read path.

### 8.3 View bar controls

Design contract. Sort, filter, group, saved-view, and surface controls live in the View bar and apply to the active surface only.

| Control | Default state | Active state | Invalid state | Clear behavior | Ordering |
| --- | --- | --- | --- | --- | ---: |
| Active surface selector | Current `view_schema_id` display label. | Same. | Surface unavailable error. | Not clearable. | 1 |
| Saved-view selector | `Unsaved view` when no saved view is active. | Saved view display name. | Fall back to base surface and show inline message. | Clears saved-view selection only, not active query unless the user selects reset. | 2 |
| Sort control | No user sort override. | Sort chips shown. | Invalid sort field blocked before persistence. | Clears all user sort overrides. | 3 |
| Group control | `Group: None`. | One group chip. | Unsupported group disabled with explanation. | Sets grouping inactive. | 4 |
| Filter control | No filters. | Filter chips shown. | Invalid filter chip marked and excluded from query submission. | Clears all filters or one selected chip. | 5 |

Design contract. Active chips MUST render in this order: group chip, sort chips in applied order, then filter chips in normalized query order.

Design contract. Saved views MUST feel like named configurations of a surface, not new canonical surfaces. A `system` saved view is not the same object as a required system view.

Design contract. Saved-view dirty state MUST be selected by this deterministic algorithm:

```pseudocode
is_saved_view_dirty(active_saved_view, current_query, current_layout):
  if active_saved_view is null:
    return false
  return normalize(current_query) != active_saved_view.query_json
      or normalize(current_layout) != active_saved_view.layout_json
```

Design contract. Dirty saved-view indication MUST NOT imply unsaved incident data. It indicates that the view configuration differs from the selected saved-view configuration.

### 8.4 Keyboard interaction matrix

Design contract. Keyboard behavior MUST use the matrix below. `Tab` order MUST enter each major region once before entering roving-focus children inside that region.

| Region or component | Entry key/path | Navigation keys | Activation | Commit | Cancel | Escape behavior | Focus return | Fallback focus |
| --- | --- | --- | --- | --- | --- | --- | --- | --- |
| Top bar built-in tabs | `Tab` to tablist | Arrow keys between tabs | `Enter` or `Space` selects | Selection commits immediately | `Esc` no-op | No-op unless menu open | Active tab | Incident identity |
| `Surfaces` control | `Tab` to button | Arrow keys in menu | `Enter` or `Space` selects | Selection commits immediately | `Esc` closes menu | Close menu | Invoking control | Active surface selector |
| `System views` switcher | `Tab` to button | Arrow keys by group and item | `Enter` or `Space` selects | Selection commits immediately | `Esc` closes menu | Close menu | Invoking control | Active surface selector |
| Saved-view selector | `Tab` to selector | Arrow keys in menu | `Enter` or `Space` selects | Selection commits immediately | `Esc` closes menu | Close menu | Invoking control | Active surface selector |
| Filter chips | `Tab` to chip group | Arrow keys between chips | `Enter` opens editor; `Delete` removes focused chip | Apply control commits | `Esc` closes editor | Close chip editor | Focused chip or chip group | Filter control |
| Grid navigation mode | `Tab` to grid | Arrow keys, Page keys, Home/End according to grid adapter | `Enter` or printable key enters edit when writable | Not applicable | `Esc` no-op | No-op | Active cell | Grid container |
| Grid edit mode | From active writable cell | Text-editor keys | Editor-specific | `Enter`, blur by explicit commit, or declared commit shortcut | `Esc` discards uncommitted editor value | Exit edit mode | Edited cell | Grid container |
| Relationship chip | `Tab` or arrow within cell | Arrow keys within chip list | `Enter` opens inspect/action menu | Action commits through owner route | `Esc` closes menu | Close menu | Invoking chip | Owning cell |
| Inspector tab/section | `Tab` to inspector | Arrow keys for tabs; headings navigable by Tab only when interactive | `Enter` or `Space` opens control | Control-specific | `Esc` closes overlay inspector only | Close overlay inspector in responsive bands | Invoking row/control | Active grid container |
| Evidence preview | Evidence action control | Preview-internal controls | `Enter` or `Space` on controls | Download or action-specific | `Esc` closes preview | Close preview | Invoking evidence affordance | Inspector Evidence section |
| Menu or popover | Invoking button | Arrow keys, Home/End | `Enter` or `Space` selects item | Selection commits unless item opens nested UI | `Esc` closes | Close menu/popover | Invoking control | Region container |
| Toast with focusable action | `Tab` to toast action | `Tab` within actions | `Enter` or `Space` activates | Action-specific | `Esc` dismisses only when safe dismiss is declared | Dismiss toast | Prior focus target | Status strip |
| Dialog | Triggering action | `Tab` trapped within dialog | `Enter` or `Space` activates focused action | Primary or destructive confirmation as declared | `Esc` safe-cancel only when declared | See §8.5 | Invoking control | Dialog safe control or shell top bar |
| Same-field conflict resolver | Conflicted cell | `Tab` through compare and actions | `Enter` or `Space` activates focused resolution action | Explicit resolution action only | `Esc` closes resolver without committing | Close resolver | Conflicted cell | Active grid container |

### 8.5 Escape priority ladder

Design contract. When multiple dismissible layers are open, `Esc` MUST resolve exactly one layer using this priority order.

| Priority | Active condition | Required result |
| ---: | --- | --- |
| 1 | Modal dialog open | Close only through safe cancel if safe cancel exists. If no safe cancel exists, `Esc` does nothing and accessible help explains the required action. |
| 2 | Same-field conflict resolver open | Close resolver without committing and return focus to conflicted cell if still present. |
| 3 | Evidence preview open | Close preview and return focus to invoking evidence affordance if still present. |
| 4 | Menu or popover open | Close menu or popover and return focus to invoking control. |
| 5 | Grid cell editor open | Discard uncommitted editor value and return to grid navigation mode. |
| 6 | Inspector open as overlay | Close inspector and return focus to invoking row or control. |
| 7 | No dismissible layer | No-op. |

Design contract. Focus restoration MUST use this fallback ladder when the invoking element no longer exists:

1. invoking cell if the row and field still exist;
2. invoking row if the field no longer exists;
3. active grid container if the row no longer exists but the surface exists;
4. active surface selector if the surface still exists;
5. top bar incident identity if no active surface can be restored.

## 9. Progressive structure, chips, and semantic states

### 9.1 Chip state selection

Design contract. Entity-mention and relationship chips MUST use this selection algorithm:

```pseudocode
select_chip_state(input):
  if input.dismissed == true:
    return dismissed
  if input.resolved_record_id is null:
    return unresolved
  if input.resolution_method == "auto":
    return auto_resolved
  return resolved
```

Design contract. Chip states MUST use the closed visual vocabulary below.

| State | Border | Required marker | Accessible name pattern | Notes |
| --- | --- | --- | --- | --- |
| `unresolved` | Dashed or dotted. | Leading `?` or visible `Unresolved`. | `Unresolved <entity type> mention: <raw text>`. | Must differ from ordinary text and resolved chips. |
| `resolved` | Solid. | No unresolved marker. | `Resolved <entity type>: <display name>`. | Shows canonical target while retaining inspection path to raw mention. |
| `auto_resolved` | Solid plus auto marker. | Visible `auto`. | `Auto-resolved <entity type>: <display name>; matched <alias text>`. | Remains inspectably marked after transient disclosure fades. |
| `dismissed` | Low-emphasis chip or token. | Visible `dismissed`. | `Dismissed mention: <raw text>`. | Display only where inspectable; excluded from active relationship values. |

Design contract. Color alone MUST NOT distinguish chip states.

### 9.2 Relationship and collection cells

Design contract. Relationship cells that mix unresolved mention tokens and canonical chips MUST preserve those as different object types. They MUST NOT be coerced into comma-delimited strings for display, editing, conflict resolution, or copy/paste presentation.

Design contract. Collection cells can display compact summaries when space is constrained. Omission of hidden collection members in compact display is conformant only when an inspectable expansion path is present and the accessible name reports the hidden count.

### 9.3 Auto-resolution disclosure

Design contract. If a value is auto-resolved, the user MUST be able to tell it was auto-resolved, inspect the match reason when such reason is available, and correct it through the same surface or inspector.

Design contract. Transient auto-resolution disclosure can fade, but the chip state MUST remain marked as `auto_resolved` until the underlying resolution state changes. Omission of fade is conformant.

## 10. Collaboration, save state, presence, and conflict design

### 10.1 Save-state labels and selection algorithm

Core restatement. Core 03 requires the UI to present exactly `Syncing`, `Saved`, or `Conflict` as the compact save-state label, and ambient collaboration state does not change the mapping.

Design contract. The status strip MUST show exactly one primary save-state label selected by the algorithm below.

```pseudocode
select_primary_save_label(state):
  if state.same_field_conflict_count > 0:
    return Conflict
  if state.queue_overflow_refused == true:
    return Conflict
  if state.non_retryable_replay_failure == true:
    return Conflict
  if state.in_flight_workbook_mutation_count > 0:
    return Syncing
  if state.pending_queue_size > 0:
    return Syncing
  if state.replay_paused_for_recovery == true:
    return Syncing
  if state.reauthentication_required_for_replay == true:
    return Syncing
  return Saved
```

Design contract. `Conflict` has precedence over `Syncing`. Presence MUST NOT change the primary save label. A secondary message can describe lower-priority state, but omission of the secondary message MUST NOT change the primary label.

| Label | Visual treatment | Required accessible representation |
| --- | --- | --- |
| `Syncing` | Subtle spinner or pending marker plus label. | Text label and state description. |
| `Saved` | Text label only by default; no celebratory animation. | Text label. |
| `Conflict` | Conflict semantic marker and local entry point. | Text label, affected count when available, and action path. |

### 10.2 Presence rendering

Design contract. Presence MUST render at header, row, and cell levels according to this capacity table.

| Scope | Capacity | Overflow label |
| --- | ---: | --- |
| Header | 5 unique users | `+N` where `N` is hidden unique users. |
| Row | 3 unique users | `+N` where `N` is hidden unique users. |
| Cell | 1 unique user | `+N` is optional when space permits; when omitted, accessible description includes hidden count. |

Design contract. Presence rendering MUST use this deterministic algorithm:

```pseudocode
render_presence(scope, presences, now, self_user_id):
  active = filter presences where presences.expires_at > now
  scoped = filter active where presence matches the requested scope
  grouped = group scoped by user_id
  representative = for each user_id select one presence by:
    mode_weight(editing=1, focused=2, viewing=3, idle=4) ascending,
    observed_at descending,
    connection_id ascending
  ordered = sort representative by:
    user_id == self_user_id descending,
    mode_weight ascending,
    observed_at descending,
    display_name normalized with Unicode NFC ascending,
    user_id ascending
  visible = first capacity(scope) entries
  overflow = count(ordered) - count(visible)
  return visible, overflow
```

Design contract. Multiple connections from the same user count as one visible user for overflow. Expired presence MUST NOT render. Presence MUST NOT lock editing, change save state, or imply ownership.

### 10.3 Same-field conflict design

Core restatement. Same-field conflicts remain unresolved until an analyst explicitly chooses a resolution. Only the affected cell enters conflict state, and the workbook save-state presentation remains `Conflict` until every unresolved same-field local conflict has been cleared or resolved.

Design contract. A conflicted cell MUST display the current saved server value plus a visible conflict marker. The analyst's unsaved local value MUST be retained separately and MUST NOT be rendered as saved.

Design contract. The resolver MUST open from the conflicted cell and include row context, field label, stable `field_key`, saved value with actor and timestamp when available, local unsaved value, optional merge summary, and direct resolution actions.

Design contract. Initial focus in a conflict resolver MUST land on a non-destructive summary or safe control, not on `Use my unsaved value`.

## 11. Evidence design

### 11.1 Evidence dimensions

Design contract. Evidence UI state MUST be computed from three independent dimensions: evidence lifecycle, upload-slot overlay, and preview capability. The UI MUST NOT conflate evidence-record state with object-blob state or preview capability.

| Dimension | Closed values |
| --- | --- |
| Evidence lifecycle | `requested`, `pending_receipt`, `received`, `available`, `released`, `failed`, `quarantined`, `inconsistent` |
| Upload-slot overlay | `none`, `pending_upload_slot`, `failed_upload_slot` |
| Preview capability | `not_applicable`, `preview_available`, `preview_blocked`, `download_only`, `access_blocked` |

Design contract. Evidence state rendering MUST use the precedence table below. Priority `1` is highest.

| Priority | Condition | Required visual result | Affordance rule |
| ---: | --- | --- | --- |
| 1 | Evidence lifecycle is `inconsistent` | Inline inconsistent state. | Preview and download blocked until repaired. |
| 2 | Evidence lifecycle is `quarantined` | Exact `quarantined` label. | Preview and download blocked unless owner security behavior later permits a separate action. |
| 3 | Evidence lifecycle is `failed` | Failure marker. | Retry or fresh-slot affordance only when owner route behavior permits it. |
| 4 | Preview capability is `access_blocked` | Access-blocked message. | No preview or download action. |
| 5 | Upload-slot overlay is `failed_upload_slot` | Upload failure marker. | Retry or fresh-slot affordance only when owner route behavior permits it. |
| 6 | Upload-slot overlay is `pending_upload_slot` | Local pending marker. | No evidence count increment solely from the slot. |
| 7 | Evidence lifecycle is `requested` | `Requested` label and requested metadata when available. | No preview action. |
| 8 | Evidence lifecycle is `pending_receipt` | Pending receipt cue. | No preview action. |
| 9 | Evidence lifecycle is `received` | Received cue. | Attach blob or awaiting-upload affordance according to owner state. |
| 10 | Evidence lifecycle is `available` and preview capability is `preview_available` | Evidence count plus preview and download affordances as allowed. | Preview and download invoke owner handle contract. |
| 11 | Evidence lifecycle is `available` and preview capability is `download_only` | Available/download-only state. | Download allowed as owner handle permits; preview not shown. |
| 12 | Evidence lifecycle is `available` and preview capability is `preview_blocked` | Inline blocked-preview state. | No silent fallback to download. Download remains separate when owner handle permits it. |
| 13 | Evidence lifecycle is `released` | Released label plus access state derived from current preview capability. | Preview/download affordance still follows owner handle behavior. |

Design contract. Evidence counts MUST include evidence records that are attached or available according to owner behavior. A `pending_upload_slot` alone MUST NOT increment evidence count.

### 11.2 Evidence preview and download

Design contract. Preview MUST open in a side or bottom preview region without forcing full-page navigation away from the grid in supported viewport bands.

Core restatement. Preview and download affordances invoke the evidence-access handle contract; returned `href` values are opaque same-origin URLs.

Design contract. The browser MUST NOT synthesize object-store URLs, parse handle tokens for semantics, or treat object-store locations as evidence identity.

Design contract. Blocked preview MUST surface explicitly. Blocked preview MUST NOT silently collapse into download.

### 11.3 Evidence density

Design contract. Evidence counts belong in grid cells or chips. Detailed custody, collection source, blob metadata, and preview/download affordances belong in the inspector or Evidence surface.

Design contract. Dense grid rows MUST show enough evidence state for triage without becoming evidence-detail cards.

## 12. Component contracts

### 12.1 Component variant matrix

Design contract. Component families MUST implement the required variants in this table.

| Component | Required variants | Required closure |
| --- | --- | --- |
| Button | `default`, `hover`, `focus_visible`, `active`, `disabled`, `loading`, `destructive`; `selected` is required only for toggle buttons and segmented-control buttons. | Disabled blocks activation; loading blocks duplicate activation; destructive never uses accent fill. |
| Icon button | All button variants plus `icon_missing`. | Accessible name required; fallback text or glyph required when icon unavailable. |
| Text input/editor | `empty`, `filled`, `focus_visible`, `invalid`, `read_only`, `disabled`, `pending`, `conflicted`. | Placeholder MUST NOT be the only label outside grid edit mode. |
| Chip | `unresolved`, `resolved`, `auto_resolved`, `dismissed`, `selected`, `focus_visible`, `disabled`. | State MUST be distinguishable without color. |
| Menu/popover | `closed`, `opening`, `open`, `item_focus`, `item_disabled`, `empty`, `error`. | Focus returns to invoker; empty menu states have text. |
| Inspector section | `collapsed`, `expanded`, `loading`, `empty`, `error`, `read_only`. | Section heading remains visible when section is not empty. |
| Toast | `queued`, `visible`, `paused`, `dismissed`, `action_focused`. | Auto-dismiss pauses while hovered or focused. |
| Dialog | `open`, `safe_cancel`, `destructive_confirm`, `blocked_until_acknowledged`. | Initial focus lands on safe control unless acknowledgement is required. |
| Banner | `visible`, `persistent`, `dismissible`, `error`, `warning`, `info`. | Persistent condition banners remain until state clears or an owner-declared dismissal path exists. |
| Inline message | `info`, `success`, `warning`, `error`, `pending`. | Message is anchored to the affected cell, row, section, or action. |

### 12.2 Buttons

Design contract. Button treatment MUST follow this table.

| Component | Use | Treatment |
| --- | --- | --- |
| Primary button | One affirmative action in a local action group. | `{components.button-primary}`. |
| Secondary button | Ordinary non-primary action. | `{components.button-secondary}`. |
| Tertiary button | Low-emphasis toolbar action. | Transparent or surface hover only. |
| Destructive button | Delete, rollback, supersede, publish-danger, merge-danger. | `{components.button-danger}`; never accent fill. |
| Icon button | Dense toolbar, chip action, inspector control. | One icon family, accessible name, visible focus. |

Design contract. Disabled buttons MUST remain focusable only when focus is required to expose explanatory help. Otherwise disabled buttons can be skipped in tab order. In both cases, disabled state MUST be available to assistive technology.

### 12.3 Inputs and editors

Design contract. Inputs MUST use neutral surfaces, compact padding, visible focus, and no decorative glow. Validation MUST appear at the field or cell that caused it.

Design contract. Read-only inputs MUST expose read-only state. Disabled inputs MUST expose disabled state and MUST NOT permit editing. Pending inputs MUST expose pending state without implying the value is saved.

### 12.4 Menus and popovers

Design contract. Menus and popovers MUST use neutral raised surfaces, compact vertical rhythm, keyboard navigation, and focus return to the invoking control.

Design contract. The `System views` menu MUST follow the grouping order in §7.2. Empty groups MUST be omitted. Omission of optional-surface groups is conformant when no optional surface in that group is implemented.

### 12.5 Inspector sections

Design contract. Inspector sections MUST use this order and purpose.

| Section | Purpose | Default treatment |
| --- | --- | --- |
| Details | Main editable or readable fields not comfortable in the grid. | Stacked compact forms with clear labels. |
| Relationships | Mentions, resolved entities, indicators, links, parties. | Chip lists with source state and action affordances. |
| Evidence | Evidence records, counts, preview/download state, custody summary. | Compact evidence cards or rows. |
| History | Change summaries, actors, timestamps, rollback eligibility. | Row-local timeline with scoped actions. |

### 12.6 Empty states

Design contract. Empty states MUST stay in the active shell. They MUST name the surface, explain the minimum useful create action, and provide an in-place create affordance when creation is permitted.

Design contract. If creation is not permitted for the caller, the empty state MUST explain that no create action is available. It MUST NOT show an unlabeled disabled create control.

### 12.7 Banners, toasts, inline messages, and dialogs

Design contract. Message patterns MUST use this table.

| Pattern | Purpose | Duration |
| --- | --- | --- |
| Banner | Persistent same-surface state such as session expiry, queue overflow, collaboration degradation, or active-surface pack degradation. | Until state clears or an owner-declared dismissal path exists. |
| Toast | Transient confirmation such as completed long queue or batch disclosure. | Auto-dismiss after 5 seconds; pause while hovered or focused. |
| Inline message | Cell-local, row-local, or action-local issue. | Until corrected, replaced, or no longer relevant. |
| Dialog | Confirmation where confirmation is the point. | Until explicit action or safe cancel. |

Design contract. Routine row creation MUST NOT require modal dialogs. Dialogs are reserved for destructive-action confirmation, merge initiation confirmation, rollback confirmation, same-field conflict resolution only when an overlay cannot preserve context, and release-scope changes when the Snapshot and Reporting Extension Profile is implemented.

## 13. Per-surface design contracts

Design contract. This table defines design-local per-surface presentation behavior only. It MUST NOT define Core-owned field membership, create defaults, writeability, or route behavior.

| Surface | `view_schema_id` | Default row emphasis | Required local state cues | Inspector default section | Empty-state create affordance | Row actions | Design non-goals |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Timeline | `cartulary.view.timeline.v1` | Fast capture, chronology, rough summary. | Rough capture, unresolved mentions, evidence count, capture/review state. | Details. | Create Timeline row when permitted. | Resolve mentions, attach evidence, inspect history. | No form-first timeline entry. |
| Hosts | `cartulary.view.hosts.v1` | Scoped host identity and triage state. | Stub/canonical state, containment/scoping cues, relationship counts. | Relationships. | Create host when permitted. | Link timeline/evidence, inspect related records. | No asset-management module replacement. |
| Identities | `cartulary.view.identities.v1` | Scoped account or persona identity. | Stub/canonical state, privilege/reset/MFA cues when surfaced by owner field registry. | Relationships. | Create identity when permitted. | Link timeline/evidence, inspect related records. | No directory-service replacement. |
| Evidence | `cartulary.view.evidence.v1` | Evidence lifecycle and access state. | Lifecycle, upload-slot overlay, preview capability, collector/source context. | Evidence. | Request or add evidence when permitted. | Preview, download, attach blob, inspect custody/history. | No raw object-store browser. |
| Notes | `cartulary.view.notes.v1` | Structured analyst text object. | Linkage, author/update state, relationship cues. | Details. | Create note when permitted. | Link to row, inspect history. | No replacement for task, decision, handoff, or evidence state. |
| Indicators | `cartulary.view.indicators.v1` | Canonical indicator with observation pivots. | Observation count, lifecycle state, source-observation pivots. | Relationships. | Create indicator only when owner behavior permits. | Pivot to observations and related records. | No threat-intel dashboard. |
| Compromise Assessments | `cartulary.view.assessments.v1` | Assessment subject and rationale. | Subject chip, confidence/status cue, support links. | Details. | Create assessment when permitted. | Pivot to assessed host/identity and prior assessments. | No generalized scoring engine. |
| Parties | `cartulary.view.parties.v1` | Coordination identity. | Party kind, link counts, requester/collector/audience use cues. | Relationships. | Create party when permitted. | Link/unlink party refs, inspect related coordination records. | No replacement for deployment users or identities. |
| Task Requests | `cartulary.view.task_requests.v1` | Queue-oriented work. | Owner, status, priority, due, blocked cues. | Details. | Create task request when permitted. | Assign, link records, inspect blockers/history. | No separate task-management module. |
| Decisions | `cartulary.view.decisions.v1` | Rationale-bearing choice. | Owner, status, review class, support refs. | Details. | Create decision when permitted. | Link support records, inspect supersession/history. | No generalized approval workflow for ordinary edits. |
| Communications Log | `cartulary.view.comm_log.v1` | Durable communication memory. | Audience, channel/meeting context, decision and action links. | Details. | Create communications-log row when permitted. | Link parties, decisions, task requests. | No chat transcript system. |
| Handoff | `cartulary.view.handoff.v1` | Continuity at work boundary. | Current state, open work, open decisions, risks, next checks. | Details. | Create handoff when permitted. | Link open tasks/decisions, acknowledge when owner behavior permits. | No mandatory per-edit handoff ritual. |
| Status Review | `cartulary.view.status_review.v1` | Coordination checkpoint. | Blocked work, pending evidence, open decisions, risk summary, next report timing. | Details. | Create status-review row when permitted. | Link tasks/evidence/decisions. | No dashboard module. |
| Lesson | `cartulary.view.lesson.v1` | Retrospective improvement. | Follow-up tasks, evidence refs, closure state. | Details. | Create lesson when permitted. | Link follow-up work and evidence. | No forced retrospective ceremony. |
| Findings | `cartulary.view.findings.v1` | Optional structured finding. | Finding state, confidence, owner, support refs when implemented. | Details. | Create finding when implemented and permitted. | Link support records. | No base-profile requirement to expose. |
| Investigative Queries | `cartulary.view.investigative_queries.v1` | Optional query library. | Platform, purpose, query text, source links when implemented. | Details. | Create query when implemented and permitted. | Copy/query-link actions when implemented. | No SIEM query executor requirement. |
| Forensic Keywords | `cartulary.view.forensic_keywords.v1` | Optional keyword library. | Pattern, reason, source links when implemented. | Details. | Create keyword when implemented and permitted. | Copy/link actions when implemented. | No YARA/Sigma rule-management requirement. |

## 14. Accessibility contract

### 14.1 Accessibility conformance matrix

Design contract. The design floor is WCAG 2.2 AA for state-bearing text and controls in the required `dark_graphite` theme.

| Area | Required behavior | Testable condition |
| --- | --- | --- |
| Contrast | State-bearing text and controls MUST meet the declared WCAG 2.2 AA target in `dark_graphite`. | Contrast check passes for every token pair used in required fixtures. |
| Focus | Focus indicator MUST be visible and at least as prominent as `{border.focus}` or a documented equivalent. | Keyboard traversal evidence confirms focus visibility. |
| Non-color state | Every state MUST expose text, shape, marker, icon with accessible name, or equivalent. | State fixtures remain distinguishable in grayscale and to screen readers. |
| Live regions | Save-state changes, conflicts, queue overflow, replay blocked, session re-authentication required, and evidence access blocked MUST follow §14.2. | Event matrix maps each event to `polite`, `assertive`, or no live announcement. |
| Icon-only controls | Every icon-only control MUST have an accessible name. | Accessibility tree contains non-empty name. |
| Reduced motion | `prefers-reduced-motion: reduce` disables or caps non-essential transitions. | Motion fixture verifies duration cap from §6.3. |
| Row labels | Accessible row names MUST include human-readable surface context and stable row relationship. Raw `record_id` alone is forbidden. | Accessibility tree check confirms human-readable context. |
| Keyboard | No hot-path operation is pointer-only. | Keyboard matrix coverage passes for §8.4 rows. |

### 14.2 Live-region event matrix

Design contract. Live-region behavior MUST use this matrix.

| Event | Required announcement | Politeness |
| --- | --- | --- |
| Save label changes to `Conflict` | Announce conflict state and affected count if available. | `assertive` |
| Save label changes to `Syncing` | Announce only when replay paused, queue overflow recovered, or re-authentication is required. | `polite` |
| Save label changes to `Saved` | Announce only after a prior non-saved state. | `polite` |
| Conflict resolver opens | Announce row context and field label. | `polite` |
| Evidence preview blocked | Announce blocked preview and available next action. | `polite` |
| Inspector opens or closes | Announce section and focus target. | `polite` |
| Queue overflow refuses admission | Announce overflow and that current edit remains local unsaved work. | `assertive` |
| Session re-authentication required for replay | Announce that unsaved work is preserved and authentication is required. | `assertive` |

Design contract. If an implementation cannot expose a live-region announcement for a required event, the corresponding accessibility criterion fails. Silent degradation is invalid.

## 15. Visual fixtures and implementation guidance

### 15.1 Visual fixture registry

Design contract. Visual fixtures are design-conformance evidence. They are not claim-bearing benchmark evidence unless Core 05 claim-publication requirements are separately satisfied.

Design contract. Every visual fixture MUST declare viewport, zoom, density, theme, scroll normalization, dynamic masks, and pass condition.

| Fixture ID | Surface | Required state | Viewport | Density | Scroll normalization | Dynamic masks | Pass condition |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `D-VFIX-001` | Shell | Base shell with default density and `dark_graphite`. | `{layout.baseViewport}` | `{density.default-mode}` | Top-left grid scroll. | Time, actor names, IDs. | Five shell regions visible. |
| `D-VFIX-002` | Timeline | Rough row with unresolved mention. | `{layout.baseViewport}` | `{density.default-mode}` | Named rough-row anchor. | IDs and timestamps. | Unresolved marker and raw capture visible. |
| `D-VFIX-003` | Timeline or Relationships | Resolved and auto-resolved chips. | `{layout.baseViewport}` | `{density.default-mode}` | Named chip row anchor. | Actor names and IDs. | Chip states distinguishable without color. |
| `D-VFIX-004` | Timeline or Relationships | Dismissed mention. | `{layout.baseViewport}` | `{density.default-mode}` | Named chip row anchor. | IDs. | Dismissed marker visible and low-emphasis. |
| `D-VFIX-005` | Any writable grid | Same-field conflict marker and resolver entry. | `{layout.baseViewport}` | `{density.default-mode}` | Conflict row anchor. | Actor names, timestamps, IDs. | Only affected cell marked; resolver entry visible. |
| `D-VFIX-006` | Status strip | `Syncing`, `Saved`, and `Conflict` states. | `{layout.baseViewport}` | `{density.default-mode}` | Strip-level crop. | None. | Exactly one primary label visible in each state. |
| `D-VFIX-007` | Shell/grid | Presence header, row, and cell states with overflow. | `{layout.baseViewport}` | `{density.default-mode}` | Presence row anchor. | Actor names/avatars unless seeded. | Capacity and `+N` behavior visible. |
| `D-VFIX-008` | Evidence | `requested`, `pending_upload_slot`, `available`, `preview_blocked`, and `inconsistent`. | `{layout.baseViewport}` | `{density.default-mode}` | Evidence row anchors. | IDs, timestamps. | Dimensions render independently and no fake attachment appears. |
| `D-VFIX-009` | Inspector | Details, Relationships, Evidence, and History sections. | `{layout.baseViewport}` | `{density.default-mode}` | Inspector open; selected row anchor. | Actor names, timestamps, IDs. | Section order matches §7.3. |
| `D-VFIX-010` | Shell | `narrow_desktop` inspector overlay. | Width band selected by §7.4 algorithm. | `{density.default-mode}` | Active row visible before overlay. | IDs. | Grid inert behind overlay; close control visible. |
| `D-VFIX-011` | Any surface | Empty successful query state. | `{layout.baseViewport}` | `{density.default-mode}` | Empty-state container anchor. | None. | Create-permitted or create-forbidden copy follows §12.6. |
| `D-VFIX-012` | Shell | Keyboard focus path across top bar, grid, inspector, and dialog. | `{layout.baseViewport}` | `{density.default-mode}` | Declared active focus sequence. | None. | Visible focus satisfies §14.1. |

Design contract. Dynamic timestamps, IDs, avatars, actor names, and cursor positions MUST be seeded deterministically or masked. A fixture without a deterministic pass condition is invalid.

### 15.2 Coding-agent and developer guidance

Design contract. Developers and coding agents MUST use this document's token registry as the first source for color, spacing, density, radius, elevation, and component visual-state decisions.

Design contract. Developers and coding agents MUST use canonical product terms from `domain.md` and stable identifiers from owner contracts when naming surfaces, states, tests, CSS variables, and component props.

Design contract. Grid styling MUST pass through the repo-local grid adapter and stable Cartulary wrapper classes, CSS variables, accessible names, and state attributes. Styling MUST NOT depend on generated vendor class-name internals.

Design contract. A design need MUST NOT create new tabs, route families, record types, lifecycle states, semantic statuses, or approval workflows. Such changes require owner-section changes or later adopted NLSpecs.

Design contract. Behavior MUST NOT be inferred from visible labels, row order, column labels, SQL names, projection names, component names, or style classes.

## 16. Boundaries, non-goals, future profiles, and external dependencies

### 16.1 Owner boundaries

Design contract. The owner boundary table below is binding for this document.

| Boundary | Owner |
| --- | --- |
| Product behavior, conformance, profile scope, and authority order | Core 00 through Core 04. |
| Claim-bearing timed or fixture-sensitive publication | Core 05. |
| Domain vocabulary and forbidden substitutions | `domain.md`, with owner sections governing behavior. |
| Public routes, wire shapes, view schemas, field keys, projection contracts, job contracts, storage boundaries | Core 01. |
| Record model, history, mentions, indicators, evidence, parties, task requests, decisions, artifacts, closed vocabularies | Core 02. |
| Workbook interactions, collaboration, save state, presence, conflict resolution, evidence workflow, workflows | Core 03. |
| Authentication, authorization, trust boundaries, deployment, evidence security, export release boundaries | Core 04. |
| Repo-local package boundaries, grid adapter, generated contracts, developer workflow | Development, bootstrap, testing, and frontend guides as subordinate implementation-support artifacts. |
| Future screenshots, Figma files, or mockups | Illustrative examples only unless later adopted as governing design artifacts. |

### 16.2 Current non-goals

Non-goal. The items in this table are intentionally outside this revision. Omission behavior is binding.

| Item | Current behavior | Omission semantics |
| --- | --- | --- |
| Light theme | Not supported. | No theme switcher entry; omission is conformant. |
| Dedicated high-contrast theme | Not supported as a separate theme. | Omission is conformant; required `dark_graphite` accessibility criteria still apply. |
| Mobile/touch-specific design | Not supported. | Below-minimum viewport behavior follows §7.4 and does not claim design conformance. |
| Report/export visual design | Not owned by this design contract. | Snapshot/reporting UI requires a separate design artifact or future section. |
| External visual reference board | Non-authoritative. | Inspiration only; cannot override token, state, or surface contracts. |
| All-surfaces-as-primary-tabs shell | Rejected for this revision. | Required system views remain in `System views`; built-in tabs remain primary at base viewport. |
| Command-palette-only system-view access | Rejected for this revision. | Required system views MUST be reachable from the shell. |

### 16.3 Future-profile candidates

Non-goal. Future-profile candidates MUST NOT be claimed as current behavior from this document.

| Candidate | Required future closure before support |
| --- | --- |
| Light theme | Complete token registry, contrast matrix, visual fixtures, switching persistence, and fallback behavior. |
| Dedicated high-contrast theme | Complete token registry, forced-color behavior, assistive-technology fixture set, and contrast criteria. |
| Mobile/touch profile | Breakpoints, gestures, target sizes, keyboard alternatives, inspector behavior, and fixtures. |
| Snapshot/reporting UI | Release states, redaction state, approval artifacts, rendered-output status, and fixture rules. |
| Expanded icon package authority | Exact package, license posture, semantic icon ID registry, and fixture coverage. |

### 16.4 Blocking design decisions

Design contract. There are no blocking design decisions in this accepted revision. If a future edit adds a blocking decision, the document status MUST change to draft-blocked and design-conformance claims MUST fail until the blocking decision is resolved or moved to non-goal/future scope with omission semantics.

## 17. Do and do-not rules

### 17.1 Do

Design contract. Implementations and design revisions MUST preserve these rules:

- Make the grid the protagonist of the workspace.
- Keep `{colors.accent}` scarce, intentional, and accessible.
- Encode state with text or shape in addition to color.
- Keep conflict, evidence, and resolution feedback local to the affected row, cell, chip, or inspector section.
- Use compact hairlines and neutral surfaces to preserve density.
- Keep system views reachable without leaving the workbook shell.
- Show the user's current save/conflict state in the status strip.
- Keep raw capture inspectable after resolution or normalization.

### 17.2 Do not

Design contract. Implementations and design revisions MUST NOT do any of the following:

- Use the accent as the warning system.
- Turn routine row creation into a modal, wizard, approval, challenge ritual, or full-page flow.
- Make all required surfaces primary tabs.
- Hide required system views behind command-palette-only discovery.
- Create dashboard modules for coordination surfaces.
- Use color-only chips, color-only conflict markers, or color-only evidence state.
- Use raw object-store URLs, SQL table names, projection names, or React component names as user-facing design concepts.
- Add neon threat maps, cinematic gradients, heavy glow, or decorative risk heatmaps to the live workbook shell.

## 18. Acceptance criteria

Design contract. A reviewer can treat this `design.md` as ready to guide design implementation only when every criterion below passes.

### 18.1 Document-structure criteria

| ID | Criterion |
| --- | --- |
| D-AC-001 | The document states that it is a design-direction contract and does not define Base Profile or extension-profile implementation conformance. |
| D-AC-002 | The document states that Core 05 is publication-only for claim-bearing timed or fixture-sensitive criteria. |
| D-AC-003 | No accepted normative section contains unresolved placeholder text. |
| D-AC-004 | Every normative paragraph uses a valid statement class or appears in a table introduced by a valid statement class. |
| D-AC-005 | Every `MAY` clause states omission behavior in the same paragraph, table row, or immediately following sentence. |
| D-AC-006 | Every table declared exhaustive contains no duplicate case key and no omitted case in its declared scope. |

### 18.2 Token and theme criteria

| ID | Criterion |
| --- | --- |
| D-AC-010 | The token registry validates against `cartulary.design_tokens.v1`. |
| D-AC-011 | All token references resolve without cycles. |
| D-AC-012 | Body text does not duplicate canonical token literal values except in explicitly non-normative examples, grammar definitions, or validation criteria. |
| D-AC-013 | The theme registry names `dark_graphite` as the only required theme and gives explicit omission semantics for `light` and `high_contrast`. |
| D-AC-014 | The iconography profile has no unresolved icon-family blocker and defines substitution, size, fallback, accessibility, and mixing rules. |

### 18.3 Surface, layout, and responsive criteria

| ID | Criterion |
| --- | --- |
| D-AC-020 | The shell-exposure registry contains all fourteen required base-profile surfaces exactly once. |
| D-AC-021 | The shell-exposure registry contains the three standardized optional workbook surfaces exactly once with omission semantics. |
| D-AC-022 | Required system views are reachable by keyboard and pointer from the shell. |
| D-AC-023 | Each viewport width and height combination falls into exactly one responsive band. |
| D-AC-024 | Each responsive band defines shell-region visibility and inspector mode. |
| D-AC-025 | Below-minimum behavior is explicitly non-conformant or degraded with safe navigation preserved. |

### 18.4 State and interaction criteria

| ID | Criterion |
| --- | --- |
| D-AC-030 | Save-state selection returns exactly one of `Syncing`, `Saved`, or `Conflict` for every input combination. |
| D-AC-031 | Cell-state precedence produces one deterministic visual primary state plus allowed co-displays. |
| D-AC-032 | Chip-state selection is deterministic for unresolved, resolved, auto-resolved, and dismissed cases. |
| D-AC-033 | Evidence lifecycle, upload-slot overlay, and preview capability render independently and deterministically. |
| D-AC-034 | Presence ordering and overflow are deterministic for equal timestamps and duplicate users. |
| D-AC-035 | View bar control order, chip order, and saved-view dirty state are deterministic. |
| D-AC-036 | `Esc` behavior follows the priority ladder. |
| D-AC-037 | Focus restoration follows the fallback ladder when the invoking element no longer exists. |

### 18.5 Component and surface criteria

| ID | Criterion |
| --- | --- |
| D-AC-040 | Every component family in §12.1 has a closed variant set. |
| D-AC-041 | Disabled and loading states are defined for action components. |
| D-AC-042 | Destructive actions never use accent fill. |
| D-AC-043 | Every required surface has one row in the per-surface design table. |
| D-AC-044 | Optional surfaces have omission behavior. |
| D-AC-045 | The per-surface design table does not define Core-owned field membership, write rules, or route behavior. |
| D-AC-046 | Empty-state behavior is defined for create-permitted and create-forbidden cases. |

### 18.6 Accessibility criteria

| ID | Criterion |
| --- | --- |
| D-AC-050 | Every required UI state has a non-color cue and accessible representation. |
| D-AC-051 | Icon-only controls have accessible names and visible focus. |
| D-AC-052 | Reduced-motion behavior is defined and testable. |
| D-AC-053 | The live-region matrix maps each listed event to `polite`, `assertive`, or no live announcement. |
| D-AC-054 | Row accessible names include human-readable surface context and are not raw `record_id` alone. |

### 18.7 Visual-fixture criteria

| ID | Criterion |
| --- | --- |
| D-AC-060 | Every `D-VFIX-*` fixture declares viewport, zoom, density, theme, scroll normalization, dynamic masks, and pass condition. |
| D-AC-061 | Dynamic fixture data is seeded or masked. |
| D-AC-062 | Fixture evidence is classified as design evidence, not claim-bearing benchmark evidence. |

### 18.8 Boundary criteria

| ID | Criterion |
| --- | --- |
| D-AC-070 | Each non-goal has observable omission behavior. |
| D-AC-071 | Future-profile candidates do not define current behavior. |
| D-AC-072 | The document forbids behavior inference from labels, row order, SQL names, projection names, vendor grid coordinates, component names, and styling classes. |
| D-AC-073 | The document can be committed as `design.md` without requiring screenshots, external reference boards, proprietary font files, or unstated implementation conventions. |