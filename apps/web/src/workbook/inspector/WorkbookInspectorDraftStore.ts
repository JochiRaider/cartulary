import { workbookSavedFieldEqual } from "../models/workbookSavedValues";
import type { WorkbookMutationAuthority } from "../mutations/workbookMutationAuthority";
import type { WorkbookAuthoringCandidate } from "../ports/WorkbookAuthoringReadPort";
import type { WorkbookQueryRow } from "../query/WorkbookQueryRow";

export type InspectorEditIdentity = Readonly<{
  viewSchemaId: string;
  recordId: string;
  fieldKey: string;
  action: string;
}>;
export type InspectorEditDraft = Readonly<{
  identity: InspectorEditIdentity;
  baseline: WorkbookQueryRow;
  value: string | null;
  references: readonly WorkbookAuthoringCandidate[];
  revision: number;
  attachment: string | null;
}>;
export type InspectorDraftCapture = Readonly<{ key: string; revision: number }>;
export function inspectorEditKey(identity: InspectorEditIdentity): string {
  return JSON.stringify([
    identity.viewSchemaId,
    identity.recordId,
    identity.fieldKey,
    identity.action,
  ]);
}
/** Raw ordinary inspector authoring only. Dispatch and domain review are separate owners. */
export class WorkbookInspectorDraftStore {
  private authority: WorkbookMutationAuthority | null = null;
  private actorId: string | null = null;
  private incidentId: string | null = null;
  private retired = false;
  private revision = 0;
  private sequence = 0;
  private readonly drafts = new Map<string, InspectorEditDraft>();
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
    if (JSON.stringify(this.authority) === JSON.stringify(authority)) return;
    this.authority = authority;
    if (authority) {
      this.actorId = authority.actorId;
      this.incidentId = authority.incidentId;
    }
    for (const [key, draft] of this.drafts)
      this.drafts.set(key, { ...draft, attachment: null });
    this.emit();
  }
  canRead() {
    return (
      !this.retired && this.authority !== null && this.authority.role !== ""
    );
  }
  canAuthor() {
    return (
      this.canRead() &&
      this.authority?.role !== "viewer" &&
      this.authority?.role !== "" &&
      !this.authority?.closed
    );
  }
  read(identity: InspectorEditIdentity): InspectorEditDraft | null {
    return this.canRead()
      ? (this.drafts.get(inspectorEditKey(identity)) ?? null)
      : null;
  }
  update(
    identity: InspectorEditIdentity,
    row: WorkbookQueryRow,
    value: string | null,
    attachment: string,
    references?: readonly WorkbookAuthoringCandidate[],
  ) {
    if (!this.canAuthor() || row.record_id !== identity.recordId) return;
    const key = inspectorEditKey(identity),
      previous = this.drafts.get(key);
    if (previous && previous.attachment !== attachment) return;
    this.drafts.set(key, {
      identity: structuredClone(identity),
      baseline: previous?.baseline ?? structuredClone(row),
      value,
      references: structuredClone(references ?? previous?.references ?? []),
      revision: ++this.sequence,
      attachment,
    });
    this.emit();
  }
  detach(attachment: string) {
    let changed = false;
    for (const [key, draft] of this.drafts)
      if (draft.attachment === attachment) {
        this.drafts.set(key, { ...draft, attachment: null });
        changed = true;
      }
    if (changed) this.emit();
  }
  resume(identity: InspectorEditIdentity, attachment: string) {
    const key = inspectorEditKey(identity),
      draft = this.read(identity);
    if (!draft || !this.canAuthor()) return;
    this.drafts.set(key, { ...draft, attachment });
    this.emit();
  }
  discard(identity: InspectorEditIdentity) {
    if (this.canRead() && this.drafts.delete(inspectorEditKey(identity)))
      this.emit();
  }
  staleFields(
    identity: InspectorEditIdentity,
    row: WorkbookQueryRow,
    dependencies: readonly string[],
  ) {
    const draft = this.read(identity);
    return draft
      ? dependencies.filter(
          (field) => !workbookSavedFieldEqual(draft.baseline, row, field),
        )
      : [];
  }
  review(
    identity: InspectorEditIdentity,
    row: WorkbookQueryRow,
    field: string,
    keepDraft: boolean,
  ) {
    const key = inspectorEditKey(identity),
      draft = this.read(identity);
    if (!draft || !this.canAuthor() || row.record_id !== identity.recordId)
      return;
    if (!keepDraft && field === identity.fieldKey) {
      this.discard(identity);
      return;
    }
    this.drafts.set(key, {
      ...draft,
      baseline: {
        ...draft.baseline,
        row_version: row.row_version,
        cells: {
          ...draft.baseline.cells,
          [field]: structuredClone(row.cells[field] ?? { value: null }),
        },
      },
      revision: ++this.sequence,
    });
    this.emit();
  }
  capture(identity: InspectorEditIdentity): InspectorDraftCapture | null {
    const draft = this.read(identity);
    return draft
      ? { key: inspectorEditKey(identity), revision: draft.revision }
      : null;
  }
  acknowledge(capture: InspectorDraftCapture | null) {
    if (
      capture &&
      this.drafts.get(capture.key)?.revision === capture.revision
    ) {
      this.drafts.delete(capture.key);
      this.emit();
    }
  }
  hasRecords(ids: readonly string[]) {
    return (
      this.canRead() &&
      [...this.drafts.values()].some((draft) =>
        ids.includes(draft.identity.recordId),
      )
    );
  }
  discardRecords(ids: readonly string[]) {
    if (!this.canRead()) return;
    for (const [key, draft] of this.drafts)
      if (ids.includes(draft.identity.recordId)) this.drafts.delete(key);
    this.emit();
  }
  retire() {
    this.retired = true;
    this.authority = null;
    this.drafts.clear();
    this.emit();
  }
}
