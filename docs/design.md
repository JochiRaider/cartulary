---
version: 0.3.0
name: Cartulary
document_class: design-direction-contract
status: adopted/closed-design-contract
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
    fontFamily: "Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, Segoe UI, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: 0
  grid:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.35
    letterSpacing: 0
  surface-title:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 18px
    fontWeight: 600
    lineHeight: 1.25
    letterSpacing: 0
  section-heading:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 16px
    fontWeight: 600
    lineHeight: 1.30
    letterSpacing: 0
  grid-cell:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.35
    letterSpacing: 0
  metadata:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 12px
    fontWeight: 500
    lineHeight: 1.30
    letterSpacing: 0
  button:
    fontFamily: "Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 13px
    fontWeight: 600
    lineHeight: 1.20
    letterSpacing: 0
  mono:
    fontFamily: "JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.40
    letterSpacing: 0
  alternate-ui:
    fontFamily: "Geist, Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.45
    letterSpacing: 0
  alternate-mono:
    fontFamily: "Geist Mono, JetBrains Mono, ui-monospace, SFMono-Regular, Menlo, Consolas, monospace"
    fontSize: 12px
    fontWeight: 400
    lineHeight: 1.40
    letterSpacing: 0
  report-narrative:
    fontFamily: "Source Serif 4, Georgia, Times New Roman, serif"
    fontSize: 16px
    fontWeight: 400
    lineHeight: 1.55
    letterSpacing: 0
  accessible-reading:
    fontFamily: "Atkinson Hyperlegible, Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 14px
    fontWeight: 400
    lineHeight: 1.50
    letterSpacing: 0
  compact-metadata:
    fontFamily: "IBM Plex Sans Condensed, Inter, ui-sans-serif, system-ui, sans-serif"
    fontSize: 12px
    fontWeight: 600
    lineHeight: 1.30
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
  icon-inline:
    size: 16px
  icon-toolbar:
    size: 20px
  focus-ring:
    border: "{border.focus}"
    offset: 2px
---

# Cartulary Design Direction

## 1. Status, authority, and scope

Design contract. This document is a design-direction contract for the Cartulary browser application.

Design contract. This document governs only observable UI design behavior in these families: visual language, token interpretation, theme scope, shell composition, responsive presentation, component visual states, local semantic UI states, accessibility presentation, visual fixture expectations, and coding-agent design guardrails.

Design contract. This document MUST NOT define Base Profile or extension-profile implementation conformance. It MUST NOT create routes, public schemas, authorization rules, evidence access semantics, lifecycle transitions, benchmark claims, storage boundaries, extension-profile behavior, record types, field registries, or public wire shapes.

Design contract. Core 00 through Core 04 govern current-profile implementation behavior. Core 05 governs claim-bearing timed or fixture-sensitive publication only. If this document and an owner section differ on product behavior, the owner section governs and this document MUST be repaired.

Design contract. This document is NLSpec-grade only inside its design-direction scope. It fully specifies the design behavior it owns, but it is not the Base Profile implementation-conformance corpus.

Design contract. `domain.md` governs repository vocabulary interpretation. This document uses domain terms for UI design only and MUST NOT redefine `party`, `artifact`, `view schema`, `saved view`, `system view`, `entity mention`, `object blob`, `workbook surface`, or any other domain concept.

## 2. Normative language and statement classes

Design contract. The key words **MUST**, **MUST NOT**, and **MAY** are normative inside this design-direction contract. **MUST** and **MUST NOT** define design-direction requirements. **MAY** defines optional design behavior only when omission behavior is specified in the same paragraph, table row, or immediately following sentence. `default` defines the required behavior when no explicit caller, user, deployment, or owner-controlled override applies.

Design contract. This revision uses no advisory normative terms. A later revision is valid only if it defines the effect of any newly introduced advisory normative term in this section.

Design contract. Accepted normative sections MUST NOT contain unresolved placeholder text. Future work MUST be classified as a non-goal, future profile candidate, or external owner dependency with explicit omission behavior.

| Statement class | Meaning | Normative-language rule |
| --- | --- | --- |
| `Design contract.` | Binding design-direction requirement owned by this document. | Uses **MUST**, **MUST NOT**, **MAY**, and `default` only as defined in §2. |
| `Core restatement.` | Non-authoritative summary of Core-owned behavior. | MUST use the syntax `Owner: <artifact> §<section>[, <REQ-* where applicable>]` and MUST NOT issue independent design-owned product behavior. |
| `Rationale.` | Explanation only. | MUST NOT contain normative keywords. |
| `Non-goal.` | Explicit omission boundary. | MUST state observable omission behavior. |
| `External dependency.` | Decision owned outside this document. | MUST name the owner or mark the revision blocked. |

Design contract. A paragraph without one of the statement-class prefixes above is non-normative only when it appears in a heading, code fence, table caption, acceptance criterion, or source note. Normative prose in all other locations MUST use one statement-class prefix.

### 2.1 Normative wording lint

Design contract. Accepted normative prose MUST fail design-document validation when it contains any pattern in the table below outside `Rationale.`, `Non-goal.`, `External dependency.`, a quoted source excerpt, a non-normative example, a grammar definition, or an acceptance criterion.

| Invalid wording pattern in normative prose | Required repair |
| --- | --- |
| `can`, `could`, `should`, `as needed`, `as appropriate`, `when available`, `when space permits`, `equivalent`, `allowed` | Replace with **MUST**, **MUST NOT**, **MAY** plus omission behavior, a deterministic algorithm, or non-normative rationale. |
| Lowercase `may` used as permission | Replace with uppercase **MAY** plus omission behavior. |
| `when available` | Replace with the exact data-presence precondition. |
| `when space permits` | Replace with a deterministic capacity or overflow rule. |
| `equivalent` | Replace with a named equivalence table or remove. |

Design contract. A `Core restatement.` whose owner cannot be located MUST be reclassified as `Rationale.` or `External dependency.` with `owner_lookup_required`; it MUST NOT remain a `Core restatement.`.

## 3. Token registry, token resolution, and theme contract

### 3.1 Front matter partition

Design contract. The YAML front matter is a single serialized object with two validation regions: document metadata and token registry. The regions MUST be disjoint.

| Front-matter region | Keys in this revision | Unknown-key behavior | Validation role |
| --- | --- | --- | --- |
| Document metadata | `version`, `name`, `document_class`, `status`, `design_contract_schema_id`, `token_registry_schema_id`, `default_theme_id`, `description` | Unknown metadata keys are invalid unless a later `cartulary.design_direction.*` revision defines them. | Parsed before token validation; metadata keys are not token IDs. |
| Token registry | `colors`, `typography`, `spacing`, `density`, `rounded`, `border`, `elevation`, `motion`, `layout`, `components` | Unknown token namespaces are invalid. | The only source for token IDs and token references. |

Design contract. `description` MAY contain embedded token references. Omission behavior: if a later revision omits `description`, no description token-resolution check applies.

Design contract. Token references in document metadata MUST be resolved for validation only and MUST NOT create token IDs. Token references inside token values MUST be resolved as token values.

Design contract. Duplicate YAML keys at any object level are invalid before token-reference resolution. Key order is not semantically significant. Canonical validation output MUST sort registry paths lexically by namespace and then token path.

### 3.2 Token registry schema contract

Design contract. The token registry schema ID is `cartulary.design_tokens.v1`. A conforming token registry MUST satisfy every row in the table below.

| Schema concern | Required contract |
| --- | --- |
| Serialization | YAML front matter parsed as a UTF-8 mapping. Duplicate keys are invalid. |
| Scalar types | Validation distinguishes string, integer, number, and object values. Browser CSS parsing is not the schema validator. |
| Required namespaces | `colors`, `typography`, `spacing`, `density`, `rounded`, `border`, `elevation`, `motion`, `layout`, and `components`. |
| Optional namespaces | None in this revision. |
| Unknown top-level token namespace | Invalid. |
| Unknown nested key | Invalid unless the owning namespace defines an open extension point. This revision defines no open extension point. |
| Token path canonical form | `namespace.path.segment`, preserving exact case, hyphenation, and punctuation from the registry. |
| Token-reference validation order | Parse front matter, validate schema, validate token path uniqueness, resolve references, validate resolved scalar grammar, then validate body token usage. |
| Failure result | Design-document validation fails; rendered design-conformance evidence MUST NOT be accepted. |

Design contract. Validation failure classes are closed to the table below.

| Failure class | Trigger |
| --- | --- |
| `invalid_frontmatter_schema` | Front matter is not a UTF-8 mapping or required metadata is malformed. |
| `duplicate_frontmatter_key` | Any mapping level contains a duplicate key after exact key comparison. |
| `unknown_token_namespace` | A top-level token namespace is outside §3.1. |
| `unknown_token_key` | A token reference names a missing path. |
| `invalid_token_value` | A value fails its namespace or scalar grammar. |
| `token_reference_cycle` | Token resolution revisits the same token reference in one stack. |
| `invalid_body_token_usage` | Body text refers to a token-owned design value without token ID or design-literal registration. |

Design contract. A conforming token registry MUST contain only the top-level namespaces in the table below.

| Namespace | Required value shape | Notes |
| --- | --- | --- |
| `colors` | Hex color, `rgba()` string, or token reference resolving to a color. | Utility tokens such as `inverse-canvas` and `inverse-ink` are not alternate themes. |
| `typography` | Object with `fontFamily`, `fontSize`, `fontWeight`, `lineHeight`, and `letterSpacing`. | `lineHeight` MAY be unitless; omission behavior is invalid because every typography token MUST declare it. `letterSpacing` MAY be `0` or a CSS pixel length; omission behavior is invalid. |
| `spacing` | CSS pixel length. | Negative spacing is invalid. |
| `density` | Object with `rowHeight` and `cellPadding`, plus `default-mode`. | Valid modes are `compact`, `default`, and `comfortable`. |
| `rounded` | CSS pixel length. | Negative radius is invalid. `9999px` is an ordinary valid pixel length used for pill geometry. |
| `border` | Border shorthand with token references. | Border color MUST resolve to a `colors` token. |
| `elevation` | `none` or box-shadow string. | Elevation MUST NOT encode semantic state. |
| `motion` | Duration or cubic-bezier string. | Durations MUST be non-negative. |
| `layout` | Declared viewport descriptor, percent string, CSS pixel length, or closed CSS-expression subset. | Breakpoint values are design-owned presentation bounds. |
| `components` | Object composed from primitive tokens and declared component literal values. | Component tokens MUST NOT redefine primitive token semantics. |

### 3.3 Token scalar grammar

Design contract. Every CSS-like scalar token MUST validate under exactly one scalar grammar in the table below.

| `value_family_id` | Accepted lexical form | Numeric bounds | Whitespace | Quoting | Canonical serialization |
| --- | --- | --- | --- | --- | --- |
| `hex_color_v1` | `#` followed by exactly six uppercase hexadecimal digits. | Channels are `00` through `FF`. | No internal whitespace. | YAML string quoting optional. | Uppercase `#RRGGBB`. |
| `rgba_color_v1` | `rgba(<r>, <g>, <b>, <a>)`. | `r`, `g`, `b` integers `0..255`; `a` number `0..1` inclusive. | Exactly one space after each comma. | YAML string quoting required. | `rgba(r, g, b, a)` with no leading zero normalization except `0` and `1`. |
| `css_px_length_v1` | Non-negative integer or decimal followed by `px`. | `0..9999`; negative values invalid. | No whitespace between number and `px`. | YAML string quoting optional unless the YAML parser would change type. | Decimal form preserved except unnecessary trailing `.0` removed. |
| `css_padding_2d_v1` | Two `css_px_length_v1` values separated by one ASCII space. | Each side follows `css_px_length_v1`. | Exactly one separator space. | YAML string quoting required. | `<vertical> <horizontal>`. |
| `css_percent_v1` | Non-negative integer followed by `%`. | `0..100`. | No whitespace. | YAML string quoting required. | `<integer>%`. |
| `css_duration_ms_v1` | Non-negative integer followed by `ms`. | `0..60000`. | No whitespace. | YAML string quoting optional. | `<integer>ms`. |
| `cubic_bezier_v1` | `cubic-bezier(<x1>, <y1>, <x2>, <y2>)`. | `x1` and `x2` in `0..1`; `y1` and `y2` in `-2..2`. | Exactly one space after each comma. | YAML string quoting required. | Preserve decimal values without trailing zeros. |
| `border_shorthand_v1` | `<width> <style> <color-token-ref>`. | Width follows `css_px_length_v1`; style closed to `solid`, `dashed`, `dotted`. | Exactly one space between members. | YAML string quoting required. | `<width> <style> {colors.*}`. |
| `box_shadow_v1` | `none` or `<x> <y> <blur> rgba(...)`. | `x` and `y` allow `-9999px..9999px`; `blur` non-negative `0..9999px`. | Exactly one space between components. | YAML string quoting required for shadow strings. | `none` or exact normalized members. |
| `viewport_descriptor_v1` | `<width>x<height> CSS px`. | Width and height positive integers `1..99999`. | One ASCII space before `CSS` and before `px`. | YAML string quoting required. | `<width>x<height> CSS px`. |
| `css_min_px_vw_v1` | `min(<px-length>, <vw>)`. | `px` follows `css_px_length_v1`; `vw` integer `1..100`. | Exactly one space after comma. | YAML string quoting required. | `min(<px>, <vw>)`. |

Design contract. This revision accepts no `em`, `rem`, `vh`, `calc()`, named colors, lowercase hex, three-digit hex, four-digit hex, eight-digit hex, gradient, CSS variable, CSS identifier escape, or browser-specific CSS function in token values.

