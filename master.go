package m3u8

import (
	"bytes"
)

type MasterPlaylist struct {
	*GenericPlaylist

	// RenditionMap is used to relate Media Playlists that contain alternative
	// Renditions of the same content.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.4.2.1.
	RenditionMap []Rendition

	// VariantStreams represents a set of Renditions that can be combined to
	// play the presentation.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.4.2.
	VariantStreams []*VariantStream

	// IFrameStreams identifies Media Playlist files containing the I-frames of
	// a multimedia presentation.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.4.3.
	IFrameStreams []*Stream

	// SessionData allows arbitrary session data to be carried in a Master
	// Playlist.
	//
	// SessionData is OPTIONAL.
	SessionData SessionData

	// SessionKeys allows encryption keys from Media Playlists to be specified
	// in a Master Playlist. This allows the client to preload these keys
	// without having to read the Media Playlist(s) first.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.4.5.
	//
	// SessionKeys is OPTIONAL.
	SessionKeys []*Key
}

func parseMasterPlaylist(base *GenericPlaylist, lines []line) (*MasterPlaylist, error) {
	var p MasterPlaylist
	for i := 0; i < len(lines); i++ {
		s, ok := lines[i].(*split)
		if !ok {
			return nil, isew(s, ErrUnexpectedURI)
		}

		switch s.tag {
		case mediaTag:
			rendition, err := parseRendition(s.meta)
			if err != nil {
				return nil, isew(s, err)
			}

			p.RenditionMap = append(p.RenditionMap, rendition)

		case streamInfTag:
			uri, ok := lines[i+1].(uri)
			if !ok {
				return nil, isew(s, ErrMissingURI)
			}

			i++

			vs, err := parseVariantStream(s.meta)
			if err != nil {
				return nil, isew(s, err)
			}

			vs.URI = string(uri)

			p.VariantStreams = append(p.VariantStreams, vs)

		case iFrameStreamInfTag:
			ifs, err := parseIFrameStream(s.meta)
			if err != nil {
				return nil, isew(s, err)
			}

			p.IFrameStreams = append(p.IFrameStreams, ifs)

		case sessionDataTag:
			sde, err := parseSessionDataEntry(s.meta)
			if err != nil {
				return nil, isew(s, err)
			}

			p.SessionData = append(p.SessionData, sde)

		case sessionKeyTag:
			key, err := parseKey(base.Version, s.meta)
			if err != nil {
				return nil, isew(s, err)
			}

			p.SessionKeys = append(p.SessionKeys, key)
		}
	}

	p.GenericPlaylist = base

	return &p, nil
}

func (*MasterPlaylist) Type() Type {
	return Master
}

func (p *MasterPlaylist) encode() ([]byte, error) {
	var buf bytes.Buffer

	if len(p.RenditionMap) > 0 {
		if err := renditions(p.RenditionMap).validate(); err != nil {
			return nil, err
		}

		for _, a := range p.RenditionMap {
			attrs, err := a.attrs()
			if err != nil {
				return nil, err
			}

			encodedAttrs, err := attrs.encode()
			if err != nil {
				return nil, err
			}

			buf.WriteString(mediaTag + ":" + encodedAttrs)
		}
	}

	return buf.Bytes(), nil
}
