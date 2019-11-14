package m3u8

import "io"

type Encoder struct {
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (e *Encoder) Encode(plist Playlist) error {
	return plist.encode(e.w)
}
