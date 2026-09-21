import { useRef } from "react";

/** One authorized inspector source, independent of loaded-window membership. */
export function useRetainedInspectorRow<Row>(input: {
  readonly recordId: string | null;
  readonly row: Row | null;
  readonly rowVersion: (row: Row) => number;
  readonly scope: string;
  readonly readable: boolean;
}): Row | null {
  const retained = useRef<{ recordId: string; row: Row; scope: string } | null>(
    null,
  );
  const authority = useRef(input.scope);
  const previousRow = useRef(input.row);
  const retiredRow = useRef<Row | null>(null);
  if (authority.current !== input.scope) {
    authority.current = input.scope;
    // Retire the old observation, not a fresh authorized row arriving in the
    // same render as the new access scope. Presentation resets are separate.
    retiredRow.current = previousRow.current;
    retained.current = null;
  }
  previousRow.current = input.row;
  if (!input.readable || !input.recordId) {
    if (!input.readable) retiredRow.current = input.row;
    retained.current = null;
    return null;
  }
  if (
    retained.current?.scope !== input.scope ||
    retained.current.recordId !== input.recordId
  )
    retained.current = null;
  if (
    input.row &&
    input.row !== retiredRow.current &&
    (!retained.current ||
      input.rowVersion(input.row) >= input.rowVersion(retained.current.row))
  ) {
    retained.current = {
      recordId: input.recordId,
      row: input.row,
      scope: input.scope,
    };
  }
  return retained.current?.row ?? null;
}
