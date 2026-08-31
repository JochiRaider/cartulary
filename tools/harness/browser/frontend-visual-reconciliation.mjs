import { createHash } from "node:crypto";
import {
  existsSync,
  readdirSync,
  readFileSync,
  statSync,
} from "node:fs";
import path from "node:path";

import {
  goldenManifestPath,
  rendererProfilePath,
  validateFrontendVisualGoldenManifest,
  visualSnapshotRoot,
} from "./frontend-visual-golden-manifest.mjs";

const captureIntentSchemaID =
  "cartulary.frontend_visual_capture_intent.v1";
const reconciliationSchemaID =
  "cartulary.frontend_visual_reconciliation.v2";
const captureAttachmentPrefix = "cartulary-visual-capture-intent-";
const snapshotRoot = visualSnapshotRoot;
const playwrightConfigPath = "apps/web/playwright.config.ts";
const visualSpecPath = "apps/web/e2e/workbook.visual.spec.ts";
const snapshotPathTemplate =
  "{snapshotDir}/{testFileDir}/{testFileName}-snapshots/{arg}{-snapshotSuffix}{ext}";

function normalizePath(value) {
  return value.replaceAll("\\", "/");
}

function repoRelative(root, value) {
  const absolute = path.isAbsolute(value) ? value : path.resolve(root, value);
  const relative = normalizePath(path.relative(root, absolute));
  if (relative === ".." || relative.startsWith("../")) {
    throw new Error(`artifact path escapes repository root: ${value}`);
  }
  return relative;
}

function readJSON(file) {
  return JSON.parse(readFileSync(file, "utf8"));
}

function sha256File(file) {
  return createHash("sha256").update(readFileSync(file)).digest("hex");
}

function uniqueSorted(values) {
  return [...new Set(values)].sort((left, right) =>
    left.localeCompare(right),
  );
}

function recursivelyListFiles(root) {
  if (!existsSync(root)) {
    return [];
  }
  return readdirSync(root, { withFileTypes: true }).flatMap((entry) => {
    const candidate = path.join(root, entry.name);
    return entry.isDirectory() ? recursivelyListFiles(candidate) : [candidate];
  });
}

function flattenPlaywrightSuites(suites, specs = []) {
  for (const suite of suites ?? []) {
    flattenPlaywrightSuites(suite.suites, specs);
    specs.push(...(suite.specs ?? []));
  }
  return specs;
}

function loadVisualCatalog(root) {
  const ownerPath = path.join(root, "tools/test_catalog_owner.json");
  const ownerCatalog = readJSON(ownerPath);
  const rows = [];
  const manifestPaths = [];
  for (const owner of ownerCatalog.owners ?? []) {
    if (owner.status !== "active") {
      continue;
    }
    const manifestPath = String(owner.manifest_path ?? "");
    if (manifestPath === "") {
      continue;
    }
    const manifest = readJSON(path.join(root, manifestPath));
    const visualRows = (manifest.rows ?? []).filter(
      (row) =>
        row.status === "active" &&
        row.runner === "playwright" &&
        row.selector?.stage === "visual",
    );
    if (visualRows.length === 0) {
      continue;
    }
    manifestPaths.push(manifestPath);
    rows.push(...visualRows);
  }
  return {
    ownerPath: "tools/test_catalog_owner.json",
    manifestPaths: uniqueSorted(manifestPaths),
    rows,
  };
}

function catalogCaptureEntries(catalog) {
  return catalog.rows.flatMap((row) => {
    const titles = row.selector?.titles ?? [];
    const scenarioIDs = row.selector?.scenario_ids ?? [];
    if (titles.length !== scenarioIDs.length) {
      throw new Error(
        `${row.row_id} must pair every visual title with exactly one scenario_id`,
      );
    }
    return titles.map((title, index) => ({
      file: row.selector.file,
      owner_id: row.owner_id,
      project_id: row.selector.project_id,
      row_id: row.row_id,
      scenario_id: scenarioIDs[index],
      title,
    }));
  });
}

