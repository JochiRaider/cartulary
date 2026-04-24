package suiteservices

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	dsnPattern = regexp.MustCompile(`(?i)\b(?:postgres|postgresql|mysql|mariadb|redis|mongodb|amqp|nats)://[^\s"'<>]+`)
	kvPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\b(password|passwd|secret|secret_key|secret_access_key|access_key|access_key_id)\b(\s*[:=]\s*)([^\s,;]+)`),
		regexp.MustCompile(`(?i)\b(password|passwd|secret|secret_key|secret_access_key|access_key|access_key_id)\b("?\s*:\s*")([^"]+)(")`),
	}
)

func SanitizeDiagnosticText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}

	sanitized := dsnPattern.ReplaceAllString(trimmed, "[redacted-dsn]")
	for _, pattern := range kvPatterns {
		sanitized = pattern.ReplaceAllStringFunc(sanitized, func(match string) string {
			groups := pattern.FindStringSubmatch(match)
			switch len(groups) {
			case 4:
				return fmt.Sprintf("%s%s[redacted]", groups[1], groups[2])
			case 5:
				return fmt.Sprintf("%s%s[redacted]%s", groups[1], groups[2], groups[4])
			default:
				return match
			}
		})
	}

	return sanitized
}

func sanitizeEvent(event Event) Event {
	event.Details = sanitizeDetailMap(event.Details)
	return event
}

func sanitizeDetailMap(details map[string]any) map[string]any {
	if len(details) == 0 {
		return details
	}

	sanitized := make(map[string]any, len(details))
	for key, value := range details {
		sanitized[key] = sanitizeDetailValue(value)
	}
	return sanitized
}

func sanitizeDetailValue(value any) any {
	switch typed := value.(type) {
	case string:
		return SanitizeDiagnosticText(typed)
	case []string:
		items := make([]string, len(typed))
		for index, item := range typed {
			items[index] = SanitizeDiagnosticText(item)
		}
		return items
	case []any:
		items := make([]any, len(typed))
		for index, item := range typed {
			items[index] = sanitizeDetailValue(item)
		}
		return items
	case map[string]any:
		return sanitizeDetailMap(typed)
	default:
		return value
	}
}
