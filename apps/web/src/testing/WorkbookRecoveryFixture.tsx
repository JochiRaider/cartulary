import {
  type ComponentProps,
  type ReactNode,
  type RefObject,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { WorkbookRecoveryBoundary } from "../shared/WorkbookRecoveryBoundary";
import { WorkbookRecoveryNavigation } from "../shared/workbookRecoveryNavigation";
import {
  WorkbookRecoveryEntry,
  WorkbookRecoveryPanel,
} from "../workbook/components/WorkbookRecoveryPanel";
import { WorkbookAuxiliaryDockProvider } from "../workbook/layout/WorkbookAuxiliaryDock";
import { WorkbookSurfaceLayout } from "../workbook/layout/WorkbookSurfaceLayout";

/** Owner content supplies its own production work-area frame. */
export function WorkbookRecoveryProvidersFixture({
  children,
  navigation: provided,
  invokerRef: providedInvoker,
  fallbackRef: providedFallback,
}: {
  readonly children: ReactNode;
  readonly navigation?: WorkbookRecoveryNavigation;
  readonly invokerRef?: RefObject<HTMLElement | null>;
  readonly fallbackRef?: RefObject<HTMLElement | null>;
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
    <WorkbookAuxiliaryDockProvider>
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
      </WorkbookRecoveryBoundary>
    </WorkbookAuxiliaryDockProvider>
  );
}

/** Isolated owners receive the production frame without a query runtime. */
export function WorkbookRecoveryFixture(
  props: ComponentProps<typeof WorkbookRecoveryProvidersFixture>,
) {
  const ownFallback = useRef<HTMLElement>(null);
  const fallbackRef = props.fallbackRef ?? ownFallback;
  return (
    <WorkbookRecoveryProvidersFixture {...props} fallbackRef={fallbackRef}>
      <div style={{ position: "relative", height: "40rem" }}>
        <WorkbookSurfaceLayout
          workAreaOnly
          viewSchemaId="test"
          viewBar={null}
          statusStrip={null}
          primaryGrid={
            <section
              aria-label="Workbook grid focus target"
              tabIndex={-1}
              ref={fallbackRef}
            >
              {props.children}
            </section>
          }
        />
      </div>
    </WorkbookRecoveryProvidersFixture>
  );
}
