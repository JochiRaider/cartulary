import type {
  ViewContract,
  ViewFieldContract,
} from "@cartulary/view-contracts";
import type {
  WorkbookProtocolCreateViewRowReceipt,
  WorkbookProtocolCreateViewRowRequest,
} from "../../adapters/workbookProtocolTypes";
import type { WorkbookAuthoringSelection } from "../../ports/WorkbookAuthoringReadPort";
import type { WorkbookQueryRow } from "../../query/WorkbookQueryRow";

/** Missing key, explicit null, and raw text are different authoring intentions. */
export type OrdinaryCreateValues = Readonly<Record<string, string | null>>;
export type OrdinaryCreatePreparation = Readonly<{
  request: WorkbookProtocolCreateViewRowRequest | null;
  errors: Readonly<Record<string, string>>;
}>;
export type OrdinaryCreateContribution = Readonly<{
  validateReceipt?(
    contract: ViewContract,
    body: string,
    receipt: WorkbookProtocolCreateViewRowReceipt,
    status: number,
  ): boolean;
  reserve?(contract: ViewContract): (() => void) | null;
  accepted?(row: WorkbookQueryRow): void;
  views: readonly string[];
  prepare(
    contract: ViewContract,
    values: OrdinaryCreateValues,
    clientTxnId: string,
  ): OrdinaryCreatePreparation;
  defaults?(
    contract: ViewContract,
    actorId: string,
  ): Readonly<Record<string, string>>;
  referenceViews(field: ViewFieldContract): readonly string[];
}>;
export type OrdinaryCreateDraft = Readonly<{
  id: number;
  revision: number;
  values: OrdinaryCreateValues;
  references: Readonly<Record<string, readonly WorkbookAuthoringSelection[]>>;
}>;
