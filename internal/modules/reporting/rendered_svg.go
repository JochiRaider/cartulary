package reporting

import (
	"bytes"
	"encoding/xml"
	"io"
	"math"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

const svgNamespace = "http://www.w3.org/2000/svg"
const xmlNamespace = "http://www.w3.org/XML/1998/namespace"

// Validate the bytes that will enter the bundle, without repairing them. In
// particular, a renderer cannot launder active content by calling it an image.
func validateRenderedSVG(data []byte) error {
	invalid := func() error {
		return newRenderValidationError("security_sandbox_violation", "external_reference_in_rendered_output")
	}
	if !utf8.Valid(data) {
		return invalid()
	}
	decoder := xml.NewDecoder(bytes.NewReader(data))
	// No custom entities or charset reader: only UTF-8 and XML's predefined
	// entities are admitted. Directives and processing instructions are rejected.
	stack := []string{}
	seenRoot := false
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			if !seenRoot || len(stack) != 0 {
				return invalid()
			}
			return nil
		}
		if err != nil {
			return invalid()
		}
		switch node := token.(type) {
		case xml.StartElement:
			if node.Name.Space != svgNamespace || !svgElementAllowed(node.Name.Local) {
				return invalid()
			}
			if len(stack) == 0 {
				if seenRoot || node.Name.Local != "svg" {
					return invalid()
				}
				seenRoot = true
			}
			attributes := map[xml.Name]bool{}
			for _, attr := range node.Attr {
				if attributes[attr.Name] || !svgAttributeAllowed(node.Name.Local, attr) {
					return invalid()
				}
				attributes[attr.Name] = true
			}
			stack = append(stack, node.Name.Local)
		case xml.EndElement:
			if len(stack) == 0 {
				return invalid()
			}
			stack = stack[:len(stack)-1]
		case xml.CharData:
			text := string(node)
			if len(stack) > 0 && svgTextElement(stack[len(stack)-1]) {
				if svgControls(text) {
					return invalid()
				}
				continue
			}
			if strings.Trim(text, " \t\r\n") == "" {
				continue
			}
			return invalid()
		case xml.Comment, xml.Directive, xml.ProcInst:
			return invalid()
		default:
			return invalid()
		}
	}
}

func svgElementAllowed(name string) bool {
	switch name {
	case "svg", "g", "defs", "linearGradient", "radialGradient", "stop", "rect", "circle", "ellipse", "line", "polyline", "polygon", "path", "text", "tspan", "title", "desc":
		return true
	}
	return false
}

func svgTextElement(name string) bool {
	return name == "text" || name == "tspan" || name == "title" || name == "desc"
}

func svgControls(value string) bool {
	for _, ch := range value {
		if ch < 0x20 || (ch >= 0x7f && ch <= 0x9f) {
			return true
		}
	}
	return false
}

var svgNumber = regexp.MustCompile(`^[+-]?(?:[0-9]+(?:\.[0-9]*)?|\.[0-9]+)(?:[eE][+-]?[0-9]+)?`)
var svgFontFamily = regexp.MustCompile(`^[\pL\pN _,'"-]+$`)
var svgPaint = regexp.MustCompile(`^(?:[A-Za-z]+|#[0-9a-fA-F]{3}|#[0-9a-fA-F]{4}|#[0-9a-fA-F]{6}|#[0-9a-fA-F]{8})$`)

func svgAttributeAllowed(element string, attr xml.Attr) bool {
	if attr.Name.Space == "" && attr.Name.Local == "xmlns" {
		return attr.Value == svgNamespace
	}
	if attr.Name.Space == "xmlns" {
		if attr.Name.Local == "xml" {
			return attr.Value == xmlNamespace
		}
		return attr.Name.Local != "xmlns" && attr.Value == svgNamespace
	}
	if attr.Name.Space == xmlNamespace {
		return attr.Name.Local == "space" && svgTextElement(element) && (attr.Value == "default" || attr.Value == "preserve")
	}
	value := attr.Value
	lower := strings.ToLower(value)
	if attr.Name.Space != "" || svgControls(value) || strings.ContainsAny(value, ":\\") ||
		strings.Contains(lower, "url(") || strings.Contains(lower, "@import") || strings.Contains(value, "--") {
		return false
	}
	switch attr.Name.Local {
	case "x", "y", "dx", "dy", "x1", "y1", "x2", "y2", "cx", "cy", "r", "rx", "ry", "width", "height", "stroke-width", "font-size", "offset":
		value = strings.Trim(value, " ")
		if strings.HasSuffix(value, "%") {
			value = strings.TrimSuffix(value, "%")
		} else {
			value = strings.TrimSuffix(value, "px")
		}
		if strings.HasSuffix(value, " ") {
			return false
		}
		numbers, ok := svgNumbers(value)
		if !ok || len(numbers) != 1 {
			return false
		}
		switch attr.Name.Local {
		case "r", "rx", "ry", "width", "height", "stroke-width", "font-size":
			return numbers[0] >= 0
		}
		return true
	case "opacity", "fill-opacity", "stroke-opacity":
		numbers, ok := svgNumbers(value)
		return ok && len(numbers) == 1 && numbers[0] >= 0 && numbers[0] <= 1
	case "viewBox":
		numbers, ok := svgNumbers(value)
		return ok && len(numbers) == 4 && numbers[2] >= 0 && numbers[3] >= 0
	case "points":
		numbers, ok := svgNumbers(value)
		return ok && len(numbers) >= 2 && len(numbers)%2 == 0
	case "d":
		return svgPath(value)
	case "transform":
		return svgTransform(value)
	case "fill", "stroke":
		return svgPaint.MatchString(value)
	case "font-family":
		return svgFontFamily.MatchString(value)
	case "font-weight":
		if value == "normal" || value == "bold" || value == "bolder" || value == "lighter" {
			return true
		}
		numbers, ok := svgNumbers(value)
		return ok && len(numbers) == 1 && numbers[0] >= 1 && numbers[0] <= 1000
	case "text-anchor":
		return value == "start" || value == "middle" || value == "end"
	case "dominant-baseline":
		switch value {
		case "auto", "text-bottom", "alphabetic", "ideographic", "middle", "central", "mathematical", "hanging", "text-top":
			return true
		}
	}
	return false
}

