package m3u8

import (
	"strconv"
	"strings"
)

type MediaType int

const (
	Audio MediaType = iota + 1
	Video
	Subtitles
	ClosedCaptions
)

type Rendition interface {
	Type() MediaType
	applyAttrs(attributes) error

	attrs() (attributes, error)

	groupID() string
	name() string
	isDefault() bool
	isAutoSelect() bool
}

type BasicRendition struct {
	// GroupID specifies the group to which the Rendition belongs.
	//
	// GroupID is REQUIRED.
	GroupID string

	// Language is a language tag that identifies the primary language used in
	// the Rendition.
	//
	// See https://tools.ietf.org/html/rfc5646.
	//
	// Language is OPTIONAL.
	Language string

	// AssociatedLanguage is a language tag that identifies a language that is
	// associated with the Rendition.
	//
	// See https://tools.ietf.org/html/rfc5646.
	//
	// AssociatedLanguage is OPTIONAL.
	AssociatedLanguage string

	// Name is a human-readable description of the Rendition.
	//
	// If the Language value is set, then this description SHOULD be in that
	// language.
	//
	// Name is REQUIRED.
	Name string

	// Default indicates whether or not to play this Rendition of the content
	// in the absence of information from the user indicating a different
	// choice.
	//
	// Default is OPTIONAL.
	Default bool

	// AutoSelect indicates whether or not to choose to play this Rendition in
	// the absence of explicit user preference because it matches the current
	// playback environment, such as chosen system language.
	//
	// Default is OPTIONAL.
	AutoSelect bool

	// Characteristics is a list of Uniform Type Identifiers.
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.4.1.
	//
	// Characteristics is OPTIONAL.
	Characteristics []string
}

func (a *BasicRendition) applyAttrs(attrs attributes) (err error) {
	a.GroupID, err = attrs.string(attrGroupID)
	if err != nil {
		return err
	}

	a.Language, err = attrs.string(attrLanguage)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	a.AssociatedLanguage, err = attrs.string(attrAssocLanguage)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	a.Name, err = attrs.string(attrName)
	if err != nil {
		return err
	}

	defaultStr, err := attrs.enum(attrDefault)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing {
		switch defaultStr {
		case "YES":
			a.Default = true
		case "NO":
		default:
			return &invalidAttributeValueError{attrDefault}
		}
	}

	autoselectStr, err := attrs.enum(attrAutoselect)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing {
		switch autoselectStr {
		case "YES":
			a.AutoSelect = true
		case "NO":
		default:
			return &invalidAttributeValueError{attrAutoselect}
		}
	}

	characteristicsStr, err := attrs.string(attrCharacteristics)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing {
		a.Characteristics = strings.Split(characteristicsStr, ",")
	}

	return nil
}

func (a BasicRendition) attrs() (attributes, error) {
	// initialize capacity for:
	// - GROUP-ID
	// - LANGUAGE
	// - ASSOC-LANGUAGE
	// - NAME
	// - DEFAULT
	// - AUTOSELECT
	// - CHARACTERISTICS
	// - TYPE
	if a.GroupID == "" {
		return nil, &missingRequiredAttrError{attrGroupID}
	}

	if a.Name == "" {
		return nil, &missingRequiredAttrError{attrName}
	}

	attrs := map[string]interface{}{
		attrGroupID: a.GroupID,
		attrName:    a.Name,
	}

	if a.Language != "" {
		attrs[attrLanguage] = a.Language
	}

	if a.AssociatedLanguage != "" {
		attrs[attrAssocLanguage] = a.AssociatedLanguage
	}

	if a.Default {
		attrs[attrDefault] = enumeratedString("YES")
	}

	if a.AutoSelect {
		attrs[attrAutoselect] = enumeratedString("YES")
	}

	if len(a.Characteristics) > 0 {
		for _, characteristic := range a.Characteristics {
			if strings.IndexRune(characteristic, ',') != -1 {
				return nil, &Error{"characteristic may not contain a comma"}
			}
		}

		attrs[attrCharacteristics] = strings.Join(a.Characteristics, ",")
	}

	return attrs, nil
}

func (a BasicRendition) groupID() string {
	return a.GroupID
}

func (a BasicRendition) name() string {
	return a.Name
}

func (a BasicRendition) isDefault() bool {
	return a.Default
}

func (a BasicRendition) isAutoSelect() bool {
	return a.AutoSelect
}

type AudioRendition struct {
	BasicRendition

	// URI identifies the Media Playlist file.
	//
	// URI is OPTIONAL.
	URI string

	// Channels is a count of audio channels indicating the maximum number of
	// independent, simultaneous audio channels present.
	Channels int
}

func (a *AudioRendition) applyAttrs(attrs attributes) (err error) {
	if err = a.BasicRendition.applyAttrs(attrs); err != nil {
		return err
	}

	a.URI, err = attrs.string(attrURI)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	channelsStr, err := attrs.string(attrChannels)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing {
		num, err := strconv.ParseInt(channelsStr, 10, 64)
		if err != nil {
			return &invalidAttributeValueError{attrChannels}
		}

		a.Channels = int(num)
	}

	return nil
}

func (a *AudioRendition) Type() MediaType {
	return Audio
}

func (a *AudioRendition) attrs() (attributes, error) {
	attrs, err := a.BasicRendition.attrs()
	if err != nil {
		return nil, err
	}

	if a.URI != "" {
		attrs[attrURI] = a.URI
	}

	if a.Channels != 0 {
		attrs[attrChannels] = strconv.Itoa(a.Channels)
	}

	return attrs, nil
}

type VideoRendition struct {
	BasicRendition

	// URI identifies the Media Playlist file.
	//
	// URI is OPTIONAL.
	URI string
}

