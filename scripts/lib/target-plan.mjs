import {
  collectEntries,
  collectSupportGoEntries,
  effectiveGoEntryPostgresFixturePolicy,
  effectiveSupportGoEntryPostgresFixturePolicy,
  goEntrySymbols,
  loadManifest,
  packageMatchesPattern,
  phaseManifestNames,
  supportGoEntrySymbols,
} from "./phase-manifest.mjs";

const targetDescriptors = [
  {
    name: "backend-unit",
    serviceBacked: false,
    checkHeavySafe: true,
    checkServiceBackedSafe: false,
    checkIsolatedSafe: false,
    canonicalAuthoritative: true,
    manifestFamilies: [
      {
        phase: "phase0",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_unit",
        packages: ["./internal/platform/..."],
        sharedReport: "backend-unit-core",
        allowEmpty: false,
      },
      {
        phase: "phase1",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_unit",
        packages: ["./internal/platform/..."],
        sharedReport: "backend-unit-core",
        allowEmpty: true,
      },
      {
        phase: "phase0",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_unit",
        packages: ["./internal/app"],
        sharedReport: "backend-unit-core",
        allowEmpty: false,
      },
      {
        phase: "phase1",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_unit",
        packages: ["./internal/modules/auth"],
        sharedReport: "backend-unit-auth",
        allowEmpty: false,
      },
      {
        phase: "phase4",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_unit",
        packages: [
          "./internal/app",
          "./internal/modules/incidents",
          "./internal/modules/entities",
          "./internal/modules/timeline",
        ],
        sharedReport: "backend-unit-core",
        allowEmpty: false,
      },
      {
        phase: "phase2",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_unit",
        packages: ["./internal/modules/incidents"],
        sharedReport: "backend-unit-core",
        allowEmpty: false,
      },
      {
        phase: "phase3",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_unit",
        packages: ["./internal/modules/timeline"],
        sharedReport: "backend-unit-core",
        allowEmpty: false,
      },
    ],
    supportFamilies: [
      {
        target: "backend_unit",
        packages: [
          "./internal/platform/...",
          "./internal/app",
          "./internal/modules/incidents",
          "./internal/modules/entities",
          "./internal/modules/timeline",
        ],
        sharedReport: "backend-unit-core",
      },
      {
        target: "backend_unit",
        packages: ["./internal/modules/auth"],
        sharedReport: "backend-unit-auth",
      },
    ],
    rawFamilies: [
      {
        id: "RAW-backend-unit-configtest",
        label: "backend-unit configtest",
        section: "unit",
        selector: "^Test",
        packages: ["./internal/testutil/configtest"],
        sharedReport: "backend-unit-configtest",
      },
    ],
  },
  {
    name: "backend-store",
    serviceBacked: true,
    checkHeavySafe: false,
    checkServiceBackedSafe: true,
    checkIsolatedSafe: false,
    canonicalAuthoritative: true,
    manifestFamilies: [
      {
        phase: "phase4",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_store",
        packages: ["./internal/modules/entities", "./internal/modules/timeline"],
        sharedReport: "backend-store-shared",
        allowEmpty: false,
      },
      {
        phase: "phase1",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_store",
        packages: ["./internal/modules/auth"],
        sharedReport: "backend-store-shared",
        allowEmpty: false,
      },
      {
        phase: "phase2",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_store",
        packages: ["./internal/modules/incidents"],
        sharedReport: "backend-store-shared",
        allowEmpty: false,
      },
      {
        phase: "phase3",
        section: "unit",
        coverage: "authoritative",
        executionDependency: "backend_store",
        packages: ["./internal/modules/timeline"],
        sharedReport: "backend-store-shared",
        allowEmpty: false,
      },
    ],
  },
  {
    name: "backend-integration",
    serviceBacked: true,
    checkHeavySafe: false,
    checkServiceBackedSafe: true,
    checkIsolatedSafe: false,
    canonicalAuthoritative: true,
    manifestFamilies: [
      {
        phase: "phase0",
        section: "integration",
        coverage: "authoritative",
        executionDependency: "backend_integration",
        packages: ["./internal/platform/..."],
        sharedReport: "backend-integration-phase0-platform",
        allowEmpty: false,
      },
      {
        phase: "phase0",
        section: "integration",
        coverage: "authoritative",
        executionDependency: "backend_integration",
        packages: ["./internal/app"],
        sharedReport: "backend-integration-phase0-app",
        allowEmpty: false,
      },
      {
        phase: "phase1",
        section: "integration",
        coverage: "authoritative",
        executionDependency: "backend_integration",
        packages: ["./internal/modules/auth"],
        sharedReport: "backend-integration-auth",
        allowEmpty: false,
      },
      {
        phase: "phase4",
        section: "integration",
        coverage: "authoritative",
        executionDependency: "backend_integration",
        packages: ["./internal/modules/entities"],
        sharedReport: "backend-integration-phase4-entities",
        allowEmpty: false,
      },
      {
        phase: "phase4",
        section: "integration",
        coverage: "authoritative",
        executionDependency: "backend_integration",
        packages: ["./internal/modules/timeline"],
        sharedReport: "backend-integration-phase4-timeline",
        allowEmpty: false,
      },
      {
        phase: "phase2",
        section: "integration",
        coverage: "authoritative",
        executionDependency: "backend_integration",
        packages: ["./internal/modules/incidents"],
        sharedReport: "backend-integration-phase2-incidents",
        allowEmpty: false,
      },
      {
        phase: "phase3",
        section: "integration",
        coverage: "authoritative",
        executionDependency: "backend_integration",
        packages: ["./internal/modules/timeline"],
        sharedReport: "backend-integration-phase3-timeline",
        allowEmpty: false,
      },
    ],
    rawFamilies: [
      {
        id: "RAW-backend-integration-testutil",
        label: "backend-integration testutil",
        section: "integration",
        selector: "^Test",
        packages: [
          "./internal/testutil/httptestx",
          "./internal/testutil/pgtest",
          "./internal/testutil/s3test",
          "./internal/testutil/testcontainersx",
          "./internal/testutil/wstest",
        ],
        sharedReport: "backend-integration-testutil",
        fixturePolicy: {
          postgres: "package_reset",
        },
      },
    ],
  },
  {
    name: "backend-integration-support",
    serviceBacked: true,
    checkHeavySafe: false,
    checkServiceBackedSafe: true,
    checkIsolatedSafe: false,
    canonicalAuthoritative: false,
    supportFamilies: [
      {
        target: "backend_integration_support",
        packages: ["./internal/platform/..."],
        sharedReport: "backend-integration-phase0-platform",
      },
      {
        target: "backend_integration_support",
        packages: ["./internal/modules/incidents"],
        sharedReport: "backend-integration-phase2-incidents",
      },
      {
        target: "backend_integration_support",
        packages: ["./internal/modules/timeline"],
        sharedReport: "backend-integration-phase3-timeline",
      },
      {
        target: "backend_integration_support",
        packages: ["./internal/modules/entities"],
        sharedReport: "backend-integration-phase4-entities",
      },
      {
        target: "backend_integration_support",
        packages: ["./internal/modules/auth"],
        sharedReport: "backend-integration-auth",
      },
    ],
  },
  {
    name: "backend-process",
    serviceBacked: true,
    checkHeavySafe: false,
    checkServiceBackedSafe: true,
    checkIsolatedSafe: false,
    canonicalAuthoritative: true,
    manifestFamilies: [
      {
        phase: "phase0",
        section: "e2e",
        coverage: "authoritative",
        executionDependency: "backend_process",
        packages: ["./cmd/server"],
        sharedReport: "backend-process-shared",
        allowEmpty: false,
      },
    ],
    rawFamilies: [
      {
        id: "RAW-backend-process-phase1-smoke",
        label: "backend-process phase1 smoke",
        section: "e2e",
        selector: "^(TestPhase1_.*_ProcessSmoke)$",
        packages: ["./cmd/server"],
        sharedReport: "backend-process-shared",
      },
    ],
  },
  {
    name: "phase0-process-e2e",
    serviceBacked: true,
    checkHeavySafe: false,
    checkServiceBackedSafe: false,
    checkIsolatedSafe: false,
    canonicalAuthoritative: false,
    manifestFamilies: [
      {
        phase: "phase0",
        section: "e2e",
        coverage: "authoritative",
        executionDependency: "backend_process",
        packages: ["./cmd/server"],
        sharedReport: "backend-process-shared",
        allowEmpty: false,
      },
    ],
  },
  {
    name: "phase1-process-smoke",
    serviceBacked: true,
    checkHeavySafe: false,
    checkServiceBackedSafe: false,
    checkIsolatedSafe: false,
    canonicalAuthoritative: false,
    rawFamilies: [
      {
        id: "RAW-phase1-process-smoke",
        label: "phase1-process-smoke",
        section: "e2e",
        selector: "^(TestPhase1_.*_ProcessSmoke)$",
        packages: ["./cmd/server"],
        sharedReport: "backend-process-shared",
      },
    ],
  },
  {
    name: "phase2-process-smoke",
    serviceBacked: true,
    checkHeavySafe: false,
    checkServiceBackedSafe: false,
    checkIsolatedSafe: false,
    canonicalAuthoritative: false,
    rawFamilies: [
      {
        id: "RAW-phase2-process-smoke",
        label: "phase2-process-smoke",
        section: "e2e",
        selector: "^(TestPhase2_ProcessSmoke_)",
        packages: ["./cmd/server"],
        sharedReport: "backend-process-shared",
      },
    ],
  },
];