### 3.4 Token ID grammar

Design contract. Token references MUST use this grammar:

```text
token_ref = "{" namespace "." token_key "}"
namespace = one of the closed top-level token namespaces in §3.1
token_key = declared key path inside that namespace, preserving exact registry spelling and using "." between nested segments
```

Design contract. The canonical token reference for a top-level primitive token is `{namespace.token-key}`. The canonical token reference for a nested token is `{namespace.parent-key.child-key}`.

### 3.5 Token resolution algorithm

Design contract. Token resolution MUST be deterministic and MUST use this algorithm:

```pseudocode
resolve_token(value, stack):
  if value is an exact token_ref:
    target = lookup(value.namespace, value.token_key)
    if target does not exist: return unknown_token_key
    if value is already in stack: return token_reference_cycle
    return resolve_token(target, stack + value)

  if value is a string containing one or more embedded token_ref substrings:
    output = value
    for each token_ref in lexical order of appearance in value:
      replacement = resolve_token(token_ref, stack)
      if replacement is an error: return replacement
      output = output with that exact substring replaced by replacement
    return output

  return value
```

Design contract. Token validation MUST fail before design-conformance evidence is accepted when any invalid state in the table below occurs.

| Invalid state | Required result |
| --- | --- |
| Unknown top-level namespace | Token validation fails with `unknown_token_namespace`. |
| Unknown token key | Token validation fails with `unknown_token_key`. |
| Token-reference cycle | Token validation fails with `token_reference_cycle`. |
| Unsupported unit for namespace value shape | Token validation fails with `invalid_token_value`. |
| Duplicate token after exact registry-path normalization | Token validation fails with `duplicate_frontmatter_key`. |
| Component token references undeclared primitive token | Token validation fails with `unknown_token_key`. |
| `density.default-mode` does not equal one declared density mode | Token validation fails with `invalid_token_value`. |
| Body token-use rule violation | Design-document validation fails with `invalid_body_token_usage`. |

### 3.6 Body token-use and design literal registry

Design contract. Body text MUST reference token IDs when it refers to a token-owned design value.

Design contract. Body text MAY contain scalar literals when the literal defines a grammar, declares a non-token design constant, gives a non-normative example, or states a validation criterion. Omission behavior: no token reference is required for those exempt scalar uses.

Design contract. A scalar literal that controls design output and appears in normative prose or a normative table MUST either be a token reference or appear in the design literal registry below.

| Literal ID | Literal value | Owning section | Purpose | Tokenization decision |
| --- | --- | --- | --- | --- |
| `literal.integer.zero` | `0` | Multiple tables | Count, index, and ratio boundary. | Not a visual token. |
| `literal.integer.one` | `1` | Multiple algorithms | Priority, count, and deterministic step boundary. | Not a visual token. |
| `literal.viewport.base.visual-fixture` | `1280x720 CSS px` | §15 | Fixture baseline viewport. | Tokenized as `{layout.baseViewport}` for design use; literal remains valid in fixture rows for exact artifact identity. |
| `literal.icon.inline.size` | `16px` | §3.10 | Dense inline icon size. | Tokenized as `{components.icon-inline.size}`. |
| `literal.icon.toolbar.size` | `20px` | §3.10 | Toolbar icon size. | Tokenized as `{components.icon-toolbar.size}`. |

### 3.7 CSS variable interface

Design contract. If an implementation exposes design tokens as CSS custom properties, it MUST use this mapping:

```text
--ct-<namespace>-<token-key>
```

Design contract. For nested component tokens, the CSS custom property mapping MUST be:

```text
--ct-component-<component-key>-<property-key>
```

Design contract. If an implementation does not expose public CSS variables, design conformance MUST be verified through rendered visual and accessibility fixture evidence rather than variable-name inspection.

### 3.8 Theme registry

Design contract. The current theme registry is closed to the rows in this table.

| Theme ID | Required | Exposure | Omission behavior |
| --- | ---: | --- | --- |
| `dark_graphite` | Yes | Default and only required theme. | Not applicable. |
| `light` | No | MUST NOT be exposed as a supported theme in this revision. | Omission is conformant. |
| `high_contrast` | No | MUST NOT be claimed as supported unless a later design revision defines complete tokens, contrast targets, and fixtures. | Omission is conformant, but `dark_graphite` still MUST satisfy §14 accessibility criteria. |

Design contract. A theme switcher MUST NOT be exposed unless every selectable theme has a complete token registry, accessibility matrix, and visual fixture coverage.

Design contract. Browser or operating-system forced-color modes MAY alter rendering outside this document's dedicated theme registry. Omission behavior: this revision does not claim a dedicated forced-color conformance profile; omission of a dedicated forced-color theme is conformant when required keyboard, non-color state, and contrast criteria remain satisfied for `dark_graphite`.

### 3.9 Density registry

Design contract. Density selection is closed to `compact`, `default`, and `comfortable`. The default density is `{density.default-mode}`.

Design contract. Workbook surfaces MUST use density modes from the shared density registry. The default workbook density remains `{density.default-mode}`, except the default Timeline grid uses `compact` density from the same shared tokens to preserve first-viewport incident-response scanning. User-selected density is valid as client or user preference only when it maps to a declared shared density mode; surfaces MUST NOT invent private row-height or padding systems.

Design contract. Large incident grids use fixed-height rows by default. Variable-height rows are valid only in inspector sections, preview areas, or non-grid detail regions. Omission of variable-height grid rows is conformant.

### 3.10 Iconography profile

Design contract. A conforming implementation MUST use exactly one icon family in the workbook shell.

Design contract. This revision uses a bounded substitution profile instead of naming a package. The selected family MUST satisfy every constraint below.

| Constraint | Requirement |
| --- | --- |
| Style | Outline-only, single-stroke icon family. |
| Stroke weight | One default stroke weight across the shell, except optical corrections inside the selected family. |
| Size grid | `{components.icon-inline.size}` for dense inline affordances and `{components.icon-toolbar.size}` for toolbar controls; no other shell icon size unless a component table declares it. |
| Filled icons | Forbidden unless a later revision enumerates a specific semantic exception. |
| Family mixing | Forbidden inside one rendered shell. |
| Semantic IDs | Icons MUST be addressed by stable semantic icon IDs in application code or design fixtures, not by visible labels. |
| Missing icon | Render the text label or a fallback glyph with the same accessible name. Silent omission is invalid. |
| Icon-only control | MUST have an accessible name and visible focus indicator. |

Design contract. Icons MAY support labels. Omission behavior: when a semantic registry row declares `label_required`, the label MUST render; when a row declares `label_optional`, omission of the label is conformant only if the icon has the declared accessible name; when a row declares `icon_only_allowed`, omission of visible label is conformant only with visible focus and accessible name.

Design contract. Icons MUST NOT replace labels for destructive actions, conflict resolution, evidence preview, evidence download, rollback, merge initiation, party link/unlink, inspector close, inspector pin, or system-view switching.

### 3.11 Semantic icon registry

Design contract. Every icon rendered for a registered meaning MUST use one semantic icon ID from this registry. Package-specific icon names MUST NOT appear in design fixtures or application-level assertions.

| `semantic_icon_id` | Required contexts | Required pairing | Accessible name | Fallback | Omission behavior | Fixture coverage |
| --- | --- | --- | --- | --- | --- | --- |
| `surface_switcher` | Top bar | `label_required` | `Surfaces` | Text `Surfaces` | Icon omission conformant when label remains. | `D-VFIX-001`, `D-VFIX-010` |
| `system_views` | Top bar | `label_required` | `System views` | Text `System views` | Icon omission conformant when label remains. | `D-VFIX-001`, `D-VFIX-010` |
| `saved_view` | View bar | `label_optional` | `Saved view` | Text `View` | Icon omission conformant. | `D-VFIX-001` |
| `sort` | Top bar | `label_optional` | `Sort` | Text `Sort` | Icon omission conformant. | `D-VFIX-001` |
| `group` | Top bar | `label_optional` | `Group` | Text `Group` | Icon omission conformant. | `D-VFIX-001` |
| `filter` | Top bar | `label_optional` | `Filter` | Text `Filter` | Icon omission conformant. | `D-VFIX-001` |
| `filter_overflow` | Top bar | `label_required` | `Filters, <N> hidden` | Text `Filters` | Icon omission conformant when label and hidden count remain. | `D-VFIX-010`, `D-VFIX-011` |
| `inspector_open` | Grid or row action | `label_optional` | `Open inspector` | Text `Inspect` | Icon omission conformant. | `D-VFIX-001` |
| `inspector_close` | Inspector | `label_required` | `Close inspector` | Text `Close inspector` | Icon omission conformant when label remains. | `D-VFIX-002` |
| `inspector_pin` | Inspector | `label_required` | `Pin inspector` | Text `Pin inspector` | Icon omission conformant when label remains. | `D-VFIX-002` |
| `inspector_unpin` | Inspector | `label_required` | `Unpin inspector` | Text `Unpin inspector` | Icon omission conformant when label remains. | `D-VFIX-002` |
| `evidence_preview` | Evidence row, inspector | `label_required` | `Preview evidence` | Text `Preview evidence` | Icon omission conformant when label remains. | `D-VFIX-006` |
| `evidence_download` | Evidence row, inspector | `label_required` | `Download evidence` | Text `Download evidence` | Icon omission conformant when label remains. | `D-VFIX-006` |
| `evidence_attach` | Row or inspector | `label_required` | `Attach evidence` | Text `Attach evidence` | Icon omission conformant when label remains. | `D-VFIX-006` |
| `evidence_blocked` | Evidence row, inspector | `label_optional` | `Evidence access blocked` | Text `Blocked` | Icon omission conformant when text remains. | `D-VFIX-006` |
| `conflict` | Cell, status strip | `label_optional` | `Conflict` | Text `Conflict` | Icon omission conformant when conflict marker remains. | `D-VFIX-003`, `D-VFIX-007` |
| `resolve_conflict` | Conflict resolver | `label_required` | `Resolve conflict` | Text `Resolve conflict` | Icon omission conformant when label remains. | `D-VFIX-003` |
| `rollback` | Inspector History | `label_required` | `Rollback` | Text `Rollback` | Icon omission conformant when label remains. | `D-VFIX-008` |
| `merge` | Inspector Relationships | `label_required` | `Merge` | Text `Merge` | Icon omission conformant when label remains. | `D-VFIX-008` |
| `delete` | Row action, inspector | `label_required` | `Delete` | Text `Delete` | Icon omission conformant when label remains. | `D-VFIX-008` |
| `party_link` | Party relationship control | `label_required` | `Link party` | Text `Link party` | Icon omission conformant when label remains. | `D-VFIX-004` |
| `party_unlink` | Party relationship control | `label_required` | `Unlink party` | Text `Unlink party` | Icon omission conformant when label remains. | `D-VFIX-004` |
| `history` | Inspector History | `label_optional` | `History` | Text `History` | Icon omission conformant. | `D-VFIX-008` |
| `presence_editing` | Presence | `label_optional` | `<display name> editing` | Text `editing` | Icon omission conformant when presence label remains. | `D-VFIX-005` |
| `presence_focused` | Presence | `label_optional` | `<display name> focused` | Text `focused` | Icon omission conformant when presence label remains. | `D-VFIX-005` |
| `presence_viewing` | Presence | `label_optional` | `<display name> viewing` | Text `viewing` | Icon omission conformant when presence label remains. | `D-VFIX-005` |
| `presence_idle` | Presence | `label_optional` | `<display name> idle` | Text `idle` | Icon omission conformant when presence label remains. | `D-VFIX-005` |
| `unresolved` | Chip | `label_optional` | `Unresolved mention` | Text `?` | Icon omission conformant when unresolved marker remains. | `D-VFIX-004` |
| `auto_resolved` | Chip | `label_optional` | `Auto-resolved mention` | Text `auto` | Icon omission conformant when auto marker remains. | `D-VFIX-004` |
| `dismissed` | Chip | `label_optional` | `Dismissed mention` | Text `dismissed` | Icon omission conformant when dismissed marker remains. | `D-VFIX-004` |
| `pending` | Status or cell | `label_optional` | `Pending` | Text `Pending` | Icon omission conformant when pending label remains. | `D-VFIX-007` |
| `success` | Status or toast | `label_optional` | `Success` | Text `Success` | Icon omission conformant when success text remains. | `D-VFIX-007` |
| `warning` | Status or inline message | `label_optional` | `Warning` | Text `Warning` | Icon omission conformant when warning text remains. | `D-VFIX-007` |
| `error` | Status or inline message | `label_optional` | `Error` | Text `Error` | Icon omission conformant when error text remains. | `D-VFIX-007` |

## 4. Product and interaction contract

### 4.1 Product thesis

Design contract. Cartulary is a workbook-native incident workspace. The design MUST feel like a serious low-friction workbook on the hot path while behaving like a disciplined case system underneath.

Design contract. The primary interaction object is the visible workbook row, cell, chip, count, preview affordance, filter state, grouping state, and status state. The storage model, route handlers, projection tables, and grid vendor coordinates MUST NOT be exposed as user-facing design concepts.

Core restatement. Cartulary preserves the spreadsheet mental model at the view layer while keeping source data relational and auditable underneath. Owner: `03_workbook_interaction_collaboration_and_workflows.md` §1, REQ-03-001 through REQ-03-003; `domain.md` §3.

