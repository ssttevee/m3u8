package m3u8

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Decoder struct {
	r      io.Reader
	Strict bool
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r:      r,
		Strict: true,
	}
}

func (d *Decoder) Decode() (Playlist, error) {
	scanner := bufio.NewScanner(d.r)
	if !scanner.Scan() {
		return nil, io.ErrUnexpectedEOF
	}

	if firstLine := scanner.Text(); firstLine != headerTag {
		fmt.Println(firstLine)
		return nil, ErrNoHeader
	}

	p, err := decode(scanner, d.Strict)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func DecodePlaylist(data []byte) (Playlist, error) {
	return NewDecoder(bytes.NewBuffer(data)).Decode()
}

type line interface {
	line() string
}

type uri string

func (u uri) line() string {
	return string(u)
}

type split struct {
	num  int
	tag  string
	meta string
}

func (s *split) line() string {
	if s.meta == "" {
		return s.tag
	}

	return s.tag + ":" + s.meta
}

func secondsToDuration(s float64) time.Duration {
	return time.Duration(s * float64(time.Second))
}

func parseDuration(str string) (time.Duration, error) {
	seconds, err := strconv.ParseFloat(str, 64)
	if err != nil {
		return 0, &Error{msg: "failed to parse duration"}
	}

	return secondsToDuration(seconds), nil
}

var rxISO8601 = regexp.MustCompile("^[+-]?\\d{4,}(?:-?(?:\\d{2}(?:-?\\d{2})?|W\\d{2}(?:-?\\d)?|\\d{3}))?(?:T\\d{2}(?::?\\d{2}(?::?\\d{2}(?:\\.\\d+)?)?)?(?:Z|[+-]\\d{2}(?::?\\d{2})?)?)?$")

func validateDate(str string) error {
	if rxISO8601.MatchString(str) {
		return nil
	}

	return &Error{msg: "invalid date format"}
}

// decode determines the playlist type, parses common tags, and buffers
// important lines for further processing.
func decode(scanner *bufio.Scanner, strict bool) (Playlist, error) {
	var pType Type
	var version int
	var lines []line

	var base GenericPlaylist

	for lineNumber := 2; scanner.Scan(); lineNumber++ {
		line := scanner.Text()
		if len(line) == 0 {
			// ignore blank lines
			continue
		}

		if line[0] != '#' {
			// this is a url line
			lines = append(lines, uri(line))
			continue
		}

		if line[:4] != tagPrefix {
			// ignore comments
			continue
		}

		var s split
		if colon := strings.IndexRune(line, ':'); colon >= 0 {
			s.tag = line[:colon]
			s.meta = line[colon+1:]
		} else {
			s.tag = line
		}

		switch s.tag {
		case versionTag:
			num, err := strconv.ParseInt(s.meta, 0, 64)
			if err != nil {
				return nil, ErrBadVersionNumber
			}

			version = int(num)
			continue

		case infTag, byterangeTag, discontinuityTag, keyTag, mapTag, programDateTimeTag, daterangeTag:
			// media segment tags
			fallthrough

		case targetdurationTag, mediaSequenceTag, discontinuitySequenceTag, endlistTag, playlistTypeTag, iFramesOnlyTag:
			// media playlist tags
			if pType == 0 {
				pType = Media
			}

			if pType != Media {
				return nil, ErrMixedTags
			}

		case mediaTag, streamInfTag, iFrameStreamInfTag, sessionDataTag, sessionKeyTag:
			// master playlist tags
			if pType == 0 {
				pType = Master
			}

			if pType != Master {
				return nil, ErrMixedTags
			}

		case independentSegmentsTag:
			base.IndependentSegments = true

		case startTag:
			attrs, err := parseAttributeList(s.meta)
			if err != nil {
				return nil, isew(&s, err)
			}

			var start Start
			timeOffset, err := attrs.float(attrTimeOffset)
			if err != nil {
				return nil, isew(&s, err)
			}

			start.TimeOffset = secondsToDuration(timeOffset)

			precise, err := attrs.enum(attrPrecise)
			if missing := isMissingAttr(err); err != nil && !missing {
				return nil, isew(&s, err)
			} else if !missing {
				switch precise {
				case "YES":
					start.Precise = true
				case "NO":
				default:
					return nil, &invalidAttributeValueError{attrPrecise}
				}
			}

		default:
			if strict {
				return nil, (*UnexpectedTagError)(&s)
			}
		}

		s.num = lineNumber
		lines = append(lines, &s)
	}

	switch pType {
	case Media:
		p, err := parseMediaPlaylist(version, lines)
		if err != nil {
			return nil, err
		}

		p.GenericPlaylist = base
		p.Version = version

		return p, nil

	case Master:
		p, err := parseMasterPlaylist(version, lines)
		if err != nil {
			return nil, err
		}

		p.GenericPlaylist = base
		p.Version = version

		return p, nil

	default:
		return nil, ErrUnknownType
	}
}
