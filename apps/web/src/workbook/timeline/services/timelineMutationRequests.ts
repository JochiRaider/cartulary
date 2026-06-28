import { fetchJSON } from "../../../services/workbookApi";
import type {
  PendingReplayPayloadIntent,
  PendingReplayUnitState,
} from "../../utils/workbookPendingQueue";

export type TimelineMutationEnvelope = {
  data: {
    view_schema_id: string;
    change_set_id?: string;
    row: unknown;
  };
};

export type TimelineMutationFetchResult = Awaited<
  ReturnType<typeof fetchJSON<TimelineMutationEnvelope>>
>;

export type TimelineWorkbookTimingRecorder = (
  name: string,
  fields?: Record<string, unknown>,
) => void;

export async function dispatchTimelinePendingReplayMutation({
  payload,
  recordTiming,
  unit,
}: {
  readonly payload: PendingReplayPayloadIntent;
  readonly recordTiming: TimelineWorkbookTimingRecorder;
  readonly unit: PendingReplayUnitState;
}): Promise<TimelineMutationFetchResult> {
  recordTiming("pending_fetch_start", {
    clientTxnId: unit.clientTxnId,
    kind: unit.kind,
    rowKey: unit.rowKey,
  });
  return fetchJSON<TimelineMutationEnvelope>(
    unit.path,
    {
      method: unit.method,
      body: JSON.stringify(payload),
    },
    {
      onJSONParsed: () => {
        recordTiming("pending_fetch_json_parsed", {
          clientTxnId: unit.clientTxnId,
          kind: unit.kind,
          rowKey: unit.rowKey,
        });
      },
      onResponse: (response) => {
        recordTiming("pending_fetch_response", {
          clientTxnId: unit.clientTxnId,
          kind: unit.kind,
          rowKey: unit.rowKey,
          serverTiming: response.headers.get("server-timing") ?? "",
          status: response.status,
        });
      },
    },
  );
}
