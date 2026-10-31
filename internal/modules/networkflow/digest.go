package networkflow

import (
	"bytes"
	"cmp"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strconv"

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

func sourceRowDigest(parserProfileID string, sourceRowNumber int64, decodedFields []string) string {
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

func normalizedRowDigest(mappingFingerprint string, values map[string]any, unmappedRaw retainedRawValues) (string, error) {
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
		if err := writeCanonical(&b, value); err != nil {
			return "", err
		}
		b.WriteByte(0)
	}
	writeDigestPart(&b, "unmapped_raw")
	if err := writeCanonical(&b, unmappedRaw); err != nil {
		return "", err
	}
	b.WriteByte(0)
	return sha256Hex(b.Bytes()), nil
}

func rowID(incidentID uuid.UUID, tableID string, sourceRowNumber int64, sourceRowDigest string, normalizedRowDigest string) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow_row_id.v1")
	writeDigestPart(&b, incidentID.String())
	writeDigestPart(&b, tableID)
	writeDigestPart(&b, strconv.FormatInt(sourceRowNumber, 10))
	writeDigestPart(&b, sourceRowDigest)
	writeDigestPart(&b, normalizedRowDigest)
	return "nfr_" + sha256Hex(b.Bytes())
}

func endpointID(incidentID uuid.UUID, endpointKind string, canonicalValue string) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow_endpoint_id.v1")
	writeDigestPart(&b, incidentID.String())
	writeDigestPart(&b, endpointKind)
	writeDigestPart(&b, canonicalValue)
	return "nfe_" + sha256Hex(b.Bytes())
}

func flowEdgeID(incidentID uuid.UUID, srcEndpointID string, dstEndpointID string, ipProtocol int32, dstPort *int32) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow_flow_edge_id.v1")
	writeDigestPart(&b, incidentID.String())
	writeDigestPart(&b, srcEndpointID)
	writeDigestPart(&b, dstEndpointID)
	writeDigestPart(&b, strconv.FormatInt(int64(ipProtocol), 10))
	if dstPort == nil {
		writeDigestPart(&b, "n")
		writeDigestPart(&b, "")
	} else {
		writeDigestPart(&b, "p")
		writeDigestPart(&b, strconv.FormatInt(int64(*dstPort), 10))
	}
	return "nff_" + sha256Hex(b.Bytes())
}

func diagnosticID(sourceSHA256, fingerprint string, sourceRowNumber int64, sourceColumnOrdinal *int64, fieldKey *string, errorCode, reasonCode string) string {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow.diagnostic_id.v1")
	writeDigestPart(&b, sourceSHA256)
	writeDigestPart(&b, fingerprint)
	writeDigestPart(&b, strconv.FormatInt(sourceRowNumber, 10))
	if sourceColumnOrdinal == nil {
		writeDigestPart(&b, "n")
	} else {
		writeDigestPart(&b, "p:"+strconv.FormatInt(*sourceColumnOrdinal, 10))
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

func mappingFingerprint(mapping approvedMapping, sourceContentSHA256 string) (string, error) {
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow_mapping_fingerprint.v1")
	writeDigestPart(&b, mapping.TargetKind)
	writeDigestPart(&b, mapping.TargetTableSchemaID)
	writeDigestPart(&b, mapping.SourceProfileID)
	writeDigestPart(&b, mapping.ParserProfileID)
	writeDigestPart(&b, sourceContentSHA256)
	writeDigestPart(&b, mapping.UnknownColumnPolicy)
	if err := writeCanonical(&b, mapping.TimestampProfile); err != nil {
		return "", err
	}
	b.WriteByte(0)
	sourceColumns := append([]sourceColumnDescriptor(nil), mapping.SourceColumns...)
	sort.SliceStable(sourceColumns, func(i, j int) bool {
		return sourceColumns[i].SourceColumnOrdinal < sourceColumns[j].SourceColumnOrdinal
	})
	for _, column := range sourceColumns {
		writeDigestPart(&b, strconv.Itoa(column.SourceColumnOrdinal))
		writeDigestPart(&b, column.RawHeaderText)
		writeDigestPart(&b, column.NormalizedHeaderForSuggestion)
		writeDigestPart(&b, column.RawHeaderSHA256)
	}
	fieldMappings := append([]fieldMapping(nil), mapping.FieldMappings...)
	sort.SliceStable(fieldMappings, func(i, j int) bool {
		return compareFieldMappings(fieldMappings[i], fieldMappings[j]) < 0
	})
	for _, fieldMapping := range fieldMappings {
		if err := writeCanonical(&b, fieldMapping); err != nil {
			return "", err
		}
		b.WriteByte(0)
	}
	return sha256Hex(b.Bytes()), nil
}

var safeDigestClassPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)

func safeDigest(keyID string, key []byte, valueClass string, canonicalValue string) (string, string, error) {
	if !safeKeyIDPattern.MatchString(keyID) || len(key) != 32 || !safeDigestClassPattern.MatchString(valueClass) {
		return "", "", fmt.Errorf("network flow safe-digest material is invalid")
	}
	var b bytes.Buffer
	writeDigestPart(&b, "cartulary.network_flow.safe_digest.v1")
	writeDigestPart(&b, valueClass)
	b.WriteString(canonicalValue)
	return hmacSHA256Hex(key, b.Bytes()), keyID, nil
}

func writeDigestPart(b *bytes.Buffer, value string) {
	b.WriteString(value)
	b.WriteByte(0)
}

func writeLengthFramedPart(b *bytes.Buffer, value string) {
	var length [8]byte
	binary.BigEndian.PutUint64(length[:], uint64(len([]byte(value))))
	b.Write(length[:])
	b.WriteString(value)
}

func registeredRowFieldKeys() []string {
	return []string{
		fieldFlowStartUTC,
		fieldFlowEndUTC,
		fieldSrcIP,
		fieldDstIP,
		fieldSrcPort,
		fieldDstPort,
		fieldIPProtocol,
		fieldBytesCount,
		fieldPacketsCount,
		fieldExporterID,
		fieldInputInterface,
		fieldOutputInterface,
		fieldTCPFlags,
		fieldApplicationLabel,
		fieldObservationSourceRef,
	}
}

// compareFieldMappings implements the typed NF-042 tuple. Ordinals are numbers,
// so their ordering must not depend on the length of their decimal spelling.
func compareFieldMappings(a, b fieldMapping) int {
	for _, comparison := range []int{
		cmp.Compare(a.FieldKey, b.FieldKey),
		cmp.Compare(a.MappingKind, b.MappingKind),
		cmp.Compare(a.SourceColumnOrdinal, b.SourceColumnOrdinal),
		cmp.Compare(a.DerivationID, b.DerivationID),
		cmp.Compare(a.IgnoreReason, b.IgnoreReason),
	} {
		if comparison != 0 {
			return comparison
		}
	}
	return 0
}
