# Cartulary Report Builder Implementation Stack Guide

**Document class:** Implementation-support guide  
**Freshness anchor:** 2026-07-05  
**Owner references:** `docs/report-composition-nlspec.md`, `docs/reporting-subsystem-nlspec.md`,
Core 00 through Core 04, and `docs/guides/cartulary-dev-guide.md`.

This guide recommends a browser implementation stack for the report-composition builder. It does not
define composition semantics, validation authority, preview authority, released output bytes, digest
calculation, redaction behavior, graph projection binding, or render sandbox behavior.

The Report Composition NLSpec owns the authoring-side data contract and builder-facing route behavior.
The Reporting Subsystem NLSpec owns render effects, Mermaid and Slidev generation, redaction,
sandboxing, preview attempts, release output, bundle hashing, and byte-affecting toolchain snapshots.
Core documents prevail over this guide whenever there is a conflict.

## 1. Authority Boundary

A builder implementation using the libraries in this guide is judged by the emitted route requests,
schema objects, validation calls, and preview calls. It is not judged by internal React component shape,
editor state, canvas state, drag library state, or browser-local render artifacts.

The following implementation rule is fixed for this guide:

| Surface | Library may own | Library MUST NOT own |
| --- | --- | --- |
| Diagram selection canvas | Interaction, viewport behavior, selection gestures, route-editing gestures, node rendering for the picker | Graph mutation semantics, graph projection identity, unnormalized layout state, Mermaid IDs, released diagram bytes |
| Authored text editor | Local editing state, input constraints, subject-reference insertion UI | Persisted rich-text state, resolved subject labels, disclosure partition inference, authored-text release permission |
| Drag ordering | Pointer and keyboard gestures, sortable-list mechanics, insert target affordances | Operation ordering semantics, anchor identity, draft lifecycle, optimistic-version conflict behavior |
| Client validation | Immediate client lint and generated TypeScript types | Conformance authority, release admissibility, canonical digest validation, semantic anchor resolution |
| Browser preview | Responsiveness and local approximation | Approval evidence, release evidence, fixture evidence, Reporting preview authority, rendered output bytes |

## 2. Recommended Stack

| Builder surface | Recommended library | Persisted output |
| --- | --- | --- |
| Diagram selection and layout canvas | `@xyflow/react` plus `@dagrejs/dagre` | `composition_diagram_decl.v1` with closed `diagram_selection_rule.v1` and optional `composition_diagram_layout.v1` |
| Authored text editor | Tiptap over ProseMirror | `authored_text.v1.body` with `{{subject:<stable_subject_ref>}}` placeholders |
| Deck, section, and block ordering | `@dnd-kit/core` and `@dnd-kit/sortable` | `composition_op.v1` objects or a versioned full-array draft replacement |
| Client schema validation | Ajv | Non-authoritative validation hints only |
| Type generation | `json-schema-to-typescript` | Generated TypeScript types downstream of owner-controlled schemas |
| Live diagram or deck preview | Mermaid in browser for auto layout; structured SVG preview for manual layout | Non-authoritative local approximation only |

Exact package versions MUST be pinned in repo-control files. This guide MUST NOT be used as evidence
that a package is installed or current. Before changing dependencies, revalidate package presence,
versions, licenses, peer dependencies, and lockfile state against `package.json`, workspace manifests,
and lockfiles.

## 3. Diagram Selection And Layout Canvas

`@xyflow/react` is acceptable for the diagram-node picker and manual layout editor because it is
designed for node-based React interfaces and provides panning, zooming, selection, dragging, and
React-component node rendering. It is not acceptable as a source of graph, projection, validation,
preview, or released-output semantics.

The canvas MUST read completed projection output or previewable projection-derived objects and emit
only closed `diagram_selection_rule.v1` data, such as `explicit_refs`, `neighborhood`,
`timeline_sequence`, or `all_with_bounds`. When manual layout is enabled, node dragging MUST emit
normalized `composition_diagram_layout.v1.node_positions[]`, and route editing MUST emit normalized
`composition_diagram_layout.v1.edge_routes[]`.

The builder MUST NOT persist any of the following in `cartulary.report_composition.v1`:

- React Flow node IDs,
- React Flow edge IDs,
- handle IDs,
- viewport state,
- screen-space x/y coordinates outside `composition_diagram_layout.v1`,
- unnormalized layout positions,
- selection-box state,
- DOM IDs,
- SVG IDs.

Node and edge creation controls MUST be disabled unless the resulting gesture compiles to an allowed
selection rule over existing projection output. Arbitrary graph mutation is out of scope.

Selection MUST target opaque graph or projection references accepted by the Reporting and Composition
contracts. Selection MUST NOT target generated Mermaid IDs, rendered DOM nodes, rendered SVG nodes, or
visual labels.

