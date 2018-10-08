package m3u8

import (
	"fmt"
)

var (
	ErrNoHeader               = &Error{"missing header tag"}
	ErrMixedTags              = &Error{"playlist contains both master and media tags"}
	ErrUnknownType            = &Error{"failed to determine playlist type"}
	ErrBadSyntax              = &Error{"invalid syntax"}
	ErrBadAttrName            = &Error{"invalid attribute name"}
	ErrBadAttrSyntax          = &Error{"invalid attribute syntax"}
	ErrBadEncryptionMethod    = &Error{"invalid encryption method"}
	ErrBadPlaylistType        = &Error{"invalid playlist type"}
	ErrNoRangeStart           = &Error{"missing range start"}
	ErrNotASegment            = &Error{"not a segment"}
	ErrUnexpectedMediaSegment = &Error{"found media segment after a " + endlistTag + " tag"}
	ErrMissingURI             = &Error{"missing uri"}
	ErrUnexpectedURI          = &Error{"unexpected uri"}
	ErrBadVersionNumber       = &Error{"invalid version number"}
)

type Error struct {
	msg string
}

func (e *Error) Error() string {
	return "m3u8: " + e.msg
}

type missingRequiredAttrError struct {
	attrName string
}

func (e *missingRequiredAttrError) msg() string {
	return `missing required attribute, "` + e.attrName + `",`
}

func (e *missingRequiredAttrError) Error() string {
	return "m3u8: " + e.msg()
}

type invalidAttributeValueError struct {
	attrName string
}

func (e *invalidAttributeValueError) msg() string {
	return `invalid value for attribute, "` + e.attrName + `",`
}

func (e *invalidAttributeValueError) Error() string {
	return "m3u8: " + e.msg()
}

type UnexpectedTagError split

func (e *UnexpectedTagError) Error() string {
	return fmt.Sprintf(`m3u8: unexpected tag in "%s"`, (*split)(e).line())
}

type CompatibilityVersionError struct {
	*split

	version int
}

func (e *CompatibilityVersionError) Error() string {
	return fmt.Sprintf(`m3u8: compatibility version number, %d, required on line %d (%s)`, e.version, e.num, e.line())
}

type InvalidSyntaxError struct {
	*split

	msg string
}

func (e *InvalidSyntaxError) Error() string {
	msg := e.msg
	if msg == "" {
		msg = "invalid syntax"
	}

	return fmt.Sprintf(`m3u8: %s on line %d (%s)`, msg, e.num, e.line())
}

func ise(s *split, msg string) *InvalidSyntaxError {
	return &InvalidSyntaxError{
		split: s,
		msg:   msg,
	}
}

func isew(s *split, err error) error {
	switch err := err.(type) {
	case *missingRequiredAttrError:
		return ise(s, err.msg())

	case *CompatibilityVersionError:
		return &CompatibilityVersionError{s, err.version}

	case *Error:
		return ise(s, err.msg)

	}

	return err
}
