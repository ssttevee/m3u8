// Package m3u8 provides abstractions to work with Apple HLS playlist files.
package m3u8

type Type int

const (
	Master Type = iota + 1
	Media
)

type Playlist interface {
	Type() Type
	encode() ([]byte, error)
}
