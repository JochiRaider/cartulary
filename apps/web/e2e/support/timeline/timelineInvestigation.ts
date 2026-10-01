import type {
  HTTPOperationID,
  HTTPOperationRequest,
} from "@cartulary/protocol-ts/http";
import type { Page } from "@playwright/test";
import { seedTimelineInvestigation } from "../../../../../tools/harness/fixtures/timeline-investigation/index.mjs";
import { csrfHeaders } from "../auth/browserSession";
import { createUploadedEvidenceFixture } from "../evidence/fixtures";
import { createIncident } from "../incidents/fixtures";
import { apiBase } from "../runtime/configuration";
import { uniqueIncidentKey } from "../runtime/fixtureIdentity";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

export async function seedVisualTimelineInvestigation(
  page: Page,
  options: {
    continuationCount: number;
    incident?: { key: string; title: string };
  },
) {
  return seedTimelineInvestigation({
    incident: options.incident ?? {
      key: uniqueIncidentKey("INVESTIGATION"),
      title: "Service-account investigation",
    },
    continuationCount: options.continuationCount,
    port: {
      createIncident: (incident: { key: string; title: string }) =>
        createIncident(page, incident.key, incident.title),
      call: async (
        operationID: HTTPOperationID,
        body: unknown,
        pathParameters: Record<string, string>,
      ) => {
        const response = await publicHttpOperation({
          operationID,
          body: body as HTTPOperationRequest<HTTPOperationID>,
          pathParameters,
          headers: await csrfHeaders(page),
          request: atJsonOrigin(page.request, apiBase),
        });
        if (!response.ok)
          throw new Error(
            `Investigation ${operationID} failed: HTTP ${response.status}`,
          );
        const payload = response.payload;
        if (
          typeof payload !== "object" ||
          payload === null ||
          !("data" in payload)
        )
          throw new Error(`Invalid ${operationID} success envelope`);
        return payload.data;
      },
      uploadEvidence: (
        incidentId: string,
        evidence: {
          content: string;
          filename: string;
          title: string;
          key: string;
        },
      ) =>
        createUploadedEvidenceFixture(page, incidentId, {
          body: Buffer.from(evidence.content),
          contentType: "text/plain",
          filename: evidence.filename,
          title: evidence.title,
          requestedAt: "2026-04-18T15:06:12Z",
          collectorPartyText: "Jordan Ellis",
          txnPrefix: evidence.key,
        }),
    },
  });
}
