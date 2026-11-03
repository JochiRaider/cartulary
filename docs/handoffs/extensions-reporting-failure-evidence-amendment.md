# Reporting failure evidence amendment

Status: adopted under the standing user authorization of 2026-11-02 after
concrete source/boundary review, Markdown lint and preservation review.
This handoff is a review artifact, not executable authority or verification input.

## Problem and selected boundary

A current capture/restore regression changes retained Graph bytes without
violating storage shape. Both render kinds reject the bytes, but their failure
writer decodes the same invalid recipe again. The Jobs runner eventually records
`job_handler_attempts_exhausted` without the required Reporting diagnostic.
Missing private input has the same dependency problem. Repeated decoding or a
partial parser for a rejected recipe would preserve that conceptual defect.

Reporting will own one independent render-failure journal. It will retain a
canonical diagnostic bound to independently retained Jobs admission, atomically
with the terminal Job and any justified failed-release state. Immutable render
input will no longer double as the failure-evidence container. Validated input
may supply snapshot/release facts; rejected or missing input may not.

## Concrete Reporting owner amendment

Add REQ-RPT-131 beside REQ-RPT-130 and reconcile REQ-RPT-117c, Table 24-A and
Table 23-B as follows:

- Release and Composition preview failures and cancellations retain exactly one
  immutable `cartulary.reporting_render_failure_record.v1` per Job. Reporting
  owns that record independently of private recipe/preparation presence and
  validity. Successful attempts have no failure record.
- The record is a closed object with exactly `schema_id`, `job_id`, `job_kind`,
  `incident_id`, `actor_user_id`, `admitted_at`, `request_sha256`, `input_binding`
  and `summary`. Every member is required; only `input_binding` is nullable. `schema_id` is the exact identifier
  above. UUIDs and the timestamp use the existing Core/Reporting canonical
  forms. `job_kind` is exactly `snapshot_reporting.release_create_v4` or
  `snapshot_reporting.composition_preview_v4`. `request_sha256` is the original
  Jobs normalized public-request digest, not a private recipe/model digest.
  `summary` is the complete canonical Table 23-B failure summary.
- `input_binding` is null when private input was not validated. Otherwise it is
  a closed object with required `admission_sha256` and nullable required
  `preparation_sha256`, binding the exact validated recipe and optional
  preparation. Both non-null values use the existing lowercase SHA-256 form.
  A retained non-null binding must still match the independently retained
  private digests and validated content. A non-null preparation digest justifies
  the failed-release row for a release Job; its absence forbids that row.
  Null `input_binding` forbids both source identities and a failed-release row.
  Unbound diagnostics do not claim that an existing preparation was validated.
- Job identity, kind, incident, initiating actor, timestamp and request digest
  come from the independently bound Jobs admission. Private recipe fields,
  renderer output and exception messages are not authorities for these facts.
  The record retains this binding through public-response expiry. Validation
  after expiry compares the still-retained Jobs identity/scope/actor/timestamp;
  it does not invent an expired request digest or rewrite the record.
- The record uses Reporting canonical JSON and its existing schema-domain hash:
  SHA-256 of the UTF-8 schema identifier, one LF byte, then the exact canonical
  record bytes. The complete encoded record is bounded at 2 MiB; the existing
  summary/issue limits remain in force. Canonical bytes and their digest are
  retained, not regenerated on replay.
- An absent required private input has `failure_code=retained_input_invalid`,
  `reason_code=retained_input_missing`. A structurally invalid, changed,
  mismatched or unsupported retained recipe/preparation has the same failure
  code and `reason_code=retained_input_mismatch`. Both use `release_state` and
  common Job code `release_render_failed`, nonretryable. No private values,
  paths, hashes, identifiers or error text become integrity-specific safe
  details. These failures do not consume the generic durable retry budget.
