import { requireViewContract } from "@cartulary/view-contracts";
import type { WorkbookInvalidationReason } from "../lifecycle/workbookInvalidation";
import { emptyWorkbookQueryState } from "../models/workbookQuery";
import type { ReferenceRequirement } from "../models/workbookSurfaceRegistration";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import type { WorkbookViewQueryPort } from "../query/WorkbookViewQueryPort";

export type ReferenceQueryBrokerContext = {
  readonly authorizationGeneration: string;
  readonly viewQuery: WorkbookViewQueryPort;
};

export type ReferenceQueryResult = {
  readonly requirement: ReferenceRequirement;
  readonly rows: readonly WorkbookQueryRow[];
};

export type ReferenceQueryInvalidationReason = Extract<
  WorkbookInvalidationReason,
  {
    readonly kind:
      | "session_unavailable"
      | "incident_access_lost"
      | "incident_role_changed"
      | "incident_closed"
      | "incident_changed"
      | "runtime_disposed";
  }
>;

export interface ReferenceQueryBrokerPort {
  execute(
    requirements: readonly ReferenceRequirement[],
    signal: AbortSignal,
  ): Promise<readonly ReferenceQueryResult[]>;
  invalidate(reason: ReferenceQueryInvalidationReason): void;
  dispose(): void;
}

type InFlightReferenceQuery = {
  readonly controller: AbortController;
  consumerCount: number;
  promise: Promise<readonly WorkbookQueryRow[]>;
  settled: boolean;
};

function abortFailure(): DOMException {
  return new DOMException("Reference query consumer is obsolete", "AbortError");
}

class ReferenceQueryBroker implements ReferenceQueryBrokerPort {
  readonly #context: ReferenceQueryBrokerContext;
  #disposed = false;
  readonly #inFlight = new Map<string, InFlightReferenceQuery>();

  constructor(context: ReferenceQueryBrokerContext) {
    this.#context = context;
  }

  async execute(
    requirements: readonly ReferenceRequirement[],
    signal: AbortSignal,
  ): Promise<readonly ReferenceQueryResult[]> {
    if (this.#disposed) {
      throw abortFailure();
    }
    const unique = [
      ...new Map(requirements.map((item) => [item.resourceId, item])).values(),
    ];
    return Promise.all(
      unique.map(async (requirement) => ({
        requirement,
        rows: await this.#consume(this.#query(requirement), signal),
      })),
    );
  }

  invalidate(_reason: ReferenceQueryInvalidationReason): void {
    this.#disposeEntries();
  }

  dispose(): void {
    this.#disposeEntries();
  }

  #disposeEntries(): void {
    if (this.#disposed) {
      return;
    }
    this.#disposed = true;
    for (const entry of this.#inFlight.values()) {
      entry.controller.abort();
    }
    this.#inFlight.clear();
  }

  #query(requirement: ReferenceRequirement): InFlightReferenceQuery {
    const key = requirement.resourceId;
    const existing = this.#inFlight.get(key);
    if (existing) {
      return existing;
    }
    const controller = new AbortController();
    const targetContract = requireViewContract(requirement.viewSchemaId);
    const entry: InFlightReferenceQuery = {
      controller,
      consumerCount: 0,
      promise: Promise.resolve([]),
      settled: false,
    };
    const pending = this.#context.viewQuery
      .query({
        contract: targetContract,
        queryState: emptyWorkbookQueryState(),
        signal: controller.signal,
      })
      .then((result) => {
        if (
          this.#disposed ||
          controller.signal.aborted ||
          result.kind === "aborted"
        ) {
          throw abortFailure();
        }
        if (result.kind === "rejected") {
          throw new Error(result.failure.message);
        }
        return result.value.rows;
      })
      .finally(() => {
        entry.settled = true;
        if (this.#inFlight.get(key) === entry) {
          this.#inFlight.delete(key);
        }
      });
    entry.promise = pending;
    this.#inFlight.set(key, entry);
    return entry;
  }

  async #consume(
    entry: InFlightReferenceQuery,
    signal: AbortSignal,
  ): Promise<readonly WorkbookQueryRow[]> {
    if (signal.aborted || this.#disposed) {
      throw abortFailure();
    }
    entry.consumerCount += 1;
    return new Promise<readonly WorkbookQueryRow[]>((resolve, reject) => {
      let released = false;
      const release = () => {
        if (released) {
          return;
        }
        released = true;
        entry.consumerCount -= 1;
        if (entry.consumerCount === 0 && !entry.settled) {
          entry.controller.abort();
        }
      };
      const onAbort = () => {
        release();
        reject(abortFailure());
      };
      signal.addEventListener("abort", onAbort, { once: true });
      void entry.promise.then(
        (value) => {
          signal.removeEventListener("abort", onAbort);
          release();
          if (this.#disposed || entry.controller.signal.aborted) {
            reject(abortFailure());
            return;
          }
          resolve(value);
        },
        (error: unknown) => {
          signal.removeEventListener("abort", onAbort);
          release();
          reject(error);
        },
      );
    });
  }
}

export function createReferenceQueryBroker(
  context: ReferenceQueryBrokerContext,
): ReferenceQueryBrokerPort {
  return new ReferenceQueryBroker(context);
}
