package packformat

import (
	"encoding/json"
	"errors"
	"slices"
	"strconv"
	"strings"
)

var licenseIDs, exceptionIDs = loadSPDX()

func loadSPDX() (map[string]bool, map[string]bool) {
	var catalog struct {
		Licenses   []string `json:"license_ids"`
		Exceptions []string `json:"exception_ids"`
	}
	if err := json.Unmarshal(projection("spdx_identifiers.v1.json"), &catalog); err != nil || len(catalog.Licenses) == 0 {
		panic("invalid pinned SPDX identifier projection")
	}
	licenses := map[string]bool{}
	exceptions := map[string]bool{}
	for _, id := range catalog.Licenses {
		licenses[strings.ToLower(id)] = true
	}
	for _, id := range catalog.Exceptions {
		exceptions[strings.ToLower(id)] = true
	}
	return licenses, exceptions
}

type licenseToken struct {
	text       string
	start, end int
}
type licenseParser struct {
	input     string
	tokens    []licenseToken
	index     int
	refs      map[string]bool
	spellings map[string]string
}

func (p *licenseParser) peek() string {
	if p.index >= len(p.tokens) {
		return ""
	}
	return p.tokens[p.index].text
}
func (p *licenseParser) take(operator string) bool {
	if p.peek() == operator || p.peek() == strings.ToLower(operator) {
		p.index++
		return true
	}
	return false
}
func (p *licenseParser) expression() bool {
	if !p.conjunction() {
		return false
	}
	for p.take("OR") {
		if !p.conjunction() {
			return false
		}
	}
	return true
}
func (p *licenseParser) conjunction() bool {
	if !p.primary() {
		return false
	}
	for p.take("AND") {
		if !p.primary() {
			return false
		}
	}
	return true
}
func (p *licenseParser) primary() bool {
	if p.take("(") {
		return p.expression() && p.take(")")
	}
	token := p.peek()
	if token == "" {
		return false
	}
	p.index++
	if strings.HasPrefix(token, "LicenseRef-") {
		id := strings.TrimPrefix(token, "LicenseRef-")
		if !licenseIDString(id) {
			return false
		}
		folded := strings.ToLower(id)
		if previous, ok := p.spellings[folded]; ok && previous != token {
			return false
		}
		p.spellings[folded] = token
		p.refs[token] = true
	} else {
		if !licenseIDs[strings.ToLower(strings.TrimSuffix(token, "+"))] {
			return false
		}
	}
	if p.peek() == "WITH" || p.peek() == "with" {
		op := p.tokens[p.index]
		if op.start == 0 || op.end >= len(p.input) || p.input[op.start-1] != ' ' || p.input[op.end] != ' ' {
			return false
		}
		p.index++
		exception := p.peek()
		if !exceptionIDs[strings.ToLower(exception)] {
			return false
		}
		p.index++
	}
	return true
}
func licenseIDString(value string) bool {
	if value == "" {
		return false
	}
	for _, c := range []byte(value) {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' || c == '.') {
			return false
		}
	}
	return true
}
func parseLicenseExpression(text string) (*licenseParser, bool) {
	if len(text) < 1 || len(text) > 1024 || strings.TrimSpace(text) != text {
		return nil, false
	}
	p := licenseParser{input: text, refs: map[string]bool{}, spellings: map[string]string{}}
	for i := 0; i < len(text); {
		if text[i] == ' ' {
			i++
			continue
		}
		if text[i] < 32 || text[i] > 126 {
			return nil, false
		}
		start := i
		if text[i] == '(' || text[i] == ')' {
			i++
		} else {
			for i < len(text) && text[i] != ' ' && text[i] != '(' && text[i] != ')' {
				if text[i] < 32 || text[i] > 126 {
					return nil, false
				}
				i++
			}
		}
		p.tokens = append(p.tokens, licenseToken{text: text[start:i], start: start, end: i})
	}
	return &p, p.expression() && p.index == len(p.tokens)
}

func licenseFindings(license License, c *semanticChecks) {
	p, valid := parseLicenseExpression(license.Expression)
	c.require(valid, ".expression")
	seen := map[string]bool{}
	previous := ""
	for i, binding := range license.Bindings {
		path := ".license_ref_bindings[" + strconv.Itoa(i) + "]"
		c.require(binding.Ref > previous, path+".license_ref")
		if valid {
			c.require(p.refs[binding.Ref], path+".license_ref")
		}
		c.require(slices.Contains(license.Notices, binding.Path), path+".notice_path")
		seen[binding.Ref] = true
		previous = binding.Ref
	}
	if valid {
		for ref := range p.refs {
			c.require(seen[ref], ".license_ref_bindings")
		}
	}
}
func validLicense(license License) bool {
	c := &semanticChecks{}
	licenseFindings(license, c)
	return !c.invalid
}

func ManifestLicenseFindings(m Manifest, emit FindingSink) error {
	if emit == nil {
		return errors.New("reference pack: missing license findings sink")
	}
	c := &semanticChecks{path: "$.license", emit: emit}
	licenseFindings(m.License, c)
	return c.err
}
