import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";
import { workbookSavedFieldEqual } from "./workbookSavedValues";

export type WorkbookGridDraftIdentity = Readonly<{
  viewSchemaId: string;
  recordId: string;
  fieldKey: string;
}>;
export type WorkbookGridDraftCapture = Readonly<{
  key: string;
  revision: number;
}>;
type WorkbookGridDraft = Readonly<{
  identity: WorkbookGridDraftIdentity;
  value: string | null;
  baseline: WorkbookQueryRow;
  revision: number;
  validation: string | null;
}>;
const workbookGridDraftKey = (identity: WorkbookGridDraftIdentity) =>
  JSON.stringify([
    identity.viewSchemaId,
    identity.recordId,
    identity.fieldKey,
    "grid",
  ]);

/** Raw grid authoring only. Runtime identity owns account/incident/client lifetime. */
export class WorkbookGridDraftStore {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private incidentId: string | null = null;
  private retired = false;
  private revision = 0;
  private readonly drafts = new Map<string, WorkbookGridDraft>();
  private readonly listeners = new Set<() => void>();
  subscribe = (listener: () => void) => {
    this.listeners.add(listener);
    return () => {
      this.listeners.delete(listener);
    };
  };
  getSnapshot = () => this.revision;
  private emit() {
    this.revision++;
    for (const listener of this.listeners) listener();
  }
  setAuthority(authority: WorkbookMutationAuthority | null) {
    if (this.retired) return;
    if (
      authority &&
      ((this.actorId && authority.actorId !== this.actorId) ||
        (this.incidentId && authority.incidentId !== this.incidentId))
    ) {
      this.retire();
      return;
    }
    if (JSON.stringify(authority) === JSON.stringify(this.authority)) return;
    this.authority = authority;
    if (authority) {
      this.actorId = authority.actorId;
      this.incidentId = authority.incidentId;
    }
    this.emit();
  }
  canRead() {
    return !this.retired && !!this.authority && this.authority.role !== "";
  }
  canAuthor() {
    return (
      this.canRead() &&
      this.authority?.role !== "viewer" &&
      !this.authority?.closed
    );
  }
  list(viewSchemaId: string) {
    return this.canRead()
      ? [...this.drafts.values()].filter(
          (draft) => draft.identity.viewSchemaId === viewSchemaId,
        )
      : [];
  }
  read(identity: WorkbookGridDraftIdentity) {
    return this.canRead()
      ? (this.drafts.get(workbookGridDraftKey(identity)) ?? null)
      : null;
  }
  update(
    identity: WorkbookGridDraftIdentity,
    row: WorkbookQueryRow,
    value: string | null,
  ) {
    if (!this.canAuthor() || row.record_id !== identity.recordId) return;
    const key = workbookGridDraftKey(identity),
      previous = this.drafts.get(key);
    if (previous?.value === value) return;
    this.drafts.set(key, {
      identity: { ...identity },
      value,
      baseline: previous?.baseline ?? structuredClone(row),
      revision: this.revision + 1,
      validation: null,
    });
    this.emit();
  }
  capture(
    identity: WorkbookGridDraftIdentity,
  ): WorkbookGridDraftCapture | null {
    const draft = this.read(identity);
    return draft
      ? { key: workbookGridDraftKey(identity), revision: draft.revision }
      : null;
  }
  acknowledge(capture: WorkbookGridDraftCapture | null | undefined) {
    if (
      capture &&
      this.drafts.get(capture.key)?.revision === capture.revision
    ) {
      this.drafts.delete(capture.key);
      this.emit();
    }
  }
  setValidation(
    capture: WorkbookGridDraftCapture | null | undefined,
    message: string,
  ) {
    if (!capture) return;
    const draft = this.drafts.get(capture.key);
    if (
      !draft ||
      draft.revision !== capture.revision ||
      draft.validation === message
    )
      return;
    this.drafts.set(capture.key, { ...draft, validation: message });
    this.emit();
  }
  discard(identity: WorkbookGridDraftIdentity) {
    if (this.canRead() && this.drafts.delete(workbookGridDraftKey(identity)))
      this.emit();
  }
  staleFields(
    identity: WorkbookGridDraftIdentity,
    row: WorkbookQueryRow,
    dependencies: readonly string[] = [],
  ) {
    const draft = this.read(identity);
    return draft
      ? [...new Set([identity.fieldKey, ...dependencies])].filter(
          (field) => !workbookSavedFieldEqual(draft.baseline, row, field),
        )
      : [];
  }
  review(identity: WorkbookGridDraftIdentity, row: WorkbookQueryRow) {
    const key = workbookGridDraftKey(identity),
      draft = this.read(identity);
    if (!draft || !this.canAuthor() || row.record_id !== identity.recordId)
      return;
    this.drafts.set(key, {
      ...draft,
      baseline: structuredClone(row),
      revision: this.revision + 1,
      validation: null,
    });
    this.emit();
  }
  /** Advance only fields proven to be written by this draft's own predecessor. */
  acceptPredecessor(
    viewSchemaId: string,
    row: WorkbookQueryRow,
    fields: readonly string[],
  ) {
    if (this.retired) return;
    let changed = false;
    for (const [key, draft] of this.drafts) {
      if (
        draft.identity.viewSchemaId !== viewSchemaId ||
        draft.identity.recordId !== row.record_id ||
        draft.baseline.row_version > row.row_version
      )
        continue;
      const cells = { ...draft.baseline.cells };
      for (const field of fields)
        if (row.cells[field]) cells[field] = structuredClone(row.cells[field]);
      this.drafts.set(key, {
        ...draft,
        baseline: { ...draft.baseline, row_version: row.row_version, cells },
      });
      changed = true;
    }
    if (changed) this.emit();
  }
  retire() {
    this.retired = true;
    this.authority = null;
    this.drafts.clear();
    this.emit();
  }
}
