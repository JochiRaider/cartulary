import { vi } from "vitest";
import { createObservationTransport } from "../workbook/adapters/createObservationTransport";
import type { ObservationSource } from "../workbook/features/indicators/observationModel";
import type {
  ObservationAuthority,
  ObservationIntent,
  ObservationReadPort,
  ObservationReceipt,
  ObservationTransportPort,
} from "../workbook/features/indicators/observationOperation";
import { WorkbookObservationOwner } from "../workbook/features/indicators/WorkbookObservationOwner";
import type { IndicatorObservation } from "../workbook/mutations/workbookMutationCommandPorts";

export const observationAuthority: ObservationAuthority = {
  actorId: "40000000-0000-4000-8000-000000000001",
  incidentId: "10000000-0000-4000-8000-000000000001",
  sessionIdentity: "session-one",
  role: "editor",
  closed: false,
};
export const observationSource: ObservationSource = {
  incidentId: observationAuthority.incidentId,
  viewSchemaId: "cartulary.view.timeline.v2",
  recordId: "20000000-0000-4000-8000-000000000001",
  rowVersion: 4,
  fieldKey: "timeline.raw_activity_text",
  text: "  α.example\r\nα.example 😀 e\u0301  ",
};
export const testObservation: IndicatorObservation = {
  observation_id: "30000000-0000-4000-8000-000000000001",
  incident_id: observationAuthority.incidentId,
  source_record_id: observationSource.recordId,
  source_field_key: observationSource.fieldKey,
  origin_kind: "manual_entry",
  origin_locator:
    "record:20000000-0000-4000-8000-000000000001:field:timeline.raw_activity_text:bytes:2-12",
  observed_text: "α.example",
  parsed_indicator_type: "domain_name",
  normalized_candidate: "α.example",
  resolution_status: "unresolved",
  resolved_indicator_record_id: null,
  row_version: 1,
  created_by_user_id: observationAuthority.actorId,
  created_at: "2026-09-11T12:00:00Z",
  resolved_by_user_id: null,
  resolved_at: null,
  resolution_method: null,
};
export const observationTargetId = "50000000-0000-4000-8000-000000000001";
export const observationTargetRow = {
  record_id: observationTargetId,
  row_version: 90,
  cells: {
    "indicator.display_value": { value: "target.example" },
    "indicator.indicator_type": { value: "domain_name" },
  },
};
export const testObservationReceipt: ObservationReceipt = {
  observation: testObservation,
  change_set_id: "60000000-0000-4000-8000-000000000001",
  replayed: false,
  affected_records: [{ record_id: observationSource.recordId, row_version: 5 }],
};
export function observationReaderFixture() {
  return {
    observations: vi.fn<ObservationReadPort["observations"]>(async () => ({
      kind: "accepted",
      value: { items: [testObservation], hasMore: false, nextCursor: null },
    })),
    records: vi.fn<ObservationReadPort["records"]>(async () => ({
      kind: "accepted",
      value: {
        items: [observationTargetRow],
        hasMore: false,
        nextCursor: null,
      },
    })),
  } satisfies ObservationReadPort;
}
export const observationCreateIntent: Extract<
  ObservationIntent,
  { action: "create" }
> = {
  action: "create",
  source: observationSource,
  selection: { startByte: 2, endByte: 12, text: "α.example" },
};
export function observationOwnerFixture() {
  let sequence = 0;
  const ids = { create: vi.fn(() => `secure-observation-${++sequence}`) },
    accepted = vi.fn();
  const owner = new WorkbookObservationOwner(
    observationAuthority.incidentId,
    ids,
    accepted,
  );
  const reader = observationReaderFixture();
  const transport = {
    ...createObservationTransport({
      apiBase: "https://original.test",
      incidentId: observationAuthority.incidentId,
    }),
    send: vi.fn<ObservationTransportPort["send"]>(async () => ({
      kind: "acknowledged",
      receipt: testObservationReceipt,
    })),
  };
  owner.configure(reader, transport);
  owner.setAuthority(observationAuthority);
  const projections = vi.fn(async () => {});
  owner.registerReconciliation(projections);
  const binding = {
    isCurrent: vi.fn(() => true),
    matchesDraft: vi.fn(() => true),
    prepare: vi.fn(async () => true),
    reconcile: vi.fn(async () => {}),
  };
  function admit(intent: ObservationIntent = observationCreateIntent) {
    const attempt = owner.admit(intent, binding);
    if (!attempt) throw new Error("fixture admission failed");
    return attempt;
  }
  const source = {
    fields: [
      {
        fieldKey: observationSource.fieldKey,
        label: "Raw activity",
        value: observationSource.text,
      },
    ],
    source: () => observationSource,
    ready: () => true,
    prepare: async () => true,
  };
  return {
    owner,
    ids,
    accepted,
    reader,
    transport,
    projections,
    binding,
    admit,
    source,
    entry: () => owner.getSnapshot().entries[0],
  };
}
