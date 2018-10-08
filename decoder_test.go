package m3u8_test

import (
	"strings"
	"testing"
	"time"

	"github.com/ssttevee/m3u8"
	"github.com/stretchr/testify/assert"
)

func TestDecodePlaylist(t *testing.T) {
	type uriDuration struct {
		uri      string
		duration time.Duration
	}

	t.Run("simple media playlist", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-TARGETDURATION:10\n#EXT-X-VERSION:4\n#EXT-X-BYTERANGE:8765213@7896\n#EXTINF:9.009,\nhttp://media.example.com/first.ts\n#EXTINF:9.009,\nhttp://media.example.com/second.ts\n#EXTINF:3.003,\nhttp://media.example.com/third.ts\n#EXT-X-ENDLIST\n"
		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Media, plist.Type(), "should be a media playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MediaPlaylist)

		assert.Equal(t, 4, mplist.Version)
		assert.Equal(t, uint64(10), mplist.TargetDuration)

		if assert.Len(t, mplist.Segments, 3) {
			for i, ud := range []*uriDuration{
				{"http://media.example.com/first.ts", 9*time.Second + 9*time.Millisecond},
				{"http://media.example.com/second.ts", 9*time.Second + 9*time.Millisecond},
				{"http://media.example.com/third.ts", 3*time.Second + 3*time.Millisecond},
			} {
				if assert.NotNil(t, mplist.Segments[i]) {
					assert.Equal(t, ud.uri, mplist.Segments[i].URI)
					assert.Equal(t, ud.duration, mplist.Segments[i].Duration)
				}
			}

			if assert.NotNil(t, mplist.Segments[0].ByteRange) {
				assert.Equal(t, int64(8765213), mplist.Segments[0].ByteRange.Length)
				assert.Equal(t, int64(7896), mplist.Segments[0].ByteRange.Start)
			}
		}
	})

	t.Run("live media playlist using https", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:8\n#EXT-X-MEDIA-SEQUENCE:2680\n\n#EXTINF:7.975,\nhttps://priv.example.com/fileSequence2680.ts\n#EXTINF:7.941,\nhttps://priv.example.com/fileSequence2681.ts\n#EXTINF:7.975,\nhttps://priv.example.com/fileSequence2682.ts\n"
		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Media, plist.Type(), "should be a media playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MediaPlaylist)

		assert.Equal(t, 3, mplist.Version)
		assert.Equal(t, uint64(8), mplist.TargetDuration)
		assert.Equal(t, uint64(2680), mplist.MediaSequence)

		if assert.Len(t, mplist.Segments, 3) {
			for i, ud := range []*uriDuration{
				{"https://priv.example.com/fileSequence2680.ts", 7*time.Second + 975*time.Millisecond},
				{"https://priv.example.com/fileSequence2681.ts", 7*time.Second + 941*time.Millisecond},
				{"https://priv.example.com/fileSequence2682.ts", 7*time.Second + 975*time.Millisecond},
			} {
				if assert.NotNil(t, mplist.Segments[i]) {
					assert.Equal(t, ud.uri, mplist.Segments[i].URI)
					assert.Equal(t, ud.duration, mplist.Segments[i].Duration)
				}
			}
		}
	})

	t.Run("media playlist with encrypted segments", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-MEDIA-SEQUENCE:7794\n#EXT-X-TARGETDURATION:15\n\n#EXT-X-KEY:METHOD=AES-128,URI=\"https://priv.example.com/key.php?r=52\"\n\n#EXTINF:2.833,\nhttp://media.example.com/fileSequence52-A.ts\n#EXTINF:15.0,\nhttp://media.example.com/fileSequence52-B.ts\n#EXTINF:13.333,\nhttp://media.example.com/fileSequence52-C.ts\n\n#EXT-X-KEY:METHOD=AES-128,URI=\"https://priv.example.com/key.php?r=53\"\n\n#EXTINF:15.0,\nhttp://media.example.com/fileSequence53-A.ts\n"
		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Media, plist.Type(), "should be a media playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MediaPlaylist)

		assert.Equal(t, 3, mplist.Version)
		assert.Equal(t, uint64(15), mplist.TargetDuration)
		assert.Equal(t, uint64(7794), mplist.MediaSequence)

		if assert.Len(t, mplist.Segments, 4) {
			for i, ud := range []*uriDuration{
				{"http://media.example.com/fileSequence52-A.ts", 2*time.Second + 833*time.Millisecond},
				{"http://media.example.com/fileSequence52-B.ts", 15 * time.Second},
				{"http://media.example.com/fileSequence52-C.ts", 13*time.Second + 333*time.Millisecond},
				{"http://media.example.com/fileSequence53-A.ts", 15 * time.Second},
			} {
				if assert.NotNil(t, mplist.Segments[i]) {
					assert.Equal(t, ud.uri, mplist.Segments[i].URI)
					assert.Equal(t, ud.duration, mplist.Segments[i].Duration)
				}
			}

			if assert.NotNil(t, mplist.Segments[0].Key) {
				assert.Equal(t, m3u8.AES128, mplist.Segments[0].Key.Method)
				assert.Equal(t, "https://priv.example.com/key.php?r=52", mplist.Segments[0].Key.URI)
			}

			if assert.NotNil(t, mplist.Segments[3].Key) {
				assert.Equal(t, m3u8.AES128, mplist.Segments[3].Key.Method)
				assert.Equal(t, "https://priv.example.com/key.php?r=53", mplist.Segments[3].Key.URI)
			}
		}
	})

	t.Run("media playlist with daterange", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-TARGETDURATION:10\n#EXT-X-VERSION:3\n#EXT-X-DATERANGE:ID=\"splice-6FFFFFF0\",START-DATE=\"2014-03-05T11:15:00Z\",PLANNED-DURATION=59.993,SCTE35-OUT=0xFC002F0000000000FF000014056FFFFFF000E011622DCAFF000052636200000000000A0008029896F50000008700000000\n#EXTINF:9.009,\nsegment-1.ts\n#EXT-X-DATERANGE:ID=\"splice-6FFFFFF0\",START-DATE=\"2014-03-05T11:15:00Z\",DURATION=59.993,SCTE35-IN=0xFC002A0000000000FF00000F056FFFFFF000401162802E6100000000000A0008029896F50000008700000000\n#EXTINF:3.003,\nsegment-2.ts\n#EXT-X-ENDLIST\n"

		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Media, plist.Type(), "should be a media playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MediaPlaylist)

		if assert.Len(t, mplist.Segments, 2) {
			if assert.NotNil(t, mplist.Segments[0]) {
				s := mplist.Segments[0]
				assert.Equal(t, "segment-1.ts", s.URI)
				assert.Equal(t, 9*time.Second+9*time.Millisecond, s.Duration)

				if assert.NotNil(t, s.DateRange) {
					assert.Equal(t, "splice-6FFFFFF0", s.DateRange.ID)
					assert.Equal(t, "2014-03-05T11:15:00Z", s.DateRange.StartDate)
					assert.Equal(t, 59*time.Second+993*time.Millisecond, s.DateRange.PlannedDuration)
					assert.Equal(t, []byte{0xFC, 0x00, 0x2F, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x14, 0x05, 0x6F, 0xFF, 0xFF, 0xF0, 0x00, 0xE0, 0x11, 0x62, 0x2D, 0xCA, 0xFF, 0x00, 0x00, 0x52, 0x63, 0x62, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x08, 0x02, 0x98, 0x96, 0xF5, 0x00, 0x00, 0x00, 0x87, 0x00, 0x00, 0x00, 0x00}, s.DateRange.SCTE35Out)
				}
			}

			if assert.NotNil(t, mplist.Segments[1]) {
				s := mplist.Segments[1]
				assert.Equal(t, "segment-2.ts", s.URI)
				assert.Equal(t, 3*time.Second+3*time.Millisecond, s.Duration)

				if assert.NotNil(t, s.DateRange) {
					assert.Equal(t, "splice-6FFFFFF0", s.DateRange.ID)
					assert.Equal(t, "2014-03-05T11:15:00Z", s.DateRange.StartDate)
					assert.Equal(t, 59*time.Second+993*time.Millisecond, s.DateRange.Duration)
					assert.Equal(t, []byte{0xFC, 0x00, 0x2A, 0x00, 0x00, 0x00, 0x00, 0x00, 0xFF, 0x00, 0x00, 0x0F, 0x05, 0x6F, 0xFF, 0xFF, 0xF0, 0x00, 0x40, 0x11, 0x62, 0x80, 0x2E, 0x61, 0x00, 0x00, 0x00, 0x00, 0x00, 0x0A, 0x00, 0x08, 0x02, 0x98, 0x96, 0xF5, 0x00, 0x00, 0x00, 0x87, 0x00, 0x00, 0x00, 0x00}, s.DateRange.SCTE35In)
				}
			}
		}
	})

	t.Run("master playlist", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1280000,AVERAGE-BANDWIDTH=1000000\nhttp://example.com/low.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=2560000,AVERAGE-BANDWIDTH=2000000\nhttp://example.com/mid.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=7680000,AVERAGE-BANDWIDTH=6000000\nhttp://example.com/hi.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=65000,CODECS=\"mp4a.40.5\"\nhttp://example.com/audio-only.m3u8\n"
		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Master, plist.Type(), "should be a master playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MasterPlaylist)

		if assert.Len(t, mplist.VariantStreams, 4) {
			assert.Equal(t, uint64(1280000), mplist.VariantStreams[0].Bandwidth)
			assert.Equal(t, uint64(1000000), mplist.VariantStreams[0].AverageBandwidth)
			assert.Equal(t, "http://example.com/low.m3u8", mplist.VariantStreams[0].URI)

			assert.Equal(t, uint64(2560000), mplist.VariantStreams[1].Bandwidth)
			assert.Equal(t, uint64(2000000), mplist.VariantStreams[1].AverageBandwidth)
			assert.Equal(t, "http://example.com/mid.m3u8", mplist.VariantStreams[1].URI)

			assert.Equal(t, uint64(7680000), mplist.VariantStreams[2].Bandwidth)
			assert.Equal(t, uint64(6000000), mplist.VariantStreams[2].AverageBandwidth)
			assert.Equal(t, "http://example.com/hi.m3u8", mplist.VariantStreams[2].URI)

			assert.Equal(t, uint64(65000), mplist.VariantStreams[3].Bandwidth)
			if assert.Len(t, mplist.VariantStreams[3].Codecs, 1) {
				assert.Equal(t, "mp4a.40.5", mplist.VariantStreams[3].Codecs[0])
			}
			assert.Equal(t, "http://example.com/audio-only.m3u8", mplist.VariantStreams[3].URI)
		}
	})

	t.Run("master playlist with i-frames", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1280000\nlow/audio-video.m3u8\n#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=86000,URI=\"low/iframe.m3u8\"\n#EXT-X-STREAM-INF:BANDWIDTH=2560000\nmid/audio-video.m3u8\n#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=150000,URI=\"mid/iframe.m3u8\"\n#EXT-X-STREAM-INF:BANDWIDTH=7680000\nhi/audio-video.m3u8\n#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=550000,URI=\"hi/iframe.m3u8\"\n#EXT-X-STREAM-INF:BANDWIDTH=65000,CODECS=\"mp4a.40.5\"\naudio-only.m3u8\n"
		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Master, plist.Type(), "should be a master playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MasterPlaylist)

		if assert.Len(t, mplist.VariantStreams, 4) {
			assert.Equal(t, uint64(1280000), mplist.VariantStreams[0].Bandwidth)
			assert.Equal(t, "low/audio-video.m3u8", mplist.VariantStreams[0].URI)

			assert.Equal(t, uint64(2560000), mplist.VariantStreams[1].Bandwidth)
			assert.Equal(t, "mid/audio-video.m3u8", mplist.VariantStreams[1].URI)

			assert.Equal(t, uint64(7680000), mplist.VariantStreams[2].Bandwidth)
			assert.Equal(t, "hi/audio-video.m3u8", mplist.VariantStreams[2].URI)

			assert.Equal(t, uint64(65000), mplist.VariantStreams[3].Bandwidth)
			if assert.Len(t, mplist.VariantStreams[3].Codecs, 1) {
				assert.Equal(t, "mp4a.40.5", mplist.VariantStreams[3].Codecs[0])
			}
			assert.Equal(t, "audio-only.m3u8", mplist.VariantStreams[3].URI)
		}

		if assert.Len(t, mplist.IFrameStreams, 3) {
			assert.Equal(t, uint64(86000), mplist.IFrameStreams[0].Bandwidth)
			assert.Equal(t, "low/iframe.m3u8", mplist.IFrameStreams[0].URI)

			assert.Equal(t, uint64(150000), mplist.IFrameStreams[1].Bandwidth)
			assert.Equal(t, "mid/iframe.m3u8", mplist.IFrameStreams[1].URI)

			assert.Equal(t, uint64(550000), mplist.IFrameStreams[2].Bandwidth)
			assert.Equal(t, "hi/iframe.m3u8", mplist.IFrameStreams[2].URI)
		}
	})

	t.Run("master playlist with alternative audio", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"aac\",NAME=\"English\",DEFAULT=YES,AUTOSELECT=YES,LANGUAGE=\"en\",URI=\"main/english-audio.m3u8\"\n#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"aac\",NAME=\"Deutsch\",DEFAULT=NO,AUTOSELECT=YES,LANGUAGE=\"de\",URI=\"main/german-audio.m3u8\"\n#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"aac\",NAME=\"Commentary\",DEFAULT=NO,AUTOSELECT=NO,LANGUAGE=\"en\",URI=\"commentary/audio-only.m3u8\"\n#EXT-X-STREAM-INF:BANDWIDTH=1280000,CODECS=\"...\",AUDIO=\"aac\"\nlow/video-only.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=2560000,CODECS=\"...\",AUDIO=\"aac\"\nmid/video-only.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=7680000,CODECS=\"...\",AUDIO=\"aac\"\nhi/video-only.m3u8\n#EXT-X-STREAM-INF:BANDWIDTH=65000,CODECS=\"mp4a.40.5\",AUDIO=\"aac\"\nmain/english-audio.m3u8\n"

		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Master, plist.Type(), "should be a master playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MasterPlaylist)

		if assert.Len(t, mplist.RenditionMap, 3) {
			if assert.NotNil(t, mplist.RenditionMap[0]) {
				if assert.Equal(t, m3u8.Audio, mplist.RenditionMap[0].Type()) {
					ar := mplist.RenditionMap[0].(*m3u8.AudioRendition)
					assert.Equal(t, "aac", ar.GroupID)
					assert.Equal(t, "English", ar.Name)
					assert.True(t, ar.Default)
					assert.True(t, ar.AutoSelect)
					assert.Equal(t, "en", ar.Language)
					assert.Equal(t, "main/english-audio.m3u8", ar.URI)
				}
			}

			if assert.NotNil(t, mplist.RenditionMap[1]) {
				if assert.Equal(t, m3u8.Audio, mplist.RenditionMap[1].Type()) {
					ar := mplist.RenditionMap[1].(*m3u8.AudioRendition)
					assert.Equal(t, "aac", ar.GroupID)
					assert.Equal(t, "Deutsch", ar.Name)
					assert.False(t, ar.Default)
					assert.True(t, ar.AutoSelect)
					assert.Equal(t, "de", ar.Language)
					assert.Equal(t, "main/german-audio.m3u8", ar.URI)
				}
			}

			if assert.NotNil(t, mplist.RenditionMap[2]) {
				if assert.Equal(t, m3u8.Audio, mplist.RenditionMap[2].Type()) {
					ar := mplist.RenditionMap[2].(*m3u8.AudioRendition)
					assert.Equal(t, "aac", ar.GroupID)
					assert.Equal(t, "Commentary", ar.Name)
					assert.False(t, ar.Default)
					assert.False(t, ar.AutoSelect)
					assert.Equal(t, "en", ar.Language)
					assert.Equal(t, "commentary/audio-only.m3u8", ar.URI)
				}
			}
		}

		if assert.Len(t, mplist.VariantStreams, 4) {
			if assert.NotNil(t, mplist.VariantStreams[0]) {
				vs := mplist.VariantStreams[0]
				assert.Equal(t, uint64(1280000), vs.Bandwidth)
				if assert.Len(t, vs.Codecs, 1) {
					assert.Equal(t, "...", vs.Codecs[0])
				}
				assert.Equal(t, m3u8.Audio, vs.Type)
				assert.Equal(t, "aac", vs.GroupID)
				assert.Equal(t, "low/video-only.m3u8", vs.URI)
			}

			if assert.NotNil(t, mplist.VariantStreams[1]) {
				vs := mplist.VariantStreams[1]
				assert.Equal(t, uint64(2560000), vs.Bandwidth)
				if assert.Len(t, vs.Codecs, 1) {
					assert.Equal(t, "...", vs.Codecs[0])
				}
				assert.Equal(t, m3u8.Audio, vs.Type)
				assert.Equal(t, "aac", vs.GroupID)
				assert.Equal(t, "mid/video-only.m3u8", vs.URI)
			}

			if assert.NotNil(t, mplist.VariantStreams[2]) {
				vs := mplist.VariantStreams[2]
				assert.Equal(t, uint64(7680000), vs.Bandwidth)
				if assert.Len(t, vs.Codecs, 1) {
					assert.Equal(t, "...", vs.Codecs[0])
				}
				assert.Equal(t, m3u8.Audio, vs.Type)
				assert.Equal(t, "aac", vs.GroupID)
				assert.Equal(t, "hi/video-only.m3u8", vs.URI)
			}

			if assert.NotNil(t, mplist.VariantStreams[3]) {
				vs := mplist.VariantStreams[3]
				assert.Equal(t, uint64(65000), vs.Bandwidth)
				if assert.Len(t, vs.Codecs, 1) {
					assert.Equal(t, "mp4a.40.5", vs.Codecs[0])
				}
				assert.Equal(t, m3u8.Audio, vs.Type)
				assert.Equal(t, "aac", vs.GroupID)
				assert.Equal(t, "main/english-audio.m3u8", vs.URI)
			}
		}
	})

	t.Run("master playlist with alternative video", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"low\",NAME=\"Main\",DEFAULT=YES,URI=\"low/main/audio-video.m3u8\"\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"low\",NAME=\"Centerfield\",DEFAULT=NO,URI=\"low/centerfield/audio-video.m3u8\"\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"low\",NAME=\"Dugout\",DEFAULT=NO,URI=\"low/dugout/audio-video.m3u8\"\n\n#EXT-X-STREAM-INF:BANDWIDTH=1280000,CODECS=\"...\",VIDEO=\"low\"\nlow/main/audio-video.m3u8\n\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"mid\",NAME=\"Main\",DEFAULT=YES,URI=\"mid/main/audio-video.m3u8\"\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"mid\",NAME=\"Centerfield\",DEFAULT=NO,URI=\"mid/centerfield/audio-video.m3u8\"\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"mid\",NAME=\"Dugout\",DEFAULT=NO,URI=\"mid/dugout/audio-video.m3u8\"\n\n#EXT-X-STREAM-INF:BANDWIDTH=2560000,CODECS=\"...\",VIDEO=\"mid\"\nmid/main/audio-video.m3u8\n\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"hi\",NAME=\"Main\",DEFAULT=YES,URI=\"hi/main/audio-video.m3u8\"\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"hi\",NAME=\"Centerfield\",DEFAULT=NO,URI=\"hi/centerfield/audio-video.m3u8\"\n#EXT-X-MEDIA:TYPE=VIDEO,GROUP-ID=\"hi\",NAME=\"Dugout\",DEFAULT=NO,URI=\"hi/dugout/audio-video.m3u8\"\n\n#EXT-X-STREAM-INF:BANDWIDTH=7680000,CODECS=\"...\",VIDEO=\"hi\"\nhi/main/audio-video.m3u8\n"

		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Master, plist.Type(), "should be a master playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MasterPlaylist)

		if assert.Len(t, mplist.RenditionMap, 9) {
			for i, r := range mplist.RenditionMap {
				if assert.NotNil(t, r) {
					if assert.Equal(t, m3u8.Video, r.Type()) {
						vr := r.(*m3u8.VideoRendition)
						switch i / 3 {
						case 0:
							assert.Equal(t, "low", vr.GroupID)
						case 1:
							assert.Equal(t, "mid", vr.GroupID)
						case 2:
							assert.Equal(t, "hi", vr.GroupID)
						}

						//if i < 3 {
						//	assert.Equal(t, "low", vr.GroupID)
						//} else if i < 6 {
						//	assert.Equal(t, "mid", vr.GroupID)
						//} else {
						//	assert.Equal(t, "hi", vr.GroupID)
						//}

						switch i % 3 {
						case 0:
							assert.Equal(t, "Main", vr.Name)
							assert.True(t, vr.Default)
						case 1:
							assert.Equal(t, "Centerfield", vr.Name)
							assert.False(t, vr.Default)
						case 2:
							assert.Equal(t, "Dugout", vr.Name)
							assert.False(t, vr.Default)
						}

						assert.Equal(t, vr.GroupID+"/"+strings.ToLower(vr.Name)+"/audio-video.m3u8", vr.URI)
					}
				}
			}
		}

		if assert.Len(t, mplist.VariantStreams, 3) {
			if assert.NotNil(t, mplist.VariantStreams[0]) {
				vs := mplist.VariantStreams[0]
				assert.Equal(t, uint64(1280000), vs.Bandwidth)
				if assert.Len(t, vs.Codecs, 1) {
					assert.Equal(t, "...", vs.Codecs[0])
				}
				assert.Equal(t, m3u8.Video, vs.Type)
				assert.Equal(t, "low", vs.GroupID)
				assert.Equal(t, "low/main/audio-video.m3u8", vs.URI)
			}

			if assert.NotNil(t, mplist.VariantStreams[1]) {
				vs := mplist.VariantStreams[1]
				assert.Equal(t, uint64(2560000), vs.Bandwidth)
				if assert.Len(t, vs.Codecs, 1) {
					assert.Equal(t, "...", vs.Codecs[0])
				}
				assert.Equal(t, m3u8.Video, vs.Type)
				assert.Equal(t, "mid", vs.GroupID)
				assert.Equal(t, "mid/main/audio-video.m3u8", vs.URI)
			}

			if assert.NotNil(t, mplist.VariantStreams[2]) {
				vs := mplist.VariantStreams[2]
				assert.Equal(t, uint64(7680000), vs.Bandwidth)
				if assert.Len(t, vs.Codecs, 1) {
					assert.Equal(t, "...", vs.Codecs[0])
				}
				assert.Equal(t, m3u8.Video, vs.Type)
				assert.Equal(t, "hi", vs.GroupID)
				assert.Equal(t, "hi/main/audio-video.m3u8", vs.URI)
			}
		}
	})

	t.Run("master playlist with session data", func(t *testing.T) {
		const data = "#EXTM3U\n#EXT-X-SESSION-DATA:DATA-ID=\"com.example.lyrics\",URI=\"lyrics.json\"\n#EXT-X-SESSION-DATA:DATA-ID=\"com.example.title\",LANGUAGE=\"en\",VALUE=\"This is an example\"\n#EXT-X-SESSION-DATA:DATA-ID=\"com.example.title\",LANGUAGE=\"es\",VALUE=\"Este es un ejemplo\"\n"

		plist, err := m3u8.DecodePlaylist([]byte(data))
		if !assert.Nil(t, err, "should sucessfully parse") {
			t.FailNow()
		}

		if !assert.Equal(t, m3u8.Master, plist.Type(), "should be a master playlist") {
			t.FailNow()
		}

		mplist := plist.(*m3u8.MasterPlaylist)

		if assert.Len(t, mplist.SessionData, 3) {
			if assert.NotNil(t, mplist.SessionData[0]) {
				sde := mplist.SessionData[0]
				assert.Equal(t, "com.example.lyrics", sde.ID)
				assert.Equal(t, "lyrics.json", sde.URI)
			}

			if assert.NotNil(t, mplist.SessionData[1]) {
				sde := mplist.SessionData[1]
				assert.Equal(t, "com.example.title", sde.ID)
				assert.Equal(t, "en", sde.Language)
				assert.Equal(t, "This is an example", sde.Value)
			}

			if assert.NotNil(t, mplist.SessionData[2]) {
				sde := mplist.SessionData[2]
				assert.Equal(t, "com.example.title", sde.ID)
				assert.Equal(t, "es", sde.Language)
				assert.Equal(t, "Este es un ejemplo", sde.Value)
			}
		}
	})
}
