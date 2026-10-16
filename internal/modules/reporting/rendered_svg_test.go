package reporting

import (
	"errors"
	"strings"
	"testing"
)

func TestRenderedSVGClosedOutputBoundary(t *testing.T) {
	wrap := func(body string) []byte {
		return []byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 800 600">` + body + `</svg>`)
	}
	for _, body := range []string{
		`<g transform="translate(10,20) scale(.5) rotate(30 2 3) matrix(1 0 0 1 0 0)"><rect x="-1" y="2" width="100%" height="24px" fill="#aabbcc" opacity="0.5"/></g>`,
		`<path d="M0 0L10-20 30.5.5h4v5C1 2 3 4 5 6s1 2 3 4Q1 2 3 4t5 6A10 20 30 0 1 50 60z" stroke="black" fill="none"/>`,
		`<path d="M0 0A10 20 30 0150 60z"/>`,
		`<text x="10" y="20" font-family="Noto Sans, sans-serif" font-size="14" font-weight="600" text-anchor="middle" dominant-baseline="central" xml:space="preserve">Inert url(https://example.invalid) &amp; &lt;script&gt;<tspan>安全</tspan></text>`,
		`<defs><linearGradient><stop offset="0%"/></linearGradient><radialGradient/></defs><circle cx="3" cy="4" r="2"/><ellipse rx="2" ry="3"/><line x1="0" y1="1" x2="2" y2="3"/><polyline points="0,0 1,1"/><polygon points="0 0 1 1 2 0"/><title>Redacted label</title><desc>Safe description</desc>`,
	} {
		if err := validateRenderedSVG(wrap(body)); err != nil {
			t.Fatalf("valid rendered subset rejected: %v", err)
		}
	}
	bad := []string{
		`<script>alert(1)</script>`, `<foreignObject/>`, `<image/>`, `<use/>`, `<animate/>`, `<unknown/>`,
		`<g onload="alert(1)"/>`, `<g OnLoad="alert(1)"/>`, `<g style="fill:red"/>`,
		`<g href="#local"/>`, `<g src="https://example.invalid"/>`, `<g xml:base="file:///"/>`,
		`<text xmlns:a="https://example.invalid" a:label="x"/>`, `<text xmlns="https://example.invalid">x</text>`,
		`<text xmlns:xml="http://www.w3.org/2000/svg"/>`, `<text xmlns:xmlns="http://www.w3.org/2000/svg"/>`,
		`<text fill="u&#114;l(#x)">x</text>`, `<text fill="URL(#x)">x</text>`, `<text fill="red;stroke:blue">x</text>`,
		`<text fill="var(--color)">x</text>`, `<text font-family="@import(x)">x</text>`,
		`<rect width="NaN"/>`, `<rect width="1e999"/>`, `<rect width="-1"/>`, `<rect width="1px%"/>`,
		`<rect width="1 px"/>`, `<rect width="&#x2003;1"/>`, `<rect width="1&#x2003;"/>`,
		`<rect width="1" width="2"/>`, `<rect width="&#10;2"/>`, `<g xml:space="preserve"/>`,
		`<path d="M0 0 L1"/>`, `<path d="L0 0"/>`, `<path d="M0 0X1 2"/>`, `<path d="M0 0 A1 1 0 2 0 1 1"/>`,
		`<path d="M0 0A10 20 30 1.0 0 50 60"/>`,
		`<path d="M0 0L1e999 2"/>`, `<path d="M0 0,"/>`, `<path d="M0 0Z1 2"/>`,
		`<g transform="translate(1),"/>`, `<g transform="matrix(1 2)"/>`, `<g transform="rotate(NaN)"/>`,
		`<g transform="translate(1) garbage"/>`, `<svg viewBox="0 0 1"/>`, `<polyline points="0 0 1"/>`,
		`<!-- comment -->`, `<?instruction value?>`, `<!DOCTYPE svg>`, `<text>&unknown;</text>`,
		`<text>&#x85;</text>`, `<text>&#x1;</text>`, `<g>Unexpected non-text content</g>`,
		"<text>A<![CDATA[\n]]>B</text>", `<title>A&#10;B</title>`,
	}
	for _, body := range bad {
		t.Run(body, func(t *testing.T) {
			err := validateBundleFilesSelfContained([]RenderBundleFile{newBundleFile("diagrams/d.svg", renderBundleRoleRenderedSVG, "image/svg+xml", wrap(body), true)})
			var failure *renderValidationError
			if !errors.As(err, &failure) || failure.FailureCode != "security_sandbox_violation" || failure.ReasonCode != "external_reference_in_rendered_output" {
				t.Fatalf("unsafe output must fail with registered bounded reason, got %v", err)
			}
			if strings.Contains(err.Error(), body) {
				t.Fatal("diagnostic must not expose SVG content")
			}
		})
	}
}

func TestRenderedSVGRejectsMalformedDocuments(t *testing.T) {
	for _, data := range [][]byte{
		nil, []byte(`<svg/>`), []byte(`<g xmlns="http://www.w3.org/2000/svg"/>`),
		[]byte(`<svg xmlns="http://www.w3.org/2000/svg"/><svg xmlns="http://www.w3.org/2000/svg"/>`),
		[]byte(`<svg xmlns="http://www.w3.org/2000/svg"><g></svg>`),
		[]byte(`<?xml version="1.0"?><svg xmlns="http://www.w3.org/2000/svg"/>`),
		[]byte(`<!DOCTYPE svg [<!ENTITY e SYSTEM "file:///etc/passwd">]><svg xmlns="http://www.w3.org/2000/svg"><text>&e;</text></svg>`),
		[]byte("<svg xmlns=\"http://www.w3.org/2000/svg\"><text>\xff</text></svg>"),
	} {
		if err := validateRenderedSVG(data); err == nil {
			t.Fatal("malformed SVG accepted")
		}
	}
}
