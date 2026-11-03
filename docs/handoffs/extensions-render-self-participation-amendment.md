# Reporting render self-participation amendment

Status: reviewed and adopted under standing maintainer authorization; adoption
is recorded as E-380 in the controlling
[tracker](extensions-module-refactor-tracker.md), before dependent implementation.

## Problem and retained production boundary

Reporting's built-in render-export participant validates a frozen model, encodes
it and returns those same bytes. Reporting immediately decodes and validates the
unchanged model. It requires `state_present=false`, while the shared emitter
matrix requires an empty result without invocation in that state. This is an
unnecessary self-callback rather than a transforming extension contribution.
Network Flow now has its own admitted typed source reader; it is independent of
this emitter and remains part of production composition.

Production inventory identifies one built-in emitter, one composition wrapper,
two worker call sites (release and preview), service/route dependency plumbing,
and no other implementation or supported build variant. Test-only invokers
simulate mutation, failure and blocking; their useful assertions must move to
current model admission and continuing renderer/publication boundaries.

## Concrete owner changes

Replace Reporting REQ-RPT-019a's emitter contract with these requirements:

- The Core-managed `snapshot_reporting` profile selects
  `snapshot_reporting_mode='no_participation'`. Its routes, Jobs, worker,
  renderer-runtime admission and Reporting state ownership remain unchanged.
  Rendering is Reporting-owned execution, not an extension source contribution.
- Release and authoritative preview execution admit the exact frozen
  `cartulary.reporting_export_model.v7` under Reporting's current derivation and
  schema-domain digest. They validate model identity, retained preparation and
  Reference Pack bindings, frozen redaction profile, schema, canonical bytes,
  field classification and item limits before rendering. A changed model cannot
  be admitted merely because a new digest is internally consistent.
- Canonical model bytes are bounded at 67,108,864, and model items at 1,048,576.
  The item count is the sum of sections, records, relationships, timeline events,
  subjects, diagrams, assets, support/source indices, presentation texts and all
  nested reporting blocks. Bounds count the actual model, without an invented
  invocation envelope. A rejected model produces no renderer call or output.
- Reporting retains the template's `export_model_materialization` deadline
  within the original total-attempt deadline and propagates cancellation.
  Validation returns an isolated admitted model so later mutation cannot alter
  the retained input. Existing failed-preparation, failed-render, atomic
  publication, proven-commit and indeterminate-commit rules remain authoritative.
- Remove `snapshot_reporting.render_export`, its `render_export_v1` participant,
  generic context/result wrappers and callback/invoker ports. Replace the
  paragraph following the current Job-input retirement rules with owned frozen
  model admission. Update RPT-AC-EXTENSIONS-001 to require claimed-only typed
  source admission, no calls for `no_participation`, and Reporting-owned model
  validation/rendering. Preserve unrelated diagram sequence participants.

Reporting REQ-RPT-019b must describe actual owner boundaries: Extensions condition
annotations govern real Extensions admission findings; ordinary Reporting model
and renderer validation uses its own closed schema/decision contracts and public
failure taxonomy. It must not create a fictitious extension-participant surface.

The server production and `cartulary_harness` build variants share the same
assembly; neither supplies another emitter. Remove the orphan
`snapshot_reporting.participant` condition surface and its unused condition.

In Extensions, `snapshot_reporting_mode='participant'` binds the current typed
Reporting source contract; `no_participation` registers no source. There is no
current generic Snapshot/Reporting `emit` operation. Retire its context/result
v1 schemas and the unused
`timeouts.extensions.snapshot_reporting_participant_seconds` setting. The removed
setting is unknown configuration, without an alias or implicit replacement.
Reporting retains its own manifest-owned deadlines; source reads retain the
caller's deadline. No transaction-participant timeout substitutes for either.

EXT-REQ-144 and Table 23-C must use explicit source selection rather than the
retired generic emission matrix. No selected source means no source read in any
claim/state condition; Core snapshot/report execution remains available. A
selected source requires participant mode, an admitted claim, current readiness
and an exact available result; otherwise Reporting rejects the selection under
its existing selected-projection error mapping. No-participation profiles never
supply source data. Neither retained unrelated profile state nor a claim causes
implicit source enumeration. This aligns the shared restatements with the
already adopted Network Flow selected-result boundary. Update the associated
acceptance/gate restatements without adding workflow labels.

Adopt `cartulary.extension_participant_specialization.v4` with only the existing
incident-portability specialization and its unchanged export/import operation
contracts. Remove the Snapshot/Reporting branch from the specialization schema,
Table 23-D/F, context/result definitions and generator/runtime handling. Reject
v3 specializations and retired emitter contexts/results at generation/admission.
The dedicated typed Reporting source contract remains v1. Update contract
inventories, contribution references and supporting/import sets accordingly;
retain the continuing Portability orchestration and transaction protocols.

## Implementation and cutover

Move continuing model identity, byte/item bounds, classification and isolation
checks into the private Reporting render preparation path. Delete the unused
emitter adapter and dependency threading instead of replacing them with another
injected callback. Keep actual renderer and Job publication ports. Preserve
semantic test row IDs while redirecting their selectors and assertions to these
continuing production boundaries. Deadline tests must use real owner deadlines
and continuing renderer/publication ports rather than a removed callback seam.

Remove authored self-contribution/implementation metadata and its imports,
update the current specialization schema, and regenerate through Make. Preserve
model v7, derivation v2, frozen Job/input/preparation formats, Recovery formats,
profile contract majors and all valid retained data. Generated registry and
binding digests change coherently. Old metadata/configuration is rejected;
there is no historical reader, conversion, environment reset or automatic data
deletion. Retired forms survive only as small negative fixtures where needed.

Review criteria: complete consumer/build-variant inventory; preserved current
render safeguards and immutable representations; no replacement generic
framework; no workflow-specific labels in normative owners; actual current
failure/deadline/isolation tests; current render/preview/release and source
lifecycle witnesses; generated parity, boundary checks and Markdown lint. The
separate Graph stored-content integrity correction remains required before
module readiness.
