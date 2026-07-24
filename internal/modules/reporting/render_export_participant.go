package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

const (
	RenderExportMaxInputBytes  int64 = 67_108_864
	RenderExportMaxOutputBytes int64 = 67_108_864
	RenderExportMaxItems             = 1_048_576
)

var ErrRenderExportParticipant = errors.New("reporting: render export participant rejected invocation")

type RenderExportContext struct {
	SchemaID                string  `json:"schema_id"`
	Operation               string  `json:"operation"`
	ProfileID               string  `json:"profile_id"`
	ContractMajor           int     `json:"contract_major"`
	ClaimState              string  `json:"claim_state"`
	StatePresent            bool    `json:"state_present"`
	StateVersion            *string `json:"state_version"`
	SnapshotRef             string  `json:"snapshot_ref"`
	AuthorizationViewSHA256 string  `json:"authorization_view_sha256"`
	RedactionProfileSHA256  string  `json:"redaction_profile_sha256"`
	TimeoutSeconds          int     `json:"timeout_seconds"`
}

type RenderExportInvocation struct {
	Context           RenderExportContext
	ImmutableModel    ExportModel
	ImmutableModelSHA string
}

type RenderExportResult struct {
	SchemaID       string `json:"schema_id"`
	Kind           string `json:"kind"`
	OutputSchema   string `json:"output_schema_id"`
	OutputSHA256   string `json:"output_sha256"`
	OutputByteSize int64  `json:"output_byte_size"`
	OutputRef      string `json:"output_ref"`
	ItemCount      int    `json:"item_count"`
	Output         []byte `json:"-"`
}

type RenderExportParticipant interface {
	Emit(context.Context, RenderExportInvocation) (RenderExportResult, error)
}

type RenderExportInvoker interface {
	Invoke(context.Context, RenderExportInvocation) (RenderExportResult, error)
}

type BuiltInRenderExportParticipant struct{}

func (BuiltInRenderExportParticipant) Emit(ctx context.Context, invocation RenderExportInvocation) (RenderExportResult, error) {
	if err := ctx.Err(); err != nil {
		return RenderExportResult{}, err
	}
	contextBytes, err := canonicalJSON(invocation.Context)
	if err != nil || int64(len(contextBytes)) > RenderExportMaxInputBytes {
		return RenderExportResult{}, fmt.Errorf("%w: context bounds", ErrRenderExportParticipant)
	}
	contextValue := invocation.Context
	if contextValue.SchemaID != RenderExportContextSchemaID ||
		contextValue.Operation != RenderExportOperationKind ||
		contextValue.ProfileID != ProfileID ||
		contextValue.ContractMajor != 1 ||
		contextValue.ClaimState != "claimed" ||
		contextValue.StatePresent ||
		contextValue.StateVersion != nil ||
		contextValue.TimeoutSeconds <= 0 ||
		contextValue.SnapshotRef != "snapshot:"+invocation.ImmutableModel.SnapshotID ||
		!sha256HexPattern.MatchString(contextValue.AuthorizationViewSHA256) ||
		!sha256HexPattern.MatchString(contextValue.RedactionProfileSHA256) {
		return RenderExportResult{}, fmt.Errorf("%w: invalid context", ErrRenderExportParticipant)
	}
	if invocation.ImmutableModel.SchemaID != ExportModelSchemaID ||
		invocation.ImmutableModel.DerivationVersion != DerivationVersion {
		return RenderExportResult{}, fmt.Errorf("%w: invalid immutable model", ErrRenderExportParticipant)
	}
	output, err := canonicalJSON(invocation.ImmutableModel)
	if err != nil {
		return RenderExportResult{}, err
	}
	if int64(len(output)) > RenderExportMaxOutputBytes {
		return RenderExportResult{}, fmt.Errorf("%w: output bounds", ErrRenderExportParticipant)
	}
	digest := hashHex(output)
	if digest != invocation.ImmutableModelSHA {
		return RenderExportResult{}, fmt.Errorf("%w: immutable model digest", ErrRenderExportParticipant)
	}
	itemCount := renderExportItemCount(invocation.ImmutableModel)
	if itemCount > RenderExportMaxItems {
		return RenderExportResult{}, fmt.Errorf("%w: item bounds", ErrRenderExportParticipant)
	}
	if err := ctx.Err(); err != nil {
		return RenderExportResult{}, err
	}
	return RenderExportResult{
		SchemaID:       RenderExportResultSchemaID,
		Kind:           "output",
		OutputSchema:   ExportModelSchemaID,
		OutputSHA256:   digest,
		OutputByteSize: int64(len(output)),
		OutputRef:      "snapshot:" + invocation.ImmutableModel.SnapshotID + "/reporting-export-model:" + digest,
		ItemCount:      itemCount,
		Output:         output,
	}, nil
}

