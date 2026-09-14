import { buildHTTPOperationPath } from "@cartulary/protocol-ts/http";
import type { OrdinaryCreateTransport } from "../features/ordinary/ordinaryCreateOperation";
import { freezeWorkbookValue } from "../utils/freezeWorkbookValue";
import { sendWorkbookRecordMutation } from "./sendWorkbookRecordMutation";
import { workbookCreateCapabilityMatches } from "./workbookCreateCapability";
import { createWorkbookOperationExecutor } from "./workbookOperationExecutor";

/** Existing public create route, with frozen bytes and fresh transport credentials. */
export function createOrdinaryCreateTransport(
  apiBase: string | undefined,
): OrdinaryCreateTransport {
  const operations = createWorkbookOperationExecutor({ apiBase });
  return {
    capture(input) {
      const pathParameters = {
        incident_id: input.authority.incidentId,
        view_schema_id: input.target.viewSchemaId,
      };
      return freezeWorkbookValue(
        structuredClone({
          ...input,
          operationID: "createViewRow" as const,
          apiBase,
          pathParameters,
          path: buildHTTPOperationPath("createViewRow", pathParameters),
          body: JSON.stringify(input.request),
        }),
      );
    },
    async verify(target, signal) {
      const result = await operations.execute({
        operationID: "getViewSchema",
        pathParameters: { view_schema_id: target.viewSchemaId },
        signal,
      });
      if (result.kind !== "accepted")
        throw new Error("Creation capability could not be verified.");
      const schema = result.value.data;
      if (!workbookCreateCapabilityMatches(schema, target))
        throw new Error(
          "Creation capability changed. Review the retained draft.",
        );
    },
    send: (attempt, signal) =>
      sendWorkbookRecordMutation(
        { ...attempt, viewSchemaId: attempt.target.viewSchemaId },
        signal,
      ),
  };
}
