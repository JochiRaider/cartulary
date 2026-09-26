import { useLayoutEffect, useRef, useState } from "react";
import { type AppWorkflowOptions, createAppWorkflows } from "./AppWorkflows";

/** One committed subscription; useAppSession owns terminal scheduling in App. */
export function useAppWorkflows(options: AppWorkflowOptions) {
  const committed = useRef(options);
  const connected = useRef<AppWorkflowOptions["sessionController"] | null>(
    null,
  );
  const [workflows] = useState(() =>
    createAppWorkflows(() => committed.current),
  );
  useLayoutEffect(() => {
    if (committed.current.sessionController !== options.sessionController)
      workflows.retire();
    committed.current = options;
    if (connected.current === options.sessionController)
      workflows.synchronize();
  });
  useLayoutEffect(() => {
    const release = workflows.connect(options.sessionController);
    connected.current = options.sessionController;
    return () => {
      connected.current = null;
      release();
    };
  }, [workflows, options.sessionController]);
  return workflows;
}
