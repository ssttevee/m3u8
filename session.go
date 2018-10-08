package m3u8

// SessionDataEntry represents the attributes associated with an
// EXT-X-SESSION-DATA tag.
//
// SessionDataEntry MUST contain either a Value or URI value, but nor both.
//
// See https://tools.ietf.org/html/rfc8216#section-4.3.4.4
type SessionDataEntry struct {
	// ID identifies a particular data value.
	//
	// ID is REQUIRED.
	ID string

	// Language is a language tag that identifies the language of Value.
	//
	// See https://tools.ietf.org/html/rfc5646.
	//
	// Language is OPTIONAL.
	Language string

	// Value contains the data identified by ID.
	Value string

	// URI identifies a JSON formatted resource.
	URI string
}

func parseSessionDataEntry(meta string) (*SessionDataEntry, error) {
	attrs, err := parseAttributeList(meta)
	if err != nil {
		return nil, err
	}

	dataID, err := attrs.string(attrDataID)
	if err != nil {
		return nil, err
	}

	language, err := attrs.string(attrLanguage)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	}

	value, err := attrs.string(attrValue)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		return &SessionDataEntry{
			ID:       dataID,
			Language: language,
			Value:    value,
		}, nil
	}

	uri, err := attrs.string(attrURI)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		return &SessionDataEntry{
			ID:       dataID,
			Language: language,
			URI:      uri,
		}, nil
	}

	return nil, &Error{"missing value or uri"}
}

// SessionData represents a set of SessionDataEntry objects and provides
// methods for manipulating them as a set.
type SessionData []*SessionDataEntry

// Entry returns the entry with the same id and language.
//
// Entry returns nil if no matching entry exists.
func (sd SessionData) Entry(id, language string) *SessionDataEntry {
	for _, sde := range sd {
		if sde.ID == id && sde.Language == language {
			return sde
		}
	}

	return nil
}

func (sd *SessionData) getOrCreateEntry(id, language string) *SessionDataEntry {
	sde := sd.Entry(id, language)
	if sde == nil {
		sde = new(SessionDataEntry)
		*sd = append(*sd, sde)
	}

	return sde
}

func (sd *SessionData) SetValue(id, language, value string) {
	sde := sd.getOrCreateEntry(id, language)
	sde.Value = value
	sde.URI = ""
}

func (sd SessionData) Value(id, language string) string {
	if sde := sd.Entry(id, language); sde != nil {
		return sde.Value
	}

	return ""
}

func (sd *SessionData) SetURI(id, language, uri string) {
	sde := sd.getOrCreateEntry(id, language)
	sde.Value = ""
	sde.URI = uri
}

func (sd SessionData) GetURI(id, language string) string {
	if sde := sd.Entry(id, language); sde != nil {
		return sde.URI
	}

	return ""
}