function sourceRef(root, kind, sourcePath, options = {}) {
  const absolute = path.join(root, sourcePath);
  return {
    kind,
    path: sourcePath,
    sha256: sha256File(absolute),
    project_ids: options.projectIDs ?? [],
    snapshot_path_template: options.snapshotTemplate ?? null,
    symbols: options.symbols ?? [],
  };
}

function resolveAttachmentPath(root, reportPath, attachmentPath) {
  if (path.isAbsolute(attachmentPath)) {
    return attachmentPath;
  }
  const candidates = [
    path.resolve(root, attachmentPath),
    path.resolve(path.dirname(reportPath), attachmentPath),
  ];
  return candidates.find((candidate) => existsSync(candidate)) ?? candidates[0];
}

function validateCapturePayload(payload) {
  const requiredStrings = [
    "capture_id",
    "capture_intent",
    "expected_golden_path",
    "project_id",
    "renderer_profile_id",
    "screenshot_assertion_location",
    "test_file",
    "test_title",
  ];
  if (payload?.schema_id !== captureIntentSchemaID) {
    throw new Error(`unexpected capture intent schema ${payload?.schema_id}`);
  }
  for (const field of requiredStrings) {
    if (typeof payload[field] !== "string" || payload[field] === "") {
      throw new Error(`capture intent requires non-empty ${field}`);
    }
  }
  if (
    typeof payload.capture_profile !== "object" ||
    payload.capture_profile === null
  ) {
    throw new Error("capture intent requires capture_profile");
  }
  if (
    !payload.expected_golden_path.startsWith(`${snapshotRoot}/`) ||
    !payload.expected_golden_path.endsWith(".png") ||
    payload.expected_golden_path.includes("../")
  ) {
    throw new Error(
      `capture intent has invalid expected_golden_path ${payload.expected_golden_path}`,
    );
  }
}

function capturePayloadsFromReports(root, reportPaths) {
  const payloads = [];
  const artifactRefs = [];
  const errors = [];
  for (const reportPath of reportPaths) {
    artifactRefs.push(repoRelative(root, reportPath));
    const report = readJSON(reportPath);
    for (const spec of flattenPlaywrightSuites(report.suites)) {
      for (const playwrightTest of spec.tests ?? []) {
        for (const result of playwrightTest.results ?? []) {
          for (const attachment of result.attachments ?? []) {
            if (
              typeof attachment?.name !== "string" ||
              !attachment.name.startsWith(captureAttachmentPrefix)
            ) {
              continue;
            }
            const attachmentPath =
              typeof attachment.path === "string" && attachment.path !== ""
                ? resolveAttachmentPath(root, reportPath, attachment.path)
                : null;
            try {
              const payload =
                attachmentPath === null
                  ? JSON.parse(
                      Buffer.from(String(attachment.body ?? ""), "base64").toString(
                        "utf8",
                      ),
                    )
                  : readJSON(attachmentPath);
              validateCapturePayload(payload);
              if (
                playwrightTest.projectName &&
                playwrightTest.projectName !== payload.project_id
              ) {
                throw new Error(
                  `capture ${payload.capture_id} project differs from its Playwright result`,
                );
              }
              payloads.push(payload);
              if (attachmentPath !== null) {
                artifactRefs.push(repoRelative(root, attachmentPath));
              }
            } catch (error) {
              errors.push(
                `${attachmentPath === null ? `${repoRelative(root, reportPath)}#${attachment.name}` : repoRelative(root, attachmentPath)}: ${error instanceof Error ? error.message : String(error)}`,
              );
            }
          }
        }
      }
    }
  }
  return { payloads, artifactRefs: uniqueSorted(artifactRefs), errors };
}

