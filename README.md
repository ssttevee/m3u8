# M3U8

Apple HLS playlist parser for Go.

The goal of this project is to create a complete vanilla m3u8 parser with idiomatic code and smart abstractions that is as close to the [specification](https://tools.ietf.org/html/rfc8216#section-4.3.2.4) as possible.

## Documentation

Much of the specification has been imported and abbreviated into the comments. Please view the [godocs](http://godoc.org/github.com/ssttevee/m3u8) for documentation.

## Examples

Parse a playlist from a stream:

```
f, err := os.Open("playlist.m3u8")
if err != nil {
	panic(err)
}

plist, err := m3u8.NewDecoder(f).Decode()
if err != nil {
	panic(err)
}

switch plist.(type) {
case *m3u8.MasterPlaylist:
	fmt.Println("I am a master playlist!")
case *m3u8.MediaPlaylist:
	fmt.Println("I am a media playlist!")
}
```

Parse a playlist from memory:

```
data := []byte(`#EXTM3U...`)

plist, err := m3u8.DecodePlaylist(data)
if err != nil {
	panic(err)
}

switch plist.(type) {
case *m3u8.MasterPlaylist:
	fmt.Println("I am a master playlist!")
case *m3u8.MediaPlaylist:
	fmt.Println("I am a media playlist!")
}
```

Parse a playlist with extraneous tags:

```
decoder := m3u8.NewDecoder(r)
decoder.Strict = false
plist, err := decoder.Decode()
if err != nil {
	panic(err)
}
```

Encoding a playlist:

```
var plist m3u8.Playlist
var buf bytes.Buffer

if err := m3u8.NewEncoder(&buf).Encode(plist); err != nil {
	panic(err)
}
```

## Todo

* Better validation when decoding/encoding.
* Abstractions for easier playlist manipulation.
* More tests!

Pull requests are welcome.

## Warning

This library is in no way stable yet. If you want to use it in your own program, please consider using [go modules](https://github.com/golang/go/wiki/Modules).