Design contract. The UI MUST preserve spreadsheet speed where it matters: direct typing, paste, compact scanning, keyboard navigation, visible rows, flexible filtering, and fast reshaping of the current working set.

Design contract. The UI MUST reject spreadsheet failure modes that damage incident work: row-position identity, silent overwrites, hidden relationship semantics, evidence paths as authority, unmanaged binary storage, and unversioned history.

Non-goal. Cartulary is not a dashboard, ticket queue, CRM, SIEM, EDR, evidence vault, long-form report editor, or command-center visualization. Omission behavior: the shell MUST NOT adopt those products' information architecture as the default workbook model.

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
| Paste or bulk-enter rows | Preserve the user's place, show interpretation and validation locally, and use staged feedback while slower work continues. |
| Resolve host, identity, or indicator references | Use inspector or same-surface enrichment; distinguish unresolved text, resolved chip, auto-resolved chip, and dismissed mention. |
| Attach or request evidence | Show requested, pending receipt, received, pending upload slot, available, failed upload, quarantined, blocked preview, and inconsistent states without implying fake attachment. |
| Preview or download evidence | Keep the workbook shell present; blocked preview remains inline and does not silently fall back to download. |
| Filter, sort, group, and save views | Treat the current view as a workbook working set; saved views remain configurations over one `view_schema_id`. |
| Resolve same-field conflict | Mark only the affected cell, keep the saved value visible, retain the local draft separately, and open a resolver from that cell. |
| Review history or rollback | Use row-local history and scoped destructive-action presentation in the inspector; when history is already open, it follows the active saved row and shows in-drawer loading or error state for that row. Routine editing MUST NOT feel like approval workflow. |
| Coordinate tasks, decisions, handoffs, status reviews, and lessons | Keep these as workbook-native surfaces, not separate task-management or workflow modules. |

## 5. Visual design principles

### 5.1 Dense graphite forensic workspace

Design contract. The default visual direction is a dark, dense, graphite workspace with warm operational emphasis. It MUST feel calm, precise, inspectable, and durable. It MUST NOT feel cyberpunk, theatrical, militarized, playful SaaS, or generic enterprise gray.

Design contract. The UI MUST use subtle surface elevation, crisp hairlines, stable row rhythm, and restrained contrast. Evidence, history, and collaboration state MUST remain inspectable without turning the workspace into a wall of alerts.

### 5.2 Accent scarcity

Design contract. `{colors.accent}` is the Cartulary accent. It MUST be used sparingly for focus rings, selected shell controls, primary affirmative action, active grid handles, and brand emphasis.

Design contract. `{colors.accent}` MUST NOT be used as the default warning color, row background, large panel fill, heatmap color, decorative gradient, or arbitrary chart color.

Design contract. Caution and warning semantics MUST use `{colors.semantic-caution}` or a semantic state token listed in §14.3 and MUST include text, shape, marker, or accessible name. Primary accent-filled buttons MUST use `{colors.on-accent}` for text or icons.

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

Design contract. The UI MUST make structure visible without demanding structure before capture. Users enter rough values first. Later normalization, resolution, review, and rollback become available through adjacent controls and stateful chips.

### 5.5 Accessibility as visual language

Design contract. Color MUST NOT be the only carrier of state. Every state-bearing color MUST be paired with at least one non-color cue: shape, marker text, accessible name, icon with accessible name, tooltip, live-region announcement, or a cue declared in §14.

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
| `{colors.semantic-conflict}` | Same-field conflict, unresolved conflict count, conflict resolver entry point. | Generic warning or destructive state. |
| `{colors.semantic-destructive}` | Delete, rollback, destructive confirmation, dangerous irreversible action. | Primary affirmative action. |
| `{colors.semantic-presence-self}` | Current user's presence marker. | Other users or semantic status. |
| `{colors.semantic-presence-other}` | Other users' presence marker. | Current user or semantic status. |

### 6.2 Typography

Design contract. Typography MUST use the registry roles below.

| Text role | Required token | Required use |
| --- | --- | --- |
| Shell and body UI | `{typography.ui}` | Default interface text. |
| Workbook grid | `{typography.grid}` | Workbook grid container text unless a specialized role applies. |
| Active surface title | `{typography.surface-title}` | Current surface and selected saved-view label. |
| Inspector section heading | `{typography.section-heading}` | Inspector sections, dialog headings, grouped menus. |
| Grid cells | `{typography.grid-cell}` | Workbook cell content and row-readable values. |
| Metadata and counters | `{typography.metadata}` | Badges, timestamps, row versions, counts, compact source labels. |
| Buttons and chips | `{typography.button}` | Button labels, chip labels, compact interactive controls. |
| Technical identifiers | `{typography.mono}` | Stable IDs, hashes, field keys, route-like diagnostic text when shown. |
| Alternate UI profile | `{typography.alternate-ui}` | Future optional alternate UI profile only; no runtime selector is exposed in this revision. |
| Alternate mono profile | `{typography.alternate-mono}` | Future optional alternate mono profile only; no runtime selector is exposed in this revision. |
| Narrative report prose | `{typography.report-narrative}` | Future report preview or export prose, not workbook hot-path text. |
| Accessibility reading profile | `{typography.accessible-reading}` | Explicit readable-content profile activated only by a reading-profile selector. |
| Compact metadata | `{typography.compact-metadata}` | Constrained metadata labels and dense status summaries. |

Design contract. `record_id`, `field_key`, route names, storage names, and hash text MUST NOT be the primary label for ordinary users. If shown, they MUST use `{typography.mono}` and appear as secondary technical metadata.

### 6.3 Motion

Design contract. Motion MUST use `{motion.duration-fast}` or `{motion.duration-normal}` for ordinary UI transitions. State feedback MUST remain legible when motion is disabled.

Design contract. Reduced-motion mode MUST cap non-essential animation at `{motion.duration-reduced-max}` and MUST remove looping spinners in favor of static pending markers plus text.

Design contract. Motion MUST NOT be required to perceive conflict, pending, saved, error, or evidence state.

## 7. Shell, surface composition, and responsive behavior

### 7.1 Shell regions

Design contract. The application shell MUST contain the regions in the table below at every supported viewport band.

| Region | Required contents | Boundary |
| --- | --- | --- |
| Top bar | Incident identity, built-in tabs or `Surfaces`, active-surface title when not already represented by a selected built-in tab, sort, group, filter controls, active chips or overflow controls, `System views`, presence summary when assigned by §7.5. | Persistent chrome, not a dashboard. |
| View bar | Saved-view selector, saved-view actions, inspector opener, and add-row control when allowed. | Belongs to active surface only. |
| Grid | Active workbook surface with `record_id`-bound rows and `field_key`-bound cells. | Primary work surface. |
| Inspector | Details, Relationships, Evidence, History, destructive and specialized row actions. | Conditional adjacent or overlay secondary surface opened through explicit controls. |
| Status strip | Save state, secondary same-surface message, presence summary or overflow when assigned by §7.5. | Capacity-limited working-state strip. |

Design contract. The default Timeline workbook shell at `{layout.baseViewport}` MUST show the top bar, compact sheet toolbar, active Timeline grid, explicit inspector opener, bottom draft row when creation is allowed, and status strip as the dominant first-viewport structure. The inspector MUST be closed by default and MUST open only through explicit controls such as the toolbar inspector control, row action menu, history action, mention action, or equivalent keyboard-accessible command. Incident summary, bootstrap defaults, membership management, promoted-field patch forms, or other administration/control surfaces MUST NOT dominate the default Timeline path above the active grid; those controls are valid only inside an explicitly opened secondary surface or a distinct administration context.

Design contract. A shell region MAY be visually collapsed only when the responsive algorithm in §7.5 assigns its controls to another reachable region. Omission behavior: a collapsed region with no assigned controls renders no visible container.

### 7.2 Shell-exposure surface registry

Core restatement. The base profile has fourteen required workbook surfaces and three standardized optional workbook surfaces. The canonical identity for each surface is the `sheet_ref` object `{ "kind": "view_schema", "id": <view_schema_id> }`. Owner: `01_architecture_storage_and_view_contracts.md` §7.4, REQ-01-307 and Table 7.4-A; `03_workbook_interaction_collaboration_and_workflows.md` §2, REQ-03-004 through REQ-03-011.

Design contract. The shell-exposure registry for this design revision is exhaustive for current-profile standardized workbook surfaces.

| Surface label | Canonical `view_schema_id` | Surface status | Primary exposure at base viewport | Switcher group | Group order | Row order | Saved-view exposure | Omission behavior |
| --- | --- | --- | --- | --- | ---: | ---: | --- | --- |
| Timeline | `cartulary.view.timeline.v1` | Required built-in sheet | Built-in tab | Built-in surfaces | 1 | 1 | Active surface view selector | Not optional. |
| Hosts | `cartulary.view.hosts.v1` | Required built-in sheet | Built-in tab | Built-in surfaces | 1 | 2 | Active surface view selector | Not optional. |
| Identities | `cartulary.view.identities.v1` | Required built-in sheet | Built-in tab | Built-in surfaces | 1 | 3 | Active surface view selector | Not optional. |
| Evidence | `cartulary.view.evidence.v1` | Required built-in sheet | Built-in tab | Built-in surfaces | 1 | 4 | Active surface view selector | Not optional. |
| Notes | `cartulary.view.notes.v1` | Required built-in sheet | Built-in tab | Built-in surfaces | 1 | 5 | Active surface view selector | Not optional. |
| Indicators | `cartulary.view.indicators.v1` | Required system view | `System views` | Scope and indicators | 2 | 1 | Active surface view selector | Not optional. |
| Compromise Assessments | `cartulary.view.assessments.v1` | Required system view | `System views` | Scope and indicators | 2 | 2 | Active surface view selector | Not optional. |
| Task Requests | `cartulary.view.task_requests.v1` | Required system view | `System views` | Coordination | 3 | 1 | Active surface view selector | Not optional. |
| Decisions | `cartulary.view.decisions.v1` | Required system view | `System views` | Coordination | 3 | 2 | Active surface view selector | Not optional. |
| Parties | `cartulary.view.parties.v1` | Required system view | `System views` | Coordination | 3 | 3 | Active surface view selector | Not optional. |
| Communications Log | `cartulary.view.comm_log.v1` | Required system view | `System views` | Coordination | 3 | 4 | Active surface view selector | Not optional. |
| Handoff | `cartulary.view.handoff.v1` | Required system view | `System views` | Coordination | 3 | 5 | Active surface view selector | Not optional. |
| Status Review | `cartulary.view.status_review.v1` | Required system view | `System views` | Review and learning | 4 | 1 | Active surface view selector | Not optional. |
| Lesson | `cartulary.view.lesson.v1` | Required system view | `System views` | Review and learning | 4 | 2 | Active surface view selector | Not optional. |
| Findings | `cartulary.view.findings.v1` | Standardized optional workbook surface | `System views` only if implemented and exposed | Optional artifact surfaces | 5 | 1 | Active surface view selector | If not implemented, it MUST NOT appear. |
| Investigative Queries | `cartulary.view.investigative_queries.v1` | Standardized optional workbook surface | `System views` only if implemented and exposed | Optional artifact surfaces | 5 | 2 | Active surface view selector | If not implemented, it MUST NOT appear. |
| Forensic Keywords | `cartulary.view.forensic_keywords.v1` | Standardized optional workbook surface | `System views` only if implemented and exposed | Optional artifact surfaces | 5 | 3 | Active surface view selector | If not implemented, it MUST NOT appear. |

Design contract. Required system views MUST NOT be command-palette-only. All required system views MUST be reachable by keyboard and pointer from the shell.

Design contract. For required system views, keyboard and pointer reachability means selection completes to the requested active workbook surface and renders that surface's grid. A visible menu option alone does not satisfy reachability.

Design contract. If a standardized optional surface is not implemented, it MUST NOT appear in the switcher. If a required surface is unavailable because of implementation defect or load failure, the UI MUST show an error state for that surface instead of silently removing it from the shell.

Design contract. Saved views MUST appear under the active surface's view selector. A saved view MUST NOT replace canonical system-view identity or create a new primary tab by default.

Design contract. Constrained UI labels MAY use `Assessments` as a display shorthand for `Compromise Assessments`. Omission behavior: the shorthand is conformant only when the accessible name or surrounding context exposes the full label.

### 7.3 Inspector contract

Design contract. The inspector MUST use the following default and bounds.

| Property | Required behavior |
| --- | --- |
| Default width | `{layout.inspectorDefaultWidth}`. |
| Minimum width | `{layout.inspectorMinWidth}`. |
| Maximum width | `{layout.inspectorMaxWidth}`. |
| Resize | User resizes within min/max bounds. Omission of resize persistence is conformant. |
| Pinning | Client-local only; MUST NOT persist in saved views. |
| Section order | Details, Relationships, Evidence, History. |
| Grid visibility | At base viewport, grid remains visible whenever inspector is open. |
| Close affordance | Visible control with accessible name `Close inspector`. |
| Pin affordance | Visible control with accessible name `Pin inspector` or `Unpin inspector`. |

Design contract. If the inspector opens as an overlay in a narrower viewport band, the grid behind the overlay MUST be inert to pointer and keyboard until the overlay closes.

### 7.4 Responsive band selection

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

