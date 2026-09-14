import {
  hostsViewSchemaId,
  identitiesViewSchemaId,
  requireViewContract,
  timelineViewSchemaId,
} from "@cartulary/view-contracts";
import type {
  AssessmentCandidatePage,
  AssessmentCandidateQuery,
  AssessmentCandidateReadPort,
} from "../features/assessments/assessmentCandidatePort";
import { assessmentSupportCandidate } from "../models/assessmentWorkbookModel";
import { buildQueryRequest } from "../models/workbookQuery";
import type { WorkbookPortResult } from "../ports/WorkbookPortResult";
import { workbookFailureLifecycle } from "../ports/WorkbookPortResult";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

/** Reduces query rows here; authoring never consumes Timeline editor/runtime rows. */
export function createAssessmentCandidateReader(options: {
  readonly apiBase: string | undefined;
  readonly incidentId: string;
  readonly recheckAuthority?: () => void;
}): AssessmentCandidateReadPort {
  const operations = createWorkbookOperationExecutor(options);
  async function page(
    view: string,
    label: string,
    input: AssessmentCandidateQuery,
  ): Promise<WorkbookPortResult<AssessmentCandidatePage>> {
    try {
      const result = await operations.execute({
        operationID: "queryWorkbookView",
        pathParameters: {
          incident_id: options.incidentId,
          view_schema_id: view,
        },
        request: {
          ...buildQueryRequest(requireViewContract(view), input.queryState),
          limit: 100,
          ...(input.cursor ? { cursor_token: input.cursor } : {}),
        },
        signal: input.signal,
      });
      if (input.signal.aborted) return { kind: "aborted" };
      if (result.kind === "rejected") {
        if (
          workbookFailureLifecycle(result.failure).kind ===
          "authority_unavailable"
        )
          options.recheckAuthority?.();
        return result;
      }
      const { data, meta } = result.value;
      if (
        data.incident_id !== options.incidentId ||
        data.view_schema_id !== view ||
        !Array.isArray(data.rows) ||
        !meta.paging ||
        meta.paging.limit !== 100 ||
        meta.paging.has_more !== (meta.paging.next_cursor !== null) ||
        (meta.paging.next_cursor !== null &&
          (typeof meta.paging.next_cursor !== "string" ||
            !meta.paging.next_cursor ||
            meta.paging.next_cursor === input.cursor))
      )
        throw new Error("Invalid candidate page");
      const candidates = data.rows.map((row) => {
        if (typeof row.record_id !== "string" || !row.record_id.trim())
          throw new Error("Invalid candidate identity");
        return assessmentSupportCandidate(
          row.record_id,
          row.cells?.[label]?.value,
        );
      });
      return {
        kind: "accepted",
        value: {
          candidates,
          hasMore: meta.paging.has_more,
          nextCursor: meta.paging.next_cursor,
        },
      };
    } catch {
      return input.signal.aborted
        ? { kind: "aborted" }
        : {
            kind: "rejected",
            failure: {
              kind: "retryable",
              message:
                "Candidates could not be loaded or verified. Retry the read.",
            },
          };
    }
  }
  return {
    subjects: (type, input) =>
      page(
        type === "host" ? hostsViewSchemaId : identitiesViewSchemaId,
        `${type}.display_name`,
        input,
      ),
    support: (input) =>
      page(timelineViewSchemaId, "timeline.activity_synopsis_text", input),
  };
}
