// Producer payloads used at the real output CLI boundary.
export const validFindings = {
    "schema_id": "cartulary.govulncheck_findings.v1",
    "tool": "govulncheck",
    "status": "fail",
    "config": null,
    "counts": {
        "raw_event_count": 2,
        "osv_count": 1,
        "finding_count": 1,
        "blocking_count": 1,
        "reachability": {
            "module": 0,
            "package": 0,
            "symbol": 1
        }
    },
    "vulnerability_ids": [
        "GO-2099-0001"
    ],
    "blocking_vulnerability_ids": [
        "GO-2099-0001"
    ],
    "findings": [
        {
            "id": "GO-2099-0001",
            "aliases": [],
            "summary": "synthetic reachable vulnerability",
            "fixed_version": "",
            "fixed_versions": [],
            "affected_packages": [],
            "reachability": "symbol",
            "blocking": true,
            "modules": [],
            "packages": [],
            "symbols": [],
            "trace": []
        }
    ]
};
export const malformedFindings = {
    "schema_id": "cartulary.govulncheck_findings.v1",
    "tool": "govulncheck",
    "status": "fail",
    "counts": {
        "finding_count": 1,
        "blocking_count": 1
    },
    "blocking_vulnerability_ids": [
        "GO-2099-0001"
    ],
    "findings": []
};
export const startupFailure = {
    "schema_id": "cartulary.browser_startup_diagnostics.v3",
    "suite_id": "suite-test",
    "browser_session_id": "session-test",
    "runtime_profile_id": "default",
    "generated_at": "2026-01-01T00:00:00Z",
    "status": "failed",
    "startup_phase": "failed",
    "message": "Port 39000 is already in use",
    "events_ref": "_shared/test-services/suite-test/browser-sessions/session-test/startup-events.jsonl",
    "events_sha256": "sha256:1111111111111111111111111111111111111111111111111111111111111111",
    "frontend_mode": "preview",
    "frontend_command_kind": "vite-preview",
    "api_origin": "http://127.0.0.1:39080",
    "public_origin": "http://127.0.0.1:39000",
    "failure_class": "infra",
    "failure_reason": "resource_conflict"
};