| Viewport band | Width and height condition | Top bar | View bar | Grid | Inspector | Status strip | Design conformance |
| --- | --- | --- | --- | --- | --- | --- | --- |
| `base` | Selected by algorithm above. | Incident identity, built-in primary tabs, active-surface query controls, `System views`, current surface title, presence summary. | Saved-view and row-action controls. | Primary and visible while inspector is open. | Adjacent right panel. | Visible primary save label and secondary message. | Claimed. |
| `narrow_desktop` | Selected by algorithm above. | Incident identity, `Surfaces`, active-surface query controls, `System views`, current surface title, presence summary. | Saved-view and row-action controls. | Primary; visible when overlays are closed. | Full-height right overlay; grid inert behind overlay. | Visible primary save label and secondary message. | Claimed. |
| `compact_desktop` | Selected by algorithm above. | Incident identity, `Surfaces`, `System views`, current surface title, active-surface controls with chips in `Filters` popover. | Saved-view and row-action controls remain reachable. | Primary when overlays are closed. | Full-screen or full-height overlay; grid inert behind overlay. | Visible primary save label; presence summary assigned here. | Claimed. |
| `below_supported_minimum` | Selected by algorithm above. | Safe navigation and session controls remain reachable. | Not required except safe save/conflict path. | Degraded; horizontal scroll or supported-viewport message permitted. | Not required. | Primary save label MUST remain visible when unsaved work exists. | Not claimed. |

Design contract. Below the supported minimum, keyboard session logout and safe navigation MUST remain available. Omission of mobile/touch-specific gestures is conformant.

### 7.5 Responsive overflow algorithm

Design contract. Responsive overflow MUST use the region assignment table below. The table is exhaustive for shell controls in this revision.

| Control family | `base` location | `narrow_desktop` location | `compact_desktop` location | `below_supported_minimum` location |
| --- | --- | --- | --- | --- |
| Incident identity | Top bar, visible text | Top bar, visible text | Top bar, visible text | Top bar, visible text |
| Built-in surfaces | Top bar primary tabs | `Surfaces` menu in top bar | `Surfaces` menu in top bar | `Surfaces` menu optional; safe navigation required |
| Required system views | `System views` control in top bar | `System views` control in top bar | `System views` control in top bar | `System views` optional; safe navigation required |
| Current surface title | Top bar when not already represented by a selected built-in tab | Top bar when not already represented by a selected built-in tab | Top bar when not already represented by a selected built-in tab | Top bar if active surface is shown |
| Presence summary | Top bar | Top bar | Status strip | Status strip only if space remains after save label |
| Saved-view selector | View bar | View bar | View bar | Not required |
| Sort control | Top bar | Top bar | Top bar | Not required |
| Group control | Top bar | Top bar | Top bar | Not required |
| Filter control | Top bar | Top bar | Top bar as `Filters` popover | Not required |
| Active group/sort/filter chips | Top bar in one row, then `Filters` overflow if capacity exceeded | Top bar in at most two rows, then `Filters` overflow | Always inside `Filters` popover | Not required |
| Primary save label | Status strip | Status strip | Status strip | Status strip when unsaved work exists |
| Secondary status message | Status strip | Status strip with truncation rule | Accessible-only summary after primary label | Not required |

Design contract. Active chip placement MUST use this deterministic algorithm:

```pseudocode
place_active_chips(band, chips):
  ordered = [group chips by applied order] + [sort chips by applied order] + [filter chips by normalized query order]
  if band == base:
    inline_capacity = 8
    inline_rows = 1
  else if band == narrow_desktop:
    inline_capacity = 6
    inline_rows = 2
  else if band == compact_desktop:
    inline_capacity = 0
    inline_rows = 0
  else:
    inline_capacity = 0
    inline_rows = 0
  inline_count = min(length(ordered), inline_capacity * inline_rows)
  inline_chips = first inline_count chips
  overflow_chips = remaining chips
  if overflow_chips is empty: overflow_control = absent
  else: overflow_control = "Filters" with accessible name "Filters, " + length(overflow_chips) + " hidden"
  return inline_chips, overflow_control
```

Design contract. Secondary status-message truncation MUST preserve the full accessible text. Visible truncation uses the first 40 Unicode scalar values followed by `…` in `narrow_desktop`; visible truncation uses the first 24 Unicode scalar values followed by `…` in `compact_desktop` only when a visible secondary summary is rendered. If no secondary message exists, no empty announcement is emitted.

## 8. Grid, top-bar query controls, view bar, and workbook interaction

### 8.1 Grid identity and addressing

Core restatement. Workbook behavior is keyed by stable identifiers such as `view_schema_id`, `record_id`, `row_version`, `field_key`, and `client_txn_id`, not by visible tab labels, row numbers, column labels, projection table names, storage table names, or React component names. Owner: `01_architecture_storage_and_view_contracts.md` §3.3.1 and §7.4; `03_workbook_interaction_collaboration_and_workflows.md` §3.

Design contract. The UI MAY display human labels as text. Omission behavior: human labels MUST NOT replace stable identifiers for interaction state, mutation affordances, focus anchoring, visual fixture selectors, or accessibility relationships when those identifiers exist.

Design contract. Presentation-only rows such as group headers, empty states, loading rows, and measurement rows MUST NOT emit mutation events and MUST NOT be presented as authoritative records.

### 8.2 Cell and row state precedence

Design contract. Cell and row visual state selection MUST use the precedence table below. Priority `1` is highest.

| Priority | State ID | Applies when | Co-display rule | Required non-color cue |
| ---: | --- | --- | --- | --- |
| 1 | `conflicted` | Same-field conflict exists for the cell. | Co-displays active focus. Overrides invalid and pending visual treatment. | Marker with accessible name `Conflict on <field label>`. |
| 2 | `invalid` | Current cell value fails field or route validation. | Co-displays active focus. Overrides dirty and pending. | Error marker and field-local message. |
| 3 | `dirty_or_pending` | Local edit or queued mutation exists and no higher-priority conflict or invalid state applies. | Co-displays active focus. | Pending marker and accessible pending label. |
| 4 | `active_cell` | Cell has keyboard focus or edit focus. | Co-displays with higher-priority state. | Visible focus outline. |
| 5 | `selected_row` | Row is selected and no cell-local primary marker replaces it. | Row-level only; does not override cell-local markers. | Row selection marker and accessible selected state. |
| 6 | `read_only_or_derived` | Cell is read-only, hidden-derived, or non-writable. | Applies only when no higher-priority error, conflict, pending, or active edit state applies. | Accessible read-only state or visible read-only marker on focus. |
| 7 | `saved` | No higher-priority state applies. | Default steady state. | Status strip `Saved`; no persistent success decoration required. |

Design contract. `invalidate_or_refresh_required` is row-block state, not a cell-state priority. When it applies, the row or block MUST show a stale marker and MUST trigger refresh through owner behavior rather than inventing a design-local read path.

### 8.3 Top-bar query and view-bar controls

Design contract. Sort, filter, and group controls live in the Top bar and apply to the active surface only. Saved-view controls live in the View bar and apply to the active surface only. The active surface title is omitted from the query rail when the selected built-in tab already provides the same visible title.

| Control | Default state | Active state | Invalid state | Clear behavior | Ordering |
| --- | --- | --- | --- | --- | ---: |
| Active surface title | Current `view_schema_id` display label when not duplicated by the selected built-in tab. | Same. | Surface unavailable error. | Not clearable. | 1 |
| Sort control | No user sort override. | Sort chips shown. | Invalid sort field blocked before persistence. | Clears all user sort overrides. | 2 |
| Group control | `Group: None`. | One group chip. | Unsupported group disabled with explanation. | Sets grouping inactive. | 3 |
| Filter control | No filters. | Filter chips shown or `Filters` overflow. | Invalid filter chip marked and excluded from query submission. | Clears all filters or one selected chip. | 4 |
| Saved-view selector | `Unsaved view` when no saved view is active. | Saved view display name. | Fall back to base surface and show inline message. | Clears saved-view selection only, not active query unless the user selects reset. | 5 |

Design contract. Active chips MUST render in this order: group chip, sort chips in applied order, then filter chips in normalized query order.

Core restatement. Saved views are incident-bound workbook configurations over exactly one `view_schema_id`; a `system` saved view is not the same object as a contract-backed system view. Owner: `03_workbook_interaction_collaboration_and_workflows.md` §2.3, REQ-03-012 through REQ-03-026.

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
| Grid navigation mode | `Tab` to grid | §8.6 key-command table | `Enter` or printable key enters edit when writable | Not applicable | `Esc` no-op | No-op | Active cell | Grid container |
| Grid edit mode | From active writable cell | §8.6 key-command table | Editor-specific | `Enter`, blur by explicit commit, or declared commit shortcut | `Esc` discards uncommitted editor value | Exit edit mode | Edited cell | Grid container |
| Relationship chip | `Tab` or arrow within cell | Arrow keys within chip list | `Enter` opens inspect/action menu | Action commits through owner route | `Esc` closes menu | Close menu | Invoking chip | Owning cell |
| Inspector tab/section | `Tab` to inspector | Arrow keys for tabs; headings tabbable only when interactive | `Enter` or `Space` opens control | Control-specific | `Esc` closes overlay inspector only | Close overlay inspector in responsive bands | Invoking row/control | Active grid container |
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
5. top bar incident identity if no active surface is restorable.

### 8.6 Grid keyboard state machine

Design contract. Grid keyboard behavior MUST use the modes below.

| Mode | Entered by | Exited by | Focus target | Mutation permitted |
| --- | --- | --- | --- | --- |
| `grid_navigation` | `Tab` into grid, committed edit, canceled edit, restored focus | Enter edit, leave grid, overlay opens | Active cell | No direct text mutation. |
| `grid_edit` | `Enter`, printable key on writable cell, explicit edit command | Commit, cancel, declared blur commit | Editor | Yes. |
| `grid_range_selection` | `Shift+Arrow*` in navigation mode | Modifier release, explicit collapse, edit entry | Active range | No direct mutation. |
| `grid_disabled_or_read_only` | Active cell is non-writable | Move focus, leave grid, overlay opens | Active cell or containing row | No mutation. |

Design contract. The key-command table is exhaustive for grid-owned key chords in this revision.

| Key chord | Mode | Preconditions | Required result | Commit/cancel behavior | Prevent browser default |
| --- | --- | --- | --- | --- | --- |
| `ArrowUp` | `grid_navigation` | Previous visible row exists. | Move active cell to same field in previous visible row. | None. | Yes. |
| `ArrowDown` | `grid_navigation` | Next visible row exists. | Move active cell to same field in next visible row. | None. | Yes. |
| `ArrowLeft` | `grid_navigation` | Previous visible cell exists in row. | Move active cell to previous visible cell. | None. | Yes. |
| `ArrowRight` | `grid_navigation` | Next visible cell exists in row. | Move active cell to next visible cell. | None. | Yes. |
| `Shift+Arrow*` | `grid_navigation` | Target visible cell exists. | Enter or extend `grid_range_selection` to target cell. | None. | Yes. |
| `PageUp` | `grid_navigation` | At least one visible body row exists. | Move up by `max(visible_body_row_count - 1, 1)`. | None. | Yes. |
| `PageDown` | `grid_navigation` | At least one visible body row exists. | Move down by `max(visible_body_row_count - 1, 1)`. | None. | Yes. |
| `Home` | `grid_navigation` | Current row has visible navigation cell. | Move to first visible navigation cell in current row. | None. | Yes. |
| `End` | `grid_navigation` | Current row has visible navigation cell. | Move to last visible navigation cell in current row. | None. | Yes. |
| `Ctrl/Cmd+Home` | `grid_navigation` | Query result has visible rows. | Move to first visible row and first visible navigation cell. | None. | Yes. |
| `Ctrl/Cmd+End` | `grid_navigation` | Query result has visible rows. | Move to last visible row and last visible navigation cell. | None. | Yes. |
| `Enter` | `grid_navigation` | Active cell writable. | Enter `grid_edit`; seed editor with existing cell value. | None. | Yes. |
| `Enter` | `grid_edit` | Editor value valid for local commit. | Commit editor value and return to `grid_navigation`. | Commit. | Yes. |
| `Shift+Enter` | `grid_navigation` | Active cell writable. | Enter `grid_edit`; seed editor with existing cell value and place caret at end. | None. | Yes. |
| `Tab` | `grid_navigation` | No editor open. | Leave grid to next major shell region. | None. | No after focus transfer. |
| `Shift+Tab` | `grid_navigation` | No editor open. | Leave grid to previous major shell region. | None. | No after focus transfer. |
| `Tab` | `grid_edit` | Editor value valid for local commit. | Commit editor value and move to next major shell region. | Commit. | Yes. |
| `Shift+Tab` | `grid_edit` | Editor value valid for local commit. | Commit editor value and move to previous major shell region. | Commit. | Yes. |
| `Escape` | `grid_edit` | Editor open. | Discard uncommitted editor value and return to `grid_navigation`. | Cancel. | Yes. |
| `Escape` | `grid_navigation` | No higher-priority dismissible layer. | No-op. | None. | No. |
| Printable character | `grid_navigation` | Active cell writable. | Enter `grid_edit`; seed editor with printed character. | None. | Yes. |
| Printable character | `grid_navigation` | Active cell read-only. | Do not mutate; expose read-only state. | None. | Yes. |
| `Backspace` | `grid_navigation` | Active cell writable and emptying is permitted by owner behavior. | Enter `grid_edit`; seed editor with empty value. | None. | Yes. |
| `Delete` | `grid_navigation` | Active cell writable and emptying is permitted by owner behavior. | Enter `grid_edit`; seed editor with empty value. | None. | Yes. |
| `Ctrl/Cmd+C` | `grid_navigation` or `grid_range_selection` | Selection exists. | Copy selected visible cell values using workbook copy presentation. | None. | Yes. |
| `Ctrl/Cmd+V` | `grid_navigation` | Clipboard has text. | Dispatch base-profile paste handling for active surface. | Paste plan governs. | Yes. |

