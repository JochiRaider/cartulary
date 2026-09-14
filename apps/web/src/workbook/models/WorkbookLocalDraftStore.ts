/** Memory-only scalar drafts keyed by a source owner's opaque row and editor identities.
 * Mounted controls, capture algorithms and listeners remain source-owned.
 */
export class WorkbookLocalDraftStore {
  readonly draftValues = new Map<string, string>();
  readonly focusKeysByRow = new Map<string, Set<string>>();

  clear(): void {
    this.draftValues.clear();
    this.focusKeysByRow.clear();
  }
}