export function classifyFrontendVisualGoldens({
  captureIntents,
  committedGoldens,
  fixtures,
  consumerRefs = new Map(),
}) {
  const capturesByGolden = Map.groupBy(
    captureIntents,
    (capture) => capture.expected_golden_path,
  );
  const fixturesByGolden = new Map();
  const designContractsByGolden = new Map();
  for (const fixture of fixtures) {
    for (const goldenPath of fixture.golden_artifacts ?? []) {
      const fixtureIDs = fixturesByGolden.get(goldenPath) ?? [];
      fixtureIDs.push(fixture.fixture_id);
      fixturesByGolden.set(goldenPath, fixtureIDs);
      if (fixture.design_contract_id) {
        const designContractIDs =
          designContractsByGolden.get(goldenPath) ?? [];
        designContractIDs.push(fixture.design_contract_id);
        designContractsByGolden.set(goldenPath, designContractIDs);
      }
    }
  }
  const allPaths = uniqueSorted([
    ...capturesByGolden.keys(),
    ...committedGoldens.keys(),
    ...fixturesByGolden.keys(),
    ...consumerRefs.keys(),
  ]);
  const goldens = allPaths.map((goldenPath) => {
    const captures = capturesByGolden.get(goldenPath) ?? [];
    const committedDigest = committedGoldens.get(goldenPath) ?? null;
    const externalConsumers = consumerRefs.get(goldenPath) ?? [];
    let classification;
    if (captures.length > 1) {
      classification = "ambiguous_mapping";
    } else if (
      committedDigest === null &&
      (captures.length === 1 || fixturesByGolden.has(goldenPath))
    ) {
      classification = "missing_golden";
    } else if (committedDigest !== null && captures.length === 1) {
      classification = "active";
    } else if (
      committedDigest !== null &&
      captures.length === 0 &&
      externalConsumers.length === 0
    ) {
      classification = "orphan";
    } else if (committedDigest !== null) {
      classification = "active";
    } else {
      classification = "ambiguous_mapping";
    }
    return {
      golden_path: goldenPath,
      sha256: committedDigest,
      consumer_capture_ids: uniqueSorted(
        captures.map((capture) => capture.capture_id),
      ),
      catalog_row_ids: uniqueSorted(
        captures.map((capture) => capture.row_id),
      ),
      owner_ids: uniqueSorted(captures.map((capture) => capture.owner_id)),
      scenario_ids: uniqueSorted(
        captures.map((capture) => capture.scenario_id),
      ),
      project_ids: uniqueSorted(
        captures.map((capture) => capture.project_id),
      ),
      fixture_ids: uniqueSorted(fixturesByGolden.get(goldenPath) ?? []),
      design_contract_ids: uniqueSorted(
        designContractsByGolden.get(goldenPath) ?? [],
      ),
      consumer_refs: uniqueSorted(externalConsumers),
      classification,
      permitted_action: classification === "active" ? "retain" : "blocked",
    };
  });
  return goldens;
}

export function resolveRegisteredFixtures(
  fixtures,
  goldens,
  catalogEntries,
  captureIntents,
) {
  const goldenByPath = new Map(
    goldens.map((golden) => [golden.golden_path, golden]),
  );
  const catalogByRow = Map.groupBy(catalogEntries, (entry) => entry.row_id);
  const selectedRowIDs = new Set(
    captureIntents.map((capture) => capture.row_id),
  );
  return fixtures.filter((fixture) => {
    const rowIDs = fixture.catalog_row_ids ?? [];
    if (
      rowIDs.length === 0 ||
      rowIDs.some((rowID) => !catalogByRow.has(rowID))
    ) {
      return true;
    }
    const scenarioTitleResolves = rowIDs.some((rowID) =>
      (catalogByRow.get(rowID) ?? []).some(
        (entry) => entry.title === fixture.playwright_scenario_title,
      ),
    );
    if (!scenarioTitleResolves) {
      return true;
    }
    const fixtureGoldens = (fixture.golden_artifacts ?? []).map(
      (goldenPath) => goldenByPath.get(goldenPath),
    );
    if (
      fixtureGoldens.some(
        (golden) => golden?.sha256 === null || golden === undefined,
      )
    ) {
      return true;
    }
    return rowIDs
      .filter((rowID) => selectedRowIDs.has(rowID))
      .some(
        (rowID) =>
          !fixtureGoldens.some((golden) =>
            golden.catalog_row_ids.includes(rowID),
          ),
      );
  });
}