Design contract. Browser defaults not explicitly prevented in the table MUST remain available only when they do not move focus, mutate data, or trigger browser navigation away from the workbook shell.

## 9. Progressive structure, chips, and semantic states

### 9.1 Chip state input schema

Design contract. `chip_state_input_v1` MUST use the schema below before §9.2 runs.

| Field | Type | Required | Null permitted | Default on omission | Invalid behavior |
| --- | --- | ---: | ---: | --- | --- |
| `dismissed` | boolean | yes | no | none | Validation fails. |
| `resolved_record_id` | stable ID string | yes | yes | none | Validation fails. |
| `resolution_method` | `manual`, `auto`, `import`, `system`, or `null` | yes | yes | none | Validation fails. |
| `entity_type` | `host`, `identity`, `indicator`, `party`, or `record` | yes | no | none | Validation fails. |
| `raw_text` | string | yes | no | none | Validation fails. |
| `display_name` | string | required when `resolved_record_id` is not null | no | none | Validation fails. |
| `alias_text` | string | required when `resolution_method='auto'` | no | none | Validation fails. |

Design contract. String comparison inside chip rendering MUST use Unicode NFC normalization for display sorting only. The normalized string MUST NOT replace preserved source text.

### 9.2 Chip state selection

Design contract. Entity-mention and relationship chips MUST use this selection algorithm:

```pseudocode
select_chip_state(input):
  validate input as chip_state_input_v1
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

Design contract. Manual resolution is represented by resolution metadata on the `resolved` state; it MUST NOT create a fifth chip state.

### 9.3 Relationship and collection cells

Design contract. Relationship cells that mix unresolved mention tokens and canonical chips MUST preserve those as different object types. They MUST NOT be coerced into comma-delimited strings for display, editing, conflict resolution, or copy/paste presentation.

Design contract. Collection cells MAY display compact summaries under the capacity rules in §7.5. Omission behavior: hidden collection members in compact display are conformant only when an inspectable expansion path is present and the accessible name reports the hidden count.

### 9.4 Auto-resolution disclosure

Design contract. If a value is auto-resolved, the user MUST be able to tell it was auto-resolved, inspect the match reason when `alias_text` is present in `chip_state_input_v1`, and correct it through the same surface or inspector.

Design contract. Transient auto-resolution disclosure MAY fade after the user has seen the affected row at least once. Omission behavior: omission of fade is conformant. The chip state MUST remain marked as `auto_resolved` until the underlying resolution state changes.

## 10. Collaboration, save state, presence, and conflict design

### 10.1 Save-state input schema and labels

Core restatement. The compact workbook save-state presentation uses exactly `Syncing`, `Saved`, or `Conflict`; ambient collaboration state does not change the mapping. Owner: `03_workbook_interaction_collaboration_and_workflows.md` §4.2, REQ-03-089.

Design contract. `save_state_input_v1` MUST use the schema below before the save-state algorithm runs.

| Field | Type | Required | Null permitted | Default on omission | Invalid behavior |
| --- | --- | ---: | ---: | --- | --- |
| `same_field_conflict_count` | non-negative integer | yes | no | none | Validation fails. |
| `queue_overflow_refused` | boolean | yes | no | none | Validation fails. |
| `non_retryable_replay_failure` | boolean | yes | no | none | Validation fails. |
| `in_flight_workbook_mutation_count` | non-negative integer | yes | no | none | Validation fails. |
| `pending_queue_size` | non-negative integer | yes | no | none | Validation fails. |
| `replay_paused_for_recovery` | boolean | yes | no | none | Validation fails. |
| `reauthentication_required_for_replay` | boolean | yes | no | none | Validation fails. |
| `secondary_message_text` | string | no | no | no secondary message | Validation fails only if non-string. |

Design contract. The status strip MUST show exactly one primary save-state label selected by the algorithm below.

```pseudocode
select_primary_save_label(state):
  validate state as save_state_input_v1
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

Design contract. `Conflict` has precedence over `Syncing`. Presence MUST NOT change the primary save label. A secondary message MAY describe lower-priority state. Omission behavior: absence of the secondary message MUST NOT change the primary label. For same-field conflicts, any ordinary visible secondary message MUST summarize the affected count in user-facing language and MUST NOT use `record_id`, `field_key`, `conflict_token`, route names, or raw error text as its primary copy.

| Label | Visual treatment | Required accessible representation |
| --- | --- | --- |
| `Syncing` | Subtle spinner or pending marker plus label. | Text label and state description. |
| `Saved` | Text label only by default; no celebratory animation. | Text label. |
| `Conflict` | Conflict semantic marker and local entry point. | Text label, affected count when `same_field_conflict_count` is greater than zero, and action path. |

### 10.2 Presence input schema

Design contract. `presence_input_v1` MUST use the schema below before presence rendering.

| Field | Type | Required | Null permitted | Default on omission | Invalid behavior |
| --- | --- | ---: | ---: | --- | --- |
| `connection_id` | stable ID string | yes | no | none | Validation fails. |
| `user_id` | stable ID string | yes | no | none | Validation fails. |
| `display_name` | string | yes | no | none | Validation fails. |
| `mode` | `editing`, `focused`, `viewing`, or `idle` | yes | no | none | Validation fails. |
| `observed_at` | UTC instant | yes | no | none | Validation fails. |
| `expires_at` | UTC instant | yes | no | none | Validation fails. |
| `sheet_ref` | owner public shape | yes | no | none | Validation fails. |
| `record_id` | stable ID string | no | yes | null | Validation fails only if non-string and non-null. |
| `field_key` | stable field key string | no | yes | null | Validation fails only if non-string and non-null. |

Design contract. Presence timestamp comparisons MUST use UTC instants. A presence with `expires_at <= now` is expired and MUST NOT render. `display_name` comparison uses Unicode NFC, then code-point ascending. Equal sort keys fall through to `connection_id` ascending. Unknown modes are invalid and MUST NOT be silently rendered as another mode.

### 10.3 Presence rendering

Design contract. Presence MUST render at header, row, and cell levels according to this capacity table.

| Scope | Capacity | Overflow label | Required source key |
| --- | ---: | --- | --- |
| Header | 5 unique users | `+N` where `N` is hidden unique users. | Exact `sheet_ref` match. |
| Row | 3 unique users | `+N` where `N` is hidden unique users. | Exact `record_id` match. |
| Cell | 2 unique users | `+N` where `N` is hidden unique users. | Exact `record_id` and `field_key` match with `mode='editing'`. |

Design contract. Presence ordering MUST use this algorithm:

```pseudocode
order_presence(presences, now):
  candidates = [p for p in presences where p.expires_at > now]
  dedupe_key = p.user_id + ":" + p.mode + ":" + p.record_id + ":" + p.field_key
  keep the candidate with latest observed_at for each dedupe_key
  sort by mode_priority(editing=1, focused=2, viewing=3, idle=4)
       then observed_at descending
       then display_name Unicode NFC code-point ascending
       then connection_id ascending
  return sorted candidates
```

Design contract. If the same user has multiple visible presence states in one scope, the highest-priority mode wins for compact avatar display. Detailed expansion MAY list all non-expired connections. Omission behavior: omission of detailed expansion is conformant when compact display and overflow count remain correct.

### 10.4 Same-field conflict design

Core restatement. A same-field conflict remains unresolved until an analyst explicitly chooses a resolution, and same-field edits MUST NOT silently overwrite each other. Owner: `03_workbook_interaction_collaboration_and_workflows.md` §3.2 through §3.3.

Design contract. Same-field conflict UI MUST satisfy the table below.

| UI element | Required behavior |
| --- | --- |
| Conflicted cell | Shows conflict marker, preserves saved value visibility, and exposes resolver entry point. |
| Local draft | Preserved as client-local unsaved work until explicit resolution or clear action. |
| Resolver | Shows saved value, local draft, base value when present in owner payload, suggested merge when present, and explicit actions. |
| Save label | Remains `Conflict` while unresolved local conflict exists. |
| Focus restoration | Returns to conflicted cell after close or resolution when the cell still exists. |
| Non-conflicting work | Other rows and cells remain editable only when the active Core 03 writeability and concurrency contract permits editing for those rows and cells. |

Design contract. Same-field conflict resolution controls MUST NOT use primary accent fill. Destructive discard actions MUST use destructive styling and label text.

## 11. Evidence design

### 11.1 Evidence dimensions

Core restatement. Evidence record lifecycle and object blob upload state are separate owner machines. `object_blobs.upload_state` is not the user-facing evidence lifecycle, and an evidence record is valid with no blob. Owner: `02_domain_model_schema_and_history.md` §13; `03_workbook_interaction_collaboration_and_workflows.md` §8.

Design contract. Evidence rendering MUST keep three independent dimensions separate.

| Dimension | Closed UI values | Source owner |
| --- | --- | --- |
| Evidence lifecycle | `requested`, `pending_receipt`, `received`, `available`, `quarantined`, `released`, `inconsistent` | Core 02 §13. |
| Upload-slot overlay | `none`, `pending_upload_slot`, `failed_upload_slot`, `quarantined_upload_slot` | Core 03 §8 and Core 02 §13. |
| Preview capability | `previewable`, `download_only`, `unsupported_preview`, `access_blocked`, `not_applicable` | Core 01 and Core 04 evidence-access owners. |

Design contract. Evidence count MUST count evidence records, not pending blob slots. A `pending_upload_slot` alone MUST NOT increment evidence count.

### 11.2 Evidence UI validity algorithm

Design contract. Evidence UI validity MUST be classified by this deterministic algorithm for every lifecycle × upload overlay × preview capability combination.

```pseudocode
classify_evidence_ui(lifecycle, upload_overlay, preview_capability):
  if lifecycle not in evidence_lifecycle_values: return invalid_owner_state
  if upload_overlay not in upload_overlay_values: return invalid_owner_state
  if preview_capability not in preview_capability_values: return invalid_owner_state

  if lifecycle == inconsistent: return owner_repair_state
  if upload_overlay == quarantined_upload_slot: return owner_repair_state
  if lifecycle == quarantined: return valid_nonpreviewable
  if preview_capability == access_blocked: return valid_nonpreviewable

  if lifecycle in [requested, pending_receipt]:
    if preview_capability != not_applicable: return invalid_owner_state
    if upload_overlay in [pending_upload_slot, failed_upload_slot]: return valid_nonpreviewable
    return valid_nonpreviewable

  if lifecycle == received:
    if preview_capability == previewable: return valid_renderable
    if preview_capability in [download_only, unsupported_preview, not_applicable]: return valid_nonpreviewable
    return invalid_owner_state

  if lifecycle in [available, released]:
    if preview_capability == not_applicable: return invalid_owner_state
    if preview_capability == previewable: return valid_renderable
    return valid_nonpreviewable
```

Design contract. The classification values MUST render according to this table.

| Classification | Required rendering | Permitted actions |
| --- | --- | --- |
| `valid_renderable` | Render lifecycle label, evidence count, preview affordance, and download affordance when owner behavior permits download. | Preview and download controls follow owner access result. |
| `valid_nonpreviewable` | Render lifecycle or upload state and suppress preview. If download is blocked, show the blocked reason inline. | Download only when owner access result permits it. |
| `owner_repair_state` | Render inconsistent, quarantined, or repair-needed state without inventing a design-local repair route. | No preview; download suppressed unless owner behavior explicitly permits. |
| `invalid_owner_state` | Render owner-supplied error or unavailable state and mark fixture evidence failed when this appears in a valid fixture. | No preview; no design-local repair action. |
| `not_representable_by_design` | Must not appear in design fixtures. | No actions. |

Design contract. The lifecycle compatibility table below constrains the algorithm.

| Lifecycle | Valid preview capability | Valid upload overlays | Invalid combinations |
| --- | --- | --- | --- |
| `requested` | `not_applicable` only | `none`, `pending_upload_slot`, `failed_upload_slot` | Any previewable or downloadable preview state. |
| `pending_receipt` | `not_applicable` only | `none`, `pending_upload_slot`, `failed_upload_slot` | Any previewable or downloadable preview state. |
| `received` | `previewable`, `download_only`, `unsupported_preview`, `not_applicable`, `access_blocked` | All overlay values except `quarantined_upload_slot` unless owner state identifies repair. | Quarantined overlay without owner repair state. |
| `available` | `previewable`, `download_only`, `unsupported_preview`, `access_blocked` | All overlay values except `quarantined_upload_slot` unless owner state identifies repair. | `not_applicable`. |
| `released` | `previewable`, `download_only`, `unsupported_preview`, `access_blocked` | `none` only unless owner state identifies repair. | `not_applicable`; pending or failed upload overlay. |
| `quarantined` | `access_blocked` or `not_applicable` | `none`, `quarantined_upload_slot` | `previewable`, `download_only`, `unsupported_preview`. |
| `inconsistent` | Any preview capability value | Any upload overlay value | None; classification is always `owner_repair_state`. |

### 11.3 Evidence preview and download

Design contract. Preview and download affordances MUST be visually distinct. Preview MUST NOT silently fall back to download.

Design contract. Blocked preview states MUST show the blocker inline at the Evidence row, evidence chip, or inspector Evidence section. The workbook shell MUST remain visible.

Design contract. A preview action MUST NOT expose raw object-store URL text to the user. If a same-origin handle path is shown for diagnostics, it MUST be secondary metadata and MUST NOT become the visible evidence identity.

