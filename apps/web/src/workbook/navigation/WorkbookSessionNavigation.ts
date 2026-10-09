import { cartularyDesignPresentation } from "@cartulary/ui-contracts";
import { sheetRefKey } from "../../shared/sheetRef";
import {
  type WorkbookNavigationIntent,
  type WorkbookReturnOrigin,
  type WorkbookSessionPin,
  workbookNavigationIntentsEqual,
} from "./WorkbookNavigationIntent";

export type {
  WorkbookReturnOrigin,
  WorkbookSessionPin,
} from "./WorkbookNavigationIntent";

const limits = cartularyDesignPresentation.workbookWorkbench;
export function workbookPinIdentity(pin: WorkbookSessionPin) {
  return JSON.stringify([
    pin.incidentId,
    sheetRefKey(pin.sheetRef),
    pin.recordId ?? null,
  ]);
}
type WorkbookNavigationOutcome =
  | "idle"
  | "pending"
  | "admitted"
  | "succeeded"
  | "failed"
  | "cancelled";
type Snapshot = {
  readonly attemptId: number;
  readonly outcome: WorkbookNavigationOutcome;
  readonly pins: readonly WorkbookSessionPin[];
  readonly trail: readonly WorkbookReturnOrigin[];
  readonly message: string | null;
  readonly readable: boolean;
  readonly pending: boolean;
};

/** Session navigation metadata only. Reads and writes stay with their owners. */
export class WorkbookSessionNavigation {
  private pins: readonly WorkbookSessionPin[] = [];
  private trail: readonly WorkbookReturnOrigin[] = [];
  private listeners = new Set<() => void>();
  // Admission settles the trail; its presentation may still be mounting/focusing.
  private intent: AbortController | null = null;
  private pending: {
    intent: WorkbookNavigationIntent;
    promise: Promise<boolean>;
  } | null = null;
  private snapshot: Snapshot = {
    attemptId: 0,
    outcome: "idle",
    pins: [],
    trail: [],
    message: null,
    readable: false,
    pending: false,
  };
  constructor(readonly incidentId: string) {}
  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  private publish(message: string | null = this.snapshot.message) {
    this.snapshot = {
      ...this.snapshot,
      pins: this.snapshot.readable ? this.pins : [],
      trail: this.snapshot.readable ? this.trail : [],
      pending:
        this.snapshot.outcome === "pending" ||
        this.snapshot.outcome === "admitted",
      message: this.snapshot.readable ? message : null,
    };
    for (const listener of this.listeners) listener();
  }
  setReadable(readable: boolean) {
    if (this.snapshot.readable === readable) return;
    if (!readable) this.cancel();
    this.snapshot = { ...this.snapshot, readable };
    this.publish(null);
  }
  pin(pin: WorkbookSessionPin) {
    if (
      !this.snapshot.readable ||
      pin.incidentId !== this.incidentId ||
      (pin.recordId && pin.sheetRef.kind !== "view_schema")
    )
      return;
    if (
      this.pins.some(
        (item) => workbookPinIdentity(item) === workbookPinIdentity(pin),
      )
    ) {
      this.publish("Already in working set");
      return;
    }
    if (this.pins.length >= limits.pin_limit) {
      this.publish("Working set is full; remove a pin first.");
      return;
    }
    this.pins = [...this.pins, structuredClone(pin)];
    this.publish("Added to working set");
  }
  remove(pin: WorkbookSessionPin) {
    this.pins = this.pins.filter(
      (item) => workbookPinIdentity(item) !== workbookPinIdentity(pin),
    );
    this.publish(null);
  }
  concealPin(pin: WorkbookSessionPin) {
    this.pins = this.pins.map((item) =>
      workbookPinIdentity(item) === workbookPinIdentity(pin)
        ? { ...item, label: "Unavailable item" }
        : item,
    );
    this.publish("This item is unavailable.");
  }
  completePresentation(
    attemptId: number,
    outcome: "succeeded" | "failed" | "cancelled" = "succeeded",
  ) {
    if (
      this.snapshot.attemptId !== attemptId ||
      this.snapshot.outcome !== "admitted" ||
      !this.snapshot.readable
    )
      return;
    this.snapshot = { ...this.snapshot, outcome };
    this.pending = null;
    this.publish();
  }
  cancel() {
    if (
      this.snapshot.outcome === "pending" ||
      this.snapshot.outcome === "admitted"
    )
      this.snapshot = { ...this.snapshot, outcome: "cancelled" };
    this.intent?.abort();
    this.intent = null;
    this.pending = null;
    this.publish(null);
  }
  clear() {
    this.cancel();
    this.pins = [];
    this.trail = [];
    this.publish(null);
  }
  /** prepare/commit is one owner-admitted navigation, never an optimistic pivot. */
  navigate(
    intent: WorkbookNavigationIntent,
    origin: WorkbookReturnOrigin,
    commit: (signal: AbortSignal) => Promise<"changed" | "same" | "failed">,
    returning = false,
  ): Promise<boolean> {
    if (!this.snapshot.readable || origin.incidentId !== this.incidentId)
      return Promise.resolve(false);
    if (
      this.pending &&
      workbookNavigationIntentsEqual(this.pending.intent, intent)
    )
      return this.pending.promise;
    this.cancel();
    const captured = structuredClone(origin);
    const controller = new AbortController();
    this.intent = controller;
    const request = {
      intent: structuredClone(intent),
      promise: Promise.resolve(false),
    };
    this.pending = request;
    this.snapshot = {
      ...this.snapshot,
      attemptId: this.snapshot.attemptId + 1,
      outcome: "pending",
    };
    request.promise = Promise.resolve()
      .then(() => commit(controller.signal))
      .then((result) => {
        if (
          this.pending !== request ||
          controller.signal.aborted ||
          !this.snapshot.readable
        )
          return false;
        if (result === "failed") {
          this.snapshot = { ...this.snapshot, outcome: "failed" };
          return false;
        }
        this.snapshot = { ...this.snapshot, outcome: "admitted" };
        if (returning) this.trail = this.trail.slice(0, -1);
        else if (result === "changed")
          this.trail = [...this.trail, captured].slice(-limits.return_limit);
        return true;
      })
      .catch(() => {
        if (this.pending === request)
          this.snapshot = { ...this.snapshot, outcome: "failed" };
        return false;
      })
      .finally(() => {
        if (this.pending === request) {
          if (this.snapshot.outcome !== "admitted") this.pending = null;
          this.publish(null);
        }
      });
    this.publish(null);
    return request.promise;
  }
}
