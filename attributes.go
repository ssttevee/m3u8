package m3u8

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

type unsignedFloat float64

func (f unsignedFloat) valid() bool {
	return f >= 0
}

type enumeratedString string

func (s enumeratedString) valid() bool {
	return strings.IndexFunc(string(s), unicode.IsSpace)+strings.IndexAny(string(s), `",`) == -2
}

func isBadQuotedStringChar(r rune) bool {
	switch r {
	case '\n', '\r', '"':
		return false
	default:
		return true
	}
}

type decimalResolution struct {
	w uint64
	h uint64
}

const (
	typeDecimalInteger             = "decimal-integer"
	typeHexadecimalSequence        = "hexadecimal-sequence"
	typeDecimalFloatingPoint       = "decimal-floating-point"
	typeSignedDecimalFloatingPoint = "signed-decimal-floating-point"
	typeQuotedString               = "quoted-string"
	typeEnumeratedString           = "enumerated-string"
	typeDecimalResolution          = "decimal-resolution"
)

func valueType(value interface{}) string {
	switch value.(type) {
	case uint64:
		return typeDecimalInteger
	case []byte:
		return typeHexadecimalSequence
	case unsignedFloat:
		return typeDecimalFloatingPoint
	case float64:
		return typeSignedDecimalFloatingPoint
	case string:
		return typeQuotedString
	case enumeratedString:
		return typeEnumeratedString
	case decimalResolution:
		return typeDecimalResolution
	default:
		return fmt.Sprintf("unsupported type (%T)", value)
	}
}

func attrError(expectedType, attrName string, value interface{}) *Error {
	return &Error{`expected ` + expectedType + ` for attribute, "` + attrName + `", but got ` + valueType(value)}
}

type attributes map[string]interface{}

func (a attributes) encode() (string, error) {
	pairs := make([]string, 0, len(a))
	for name, value := range a {
		if value == nil {
			continue
		}

		var encoded string
		switch v := value.(type) {
		case uint64:
			encoded = strconv.FormatUint(v, 10)

		case []byte:
			encoded = "0x" + hex.EncodeToString(v)

		case unsignedFloat:
			if !v.valid() {
				return "", &Error{fmt.Sprintf(`illegal %s: %g`, typeDecimalFloatingPoint, v)}
			}

			encoded = strconv.FormatFloat(float64(v), 'f', -1, 64)

		case float64:
			encoded = strconv.FormatFloat(float64(v), 'f', -1, 64)

		case string:
			if strings.IndexFunc(v, isBadQuotedStringChar) == -1 {
				return "", &Error{fmt.Sprintf(`illegal %s: %s`, typeQuotedString, v)}
			}

			encoded = `"` + v + `"`

		case enumeratedString:
			if !v.valid() {
				return "", &Error{fmt.Sprintf(`illegal %s: %s`, typeEnumeratedString, v)}
			}

			encoded = string(v)

		case decimalResolution:
			encoded = strconv.FormatUint(v.w, 10) + "x" + strconv.FormatUint(v.h, 10)

		default:
			return "", &Error{fmt.Sprintf(`unexpected attribute value type: %T`, v)}
		}

		pairs = append(pairs, name+"="+encoded)
	}

	return strings.Join(pairs, ","), nil
}

func (a attributes) has(name string) bool {
	_, ok := a[name]
	return ok
}

func (a attributes) value(name string) (interface{}, error) {
	v, ok := a[name]
	if !ok {
		return nil, &missingRequiredAttrError{name}
	}

	return v, nil
}

func (a attributes) integer(name string) (uint64, error) {
	v, err := a.value(name)
	if err != nil {
		return 0, err
	}

	if n, ok := v.(uint64); ok {
		return n, nil
	}

	return 0, attrError(typeDecimalInteger, name, v)
}

func (a attributes) bytes(name string) ([]byte, error) {
	v, err := a.value(name)
	if err != nil {
		return nil, err
	}

	if data, ok := v.([]byte); ok {
		return data, nil
	}

	return nil, attrError(typeHexadecimalSequence, name, v)
}

func (a attributes) float(name string) (float64, error) {
	v, err := a.value(name)
	if err != nil {
		return 0, err
	}

	if f, ok := v.(unsignedFloat); ok {
		return float64(f), nil
	}

	return 0, attrError(typeDecimalFloatingPoint, name, v)
}