// Read SVG numbers, including implicit separators before a sign or dot. Never
// admit NaN/Infinity or overflow that the browser could interpret differently.
func svgNumbers(value string) ([]float64, bool) {
	result := []float64{}
	value = strings.Trim(value, " ")
	for value != "" {
		token := svgNumber.FindString(value)
		if token == "" {
			return nil, false
		}
		number, err := strconv.ParseFloat(token, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return nil, false
		}
		result = append(result, number)
		value = strings.TrimLeft(value[len(token):], " ")
		if strings.HasPrefix(value, ",") {
			value = strings.TrimLeft(value[1:], " ")
			if value == "" {
				return nil, false
			}
		}
	}
	return result, len(result) > 0
}

func svgTransform(value string) bool {
	value = strings.Trim(value, " ")
	if value == "" {
		return false
	}
	for value != "" {
		open, close := strings.IndexByte(value, '('), strings.IndexByte(value, ')')
		if open < 1 || close < open {
			return false
		}
		name := strings.Trim(value[:open], " ")
		numbers, ok := svgNumbers(value[open+1 : close])
		if !ok {
			return false
		}
		switch name {
		case "matrix":
			if len(numbers) != 6 {
				return false
			}
		case "translate", "scale":
			if len(numbers) < 1 || len(numbers) > 2 {
				return false
			}
		case "rotate":
			if len(numbers) != 1 && len(numbers) != 3 {
				return false
			}
		case "skewX", "skewY":
			if len(numbers) != 1 {
				return false
			}
		default:
			return false
		}
		value = strings.Trim(value[close+1:], " ")
		if strings.HasPrefix(value, ",") {
			value = strings.Trim(value[1:], " ")
			if value == "" {
				return false
			}
		}
	}
	return true
}

func svgPath(value string) bool {
	value = strings.Trim(value, " ")
	if value == "" || (value[0] != 'M' && value[0] != 'm') {
		return false
	}
	for value != "" {
		command := value[0]
		if command >= 'a' && command <= 'z' {
			command -= 'a' - 'A'
		}
		arity := 0
		switch command {
		case 'M', 'L', 'T':
			arity = 2
		case 'H', 'V':
			arity = 1
		case 'C':
			arity = 6
		case 'S', 'Q':
			arity = 4
		case 'A':
			arity = 7
		case 'Z':
			value = strings.Trim(value[1:], " ")
			continue
		default:
			return false
		}
		value = value[1:]
		end := strings.IndexAny(value, "MmLlHhVvCcSsQqTtAaZz")
		if end < 0 {
			end = len(value)
		}
		var numbers []float64
		var ok bool
		if command == 'A' {
			numbers, ok = svgArcNumbers(value[:end])
		} else {
			numbers, ok = svgNumbers(value[:end])
		}
		if !ok || len(numbers)%arity != 0 {
			return false
		}
		if command == 'A' {
			for i := 0; i < len(numbers); i += 7 {
				if numbers[i] < 0 || numbers[i+1] < 0 || (numbers[i+3] != 0 && numbers[i+3] != 1) || (numbers[i+4] != 0 && numbers[i+4] != 1) {
					return false
				}
			}
		}
		value = strings.Trim(value[end:], " ")
	}
	return true
}

// Arc flags are single characters in the SVG grammar. Parsing them as ordinary
// floats would admit invalid forms such as 1.0 and reject compact valid flags.
func svgArcNumbers(value string) ([]float64, bool) {
	values := []float64{}
	value = strings.Trim(value, " ")
	for value != "" {
		index := len(values) % 7
		token := svgNumber.FindString(value)
		if index == 3 || index == 4 {
			if value[0] != '0' && value[0] != '1' {
				return nil, false
			}
			token = value[:1]
		}
		if token == "" {
			return nil, false
		}
		number, err := strconv.ParseFloat(token, 64)
		if err != nil || math.IsNaN(number) || math.IsInf(number, 0) {
			return nil, false
		}
		values = append(values, number)
		value = strings.TrimLeft(value[len(token):], " ")
		if strings.HasPrefix(value, ",") {
			value = strings.TrimLeft(value[1:], " ")
			if value == "" {
				return nil, false
			}
		}
	}
	return values, len(values) > 0
}
