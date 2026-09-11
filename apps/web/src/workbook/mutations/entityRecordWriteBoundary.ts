/** Local admission only. Server destructive locks and authorization remain authoritative. */
export type EntityRecordWriteTarget = {
  readonly recordIds: readonly string[];
  // Creates/upserts can identify an existing entity only at the server.
  readonly unknownEntityType?: "host" | "identity";
};
export interface EntityRecordWriteBoundary {
  begin(target: EntityRecordWriteTarget): (() => void) | null;
  acceptVersion(recordId: string, rowVersion: number): void;
}
