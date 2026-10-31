package networkflow

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

func TestImportSourceInspectionHeaderOnly(t *testing.T) {
	_, preview, _ := approvedVariantFixture(t, sourceProfileCiscoSNANetFlowCSV)
	mapping, err := decodeApprovedMapping(preview.OwnerMapping, defaultLimits())
	if err != nil {
		t.Fatal(err)
	}
	for _, tail := range []string{"", validPreviewRow, "\"unclosed\n", "\xff"} {
		source := strings.Join(requiredFields(sourceProfileCiscoSNANetFlowCSV), ",") + "\n" + tail
		f := &importFacade{sourceStore: previewSource{csv: source}, limits: defaultLimits()}
		request := imports.ExtensionImportSourceRequest{TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID, SourceCapability: imports.ImportSourceCapability{SourceContentSHA256: sha256Hex([]byte(source))}}
		result, err := f.InspectImportUnitSource(context.Background(), request)
		if err != nil {
			t.Fatalf("inspection parsed data records: %v", err)
		}
		if err := f.ValidateImportUnitSourceResult(result); err != nil {
			t.Fatalf("inspection result rejected: %v", err)
		}
		columns, err := decodeSourceColumns(mappingTestJSON(t, result.OwnerResult["source_columns"]), defaultLimits())
		if err != nil {
			t.Fatalf("producer-derived zero-count control rejected: %v", err)
		}
		for _, column := range columns {
			if len(column.SampleValues) != 0 {
				t.Fatal("inspection sampled a data record")
			}
		}
		if _, err := decodeSourceColumns(mappingTestJSON(t, columns[:1]), defaultLimits()); err != nil {
			t.Fatalf("single-column zero-count control rejected: %v", err)
		}
		for _, equal := range []bool{false, true} {
			badColumns := append([]sourceColumnDescriptor(nil), columns...)
			for i := range badColumns {
				if equal || i == 0 {
					badColumns[i].SampleValues = mapping.SourceColumns[i].SampleValues
				}
			}
			bad := result
			bad.OwnerResult = make(map[string]any, len(result.OwnerResult))
			for key, value := range result.OwnerResult {
				bad.OwnerResult[key] = value
			}
			bad.OwnerResult["source_columns"] = badColumns
			if err := f.ValidateImportUnitSourceResult(bad); !errors.Is(err, errInvalidImportObject) {
				t.Fatalf("inspection admitted samples (equal=%v): %v", equal, err)
			}
		}
		if result.OwnerResult["source_content_sha256"] != request.SourceCapability.SourceContentSHA256 {
			t.Fatal("inspection changed source identity")
		}
		request.SourceCapability.SourceContentSHA256 = strings.Repeat("0", 64)
		if _, err := f.InspectImportUnitSource(context.Background(), request); err == nil {
			t.Fatal("inspection certified corrupt content")
		}
		result.OwnerResult["unexpected"] = true
		if f.ValidateImportUnitSourceResult(result) == nil {
			t.Fatal("inspection admitted unknown response member")
		}
	}
	for _, header := range []string{"\"unclosed", "\xff\n"} {
		f := &importFacade{sourceStore: previewSource{csv: header}, limits: defaultLimits()}
		if _, err := f.InspectImportUnitSource(context.Background(), imports.ExtensionImportSourceRequest{TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID}); err == nil {
			t.Fatal("invalid header admitted")
		}
	}
}
