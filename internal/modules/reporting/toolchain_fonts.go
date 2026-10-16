package reporting

import (
	"errors"
	"path"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/renderprocess"
)

// The package store can expose webfonts independently of fontconfig. Include
// every installed font and stylesheet, even when this particular deck does not
// reference it, rather than claiming the system font directory is complete.
func installedToolchainFontFiles(files []renderprocess.InstallationFile) []ToolchainFontFile {
	fonts := []ToolchainFontFile{}
	for _, file := range files {
		include := strings.HasPrefix(file.Path, "runtime/fonts/")
		if strings.HasPrefix(file.Path, "runtime/renderer/") {
			switch strings.ToLower(path.Ext(file.Path)) {
			case ".css", ".woff", ".woff2", ".ttf", ".otf", ".ttc", ".otc", ".eot":
				include = true
			}
		}
		if include {
			fonts = append(fonts, ToolchainFontFile{Path: "installed/" + file.Path, SHA256: file.SHA256})
		}
	}
	sort.Slice(fonts, func(i, j int) bool { return fonts[i].Path < fonts[j].Path })
	return fonts
}

func templateFontFiles(files []RenderBundleFile) []ToolchainFontFile {
	fonts := []ToolchainFontFile{}
	for _, file := range files {
		if file.Role == renderBundleRoleLocalTheme && (file.MediaType == themeCSSMedia || file.MediaType == "font/woff2") {
			fonts = append(fonts, ToolchainFontFile{Path: "template/" + file.Path, SHA256: file.SHA256})
		}
	}
	sort.Slice(fonts, func(i, j int) bool { return fonts[i].Path < fonts[j].Path })
	return fonts
}

func validateToolchainFonts(snapshot ToolchainSnapshot, files []RenderBundleFile) error {
	invalid := errors.New("reporting: invalid usable font inventory")
	if len(snapshot.FontFiles) == 0 || len(snapshot.FontFiles) > 1024 {
		return invalid
	}
	previous := ""
	installed := 0
	declared := []ToolchainFontFile{}
	for _, font := range snapshot.FontFiles {
		if !portableMemberPath(font.Path) || font.Path <= previous || !sha256HexPattern.MatchString(font.SHA256) {
			return invalid
		}
		previous = font.Path
		if strings.HasPrefix(font.Path, "installed/") {
			installed++
		} else if strings.HasPrefix(font.Path, "template/theme/") {
			declared = append(declared, font)
		} else {
			return invalid
		}
	}
	raw, err := canonicalJSON(snapshot.FontFiles)
	if err != nil || installed == 0 || hashHex(raw) != snapshot.FontManifestSHA256 || !sameReportingValue(declared, templateFontFiles(files)) {
		return invalid
	}
	return nil
}

func bindTemplateFonts(snapshot *ToolchainSnapshot, files []RenderBundleFile) error {
	if err := validateToolchainFonts(*snapshot, nil); err != nil {
		return err
	}
	snapshot.FontFiles = append(append([]ToolchainFontFile{}, snapshot.FontFiles...), templateFontFiles(files)...)
	sort.Slice(snapshot.FontFiles, func(i, j int) bool { return snapshot.FontFiles[i].Path < snapshot.FontFiles[j].Path })
	raw, err := canonicalJSON(snapshot.FontFiles)
	if err != nil {
		return err
	}
	snapshot.FontManifestSHA256 = hashHex(raw)
	return nil
}
