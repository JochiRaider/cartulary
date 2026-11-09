package schema

import (
	"encoding/json"
	"strconv"
	"strings"
)

func integer(value any, min, max int64) (int64, bool) {
	n, ok := value.(json.Number)
	if !ok {
		return 0, false
	}
	text := n.String()
	if text == "" || strings.ContainsAny(text, ".eE+") || (len(text) > 1 && text[0] == '0') || text == "-0" {
		return 0, false
	}
	v, err := strconv.ParseInt(text, 10, 64)
	return v, err == nil && v >= min && v <= max
}
