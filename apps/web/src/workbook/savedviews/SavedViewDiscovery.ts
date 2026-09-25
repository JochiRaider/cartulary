import type { SavedViewResource } from "../models/workbookSavedViews";
import type {
  SavedViewObserver,
  SavedViewProblem,
  WorkbookSavedViewPort,
} from "../ports/WorkbookSavedViewPort";

const savedViewDiscoveryLimit = 50;
const savedViewPreviousCursorLimit = 10;
type Destination = {
  cursor: string | null;
  previous: readonly (string | null)[];
};
export type SavedViewDiscoveryAction =
  | "first"
  | "previous"
  | "next"
  | "refresh"
  | "retry";
export type SavedViewDiscoverySnapshot = {
  readonly viewSchemaId: string | null;
  readonly open: boolean;
  readonly accepted: boolean;
  readonly candidates: readonly SavedViewResource[];
  readonly cursor: string | null;
  readonly previous: readonly (string | null)[];
  readonly nextCursor: string | null;
  readonly pending: boolean;
  readonly pendingAction: SavedViewDiscoveryAction | null;
  readonly stale: boolean;
  readonly problem: SavedViewProblem | null;
};
export const emptySavedViewDiscovery = (): SavedViewDiscoverySnapshot => ({
  viewSchemaId: null,
  open: false,
  accepted: false,
  candidates: [],
  cursor: null,
  previous: [],
  nextCursor: null,
  pending: false,
  pendingAction: null,
  stale: false,
  problem: null,
});

/** One accepted page and bounded checkpoints; discovery never owns selection. */
export class SavedViewDiscovery {
  private state = emptySavedViewDiscovery();
  private generation = 0;
  private read: { cancel: () => void } | null = null;
  private retryDestination: Destination | null = null;
  constructor(
    private readonly ports: {
      port: () => WorkbookSavedViewPort | null;
      observe: SavedViewObserver;
      changed: () => void;
      failed: (problem: SavedViewProblem) => void;
    },
  ) {}
  getSnapshot = () => this.state;
  private publish(patch: Partial<SavedViewDiscoverySnapshot>) {
    this.state = { ...this.state, ...patch };
    this.ports.changed();
  }
  setSchema = (viewSchemaId: string | null) => {
    if (this.state.viewSchemaId === viewSchemaId) return;
    this.clear();
    this.publish({ viewSchemaId });
  };
  open = () => {
    if (this.state.open) return;
    this.publish({ open: true });
    void this.load(
      { cursor: this.state.cursor, previous: this.state.previous },
      null,
    );
  };
  close = () => {
    this.cancel();
    this.publish({ open: false, pending: false, pendingAction: null });
  };
  first = () => this.load({ cursor: null, previous: [] }, "first");
  refresh = () => this.load({ cursor: null, previous: [] }, "refresh");
  next = () => {
    if (this.state.pending || this.state.nextCursor === null) return;
    return this.load(
      {
        cursor: this.state.nextCursor,
        previous: [...this.state.previous, this.state.cursor].slice(
          -savedViewPreviousCursorLimit,
        ),
      },
      "next",
    );
  };
  previous = () => {
    if (this.state.pending || this.state.previous.length === 0) return;
    return this.load(
      {
        cursor: this.state.previous.at(-1) ?? null,
        previous: this.state.previous.slice(0, -1),
      },
      "previous",
    );
  };
  retry = () =>
    this.retryDestination
      ? this.load(this.retryDestination, "retry")
      : this.first();
  revalidate = () =>
    this.load(
      { cursor: this.state.cursor, previous: this.state.previous },
      null,
    );
  invalidate = (removedId?: string) => {
    this.cancel();
    this.publish({
      pending: false,
      pendingAction: null,
      stale: this.state.accepted,
      ...(removedId
        ? {
            candidates: this.state.candidates.filter(
              (r) => r.saved_view_id !== removedId,
            ),
          }
        : {}),
    });
  };
  clear = () => {
    this.cancel();
    this.retryDestination = null;
    this.state = emptySavedViewDiscovery();
    this.ports.changed();
  };
  private cancel() {
    this.generation++;
    this.read?.cancel();
    this.read = null;
  }
  private async load(
    destination: Destination,
    action: SavedViewDiscoveryAction | null,
  ) {
    const port = this.ports.port();
    const schema = this.state.viewSchemaId;
    if (!port || !schema || !this.state.open || this.state.pending) return;
    this.cancel();
    const generation = this.generation;
    this.retryDestination = destination;
    this.publish({ pending: true, pendingAction: action, problem: null });
    const read = this.ports.observe((signal) =>
      port.listPage({
        viewSchemaId: schema,
        limit: savedViewDiscoveryLimit,
        cursorToken: destination.cursor,
        signal,
      }),
    );
    this.read = read;
    const outcome = await read.result;
    if (
      generation !== this.generation ||
      schema !== this.state.viewSchemaId ||
      this.read !== read
    )
      return;
    this.read = null;
    if (outcome.kind === "completed" && outcome.value.kind === "accepted") {
      const page = outcome.value.value;
      const ids = new Set(page.savedViews.map((r) => r.saved_view_id));
      if (
        page.savedViews.length <= savedViewDiscoveryLimit &&
        ids.size === page.savedViews.length &&
        page.savedViews.every((r) => r.view_schema_id === schema) &&
        (page.nextCursor === null ||
          (page.nextCursor.trim() !== "" &&
            page.nextCursor !== destination.cursor &&
            !destination.previous.includes(page.nextCursor)))
      ) {
        this.retryDestination = null;
        this.publish({
          accepted: true,
          candidates: page.savedViews,
          cursor: destination.cursor,
          previous: destination.previous,
          nextCursor: page.nextCursor,
          pending: false,
          pendingAction: null,
          stale: false,
          problem: null,
        });
        return;
      }
    }
    const problem: SavedViewProblem =
      outcome.kind === "completed"
        ? outcome.value.kind !== "accepted"
          ? outcome.value.failure
          : {
              kind: "invalid_contract",
              message:
                "Saved-view discovery returned an invalid page or repeated cursor.",
            }
        : {
            kind: "transport",
            message: "Saved views could not be loaded. Retry this page.",
          };
    this.publish({
      pending: false,
      pendingAction: null,
      stale: this.state.accepted,
      problem,
    });
    this.ports.failed(problem);
  }
}
