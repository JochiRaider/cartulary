package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	UploadEnvelopeReasonUnsupportedUploadEnvelope = "unsupported_upload_envelope"
	UploadEnvelopeReasonMissingRequiredPart       = "missing_required_part"
	UploadEnvelopeReasonDuplicatePart             = "duplicate_part"
	UploadEnvelopeReasonUnexpectedPart            = "unexpected_part"
	UploadEnvelopeReasonInvalidPartContentType    = "invalid_part_content_type"
	UploadEnvelopeReasonInvalidMetadataEncoding   = "invalid_metadata_encoding"
	UploadEnvelopeReasonMalformedMetadataJSON     = "malformed_metadata_json"
	UploadEnvelopeReasonRequestNotObject          = "request_not_object"
)

var (
	metadataUploadEnvelopeContentTypes = []string{"application/json", "application/json; charset=utf-8"}
	sharedUploadEnvelopeReasonCodes    = []string{
		UploadEnvelopeReasonUnsupportedUploadEnvelope,
		UploadEnvelopeReasonMissingRequiredPart,
		UploadEnvelopeReasonDuplicatePart,
		UploadEnvelopeReasonUnexpectedPart,
		UploadEnvelopeReasonInvalidPartContentType,
		UploadEnvelopeReasonInvalidMetadataEncoding,
		UploadEnvelopeReasonMalformedMetadataJSON,
	}
	errDuplicateJSONKey = errors.New("duplicate json object key")
	errJSONNotObject    = errors.New("json value is not an object")
)

type UploadEnvelopePolicy struct {
	FileContentTypes []string
}

type UploadEnvelope struct {
	Metadata              map[string]json.RawMessage
	MetadataRaw           json.RawMessage
	File                  []byte
	FileSHA256Hex         string
	FileContentType       string
	FileContentTypeHeader string
	FileName              string
}

type UploadEnvelopeError struct {
	ReasonCode          string
	PartName            string
	ReceivedContentType *string
	AllowedContentTypes []string
}

func (e *UploadEnvelopeError) Error() string {
	if e == nil {
		return ""
	}
	if e.PartName == "" {
		return e.ReasonCode
	}
	return fmt.Sprintf("%s: %s", e.ReasonCode, e.PartName)
}

func (e *UploadEnvelopeError) Details() map[string]any {
	details := map[string]any{
		"reason_code": e.ReasonCode,
	}
	if e.PartName != "" {
		details["part_name"] = e.PartName
	}
	if e.ReasonCode == UploadEnvelopeReasonInvalidPartContentType {
		if e.ReceivedContentType == nil {
			details["received_content_type"] = nil
		} else {
			details["received_content_type"] = *e.ReceivedContentType
		}
		details["allowed_content_types"] = append([]string(nil), e.AllowedContentTypes...)
	}
	return details
}

func SharedUploadEnvelopeReasonCodes() []string {
	return append([]string(nil), sharedUploadEnvelopeReasonCodes...)
}

func MetadataUploadEnvelopeContentTypes() []string {
	return append([]string(nil), metadataUploadEnvelopeContentTypes...)
}

func ParseUploadEnvelope(r *http.Request, policy UploadEnvelopePolicy) (UploadEnvelope, *UploadEnvelopeError) {
	mediaType, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	if err != nil || !strings.EqualFold(mediaType, "multipart/form-data") || params["boundary"] == "" {
		return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonUnsupportedUploadEnvelope, "", nil, nil)
	}

	allowedFileTypes := canonicalFileContentTypes(policy.FileContentTypes)
	reader := multipart.NewReader(r.Body, params["boundary"])
	var envelope UploadEnvelope
	var sawMetadata bool
	var sawFile bool

	for {
		part, err := reader.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonUnsupportedUploadEnvelope, "", nil, nil)
		}

		partName := part.FormName()
		switch partName {
		case "metadata":
			if sawMetadata {
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonDuplicatePart, "metadata", nil, nil)
			}
			sawMetadata = true
			if isNestedMultipart(part.Header.Get("Content-Type")) {
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonUnsupportedUploadEnvelope, "metadata", nil, nil)
			}
			if ok := metadataContentTypeAllowed(part.Header.Get("Content-Type")); !ok {
				var received *string
				if rawContentType := part.Header.Get("Content-Type"); rawContentType != "" {
					received = &rawContentType
				}
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonInvalidPartContentType, "metadata", received, metadataUploadEnvelopeContentTypes)
			}
			metadataBytes, err := io.ReadAll(part)
			if err != nil {
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonMalformedMetadataJSON, "metadata", nil, nil)
			}
			metadata, apiErr := parseMetadataJSON(metadataBytes)
			if apiErr != nil {
				return UploadEnvelope{}, apiErr
			}
			envelope.Metadata = metadata
			envelope.MetadataRaw = append(json.RawMessage(nil), metadataBytes...)
		case "file":
			if sawFile {
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonDuplicatePart, "file", nil, nil)
			}
			sawFile = true
			if isNestedMultipart(part.Header.Get("Content-Type")) {
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonUnsupportedUploadEnvelope, "file", nil, nil)
			}
			rawContentType := part.Header.Get("Content-Type")
			normalizedContentType, ok := normalizeContentType(rawContentType)
			if !ok || !stringInSlice(normalizedContentType, allowedFileTypes) {
				var received *string
				if rawContentType != "" {
					received = &rawContentType
				}
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonInvalidPartContentType, "file", received, allowedFileTypes)
			}
			fileBytes, err := io.ReadAll(part)
			if err != nil {
				return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonUnsupportedUploadEnvelope, "file", nil, nil)
			}
			hash := sha256.Sum256(fileBytes)
			envelope.File = fileBytes
			envelope.FileSHA256Hex = hex.EncodeToString(hash[:])
			envelope.FileContentType = normalizedContentType
			envelope.FileContentTypeHeader = rawContentType
			envelope.FileName = part.FileName()
		default:
			return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonUnexpectedPart, "", nil, nil)
		}
	}

	if !sawMetadata {
		return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonMissingRequiredPart, "metadata", nil, nil)
	}
	if !sawFile {
		return UploadEnvelope{}, uploadEnvelopeError(UploadEnvelopeReasonMissingRequiredPart, "file", nil, nil)
	}
	return envelope, nil
}

