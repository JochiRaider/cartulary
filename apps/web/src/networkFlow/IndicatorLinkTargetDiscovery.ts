import { boundedRead } from "../services/asyncObservation";

import type {
  IndicatorLinkTargetOption,
  IndicatorLinkTargetPort,
  IndicatorLinkTargetQuery,
} from "../services/networkFlowIndicatorAdapter";

export type IndicatorLinkDiscoveryContext = Omit<
  IndicatorLinkTargetQuery,
  "cursor"
> & { readonly key: string };
export type IndicatorLinkDiscoverySnapshot = {
  readonly phase: "idle" | "loading" | "ready" | "failed";
  readonly items: readonly IndicatorLinkTargetOption[];
  readonly nextCursor: string | null;
  readonly error: string | null;
};
const initial = (): IndicatorLinkDiscoverySnapshot => ({
  phase: "idle",
  items: [],
  nextCursor: null,
  error: null,
});

/** A single bounded Core query page, independent of submitted write outcomes. */
export class IndicatorLinkTargetDiscovery {
  private state = initial();
  private listeners = new Set<() => void>();
  private port: IndicatorLinkTargetPort | null = null;
  private context: () => IndicatorLinkDiscoveryContext | null = () => null;
  private key: string | null = null;
  private request: AbortController | null = null;
  private cursor: string | null = null;
  readonly getSnapshot = () => this.state;
  readonly subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };
  private update(patch: Partial<IndicatorLinkDiscoverySnapshot>) {
    this.state = { ...this.state, ...patch };
    for (const listener of this.listeners) listener();
  }
  bind(
    port: IndicatorLinkTargetPort,
    context: () => IndicatorLinkDiscoveryContext | null,
  ) {
    this.port = port;
    this.context = context;
    this.revalidate();
  }
  readonly revalidate = () => {
    const key = this.context()?.key ?? null;
    if (key !== this.key) {
      this.clear();
      this.key = key;
    }
  };
  readonly load = (cursor: string | null = null): void => {
    this.revalidate();
    const context = this.context();
    const port = this.port;
    if (context === null || port === null || this.request !== null) return;
    const request = new AbortController();
    this.request = request;
    this.cursor = cursor;
    this.update({ phase: "loading", items: [], nextCursor: null, error: null });
    void boundedRead(
      (signal) =>
        port(
          {
            incidentId: context.incidentId,
            candidateValue: context.candidateValue,
            cursor,
          },
          signal,
        ),
      request.signal,
    )
      .then((page) => {
        if (
          !request.signal.aborted &&
          this.request === request &&
          this.context()?.key === context.key
        )
          this.update({
            phase: "ready",
            items: page.items,
            nextCursor: page.nextCursor,
            error: null,
          });
      })
      .catch(() => {
        if (
          !request.signal.aborted &&
          this.request === request &&
          this.context()?.key === context.key
        )
          this.update({
            phase: "failed",
            error:
              "Compatible indicators could not be loaded. Retry this page or enter a known indicator ID.",
          });
      })
      .finally(() => {
        if (this.request === request) this.request = null;
      });
  };
  readonly next = () => {
    if (this.state.nextCursor !== null) this.load(this.state.nextCursor);
  };
  readonly retry = () => this.load(this.cursor);
  readonly clear = () => {
    this.request?.abort();
    this.request = null;
    this.cursor = null;
    this.key = null;
    this.state = initial();
    for (const listener of this.listeners) listener();
  };
}
