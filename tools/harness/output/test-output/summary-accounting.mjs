// Deterministic accounting. Inputs are caller-owned observations; no environment or I/O.
import { classifyExecutionFailure, failureFieldsForJSON } from '../../contract/failure-taxonomy.mjs';
import { normalizeFixtureSummary } from '../../diagnostics/fixture-reporting.mjs';
import { testCoverageBuckets, testCoverageBucketSet, validStepCountingModes } from '../../contract/test-output-context.mjs';
import { addDurationFields, createDurationFields, durationFieldsForJSON, readSummaryDurationFields, teardownStatus } from './timing.mjs';
function createCounts() {
    const counts = {
        tests: 0,
        failed: 0,
        non_test: 0,
        non_test_failed: 0,
        packages: 0,
    };
    for (const coverage of testCoverageBuckets) {
        counts[coverage] = 0;
        counts[`${coverage}_failed`] = 0;
    }
    return counts;
}
function normalizeTestCoverage(value, fallback = "unmapped") {
    const normalized = String(value ?? "").trim();
    if (testCoverageBucketSet.has(normalized)) {
        return normalized;
    }
    return testCoverageBucketSet.has(fallback) ? fallback : "unmapped";
}
function failedTestCount(counts = {}) {
    return testCoverageBuckets.reduce((total, coverage) => total + (counts[`${coverage}_failed`] ?? 0), 0);
}
export function clampDurationMs(value) {
    if (!Number.isFinite(value) || value < 0) {
        return 0;
    }
    return value;
}
function normalizeAccountingMode(value) {
    if (value === "actual" || value === "reused" || value === "derived") {
        return value;
    }
    return "actual";
}
function normalizeStepCountingMode(value) {
    if (validStepCountingModes.has(value)) {
        return value;
    }
    return "counted";
}
export function createAccountingModes() {
    return {
        actual: 0,
        reused: 0,
        derived: 0,
    };
}
export function mergeAccountingModes(target, source) {
    for (const mode of Object.keys(target)) {
        target[mode] += clampDurationMs(source?.[mode] ?? 0);
    }
}
export function resolveAccountingModes(accountingModes, fallbackActualSteps = 0) {
    const modes = createAccountingModes();
    if (!accountingModes) {
        modes.actual = clampDurationMs(fallbackActualSteps);
        return modes;
    }
    for (const mode of Object.keys(modes)) {
        modes[mode] = clampDurationMs(accountingModes[mode] ?? 0);
    }
    return modes;
}
export function computeWindowDurationMs(startTime, endTime) {
    if (!startTime || !endTime) {
        return 0;
    }
    const startMs = Date.parse(startTime);
    const endMs = Date.parse(endTime);
    if (!Number.isFinite(startMs) || !Number.isFinite(endMs) || endMs < startMs) {
        return 0;
    }
    return endMs - startMs;
}
function isDiagnosticWrapperFailure(failure) {
    return (failure?.source === "shell" &&
        failure?.runner === "shell" &&
        failure?.label === "(shell command)" &&
        failure?.failure_class === "harness" &&
        failure?.failure_reason === "unknown_failure" &&
        String(failure?.message ?? "").startsWith("command exited with status "));
}
function demoteDiagnosticWrapperFailures(failures) {
    const hasRicherFailure = failures.some((failure) => !isDiagnosticWrapperFailure(failure) &&
        failure?.failure_class &&
        failure?.failure_reason);
    if (!hasRicherFailure) {
        return 0;
    }
    let writeIndex = 0;
    let demoted = 0;
    for (const failure of failures) {
        if (isDiagnosticWrapperFailure(failure)) {
            demoted += 1;
            continue;
        }
        failures[writeIndex] = failure;
        writeIndex += 1;
    }
    failures.length = writeIndex;
    return demoted;
}
export function normalizeCounts(counts = {}) {
    const normalized = {
        steps: clampDurationMs(counts.steps ?? 0),
        tests: clampDurationMs(counts.tests ?? 0),
        failed: clampDurationMs(counts.failed ?? 0),
        non_test: clampDurationMs(counts.non_test ?? 0),
        non_test_failed: clampDurationMs(counts.non_test_failed ?? 0),
        packages: clampDurationMs(counts.packages ?? 0),
    };
    for (const coverage of testCoverageBuckets) {
        normalized[coverage] = clampDurationMs(counts[coverage] ?? 0);
        normalized[`${coverage}_failed`] = clampDurationMs(counts[`${coverage}_failed`] ?? 0);
    }
    return normalized;
}
function addCounts(target, source) {
    const normalized = normalizeCounts(source);
    for (const key of Object.keys(normalized)) {
        target[key] += normalized[key];
    }
}
function createStepAggregate() {
    return {
        steps: 0,
        ...createCounts(),
        ...createDurationFields(),
    };
}
export function countsForJSON(aggregate) {
    const counts = {
        steps: aggregate.steps,
        tests: aggregate.tests,
        failed: aggregate.failed,
        non_test: aggregate.non_test,
        non_test_failed: aggregate.non_test_failed,
        packages: aggregate.packages,
    };
    for (const coverage of testCoverageBuckets) {
        counts[coverage] = aggregate[coverage];
        counts[`${coverage}_failed`] = aggregate[`${coverage}_failed`];
    }
    return counts;
}
export function findSlowestTarget(targetSummaries) {
    return targetSummaries.reduce((current, summary) => {
        const view = targetSummaryAccountingView(summary);
        const durationMs = clampDurationMs(view.critical_path_wall_duration_ms ?? view.wall_duration_ms ?? view.logical_duration_ms ?? 0);
        if (!current || durationMs > current.critical_path_wall_duration_ms) {
            return {
                target: summary.target,
                critical_path_wall_duration_ms: durationMs,
                basis: "critical_path_wall_duration_ms",
            };
        }
        return current;
    }, null);
}
export function findSlowestLifecycleBucket(targetSummaries) {
    return targetSummaries.reduce((current, summary) => {
        const bucket = targetSummaryAccountingView(summary).slowest_lifecycle_bucket;
        if (!bucket)
            return current;
        const candidate = {
            target: summary.target,
            name: bucket.name,
            duration_ms: clampDurationMs(bucket.duration_ms ?? 0),
        };
        return !current || candidate.duration_ms > current.duration_ms ? candidate : current;
    }, null);
}
function sectionFromFlatSummary(summary, fallbackTarget) {
    const durations = readSummaryDurationFields(summary);
    const counts = normalizeCounts(summary?.counts ?? {});
    let failures = summary?.failures ?? [];
    if (failures.length === 0 && summary?.status && summary.status !== "pass") {
        const failedTests = failedTestCount(counts);
        failures =
            failedTests > 0
                ? [
                    {
                        failure_class: "product",
                        kind: "test",
                        target: summary?.target ?? fallbackTarget,
                        message: "reported test failure",
                    },
                ]
                : counts.non_test_failed > 0
                    ? [
                        {
                            failure_class: classifyExecutionFailure(summary?.target ?? fallbackTarget),
                            kind: "failure",
                            target: summary?.target ?? fallbackTarget,
                            message: "reported non-test failure",
                        },
                    ]
                    : [];
    }
    const failureFields = failureFieldsForJSON(failures, counts);
    return {
        target: summary?.target ?? fallbackTarget,
        status: summary?.status ?? "",
        start_time: summary?.start_time ?? "",
        end_time: summary?.end_time ?? "",
        ...durations,
        accounting_modes: resolveAccountingModes(summary?.accounting_modes, counts.steps),
        counts,
        failure_class: summary?.failure_class ?? failureFields.failure_class,
        failure_reason: summary?.failure_reason ?? failureFields.failure_reason,
        failure_classes: summary?.failure_classes ?? failureFields.failure_classes,
        failure_reasons: summary?.failure_reasons ?? failureFields.failure_reasons,
        failures: summary?.failures ?? failureFields.failures,
        failure_headline: summary?.failure_headline ?? failureFields.failure_headline,
        slowest_lifecycle_bucket: summary?.slowest_lifecycle_bucket ?? null,
        timing_failures: summary?.timing_failures ?? [],
        teardown_status: summary?.teardown_status ??
            teardownStatus(durations.teardown_duration_ms, []),
        teardown_failures: summary?.teardown_failures ?? [],
        fixture: normalizeFixtureSummary(summary?.target ?? fallbackTarget, summary?.fixture),
        artifacts: {
            dir: summary?.artifacts?.dir ??
                "",
            timing_json: summary?.artifacts?.timing_json ?? "",
        },
    };
}
export function targetSummarySection(summary, sectionName, fallbackTarget) {
    if (summary?.[sectionName]) {
        return sectionFromFlatSummary(summary[sectionName], summary.target ?? fallbackTarget);
    }
    return sectionFromFlatSummary(summary, fallbackTarget);
}
export function targetSummaryAccountingView(summary, fallbackTarget = summary?.target ?? "") {
    return targetSummarySection(summary, "totals", fallbackTarget);
}
export function combineSummarySections(target, sections, status = "pass") {
    const aggregate = createStepAggregate();
    const accountingModes = createAccountingModes();
    const timingFailures = [];
    const teardownFailures = [];
    const failures = [];
    let startTime = "";
    let endTime = "";
    let failed = status !== "pass";
    for (const section of sections) {
        const view = targetSummaryAccountingView(section, section?.target ?? target);
        const viewFailures = [...(view.failures ?? [])];
        if (view.status && view.status !== "pass" && viewFailures.length === 0) {
            viewFailures.push({
                failure_class: "harness",
                failure_reason: "child_target_failure",
                kind: "failure",
                source: "target-summary",
                target,
                child_target: view.target ?? section?.target ?? "",
                label: view.target ?? section?.target ?? "",
                message: "child target reported failure status without failure details",
                artifact: view.artifacts?.dir ?? "",
            });
        }
        addCounts(aggregate, view.counts);
        addDurationFields(aggregate, view);
        mergeAccountingModes(accountingModes, view.accounting_modes);
        timingFailures.push(...(view.timing_failures ?? []));
        teardownFailures.push(...(view.teardown_failures ?? []));
        failures.push(...viewFailures);
        if (startTime === "" || (view.start_time && view.start_time < startTime)) {
            startTime = view.start_time ?? "";
        }
        if (endTime === "" || (view.end_time && view.end_time > endTime)) {
            endTime = view.end_time ?? "";
        }
        if (view.status && view.status !== "pass") {
            failed = true;
        }
    }
    const windowWallDurationMs = computeWindowDurationMs(startTime, endTime);
    const wallDurationMs = windowWallDurationMs > 0
        ? windowWallDurationMs
        : aggregate.wall_duration_ms;
    const criticalPathWallDurationMs = wallDurationMs;
    const failureFields = failureFieldsForJSON(failures, countsForJSON(aggregate));
    return {
        target,
        status: failed ? "fail" : "pass",
        start_time: startTime,
        end_time: endTime,
        ...durationFieldsForJSON(aggregate, {
            wall_duration_ms: wallDurationMs,
            critical_path_wall_duration_ms: criticalPathWallDurationMs,
        }),
        accounting_modes: accountingModes,
        counts: countsForJSON(aggregate),
        ...failureFields,
        slowest_lifecycle_bucket: findSlowestLifecycleBucket(sections),
        timing_failures: timingFailures,
        teardown_status: teardownStatus(aggregate.teardown_duration_ms, teardownFailures),
        teardown_failures: teardownFailures,
    };
}
export function summarizeTargetSteps(target, summaries) {
    summaries = [...summaries].sort((a, b) => a.start_time.localeCompare(b.start_time));
    const owners = new Set();
    const inventoryByCoverage = Object.fromEntries(testCoverageBuckets.map((coverage) => [coverage, []]));
    const counts = {
        steps: summaries.length,
        ...createCounts(),
    };
    let startTime = "";
    let endTime = "";
    let actualStartTime = "";
    let actualEndTime = "";
    const durations = createDurationFields();
    const accountingModes = createAccountingModes();
    const failures = [];
    let failed = false;
    for (const summary of summaries) {
        const accountingMode = normalizeAccountingMode(summary.accounting_mode);
        const countingMode = normalizeStepCountingMode(summary.counting_mode ?? "counted");
        const summaryDurations = readSummaryDurationFields(summary, accountingMode);
        if (startTime === "" || summary.start_time < startTime) {
            startTime = summary.start_time;
        }
        if (endTime === "" || summary.end_time > endTime) {
            endTime = summary.end_time;
        }
        if (accountingMode === "actual") {
            if (actualStartTime === "" || summary.start_time < actualStartTime) {
                actualStartTime = summary.start_time;
            }
            if (actualEndTime === "" || summary.end_time > actualEndTime) {
                actualEndTime = summary.end_time;
            }
        }
        addDurationFields(durations, summaryDurations);
        accountingModes[accountingMode] += 1;
        if (countingMode !== "none") {
            counts.tests += summary.counts?.tests ?? 0;
            for (const coverage of testCoverageBuckets) {
                counts[coverage] += summary.counts?.[coverage] ?? 0;
                counts[`${coverage}_failed`] +=
                    summary.counts?.[`${coverage}_failed`] ?? 0;
            }
        }
        counts.failed += summary.counts?.failed ?? 0;
        counts.non_test += summary.counts?.non_test ?? 0;
        counts.non_test_failed += summary.counts?.non_test_failed ?? 0;
        failures.push(...(summary.failures ?? []));
        for (const owner of summary.owners ?? []) {
            owners.add(owner);
        }
        if (countingMode !== "none") {
            for (const item of summary.inventory ?? []) {
                const coverage = normalizeTestCoverage(item.coverage);
                if (inventoryByCoverage[coverage]) {
                    inventoryByCoverage[coverage].push({ ...item, coverage });
                }
            }
        }
        if (summary.status !== "pass") {
            failed = true;
        }
    }
    const demotedFailureCount = demoteDiagnosticWrapperFailures(failures);
    if (demotedFailureCount > 0) {
        counts.failed = Math.max(0, counts.failed - demotedFailureCount);
        counts.non_test = Math.max(0, counts.non_test - demotedFailureCount);
        counts.non_test_failed = Math.max(0, counts.non_test_failed - demotedFailureCount);
    }
    counts.packages = owners.size;
    const actualWindowWallDurationMs = computeWindowDurationMs(actualStartTime, actualEndTime);
    const wallDurationMs = actualWindowWallDurationMs > 0
        ? actualWindowWallDurationMs
        : durations.wall_duration_ms;
    return {
        target,
        summaries,
        counts,
        startTime,
        endTime,
        durations: durationFieldsForJSON(durations, {
            wall_duration_ms: wallDurationMs,
            critical_path_wall_duration_ms: wallDurationMs,
        }),
        executedDurationMs: durations.executed_duration_ms,
        logicalDurationMs: durations.logical_duration_ms,
        wallDurationMs,
        criticalPathWallDurationMs: wallDurationMs,
        reusedDurationMs: durations.reused_duration_ms,
        derivedDurationMs: durations.derived_duration_ms,
        teardownDurationMs: durations.teardown_duration_ms,
        accountingModes,
        failures,
        failed,
        inventoryByCoverage,
    };
}
function addSummaryToAggregate(aggregate, accountingModes, summary) {
    const view = targetSummaryAccountingView(summary);
    addCounts(aggregate, view.counts);
    addDurationFields(aggregate, view);
    mergeAccountingModes(accountingModes, resolveAccountingModes(view.accounting_modes, view.counts?.steps ?? 0));
}
export function summarizeTargetSummaries(summaries, missingTargetSummaries, requestedStatus = "pass", failureContext = {}) {
    const aggregate = createStepAggregate();
    const accountingModes = createAccountingModes();
    let failed = requestedStatus === "fail" || missingTargetSummaries.length > 0;
    let startTime = "";
    let endTime = "";
    const timingFailures = [];
    const teardownFailures = [];
    const failures = [];
    for (const summary of summaries) {
        const view = targetSummaryAccountingView(summary);
        addSummaryToAggregate(aggregate, accountingModes, summary);
        timingFailures.push(...(view.timing_failures ?? []));
        teardownFailures.push(...(view.teardown_failures ?? []));
        failures.push(...(view.failures ?? []));
        if (startTime === "" || (view.start_time && view.start_time < startTime)) {
            startTime = view.start_time ?? "";
        }
        if (endTime === "" || (view.end_time && view.end_time > endTime)) {
            endTime = view.end_time ?? "";
        }
        if (summary.status !== "pass") {
            failed = true;
        }
    }
    const abortedAfter = failureContext.abortedAfter ?? "";
    const recordsMissingTargetFailures = missingTargetSummaries.length > 0 &&
        !(requestedStatus === "fail" && abortedAfter && abortedAfter !== "-");
    if (requestedStatus === "fail" &&
        aggregate.failed === 0 &&
        recordsMissingTargetFailures) {
        aggregate.steps += 1;
        aggregate.failed += 1;
        aggregate.non_test += 1;
        aggregate.non_test_failed += 1;
    }
    if (requestedStatus === "fail" &&
        failures.length === 0 &&
        !recordsMissingTargetFailures) {
        if (aggregate.failed === 0) {
            aggregate.steps += 1;
            aggregate.failed += 1;
            aggregate.non_test += 1;
            aggregate.non_test_failed += 1;
        }
        const failureTarget = failureContext.abortedAfter && failureContext.abortedAfter !== "-"
            ? failureContext.abortedAfter
            : (failureContext.target ?? "");
        failures.push({
            failure_class: classifyExecutionFailure(failureTarget, failureContext.label ?? "", failureContext.command ?? ""),
            kind: "failure",
            source: "run-summary",
            target: failureContext.target ?? "",
            label: failureTarget,
            message: failureTarget
                ? `${failureTarget} failed before a test failure was attributed`
                : "run failed before a test failure was attributed",
        });
    }
    for (const missingTarget of missingTargetSummaries) {
        if (requestedStatus === "fail" && abortedAfter && abortedAfter !== "-") {
            continue;
        }
        const abortedClass = abortedAfter && abortedAfter !== "-"
            ? classifyExecutionFailure(abortedAfter, failureContext.label ?? "")
            : "";
        const missingClass = abortedClass ||
            (classifyExecutionFailure(missingTarget) === "timing"
                ? "timing"
                : "artifact");
        failures.push({
            failure_class: missingClass,
            kind: missingClass === "timing" ? "timing" : "artifact",
            source: "run-summary",
            target: failureContext.target ?? "",
            label: missingTarget,
            message: `missing target summary: ${missingTarget}`,
            artifact: failureContext.missingArtifacts?.[missingTarget] ?? "",
        });
    }
    const windowWallDurationMs = computeWindowDurationMs(startTime, endTime);
    const wallDurationMs = windowWallDurationMs > 0
        ? windowWallDurationMs
        : aggregate.wall_duration_ms;
    const criticalPathWallDurationMs = wallDurationMs;
    const failureFields = failureFieldsForJSON(failures, countsForJSON(aggregate));
    return {
        aggregate,
        accountingModes,
        failed,
        startTime,
        endTime,
        wallDurationMs,
        criticalPathWallDurationMs,
        timingFailures,
        teardownFailures,
        ...failureFields,
    };
}