func uploadEnvelopeError(reasonCode string, partName string, receivedContentType *string, allowedContentTypes []string) *UploadEnvelopeError {
	return &UploadEnvelopeError{
		ReasonCode:          reasonCode,
		PartName:            partName,
		ReceivedContentType: receivedContentType,
		AllowedContentTypes: append([]string(nil), allowedContentTypes...),
	}
}

func parseMetadataJSON(metadataBytes []byte) (map[string]json.RawMessage, *UploadEnvelopeError) {
	if bytes.HasPrefix(metadataBytes, []byte{0xef, 0xbb, 0xbf}) || !utf8.Valid(metadataBytes) {
		return nil, uploadEnvelopeError(UploadEnvelopeReasonInvalidMetadataEncoding, "metadata", nil, nil)
	}
	if err := validateJSONObjectNoDuplicateKeys(metadataBytes); err != nil {
		if errors.Is(err, errJSONNotObject) {
			return nil, uploadEnvelopeError(UploadEnvelopeReasonRequestNotObject, "metadata", nil, nil)
		}
		return nil, uploadEnvelopeError(UploadEnvelopeReasonMalformedMetadataJSON, "metadata", nil, nil)
	}
	var metadata map[string]json.RawMessage
	if err := json.Unmarshal(metadataBytes, &metadata); err != nil {
		return nil, uploadEnvelopeError(UploadEnvelopeReasonMalformedMetadataJSON, "metadata", nil, nil)
	}
	return metadata, nil
}

func metadataContentTypeAllowed(raw string) bool {
	mediaType, params, err := mime.ParseMediaType(raw)
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return false
	}
	if len(params) == 0 {
		return true
	}
	if len(params) == 1 && strings.EqualFold(params["charset"], "utf-8") {
		return true
	}
	return false
}

func isNestedMultipart(raw string) bool {
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return false
	}
	return strings.HasPrefix(strings.ToLower(mediaType), "multipart/")
}

func canonicalFileContentTypes(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		normalized, ok := normalizeContentType(value)
		if !ok {
			normalized = strings.ToLower(strings.TrimSpace(value))
		}
		if normalized == "" {
			continue
		}
		seen[normalized] = struct{}{}
	}
	canonical := make([]string, 0, len(seen))
	for value := range seen {
		canonical = append(canonical, value)
	}
	sort.Strings(canonical)
	return canonical
}

func normalizeContentType(raw string) (string, bool) {
	if strings.TrimSpace(raw) == "" {
		return "", false
	}
	mediaType, _, err := mime.ParseMediaType(raw)
	if err != nil {
		return "", false
	}
	mediaType = strings.ToLower(strings.TrimSpace(mediaType))
	if !strings.Contains(mediaType, "/") {
		return "", false
	}
	return mediaType, true
}

func stringInSlice(value string, values []string) bool {
	index := sort.SearchStrings(values, value)
	return index < len(values) && values[index] == value
}

func validateJSONObjectNoDuplicateKeys(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok || delimiter != '{' {
		return errJSONNotObject
	}
	if err := consumeJSONObject(decoder); err != nil {
		return err
	}
	if token, err := decoder.Token(); err != nil {
		if errors.Is(err, io.EOF) {
			return nil
		}
		return err
	} else if token != nil {
		return errors.New("metadata json contains multiple values")
	}
	return nil
}

func consumeJSONObject(decoder *json.Decoder) error {
	seen := make(map[string]struct{})
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return errors.New("json object key is not a string")
		}
		if _, ok := seen[key]; ok {
			return errDuplicateJSONKey
		}
		seen[key] = struct{}{}
		if err := consumeJSONValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != '}' {
		return errors.New("json object not terminated")
	}
	return nil
}

func consumeJSONArray(decoder *json.Decoder) error {
	for decoder.More() {
		if err := consumeJSONValue(decoder); err != nil {
			return err
		}
	}
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if delimiter, ok := token.(json.Delim); !ok || delimiter != ']' {
		return errors.New("json array not terminated")
	}
	return nil
}

func consumeJSONValue(decoder *json.Decoder) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	delimiter, ok := token.(json.Delim)
	if !ok {
		return nil
	}
	switch delimiter {
	case '{':
		return consumeJSONObject(decoder)
	case '[':
		return consumeJSONArray(decoder)
	default:
		return errors.New("unexpected json delimiter")
	}
}
