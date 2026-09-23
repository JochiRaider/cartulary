import type {
  IndicatorLifecycleInterval,
  IndicatorLifecycleReceipt,
} from "../workbook/adapters/indicatorLifecycleProtocol";
import {
  emptyLifecycleValues,
  type LifecycleDraft,
} from "../workbook/features/indicators/indicatorLifecycleModel";
import type { LifecycleAuthority } from "../workbook/features/indicators/indicatorLifecycleOperation";
import type { WorkbookQueryRow } from "../workbook/query/WorkbookQueryRow";

const lifecycleIncident = "00000000-0000-4000-8000-000000000001";
const lifecycleActor = "00000000-0000-4000-8000-000000000002";
export const lifecycleIndicator = "00000000-0000-4000-8000-000000000003";
export const lifecycleAuthority: LifecycleAuthority = {
  incidentId: lifecycleIncident,
  actorId: lifecycleActor,
  role: "editor",
  sessionIdentity: "session-one",
  closed: false,
};
export function lifecycleInterval(): IndicatorLifecycleInterval {
  return {
    interval_id: "00000000-0000-4000-8000-000000000004",
    incident_id: lifecycleIncident,
    indicator_record_id: lifecycleIndicator,
    lifecycle_state: "active",
    valid_from: "2026-09-11T12:00:00Z",
    valid_to: null,
    confidence: 0,
    rationale: "Observed context",
    support_refs: [],
    assessor: "Analyst supplied name",
    assessed_at: "2026-09-11T14:00:00Z",
    row_version: 1,
    created_by_user_id: lifecycleActor,
    created_at: "2026-09-11T14:00:00Z",
  };
}
export function lifecycleReceipt(): IndicatorLifecycleReceipt {
  return {
    interval: lifecycleInterval(),
    affected_records: [{ record_id: lifecycleIndicator, row_version: 2 }],
    change_set_id: "00000000-0000-4000-8000-000000000005",
    replayed: false,
  };
}
export function lifecycleRow(version = 1): WorkbookQueryRow {
  return {
    record_id: lifecycleIndicator,
    row_version: version,
    cells: { "indicator.display_value": { value: "example.test" } },
  } as WorkbookQueryRow;
}
export function lifecycleDraft(): LifecycleDraft {
  return {
    recordId: lifecycleIndicator,
    label: "example.test",
    baseRowVersion: 1,
    revision: 1,
    values: {
      ...emptyLifecycleValues,
      validFrom: "2026-09-11T12:00",
      confidence: "0",
      rationale: "Observed context",
      assessor: "Analyst supplied name",
    },
  };
}
