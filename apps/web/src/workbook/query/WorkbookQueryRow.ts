export type WorkbookQueryRow = {
  readonly record_id: string;
  readonly row_version: number;
  readonly cells: Record<string, { readonly value: unknown }>;
  readonly group_values?: Record<string, unknown>;
};
