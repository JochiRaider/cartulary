import {
  listWorkbookSurfaceContracts,
  requireViewContract,
} from "@cartulary/view-contracts";
import { coreIndicatorTypes } from "../../adapters/observationProtocol";
import type { IndicatorObservation } from "../../mutations/workbookMutationCommandPorts";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

export { coreIndicatorTypes as observationTypes };
export const observationIndicatorView = "cartulary.view.indicators.v1";
export type ObservationType = (typeof coreIndicatorTypes)[number];
export type ObservationSource = Readonly<{
  incidentId: string;
  viewSchemaId: string;
  recordId: string;
  rowVersion: number;
  fieldKey: string;
  text: string;
}>;
export type ObservationSelection = Readonly<{
  startByte: number;
  endByte: number;
  text: string;
}>;
export type ObservationTarget = Readonly<{
  recordId: string;
  label: string;
  type: string;
}>;
export type ObservationDraft = Readonly<{
  key: string;
  revision: number;
  source: ObservationSource | null;
  selection: ObservationSelection | null;
  parsedType: ObservationType | "";
  target: ObservationTarget | null;
}>;

export function observationUUID(value: unknown): value is string {
  return (
    typeof value === "string" &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/u.test(
      value,
    ) &&
    value !== "00000000-0000-0000-0000-000000000000"
  );
}
export function observationVersion(value: unknown): value is number {
  return typeof value === "number" && Number.isSafeInteger(value) && value > 0;
}
export function observationTimeKey(value: string): string | null {
  const match =
    /^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2})(?:\.(\d{1,9}))?(Z|[+-](?:[01]\d|2[0-3]):[0-5]\d)$/u.exec(
      value,
    );
  if (
    !match ||
    !Number.isFinite(Date.parse(value)) ||
    new Date(`${match[1]}Z`).toISOString().slice(0, 19) !== match[1]
  )
    return null;
  return `${new Date(value).toISOString().slice(0, 19)}.${(match[2] ?? "").padEnd(9, "0")}`;
}
export function observationOrder(
  a: IndicatorObservation,
  b: IndicatorObservation,
): boolean {
  const left = observationTimeKey(a.created_at),
    right = observationTimeKey(b.created_at);
  return (
    left !== null &&
    right !== null &&
    (left > right || (left === right && a.observation_id >= b.observation_id))
  );
}
export function freezeObservation<T>(value: T): T {
  if (value !== null && typeof value === "object") {
    for (const child of Object.values(value)) freezeObservation(child);
    Object.freeze(value);
  }
  return value;
}

/** Matches the source-text port's active-view/read-kind rule, independent of editors. */
export function observationSourceFields(viewId: string, row: WorkbookQueryRow) {
  const contract = requireViewContract(viewId);
  const surface = listWorkbookSurfaceContracts().find(
    (item) => item.viewSchemaId === viewId,
  );
  if (!surface) return [];
  return contract.fields.flatMap((field) => {
    const value = row.cells[field.fieldKey]?.value;
    if (field.readKind !== "text" || typeof value !== "string") return [];
    const owners = listWorkbookSurfaceContracts().filter(
      (candidate) =>
        candidate.sourceRecordTypes.some((type) =>
          surface.sourceRecordTypes.includes(type),
        ) &&
        candidate.contract.fields.some(
          (entry) =>
            entry.fieldKey === field.fieldKey && entry.readKind === "text",
        ),
    );
    return owners.length === 1 && owners[0]?.viewSchemaId === viewId
      ? [{ fieldKey: field.fieldKey, label: field.label, value }]
      : [];
  });
}

/** HTML textarea positions count UTF-16 units and expose CR/CRLF as LF.
 * Each displayed boundary maps directly to the original, never to a text match. */
export function observationTextMap(original: string) {
  let display = "";
  const boundaries = [0];
  for (let index = 0; index < original.length; index++) {
    const char = original[index];
    if (char === "\r") {
      if (original[index + 1] === "\n") index++;
      display += "\n";
    } else display += char;
    boundaries.push(index + 1);
  }
  return { display, boundaries };
}
function scalarBoundary(text: string, offset: number) {
  return !(
    offset > 0 &&
    offset < text.length &&
    /[\uD800-\uDBFF]/u.test(text[offset - 1] ?? "") &&
    /[\uDC00-\uDFFF]/u.test(text[offset] ?? "")
  );
}
export function observationSelection(
  original: string,
  start: number,
  end: number,
): ObservationSelection | null {
  const { boundaries } = observationTextMap(original);
  const from = boundaries[start],
    to = boundaries[end];
  if (
    !Number.isInteger(start) ||
    !Number.isInteger(end) ||
    from === undefined ||
    to === undefined ||
    from >= to ||
    !scalarBoundary(original, from) ||
    !scalarBoundary(original, to) ||
    /[\uD800-\uDBFF](?![\uDC00-\uDFFF])|(?<![\uD800-\uDBFF])[\uDC00-\uDFFF]/u.test(
      original,
    )
  )
    return null;
  const text = original.slice(from, to);
  if (text.includes("\0")) return null;
  const encoder = new TextEncoder();
  return {
    startByte: encoder.encode(original.slice(0, from)).length,
    endByte: encoder.encode(original.slice(0, to)).length,
    text,
  };
}
export function sameObservationSource(
  a: ObservationSource | null,
  b: ObservationSource | null,
): boolean {
  return (
    a === b ||
    (!!a &&
      !!b &&
      a.incidentId === b.incidentId &&
      a.viewSchemaId === b.viewSchemaId &&
      a.recordId === b.recordId &&
      a.rowVersion === b.rowVersion &&
      a.fieldKey === b.fieldKey &&
      a.text === b.text)
  );
}
export function observationCanTransition(
  item: IndicatorObservation,
  action: "resolve" | "dismiss" | "restore",
  target?: string,
): boolean {
  if (action === "restore") return item.resolution_status === "dismissed";
  if (action === "dismiss")
    return (
      item.resolution_status === "unresolved" ||
      item.resolution_status === "resolved"
    );
  return (
    observationUUID(target) &&
    (item.resolution_status === "unresolved" ||
      (item.resolution_status === "resolved" &&
        item.resolved_indicator_record_id !== target))
  );
}
