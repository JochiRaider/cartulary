/** Existing account/session/incident authority plus its runtime invalidation epoch. */
export type WorkbookReadScope = Readonly<{
  actorId: string;
  sessionIdentity: string;
  incidentId: string;
  epoch: number;
}>;
export type WorkbookRowObservation = Readonly<{
  recordId: string;
  rowVersion: number;
  scope: WorkbookReadScope;
}>;

export type WorkbookReadScopeSource = () => WorkbookReadScope | null;

export type WorkbookQueryRow = {
  readonly observation?: WorkbookRowObservation;
  readonly record_id: string;
  readonly row_version: number;
  readonly cells: Record<string, { readonly value: unknown }>;
  readonly group_values?: Record<string, unknown>;
};