func (a *VideoRendition) applyAttrs(attrs attributes) (err error) {
	if err = a.BasicRendition.applyAttrs(attrs); err != nil {
		return err
	}

	a.URI, err = attrs.string(attrURI)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	return nil
}

func (a *VideoRendition) Type() MediaType {
	return Video
}

func (a *VideoRendition) attrs() (attributes, error) {
	attrs, err := a.BasicRendition.attrs()
	if err != nil {
		return nil, err
	}

	if a.URI != "" {
		attrs[attrURI] = a.URI
	}

	return attrs, nil
}

type SubtitlesRendition struct {
	BasicRendition

	// URI identifies the Media Playlist file.
	//
	// URI is OPTIONAL.
	URI string

	// Forced indicates that the subtitles are considered essential to play.
	//
	// If Forced is false, the Rendition contains content that is intended to
	// be played in response to explicit user request.
	//
	// Default is OPTIONAL.
	Forced bool
}

func (a *SubtitlesRendition) applyAttrs(attrs attributes) (err error) {
	if err = a.BasicRendition.applyAttrs(attrs); err != nil {
		return err
	}

	a.URI, err = attrs.string(attrURI)
	if err != nil && !isMissingAttr(err) {
		return err
	}

	forcedStr, err := attrs.enum(attrForced)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing {
		switch forcedStr {
		case "YES":
			a.Forced = true
		case "NO":
		default:
			return &invalidAttributeValueError{attrForced}
		}
	}

	return nil
}

func (a *SubtitlesRendition) Type() MediaType {
	return Subtitles
}

func (a *SubtitlesRendition) attrs() (attributes, error) {
	attrs, err := a.BasicRendition.attrs()
	if err != nil {
		return nil, err
	}

	if a.URI != "" {
		attrs[attrURI] = a.URI
	}

	if a.Forced {
		attrs[attrForced] = enumeratedString("YES")
	}

	return attrs, nil
}

type ClosedCaptionsRendition struct {
	BasicRendition

	// InstreamID specifies a Rendition within the segments in the Media
	// Playlist.
	//
	// Value must be one of "CC1", "CC2", "CC3", "CC4", or "SERVICEn" where n
	// MUST be an integer between 1 and 63 (e.g., "SERVICE3" or "SERVICE42").
	//
	// InstreamID is REQUIRED.
	InstreamID string
}

func (a *ClosedCaptionsRendition) applyAttrs(attrs attributes) (err error) {
	if err = a.BasicRendition.applyAttrs(attrs); err != nil {
		return err
	}

	a.InstreamID, err = attrs.enum(attrInstreamID)
	if missing := isMissingAttr(err); err != nil && !missing {
		return err
	} else if !missing && !isValidInstreamID(a.InstreamID) {
		return &invalidAttributeValueError{attrInstreamID}
	}

	return nil
}

func (a *ClosedCaptionsRendition) Type() MediaType {
	return ClosedCaptions
}

func (a *ClosedCaptionsRendition) attrs() (attributes, error) {
	attrs, err := a.BasicRendition.attrs()
	if err != nil {
		return nil, err
	}

	if a.InstreamID == "" {
		if !isValidInstreamID(a.InstreamID) {
			return nil, &invalidAttributeValueError{attrInstreamID}
		}

		attrs[attrURI] = a.InstreamID
	}

	return attrs, nil
}

func isValidInstreamID(str string) bool {
	if strings.HasPrefix(str, "CC") {
		if n, err := strconv.ParseInt(str[2:], 10, 64); err == nil && n >= 1 && n <= 4 {
			return true
		}
	} else if strings.HasPrefix(str, "SERVICE") {
		if n, err := strconv.ParseInt(str[7:], 10, 64); err == nil && n >= 1 && n <= 63 {
			return true
		}
	}

	return false
}

type renditions []Rendition

func (as renditions) validate() error {
	if len(as) == 0 {
		return nil
	}

	// TODO validate rendition groups: https://tools.ietf.org/html/rfc8216#section-4.3.4.1.1

	type renditionGroup struct {
		hasDefault bool
		names      map[string]interface{}
	}

	typeGroups := map[MediaType]map[string]renditionGroup{}
	for _, a := range as {
		t, gid, n, d := a.Type(), a.groupID(), a.name(), a.isDefault()
		groups, ok := typeGroups[t]
		if !ok {
			groups = map[string]renditionGroup{}
			typeGroups[t] = groups
		}

		group, ok := groups[gid]
		if !ok {
			group = renditionGroup{
				names: map[string]interface{}{},
			}

			groups[gid] = group
		}

		if _, ok := group.names[n]; !ok {
			if group.hasDefault && d {
				return &Error{"a rendition group must not have more than one default"}
			}

			group.names[n] = nil
			group.hasDefault = d
		} else {
			return &Error{"all renditions in the same group must have different names"}
		}
	}

	return nil
}

func parseRendition(meta string) (Rendition, error) {
	attrs, err := parseAttributeList(meta)
	if err != nil {
		return nil, err
	}

	renditionType, err := attrs.enum(attrType)
	if err != nil {
		return nil, err
	}

	var rendition Rendition
	switch renditionType {
	case "AUDIO":
		rendition = new(AudioRendition)
	case "VIDEO":
		rendition = new(VideoRendition)
	case "SUBTITLES":
		rendition = new(SubtitlesRendition)
	case "CLOSED-CAPTIONS":
		rendition = new(ClosedCaptionsRendition)
	default:
		return nil, &invalidAttributeValueError{attrType}
	}

	if err = rendition.applyAttrs(attrs); err != nil {
		return nil, err
	}

	return rendition, nil
}