func AdmitRenderExportResult(invocation RenderExportInvocation, result RenderExportResult) (ExportModel, string, error) {
	if result.SchemaID != RenderExportResultSchemaID ||
		result.Kind != "output" ||
		result.OutputSchema != ExportModelSchemaID ||
		!sha256HexPattern.MatchString(result.OutputSHA256) ||
		result.OutputSHA256 != hashHex(result.Output) ||
		result.OutputByteSize != int64(len(result.Output)) ||
		result.OutputByteSize <= 0 ||
		result.OutputByteSize > RenderExportMaxOutputBytes ||
		result.ItemCount < 0 ||
		result.ItemCount > RenderExportMaxItems ||
		result.OutputRef != "snapshot:"+invocation.ImmutableModel.SnapshotID+"/reporting-export-model:"+result.OutputSHA256 {
		return ExportModel{}, "", fmt.Errorf("%w: invalid result", ErrRenderExportParticipant)
	}
	decoder := json.NewDecoder(bytes.NewReader(result.Output))
	decoder.DisallowUnknownFields()
	var model ExportModel
	if err := decoder.Decode(&model); err != nil {
		return ExportModel{}, "", fmt.Errorf("%w: decode output: %v", ErrRenderExportParticipant, err)
	}
	if err := requireJSONEOF(decoder); err != nil {
		return ExportModel{}, "", fmt.Errorf("%w: trailing output", ErrRenderExportParticipant)
	}
	if model.SchemaID != ExportModelSchemaID ||
		model.DerivationVersion != DerivationVersion ||
		model.SnapshotID != invocation.ImmutableModel.SnapshotID ||
		model.IncidentID != invocation.ImmutableModel.IncidentID ||
		renderExportItemCount(model) != result.ItemCount {
		return ExportModel{}, "", fmt.Errorf("%w: output binding", ErrRenderExportParticipant)
	}
	canonical, err := canonicalJSON(model)
	if err != nil || !bytes.Equal(canonical, result.Output) {
		return ExportModel{}, "", fmt.Errorf("%w: non-canonical output", ErrRenderExportParticipant)
	}
	model.Fields = model.RedactionFields()
	return model, result.OutputSHA256, nil
}

func renderExportItemCount(model ExportModel) int {
	count := len(model.Sections) + len(model.Records) + len(model.Relationships) +
		len(model.TimelineEvents) + len(model.Subjects) + len(model.Diagrams) +
		len(model.Assets) + len(model.SupportIndex)
	var visit func([]ReportingBlock)
	visit = func(blocks []ReportingBlock) {
		count += len(blocks)
		for _, block := range blocks {
			visit(block.Children)
		}
	}
	for _, section := range model.Sections {
		visit(section.Blocks)
	}
	return count
}

func requireJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("additional JSON value")
		}
		return err
	}
	return nil
}

func renderExportTimeoutSeconds(timeout time.Duration) int {
	seconds := int(timeout / time.Second)
	if seconds < 1 {
		return 1
	}
	return seconds
}