### 11.4 Evidence density

Design contract. Evidence badges in dense grid cells MUST show count plus highest-priority state marker. Priority order is `inconsistent`, `quarantined`, `access_blocked`, `failed_upload_slot`, `pending_upload_slot`, `available`, `released`, `received`, `pending_receipt`, `requested`.

Design contract. If a row has more evidence records than visible badge capacity, the badge MUST show the visible count and an inspectable overflow path with accessible name `Evidence, <total count> records`.

## 12. Component contracts

### 12.1 Component variant matrix

Design contract. The component variant matrix below is exhaustive for required component families in this revision.

| Component family | Required variants |
| --- | --- |
| Button | `primary`, `secondary`, `danger`, `ghost`, `disabled`, `loading`. |
| Icon button | `default`, `selected`, `danger`, `disabled`, `loading`. |
| Text input/editor | `default`, `focused`, `invalid`, `read_only`, `dirty`, `disabled`. |
| Chip | `neutral`, `unresolved`, `resolved`, `auto_resolved`, `dismissed`, `selected`, `disabled`. |
| Menu item | `default`, `focused`, `selected`, `disabled`, `destructive`. |
| Inspector section | `default`, `empty`, `loading`, `error`, `dirty`. |
| Toast | `info`, `success`, `warning`, `error`, `undoable`. |
| Dialog | `standard`, `destructive`, `blocking`, `safe_cancel`. |
| Banner/inline message | `info`, `warning`, `error`, `success`. |

### 12.2 Component-state property matrix

Design contract. Every required component state MUST map to visual properties, accessible state, activation, focusability, motion, and precedence through this matrix.

Design contract. The focusability value `conditional_action_focus` means the component instance MUST be `tabbable` when it renders a visible user-triggered action and MUST be `programmatic_only` when it renders information only. The focus-indicator value `conditional_action_focus_indicator` means `{components.focus-ring.border}` for the same visible user-triggered action condition and `none` otherwise.

Design contract. An activation value of `Optional action` means the component instance MAY render no action. Omission behavior: if no action is rendered, the component MUST present information only, MUST NOT enter the tab order except through `conditional_action_focus`, and MUST NOT expose a no-op action.

