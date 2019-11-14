// Package m3u8 provides abstractions to work with Apple HLS playlist files.
package m3u8

import "io"

type Type int

const (
	Master Type = iota + 1
	Media
)

type Playlist interface {
	Type() Type
	encode(io.Writer) error
}
