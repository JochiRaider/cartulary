import { useCallback, useEffect, useRef, useState } from "react";
import type {
  NetworkFlowIndicatorSelector,
  NetworkFlowIndicatorTarget,
} from "./networkFlowClient";
import {
  getNetworkFlowBindingSourceRowLimit,
  linkNetworkFlowIndicator,
} from "./networkFlowClient";
import {
  type NetworkFlowWorkspaceError,
  networkFlowErrorFromUnknown,
} from "./networkFlowErrors";

export type NetworkFlowIndicatorLinkCandidate = {
  readonly candidateValue: string;
  readonly key: string;
  readonly label: string;
  readonly selector: NetworkFlowIndicatorSelector;
};

export function useNetworkFlowIndicatorLinkController({
  activeCandidateKey,
  apiBase,
  enabled,
  incidentId,
  onError,
  onGraphStale,
  onMessage,
}: {
  readonly activeCandidateKey: string | null;
  readonly apiBase: string | undefined;
  readonly enabled: boolean;
  readonly incidentId: string;
  readonly onError: (error: NetworkFlowWorkspaceError | null) => void;
  readonly onGraphStale: () => void;
  readonly onMessage: (message: string) => void;
}) {
  const [bindingSourceRowLimit, setBindingSourceRowLimit] = useState(0);
  const [linking, setLinking] = useState(false);
  const candidateKeyRef = useRef(activeCandidateKey);
  const generationRef = useRef(0);
  candidateKeyRef.current = activeCandidateKey;

  useEffect(() => {
    void activeCandidateKey;
    void enabled;
    void incidentId;
    generationRef.current += 1;
    setLinking(false);
  }, [activeCandidateKey, enabled, incidentId]);

  useEffect(() => {
    if (!enabled) {
      setBindingSourceRowLimit(0);
      return;
    }
    const controller = new AbortController();
    void getNetworkFlowBindingSourceRowLimit({
      apiBase,
      incidentId,
      signal: controller.signal,
    })
      .then((limit) => {
        if (!controller.signal.aborted) {
          setBindingSourceRowLimit(limit);
        }
      })
      .catch((caught: unknown) => {
        if (!controller.signal.aborted) {
          setBindingSourceRowLimit(0);
          onError(
            networkFlowErrorFromUnknown(
              caught,
              "Network Flow link limits could not be loaded.",
            ),
          );
        }
      });
    return () => controller.abort();
  }, [apiBase, enabled, incidentId, onError]);

  const link = useCallback(
    async (options: {
      readonly candidate: NetworkFlowIndicatorLinkCandidate;
      readonly confirmExactValue: string;
      readonly target: NetworkFlowIndicatorTarget;
    }): Promise<boolean> => {
      if (
        !enabled ||
        linking ||
        options.confirmExactValue !== options.candidate.candidateValue ||
        candidateKeyRef.current !== options.candidate.key
      ) {
        return false;
      }
      generationRef.current += 1;
      const generation = generationRef.current;
      const candidateKey = options.candidate.key;
      setLinking(true);
      onError(null);
      try {
        const result = await linkNetworkFlowIndicator({
          apiBase,
          confirmExactValue: options.confirmExactValue,
          incidentId,
          selector: options.candidate.selector,
          target: options.target,
        });
        if (
          generation !== generationRef.current ||
          candidateKey !== candidateKeyRef.current
        ) {
          return false;
        }
        onMessage(
          result.duplicate
            ? "Indicator link already exists."
            : "Indicator link created.",
        );
        onError(null);
        setLinking(false);
        return true;
      } catch (caught) {
        if (
          generation !== generationRef.current ||
          candidateKey !== candidateKeyRef.current
        ) {
          return false;
        }
        const requestError = networkFlowErrorFromUnknown(
          caught,
          "Network Flow indicator link failed.",
        );
        if (requestError.code === "network_flow_graph_query_stale") {
          onGraphStale();
        }
        onError(requestError);
        setLinking(false);
        return false;
      }
    },
    [apiBase, enabled, incidentId, linking, onError, onGraphStale, onMessage],
  );

  return { bindingSourceRowLimit, link, linking };
}