func (a attributes) signedFloat(name string) (float64, error) {
	v, err := a.value(name)
	if err != nil {
		return 0, err
	}

	if f, ok := v.(unsignedFloat); ok {
		return float64(f), nil
	} else if f, ok := v.(float64); ok {
		return f, nil
	}

	return 0, attrError(typeSignedDecimalFloatingPoint, name, v)
}

func (a attributes) string(name string) (string, error) {
	v, err := a.value(name)
	if err != nil {
		return "", err
	}

	if str, ok := v.(string); ok {
		return str, nil
	}

	return "", attrError(typeQuotedString, name, v)
}

func (a attributes) enum(name string) (string, error) {
	v, err := a.value(name)
	if err != nil {
		return "", err
	}

	if str, ok := v.(enumeratedString); ok {
		return string(str), nil
	}

	return "", attrError(typeEnumeratedString, name, v)
}

func (a attributes) resolution(name string) (w uint64, h uint64, _ error) {
	v, err := a.value(name)
	if err != nil {
		return 0, 0, err
	}

	if res, ok := v.(decimalResolution); ok {
		return res.w, res.h, nil
	}

	return 0, 0, attrError(typeDecimalResolution, name, v)
}

var (
	rxAttributeName              = regexp.MustCompile(`^([A-Z0-9-]+)=`)
	rxDecimalInteger             = regexp.MustCompile(`^(\d{1,20})(?:,|$)`)
	rxHexadecimalSequence        = regexp.MustCompile(`^(0[xX][0-9A-F]*)(?:,|$)`)
	rxDecimalFloatingPoint       = regexp.MustCompile(`^(\d+(?:\.\d+)?)(?:,|$)`)
	rxSignedDecimalFloatingPoint = regexp.MustCompile(`^(-?\d+(?:\.\d+)?)(?:,|$)`)
	rxQuotedString               = regexp.MustCompile(`^("[^\n\r"]*")(?:,|$)`)
	rxEnumeratedString           = regexp.MustCompile(`^([^\s",]*)(?:,|$)`)
	rxDecimalResolution          = regexp.MustCompile(`^(\d+x\d+)(?:,|$)`)
)

func parseAttributeList(str string) (attributes, error) {
	attrs := make(map[string]interface{})
	for pos := 0; pos < len(str); {
		name := rxAttributeName.FindString(str[pos:])
		if name == "" {
			return nil, ErrBadAttrName
		}

		pos += len(name)

		name = name[:len(name)-1]

		var matches []string
		if matches = rxDecimalInteger.FindStringSubmatch(str[pos:]); matches != nil {
			n, _ := strconv.ParseUint(matches[1], 10, 64)
			attrs[name] = n
		} else if matches = rxHexadecimalSequence.FindStringSubmatch(str[pos:]); matches != nil {
			chars := matches[1][2:]
			if len(chars)%2 == 1 {
				chars = "0" + chars
			}

			attrs[name], _ = hex.DecodeString(chars)
		} else if matches = rxDecimalFloatingPoint.FindStringSubmatch(str[pos:]); matches != nil {
			n, _ := strconv.ParseFloat(matches[1], 64)
			attrs[name] = unsignedFloat(n)
		} else if matches = rxSignedDecimalFloatingPoint.FindStringSubmatch(str[pos:]); matches != nil {
			attrs[name], _ = strconv.ParseFloat(matches[1], 64)
		} else if matches = rxQuotedString.FindStringSubmatch(str[pos:]); matches != nil {
			attrs[name] = matches[1][1 : len(matches[1])-1]
		} else if matches = rxEnumeratedString.FindStringSubmatch(str[pos:]); matches != nil {
			attrs[name] = enumeratedString(matches[1])
		} else if matches = rxDecimalResolution.FindStringSubmatch(str[pos:]); matches != nil {
			x := strings.IndexRune(matches[1], 'x')
			w, _ := strconv.ParseUint(matches[1][:x], 10, 64)
			h, _ := strconv.ParseUint(matches[1][x+1:], 10, 64)
			attrs[name] = decimalResolution{w, h}
		} else {
			return nil, ErrBadAttrSyntax
		}

		pos += len(matches[0])
	}

	return attrs, nil
}

func isMissingAttr(err error) bool {
	_, ok := err.(*missingRequiredAttrError)
	return ok
}
