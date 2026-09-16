import {
  type ReactNode,
  type RefObject,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { WorkbookRecoveryBoundary } from "../shared/WorkbookRecoveryBoundary";
import {
  WorkbookWorkAreaOverlayHost,
  WorkbookWorkAreaOverlayProvider,
} from "../shared/WorkbookWorkAreaOverlay";
import { WorkbookRecoveryNavigation } from "../shared/workbookRecoveryNavigation";
import {
  WorkbookRecoveryEntry,
  WorkbookRecoveryPanel,
} from "../workbook/components/WorkbookRecoveryPanel";

/** Isolated owners use the same entry, attachment and geometry as the production shell. */
export function WorkbookRecoveryFixture({
  children,
  navigation: provided,
  invokerRef: providedInvoker,
  fallbackRef: providedFallback,
  standalone = true,
}: {
  readonly children: ReactNode;
  readonly navigation?: WorkbookRecoveryNavigation;
  readonly invokerRef?: RefObject<HTMLElement | null>;
  readonly fallbackRef?: RefObject<HTMLElement | null>;
  readonly standalone?: boolean;
}) {
  const own = useMemo(() => new WorkbookRecoveryNavigation(), []);
  const navigation = provided ?? own;
  const ownInvoker = useRef<HTMLElement>(null);
  const ownFallback = useRef<HTMLElement>(null);
  const invokerRef = providedInvoker ?? ownInvoker;
  const fallbackRef = providedFallback ?? ownFallback;
  const [host, setHost] = useState<HTMLDivElement | null>(null);
  useEffect(() => () => own.dispose(), [own]);
  return (
    <WorkbookWorkAreaOverlayProvider>
      <WorkbookRecoveryBoundary
        navigation={navigation}
        detailHost={host}
        invokerRef={invokerRef}
      >
        <WorkbookRecoveryEntry
          navigation={navigation}
          invokerRef={invokerRef}
        />
        <WorkbookRecoveryPanel
          navigation={navigation}
          registerDetailHost={setHost}
          invokerRef={invokerRef}
          fallbackRef={fallbackRef}
        />
        {children}
        {standalone ? (
          <div style={{ position: "relative", height: "40rem" }}>
            <section
              aria-label="Workbook grid focus target"
              tabIndex={-1}
              ref={fallbackRef}
            />
            <WorkbookWorkAreaOverlayHost />
          </div>
        ) : null}
      </WorkbookRecoveryBoundary>
    </WorkbookWorkAreaOverlayProvider>
  );
}
