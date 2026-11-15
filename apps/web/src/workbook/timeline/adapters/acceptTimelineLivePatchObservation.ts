import { acceptWorkbookRowObservation } from "../../query/acceptWorkbookRowObservation";
import type {
  WorkbookQueryRow,
  WorkbookReadScope,
} from "../../query/WorkbookQueryRow";
import { workbookRowIsAdmissible } from "../../query/workbookRowObservation";

/**
 * A live Timeline row patch is a source observation. A prior row that was
 * admissible for the same record and read scope keeps accepted authority at
 * the patched version; any other patched row carries no new evidence.
 */
export function acceptTimelineLivePatchObservation<
  Row extends WorkbookQueryRow,
>(
  previous: WorkbookQueryRow,
  patched: Row,
  recordId: string,
  scope: WorkbookReadScope | null,
): Row {
  return workbookRowIsAdmissible(previous, recordId, scope)
    ? acceptWorkbookRowObservation(patched, scope)
    : patched;
}
