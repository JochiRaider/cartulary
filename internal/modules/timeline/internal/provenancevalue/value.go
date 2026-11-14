package provenancevalue

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const MaxValueBytes = 65536
const MaxEntriesPerRecord = 4096

var metadataKeys = map[string]bool{
	"source_kind": true, "paste_client_txn_id": true, "import_session_id": true, "import_unit_id": true,
	"mapping_fingerprint": true, "source_file_kind": true, "source_content_sha256": true,
	"parser_profile_id": true, "parser_version": true, "locator_kind": true, "locator": true, "source_rect_a1": true,
}

type Values struct {
	SourceKind   string
	Metadata     map[string]string
	Header       *string
	RawValue     string
	CellKind     *string
	MetadataJSON []byte
	HeaderJSON   []byte
	Digest       [sha256.Size]byte
}

// ParseValues admits the closed value shape and both logical/storage byte bounds.
// It does not repair strings or determine which source-kind tuple is complete.
func ParseValues(kind any, metadata any, header any, raw any, cellKind any) (Values, bool) {
	var value Values
	var ok bool
	value.SourceKind, ok = kind.(string)
	if !ok || !kindText(value.SourceKind) {
		return Values{}, false
	}
	value.RawValue, ok = raw.(string)
	if !ok || !retainedText(value.RawValue) || len(value.RawValue) > MaxValueBytes {
		return Values{}, false
	}
	if header != nil {
		text, ok := header.(string)
		if !ok || !retainedText(text) {
			return Values{}, false
		}
		value.Header = &text
	}
	if cellKind != nil {
		text, ok := cellKind.(string)
		if !ok || !kindText(text) {
			return Values{}, false
		}
		value.CellKind = &text
	}
	object, ok := metadata.(map[string]any)
	if !ok || len(object) == 0 {
		return Values{}, false
	}
	value.Metadata = make(map[string]string, len(object))
	for key, raw := range object {
		text, ok := raw.(string)
		if !metadataKeys[key] || !ok || !retainedText(text) {
			return Values{}, false
		}
		value.Metadata[key] = text
	}
	if _, present := value.Metadata["source_kind"]; !present {
		return Values{}, false
	}
	var err error
	value.MetadataJSON, err = json.Marshal(value.Metadata)
	if err != nil || len(value.MetadataJSON) > MaxValueBytes {
		return Values{}, false
	}
	value.HeaderJSON, err = json.Marshal(value.Header)
	if err != nil || len(value.HeaderJSON) > MaxValueBytes {
		return Values{}, false
	}
	// Metadata keys are closed ASCII strings and all values are text. JSONB
	// differs from compact JCS only by its colon/comma spaces; header is scalar.
	storedMetadata, err := canonicaljson.Marshal(value.Metadata)
	if err != nil || len(storedMetadata)+2*len(value.Metadata)-1 > MaxValueBytes {
		return Values{}, false
	}
	storedHeader, err := canonicaljson.Marshal(value.Header)
	if err != nil || len(storedHeader) > MaxValueBytes {
		return Values{}, false
	}
	value.Digest = sha256.Sum256(value.MetadataJSON)
	return value, true
}

func (v Values) Coherent() bool {
	if v.Metadata["source_kind"] != v.SourceKind {
		return false
	}
	for key, value := range v.Metadata {
		if value == "" {
			return false
		}
		switch key {
		case "import_session_id", "import_unit_id":
			if !CanonicalUUID(value) {
				return false
			}
		case "mapping_fingerprint", "source_content_sha256":
			if !LowerDigest(value) {
				return false
			}
		}
	}
	required := []string{}
	switch v.SourceKind {
	case "clipboard_paste":
		required = []string{"paste_client_txn_id", "mapping_fingerprint"}
	case "file_import":
		required = []string{"import_session_id", "import_unit_id", "mapping_fingerprint", "source_file_kind", "source_content_sha256", "parser_profile_id", "parser_version", "locator_kind", "locator", "source_rect_a1"}
	}
	for _, key := range required {
		if v.Metadata[key] == "" {
			return false
		}
	}
	return true
}

func CanonicalUUID(text string) bool {
	value, err := uuid.Parse(text)
	return err == nil && value != uuid.Nil && value.String() == text
}
func LowerDigest(text string) bool {
	if len(text) != 64 || strings.ToLower(text) != text {
		return false
	}
	_, err := hex.DecodeString(text)
	return err == nil
}
func retainedText(text string) bool { return utf8.ValidString(text) && !strings.ContainsRune(text, 0) }
func kindText(text string) bool {
	if !retainedText(text) || len(text) == 0 || len(text) > 64 {
		return false
	}
	for _, r := range text {
		if r <= 31 || (r >= 127 && r <= 159) {
			return false
		}
	}
	return true
}