function compareStrings(left, right) {
  return left.localeCompare(right);
}

function compareRows(left, right) {
  return (
    compareStrings(left.target, right.target) ||
    compareStrings(left.manifest_phase, right.manifest_phase) ||
    compareStrings(left.section, right.section) ||
    compareStrings(left.id, right.id)
  );
}

function rowBase(descriptor) {
  return {
    target: descriptor.name,
    service_backed: descriptor.serviceBacked,
    runner_family: "go_test",
    check_heavy_safe: descriptor.checkHeavySafe,
    check_service_backed_safe: descriptor.checkServiceBackedSafe,
    check_isolated_safe: descriptor.checkIsolatedSafe,
    canonical_authoritative: descriptor.canonicalAuthoritative,
  };
}

function selectGoEntries(root, family) {
  const { manifest } = loadManifest(root, family.phase);
  return collectEntries(manifest).filter(
    (entry) =>
      entry.section === family.section &&
      entry.runner === "go_test" &&
      entry.coverage === family.coverage &&
      entry.execution_dependency === family.executionDependency &&
      family.packages.some((pattern) => packageMatchesPattern(entry.package, pattern)),
  );
}

function selectSupportEntries(root, phase, family) {
  const { manifest } = loadManifest(root, phase);
  return collectSupportGoEntries(manifest).filter(
    (entry) =>
      entry.target === family.target &&
      family.packages.some((pattern) => packageMatchesPattern(entry.package, pattern)),
  );
}

