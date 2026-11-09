package producer

import (
	"archive/zip"
	"bytes"
	"errors"
	"fmt"
	"io"
	"path"
	"strconv"
	"strings"
)

// OOXML is read as data only. No relationship is fetched and no formula, macro,
// embedded object, or image is executed. Bounds cover expanded bytes, not only
// the compressed input; duplicate names cannot change locator interpretation.
func windowsOfficeMembers(b []byte) (map[string][]byte, error) {
	z, e := zip.NewReader(bytes.NewReader(b), int64(len(b)))
	if e != nil {
		return nil, e
	}
	if len(z.File) > 2048 {
		return nil, errors.New("OOXML member count exceeded")
	}
	out := map[string][]byte{}
	var total uint64
	for _, f := range z.File {
		name := f.Name
		if name == "" || strings.Contains(name, "\\") || strings.HasPrefix(name, "/") || path.Clean(name) != name || strings.HasPrefix(name, "../") || !f.Mode().IsRegular() || f.Flags&1 != 0 {
			return nil, errors.New("unsafe OOXML member")
		}
		if _, ok := out[name]; ok {
			return nil, errors.New("duplicate OOXML member")
		}
		total += f.UncompressedSize64
		if f.UncompressedSize64 > uint64(windowsArtifactLimit) || total > 256<<20 {
			return nil, errors.New("OOXML expansion budget")
		}
		r, e := f.Open()
		if e != nil {
			return nil, e
		}
		raw, e := io.ReadAll(io.LimitReader(r, windowsArtifactLimit+1))
		ce := r.Close()
		if e != nil {
			return nil, e
		}
		if ce != nil {
			return nil, ce
		}
		if uint64(len(raw)) != f.UncompressedSize64 {
			return nil, errors.New("OOXML member size differs")
		}
		if strings.HasSuffix(name, ".xml") || strings.HasSuffix(name, ".rels") {
			if bytes.Contains(bytes.ToUpper(raw), []byte("<!DOCTYPE")) || bytes.Contains(bytes.ToUpper(raw), []byte("<!ENTITY")) {
				return nil, errors.New("OOXML DTD or entity prohibited")
			}
		}
		out[name] = raw
	}
	return out, nil
}

func officeText(n *windowsXML) string {
	if n == nil {
		return ""
	}
	switch n.Name {
	case "t":
		return n.Text
	case "tab":
		return "\t"
	case "br", "cr":
		return "\n"
	}
	var b strings.Builder
	for _, c := range n.Children {
		b.WriteString(officeText(c))
	}
	if n.Name == "p" {
		b.WriteByte('\n')
	}
	return b.String()
}

type windowsWorksheet struct {
	Cells map[string]string
	Rows  map[int]string
	Part  string
}
type windowsOffice struct {
	Members map[string][]byte
	Body    []*windowsXML
	Sheets  map[string]windowsWorksheet
}

