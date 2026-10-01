import type { WorkbookMutationAuthority } from "../../mutations/workbookMutationAuthority";
import type {
  TimelineBulkTagAuthoringPort,
  TimelineBulkTagAuthoringSnapshot,
} from "../ports/TimelineBulkTagAuthoringPort";

/** Timeline-only raw command authoring, independent of selection and batches. */
export class TimelineBulkTagAuthoring {
  private draft = { raw: "", revision: 0 };
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private generation = 0;
  private attachment: object | null = null;
  private retired = false;
  private snapshot: TimelineBulkTagAuthoringSnapshot | null = null;
  private readonly listeners = new Set<() => void>();

  constructor(private readonly incidentId: string) {}

  getSnapshot = () => this.snapshot;
  subscribe = (listener: () => void) => {
    if (this.retired) return () => {};
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };

  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (this.retired) return;
    if (
      authority &&
      (authority.incidentId !== this.incidentId ||
        (this.actorId !== null && authority.actorId !== this.actorId))
    ) {
      this.retire();
      return;
    }
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    this.authority = authority ? Object.freeze({ ...authority }) : null;
    if (authority) this.actorId = authority.actorId;
    this.generation++;
    this.publish();
  }
  suspend() {
    this.setAuthority(null);
  }
  closeIncident() {
    if (this.authority) this.setAuthority({ ...this.authority, closed: true });
  }
  retire() {
    if (this.retired) return;
    this.retired = true;
    this.draft = { raw: "", revision: this.draft.revision + 1 };
    this.authority = null;
    this.actorId = null;
    this.attachment = null;
    this.generation++;
    this.publish();
    this.listeners.clear();
  }

  /** Preparing a binding has no effects; its layout effect commits the token. */
  bind(): TimelineBulkTagAuthoringPort {
    const token = {};
    const isCurrent = (generation: number | undefined) =>
      !this.retired &&
      this.attachment === token &&
      this.snapshot !== null &&
      generation === this.generation;
    const replace = (raw: string) => {
      this.draft = { raw, revision: this.draft.revision + 1 };
      this.publish();
      return true;
    };
    return {
      getSnapshot: this.getSnapshot,
      subscribe: this.subscribe,
      attach: () => {
        if (!this.retired) this.attachment = token;
        return () => {
          if (this.attachment === token) this.attachment = null;
        };
      },
      isCurrent,
      update: (raw, generation) =>
        isCurrent(generation) &&
        this.snapshot?.canEdit === true &&
        replace(raw),
      clear: (generation) => isCurrent(generation) && replace(""),
    };
  }

  private publish() {
    this.snapshot =
      !this.retired && this.authority?.role
        ? Object.freeze({
            ...this.draft,
            generation: this.generation,
            canEdit:
              !this.authority.closed &&
              ["editor", "reviewer", "admin"].includes(this.authority.role),
          })
        : null;
    for (const listener of this.listeners) listener();
  }
}
