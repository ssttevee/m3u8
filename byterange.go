package m3u8

import (
	"strconv"
	"strings"
)

type ByteRange struct {
	Start  int64
	Length int64
}

func parseByteRange(str string) (_ *ByteRange, err error) {
	var br ByteRange

	lenStr := str

	if at := strings.IndexRune(str, '@'); at != -1 {
		start, err := strconv.ParseUint(str[at+1:], 10, 64)
		if err != nil {
			return nil, &Error{"failed to parse range start"}
		}

		br.Start = int64(start)

		lenStr = str[:at]
	} else {
		br.Start = -1
	}

	length, err := strconv.ParseUint(lenStr, 10, 64)
	if err != nil {
		return nil, &Error{"failed to parse range length"}
	}

	br.Length = int64(length)

	if lenStr == str {
		err = ErrNoRangeStart
	}

	return &br, err
}

func (r ByteRange) closed() bool {
	return r.Length > 0 && r.Start >= 0
}

func (r ByteRange) String() string {
	if r.Start < 0 {
		return strconv.FormatInt(r.Length, 10)
	}

	return strconv.FormatInt(r.Length, 10) + "@" + strconv.FormatInt(r.Start, 10)
}
