package m3u8

import (
	"fmt"
	"io"
	"strconv"
)

type EncryptionMethod int

const (
	NoEncryption EncryptionMethod = iota + 1
	AES128
	SampleAES
)

func (m EncryptionMethod) String() string {
	switch m {
	case NoEncryption:
		return "NONE"
	case AES128:
		return "AES-128"
	case SampleAES:
		return "SAMPLE-AES"
	}

	panic("invalid encryption method")
}

func ParseEncryptionMethod(str string) (EncryptionMethod, error) {
	switch str {
	case "NONE":
		return NoEncryption, nil
	case "AES-128":
		return AES128, nil
	case "SAMPLE-AES":
		return SampleAES, nil
	}

	return 0, ErrBadEncryptionMethod
}

type Map struct {
	URI       string
	ByteRange *ByteRange

	s *split
}

func (m *Map) attrs() (attributes, error) {
	attrs := attributes{
		attrURI: m.URI,
	}

	if m.ByteRange != nil {
		attrs[byterangeTag] = m.ByteRange.String()
	}

	return attrs, nil
}

type PlaylistType int

const (
	Event PlaylistType = iota + 1
	VOD
)

func (t PlaylistType) String() string {
	switch t {
	case Event:
		return "EVENT"
	case VOD:
		return "VOD"
	}

	panic("invalid playlist type")
}

func ParsePlaylistType(str string) (PlaylistType, error) {
	switch str {
	case "EVENT":
		return Event, nil
	case "VOD":
		return VOD, nil
	}

	return 0, ErrBadPlaylistType
}

type MediaPlaylist struct {
	*GenericPlaylist

	Segments []*MediaSegment

	// TargetDuration specifies the maximum Media Segment duration.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.3.1.
	TargetDuration uint64

	// MediaSequence indicates the Media Sequence Number of the first Media
	// Segment that appears in a Playlist file.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.3.2.
	MediaSequence uint64

	// DiscontinuousSequence allows synchronization between different
	// Renditions of the same Variant Stream or different Variant Streams that
	// have Discontinuity == true in their Media Playlists
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.3.3.
	DiscontinuousSequence uint64

	// PlaylistType provides mutability information about the Media Playlist.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.3.5.
	PlaylistType PlaylistType

	// IFramesOnly indicates that each Media Segment in the Playlist describes
	// a single I-frame.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.3.6.
	IFramesOnly bool
}

func parseMediaPlaylist(base *GenericPlaylist, lines []line) (_ *MediaPlaylist, err error) {
	var p MediaPlaylist
	var endlist bool
	for i := 0; i < len(lines); i++ {
		if skip, err := parseMediaSegment(&p, base.Version, lines[i:]); err != nil && err != ErrNotASegment {
			return nil, err
		} else if err == nil {
			if endlist {
				return nil, ErrUnexpectedMediaSegment
			}

			i += skip
			continue
		}

		s := lines[i].(*split)

		switch s.tag {
		case targetdurationTag:
			if !rxDecimalInteger.MatchString(s.meta) {
				return nil, isew(s, ErrBadSyntax)
			}

			p.TargetDuration, _ = strconv.ParseUint(s.meta, 10, 64)

		case mediaSequenceTag:
			if len(p.Segments) > 0 {
				return nil, ise(s, "this tag must appear before the first media segment")
			}

			if !rxDecimalInteger.MatchString(s.meta) {
				return nil, isew(s, ErrBadSyntax)
			}

			p.MediaSequence, _ = strconv.ParseUint(s.meta, 10, 64)

		case discontinuitySequenceTag:
			if len(p.Segments) > 0 {
				return nil, ise(s, "this tag must appear before the first media segment")
			}

			if !rxDecimalInteger.MatchString(s.meta) {
				return nil, isew(s, ErrBadSyntax)
			}

			p.DiscontinuousSequence, _ = strconv.ParseUint(s.meta, 10, 64)

		case endlistTag:
			endlist = true

		case playlistTypeTag:
			p.PlaylistType, err = ParsePlaylistType(s.meta)
			if err != nil {
				return nil, isew(s, err)
			}

		case iFramesOnlyTag:
			p.IFramesOnly = true

		}
	}

	p.GenericPlaylist = base

	return &p, nil
}

func (p *MediaPlaylist) last() *MediaSegment {
	if n := len(p.Segments); n > 0 {
		return p.Segments[n-1]
	}

	return nil
}

func (*MediaPlaylist) Type() Type {
	return Media
}

func (p *MediaPlaylist) encode(w io.Writer) error {
	if err := p.GenericPlaylist.encode(w); err != nil {
		return err
	}

	if _, err := fmt.Fprintf(w, targetdurationTag+":%d\n", p.TargetDuration); err != nil {
		return err
	}

	if p.MediaSequence > 0 {
		if _, err := fmt.Fprintf(w, mediaSequenceTag+":%d\n", p.MediaSequence); err != nil {
			return err
		}
	}

	if p.DiscontinuousSequence > 0 {
		if _, err := fmt.Fprintf(w, discontinuitySequenceTag+":%d\n", p.DiscontinuousSequence); err != nil {
			return err
		}
	}

	if p.PlaylistType != 0 {
		if _, err := fmt.Fprintln(w, playlistTypeTag+":"+p.PlaylistType.String()); err != nil {
			return err
		}
	}

	if p.IFramesOnly {
		if _, err := fmt.Fprintln(w, iFramesOnlyTag); err != nil {
			return err
		}
	}

	if len(p.Segments) > 0 {
		// TODO validate segments

		for _, stream := range p.Segments {
			if err := stream.encode(w); err != nil {
				return err
			}
		}
	}

	if _, err := fmt.Fprintln(w, endlistTag); err != nil {
		return err
	}

	return nil
}
