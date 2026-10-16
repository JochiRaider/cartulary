// Package reportingassets supplies inert licensed bytes for renderer and
// application fixtures. It owns no product registry or rendering capability.
package reportingassets

import _ "embed"

//go:embed AtkinsonHyperlegible-Regular.woff2
var themeFont []byte

func ThemeFont() []byte { return append([]byte(nil), themeFont...) }