function supportID(phase, target, file, symbol) {
  const normalized = `${phase}-${target}-${file}-${symbol}`.replace(/[^A-Za-z0-9]+/g, "-");
  return `SUPPORT-${normalized.replace(/^-|-$/g, "")}`;
}

function manifestRowsForFamily(root, descriptor, family) {
  const entries = selectGoEntries(root, family);
  if (entries.length === 0) {
    if (family.allowEmpty) {
      return [];
    }
    throw new Error(
      `target ${descriptor.name} family ${family.phase} ${family.section} ${family.executionDependency} selected no manifest rows`,
    );
  }
  return entries.map((entry) => ({
    ...rowBase(descriptor),
    id: entry.id,
    manifest_phase: family.phase,
    section: entry.section,
    coverage: entry.coverage,
    execution_dependency: entry.execution_dependency,
    packages: [...family.packages],
    support_only: false,
    support_selector: null,
    raw_selector: null,
    shared_report: family.sharedReport,
    file: entry.file,
    package: entry.package,
    symbols: goEntrySymbols(entry),
    evidence_layer: entry.evidence_layer,
    fixture_policy: {
      postgres: effectiveGoEntryPostgresFixturePolicy(entry),
    },
  }));
}

function supportRowsForFamily(root, descriptor, family) {
  const rows = [];
  for (const phase of phaseManifestNames(root)) {
    for (const entry of selectSupportEntries(root, phase, family)) {
      for (const symbol of supportGoEntrySymbols(entry)) {
        rows.push({
          ...rowBase(descriptor),
          id: supportID(phase, entry.target, entry.file, symbol),
          manifest_phase: phase,
          section: entry.section,
          coverage: "support",
          execution_dependency: entry.target,
          packages: [...family.packages],
          support_only: true,
          support_selector: entry.selection_pattern,
          raw_selector: null,
          shared_report: family.sharedReport,
          file: entry.file,
          package: entry.package,
          symbols: [symbol],
          evidence_layer: "support",
          fixture_policy: {
            postgres: effectiveSupportGoEntryPostgresFixturePolicy(entry),
          },
        });
      }
    }
  }
  return rows;
}

function rawRowsForFamily(descriptor, family) {
  return [
    {
      ...rowBase(descriptor),
      id: family.id,
      manifest_phase: "",
      section: family.section,
      coverage: "raw",
      execution_dependency: "",
      packages: [...family.packages],
      support_only: false,
      support_selector: null,
      raw_selector: family.selector,
      shared_report: family.sharedReport,
      file: "",
      package: "",
      symbols: [],
      evidence_layer: "raw",
      label: family.label,
      fixture_policy: family.fixturePolicy ?? {},
    },
  ];
}

export function collectTargetPlanRows(root = process.cwd()) {
  const rows = [];
  for (const descriptor of targetDescriptors) {
    for (const family of descriptor.manifestFamilies ?? []) {
      rows.push(...manifestRowsForFamily(root, descriptor, family));
    }
    for (const family of descriptor.supportFamilies ?? []) {
      rows.push(...supportRowsForFamily(root, descriptor, family));
    }
    for (const family of descriptor.rawFamilies ?? []) {
      rows.push(...rawRowsForFamily(descriptor, family));
    }
  }
  return rows.sort(compareRows);
}

export function collectTargetNames() {
  return targetDescriptors.map((descriptor) => descriptor.name);
}

export function findTargetDescriptor(target) {
  return targetDescriptors.find((descriptor) => descriptor.name === target) ?? null;
}

export function knownManifestPhases(root = process.cwd()) {
  return phaseManifestNames(root);
}
