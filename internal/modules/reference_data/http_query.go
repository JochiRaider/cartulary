package reference_data

import (
	"net/url"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

// Non-list operations have an empty query contract. Parse the complete raw
// input before classifying keys: URL.Query silently discards malformed pairs.
func validateEmptyAdministrativeQuery(raw string) *httpapi.APIError {
	query, err := url.ParseQuery(raw)
	if err != nil {
		return invalidReferencePackRequest("query", "malformed_query")
	}
	for key, values := range query {
		if !utf8.ValidString(key) {
			return invalidReferencePackRequest("query", "malformed_query")
		}
		for _, value := range values {
			if !utf8.ValidString(value) {
				return invalidReferencePackRequest("query", "malformed_query")
			}
		}
	}
	if apiErr := httpapi.ValidateSingletonReadQuery(query); apiErr != nil {
		return apiErr
	}
	if len(query) != 0 {
		return invalidReferencePackRequest("query", "unknown_query_member")
	}
	return nil
}
