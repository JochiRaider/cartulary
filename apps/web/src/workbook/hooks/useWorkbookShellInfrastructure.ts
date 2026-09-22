import { useEffect, useMemo, useRef, useSyncExternalStore } from "react";
import type { ExtensionAvailabilityController } from "../../extensions/extensionAvailability";
import { createWorkbookClipboardPasteAdapter } from "../adapters/createWorkbookClipboardPasteAdapter";
import { createWorkbookIncidentAdapter } from "../adapters/createWorkbookIncidentAdapter";
import { createWorkbookReferenceMemberReader } from "../adapters/createWorkbookReferenceMemberReader";
import { createWorkbookStartupAdapter } from "../adapters/createWorkbookStartupAdapter";
import { createWorkbookViewQueryAdapter } from "../adapters/createWorkbookViewQueryAdapter";
import { createWorkbookMutationCommandPorts } from "../mutations/createWorkbookMutationCommandPorts";
import { createBrowserSecureTransactionIdPort } from "../mutations/secureTransactionId";
import type { WorkbookMutationRuntime } from "../runtime/WorkbookMutationRuntime";
import type { SavedViewBinding } from "../savedviews/savedViewOperationModel";
import type { WorkbookSavedViewController } from "../savedviews/WorkbookSavedViewController";
import { createWorkbookReferenceReader } from "../services/workbookReferenceReader";
import { timelineCaptureOwnerFor } from "../timeline/actions/timelineCaptureOwnerFor";
import { timelineMentionOwnerFor } from "../timeline/actions/timelineMentionOwnerFor";
import { useWorkbookShellRuntime } from "./useWorkbookShellRuntime";

type WorkbookShellInfrastructureOptions = {
  readonly acceptedAuthority: Readonly<{
    userId: string | null;
    role: string | null;
    revision: number;
  }>;
  readonly sessionIdentity: string | null;
  readonly savedViewOwner: WorkbookSavedViewController;
  readonly bindWorkbookSavedViews: (binding: SavedViewBinding | null) => void;
  readonly authorizationRecovered: SavedViewBinding["authorizationRecovered"];
  readonly apiBase: string | undefined;
  readonly extensionAvailability: ExtensionAvailabilityController;
  readonly incidentId: string;
  readonly mutationRuntime: WorkbookMutationRuntime;
  readonly onExtensionAvailabilityChange: () => void;
  readonly onAuthorityUncertain: (() => void) | undefined;
};

/** Pure adapters and presentation state over the already committed retained runtime. */
export function useWorkbookShellInfrastructure({
  acceptedAuthority,
  sessionIdentity,
  savedViewOwner,
  bindWorkbookSavedViews,
  authorizationRecovered,
  apiBase,
  extensionAvailability,
  incidentId,
  mutationRuntime,
  onExtensionAvailabilityChange,
  onAuthorityUncertain,
}: WorkbookShellInfrastructureOptions) {
  const surfaceSelectionVersionRef = useRef(0);
  const timelineCapture = timelineCaptureOwnerFor(mutationRuntime);
  const timelineMentions = timelineMentionOwnerFor(mutationRuntime);
  const clipboardPastePort = useMemo(
    () => createWorkbookClipboardPasteAdapter(mutationRuntime.batches),
    [mutationRuntime],
  );
  const incidentPort = useMemo(
    () => createWorkbookIncidentAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  const transactionIds = useMemo(createBrowserSecureTransactionIdPort, []);
  const mutationCommands = useMemo(
    () =>
      createWorkbookMutationCommandPorts({
        readScope: () => mutationRuntime.recordReadScope,
        batches: mutationRuntime.batches,
        apiBase,
        incidentId,
        transactionIds,
      }),
    [apiBase, incidentId, mutationRuntime, transactionIds],
  );
  const startupPort = useMemo(
    () => createWorkbookStartupAdapter({ apiBase, incidentId }),
    [apiBase, incidentId],
  );
  // Rebind reads only when accepted read authority changes. Rechecking the
  // same authority must not abort an acknowledged operation's refresh.
  const readScopeKey = useSyncExternalStore(mutationRuntime.subscribe, () =>
    JSON.stringify(mutationRuntime.recordReadScope),
  );
  const viewQuery = useMemo(
    () =>
      createWorkbookViewQueryAdapter({
        apiBase,
        incidentId,
        readScope: () => {
          const scope = mutationRuntime.recordReadScope;
          return acceptedAuthority.role &&
            scope?.actorId === acceptedAuthority.userId &&
            scope?.sessionIdentity === sessionIdentity &&
            JSON.stringify(scope) === readScopeKey
            ? scope
            : null;
        },
      }),
    [
      apiBase,
      incidentId,
      mutationRuntime,
      acceptedAuthority.role,
      acceptedAuthority.userId,
      sessionIdentity,
      readScopeKey,
    ],
  );
  const workbookRuntime = useWorkbookShellRuntime({
    incidentId,
    onAuthorityUncertain,
    surfaceSelectionVersionRef,
    extensionAvailability,
    onExtensionAvailabilityChange,
    savedViewOwner,
    bindWorkbookSavedViews,
    authorizationRecovered,
    apiBase,
    startupPort,
  });

  return {
    clipboardPastePort,
    incidentPort,
    mutationCommands,
    mutationRuntime,
    timelineCapture,
    timelineMentions,
    viewQuery,
    workbookRuntime,
  };
}

export function useWorkbookReferenceReader(
  authorityScope: string,
  viewQuery: ReturnType<typeof createWorkbookViewQueryAdapter>,
  apiBase: string | undefined,
  incidentId: string,
) {
  const reader = useMemo(
    () =>
      createWorkbookReferenceReader({
        authorityScope,
        viewQuery,
        readMembers: createWorkbookReferenceMemberReader({
          apiBase,
          incidentId,
        }),
      }),
    [authorityScope, viewQuery, apiBase, incidentId],
  );
  useEffect(() => () => reader.dispose(), [reader]);
  return reader;
}
