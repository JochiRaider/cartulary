package telemetry

import (
	"regexp"
	"strings"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
)

func SafeAttributes(attrs ...attribute.KeyValue) []attribute.KeyValue {
	if len(attrs) == 0 {
		return nil
	}
	safe := make([]attribute.KeyValue, 0, len(attrs))
	for _, attr := range attrs {
		if safeAttribute(attr) {
			safe = append(safe, attr)
		}
	}
	return safe
}

func safeAttribute(attr attribute.KeyValue) bool {
	key := string(attr.Key)
	if key == "" || attr.Value.Type() == attribute.EMPTY {
		return false
	}
	switch key {
	case "service.name", "service.namespace", "deployment.environment.name":
		return safeString(attr, safeAttributeToken)
	case "service.version":
		return safeString(attr, safeServiceVersion)
	case "service.instance.id":
		return safeString(attr, safeUUIDV4)
	case "db.system.name":
		return safeStringIn(attr, "postgresql")
	case "http.request.method":
		return safeString(attr, safeHTTPMethodToken)
	case "http.route":
		return safeString(attr, safeRouteTemplate)
	case "http.response.status_code":
		return attr.Value.Type() == attribute.INT64 && attr.Value.AsInt64() >= 100 && attr.Value.AsInt64() <= 599
	case "cartulary.deployment.profile":
		return safeStringIn(attr, "disconnected", "on_prem", "cloud")
	case "cartulary.profile.claims":
		return safeString(attr, safeProfileClaims)
	case "cartulary.module":
		return safeStringIn(attr, "httpapi", "workbook", "collaboration", "jobs", "postgres", "objectstore", "telemetry")
	case "cartulary.route_family":
		return safeStringIn(attr, "web.root", "health", "readiness", "web.asset", "auth", "incidents", "records", "jobs", "view_schemas", "api", "websocket", "unmatched")
	case "cartulary.view_schema_id", "cartulary.record_type", "cartulary.error_code":
		return safeString(attr, safeAttributeTokenNoForbiddenID)
	case "cartulary.operation":
		return safeString(attr, safeOperation)
	case "cartulary.result":
		return safeStringIn(attr, "success", "rejected", "conflict", "canceled", "failed", "timeout", "dropped")
	case "cartulary.error_class":
		return safeStringIn(attr, safeErrorClasses...)
	case "cartulary.websocket.event_type":
		return safeStringIn(attr, "record_changed", "extension_resource_changed", "job_progress", "presence_delta", "presence_snapshot", "hello_ack", "resume_ack", "ping", "session_revoked", "error", "other")
	case "cartulary.job_kind":
		return safeString(attr, safeAttributeTokenNoForbiddenID)
	case "cartulary.job_terminal_status":
		return safeStringIn(attr, "succeeded", "failed", "canceled", "expired")
	case "cartulary.signal_kind":
		return safeStringIn(attr, "traces", "metrics", "logs")
	case "cartulary.telemetry.exporter_kind":
		return safeStringIn(attr, "none", "otlp_http", "otlp_grpc")
	case "cartulary.drop_reason":
		return safeStringIn(attr, "queue_full", "redaction_rejected", "exporter_permanent_discard", "shutdown_timeout", "recursion_guard", "metric_overflow")
	case "cartulary.incident.hash64":
		return safeString(attr, safeIncidentHash64)
	default:
		return false
	}
}

func safeString(attr attribute.KeyValue, validator func(string) bool) bool {
	if attr.Value.Type() != attribute.STRING {
		return false
	}
	value := attr.Value.AsString()
	return value != "" && !containsForbiddenValueShape(value) && validator(value)
}

func safeStringIn(attr attribute.KeyValue, allowed ...string) bool {
	if attr.Value.Type() != attribute.STRING {
		return false
	}
	value := attr.Value.AsString()
	for _, candidate := range allowed {
		if value == candidate {
			return true
		}
	}
	return false
}

func safeAttributeToken(value string) bool {
	return telemetrySafeTokenPattern.MatchString(value)
}

