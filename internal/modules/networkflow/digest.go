package networkflow

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

func hmacSHA256Hex(key []byte, data []byte) string {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write(data)
	return hex.EncodeToString(mac.Sum(nil))
}

func SourceRowDigest(parserProfileID string, sourceRowNumber int64, decodedFields []string) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow.source_row_digest.v1")
	writeDigestPart(&b, parserProfileID)
	writeDigestPart(&b, strconv.FormatInt(sourceRowNumber, 10))
	writeDigestPart(&b, strconv.Itoa(len(decodedFields)))
	for index, field := range decodedFields {
		writeDigestPart(&b, strconv.Itoa(index+1))
		writeDigestPart(&b, strconv.Itoa(len([]byte(field))))
		writeDigestPart(&b, field)
	}
	return sha256Hex(b.Bytes())
}

func NormalizedRowDigest(mappingFingerprint string, values map[string]any, unmappedRaw map[string]any) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow.normalized_row_digest.v1")
	writeDigestPart(&b, mappingFingerprint)
	for _, fieldKey := range registeredRowFieldKeys() {
		writeDigestPart(&b, fieldKey)
		value, ok := values[fieldKey]
		if !ok || value == nil {
			writeDigestPart(&b, "n")
			writeDigestPart(&b, "null")
			continue
		}
		writeDigestPart(&b, "p")
		b.Write(canonicalJSON(value))
		b.WriteByte(0)
	}
	writeDigestPart(&b, "unmapped_raw")
	b.Write(canonicalJSON(unmappedRaw))
	b.WriteByte(0)
	return sha256Hex(b.Bytes())
}

func RowID(incidentID uuid.UUID, tableID string, sourceRowNumber int64, sourceRowDigest string, normalizedRowDigest string) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow_row_id.v1")
	writeDigestPart(&b, incidentID.String())
	writeDigestPart(&b, tableID)
	writeDigestPart(&b, strconv.FormatInt(sourceRowNumber, 10))
	writeDigestPart(&b, sourceRowDigest)
	writeDigestPart(&b, normalizedRowDigest)
	return "nfr_" + sha256Hex(b.Bytes())
}

func DiagnosticID(sourceRowNumber int64, sourceColumnOrdinal *int64, rawHeaderSHA256 *string, fieldKey *string, errorCode string, reasonCode string) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow_diagnostic_id.v1")
	writeDigestPart(&b, strconv.FormatInt(sourceRowNumber, 10))
	if sourceColumnOrdinal == nil {
		writeDigestPart(&b, "n")
	} else {
		writeDigestPart(&b, "p:"+strconv.FormatInt(*sourceColumnOrdinal, 10))
	}
	if rawHeaderSHA256 == nil {
		writeDigestPart(&b, "n")
	} else {
		writeDigestPart(&b, "p:"+*rawHeaderSHA256)
	}
	if fieldKey == nil {
		writeDigestPart(&b, "n")
	} else {
		writeDigestPart(&b, "p:"+*fieldKey)
	}
	writeDigestPart(&b, errorCode)
	writeDigestPart(&b, reasonCode)
	return "nfd_" + sha256Hex(b.Bytes())
}

func MappingFingerprint(mapping ApprovedMapping, sourceContentSHA256 string) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow_mapping_fingerprint.v1")
	writeDigestPart(&b, mapping.TargetKind)
	writeDigestPart(&b, mapping.TargetTableSchemaID)
	writeDigestPart(&b, mapping.SourceProfileID)
	writeDigestPart(&b, mapping.ParserProfileID)
	writeDigestPart(&b, sourceContentSHA256)
	writeDigestPart(&b, mapping.UnknownColumnPolicy)
	b.Write(canonicalJSON(mapping.TimestampProfile))
	b.WriteByte(0)
	sourceColumns := append([]SourceColumnDescriptor(nil), mapping.SourceColumns...)
	sort.SliceStable(sourceColumns, func(i, j int) bool {
		return sourceColumns[i].SourceColumnOrdinal < sourceColumns[j].SourceColumnOrdinal
	})
	for _, column := range sourceColumns {
		writeDigestPart(&b, strconv.Itoa(column.SourceColumnOrdinal))
		writeDigestPart(&b, column.RawHeaderText)
		writeDigestPart(&b, column.NormalizedHeaderForSuggestion)
		writeDigestPart(&b, column.RawHeaderSHA256)
	}
	fieldMappings := append([]FieldMapping(nil), mapping.FieldMappings...)
	sort.SliceStable(fieldMappings, func(i, j int) bool {
		return mappingSortKey(fieldMappings[i]) < mappingSortKey(fieldMappings[j])
	})
	for _, fieldMapping := range fieldMappings {
		b.Write(canonicalJSON(fieldMapping))
		b.WriteByte(0)
	}
	return sha256Hex(b.Bytes())
}

