package m3u8

import (
	"strings"
	"time"
)

// MediaSegment represents a Media Segment.
//
// See https://tools.ietf.org/html/rfc8216#section-3 for information about
// Media Segments.
//
// See https://tools.ietf.org/html/rfc8216#section-4.3.2 for information about
// Media Segment Tags.
type MediaSegment struct {
	// URI is the resource of the Media Segment.
	URI string

	// Duration specifies the duration of the Media Segment.
	Duration time.Duration

	// Title is an optional human-readable informative title of the Media
	// Segment.
	Title string

	// ByteRange indicates that a Media Segment is a sub-range of the resource
	// identified by the URI value. It applies only to this Media Segment.
	//
	// A Media Segment without a ByteRange value consists of the entire
	// resource identified by the URI value.
	//
	// Use of ByteRange REQUIRES a compatibility version number of 4 or
	// greater.
	ByteRange *ByteRange

	// Discontinuity indicates a discontinuity between the Media Segment that
	// follows it and the one that preceded it.
	//
	// Discontinuity MUST be true if there is a change in any of the following
	// characteristics:
	//
	// - file format
	//
	// - number, type, and identifiers of tracks
	//
	// - timestamp sequence
	//
	// - encoding parameters
	//
	// - encoding sequence
	Discontinuity bool

	// Key specifies how to decrypt encrypted Media Segments. It applies to
	// every Media Segment and to every Media Initialization Section declared
	// by a non-nil Map value that appears between it and the next non-nil Key
	// value in the Playlist file with the same KeyFormat value (or the end of
	// the Playlist file).
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.2.4.
	Key *Key

	// Map specifies how to obtain the Media Initialization Section required to
	// parse the applicable Media Segments. It applies to every Media Segment
	// that appears after it in the Playlist until the next Media Segment with
	// a non-nil Map value or until the end of the Playlist.
	Map *Map

	// ProgramDateTime associates the first sample of a Media Segment with an
	// absolute date and/or time. It applies only to the next Media Segment.
	//
	// NOTE: This is left as a string because there is no easy way to parse the
	// expansive variations of ISO 8601:2004 formats and section 4.3.2.6 of rfc
	// 8216 does not require a specific accuracy.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.2.6.
	ProgramDateTime string

	// DateRange associates a Date Range (i.e., a range of time defined by a
	// starting and ending date) with a set of properties.
	DateRange *DateRange
}

func parseMediaSegment(p *MediaPlaylist, version int, lines []line) (skip int, err error) {
	var segment MediaSegment

LinesLoop:
	for i, line := range lines {
		if uri, ok := line.(uri); ok {
			if i == 0 {
				return 0, ErrUnexpectedURI
			}

			segment.URI = string(uri)
			p.Segments = append(p.Segments, &segment)

			return i, nil
		}

		s := line.(*split)

		var attrs attributes
		switch s.tag {
		case infTag:
			comma := strings.IndexRune(s.meta, ',')
			if comma == -1 {
				return 0, ise(s, "missing comma")
			}

			duration := s.meta[:comma]
			segment.Title = s.meta[comma+1:]

			if version < 3 {
				if decimal := strings.IndexRune(duration, '.'); decimal != -1 {
					return 0, &CompatibilityVersionError{s, 3}
				}
			}

			segment.Duration, err = parseDuration(duration)
			if err != nil {
				return 0, isew(s, err)
			}

		case byterangeTag:
			if version < 4 {
				return 0, &CompatibilityVersionError{s, 4}
			}

			segment.ByteRange, err = parseByteRange(s.meta)
			if err != nil && (err != ErrNoRangeStart || !p.last().hasDependableRange()) {
				return 0, isew(s, err)
			}

		case discontinuityTag:
			segment.Discontinuity = true

		case keyTag:
			segment.Key, err = parseKey(version, s.meta)
			if err != nil {
				return 0, isew(s, err)
			}

		case mapTag:
			attrs, err = parseAttributeList(s.meta)
			if err != nil {
				return 0, isew(s, err)
			}

			var m Map
			m.URI, err = attrs.string(attrURI)
			if err != nil {
				return 0, isew(s, err)
			}

			var strbr string
			strbr, err = attrs.string(attrByteRange)
			if err != nil && !isMissingAttr(err) {
				return 0, isew(s, err)
			}

			m.ByteRange, err = parseByteRange(strbr)
			if err != nil {
				return 0, isew(s, err)
			}

			m.s = s

		case programDateTimeTag:
			if err = validateDate(s.meta); err != nil {
				return 0, isew(s, err)
			}

			segment.ProgramDateTime = s.meta

		case daterangeTag:
			segment.DateRange, err = parseDateRange(s.meta)
			if err != nil {
				return 0, isew(s, err)
			}
		default:
			if i == 0 {
				break LinesLoop
			}

			return 0, &Error{msg: `unexpected tag, "` + s.tag + `"`}
		}
	}

	return 0, ErrNotASegment
}

func (s *MediaSegment) hasDependableRange() bool {
	if s == nil {
		return false
	}

	if s.ByteRange == nil {
		return false
	}

	return s.ByteRange.closed()
}