func safeAttributeTokenNoForbiddenID(value string) bool {
	return safeAttributeToken(value) && !uuidLikePattern.MatchString(value)
}

func safeServiceVersion(value string) bool {
	return value == VersionUnknown || semverLikePattern.MatchString(value)
}

func safeUUIDV4(value string) bool {
	parsed, err := uuid.Parse(value)
	return err == nil && parsed.Version() == 4 && parsed != uuid.Nil && value == strings.ToLower(parsed.String())
}

func safeHTTPMethodToken(value string) bool {
	switch value {
	case "GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "UNKNOWN", "OTHER":
		return true
	default:
		return false
	}
}

func safeRouteTemplate(value string) bool {
	if value == "/" {
		return true
	}
	if !strings.HasPrefix(value, "/") || strings.ContainsAny(value, "?#") || uuidLikePattern.MatchString(value) {
		return false
	}
	segments := strings.Split(strings.Trim(value, "/"), "/")
	for _, segment := range segments {
		if segment == "" || segment == "." || segment == ".." {
			return false
		}
		if telemetrySafeRouteTemplateSegmentPattern.MatchString(segment) {
			continue
		}
		if telemetrySafeRouteTokenPattern.MatchString(segment) {
			continue
		}
		return false
	}
	return true
}

func safeProfileClaims(value string) bool {
	seenBase := false
	tokens := strings.Split(value, ",")
	for index, token := range tokens {
		if !resolvedClaimProfileIDPattern.MatchString(token) {
			return false
		}
		if index > 0 && tokens[index-1] >= token {
			return false
		}
		if token == "base" {
			if seenBase {
				return false
			}
			seenBase = true
		}
	}
	return seenBase
}

func safeOperation(value string) bool {
	switch value {
	case "connect", "query", "create", "patch", "enqueue", "run", "exec", "query_row", "begin_tx",
		"create_upload_target", "complete_upload_target", "put_object", "get_object", "get_object_range",
		"head_object", "list_prefix", "delete_object", "ensure_bucket_for_dev_test", "rebuild", "unknown":
		return true
	default:
		return false
	}
}

func safeIncidentHash64(value string) bool {
	return incidentHashPattern.MatchString(value)
}

func containsForbiddenValueShape(value string) bool {
	lower := strings.ToLower(value)
	switch {
	case strings.Contains(value, "://"):
		return true
	case strings.ContainsAny(value, "\n\r\t"):
		return true
	case strings.Contains(value, "?") || strings.Contains(value, "#"):
		return true
	case strings.Contains(lower, "bearer ") || strings.Contains(lower, "password") || strings.Contains(lower, "authorization"):
		return true
	case strings.HasPrefix(value, "/tmp/") || strings.HasPrefix(value, "/var/") || strings.HasPrefix(value, "/home/"):
		return true
	default:
		return false
	}
}

var (
	telemetrySafeTokenPattern                = regexp.MustCompile(`^[A-Za-z0-9._-]{1,128}$`)
	telemetrySafeRouteTemplateSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9._~-]{1,64}$`)
	telemetrySafeRouteTokenPattern           = regexp.MustCompile(`^\{[A-Za-z0-9_]{1,64}\}$`)
	uuidLikePattern                          = regexp.MustCompile(`(?i)[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`)
	semverLikePattern                        = regexp.MustCompile(`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?(\+[0-9A-Za-z-]+(\.[0-9A-Za-z-]+)*)?$`)
	incidentHashPattern                      = regexp.MustCompile(`^[0-9a-f]{16}$`)
)

var safeErrorClasses = []string{
	"request_invalid",
	"authentication",
	"authorization",
	"capability_unavailable",
	"concurrency_conflict",
	"lifecycle_conflict",
	"not_found",
	"expired_or_consumed",
	"policy_rejected",
	"dependency_unavailable",
	"invariant_violation",
	"timeout",
	"serialization_conflict",
	"constraint_violation",
	"exporter_transient",
	"exporter_permanent",
	"redaction_rejected",
	"queue_full",
	"shutdown_timeout",
	"recursion_guard",
	"internal_error",
}
