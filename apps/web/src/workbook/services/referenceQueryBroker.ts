import { requireViewContract } from "@cartulary/view-contracts";
import { apiPath } from "../../services/browserApi";
import {
  fetchWorkbookJSON,
  parseErrorMessage,
  readEnvelope,
} from "../../services/workbookApi";
import {
  buildQueryRequest,
  emptyWorkbookQueryState,
} from "../models/workbookQuery";
import type { ReferenceRequirement } from "../models/workbookSurfaceRegistration";
import type { EntityApiRow } from "../timeline/models/workbookTimelineModel";

type ViewQueryEnvelope = {
  data: {
    view_schema_id: string;
    rows: EntityApiRow[];
  };
};

export type ReferenceQueryContext = {
  readonly apiBase?: string | undefined;
  readonly authorizationEpoch: string;
  readonly incidentId: string;
};

type ReferenceQueryResult = {
  readonly requirement: ReferenceRequirement;
  readonly rows: readonly EntityApiRow[];
};

function abortFailure(): DOMException {
  return new DOMException("Reference query consumer is obsolete", "AbortError");
}

export class ReferenceQueryBroker {
  readonly #inFlight = new Map<string, Promise<readonly EntityApiRow[]>>();

  async execute(
    requirements: readonly ReferenceRequirement[],
    context: ReferenceQueryContext,
    signal: AbortSignal,
  ): Promise<readonly ReferenceQueryResult[]> {
    const unique = [
      ...new Map(requirements.map((item) => [item.resourceId, item])).values(),
    ];
    return Promise.all(
      unique.map(async (requirement) => ({
        requirement,
        rows: await this.#consume(this.#query(requirement, context), signal),
      })),
    );
  }

  #query(
    requirement: ReferenceRequirement,
    context: ReferenceQueryContext,
  ): Promise<readonly EntityApiRow[]> {
    const key = [
      context.apiBase ?? "",
      context.incidentId,
      context.authorizationEpoch,
      requirement.resourceId,
    ].join("\u0000");
    const existing = this.#inFlight.get(key);
    if (existing) {
      return existing;
    }
    const targetContract = requireViewContract(requirement.viewSchemaId);
    const pending = fetchWorkbookJSON<ViewQueryEnvelope>(
      apiPath(
        context.apiBase,
        `/api/v1/incidents/${context.incidentId}/views/${requirement.viewSchemaId}/query`,
      ),
      {
        method: "POST",
        body: JSON.stringify(
          buildQueryRequest(targetContract, emptyWorkbookQueryState()),
        ),
      },
    )
      .then((result) => {
        if (!result.ok) {
          throw new Error(parseErrorMessage(result.payload));
        }
        return readEnvelope<ViewQueryEnvelope>(result.payload).data.rows;
      })
      .finally(() => {
        this.#inFlight.delete(key);
      });
    this.#inFlight.set(key, pending);
    return pending;
  }

  async #consume<T>(pending: Promise<T>, signal: AbortSignal): Promise<T> {
    if (signal.aborted) {
      throw abortFailure();
    }
    return new Promise<T>((resolve, reject) => {
      const onAbort = () => reject(abortFailure());
      signal.addEventListener("abort", onAbort, { once: true });
      void pending.then(
        (value) => {
          signal.removeEventListener("abort", onAbort);
          resolve(value);
        },
        (error: unknown) => {
          signal.removeEventListener("abort", onAbort);
          reject(error);
        },
      );
    });
  }
}

export const referenceQueryBroker = new ReferenceQueryBroker();
