import { createContext, type ReactNode, useContext } from "react";
import { createPortal } from "react-dom";

// Only a presentation destination. Query/editor state stays with its controller.
export const WorkbookQuerySummaryHost = createContext<
  HTMLElement | null | undefined
>(undefined);

export function WorkbookQuerySummarySlot({
  children,
}: {
  readonly children: ReactNode;
}) {
  const host = useContext(WorkbookQuerySummaryHost);
  if (host === undefined) return children;
  return host === null ? null : createPortal(children, host);
}