| Component state | Background | Foreground | Border | Focus indicator | Non-color cue | ARIA or accessible state | Activation rule | Focusability | Motion | Precedence |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | ---: |
| `button.primary` | `{components.button-primary.backgroundColor}` | `{components.button-primary.textColor}` | `none` | `{components.focus-ring.border}` | Label text | Accessible name from label | Activates primary action | `tabbable` | `{motion.duration-fast}` | 50 |
| `button.secondary` | `{components.button-secondary.backgroundColor}` | `{components.button-secondary.textColor}` | `{components.button-secondary.border}` | `{components.focus-ring.border}` | Label text | Accessible name from label | Activates secondary action | `tabbable` | `{motion.duration-fast}` | 50 |
| `button.danger` | `{components.button-danger.backgroundColor}` | `{components.button-danger.textColor}` | `none` | `{components.focus-ring.border}` | Destructive label | Accessible name includes destructive verb | Activates destructive confirmation or action | `tabbable` | `{motion.duration-fast}` | 40 |
| `button.ghost` | `transparent` | `{colors.ink-muted}` | `none` | `{components.focus-ring.border}` | Label text | Accessible name from label | Activates low-emphasis action | `tabbable` | `{motion.duration-fast}` | 50 |
| `button.disabled` | `{colors.surface-2}` | `{colors.ink-tertiary}` | `{border.hairline}` | `none` | Disabled text or state | `aria-disabled=true` | Blocked | `not_focusable` | `none` | 10 |
| `button.loading` | `{colors.surface-2}` | `{colors.ink-muted}` | `{border.hairline}` | `{components.focus-ring.border}` | Pending marker and text | `aria-busy=true` | Blocked until loading clears | `tabbable` | `{motion.duration-fast}` | 20 |
| `icon_button.default` | `transparent` | `{colors.ink-muted}` | `none` | `{components.focus-ring.border}` | Semantic icon ID | Accessible name from §3.11 | Activates declared action | `tabbable` | `{motion.duration-fast}` | 50 |
| `icon_button.selected` | `{colors.surface-3}` | `{colors.ink}` | `{border.hairline}` | `{components.focus-ring.border}` | Selected marker | `aria-pressed=true` for toggle instances; otherwise none | Activates declared action | `tabbable` | `{motion.duration-fast}` | 40 |
| `icon_button.danger` | `transparent` | `{colors.semantic-destructive}` | `none` | `{components.focus-ring.border}` | Destructive icon and name | Accessible name includes destructive verb | Opens confirmation or destructive action | `tabbable` | `{motion.duration-fast}` | 30 |
| `icon_button.disabled` | `transparent` | `{colors.ink-tertiary}` | `none` | `none` | Disabled state | `aria-disabled=true` | Blocked | `not_focusable` | `none` | 10 |
| `text_input.default` | `{components.text-input.backgroundColor}` | `{components.text-input.textColor}` | `{components.text-input.border}` | `none` | Label or placeholder | Accessible name from label | Edits text | `tabbable` | `none` | 50 |
| `text_input.focused` | `{components.text-input.backgroundColor}` | `{components.text-input.textColor}` | `{border.focus}` | `{components.focus-ring.border}` | Caret and focus ring | Focused input | Edits text | `tabbable` | `none` | 30 |
| `text_input.invalid` | `{components.text-input.backgroundColor}` | `{colors.ink}` | `2px solid {colors.semantic-conflict}` | `{components.focus-ring.border}` | Error marker and message | `aria-invalid=true` | Edits text until valid | `tabbable` | `none` | 20 |
| `text_input.read_only` | `{colors.surface-2}` | `{colors.ink-muted}` | `{border.hairline}` | `{components.focus-ring.border}` | Read-only marker | `readonly` or `aria-readonly=true` | No mutation | `tabbable` | `none` | 40 |
| `chip.unresolved` | `{components.chip.backgroundColor}` | `{components.chip.textColor}` | `1px dashed {colors.semantic-caution}` | `{components.focus-ring.border}` | `?` or `Unresolved` | Name pattern from §9.2 | Opens resolution path | `conditional_action_focus` | `none` | 20 |
| `chip.resolved` | `{components.chip.backgroundColor}` | `{colors.ink}` | `{components.chip.border}` | `{components.focus-ring.border}` | Canonical label | Name pattern from §9.2 | Opens inspection path | `conditional_action_focus` | `none` | 50 |
| `chip.auto_resolved` | `{components.chip.backgroundColor}` | `{colors.ink}` | `{components.chip.border}` | `{components.focus-ring.border}` | `auto` marker | Name pattern from §9.2 | Opens correction path | `conditional_action_focus` | `none` | 30 |
| `chip.dismissed` | `transparent` | `{colors.ink-tertiary}` | `{border.hairline}` | `{components.focus-ring.border}` | `dismissed` marker | Name pattern from §9.2 | Opens inspection path only | `conditional_action_focus` | `none` | 40 |
| `menu_item.focused` | `{colors.surface-3}` | `{colors.ink}` | `none` | `{components.focus-ring.border}` | Highlight plus text | `none` | Activates item | `tabbable` inside menu | `none` | 30 |
| `menu_item.selected` | `{colors.surface-3}` | `{colors.ink}` | `{border.hairline}` | `{components.focus-ring.border}` | Check or selected text | `aria-selected=true` | Selects item | `tabbable` inside menu | `none` | 30 |
| `toast.warning` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-caution}` | `conditional_action_focus_indicator` | Warning label | Live-region rule in §14.2 | Optional action | `conditional_action_focus` | `{motion.duration-fast}` | 30 |
| `toast.error` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-conflict}` | `conditional_action_focus_indicator` | Error label | Live-region rule in §14.2 | Optional action | `conditional_action_focus` | `{motion.duration-fast}` | 20 |
| `dialog.destructive` | `{colors.surface-1}` | `{colors.ink}` | `1px solid {colors.semantic-destructive}` | trapped focus | Destructive title and action label | `role=dialog`; destructive action named | Explicit confirmation | focus trapped | `{motion.duration-fast}` | 10 |
| `icon_button.loading` | `transparent` | `{colors.ink-muted}` | `none` | `{components.focus-ring.border}` | Pending marker and accessible name | `aria-busy=true` | Blocked until loading clears | `tabbable` | `{motion.duration-fast}` | 20 |
| `text_input.dirty` | `{components.text-input.backgroundColor}` | `{components.text-input.textColor}` | `1px dashed {colors.semantic-caution}` | `{components.focus-ring.border}` | Dirty marker | Described by pending or unsaved text | Edits text | `tabbable` | `none` | 35 |
| `text_input.disabled` | `{colors.surface-2}` | `{colors.ink-tertiary}` | `{border.hairline}` | `none` | Disabled state text | `aria-disabled=true` | No mutation | `not_focusable` | `none` | 10 |
| `chip.neutral` | `{components.chip.backgroundColor}` | `{components.chip.textColor}` | `{components.chip.border}` | `{components.focus-ring.border}` | Text label | Accessible name from label | Opens inspection path for action-bearing chip; otherwise information-only | `conditional_action_focus` | `none` | 60 |
| `chip.selected` | `{colors.surface-3}` | `{colors.ink}` | `{border.strong}` | `{components.focus-ring.border}` | Selected marker | `aria-selected=true` or selected text | Selects or opens declared action | `tabbable` | `none` | 30 |
| `chip.disabled` | `{colors.surface-2}` | `{colors.ink-tertiary}` | `{border.hairline}` | `none` | Disabled text | `aria-disabled=true` | Blocked | `not_focusable` | `none` | 10 |
| `menu_item.default` | `{colors.surface-1}` | `{colors.ink}` | `none` | `none` | Item text | Accessible name from item text | Activates item | `tabbable` inside menu | `none` | 50 |
| `menu_item.disabled` | `{colors.surface-1}` | `{colors.ink-tertiary}` | `none` | `none` | Disabled item text | `aria-disabled=true` | Blocked | `not_focusable` | `none` | 10 |
| `menu_item.destructive` | `{colors.surface-1}` | `{colors.semantic-destructive}` | `none` | `{components.focus-ring.border}` | Destructive verb text | Accessible name includes destructive verb | Opens confirmation or destructive action | `tabbable` inside menu | `none` | 20 |
| `inspector_section.default` | `transparent` | `{colors.ink}` | `none` | `{components.focus-ring.border}` | Section heading | Heading and region label | Expands, collapses, or focuses section controls | `conditional_action_focus` | `none` | 50 |
| `inspector_section.empty` | `transparent` | `{colors.ink-muted}` | `none` | `none` | Empty text | Empty message text | Optional action | `conditional_action_focus` | `none` | 60 |
| `inspector_section.loading` | `transparent` | `{colors.ink-muted}` | `none` | `none` | Loading text | `aria-busy=true` | Blocked | `programmatic_only` | `{motion.duration-fast}` | 20 |
| `inspector_section.error` | `transparent` | `{colors.semantic-conflict}` | `none` | `{components.focus-ring.border}` | Error text | Error message text | Optional action | `conditional_action_focus` | `none` | 10 |
| `inspector_section.dirty` | `transparent` | `{colors.ink}` | `1px dashed {colors.semantic-caution}` | `{components.focus-ring.border}` | Dirty marker | Described by unsaved state text | Section controls remain usable | `conditional_action_focus` | `none` | 30 |
| `toast.info` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-info}` | `conditional_action_focus_indicator` | Info label | Live-region rule in §14.2 | Optional action | `conditional_action_focus` | `{motion.duration-fast}` | 50 |
| `toast.success` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-success}` | `conditional_action_focus_indicator` | Success label | Live-region rule in §14.2 | Optional action | `conditional_action_focus` | `{motion.duration-fast}` | 50 |
| `toast.undoable` | `{colors.surface-2}` | `{colors.ink}` | `{border.hairline}` | `conditional_action_focus_indicator` | Undo action text | Accessible undo action | Activates undo action while present | `tabbable` | `{motion.duration-fast}` | 30 |
| `dialog.standard` | `{colors.surface-1}` | `{colors.ink}` | `{border.strong}` | trapped focus | Dialog title | `role=dialog`; title labels dialog | Explicit action or safe cancel | focus trapped | `{motion.duration-fast}` | 50 |
| `dialog.blocking` | `{colors.surface-1}` | `{colors.ink}` | `1px solid {colors.semantic-caution}` | trapped focus | Blocking explanation | `role=dialog`; blocking reason described | Blocked until required action is satisfied | focus trapped | `{motion.duration-fast}` | 5 |
| `dialog.safe_cancel` | `{colors.surface-1}` | `{colors.ink}` | `{border.strong}` | trapped focus | Safe cancel button | Safe cancel accessible name | Cancels without committing | focus trapped | `none` | 30 |
| `banner_inline.info` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-info}` | `conditional_action_focus_indicator` | Info marker and text | Info text or region label | Optional action | `conditional_action_focus` | `none` | 50 |
| `banner_inline.warning` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-caution}` | `conditional_action_focus_indicator` | Warning marker and text | Warning text or region label | Optional action | `conditional_action_focus` | `none` | 20 |
| `banner_inline.error` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-conflict}` | `conditional_action_focus_indicator` | Error marker and text | Error text or region label | Optional action | `conditional_action_focus` | `none` | 10 |
| `banner_inline.success` | `{colors.surface-2}` | `{colors.ink}` | `1px solid {colors.semantic-success}` | `conditional_action_focus_indicator` | Success marker and text | Success text or region label | Optional action | `conditional_action_focus` | `none` | 50 |

Design contract. Component states not listed in the matrix are invalid in required fixtures unless a later design revision adds a row.

### 12.3 Compound-state precedence

Design contract. Compound component states MUST resolve through this precedence table.

| Combination | Required winner | Required co-display |
| --- | --- | --- |
| `disabled + hover` | `disabled` | No hover visual. |
| `loading + active` | `loading` | Active visual suppressed. |
| `focus_visible + conflicted` | `conflicted` primary | Focus ring co-displays. |
| `invalid + pending` | `invalid` primary | Pending marker MAY appear as secondary marker. Omission behavior: pending marker omission is conformant when invalid state and accessible message remain. |
| `destructive + disabled` | `disabled` primary | Destructive text MAY remain only if contrast passes. Omission behavior: destructive text omission is conformant when disabled state and accessible name remain. |

### 12.4 Buttons

Design contract. Primary buttons MUST use `{components.button-primary}`. Secondary buttons MUST use `{components.button-secondary}`. Destructive buttons MUST use `{components.button-danger}` or a destructive dialog action row in §12.2.

Design contract. Destructive action controls MUST include destructive verb text. Icon-only destructive actions are invalid.

Design contract. Loading buttons MUST preserve their accessible name and add pending state. Disabled buttons MUST NOT submit actions.

### 12.5 Inputs and editors

Design contract. Grid editors MUST preserve visible cell context and row identity while editing. Editor chrome MUST NOT cover the entire grid unless the viewport band already uses overlay mode for inspector or preview.

Design contract. Invalid editor state MUST show field-local message and accessible invalid state. It MUST NOT rely on red border alone.

### 12.6 Menus and popovers

Design contract. Menus and popovers MUST be anchored to invoking controls, close through §8.5, and restore focus through §8.5.

Design contract. Menu items that open nested UI MUST state that result in accessible name or description.

### 12.7 Inspector sections

Design contract. Inspector sections MUST render in this order: Details, Relationships, Evidence, History.

Design contract. Empty sections MUST show a concise empty state and, when the current user has an available create or link action under owner behavior, an action entry point. If no action is available, the empty state MUST say why no action is available.

### 12.8 Empty states, messages, and dialogs

Design contract. Empty grids MUST distinguish successful empty query, no permission, unavailable surface, and loading state.

| Empty state | Required message posture | Action rule |
| --- | --- | --- |
| Successful empty query | Neutral. | Show create action only when owner behavior permits creation. |
| Filtered empty result | Neutral with filter context. | Show clear filters action. |
| No permission | Restrained error. | No create action. |
| Surface unavailable | Error with retry or repair path if owner behavior exposes one. | No design-local repair route. |
| Loading | Pending state. | No fake rows. |

Design contract. Toasts MUST NOT carry critical unresolved state as the only visible location. Critical unresolved state MUST also appear in the relevant row, cell, inspector section, or status strip.

Design contract. Dialogs are reserved for destructive confirmation, blocking authentication/session messages, unsupported viewport messages, and multi-step owner flows. Routine row edits MUST NOT open dialogs.

## 13. Per-surface design contracts

Design contract. The table below defines surface design posture only. It MUST NOT define Core-owned field membership, write rules, route behavior, or lifecycle behavior.

| Surface | Required design posture | Primary state markers | Inspector emphasis | Empty-state posture |
| --- | --- | --- | --- | --- |
| Timeline | Fast rough capture, chronological scanning, progressive mention resolution. | Reviewed, conflicted, unresolved mentions, evidence count. | Details, Relationships, Evidence, History. | Permit rough creation when owner behavior permits creation. |
| Hosts | Scope and asset posture, links to events, evidence, assessments. | Criticality, containment, unresolved refs, linked evidence. | Relationships and History. | Explain no hosts discovered or captured. |
| Identities | Account or persona scoping, reset/MFA state, links to events. | Privilege, MFA/reset, unresolved refs, linked evidence. | Relationships and History. | Explain no identities captured. |
| Evidence | Evidence envelope state, upload and preview status, collector/source references. | Lifecycle, upload overlay, preview capability, custody flags. | Evidence and History. | Offer request or attach path only when owner behavior permits. |
| Notes | Source-preserving analyst material. | Linked records, tags, unresolved mentions. | Details and Relationships. | Offer note creation when owner behavior permits. |
| Indicators | Canonical indicator rows and source-bound observation pivots. | Lifecycle, first/last observed, source count. | Relationships and History. | Explain no canonical indicators captured. |
| Compromise Assessments | Review-oriented host or identity assessment posture. | Status, confidence, subject, history. | Details and History. | Explain no assessments captured. |
| Task Requests | Queue-oriented operational work surface. | Status, priority, blocked, owner, due. | Details and Relationships. | Offer task creation when owner behavior permits. |
| Decisions | Rationale-bearing decisions and review state. | Status, owner, superseded, decided time. | Details and History. | Offer decision creation when owner behavior permits. |
| Parties | Coordination identities, requester/collector/source/audience references. | Party kind, linked references. | Relationships. | Offer party creation when owner behavior permits. |
| Communications Log | Durable communication memory. | Audience, channel, linked decisions/actions. | Relationships. | Offer log creation when owner behavior permits. |
| Handoff | Shift or phase continuity. | Acknowledgement, open tasks, open decisions, open risks. | Details and Relationships. | Offer handoff creation when owner behavior permits. |
| Status Review | Coordination checkpoint surface. | Blocked tasks, pending evidence, open decisions, next report. | Details and Relationships. | Offer status-review creation when owner behavior permits. |
| Lesson | Learning and follow-through. | Closure state, owner, follow-up tasks. | Details and Relationships. | Offer lesson creation when owner behavior permits. |
| Findings | Optional structured findings and hypotheses surface. | State, confidence, owner. | Details and Relationships. | If not implemented, absent from switcher; if implemented and empty, explain no findings captured. |
| Investigative Queries | Optional query library surface. | Platform, purpose, creator. | Details. | If not implemented, absent from switcher; if implemented and empty, explain no queries captured. |
| Forensic Keywords | Optional keyword surface. | Pattern, reason, linked records. | Details. | If not implemented, absent from switcher; if implemented and empty, explain no keywords captured. |

Design contract. Optional surfaces in this table MUST NOT appear unless implemented and exposed. Omission of an optional surface is conformant and MUST NOT affect required surfaces.

## 14. Accessibility contract

### 14.1 Accessibility conformance matrix

Design contract. `dark_graphite` MUST satisfy WCAG 2.2 AA for all required state-bearing text, controls, focus indicators, and non-text state markers. This design contract owns the local token-pair matrix in §14.3; it does not redefine the external standard.

| Requirement family | Required behavior |
| --- | --- |
| Keyboard access | Every action reachable by pointer is reachable by keyboard unless the action is pointer-only by owner behavior and has a keyboard alternative. |
| Focus visibility | Focus indicator uses `{border.focus}` or a token-pair row in §14.3. |
| State communication | Conflict, pending, invalid, read-only, selected, disabled, evidence blocked, auto-resolved, dismissed, and unresolved states have non-color cues. |
| Accessible names | Icon-only controls use §3.11 accessible names. Rows include human-readable surface context and are not named only by `record_id`. |
| Reduced motion | §6.3 governs reduced-motion behavior. |
| Live regions | §14.2 governs announcements. |

### 14.2 Live-region event matrix

Design contract. Live-region behavior MUST use this matrix.

| Event | Live-region behavior | Required announcement content |
| --- | --- | --- |
| Save state changes to `Syncing` | Polite. | `Syncing changes`. |
| Save state changes to `Saved` | Polite only after prior non-saved state. | `Saved`. |
| Save state changes to `Conflict` | Assertive. | `Conflict. <count> unresolved` when count is present. |
| Same-field conflict opens | Assertive. | Field label and conflict state. |
| Evidence preview blocked | Polite. | Evidence title or row context plus blocker. |
| Evidence upload failed | Assertive. | Evidence or upload context plus failure state. |
| Presence update only | No live announcement. | None. |
| Auto-resolution batch complete | Polite. | Count auto-resolved and count unresolved. |
| Toast warning or error | Follows toast severity. | Message text plus action name if present. |
| Menu open/close | No live announcement beyond focus and ARIA state. | None. |

### 14.3 Contrast pair matrix

Design contract. Required foreground/background pairings are closed to this matrix for required fixtures. Unlisted foreground/background pairings are invalid in required fixtures.

| `contrast_pair_id` | Foreground token | Background or computed background | State context | Minimum ratio | Text size class | Alpha composition | Accepted contexts | Fixture coverage |
| --- | --- | --- | --- | ---: | --- | --- | --- | --- |
| `cp.body.canvas` | `{colors.ink}` | `{colors.canvas}` | Text | 4.5:1 | Normal | None | Shell body, grid surround | `D-VFIX-001` |
| `cp.body.surface1` | `{colors.ink}` | `{colors.surface-1}` | Text | 4.5:1 | Normal | None | Inspector, menus, panels | `D-VFIX-002` |
| `cp.muted.surface1` | `{colors.ink-muted}` | `{colors.surface-1}` | Text | 4.5:1 | Normal | None | Metadata and secondary labels | `D-VFIX-001` |
| `cp.subtle.surface2` | `{colors.ink-subtle}` | `{colors.surface-2}` | Text | 4.5:1 | Normal | None | Low-emphasis metadata | `D-VFIX-001` |
| `cp.accent.onaccent` | `{colors.on-accent}` | `{colors.accent}` | Text/control | 4.5:1 | Normal | None | Primary button text | `D-VFIX-009` |
| `cp.focus.surface` | `{colors.hairline-focus}` | `{colors.surface-1}` | Focus indicator | 3:1 | Non-text | None | Focus ring against panel | `D-VFIX-009` |
| `cp.conflict.surface` | `{colors.semantic-conflict}` | `{colors.surface-1}` | State marker | 3:1 | Non-text | None | Conflict marker, error border | `D-VFIX-003` |
| `cp.caution.surface` | `{colors.semantic-caution}` | `{colors.surface-1}` | State marker | 3:1 | Non-text | None | Pending, warning, unsupported preview | `D-VFIX-006` |
| `cp.success.surface` | `{colors.semantic-success}` | `{colors.surface-1}` | State marker | 3:1 | Non-text | None | Successful attach, completed task | `D-VFIX-007` |
| `cp.destructive.surface` | `{colors.semantic-destructive}` | `{colors.surface-1}` | Text/control | 4.5:1 | Normal | None | Destructive action labels | `D-VFIX-008` |
| `cp.overlay.text` | `{colors.ink}` | `{colors.overlay-scrim}` composited over `{colors.canvas}` | Overlay text | 4.5:1 | Normal | Source-over alpha composition | Overlay inspector/dialog text | `D-VFIX-002` |

Design contract. Alpha colors MUST be composited against the declared background before contrast calculation. Disabled states that communicate information MUST meet the declared text or non-text target for that state. Focus indicators MUST be evaluated against adjacent colors, not only against the main background.

## 15. Visual fixtures and implementation guidance

### 15.1 Visual fixture execution contract

Design contract. Visual fixture evidence is design evidence only. It MUST NOT be treated as Core 05 claim-bearing timed or fixture-sensitive publication evidence unless Core 05 publication requirements are separately satisfied.

Design contract. Every visual fixture MUST use this execution contract.

| Field | Required value or rule |
| --- | --- |
| Browser baseline | Chromium, headed mode, exact build supplied by the active visual harness artifact. If no exact build is recorded, fixture evidence is invalid. |
| Device scale factor | `1`. |
| Zoom | `{layout.zoomDefault}`. |
| Theme | `dark_graphite`. |
| Density | `{density.default-mode}` unless fixture row declares another density from §3.9. |
| Font fallback | Registry font stack; custom font files MUST NOT be distributed to users as evidence artifacts. |
| Seed data | Stable fixture dataset ID declared in the row. |
| Dynamic masks | Closed grammar: `mask(selector=<stable selector>, region=<text\|bbox\|attribute>, replacement=<fixed string>)`. |
| Scroll anchor | `top_left`, `row:<record_id>`, `cell:<record_id>:<field_key>`, `right_edge`, or `status_strip`. Missing anchor is fixture failure. |
| Crop rule | `full_viewport`, `selector:<stable selector>`, or `region:<x,y,width,height>`. |
| Diff tolerance | `0` pixel mismatch unless the fixture row declares a non-zero tolerance with rationale. |
| Artifact naming | `design/<fixture_id>/<viewport>/<theme>/<density>/<artifact_kind>.png` plus metadata JSON. |
| Failure output | Retain expected, actual, diff, metadata, mask summary, and validation summary. |

### 15.2 Visual fixture registry

Design contract. The visual fixture registry is closed to the rows below for this revision.

| Fixture ID | Required state | Viewport | Zoom | Density | Theme | Scroll normalization | Dynamic masks | Crop rule | Pass condition |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `D-VFIX-001` | Default Timeline workbook shell with top-bar query controls, compact view bar, compact Timeline grid, row gutter, header affordances, selected row, focused Summary cell, row-context inspector, status strip, Core 01 default Timeline fields, and no admin-card dominance above the active grid. | `1440x900 CSS px` | `{layout.zoomDefault}` | `compact` | `dark_graphite` | `top_left` | Actor names, timestamps, IDs. | `full_viewport` | All shell regions visible in the first viewport, admin/control content absent unless explicitly opened, and token pairs pass §14.3. |
| `D-VFIX-002` | Inspector open adjacent at base viewport. | `{layout.baseViewport}` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `cell:rec_timeline_001:timeline.summary` | Actor names, timestamps, IDs. | `full_viewport` | Grid remains visible; inspector sections in required order. |
| `D-VFIX-003` | Same-field conflict cell and resolver. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `cell:rec_timeline_conflict:timeline.summary` | Actor names, timestamps, IDs. | `selector:[data-design-fixture='conflict']` | Conflict marker, local draft, saved value, and actions visible. |
| `D-VFIX-004` | Unresolved, resolved, auto-resolved, and dismissed chips. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `row:rec_timeline_mentions` | Actor names, IDs. | `selector:[data-design-fixture='chips']` | Four chip states have distinct non-color cues. |
| `D-VFIX-005` | Presence at header, row, and cell with overflow. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `row:rec_timeline_presence` | Actor names. | `full_viewport` | Presence ordering and `+N` labels follow §10.3. |
| `D-VFIX-006` | Evidence states: available, blocked preview, pending upload, failed upload, quarantined. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `row:rec_evidence_matrix` | Filenames, hashes, timestamps. | `selector:[data-design-fixture='evidence']` | Evidence badges and actions follow §11. |
| `D-VFIX-007` | Save-state strip for `Syncing`, `Saved`, and `Conflict`. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `status_strip` | Timestamps. | `selector:[data-design-fixture='status-strip']` | One primary save label visible and accessible. |
| `D-VFIX-008` | Destructive actions in History and Relationships inspector. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `cell:rec_timeline_history:timeline.summary` | Actor names, IDs. | `selector:[data-design-fixture='destructive-actions']` | Destructive actions have label text and destructive styling. |
| `D-VFIX-009` | Component state matrix sample. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `top_left` | None. | `selector:[data-design-fixture='components']` | Required component states render and pass §14.3. |
| `D-VFIX-010` | Narrow desktop shell. | `1024x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `top_left` | Actor names, timestamps, IDs. | `full_viewport` | Built-in tabs collapse to `Surfaces`; required controls remain reachable. |
| `D-VFIX-011` | Compact desktop shell. | `768x640 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `top_left` | Actor names, timestamps, IDs. | `full_viewport` | Chips move to `Filters`; presence moves to status strip. |
| `D-VFIX-012` | Successful empty query. | `1280x720 CSS px` | `{layout.zoomDefault}` | `{density.default-mode}` | `dark_graphite` | `top_left` | IDs. | `selector:[data-design-fixture='empty-state']` | Empty state distinguishes filtered empty and successful empty. |

### 15.3 Coding-agent and developer guidance

Design contract. Coding agents and implementers MUST use stable selectors based on design fixture IDs, `view_schema_id`, `record_id`, `field_key`, and semantic icon IDs. They MUST NOT assert against incidental DOM hierarchy, visible row number, SQL table name, projection table name, or package-specific icon name.

Design contract. Generated screenshots, fixtures, or implementation examples MUST NOT include live incident content, raw evidence bytes, credentials, object-store keys, or unmasked actor names.

## 16. Boundaries, non-goals, future profiles, and external dependencies

### 16.1 Owner boundaries

Design contract. The boundary table below is binding for this document.

| Subject | Owner |
| --- | --- |
| Implementation conformance, profile scope, and authority order | Core 00 through Core 04. |
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
| External visual reference board | Non-authoritative. | Inspiration only; it cannot override token, state, or surface contracts. |
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
- Use raw object-store URLs, SQL table names, projection names, route helper names, vendor grid coordinates, package-specific icon names, component names, or styling classes as user-facing design concepts.
- Add neon threat maps, cinematic gradients, heavy glow, or decorative risk heatmaps to the live workbook shell.

## 18. Acceptance criteria

Design contract. This `design.md` is ready to guide design implementation only when every criterion below passes.

### 18.1 Editorial validation criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-001` | §1 | Document review | The document states that it is a design-direction contract and does not define Base Profile or extension-profile implementation conformance. | Scope statement missing or contradicted. |
| `D-AC-002` | §1 and §16 | Document review | The document states that Core 05 is publication-only for claim-bearing timed or fixture-sensitive criteria. | Core 05 is treated as runtime design owner. |
| `D-AC-003` | All accepted sections | Lint | No accepted normative section contains unresolved placeholder text. | Placeholder or unresolved blocker appears without non-goal or external dependency classification. |
| `D-AC-004` | §2 | Lint | Every normative paragraph uses a valid statement class or appears in a table introduced by a valid statement class. | Unmarked normative prose exists. |
| `D-AC-005` | §2 | Lint | Every **MAY** clause states omission behavior in the same paragraph, table row, or immediately following sentence. | Optional behavior lacks omission behavior. |
| `D-AC-006` | All exhaustive tables | Table audit | Every table declared exhaustive contains no duplicate case key and no omitted case in its declared scope. | Duplicate or missing declared case. |
| `D-AC-007` | All `Core restatement.` paragraphs | Owner-traceability audit | Every `Core restatement.` contains an exact owner reference or is reclassified. | Owner missing, broad, or unverifiable. |
| `D-AC-008` | §2.1 | Wording lint | No accepted normative prose contains informal permissive language outside declared exceptions. | Invalid wording pattern remains. |

