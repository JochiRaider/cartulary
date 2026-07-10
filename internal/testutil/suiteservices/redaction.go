package suiteservices

import (
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/harnessredact"
)

func SanitizeDiagnosticText(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	return harnessredact.String(trimmed)
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
		sanitized[key] = harnessredact.Value(sanitizeDetailValue(value), key)
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
