import { boundedRead } from "../../services/asyncObservation";
import type {
  EvidenceMetadataResult,
  WorkbookEvidenceMetadataPort,
} from "../evidence/WorkbookEvidenceMetadataPort";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { sameWorkbookReadScope } from "../query/workbookRowObservation";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

const invalid: EvidenceMetadataResult = {
  kind: "rejected",
  failure: {
    kind: "invalid_contract",
    message: "Evidence information is unavailable.",
  },
};
const uuid = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

export function createWorkbookEvidenceMetadataAdapter(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly readScope: WorkbookReadScopeSource;
}): WorkbookEvidenceMetadataPort {
  const operations = createWorkbookOperationExecutor({
    apiBase: options.apiBase,
  });
  return {
    async query(input) {
      const scope = options.readScope();
      if (
        input.signal.aborted ||
        !scope ||
        scope.incidentId !== options.incidentId
      )
        return { kind: "aborted" };
      const ids = [...input.recordIds];
      if (
        ids.length < 1 ||
        ids.length > 100 ||
        new Set(ids).size !== ids.length ||
        ids.some(
          (id) =>
            !uuid.test(id) || id === "00000000-0000-0000-0000-000000000000",
        )
      )
        return invalid;
      try {
        const outcome = await boundedRead(
          (signal) =>
            operations.execute({
              operationID: "queryEvidenceMetadata",
              pathParameters: { incident_id: options.incidentId },
              request: { record_ids: ids as [string, ...string[]] },
              signal,
            }),
          input.signal,
        );
        if (
          input.signal.aborted ||
          !sameWorkbookReadScope(scope, options.readScope())
        )
          return { kind: "aborted" };
        if (outcome.kind === "rejected") return outcome;
        const { incident_id, items } = outcome.value.data;
        if (
          incident_id !== options.incidentId ||
          items.length !== ids.length ||
          items.some(
            (item, index) =>
              item.record_id !== ids[index] ||
              !Number.isSafeInteger(item.row_version) ||
              item.row_version < 1 ||
              item.preview_eligibility.eligible !==
                (item.preview_eligibility.reason_code === null) ||
              (!item.has_blob && item.upload_state !== null) ||
              (item.preview_eligibility.eligible &&
                (!item.has_blob ||
                  item.upload_state !== "available" ||
                  (item.lifecycle_state !== "available" &&
                    item.lifecycle_state !== "released"))),
          )
        )
          return invalid;
        return {
          kind: "accepted",
          items: items.map((item) => ({
            recordId: item.record_id,
            rowVersion: item.row_version,
            title: item.title,
            lifecycleState: item.lifecycle_state,
            hasBlob: item.has_blob,
            uploadState: item.upload_state,
            previewEligibility: {
              eligible: item.preview_eligibility.eligible,
              reasonCode: item.preview_eligibility.reason_code,
            },
          })),
        };
      } catch {
        if (
          input.signal.aborted ||
          !sameWorkbookReadScope(scope, options.readScope())
        )
          return { kind: "aborted" };
        return {
          kind: "rejected",
          failure: {
            kind: "retryable",
            message: "Evidence information is unavailable.",
          },
        };
      }
    },
  };
}
