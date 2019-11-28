package m3u8

import (
	"strings"
)

type Stream struct {
	// URI identifies the Media Playlist file.
	//
	// URI is REQUIRED.
	URI string

	// Bandwidth represents the peak segment bit rate of the Variant Stream.
	//
	// Bandwidth is REQUIRED.
	Bandwidth uint64

	// AverageBandwidth represents the average segment bit rate of the Variant
	// Stream.
	//
	// AverageBandwidth is OPTIONAL.
	AverageBandwidth uint64

	// Codecs is a list of formats, where each format specifies a media sample
	// type that is present in one or more Renditions specified by the Variant
	// Stream.
	//
	// Codecs is REQUIRED.
	Codecs []string

	// Width describes the optimal pixel width at which to display all the
	// video in the Variant Stream.
	//
	// Width is OPTIONAL.
	Width uint64

	// Height describes the optimal pixel height at which to display all the
	// video in the Variant Stream.
	//
	// Height is OPTIONAL.
	Height uint64

	// HDCP indicates that the Variant Stream could fail to play unless the
	// output is protected by High-bandwidth Digital Content Protection Type 0
	// or equivalent.
	//
	// See https://tools.ietf.org/html/rfc8216#ref-HDCP.
	//
	// HDCP is OPTIONAL.
	HDCP bool

	// GroupID indicates the set of Renditions that SHOULD be used when playing
	// the presentation.
	//
	// It MUST match the value of the GroupID value of a Rendition of the same
	// type in the Master Playlist.
	//
	// If GroupID is set, it indicates that alternative Renditions of the
	// content are available for playback of that Variant Stream.
	//
	// GroupID is OPTIONAL.
	GroupID string

	// ProgramID uniquely identifies a particular presentation within the scope
	// of the Playlist file.
	//
	// See https://tools.ietf.org/html/draft-pantos-http-live-streaming-11#section-3.4.10.
	//
	// ProgramID was removed in protocol version 6.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.2.1.
	ProgramID uint64
}

func (s *Stream) applyAttributes(attrs attributes) (err error) {
	s.Bandwidth, err = attrs.integer(attrBandwidth)
	if err != nil {
		return err
	}

	s.AverageBandwidth, err = attrs.integer(attrAverageBandwidth)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	codecs, err := attrs.string(attrCodecs)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing {
		s.Codecs = strings.Split(codecs, ",")
	}

	s.Width, s.Height, err = attrs.resolution(attrResolution)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	hdcpLevel, err := attrs.string(attrHDCPLevel)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing {
		switch hdcpLevel {
		case "TYPE-0":
			s.HDCP = true
		case "NONE":
		default:
			return &invalidAttributeValueError{attrHDCPLevel}
		}
	}

	s.ProgramID, err = attrs.integer(attrProgramID)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	return nil
}

func (s *Stream) attrs() (attributes, error) {
	attrs := attributes{
		attrBandwidth: s.Bandwidth,
	}

	if s.AverageBandwidth > 0 {
		attrs[attrAverageBandwidth] = s.AverageBandwidth
	}

	if s.Width > 0 && s.Height > 0 {
		attrs[attrResolution] = decimalResolution{
			w: s.Width,
			h: s.Height,
		}
	}

	if len(s.Codecs) > 0 {
		attrs[attrCodecs] = strings.Join(s.Codecs, ",")
	}

	return attrs, nil
}

// VariantStream represents a Variant Stream, which is a set of Renditions that
// can be combined to play the presentation.
//
// Note: The AUDIO, VIDEO, SUBTITLES and CLOSED-CAPTIONS attributes of the
// EXT-X-STREAM-INF have been replaced with the Type and GroupID fields. Since
// only one of the aforementioned attributes are ever used at once and all are
// used to indicate a group id, it makes more sense to have this kind of data
// structure.
//
// See https://tools.ietf.org/html/rfc8216#section-4.3.4.2.
type VariantStream struct {
	Stream

	// FrameRate describes the maximum frame rate for all the video in the
	// Variant Stream, rounded to three decimal places.
	//
	// FrameRate is OPTIONAL.
	FrameRate float64

	// Type indicates the type of the Variant Stream.
	//
	// Type must be set if GroupID is set.
	Type MediaType
}

func (s *VariantStream) attrs() (attributes, error) {
	attrs, err := s.Stream.attrs()
	if err != nil {
		return nil, err
	}

	if s.FrameRate > 0 {
		attrs[attrFrameRate] = s.FrameRate
	}

	if s.GroupID != "" {
		switch s.Type {
		case Audio:
			attrs[attrAudio] = s.GroupID
		case Video:
			attrs[attrVideo] = s.GroupID
		case Subtitles:
			attrs[attrSubtitles] = s.GroupID
		case ClosedCaptions:
			attrs[attrClosedCaptions] = s.GroupID
		}
	}

	return attrs, nil
}

func parseVariantStream(meta string) (*VariantStream, error) {
	attrs, err := parseAttributeList(meta)
	if err != nil {
		return nil, err
	}

	var vs VariantStream
	if err := vs.applyAttributes(attrs); err != nil {
		return nil, err
	}

	vs.FrameRate, err = attrs.float(attrFrameRate)
	if err != nil && !isMissingAttr(err) {
		return nil, err
	}

	if vs.GroupID, err = attrs.string(attrAudio); err == nil {
		vs.Type = Audio
	} else if !isMissingAttr(err) {
		return nil, err
	} else if vs.GroupID, err = attrs.string(attrVideo); err == nil {
		vs.Type = Video
	} else if !isMissingAttr(err) {
		return nil, err
	} else if vs.GroupID, err = attrs.string(attrSubtitles); err == nil {
		vs.Type = Subtitles
	} else if !isMissingAttr(err) {
		return nil, err
	} else if vs.GroupID, err = attrs.string(attrClosedCaptions); err == nil {
		vs.Type = ClosedCaptions
	} else if !isMissingAttr(err) {
		return nil, err
	}

	return &vs, nil
}

func parseIFrameStream(meta string) (*Stream, error) {
	attrs, err := parseAttributeList(meta)
	if err != nil {
		return nil, err
	}

	var s Stream
	if s.URI, err = attrs.string(attrURI); err != nil {
		return nil, err
	}

	if err := s.applyAttributes(attrs); err != nil {
		return nil, err
	}

	return &s, nil
}
