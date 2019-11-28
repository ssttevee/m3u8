package m3u8

import (
	"strings"
	"time"
)

type DateRange struct {
	// ID uniquely identifies a Date Range in the Playlist.
	//
	// ID is REQUIRED.
	ID string

	// Class specifies some set of attributes and their associated value
	// semantics. All Date Ranges with the same Class value MUST adhere to
	// these semantics.
	//
	// Class is REQUIRED.
	Class string

	// StartDate is a ISO-8601 formatted date at which the Date Range begins.
	//
	// StartDate is REQUIRED.
	StartDate string

	// EndDate is a ISO-8601 formatted date at which the Date Range ends.
	//
	// It MUST be equal to or later than the value of StartDate.
	//
	// EndDate is OPTIONAL.
	EndDate string

	// Duration is the duration of the Date Range.
	//
	// Duration is OPTIONAL.
	Duration time.Duration

	// PlannedDuration is the expected duration of the Date Range.
	//
	// PlannedDuration is OPTIONAL.
	PlannedDuration time.Duration

	// ClientAttributes are client-defined attributes. The map keys may only
	// use uppercase alphanumeric characters and hyphens. The map values may
	// only be of type string, []byte or float64.
	//
	// ClientAttributes are OPTIONAL.
	ClientAttributes map[string]interface{}

	// SCTE35Command is the big-endian binary representation of the
	// splice_info_section().
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.2.7.1.
	SCTE35Command []byte

	// SCTE35Out is the big-endian binary representation of the "out"
	// splice_info_section().
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.2.7.1.
	SCTE35Out []byte

	// SCTE35In is the big-endian binary representation of the "in"
	// splice_info_section().
	//
	// See https://tools.ietf.org/html/rfc8216#section-4.3.2.7.1.
	SCTE35In []byte

	// EndOnNext indicates that the end of the range is equal to the StartDate
	// of the Following Range. The Following Range is the Date Range with the
	// same Class value that has the earliest StartDate after the StartDate of
	// the range in question.
	//
	// EndOnNext is OPTIONAL.
	EndOnNext bool
}

func parseDateRange(meta string) (*DateRange, error) {
	attrs, err := parseAttributeList(meta)
	if err != nil {
		return nil, err
	}

	var dr DateRange
	dr.ID, err = attrs.string(attrID)
	if err != nil {
		return nil, err
	}

	dr.Class, err = attrs.string(attrClass)
	if err != nil && !isMissingAttr(err) {
		return nil, err
	}

	dr.StartDate, err = attrs.string(attrStartDate)
	if err != nil {
		return nil, err
	}

	if err = validateDate(dr.StartDate); err != nil {
		return nil, err
	}

	dr.EndDate, err = attrs.string(attrEndDate)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		if err = validateDate(dr.EndDate); err != nil {
			return nil, err
		}
	}

	var duration float64
	duration, err = attrs.float(attrDuration)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		dr.Duration = secondsToDuration(duration)
	}

	var plannedDuration float64
	plannedDuration, err = attrs.float(attrPlannedDuration)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		dr.PlannedDuration = secondsToDuration(plannedDuration)
	}

	clientAttributes := make(map[string]interface{})
	for name, value := range attrs {
		if name[:2] != "X-" {
			continue
		}

		switch v := value.(type) {
		case unsignedFloat:
			clientAttributes[name] = float64(v)

		case string, []byte:
			clientAttributes[name] = value

		default:
			return nil, &Error{"illegal client attribute value"}
		}
	}

	if len(clientAttributes) > 0 {
		dr.ClientAttributes = clientAttributes
	}

	var scte35Cmd, scte35Out, scte35In []byte
	scte35Cmd, err = attrs.bytes(attrSCTE35Command)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		dr.SCTE35Command = scte35Cmd
	}

	scte35Out, err = attrs.bytes(attrSCTE35Out)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		dr.SCTE35Out = scte35Out
	}

	scte35In, err = attrs.bytes(attrSCTE35In)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		dr.SCTE35In = scte35In
	}

	var endOnNext string
	endOnNext, err = attrs.enum(attrEndOnNext)
	if missing := isMissingAttr(err); err != nil && !missing {
		return nil, err
	} else if !missing {
		if endOnNext != "YES" {
			return nil, &invalidAttributeValueError{attrEndOnNext}
		}

		if dr.Class == "" {
			return nil, &Error{`this tag must have attribute, "` + attrClass + `", with attribute, "` + attrEndOnNext + `",`}
		}

		if dr.Duration != 0 {
			return nil, &Error{`this tag may not have attribute, "` + attrDuration + `", with attribute, "` + attrEndOnNext + `",`}
		}

		if dr.EndDate != "" {
			return nil, &Error{`this tag may not have attribute, "` + attrEndDate + `", with attribute, "` + attrEndOnNext + `",`}
		}
	}

	return &dr, nil
}

func (r DateRange) attrs() (attributes, error) {
	attrs := attributes{
		attrID:        r.ID,
		attrStartDate: r.StartDate,
	}

	if r.Class != "" {
		attrs[attrClass] = r.Class
	}

	if r.EndDate != "" {
		attrs[attrEndDate] = r.EndDate
	}

	if r.Duration > 0 {
		attrs[attrDuration] = unsignedFloat(r.Duration.Seconds())
	}

	if r.PlannedDuration > 0 {
		attrs[attrPlannedDuration] = unsignedFloat(r.PlannedDuration.Seconds())
	}

	if len(r.ClientAttributes) > 0 {
		for name, value := range r.ClientAttributes {
			attrs[strings.ToUpper(name)] = value
		}
	}

	if len(r.SCTE35Command) > 0 {
		attrs[attrSCTE35Command] = r.SCTE35Command
	}

	if len(r.SCTE35Out) > 0 {
		attrs[attrSCTE35Out] = r.SCTE35Out
	}

	if len(r.SCTE35In) > 0 {
		attrs[attrSCTE35In] = r.SCTE35In
	}

	if r.EndOnNext {
		attrs[attrEndOnNext] = enumeratedString("YES")
	}

	return attrs, nil
}
