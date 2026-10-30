package networkflow

import (
	"context"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/imports"
)

func TestImportSourceInspectionHeaderOnly(t *testing.T) {
	for _, tail := range []string{"", validPreviewRow, "\"unclosed\n", "\xff"} {
		source := strings.Join(requiredCiscoFields(), ",") + "\n" + tail
		f := &importFacade{sourceStore: previewSource{csv: source}, limits: defaultLimits()}
		request := imports.ExtensionImportSourceRequest{TargetKind: targetKindNetworkFlowTable, ExtensionProfileID: ProfileID, SourceCapability: imports.ImportSourceCapability{SourceContentSHA256: sha256Hex([]byte(source))}}
		result, err := f.InspectImportUnitSource(context.Background(), request)
		if err != nil {
			t.Fatalf("inspection parsed data records: %v", err)
		}
		if err := f.ValidateImportUnitSourceResult(result); err != nil {
			t.Fatalf("inspection result rejected: %v", err)
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
