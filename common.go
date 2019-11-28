package m3u8

import (
	"fmt"
	"io"
	"time"
)

// Start represents the attributes associated with an EXT-X-START tag.
//
// See https://tools.ietf.org/html/rfc8216#section-4.3.5.2.
type Start struct {
	// TimeOffset indicates a time offset from the beginning of the
	// Playlist.
	//
	// A negative value indicates a negative time offset from the end of the
	// last Media Segment in the Playlist.
	//
	// TimeOffset is REQUIRED.
	TimeOffset time.Duration

	// Precise indicates whether or not to start playback at the Media Segment
	// containing the Time Offset.
	//
	// Precise is OPTIONAL.
	Precise bool
}

func (s *Start) attrs() (attributes, error) {
	attrs := attributes{
		attrTimeOffset: s.TimeOffset.Seconds(),
	}

	if s.Precise {
		attrs[attrPrecise] = true
	}

	return attrs, nil
}

// GenericPlaylist encompasses the tags described in rfc8216 section 4.3.5.
//
// See https://tools.ietf.org/html/rfc8216#section-4.3.5
type GenericPlaylist struct {
	// IndependentSegments indicates that all media samples in a Media Segment
	// can be decoded without information from other segments.
	IndependentSegments bool

	// Start indicates a preferred point at which to start playing a Playlist.
	//
	// Start is OPTIONAL.
	Start *Start

	Version int
}

func (p *GenericPlaylist) encode(w io.Writer) error {
	if _, err := fmt.Fprintln(w, headerTag); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, versionTag+":%d\n", p.Version); err != nil {
		return err
	}

	if p.IndependentSegments {
		if _, err := fmt.Fprintln(w, independentSegmentsTag); err != nil {
			return err
		}
	}

	if p.Start != nil {
		attrs, err := p.Start.attrs()
		if err != nil {
			return err
		}

		encodedAttrs, err := attrs.encode()
		if err != nil {
			return err
		}

		if _, err := fmt.Fprintln(w, startTag+":"+encodedAttrs); err != nil {
			return err
		}
	}

	return nil
}
