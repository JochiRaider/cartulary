import { isDeepStrictEqual } from "node:util";
import { expect, it } from "vitest";
import {
  decodeFixtureJSON,
  runNetworkFlowFixture,
} from "../testing/networkFlowFixtureTestSupport";
import type {
  NetworkFlowMappingColumn,
  NetworkFlowUnknownColumnPolicy,
} from "./networkFlowImportModel";
import {
  buildNetworkFlowMappingCandidate,
  createNetworkFlowMappingDraft,
  networkFlowMappingDraftReadyForPreview,
  networkFlowMappingIssues,
  sourceColumnLabel,
  withNetworkFlowColumnChoice,
} from "./networkFlowImportModel";

type AliasCase = {
  name: string;
  columns: NetworkFlowMappingColumn[];
  unknown_column_policy: NetworkFlowUnknownColumnPolicy;
  choices: { ordinal: number; choice: string | null }[];
};

it("Alias fixture compares deterministic suggestions blocking collisions and preview mappings", async () => {
  await runNetworkFlowFixture(
    { fixtureID: "NF-FIX-011-alias-collision", revision: 4 },
    async (fixture) => {
      const cases = (await decodeFixtureJSON(
        fixture.read("source/client-cases.json"),
      )) as AliasCase[];
      const outcomes = cases.map((test) => {
        let draft = {
          ...createNetworkFlowMappingDraft(test.columns),
          unknownColumnPolicy: test.unknown_column_policy,
        };
        for (const step of test.choices) {
          draft = withNetworkFlowColumnChoice(draft, step.ordinal, step.choice);
        }
        return {
          case: test.name,
          request: test,
          result: {
            column_choices: draft.columnChoices,
            blocking_alias_ordinals: draft.unresolvedAliasCollisionOrdinals,
            preview_enabled:
              networkFlowMappingDraftReadyForPreview(draft) &&
              networkFlowMappingIssues(draft).length === 0,
            candidate: buildNetworkFlowMappingCandidate(draft, test.columns),
            source_labels: test.columns.map(sourceColumnLabel),
          },
        };
      });
      const actual = {
        fixture_id: "NF-FIX-011-alias-collision",
        freeze_revision: 4,
        outcomes,
      };
      fixture.record(
        "actual-client.json",
        JSON.stringify(actual, null, 2) + "\n",
      );
      for (const name of [
        "expected/alias-collision.client.expected.json",
        "transcripts/alias-collision.client.transcript.json",
      ]) {
        const expected = await decodeFixtureJSON(fixture.read(name));
        if (!isDeepStrictEqual(expected, actual))
          throw new Error("Client fixture comparison differs: " + name);
        const wrong = structuredClone(actual);
        const first = wrong.outcomes[0];
        if (!first) throw new Error("Alias fixture has no observable outcomes");
        first.result.preview_enabled = !first.result.preview_enabled;
        expect(isDeepStrictEqual(expected, wrong)).toBe(false);
      }
    },
  );
});