- A summary includes a snapshot or release identity only when its exact binding
  was validated. Otherwise both fields are null and
  `safe_details.pre_release=true`; an untrusted identity is not “known” for
  Table 23-B. `created_at` is the independently retained admission timestamp.
  No synthetic release or fabricated model/preparation facts are permitted.
- A validated retained preparation permits the existing failed-release outcome.
  Without that proof the journal and Job are sufficient; a release row must not
  be manufactured merely to retain an error. Authority denial before private
  work remains recordable without reading private case content. A preview never
  allocates a release identity.
- Private byte decoding, hashing, materialization and diagnostic serialization
  occur outside the Jobs terminal transaction. An owner-created candidate
  carries bounded validated metadata, immutable diagnostic bytes and the exact
  observed input version/presence. The final transaction revalidates the live
  execution and candidate bindings/presence/version, then atomically writes the
  journal, optional justified failed release and terminal Job. It does not
  decode the rejected recipe to decide whether its failure can be retained.
- Storage/transport failure is not evidence of input corruption. Cancellation,
  timeout, cleanup-failure precedence and indeterminate-commit fatal handling
  retain their existing contracts. A proven cancellation can retain its true
  cancellation diagnostic even when source identity is unavailable. Successful
  committed outcomes are immutable; exact failure replay compares the original
  record and never reconstructs missing evidence.
- A truthful failure record does not make corrupt/missing private input valid.
  Claim admission still rejects invalid required Reporting state. Structural
  Recovery preserves independent journal bytes and current binding proofs;
  it does not run inactive Reporting semantics or repair private input.

Add the failure code and its two reasons to Tables 23-E/F. The validation-summary
wire shape remains v1; only the explicitly adopted closed values and conditions
for nullable identity change. Snapshot Jobs keep their existing common failure
contract and do not acquire render-failure records.

## Core and current storage amendment

Core 01 REQ-01-634 references REQ-RPT-131 for Reporting terminal owner evidence;
Jobs retains exclusive execution, cancellation, terminal-transition and commit
proof ownership. Reporting selects its worker kind from the independently bound
Jobs definition rather than an untrusted private row.

Introduce Reporting-owned authoritative table `reporting_render_failure_records`
with `job_id` as primary key and restrictive reference to `jobs`, required bounded
canonical `record_json` bytes and required `record_sha256`. It is a required
current Recovery table, including while Reporting is unclaimed. No generic
cross-owner failure framework or second diagnostic representation is introduced.

Retire `reporting_job_payloads.render_failure_summary` and
`render_failure_sha256` and their pair constraint. Preserve all earlier migration
files. The new transactional Up migration locks the old owner table and rejects
any non-null legacy diagnostic member before changing schema. It does not copy,
convert, delete or reinterpret existing diagnostic history. Compatible empty
columns may be removed. Historical migration Down behavior remains confined to
the Core-owned disposable test cleanup contract.

Core's catalog facts become 148 total tables and 116 required tables. Authored
catalogs, fixtures, schema head, current generation and dependent binding digests
must be regenerated as one coherent cutover. Prior incompatible backup identity
pairs reject before mutation. Runtime uses the selected frozen catalog for exact
membership; no new independently maintained runtime cardinality is added.

## Review and validation exits

The correction must cover current capture/restore and live execution with missing
private input, altered recipe bytes, independent binding mismatch and altered
preparation. Check exact nonretryable diagnostics, atomic journal/Job retention,
no rendering or success proof, no rewriting rejected bytes, stale candidates,
missing/changed terminal records, public-response expiry, cancellation races,
authority denial and permitted committed outcomes. Existing valid failures and
prepared/unprepared behavior retain their meaning.

Generation and schema checks must include the new packaged record definition,
exact Recovery membership and rejection of the retired diagnostic shape. Prove
that a nonempty old diagnostic blocks migration without mutation and that the
current empty cutover works. Preserve current snapshot/binding, staged-inventory,
immutable-catalog and route-receipt regressions. Real process/browser qualification
remains in its separately required workstreams.
