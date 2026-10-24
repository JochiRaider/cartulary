import type {
  APIResponse,
  BrowserContext,
  Page,
  Route,
} from "@playwright/test";
import { expect } from "@playwright/test";
import type { OwnedBrowserScope } from "../runtime/cleanup";
import { apiBase } from "../runtime/configuration";
import { publicHttpOperation } from "../transport/publicHttpOperationClient";
import { atJsonOrigin } from "../transport/publicJsonClient";

// These feature admissions return the public Job envelope. Presentation fixtures
// intercept above this context route, so synthetic Jobs never enter this owner.
const jobAdmissionPath =
  /^\/api\/v1\/(?:object-blobs\/[^/]+\/verify|incident-bundles\/(?:import|export)|reference-packs\/(?:import|refresh|[^/]+\/[^/]+\/(?:activate|disable|reverify))|import-sessions(?:\/[^/]+\/apply)?|snapshots|releases)$/u;
const owners = new WeakMap<BrowserContext, OwnedJobs>();

/** Own admission before dispatch, including replies lost before browser delivery. */
export class OwnedJobs {
  readonly #pending = new Set<Promise<void>>();
  readonly #jobs = new Map<string, Promise<void>>();
  #stopping = false;
  #failed = false;
  #completion: Promise<void> | undefined;

  constructor(
    private readonly observeTerminal: (
      jobId: string,
      headers: Record<string, string>,
    ) => Promise<void>,
  ) {}

  admit<T>(
    dispatch: () => Promise<T>,
    identify: (
      response: T,
    ) => Promise<{ jobId: string; headers: Record<string, string> } | null>,
  ): Promise<T> {
    if (this.#stopping) throw new Error("Feature job admission has stopped");
    const response = Promise.resolve()
      .then(dispatch)
      .then(async (result) => {
        const job = await identify(result);
        if (job !== null && !this.#jobs.has(job.jobId)) {
          const terminal = Promise.resolve()
            .then(() => this.observeTerminal(job.jobId, job.headers))
            .catch(() => {
              this.#failed = true;
            });
          this.#jobs.set(job.jobId, terminal);
        }
        return result;
      });
    const settled = response.then(
      () => {},
      () => {
        this.#failed = true;
      },
    );
    this.#pending.add(settled);
    void settled.then(() => this.#pending.delete(settled));
    return response;
  }

  settle(): Promise<void> {
    if (this.#completion !== undefined) return this.#completion;
    this.#stopping = true;
    this.#completion = (async () => {
      await Promise.all(this.#pending);
      await Promise.all(this.#jobs.values());
      this.#jobs.clear();
      if (this.#failed) throw new Error("Feature job settlement failed");
    })();
    return this.#completion;
  }

  fetch(route: Route): Promise<APIResponse> {
    return this.admit(
      async () => {
        const headers = await route.request().allHeaders();
        const response = await route.fetch();
        return { response, headers: { cookie: headers.cookie ?? "" } };
      },
      async ({ response, headers }) => {
        if (response.status() !== 202) return null;
        const envelope: unknown = await response.json();
        if (!envelope || typeof envelope !== "object" || !("data" in envelope))
          throw new Error("Invalid feature job admission envelope");
        const job = envelope.data;
        if (
          !job ||
          typeof job !== "object" ||
          !("job_id" in job) ||
          typeof job.job_id !== "string" ||
          !job.job_id
        )
          throw new Error("Invalid feature job admission identity");
        // Use the authority that actually dispatched the request, independent of
        // later account replacement or cookie changes in the owning context.
        return { jobId: job.job_id, headers };
      },
    ).then(({ response }) => response);
  }
}

export async function installOwnedJobs(
  context: BrowserContext,
  scope: OwnedBrowserScope,
  timeoutMs: number,
) {
  const jobs = new OwnedJobs(async (jobId, headers) => {
    await expect
      .poll(
        async () => {
          const result = await publicHttpOperation({
            operationID: "getJob",
            pathParameters: { job_id: jobId },
            headers,
            request: atJsonOrigin(context.request, apiBase),
          });
          if (!result.ok)
            throw new Error("Owned feature job observation failed");
          return result.payload.data.status;
        },
        { timeout: timeoutMs },
      )
      .toMatch(/^(succeeded|failed|canceled)$/u);
  });
  owners.set(context, jobs);
  scope.settleBeforeClosing(() => jobs.settle(), context);
  await scope.acquireRoute(() =>
    context.route(
      (url) => jobAdmissionPath.test(url.pathname),
      async (route) => {
        if (route.request().method() !== "POST") {
          await route.fallback();
          return;
        }
        try {
          await route.fulfill({ response: await jobs.fetch(route) });
        } catch {
          scope.recordRouteFailure();
          throw new Error("Owned feature job route failed");
        }
      },
    ),
  );
}

/** A scenario which fetches upstream itself must register that real admission. */
export function fetchOwnedJobResponse(
  page: Page,
  route: Route,
): Promise<APIResponse> {
  if (
    route.request().method() !== "POST" ||
    !jobAdmissionPath.test(new URL(route.request().url()).pathname)
  )
    return route.fetch();
  return ownedJobsFor(page).fetch(route);
}

export function ownedJobsFor(page: Page): OwnedJobs {
  const jobs = owners.get(page.context());
  if (!jobs) throw new Error("Feature jobs have no resource owner");
  return jobs;
}