Manual layout MUST target retained selected vertex and edge refs exactly as the owner NLSpecs define
them. It MUST NOT create graph vertices, create graph edges, remove graph edges, change semantic edge
endpoints, or use React Flow IDs as persisted targets.

`@dagrejs/dagre` is preferred for local picker layout over the older unscoped `dagre` package. Dagre
output is a local suggestion only. Persisted manual layout is valid only after it is normalized into
`composition_diagram_layout.v1` and passes server validation.

`elkjs` SHOULD NOT be introduced under the current permissive-runtime-dependency preference unless the
repository dependency policy explicitly accepts EPL-2.0 for this use.

## 4. Authored Text Editor

Tiptap over ProseMirror is acceptable for authored presentation text because ProseMirror schemas can
constrain allowed node and mark structure. ProseMirror JSON and Tiptap extension state are transient
editor state only.

The builder MUST persist only `authored_text.v1` objects. It MUST NOT persist ProseMirror JSON,
Tiptap extension metadata, editor transaction logs, comment state, collaboration state, or document
history state.

The subject-reference node MUST serialize to the exact placeholder grammar:

```text
{{subject:<stable_subject_ref>}}
```

The editor MUST NOT serialize resolved display names, hostnames, account names, party labels, emails,
IP addresses, token display values, or other raw subject values as substitutions for subject refs.

Role-specific NFC normalization, control-character rejection, LF rules, and before/after-placeholder
limits for `title_override`, `speaker_notes`, and `authored_text` SHOULD be enforced locally for
responsiveness and MUST be rechecked by server validation. The UI MUST require an explicit
`disclosure_partition_ref`; it MUST NOT infer the partition from current user, role, recipient, section,
slide, or preview context.

Tiptap collaboration, comments, managed document history, AI features, Hocuspocus, and Yjs co-editing
are future-only for this builder surface unless the owner specifications are revised.

Lexical remains a credible alternate editor implementation. If selected, it has the same observable
serialization contract and the same prohibition on persisting editor-native state.

## 5. Drag Ordering

`dnd-kit` is acceptable for section, slide, block, and drag-to-insert gestures because those gestures
can compile to Composition-owned operations without making the drag library a layout authority.

Every committed drag result MUST compile to exactly one supported `composition_op.v1` operation or to
a full-array draft update that preserves server-owned semantics.

The UI MUST translate visual drag targets into semantic anchors:

- `section_anchor`,
- `record_anchor`,
- `block_anchor`,
- `diagram_anchor`.

The builder MUST NOT persist slide ordinal, visual index, DOM order, CSS selector, grid coordinate,
dnd-kit sortable ID, or rendered slide ID as an operation target.

Reordering MUST preserve the composition spec's operation-order semantics and draft/version lifecycle.
If an update replaces `deck_ops[]`, `diagram_decls[]`, or `authored_texts[]`, the client MUST treat the
replacement as a versioned draft mutation, not as silent local state.

## 6. Client Validation And Types

Ajv is acceptable for immediate client feedback over JSON Schemas generated from owner-controlled
contracts. Ajv validation is client lint only. It MUST NOT be represented as conformance authority,
release admissibility, preview authority, or approval evidence.

`json-schema-to-typescript` is acceptable for generated TypeScript types. TypeScript interfaces MUST
remain downstream of JSON Schemas and MUST NOT become the source of truth.

Client validation cannot prove every condition required by the owner specifications. Server validation
and Reporting behavior remain authoritative for at least:

- duplicate JSON member rejection when input begins as raw JSON bytes,
- semantic anchor resolution,
- graph projection binding,
- release context,
- redaction-profile permission,
- `allow_authored_presentation_text`,
- tokenizable-subject placeholder resolution,
- canonical byte digest validation,
- authoritative preview admissibility,
- render sandbox and toolchain behavior.

If a client surface accepts raw JSON editing, duplicate-member detection MUST occur before ordinary
JSON object materialization because duplicate keys are no longer observable after normal `JSON.parse`
semantics.

## 7. Browser Mermaid Preview

Mermaid in browser is acceptable only as a non-authoritative local approximation for
`layout_mode='auto'`.

The browser MUST NOT offer raw Mermaid authoring as persisted composition input. Browser Mermaid
output MUST be derived from the same composition-data interpretation path where feasible, or else
clearly treated as a local approximation that cannot be used for approval, release, fixture, or digest
evidence. A local scratchpad MAY exist only when save converts into `composition_diagram_decl.v1` and
`composition_diagram_layout.v1` and discards the raw Mermaid text.

The exact Mermaid version used for authoritative output is owned by the Reporting render toolchain
snapshot. Browser Mermaid package selection MUST NOT change released output bytes.

