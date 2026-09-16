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
export type SavedViewResourceSlot =
  | "selected"
  | "activation"
  | "operation"
  | "home"
  | "default";
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
  private reads = new Map<string, { cancel: () => void }>();
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
  cancel = (id: string) => {
    this.reads.get(id)?.cancel();
    this.reads.delete(id);
    const value = this.get(id);
    if (value?.pending) this.publish(id, { ...value, pending: false });
  };
  retain = (slot: SavedViewResourceSlot, id: string | null) => {
    if (this.slots.get(slot) === (id ?? undefined)) return;
    if (id === null) this.slots.delete(slot);
    else this.slots.set(slot, id);
    const keep = new Set(this.slots.values());
    const next = new Map(this.state);
    for (const key of next.keys())
      if (!keep.has(key)) {
        this.reads.get(key)?.cancel();
        this.reads.delete(key);
        next.delete(key);
      }
    if (id !== null && !next.has(id)) next.set(id, empty());
    this.state = next;
    this.ports.changed();
  };
  private publish(id: string, value: SavedViewObservation) {
    if (![...this.slots.values()].includes(id)) return;
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
    this.reads.get(id)?.cancel();
    this.reads.delete(id);
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
    for (const read of this.reads.values()) read.cancel();
    this.reads.clear();
    this.state = new Map(
      [...this.state].map(([id, value]) => [id, { ...value, pending: false }]),
    );
    this.ports.changed();
  };
  unavailable = (id: string, problem: SavedViewProblem) => {
    this.reads.get(id)?.cancel();
    this.reads.delete(id);
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
    this.state = new Map();
    this.ports.changed();
  };
  read = async (id: string): Promise<SavedViewResource | null> => {
    const port = this.ports.port();
    const prior = this.get(id);
    if (!port || !prior) return null;
    this.reads.get(id)?.cancel();
    const generation = this.generation;
    this.publish(id, { ...prior, pending: true, problem: null });
    const read = this.ports.observe((signal) =>
      port.getResource({ savedViewId: id, signal }),
    );
    this.reads.set(id, read);
    const outcome = await read.result;
    if (
      generation !== this.generation ||
      this.reads.get(id) !== read ||
      !this.get(id)
    )
      return null;
    this.reads.delete(id);
    if (
      outcome.kind === "completed" &&
      outcome.value.kind === "accepted" &&
      outcome.value.value.saved_view_id === id &&
      this.ports.visible(outcome.value.value) &&
      (!prior.resource ||
        outcome.value.value.saved_view_version >=
          prior.resource.saved_view_version)
    ) {
      this.accept(outcome.value.value);
      return this.get(id)?.resource ?? null;
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
    else this.publish(id, { ...prior, pending: false, problem });
    this.ports.failed(id, problem);
    return null;
  };
}