### 18.2 Token and schema validation criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-010` | §3.1 and §3.2 | Token-schema validation | Document metadata and token registry are disjoint validation regions. | Metadata is exposed as token namespace or token key. |
| `D-AC-011` | §3.2 | Token-schema validation | Unknown keys, duplicate keys, unsupported scalar forms, cycles, and invalid body token usage fail validation. | Invalid token state is accepted. |
| `D-AC-012` | §3.3 | Scalar grammar validation | Every CSS-like token value validates under exactly one scalar grammar. | Token value validates under zero or multiple grammars. |
| `D-AC-013` | §3.5 | Token-resolution validation | All token references resolve without cycles. | Unknown reference or cycle exists. |
| `D-AC-014` | §3.6 | Body token-use validation | Body text refers to token-owned design values by token ID or design-literal registry row. | Undeclared token-owned scalar controls design output. |
| `D-AC-015` | §3.8 | Theme validation | `dark_graphite` is the only required theme; `light` and `high_contrast` have omission semantics. | Theme exposure or omission behavior is ambiguous. |
| `D-AC-016` | §3.10 and §3.11 | Icon registry validation | Every semantic icon ID used by the shell appears in the semantic icon registry. | Icon usage lacks semantic ID or fallback. |

### 18.3 Surface, layout, and responsive criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-020` | §7.2 | Surface-registry audit | The shell-exposure registry contains all fourteen required base-profile surfaces exactly once. | Required surface missing or duplicated. |
| `D-AC-021` | §7.2 | Surface-registry audit | The shell-exposure registry contains the three standardized optional workbook surfaces exactly once with omission semantics. | Optional surface missing, duplicated, or lacks omission behavior. |
| `D-AC-022` | §7.2 | Keyboard and pointer fixture | Required system views are reachable by keyboard and pointer from the shell, and selection completes to the requested active surface grid. | Required system view requires command palette, hidden route, or a visible option that does not complete surface activation. |
| `D-AC-023` | §7.4 | Responsive algorithm validation | Each viewport width and height combination falls into exactly one responsive band. | Overlap or gap exists. |
| `D-AC-024` | §7.5 | Responsive fixture | Responsive overflow selects the same rendered location, truncation, popover, and accessible label for each declared viewport. | Same viewport permits divergent layouts. |
| `D-AC-025` | §7.4 | Below-minimum fixture | Below-minimum behavior is explicitly non-conformant or degraded with safe navigation preserved. | Below-minimum viewport claims design conformance or loses safe navigation. |

### 18.4 State and interaction criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-030` | §10.1 | Algorithm test | Save-state selection returns exactly one of `Syncing`, `Saved`, or `Conflict` for every valid input combination and rejects invalid schema inputs. | Ambiguous label or unhandled input. |
| `D-AC-031` | §8.2 | State precedence test | Cell-state precedence produces one deterministic visual primary state plus declared co-displays. | State priority conflict or color-only state. |
| `D-AC-032` | §9.1 and §9.2 | Algorithm test | Chip-state selection is deterministic and rejects omitted, null, and invalid inputs according to schema. | Unhandled chip input or silent coercion. |
| `D-AC-033` | §11.2 | Evidence matrix test | Evidence UI validity classifies every lifecycle × overlay × preview combination. | Any combination unclassified or silently coerced. |
| `D-AC-034` | §10.2 and §10.3 | Presence algorithm test | Presence ordering and overflow are deterministic for equal timestamps and duplicate users. | Ordering depends on array order or map iteration. |
| `D-AC-035` | §8.3 | Shell-control test | Top-bar query control order, chip order, and saved-view dirty state are deterministic. | Control ordering or dirty state differs across implementations. |
| `D-AC-036` | §8.5 | Keyboard test | `Esc` behavior follows the priority ladder and closes one layer only. | `Esc` closes multiple layers or wrong layer. |
| `D-AC-037` | §8.5 | Focus test | Focus restoration follows the fallback ladder when the invoking element no longer exists. | Focus lost or restored to an undefined target. |
| `D-AC-038` | §8.6 | Grid keyboard test | Grid keyboard behavior is defined for every key chord in every declared mode. | Key chord delegates to unspecified adapter behavior. |

### 18.5 Component and surface criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-040` | §12.1 | Component audit | Every component family in §12.1 has a closed variant set. | Component family or variant open-ended. |
| `D-AC-041` | §12.2 | Component fixture | Component-state matrix contains one row or inherited rule for every required component variant. | Required variant has no visual, accessibility, activation, focus, or precedence mapping. |
| `D-AC-042` | §12.3 | Component fixture | Compound-state precedence cases resolve deterministically. | Compound state renders inconsistently. |
| `D-AC-043` | §13 | Surface audit | Every required surface has one row in the per-surface design table. | Required surface missing or duplicated. |
| `D-AC-044` | §13 | Surface audit | Optional surfaces have omission behavior. | Optional surface omission ambiguous. |
| `D-AC-045` | §13 | Owner-boundary audit | The per-surface design table does not define Core-owned field membership, write rules, or route behavior. | Design table becomes product behavior owner. |
| `D-AC-046` | §12.8 | Empty-state fixture | Empty-state behavior is defined for create-permitted and create-forbidden cases. | Empty state lacks reason or action rule. |

### 18.6 Accessibility criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-050` | §14.1 | Accessibility report | Every required UI state has a non-color cue and accessible representation. | Color-only state exists. |
| `D-AC-051` | §3.11 and §14.1 | Accessibility report | Icon-only controls have accessible names and visible focus. | Icon-only control lacks name or focus. |
| `D-AC-052` | §6.3 | Reduced-motion fixture | Reduced-motion behavior is defined and testable. | Motion required to perceive state. |
| `D-AC-053` | §14.2 | Live-region test | The live-region matrix maps each listed event to `polite`, `assertive`, or no live announcement. | Event announcement behavior undefined. |
| `D-AC-054` | §14.1 | Accessibility report | Row accessible names include human-readable surface context and are not raw `record_id` alone. | Row accessible name uses only ID or implementation coordinate. |
| `D-AC-055` | §14.3 | Contrast report | Contrast pair matrix covers every token pair used in required fixtures and forbids unlisted pairings. | Fixture uses unlisted pair or failing ratio. |

### 18.7 Visual-fixture execution criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-060` | §15.1 and §15.2 | Visual fixture metadata | Every `D-VFIX-*` fixture declares viewport, zoom, density, theme, scroll normalization, dynamic masks, crop rule, diff predicate, and artifact naming. | Missing fixture execution field. |
| `D-AC-061` | §15.1 | Visual fixture metadata | Dynamic fixture data is seeded or masked. | Actor names, timestamps, IDs, cursor positions, or local browser defaults are unmasked. |
| `D-AC-062` | §15.1 | Evidence-class audit | Fixture evidence is classified as design evidence, not claim-bearing benchmark evidence. | Fixture evidence is represented as Core 05 claim evidence without Core 05 compliance. |
| `D-AC-063` | §15.2 | Visual fixture registry | Every fixture row has exact viewport dimensions, not band-only declarations. | Fixture viewport is open-ended. |
| `D-AC-064` | §7.1 and §15.2 | Visual fixture review | `D-VFIX-001` captures the fixed first viewport for the default Timeline workbook shell with top-bar query controls, compact view bar, compact grid, row gutter, header affordances, selected row, focused Summary cell, row-context inspector, status strip, Core 01 default Timeline fields, and no admin/control card stack above the grid. | The fixture captures a dashboard/admin-card layout, uses non-Core default Timeline columns, or lacks a required workbook shell region. |

### 18.8 Boundary criteria

| ID | Requirement target | Evidence artifact | Pass condition | Failure condition |
| --- | --- | --- | --- | --- |
| `D-AC-070` | §16.2 | Boundary review | Each non-goal has observable omission behavior. | Non-goal lacks omission behavior. |
| `D-AC-071` | §16.3 | Boundary review | Future-profile candidates do not define current behavior. | Future profile is treated as current requirement. |
| `D-AC-072` | §17.2 | Boundary review | The document forbids behavior inference from labels, row order, SQL names, projection names, vendor grid coordinates, component names, and styling classes. | Implementation coordinate becomes design authority. |
| `D-AC-073` | §16.1 | Owner-boundary review | No added section creates route, schema, authorization, lifecycle, storage, or Core conformance behavior. | Design document defines Core-owned behavior. |
