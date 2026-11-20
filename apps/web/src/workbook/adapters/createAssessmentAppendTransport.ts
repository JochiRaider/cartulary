import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import { assessmentsViewSchemaId } from "@cartulary/view-contracts";
import { fetchHTTPOperation } from "../../services/browserApi";
import {
  type AssessmentAppendTransport,
  freezeAssessment,
} from "../features/assessments/assessmentOperation";
import { buildAssessmentCreatePayload } from "../models/assessmentWorkbookModel";
import { acceptWorkbookRowObservation } from "../query/acceptWorkbookRowObservation";
import type { WorkbookReadScopeSource } from "../query/WorkbookQueryRow";
import { explicitWorkbookFailureOutcome } from "./workbookOperationErrorPolicy";
import { acceptedRecordMutation } from "./workbookRecordPatchTransport";

export function createAssessmentAppendTransport(
  apiBase: string | undefined,
  readScope?: WorkbookReadScopeSource,
): AssessmentAppendTransport {
  return {
    capture(review, clientTxnId) {
      const request = buildAssessmentCreatePayload(
        review.draft.values,
        clientTxnId,
      );
      if (!request) throw new Error("Complete the required assessment fields.");
      return freezeAssessment({
        clientTxnId,
        review: structuredClone(review),
        request,
        body: JSON.stringify(request),
        apiBase,
        path: buildHTTPOperationPath("createViewRow", {
          incident_id: review.authority.incidentId,
          view_schema_id: assessmentsViewSchemaId,
        }),
      });
    },
    async send(attempt, signal) {
      const scope = readScope?.() ?? null;
      let status: number | null = null;
      try {
        const pathParameters = {
          incident_id: attempt.review.authority.incidentId,
          view_schema_id: assessmentsViewSchemaId,
        };
        if (
          attempt.path !==
          buildHTTPOperationPath("createViewRow", pathParameters)
        )
          return { kind: "uncertain" };
        const result = await fetchHTTPOperation({
          apiBase: attempt.apiBase,
          operationID: "createViewRow",
          pathParameters,
          init: { method: "POST", body: attempt.body, signal },
          onResponse: (response) => {
            status = response.status;
          },
        });
        if (!result.ok) {
          return explicitWorkbookFailureOutcome(
            status,
            result.payload.error,
            "createViewRow",
          );
        }
        return acceptedRecordMutation(
          result.payload.data,
          assessmentsViewSchemaId,
        )
          ? {
              kind: "accepted",
              receipt: freezeAssessment(
                structuredClone({
                  ...result.payload,
                  data: {
                    ...result.payload.data,
                    row: acceptWorkbookRowObservation(
                      result.payload.data.row,
                      scope,
                    ),
                  },
                }),
              ),
            }
          : { kind: "uncertain" };
      } catch {
        return { kind: "uncertain" };
      }
    },
  };
}
