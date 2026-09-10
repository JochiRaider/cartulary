import {
  act,
  cleanup,
  fireEvent,
  render,
  screen,
} from "@testing-library/react";
import { useState } from "react";
import { afterEach, expect, it, vi } from "vitest";
import { explorationFixture } from "./explorationTestFixtures";
import {
  type ExplorationContributorPage,
  NetworkFlowExplorationPanel,
} from "./NetworkFlowExplorationPanel";
import {
  type ExplorationAction,
  emptyExploration,
  explorationFocusCurrent,
  transitionExploration,
} from "./networkFlowExplorationNavigation";

const page: ExplorationContributorPage = {
  items: [],
  loadState: "ready",
  loadGenerationKey: 0,
  error: null,
  pageNumber: 1,
  canNext: false,
  canPrevious: false,
  pending: null,
  failed: null,
  notice: null,
  recovery: "none",
  nextPage: () => {},
  canRefresh: false,
  canRestart: false,
  refresh: () => {},
  previousPage: () => {},
  retry: () => {},
  restart: () => {},
};
afterEach(cleanup);
it("withdraws removed graph focus safely and cancels an older semantic focus intent", () => {
  const graph = explorationFixture(2, 1, true);
  let navigate: (action: ExplorationAction) => void = () => {};
  const bind = vi.fn(() => () => {});
  function Subject() {
    const [navigation, setNavigation] = useState(() =>
      transitionExploration(emptyExploration("reader"), {
        type: "accept",
        result: graph,
      }),
    );
    navigate = (action) =>
      setNavigation((current) => transitionExploration(current, action));
    return (
      <>
        <button type="button">Newer interaction</button>
        <NetworkFlowExplorationPanel
          navigation={navigation}
          contributorPage={page}
          status={{ loadState: "ready", stale: false, validationMessage: null }}
          tables={[]}
          canLink
          onNavigate={navigate}
          onRefreshGraph={() => {}}
          onLinkEdge={() => {}}
          onLinkVertex={() => {}}
          isFocusCurrent={(intent) =>
            explorationFocusCurrent(navigation, intent)
          }
          bindFocusRestoration={bind}
        />
      </>
    );
  }
  render(<Subject />);
  const selected = screen.getAllByRole("button", { name: /^Select edge/u })[0];
  if (!selected) throw new Error("Missing edge control");
  selected.focus();
  fireEvent.click(selected);
  expect(screen.queryByText("Unavailable endpoint")).toBeNull();
  const close = screen.getByRole("button", {
    name: "Close graph contributors",
  });
  close.focus();
  act(() => navigate({ type: "clear" }));
  expect(document.activeElement).toBe(
    screen.getByRole("region", { name: "Network Flow graph" }),
  );
  act(() => navigate({ type: "accept", result: graph }));
  const current = screen.getAllByRole("button", { name: /^Select edge/u })[0];
  if (!current) throw new Error("Missing current edge");
  fireEvent.click(current);
  const newer = screen.getByRole("button", { name: "Newer interaction" });
  act(() => {
    navigate({ type: "close" });
    navigate({ type: "bucket", index: 2 });
    newer.focus();
  });
  expect(document.activeElement).toBe(newer);
  expect(
    screen.queryByRole("complementary", { name: "Graph contributors" }),
  ).toBeNull();
});
