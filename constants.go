package m3u8

const (
	tagPrefix = "#EXT"

	// basic tags
	headerTag  = tagPrefix + "M3U"
	versionTag = tagPrefix + "-X-VERSION"

	// media segment tags
	infTag             = tagPrefix + "INF"
	byterangeTag       = tagPrefix + "-X-BYTERANGE"
	discontinuityTag   = tagPrefix + "-X-DISCONTINUITY"
	keyTag             = tagPrefix + "-X-KEY"
	mapTag             = tagPrefix + "-X-MAP"
	programDateTimeTag = tagPrefix + "-X-PROGRAM-DATE-TIME"
	daterangeTag       = tagPrefix + "-X-DATERANGE"

	// media playlist tags
	targetdurationTag        = tagPrefix + "-X-TARGETDURATION"
	mediaSequenceTag         = tagPrefix + "-X-MEDIA-SEQUENCE"
	discontinuitySequenceTag = tagPrefix + "-X-DISCONTINUITY-SEQUENCE"
	endlistTag               = tagPrefix + "-X-ENDLIST"
	playlistTypeTag          = tagPrefix + "-X-PLAYLIST-TYPE"
	iFramesOnlyTag           = tagPrefix + "-X-I-FRAMES-ONLY"

	// master playlist tags
	mediaTag           = tagPrefix + "-X-MEDIA"
	streamInfTag       = tagPrefix + "-X-STREAM-INF"
	iFrameStreamInfTag = tagPrefix + "-X-I-FRAME-STREAM-INF"
	sessionDataTag     = tagPrefix + "-X-SESSION-DATA"
	sessionKeyTag      = tagPrefix + "-X-SESSION-KEY"

	// media or master playlist tags
	independentSegmentsTag = tagPrefix + "-X-INDEPENDENT-SEGMENTS"
	startTag               = tagPrefix + "-X-START"
)

const (
	attrAssocLanguage     = "ASSOC-LANGUAGE"
	attrAudio             = "AUDIO"
	attrAutoselect        = "AUTOSELECT"
	attrAverageBandwidth  = "AVERAGE-BANDWIDTH"
	attrBandwidth         = "BANDWIDTH"
	attrByteRange         = "BYTERANGE"
	attrClass             = "CLASS"
	attrCharacteristics   = "CHARACTERISTICS"
	attrChannels          = "CHANNELS"
	attrClosedCaptions    = "CLOSED-CAPTIONS"
	attrCodecs            = "CODECS"
	attrDataID            = "DATA-ID"
	attrDefault           = "DEFAULT"
	attrDuration          = "DURATION"
	attrEndDate           = "END-DATE"
	attrEndOnNext         = "END-ON-NEXT"
	attrForced            = "FORCED"
	attrFrameRate         = "FRAME-RATE"
	attrGroupID           = "GROUP-ID"
	attrHDCPLevel         = "HDCP-LEVEL"
	attrID                = "ID"
	attrInstreamID        = "INSTREAM-ID"
	attrIV                = "IV"
	attrKeyFormat         = "KEYFORMAT"
	attrKeyFormatVersions = "KEYFORMATVERSIONS"
	attrLanguage          = "LANGUAGE"
	attrMethod            = "METHOD"
	attrName              = "NAME"
	attrPlannedDuration   = "PLANNED-DURATION"
	attrPrecise           = "PRECISE"
	attrProgramID         = "PROGRAM-ID"
	attrResolution        = "RESOLUTION"
	attrSCTE35Command     = "SCTE35-CMD"
	attrSCTE35In          = "SCTE35-IN"
	attrSCTE35Out         = "SCTE35-OUT"
	attrStartDate         = "START-DATE"
	attrSubtitles         = "SUBTITLES"
	attrTimeOffset        = "TIME-OFFSET"
	attrType              = "TYPE"
	attrURI               = "URI"
	attrValue             = "VALUE"
	attrVideo             = "VIDEO"
)
