import type { WorkbookCommittedRecordPort } from "../../query/WorkbookCommittedRecordPort";
import type { WorkbookIndicatorCreateOwner } from "./WorkbookIndicatorCreateOwner";
import type { WorkbookIndicatorLifecycleOwner } from "./WorkbookIndicatorLifecycleOwner";
import type { WorkbookObservationOwner } from "./WorkbookObservationOwner";

/** The three Indicator sources share one committed-record view. */
export function createIndicatorCommittedRecords({
  lifecycle,
  observations,
  create,
  history,
}: {
  readonly lifecycle: WorkbookIndicatorLifecycleOwner;
  readonly observations: WorkbookObservationOwner;
  readonly create: WorkbookIndicatorCreateOwner;
  readonly history: { latestVersion(recordId: string): number | null };
}): WorkbookCommittedRecordPort {
  const records: WorkbookCommittedRecordPort = {
    subscribe: (listener) => {
      let authorized = !!records.getSnapshot().authority;
      const changed = () => {
        const next = !!records.getSnapshot().authority;
        // Sequential initialization is not revocation. Once active, loss of
        // any source authority must conceal protected rows immediately.
        if (next || authorized) {
          authorized = next;
          listener();
        }
      };
      const a = lifecycle.subscribe(changed);
      const b = observations.subscribe(changed);
      const c = create.subscribe(changed);
      return () => {
        a();
        b();
        c();
      };
    },
    getSnapshot: () =>
      !create.getSnapshot().authority
        ? create.getSnapshot()
        : observations.getSnapshot().authority
          ? lifecycle.getSnapshot()
          : observations.getSnapshot(),
    latestVersion: (id) =>
      Math.max(
        lifecycle.latestVersion(id) ?? 0,
        observations.latestVersion(id) ?? 0,
        create.latestVersion(id) ?? 0,
        history.latestVersion(id) ?? 0,
      ) || null,
    latestRow: (id) => {
      const rows = [
        lifecycle.latestRow(id),
        observations.latestRow(id),
        create.latestRow(id),
      ].filter(
        (row) =>
          row !== null && row.row_version >= (records.latestVersion(id) ?? 0),
      );
      return (
        rows.sort((a, b) => (b?.row_version ?? 0) - (a?.row_version ?? 0))[0] ??
        null
      );
    },
    acceptRow: (row) => {
      if (
        !records.getSnapshot().authority ||
        row.row_version < (records.latestVersion(row.record_id) ?? 0)
      )
        return records.latestRow(row.record_id);
      lifecycle.acceptRow(row);
      observations.acceptRow(row);
      create.acceptRow(row);
      return records.latestRow(row.record_id);
    },
  };
  return records;
}
