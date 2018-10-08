package m3u8

import (
	"strconv"
	"strings"
)

// Key represents the attributes associated with an EXT-X-Key tag.
//
// See https://tools.ietf.org/html/rfc8216#section-4.3.2.4.
type Key struct {
	// Method specifies the encryption method.
	//
	// Method is REQUIRED.
	Method EncryptionMethod

	// URI specifies how to obtain the key.
	//
	// URI is REQUIRED unless Method is NoEncryption
	URI string

	// IV specifies a 128-bit unsigned integer Initialization Vector to be used
	// with the key.
	IV *[16]byte

	// KeyFormat specifies how the key is represented in the resource
	// identified by the URI.
	//
	// See https://tools.ietf.org/html/rfc8216#section-5.
	//
	// KeyFormat is OPTIONAL; a zero-value indicates an implicit value of
	// "identity".
	KeyFormat string

	// KeyFormatVersions can be used to indicate which version(s) of a
	// particular Key Format this instance complies with, if there is more than
	// one defined.
	//
	// KeyFormatVersions is OPTIONAL.
	KeyFormatVersions []uint
}

func parseKey(version int, meta string) (*Key, error) {
	attrs, err := parseAttributeList(meta)
	if err != nil {
		return nil, err
	}

	var k Key
	var method string
	method, err = attrs.enum(attrMethod)
	if err != nil {
		return nil, err
	}

	k.Method, err = ParseEncryptionMethod(method)
	if err != nil {
		return nil, &invalidAttributeValueError{attrMethod}
	}

	k.URI, err = attrs.string(attrURI)
	if err != nil && (!isMissingAttr(err) || k.Method != NoEncryption) {
		return nil, err
	}

	var iv []byte
	iv, err = attrs.bytes(attrIV)
	if err != nil && !isMissingAttr(err) {
		return nil, err
	}

	var b [16]byte
	copy(b[:], iv)
	k.IV = &b

	k.KeyFormat, err = attrs.string(attrKeyFormat)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing && version < 5 {
		return nil, &CompatibilityVersionError{version: 5}
	}

	keyFormatVersions, err := attrs.string(attrKeyFormatVersions)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing && version < 5 {
		return nil, &CompatibilityVersionError{version: 5}
	} else if !missing {
		for _, str := range strings.Split(keyFormatVersions, "/") {
			ver, err := strconv.ParseUint(str, 10, 64)
			if err != nil {
				return nil, &invalidAttributeValueError{attrKeyFormatVersions}
			}

			k.KeyFormatVersions = append(k.KeyFormatVersions, uint(ver))
		}
	}

	return &k, nil
}
