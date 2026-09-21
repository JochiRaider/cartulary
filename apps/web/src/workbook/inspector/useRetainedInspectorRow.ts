import { useRef } from "react";
import type {
  WorkbookQueryRow,
  WorkbookReadScope,
} from "../query/WorkbookQueryRow";
import {
  sameWorkbookReadScope,
  workbookRowIsAdmissible,
} from "../query/workbookRowObservation";

/** Retains one source-accepted observation independently of loaded-window membership. */
export function useRetainedInspectorRow<Row>(input: {
  readonly recordId: string | null;
  readonly rows: readonly (Row | null | undefined)[];
  readonly sourceRow: (row: Row) => WorkbookQueryRow | null;
  readonly scope: WorkbookReadScope | null;
  readonly readable: boolean;
}): Row | null {
  const retained = useRef<{
    recordId: string;
    row: Row;
    scope: WorkbookReadScope;
  } | null>(null);
  if (!input.readable || !input.scope) {
    retained.current = null;
    return null;
  }
  if (!input.recordId) {
    retained.current = null;
    return null;
  }
  if (
    !sameWorkbookReadScope(retained.current?.scope, input.scope) ||
    retained.current?.recordId !== input.recordId
  )
    retained.current = null;
  for (const row of input.rows) {
    if (!row) continue;
    const source = input.sourceRow(row);
    if (
      !source ||
      !workbookRowIsAdmissible(source, input.recordId, input.scope)
    )
      continue;
    const prior = retained.current && input.sourceRow(retained.current.row);
    if (!prior || source.row_version >= prior.row_version)
      retained.current = { recordId: input.recordId, row, scope: input.scope };
  }
  return retained.current?.row ?? null;
}