func readWindowsOffice(kind string, b []byte) (windowsOffice, error) {
	out := windowsOffice{Sheets: map[string]windowsWorksheet{}}
	var e error
	out.Members, e = windowsOfficeMembers(b)
	if e != nil {
		return out, e
	}
	if kind == "original_docx" {
		doc, e := parseWindowsXML(out.Members["word/document.xml"])
		if e != nil {
			return out, e
		}
		body := xmlOne(doc, "body")
		if body == nil {
			return out, errors.New("word body missing")
		}
		out.Body = body.Children
		return out, nil
	}
	book, e := parseWindowsXML(out.Members["xl/workbook.xml"])
	if e != nil {
		return out, e
	}
	rels, e := parseWindowsXML(out.Members["xl/_rels/workbook.xml.rels"])
	if e != nil {
		return out, e
	}
	targets := map[string]string{}
	for _, rel := range xmlChildren(rels, "Relationship") {
		if rel.Attr["TargetMode"] == "External" {
			continue
		}
		target := path.Join("xl", rel.Attr["Target"])
		if strings.HasPrefix(rel.Attr["Target"], "/") {
			target = strings.TrimPrefix(rel.Attr["Target"], "/")
		}
		if _, ok := out.Members[target]; ok {
			targets[rel.Attr["Id"]] = target
		}
	}
	shared := []string{}
	if raw, ok := out.Members["xl/sharedStrings.xml"]; ok {
		sst, e := parseWindowsXML(raw)
		if e != nil {
			return out, e
		}
		for _, si := range xmlChildren(sst, "si") {
			shared = append(shared, officeText(si))
		}
	}
	for _, sheet := range xmlDescendants(book, "sheet") {
		part := targets[sheet.Attr["id"]]
		raw, ok := out.Members[part]
		if !ok {
			return out, errors.New("worksheet relationship missing")
		}
		x, e := parseWindowsXML(raw)
		if e != nil {
			return out, e
		}
		s := windowsWorksheet{Cells: map[string]string{}, Rows: map[int]string{}, Part: part}
		for _, row := range xmlDescendants(x, "row") {
			n, e := strconv.Atoi(row.Attr["r"])
			if e != nil || n < 1 {
				return out, errors.New("invalid worksheet row")
			}
			if _, ok := s.Rows[n]; ok {
				return out, errors.New("duplicate worksheet row")
			}
			s.Rows[n] = row.Raw
			for _, cell := range xmlChildren(row, "c") {
				if xmlOne(cell, "f") != nil {
					return out, errors.New("formula is not a literal source value")
				}
				value := ""
				if v := xmlOne(cell, "v"); v != nil {
					value = v.Text
				}
				if cell.Attr["t"] == "s" {
					i, e := strconv.Atoi(value)
					if e != nil || i < 0 || i >= len(shared) {
						return out, errors.New("invalid shared string")
					}
					value = shared[i]
				}
				if cell.Attr["t"] == "inlineStr" {
					value = officeText(xmlOne(cell, "is"))
				}
				if _, ok := s.Cells[cell.Attr["r"]]; ok {
					return out, errors.New("duplicate worksheet cell")
				}
				s.Cells[cell.Attr["r"]] = value
			}
		}
		if _, ok := out.Sheets[sheet.Attr["name"]]; ok {
			return out, errors.New("duplicate worksheet name")
		}
		out.Sheets[sheet.Attr["name"]] = s
	}
	return out, nil
}
func (o windowsOffice) fragment(start, end int) (string, string, error) {
	if start < 1 || end < start || end > len(o.Body) {
		return "", "", errors.New("word body range out of bounds")
	}
	var raw, text strings.Builder
	for _, n := range o.Body[start-1 : end] {
		raw.WriteString(n.Raw)
		text.WriteString(officeText(n))
	}
	return raw.String(), text.String(), nil
}
func officeRangeLocator(start, end int) string {
	return fmt.Sprintf("ooxml/word/document.xml/body/%d/%d", start, end)
}

// The supplied extraction renders table cells with vertical bars and separates
// paragraphs within a cell with slashes. Reproduce that documented presentation
// from cell structure; do not erase punctuation from source claims to make them
// match. The structured table audit independently checks every original cell.
func officeReferenceText(nodes []*windowsXML) string {
	parts := []string{}
	for _, n := range nodes {
		if n.Name != "tbl" {
			parts = append(parts, strings.TrimSpace(officeText(n)))
			continue
		}
		for _, row := range xmlChildren(n, "tr") {
			cells := []string{}
			for _, cell := range xmlChildren(row, "tc") {
				paragraphs := []string{}
				for _, p := range xmlChildren(cell, "p") {
					paragraphs = append(paragraphs, strings.TrimSpace(officeText(p)))
				}
				cells = append(cells, strings.Join(paragraphs, " / "))
			}
			parts = append(parts, strings.Join(cells, " | "))
		}
	}
	return strings.Join(parts, "\n")
}
