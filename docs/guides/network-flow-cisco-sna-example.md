# Cisco Secure Network Analytics upload example

Use the [original synthetic CSV](../../fixtures/network-flow/NF-FIX-029-cisco-sna-flow-search/source/cisco_sna_netflow_example_1200_rows.csv)
to import a day of flows and build a saved Network Analysis graph. Its 116
headers come from a Cisco Secure Network Analytics export; its data is
synthetic. The file is preserved exactly as supplied, including timestamps
such as `2026-10-29T00:00:07.117+0000` and blank ICMP ports.

## Upload and save a graph

1. Open an active incident in a deployment with Network Flow enabled. Select
   **More views → Network Analysis** and upload the linked CSV.
2. In the mapping panel, select **Cisco Secure Network Analytics Flow Search CSV**.
   This selects the export's nine mapped columns, compact-offset timestamp mode
   and millisecond precision. Optionally name the table **Cisco SNA synthetic example**.
3. Keep **Preserve unmapped raw** so the remaining 107 columns stay with each
   row's source provenance. Preview the mapping: the first 50 records should
   all be accepted, with no diagnostics.
4. Choose **Approve and apply**. The complete import should accept **1,200 rows**,
   reject **0**, and open the resulting table.
5. Switch to **Graph**, with the active table selected and no filters or time
   buckets. Expect **173 endpoint vertices** and **944 aggregated edges**.
6. Open **Saved graphs → Save current graph**, enter a name such as
   **Cisco SNA — 29 October 2026**, then choose **Save graph**. Wait for
   **Materialization succeeded**.
7. Open the same incident in a fresh browser tab. Select **More views → Network
   Analysis → Graph → Saved graphs**, then select the saved graph. Its persisted
   graph and projection identities, bindings, query and complete results should
   match the saved result without uploading or materializing again.
8. Choose **Refresh**, confirm **Refresh graph**, and wait for success. This
   creates a separate job and advances the generation while preserving the
   semantic query and totals. An unchanged immutable projection can be reused.

For a smaller exploration, filter the destination port to `53` to inspect
peer-side DNS traffic, or select one endpoint and inspect its contributors.
Remove filters before checking the complete-example totals below. This is an
example dataset, not evidence of suspicious activity or a vendor-completeness
claim.

## Interpret the graph

The profile maps **search subject → peer**. A subject can be a client or a
server; an arrow does not establish client/server roles, packet direction, or
which endpoint sent the connection's bytes. Each flow contributes
`connection.transferBytes` and `connection.transferPackets` exactly once.
The graph groups by subject IP, peer IP, IP protocol and peer port; it does not
group by source port. Multiple rows can therefore contribute to one edge.

| Export header | Network Flow value |
| --- | --- |
| `firstActiveTime` | Flow start, converted to UTC |
| `lastActiveTime` | Flow end, converted to UTC |
| `searchSubject.ipAddress` | Source endpoint |
| `peer.ipAddress` | Destination endpoint |
| `searchSubject.portProtocol.port` | Source port; blank becomes null |
| `peer.portProtocol.port` | Destination port; blank becomes null |
| `searchSubject.portProtocol.protocol` | IP protocol |
| `connection.transferBytes` | Connection bytes |
| `connection.transferPackets` | Connection packets |

The two protocol columns agree in this example. The peer protocol column,
application labels, hostnames, endpoint-relative interfaces, NAT values and
other export columns remain raw provenance. Endpoint-relative interfaces are
not reinterpreted as fixed ingress/egress interfaces. Blank ports are unknown
or inapplicable, distinct from port zero.

## Expected results and automated proof

| Measure | Expected value |
| --- | ---: |
| Imported records | 1,200 |
| Rejected records | 0 |
| Endpoints | 173 |
| Default graph edges | 944 |
| Sum of edge flow counts | 1,200 |
| Connection bytes | 71,114,484,212 |
| Connection packets | 94,893,741 |
| TCP / UDP / ICMP records | 967 / 203 / 30 |
| ICMP records with null ports | 30 |

The timestamps span `2026-10-29T00:00:07.117Z` through
`2026-10-29T23:56:25.85Z`. The original file is 1,275,537 bytes with SHA-256
`b03c98cf0124f426e05d1b87f6ca1cf38c5f8a145e22c7dec6d3f527b99e70c3`.

Frozen bundle `NF-FIX-029-cisco-sna-flow-search` supplies the unchanged upload,
explicit mapping and independently calculated expectations. Its backend test
uses real Imports approval, apply, persistence, replay and Graph Projection.
It compares all persisted row members, using a complete sorted-key JSON digest
for each row's raw provenance, complete public pages and authenticated cursors,
inspection/preview/mapping/table contracts, and full graph and adapter input
against independent expectations. Targeted negative controls reject altered
metadata, bindings, defaults, identities, ordering and missing/extra members.
The browser test uploads the same admitted bytes, checks counts and counter
sums, creates a saved graph, restores its complete persisted declaration and
result in a fresh page, then explicitly refreshes it. Both
retain fixture input hashes and comparison evidence beneath their selected
harness run roots. Tests never read this guide.

The exact registered rows are:

- `module.networkflow.integration.original_cisco_flow_search_export_imports_1200_r_6e0167c488`
- `module.networkflow.browser.cisco_flow_search_upload_builds_and_saves_the_co_06fe25b179`

Use `make task-guide OWNER=module.networkflow` and select these rows through
`make test-slice`, with a fresh explicit `CARTULARY_TEST_RUN_ID` for each run.
The [controlling tracker](../handoffs/networkflow-module-refactor-tracker.md)
records verification outcomes and compatibility details.

## Compatibility

The separate `cisco_sna_flow_search_csv_v1` profile accepts compact numeric UTC
offsets through `cisco_sna_datetime_v1`; it does not relax the original NetFlow
profile's RFC3339 grammar or reinterpret existing approved mappings. Existing
identities, historical migrations and database columns remain unchanged.
New imports require a server and browser supporting this profile. Preserve that
support when operating databases containing these imports; reverting the source
alone is appropriate only before creating durable imports with the new profile.
No database reset, data conversion, application population or deployment is part
of adopting this repository example.