func SafeDigest(keyID string, key []byte, valueClass string, canonicalValue string) (string, string) {
	if keyID == "" || len(key) == 0 {
		keyID = "network-flow-ws14"
		key = []byte("network-flow-ws14-engineering-only-safe-digest-key")
	}
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow.safe_digest.v1")
	writeDigestPart(&b, valueClass)
	b.WriteString(canonicalValue)
	return hmacSHA256Hex(key, b.Bytes()), keyID
}

func writeDigestPart(b *bytes.Buffer, value string) {
	b.WriteString(value)
	b.WriteByte(0)
}

func canonicalJSON(value any) []byte {
	switch v := value.(type) {
	case nil:
		return []byte("null")
	case json.RawMessage:
		if len(v) == 0 {
			return []byte("null")
		}
		var decoded any
		if err := json.Unmarshal(v, &decoded); err != nil {
			return []byte("null")
		}
		return canonicalJSON(decoded)
	case string:
		data, _ := json.Marshal(v)
		return data
	case bool:
		if v {
			return []byte("true")
		}
		return []byte("false")
	case int:
		return []byte(strconv.Itoa(v))
	case int32:
		return []byte(strconv.FormatInt(int64(v), 10))
	case int64:
		return []byte(strconv.FormatInt(v, 10))
	case float64:
		if v == float64(int64(v)) {
			return []byte(strconv.FormatInt(int64(v), 10))
		}
		data, _ := json.Marshal(v)
		return data
	case map[string]any:
		return canonicalJSONObject(v)
	case []any:
		return canonicalJSONArray(v)
	default:
		data, err := json.Marshal(v)
		if err != nil {
			encoded, _ := json.Marshal(fmt.Sprint(v))
			return encoded
		}
		var decoded any
		if err := json.Unmarshal(data, &decoded); err != nil {
			return data
		}
		return canonicalJSON(decoded)
	}
}

func canonicalJSONObject(values map[string]any) []byte {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	var b bytes.Buffer
	b.WriteByte('{')
	for index, key := range keys {
		if index > 0 {
			b.WriteByte(',')
		}
		keyJSON, _ := json.Marshal(key)
		b.Write(keyJSON)
		b.WriteByte(':')
		b.Write(canonicalJSON(values[key]))
	}
	b.WriteByte('}')
	return b.Bytes()
}

func canonicalJSONArray(values []any) []byte {
	var b bytes.Buffer
	b.WriteByte('[')
	for index, value := range values {
		if index > 0 {
			b.WriteByte(',')
		}
		b.Write(canonicalJSON(value))
	}
	b.WriteByte(']')
	return b.Bytes()
}

func registeredRowFieldKeys() []string {
	return []string{
		FieldFlowStartUTC,
		FieldFlowEndUTC,
		FieldSrcIP,
		FieldDstIP,
		FieldSrcPort,
		FieldDstPort,
		FieldIPProtocol,
		FieldBytesCount,
		FieldPacketsCount,
		FieldExporterID,
		FieldInputInterface,
		FieldOutputInterface,
		FieldTCPFlags,
		FieldApplicationLabel,
		FieldObservationSourceRef,
	}
}

func mappingSortKey(mapping FieldMapping) string {
	switch mapping.MappingKind {
	case MappingKindSourceColumn:
		return strings.Join([]string{mapping.FieldKey, "source_column", strconv.Itoa(mapping.SourceColumnOrdinal), "", ""}, "\x00")
	case MappingKindSystemDerivation:
		return strings.Join([]string{mapping.FieldKey, "system_derivation", "0", mapping.DerivationID, ""}, "\x00")
	case MappingKindIgnoredSourceColumn:
		return strings.Join([]string{"", "ignored_source_column", strconv.Itoa(mapping.SourceColumnOrdinal), "", mapping.IgnoreReason}, "\x00")
	default:
		return strings.Join([]string{mapping.FieldKey, mapping.MappingKind, strconv.Itoa(mapping.SourceColumnOrdinal), mapping.DerivationID, mapping.IgnoreReason}, "\x00")
	}
}