Browser Mermaid rendering MUST NOT relax Reporting's serializer and sandbox constraints. The builder
MUST NOT admit Mermaid init blocks, comments, raw HTML, labels that require HTML, arbitrary renderer
syntax, click actions, unsafe SVG constructs, network reads, package-resolution behavior, or
renderer-local syntax as persisted composition source.

For `layout_mode='manual'`, local live preview MUST render from the structured layout model, not from
raw Mermaid text. The preview may use browser SVG or canvas, but persisted state remains the closed
composition layout schema and authoritative preview remains Reporting-owned.

Authoritative preview MUST call the composition preview route. That route delegates to Reporting as an
`internal_draft` render attempt and does not create immutable release bytes.

## 8. Builder Acceptance Criteria

A builder following this guide is acceptable only if all criteria in this section pass.

| ID | Criterion |
| --- | --- |
| `RB-AC-001` | No persisted composition document contains React Flow IDs, ProseMirror JSON, Tiptap metadata, dnd-kit sortable IDs, DOM selectors, CSS selectors, Mermaid IDs, SVG IDs, viewport positions, unnormalized layout coordinates, or generated source bytes. |
| `RB-AC-002` | Every committed builder control maps to exactly one declared `composition_op.v1` operation, `authored_text.v1` object, `composition_diagram_decl.v1` object, `composition_diagram_layout.v1` object, validation request, or authoritative preview request. |
| `RB-AC-003` | A full builder UI provides authoring paths for every current operation kind, every authored-text role, diagram declarations, manual diagram layout where exposed, server validation, and authoritative preview. A partial builder may exist but MUST NOT claim full builder UI conformance. |
| `RB-AC-004` | Client lint may block obviously invalid input early, but only the server validation summary may mark a composition valid for preview, release, or approval workflow. |
| `RB-AC-005` | Local preview may improve responsiveness, but authoritative preview uses the server preview route and Reporting's `internal_draft` render path. |
| `RB-AC-006` | Invalid states are made unrepresentable where feasible, and server validation remains mandatory for semantic anchor resolution, graph projection binding, manual layout target and coordinate validity, disclosure partition validity, authored-text external-release permission, placeholder resolution, digest validation, sandbox behavior, and toolchain behavior. |

## 9. Version And Source Notes

These notes are implementation-support inputs, not normative behavior. They were reviewed for freshness
against external sources with the intended anchor date of 2026-07-05. Revalidate all volatile package
facts before changing dependency pins.

| Topic | Source note | Implementation consequence |
| --- | --- | --- |
| React Flow | The React Flow site describes the library as MIT-licensed open source and describes built-in dragging, zooming, panning, selection, and React-component nodes. GitHub release data showed `@xyflow/react@12.10.2`, `12.11.0`, and `12.11.1` in the release feed. | Pin an exact `@xyflow/react` version. Use it for selection and manual-layout gestures only after normalizing persisted data into owner schemas. |
| Dagre | The `dagrejs/dagre` repository describes directed graph layout for JavaScript and uses a permissive license. | Prefer `@dagrejs/dagre` for transient local layout. |
| ELK | Package metadata for `elkjs` identifies EPL-2.0. | Avoid under the current permissive-runtime-dependency preference unless dependency policy changes. |
| ProseMirror and Tiptap | ProseMirror documents schema-constrained document structure. npm reported `@tiptap/core` `3.27.2` during review. | Pin an exact Tiptap version; do not rely on `latest`. |
| dnd-kit | The project describes a React toolkit for drag-and-drop interfaces and is MIT-licensed. | Use it for gestures that compile to Composition-owned operations. |
| Ajv | The project supports JSON Schema validation, including draft 2020-12. | Use it for client lint only. |
| json-schema-to-typescript | The project compiles JSON Schema to TypeScript declarations. | Generate TypeScript downstream from owner schemas. |
| Mermaid | Mermaid documents `securityLevel` behavior, including `strict` handling. | Browser Mermaid remains a local approximation for auto layout; Reporting owns authoritative Mermaid and manual-layout SVG output. |

Sources:

- [React Flow](https://reactflow.dev/)
- [xyflow releases](https://github.com/xyflow/xyflow/releases)
- [dagrejs/dagre](https://github.com/dagrejs/dagre)
- [elkjs package metadata](https://www.jsdelivr.com/package/npm/elkjs)
- [ProseMirror guide](https://prosemirror.net/docs/guide/)
- [`@tiptap/core` npm package](https://www.npmjs.com/package/%40tiptap/core)
- [dnd-kit](https://github.com/clauderic/dnd-kit)
- [Ajv](https://github.com/ajv-validator/ajv)
- [json-schema-to-typescript](https://github.com/bcherny/json-schema-to-typescript)
- [Mermaid `securityLevel`](https://mermaid.js.org/config/schema-docs/config-properties-securitylevel.html)
