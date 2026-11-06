import type { NetworkFlowTable } from "../services/networkFlowContractAdapter";
import type { TableAuthority } from "./networkFlowTableOperation";
export const tableIncidentId = "11111111-1111-4111-8111-111111111111";
const tableActorId = "22222222-2222-4222-8222-222222222222";
export const tableAuthority = (): TableAuthority => ({
  incidentId: tableIncidentId,
  actorId: tableActorId,
  sessionIdentity: "session-1",
  role: "admin",
  open: true,
  available: true,
  profileAvailable: true,
  availabilityTag: { epochId: "epoch", generation: 1n },
});
export function tableFixture(
  id = "a",
  version = 1,
  name = "Flows",
): NetworkFlowTable {
  return {
    network_flow_table_id: `nft_${id.repeat(32)}`,
    incident_id: tableIncidentId,
    display_name: name,
    table_version: version,
    table_status: "active",
    committed_source_id: "33333333-3333-4333-8333-333333333333",
    source_interpretation_sha256: "d".repeat(64),
    source_raw_cell_scalar_limit: 4096,
    source_content_sha256: "a".repeat(64),
    source_filename_display: "flows.csv",
    source_filename_digest: "b".repeat(64),
    source_filename_digest_key_id: "key-1",
    mapping_fingerprint: "c".repeat(64),
    source_profile_id: "cisco_sna_netflow_csv_v2",
    counter_semantics: "source_reported_flow_total",
    parser_profile_id: "rfc4180_headered_csv_v1",
    row_count_accepted: 1,
    row_count_rejected: 0,
    diagnostics_truncated: false,
    created_by_user_id: tableActorId,
    created_at: "2026-09-09T20:00:00Z",
    updated_at: "2026-09-09T20:00:00Z",
    deleted_at: null,
  };
}