export function buildFrontendVisualReconciliation({
  root,
  reportPaths,
  attemptPassed,
  snapshotDirectory = path.join(root, snapshotRoot),
  rendererAttestationPaths = [],
  candidateGoldenManifest = null,
}) {
  const catalog = loadVisualCatalog(root);
  const catalogEntries = catalogCaptureEntries(catalog);
  const registryPath = path.join(
    root,
    "tools/frontend_visual_fixture_registry.json",
  );
  const registry = readJSON(registryPath);
  const { payloads, artifactRefs, errors } = capturePayloadsFromReports(
    root,
    reportPaths,
  );
  const catalogEntryByKey = Map.groupBy(
    catalogEntries,
    (entry) => `${entry.file}\u0000${entry.project_id}\u0000${entry.title}`,
  );
  const captureByID = new Map();
  for (const payload of payloads) {
    const candidates =
      catalogEntryByKey.get(
        `${payload.test_file}\u0000${payload.project_id}\u0000${payload.test_title}`,
      ) ?? [];
    if (candidates.length !== 1) {
      errors.push(
        `${payload.capture_id}: expected one catalog row/scenario, resolved ${candidates.length}`,
      );
      continue;
    }
    const candidate = candidates[0];
    const capture = {
      capture_id: payload.capture_id,
      owner_id: candidate.owner_id,
      row_id: candidate.row_id,
      scenario_id: candidate.scenario_id,
      project_id: payload.project_id,
      renderer_profile_id: payload.renderer_profile_id,
      screenshot_assertion_location: payload.screenshot_assertion_location,
      assertion_file: payload.test_file,
      capture_intent: payload.capture_intent,
      screenshot_name_source: payload.capture_intent,
      capture_profile: payload.capture_profile,
      expected_golden_path: payload.expected_golden_path,
      test_title: payload.test_title,
    };
    const existing = captureByID.get(capture.capture_id);
    if (existing && JSON.stringify(existing) !== JSON.stringify(capture)) {
      errors.push(`${capture.capture_id}: conflicting duplicate capture intent`);
      continue;
    }
    captureByID.set(capture.capture_id, capture);
  }
  const captureIntents = [...captureByID.values()].sort((left, right) =>
    left.capture_id.localeCompare(right.capture_id),
  );
  const committedGoldens = new Map(
    recursivelyListFiles(snapshotDirectory)
      .filter((file) => file.endsWith(".png") && statSync(file).isFile())
      .map((file) => [
        `${snapshotRoot}/${path.basename(file)}`,
        sha256File(file),
      ]),
  );
  const goldens = classifyFrontendVisualGoldens({
    captureIntents,
    committedGoldens,
    fixtures: registry.fixtures ?? [],
  });
  const unresolvedFixtures = resolveRegisteredFixtures(
    registry.fixtures ?? [],
    goldens,
    catalogEntries,
    captureIntents,
  );
  const countClassification = (classification) =>
    goldens.filter((golden) => golden.classification === classification).length;
  const counts = {
    capture_intents: captureIntents.length,
    committed_goldens: committedGoldens.size,
    active: countClassification("active"),
    orphan: countClassification("orphan"),
    missing_golden: countClassification("missing_golden"),
    ambiguous_mapping: countClassification("ambiguous_mapping"),
    registered_fixtures: (registry.fixtures ?? []).length,
    unresolved_registered_fixtures: unresolvedFixtures.length,
  };
  if (!attemptPassed) {
    errors.push("visual target attempt did not complete successfully");
  }
  if (captureIntents.length === 0) {
    errors.push("visual target emitted no capture intents");
  }
  if (counts.missing_golden > 0) {
    errors.push(`${counts.missing_golden} active or registered golden(s) are missing`);
  }
  if (counts.ambiguous_mapping > 0) {
    errors.push(`${counts.ambiguous_mapping} golden mapping(s) are ambiguous`);
  }
  if (counts.unresolved_registered_fixtures > 0) {
    errors.push(
      `${counts.unresolved_registered_fixtures} registered fixture(s) do not resolve`,
    );
  }
  const rendererProfile = readJSON(path.join(root, rendererProfilePath));
  if (rendererAttestationPaths.length === 0) {
    errors.push("visual target retained no renderer attestation");
  }
  for (const attestationPath of rendererAttestationPaths) {
    try {
      const attestation = readJSON(attestationPath);
      if (JSON.stringify(attestation) !== JSON.stringify(rendererProfile)) {
        errors.push(`${repoRelative(root, attestationPath)}: renderer attestation mismatch`);
      } else {
        artifactRefs.push(repoRelative(root, attestationPath));
      }
    } catch (error) {
      errors.push(
        `${repoRelative(root, attestationPath)}: ${error instanceof Error ? error.message : String(error)}`,
      );
    }
  }
  const distinctCaptureRendererIDs = uniqueSorted(
    captureIntents.map((capture) => capture.renderer_profile_id),
  );
  if (
    distinctCaptureRendererIDs.length !== 1 ||
    distinctCaptureRendererIDs[0] !== rendererProfile.profile_id
  ) {
    errors.push("visual capture intents do not use the active renderer profile");
  }
  const manifestFile = path.join(root, goldenManifestPath);
  let manifest = candidateGoldenManifest;
  let manifestBytes = null;
  let manifestStatus = "pass";
  if (manifest === null) {
    if (!existsSync(manifestFile)) {
      manifestStatus = "missing";
      errors.push("frontend visual golden manifest is missing");
    } else {
      manifestBytes = readFileSync(manifestFile);
      manifest = JSON.parse(manifestBytes.toString("utf8"));
    }
  } else {
    manifestBytes = Buffer.from(`${JSON.stringify(manifest, null, 2)}\n`, "utf8");
  }
  if (manifest !== null) {
    const manifestErrors = validateFrontendVisualGoldenManifest(manifest, {
      root,
      snapshotRoot: snapshotDirectory,
    });
    if (manifestErrors.length > 0) {
      manifestStatus = "mismatch";
      errors.push(...manifestErrors);
    }
  }
  const projectIDs = uniqueSorted(
    catalogEntries.map((entry) => entry.project_id),
  );
  const configSource = readFileSync(path.join(root, playwrightConfigPath), "utf8");
  if (!configSource.includes(snapshotPathTemplate)) {
    errors.push("Playwright snapshot template does not match the reconciler");
  }
  const sourceRefs = [
    sourceRef(root, "catalog_owner", catalog.ownerPath),
    ...catalog.manifestPaths.map((manifestPath) =>
      sourceRef(root, "family_manifest", manifestPath),
    ),
    sourceRef(
      root,
      "fixture_registry",
      "tools/frontend_visual_fixture_registry.json",
    ),
    sourceRef(root, "renderer_profile", rendererProfilePath),
    {
      kind: "golden_manifest",
      path: goldenManifestPath,
      sha256: manifestBytes === null
        ? "0".repeat(64)
        : createHash("sha256").update(manifestBytes).digest("hex"),
      project_ids: [],
      snapshot_path_template: null,
      symbols: [],
    },
    sourceRef(root, "playwright_config", playwrightConfigPath, {
      projectIDs,
      snapshotTemplate: snapshotPathTemplate,
    }),
    sourceRef(root, "screenshot_helper", visualSpecPath, {
      symbols: [
        "assertVisualRegression",
        "assertViewportVisualRegression",
      ],
    }),
  ];
  return {
    schema_id: reconciliationSchemaID,
    status: errors.length === 0 ? "pass" : "fail",
    renderer: {
      profile_id: rendererProfile.profile_id,
      container_image: rendererProfile.container_image,
      platform: rendererProfile.platform,
      playwright_version: rendererProfile.playwright_version,
      chromium_revision: rendererProfile.chromium_revision,
      chromium_version: rendererProfile.chromium_version,
      font_manifest_sha256: rendererProfile.font_manifest_sha256,
      locale: rendererProfile.locale,
      device_scale_factor: rendererProfile.device_scale_factor,
      color_scheme: rendererProfile.color_scheme,
      attestation_count: rendererAttestationPaths.length,
    },
    golden_manifest: {
      path: goldenManifestPath,
      sha256: manifestBytes === null
        ? null
        : createHash("sha256").update(manifestBytes).digest("hex"),
      renderer_profile_id: manifest?.renderer_profile_id ?? null,
      status: manifestStatus,
    },
    source_refs: sourceRefs,
    capture_intents: captureIntents,
    goldens,
    counts,
    artifact_refs: uniqueSorted(artifactRefs),
    errors: uniqueSorted(errors),
  };
}
