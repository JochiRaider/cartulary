import {
  type HTTPOperationResponse,
  httpOperationBindings,
} from "@cartulary/protocol-ts/http";
import { fetchHTTPOperation } from "../../services/browserApi";
import type { WorkbookOperationOutcome } from "../mutations/workbookOperationOutcome";
import type {
  WorkbookOperationExecution,
  WorkbookOperationExecutor,
  WorkbookOperationID,
} from "./workbookOperationContract";
import { classifyWorkbookOperationFailure } from "./workbookOperationErrorPolicy";

export function createWorkbookOperationExecutor(options: {
  readonly apiBase: string | undefined;
}): WorkbookOperationExecutor {
  return {
    async execute<OperationID extends WorkbookOperationID>(
      input: WorkbookOperationExecution<OperationID>,
    ): Promise<WorkbookOperationOutcome<HTTPOperationResponse<OperationID>>> {
      const request = input.request;
      const result = await fetchHTTPOperation<
        HTTPOperationResponse<OperationID>
      >({
        apiBase: options.apiBase,
        operationID: input.operationID,
        ...(input.observeTransport?.onJSONParsed === undefined
          ? {}
          : { onJSONParsed: input.observeTransport.onJSONParsed }),
        ...(input.observeTransport?.onResponseStatus === undefined
          ? {}
          : {
              onResponse: (response: Response) =>
                input.observeTransport?.onResponseStatus?.(response.status),
            }),
        pathParameters: input.pathParameters,
        query: input.query,
        init: {
          method: httpOperationBindings[input.operationID].method,
          ...(input.signal === undefined ? {} : { signal: input.signal }),
          ...(request === undefined ? {} : { body: JSON.stringify(request) }),
        },
      });
      return result.ok
        ? { kind: "accepted", value: result.payload }
        : {
            kind: "rejected",
            failure: classifyWorkbookOperationFailure(
              result.status,
              result.payload,
              input.operationID,
            ),
          };
    },
  };
}
