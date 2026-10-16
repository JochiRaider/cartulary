package reportingfixture

import (
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingassets"
)

// TemplatePresentationPack provides static public presentation bytes for ordinary
// admission/portability tests. It grants no storage or renderer capability.
func TemplatePresentationPack() reporting.TemplatePack {
	contract, _ := reporting.ResolveTemplateContract(reporting.DefaultTemplateID, reporting.DefaultTemplateVersion)
	contract.TemplateID, contract.TemplateVersion = "test.public-assets", "1"
	contract.SupportedOutputKinds = []string{reporting.OutputKindSlidev}
	contract.SupportedOutputOptions = contract.SupportedOutputOptions[1:]
	contract.AllowedExportModelBindings = []string{"$.assets['mark']", "$.sections[*]"}
	contract.Sections[0].Bindings = []string{"$.assets['mark']", "$.sections[*]"}
	data := []byte(`<svg xmlns="http://www.w3.org/2000/svg" width="20" height="20"><rect width="20" height="20" fill="#238853"/></svg>`)
	contract.Assets = []reporting.TemplateAsset{{SchemaID: "template_asset_item.v1", AssetID: "mark", AssetRole: "local_asset", BundlePath: "assets/mark/mark.svg", SourcePath: "images/mark.svg", MediaType: "image/svg+xml", SHA256: digest(data), ByteSize: int64(len(data)), RequiredForRelease: true}}
	files := []reporting.TemplatePackFile{{SourcePath: "images/mark.svg", Bytes: data}}
	for _, file := range []struct {
		id, path, media string
		bytes           []byte
	}{
		{"body", "theme/fonts/body.woff2", "font/woff2", reportingassets.ThemeFont()},
		{"theme", "theme/style.css", "text/css; charset=utf-8", []byte(`@font-face {font-family: "rpt-body";src: url("fonts/body.woff2") format("woff2");font-style: normal;font-weight: 400;font-display: block;}
:root {--cartulary-report-text: #183650;--cartulary-report-font-family: "rpt-body";}`)},
	} {
		contract.Assets = append(contract.Assets, reporting.TemplateAsset{SchemaID: "template_asset_item.v1", AssetID: file.id, AssetRole: "local_theme", BundlePath: file.path, SourcePath: file.path, MediaType: file.media, SHA256: digest(file.bytes), ByteSize: int64(len(file.bytes)), RequiredForRelease: true})
		files = append(files, reporting.TemplatePackFile{SourcePath: file.path, Bytes: file.bytes})
	}
	return reporting.TemplatePack{Manifest: contract, Files: files}
}
