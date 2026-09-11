/** Lifecycle/status bridge only. Timeline owns its action policy and transport. */
export interface WorkbookTimelineActionRuntimePort {
  readonly pendingCount: number;
  readonly blockedCount: number;
  subscribe(listener: () => void): () => void;
  blocksRecord(recordId: string): boolean;
  acceptVersion(recordId: string, rowVersion: number): void;
  suspend(): void;
  closeIncident(): void;
  retire(): void;
}
