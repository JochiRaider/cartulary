package reporting

import (
	"encoding/binary"
	"path"
	"regexp"
	"strings"
)

const themeCSSMedia = "text/css; charset=utf-8"

var themeAssetID = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,63}$`)
var themeColor = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var themeFontSize = regexp.MustCompile(`^(1[89]|2[0-9]|3[0-2])px$`)

func themeFileDeclaration(name, media string) TemplateAsset {
	return TemplateAsset{AssetID: strings.TrimSuffix(path.Base(name), path.Ext(name)), AssetRole: "local_theme", BundlePath: name, SourcePath: name, MediaType: media}
}

func validateThemeBundleFiles(files []RenderBundleFile) error {
	contract := TemplateContract{SupportedOutputKinds: []string{OutputKindSlidev}}
	contents := []TemplatePackFile{}
	for _, file := range files {
		if file.Role != renderBundleRoleLocalTheme {
			continue
		}
		asset := themeFileDeclaration(file.Path, file.MediaType)
		if !templateThemeMember(asset) {
			return newRenderValidationError("bundle_manifest_invalid", "invalid_asset_media_type")
		}
		if err := validateTemplateFileContent(file.MediaType, file.Bytes); err != nil {
			return err
		}
		contract.Assets = append(contract.Assets, asset)
		contents = append(contents, TemplatePackFile{SourcePath: file.Path, Bytes: file.Bytes})
	}
	return validateTemplateTheme(contract, contents)
}

func templateThemeMember(asset TemplateAsset) bool {
	if asset.AssetRole != "local_theme" || !portableMemberPath(asset.BundlePath) {
		return false
	}
	if asset.MediaType == themeCSSMedia {
		return asset.BundlePath == "theme/style.css"
	}
	if !themeAssetID.MatchString(asset.AssetID) {
		return false
	}
	switch asset.MediaType {
	case "font/woff2":
		return asset.BundlePath == "theme/fonts/"+asset.AssetID+".woff2"
	case "image/png":
		return asset.BundlePath == "theme/images/"+asset.AssetID+".png"
	case "image/svg+xml":
		return asset.BundlePath == "theme/images/"+asset.AssetID+".svg"
	default:
		return false
	}
}

func validateThemeFont(data []byte) error {
	invalid := func() error { return newRenderValidationError("bundle_manifest_invalid", "invalid_asset_media_type") }
	if len(data) < 48 || len(data) > 16<<20 || string(data[:4]) != "wOF2" {
		return invalid()
	}
	u32 := func(at int) uint32 { return binary.BigEndian.Uint32(data[at : at+4]) }
	u16 := func(at int) uint16 { return binary.BigEndian.Uint16(data[at : at+2]) }
	if (u32(4) != 0x00010000 && string(data[4:8]) != "OTTO") || int64(u32(8)) != int64(len(data)) || u16(12) == 0 || u16(12) > 256 || u16(14) != 0 || u32(16) == 0 || u32(16) > 64<<20 || u32(20) == 0 || int64(u32(20)) > int64(len(data)-48) {
		return invalid()
	}
	for _, offset := range []int{28, 32, 36, 40, 44} {
		if u32(offset) != 0 {
			return invalid()
		}
	}
	return nil
}

// This recognizes a small presentation language encoded as CSS. It is not a
// general CSS sanitizer: no escapes, selectors, arbitrary functions or URLs can
// acquire meaning outside the closed values checked here. Accepted bytes are
// retained verbatim and interpreted only inside the renderer sandbox.
func validateTemplateTheme(contract TemplateContract, files []TemplatePackFile) error {
	invalid := func() error { return newRenderValidationError("bundle_manifest_invalid", "invalid_asset_media_type") }
	bySource := map[string][]byte{}
	for _, file := range files {
		bySource[file.SourcePath] = file.Bytes
	}
	fonts, images := map[string]string{}, map[string]string{}
	var css []byte
	count := 0
	for _, asset := range contract.Assets {
		if asset.AssetRole != "local_theme" {
			continue
		}
		count++
		switch asset.MediaType {
		case themeCSSMedia:
			if css != nil {
				return invalid()
			}
			css = bySource[asset.SourcePath]
		case "font/woff2":
			fonts[`"rpt-`+asset.AssetID+`"`] = `url("fonts/` + asset.AssetID + `.woff2") format("woff2")`
		case "image/png", "image/svg+xml":
			images[`url("`+strings.TrimPrefix(asset.BundlePath, "theme/")+`")`] = asset.AssetID
		default:
			return invalid()
		}
	}
	if count == 0 {
		return nil
	}
	if len(contract.SupportedOutputKinds) != 1 || contract.SupportedOutputKinds[0] != OutputKindSlidev || len(css) == 0 || len(css) > 8192 || len(fonts) > 2 || len(images) > 1 {
		return invalid()
	}
	for _, c := range css {
		if c > 126 || c < 32 && c != '\n' && c != '\t' || c == '\\' {
			return invalid()
		}
	}
	if strings.Contains(string(css), "/*") || strings.Contains(string(css), "*/") {
		return invalid()
	}
	remaining := strings.TrimSpace(string(css))
	faces, used := map[string]bool{}, map[string]bool{}
	rootSeen, logoUsed := false, false
	for remaining != "" {
		head, rest, ok := strings.Cut(remaining, "{")
		if !ok {
			return invalid()
		}
		body, tail, ok := strings.Cut(rest, "}")
		if !ok || strings.Contains(body, "{") {
			return invalid()
		}
		remaining = strings.TrimSpace(tail)
		props := map[string]string{}
		body = strings.TrimSpace(body)
		if !strings.HasSuffix(body, ";") {
			return invalid()
		}
		for _, declaration := range strings.Split(strings.TrimSuffix(body, ";"), ";") {
			name, value, ok := strings.Cut(declaration, ":")
			name, value = strings.TrimSpace(name), strings.TrimSpace(value)
			if !ok || name == "" || value == "" || props[name] != "" {
				return invalid()
			}
			props[name] = value
		}
		switch strings.TrimSpace(head) {
		case "@font-face":
			family := props["font-family"]
			if rootSeen || len(props) != 5 || faces[family] || fonts[family] == "" || props["src"] != fonts[family] || props["font-style"] != "normal" || props["font-weight"] != "400" || props["font-display"] != "block" {
				return invalid()
			}
			faces[family] = true
		case ":root":
			if rootSeen {
				return invalid()
			}
			rootSeen = true
			for name, value := range props {
				switch name {
				case "--cartulary-report-background", "--cartulary-report-text", "--cartulary-report-accent":
					if !themeColor.MatchString(value) {
						return invalid()
					}
				case "--cartulary-report-font-family", "--cartulary-report-heading-font-family":
					if value != "sans-serif" && value != "serif" && value != "monospace" {
						if !faces[value] {
							return invalid()
						}
						used[value] = true
					}
				case "--cartulary-report-font-size":
					if !themeFontSize.MatchString(value) {
						return invalid()
					}
				case "--cartulary-report-logo":
					if images[value] == "" {
						return invalid()
					}
					logoUsed = true
				default:
					return invalid()
				}
			}
		default:
			return invalid()
		}
	}
	if !rootSeen || len(used) != len(fonts) || len(faces) != len(fonts) || (len(images) == 1) != logoUsed {
		return invalid()
	}
	return nil
}
