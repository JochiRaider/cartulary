# IR incident fixture corpus

The maintained originals live in `fixtures/ir-incident/v1`. Its typed manifest
binds the five newly adopted files and pins expectation revision 1 of
`network-flow-orientation` by its manifest digest. The three Cisco exports are
read from that dependency; there are no new copies of their CSV bytes.

The current implementation supports complete Network Flow incident portability:
public major 10, retained state 7, Core source-port major 3, Incident Bundle 6,
and Recovery generation `recovery.incident_portability.v1`. Network Flow is
included automatically when retained state exists. The destination must have
compatible support enabled; import never enables it or silently drops data.
ZIP is exported; ZIP, TAR and TAR.GZ carry the same version-6 member format.
See the handoff for exact passing development evidence and final qualification.

## Inputs and interpretation

| Original | Logical data records | Supported use |
| --- | ---: | --- |
| `timeline.csv` | 1,500 | Imports file upload to Timeline v2 |
| `hosts.csv` | 30 | Imports file upload to Hosts v1 |
| `identities.csv` | 20 | Imports file upload to Identities v1 |
| `aliases.csv` | 222 associations | Host/Identity `add_alias` collection operations |
| `timeline.tsv` | 1,500 | Clipboard input; not an upload format |
| dependency `day1.csv` | 2,000 | Cisco SNA Flow Search CSV v2 |
| dependency `day2.csv` | 2,000 | Cisco SNA Flow Search CSV v2 |
| dependency `overlap.csv` | 500 | Cisco SNA Flow Search CSV v2 |

The manifest supplies explicit source-column-to-field mappings. Timeline uses
the ten visible text fields; the host and identity headers are stable writable
field keys. The integration test approves those mappings through Imports.
No source operational text creates inferred relationships, evidence records,
indicator records, or authenticated users.

The CSV and TSV contain the same decoded cells. Quoted multiline text, embedded
tabs, CRLF separators and decomposed Unicode are significant. Go's standard CSV
reader folds quoted CRLF; the independent test reader shields quoted CR bytes
before using that library and restores them afterward. The product lexer is
not used by the expectation builder. Logical source record 183,
`NSR-EV-B0258`, explicitly exercises CRLF and decomposed Unicode. Frozen cases
use logical data-record numbers and source evidence-reference identities,
never physical line numbers or grid positions.

Git line-ending conversion is disabled for these originals, the orientation
CSV corpus and the original Cisco example. The existing checkout had CRLF
copies whose hashes matched the orientation manifest, while its committed
blobs were LF-normalized. Preserving the supplied bytes in Git repairs that
fresh-checkout discrepancy without changing source-profile semantics.

## Independent expectations and admission

`internal/testutil/irfixtures/expectations.go` recomputes source facts using
standard-library CSV, timestamps, string operations and counters. It imports
no product normalization, identity, matching or graph code. The reviewed
`expected.json` is authored input. Tests compare it and never overwrite it.
The initial expected facts were also calculated independently with Python's
standard CSV and datetime libraries. The existing orientation corpus retains
its independent per-observation and per-edge expectations.

Go and browser admission validate the same closed manifest schema and member
hashes before product setup. Owner-specific manifest handling uses shared,
small file-admission primitives for confined paths, regular files, verified
copies and read-only materialization. Missing or changed sources, expectations
or dependency members fail admission. The helpers contain no scenario engine.

Go corpus consumers retain admitted inputs and `comparison.json` beneath the
selected harness run. Browser admission retains its copies in the test's
output directory and attaches the manifest identity. Test discovery never
executes admission. Catalog `fixture_input_roots` closes fixture bytes and
dependency expectations into cache identity; executable source indexing
continues to exclude documentation.

The numerical anchors are 1,450 populated consistent time pairs at offset
−240 minutes and 50 empty pairs; 222 alias associations with 220 distinct
values; 4,500 flow observations with 4,000 distinct source records;
3,903,637,992 bytes and 4,412,353 packets; 667 server-first observations;
284 missing-orientation observations; and 651/660/324 per-table edges.
The 500 overlap observations are retained separately, including provenance.
These volumes make no performance qualification claim.

## Synthetic supplement

`supplement/records.json` identifies synthetic workflow inputs and explicit
mentions for `research-files` and `nsr-admin`. The two harmless text blobs are
used for attachment/replacement-history coverage; the two tasks exercise
dependency/completion behavior. The connected public scenario creates the
pending evidence request, linked note, decision, assessment history, explicit
indicator observation and flow binding, party reference, saved view and report
composition. Synthetic descriptions remain separate from original observations.
The composition selects the exact day-1 graph, freezes a snapshot, renders an
approved release, and verifies retained files after later imports and transfer.

The shared aliases remain ambiguous: `research-files` names NSR-FS-01 and
NSR-FS-02; `nsr-admin` names the Alvarez and Bergström identities. The original
Timeline contains neither ambiguous alias. Neither graph adjacency nor these
synthetic descriptions establish compromise or evidence validity.

## Verification and operational boundary

Discover exact selections with `make task-guide` and `make target-plan`, then
use `make test-slice` with a fresh explicit `CARTULARY_TEST_RUN_ID`. Primary
catalog rows are:

- `module.imports.unit.ir_corpus_admission`
- `module.imports.integration.ir_corpus_core`
- `module.imports.browser.ir_corpus_import`
- `module.timeline.unit.ir_clipboard_boundaries`
- `module.timeline.integration.ir_public_clipboard`
- `module.networkflow.unit.orientation_corpus`
- `module.networkflow.integration.orientation_corpus`
- `web.architecture.boundary_support.ir_fixture_admission`
- `module.incidentbundles.integration.network_flow_portability`
- `module.incidentbundles.integration.network_flow_historical_portability`
- `module.incidentbundles.integration.network_flow_export_snapshot`
- `module.incidentbundles.integration.large_member_public_transfer`
- `module.recovery.browser.ir_transferred_investigation`
- `module.recovery.browser.ir_restored_investigation`

The clipboard integration checks public 499/500-row writes and exact replay;
501 rows and the full TSV reject without partial records. Core import checks
all fields, pagination, time conversion and alias ambiguity. The full browser
journey uploads the original Core and all three flow CSVs and approves mappings.
Destination and restored browser journeys initialize their own incidents,
inspect exact selected graph history, and download retained approved report files.

Committed-source inspection retains **exact uploaded bytes**, including unmapped
content. Source authorization remains target-specific. Workflow cleanup does not
remove those immutable sources. The historical actor roster is attribution only:
it grants no accounts, authentication, permissions or memberships.

Graph refresh is explicit. Retired graphs and graphs whose selected tables were
soft-deleted retain their last selected results; retirement prevents refresh and
does not override Reporting's selection rules. Import rebuilds the selected result
without manufacturing a completed job or resolving new all-active sources.

See the [implementation handoff](../handoffs/ir-fixture-adoption-implementation.md)
for source identities, retained results, comparisons and qualification status.
All test resources are harness-owned and isolated; no live environment is reset.
A new deployment must be initialized explicitly with empty current-format storage.
Keep the prior executable and its data together for historical access. Existing
retained state, older bundles and older Recovery generations are rejected before
mutation; no backfill, conversion, dual reader or in-place downgrade is supplied.
Transfer is incident-scoped and excludes operational identity and credentials.
Recovery retains its separate deployment inventory, key requirements, session
invalidation and serving-admission lifecycle.
