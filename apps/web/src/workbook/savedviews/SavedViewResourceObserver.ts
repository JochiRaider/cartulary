import type { SavedViewResource } from "../models/workbookSavedViews";
import type {
  SavedViewObserver,
  SavedViewProblem,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";

export type SavedViewObservation = {
  readonly resource: SavedViewResource | null;
  readonly status: "unobserved" | "ready" | "unavailable";
  readonly pending: boolean;
  readonly problem: SavedViewProblem | null;
  readonly revision: number;
};
type SavedViewResourceSlot = "selected" | "operation" | "home" | "default";
export type SavedViewObservationResult =
  | { readonly kind: "accepted"; readonly value: SavedViewResource }
  | { readonly kind: "rejected"; readonly failure: SavedViewProblem }
  | { readonly kind: "aborted" };
export type SavedViewObservationHandle = {
  readonly result: Promise<SavedViewObservationResult>;
  readonly release: () => void;
};
const empty = (): SavedViewObservation => ({
  resource: null,
  status: "unobserved",
  pending: false,
  problem: null,
  revision: 0,
});

/** Addressed observations retained only by concrete consumers, never by pages. */
export class SavedViewResourceObserver {
  private state: ReadonlyMap<string, SavedViewObservation> = new Map();
  private slots = new Map<SavedViewResourceSlot, string>();
  private handles = new Map<
    symbol,
    { id: string; retains: boolean; cancel: () => void }
  >();
  private reads = new Map<string, symbol>();
  private generation = 0;
  private revision = 0;
  constructor(
    private readonly ports: {
      port: () => WorkbookSavedViewPort | null;
      observe: SavedViewObserver;
      visible: (resource: SavedViewResource) => boolean;
      changed: () => void;
      failed: (id: string, problem: SavedViewProblem) => void;
    },
  ) {}
  getSnapshot = () => this.state;
  get = (id: string | null) => (id === null ? undefined : this.state.get(id));
  private cancelReads(id: string) {
    this.reads.delete(id);
    for (const handle of this.handles.values())
      if (handle.id === id) handle.cancel();
  }
  private collect() {
    const keep = new Set([
      ...this.slots.values(),
      ...[...this.handles.values()]
        .filter((handle) => handle.retains)
        .map((handle) => handle.id),
    ]);
    for (const [token, handle] of this.handles) {
      if (!keep.has(handle.id)) {
        if (this.reads.get(handle.id) === token) this.reads.delete(handle.id);
        handle.cancel();
      }
    }
    const next = new Map(this.state);
    for (const id of next.keys()) if (!keep.has(id)) next.delete(id);
    this.state = next;
    this.ports.changed();
  }
  retain = (slot: SavedViewResourceSlot, id: string | null) => {
    if (this.slots.get(slot) === (id ?? undefined)) return;
    if (id === null) this.slots.delete(slot);
    else this.slots.set(slot, id);
    if (id !== null && !this.state.has(id))
      this.state = new Map(this.state).set(id, empty());
    this.collect();
  };
  private publish(id: string, value: SavedViewObservation) {
    if (!this.state.has(id)) return;
    this.state = new Map(this.state).set(id, value);
    this.ports.changed();
  }
  accept = (resource: SavedViewResource) => {
    const id = resource.saved_view_id;
    const old = this.get(id);
    if (
      !old ||
      !this.ports.visible(resource) ||
      (old.resource &&
        old.resource.saved_view_version > resource.saved_view_version)
    )
      return;
    this.cancelReads(id);
    this.publish(id, {
      resource: structuredClone(resource),
      status: "ready",
      pending: false,
      problem: null,
      revision: ++this.revision,
    });
  };
  invalidate = () => {
    this.generation++;
    for (const handle of this.handles.values()) handle.cancel();
    this.reads.clear();
    this.state = new Map(
      [...this.state].map(([id, value]) => [id, { ...value, pending: false }]),
    );
    this.ports.changed();
  };
  unavailable = (id: string, problem: SavedViewProblem) => {
    this.cancelReads(id);
    this.publish(id, {
      resource: null,
      status: "unavailable",
      pending: false,
      problem,
      revision: ++this.revision,
    });
  };
  clear = () => {
    this.invalidate();
    this.slots.clear();
    this.handles.clear();
    this.state = new Map();
    this.ports.changed();
  };
  /** The handle owns only its read and retention, even for same-resource peers. */
  observe = (id: string, signal?: AbortSignal): SavedViewObservationHandle =>
    this.observeResource(id, signal, true);
  private observeResource = (
    id: string,
    signal: AbortSignal | undefined,
    retains: boolean,
  ): SavedViewObservationHandle => {
    const token = Symbol(id);
    const generation = this.generation;
    let released = false;
    let cancel = () => {};
    const release = () => {
      if (released) return;
      released = true;
      signal?.removeEventListener("abort", release);
      cancel();
      this.handles.delete(token);
      if (this.reads.get(id) === token) {
        this.reads.delete(id);
        const current = this.get(id);
        if (current?.pending) this.publish(id, { ...current, pending: false });
      }
      this.collect();
    };
    const port = this.ports.port();
    if (!port || signal?.aborted)
      return { result: Promise.resolve({ kind: "aborted" }), release };
    this.handles.set(token, { id, retains, cancel: () => cancel() });
    const prior = this.get(id) ?? empty();
    this.state = new Map(this.state).set(id, prior);
    this.reads.set(id, token);
    this.publish(id, { ...prior, pending: true, problem: null });
    const read = this.ports.observe((readSignal) =>
      port.getResource({ savedViewId: id, signal: readSignal }),
    );
    cancel = read.cancel;
    signal?.addEventListener("abort", release, { once: true });
    if (signal?.aborted) release();
    const result = read.result.then((outcome): SavedViewObservationResult => {
      if (
        released ||
        generation !== this.generation ||
        this.reads.get(id) !== token ||
        !this.get(id)
      )
        return { kind: "aborted" };
      this.reads.delete(id);
      const current = this.get(id) ?? prior;
      if (outcome.kind === "cancelled") {
        this.publish(id, { ...current, pending: false });
        return { kind: "aborted" };
      }
      if (
        outcome.kind === "completed" &&
        outcome.value.kind === "accepted" &&
        outcome.value.value.saved_view_id === id &&
        this.ports.visible(outcome.value.value) &&
        (!current.resource ||
          outcome.value.value.saved_view_version >=
            current.resource.saved_view_version)
      ) {
        const resource = structuredClone(outcome.value.value);
        this.publish(id, {
          resource,
          status: "ready",
          pending: false,
          problem: null,
          revision: ++this.revision,
        });
        return { kind: "accepted", value: resource };
      }
      const problem: SavedViewProblem =
        outcome.kind === "completed"
          ? outcome.value.kind !== "accepted"
            ? outcome.value.failure
            : {
                kind: "invalid_contract",
                message:
                  "The resource observation did not match its authority or identity.",
              }
          : {
              kind: "transport",
              message:
                "This saved view could not be refreshed. Retry the resource read.",
            };
      if (problem.kind === "unavailable_target") this.unavailable(id, problem);
      else this.publish(id, { ...current, pending: false, problem });
      this.ports.failed(id, problem);
      return { kind: "rejected", failure: problem };
    });
    return { result, release };
  };
  /** Refresh a retained source-owner resource; no presentation lifetime is borrowed. */
  read = async (id: string): Promise<SavedViewResource | null> => {
    if (!this.get(id)) return null;
    const handle = this.observeResource(id, undefined, false);
    try {
      const result = await handle.result;
      return result.kind === "accepted" ? result.value : null;
    } finally {
      handle.release();
    }
  };
}
