package mp3

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

/*
MP3 Header Mapping based on: https://www.datavoyage.com/mpgscript/mpeghdr.htm
Other good link: https://www.codeproject.com/Articles/8295/MPEG-Audio-Frame-Header
*/

const (
	MPEG_VERSION_2_5 byte = iota << 3 // 00000000
	MPEG_RESERVED                     // 00001000
	MPEG_VERSION_2                    // 00010000
	MPEG_VERSION_1                    // 00011000
)

const (
	LAYER_RESERVED byte = iota << 1 // 00000000
	LAYER_3                         // 00000010
	LAYER_2                         // 00000100
	LAYER_1                         // 00000110
)

const (
	CH_MODE_STEREO byte = iota << 6 // 00000000

	// Stereo
	CH_MODE_JOINT // 01000000
	// Stereo
	CH_MODE_DUAL // 10000000
	// Mono
	CH_MODE_SINGLE // 11000000
)

var V1L3_BITRATE_KBPS = [...]int16{
	0, 32, 40, 48, 56, 64, 80, 96, 112, 128, 160, 192, 224, 256, 320, -1,
}

var V2L2L3_BITRATE_KBPS = [...]int16{
	0, 8, 16, 24, 32, 40, 48, 56, 64, 80, 96, 112, 128, 144, 160, -1,
}

var MPEG1_SAMPLE_RATE_HZ = [...]int32{
	44100, 48000, 32000, -1,
}

var MPEG2_SAMPLE_RATE_HZ = [...]int32{
	22050, 24000, 16000, -1,
}

var MPEG25_SAMPLE_RATE_HZ = [...]int32{
	11025, 12000, 8000, -1,
}

const MPEG1_LAYER3_DEFAULT_SAMPLE_RATE int16 = 1152

var (
	errInvalidSync       = errors.New("Invalid Sync")
	errInvalidVersion    = errors.New("Invalid Version")
	errInvalidLayer      = errors.New("Invalid Layer")
	errInvalidBitrate    = errors.New("Invalid Bitrate Index")
	errInvalidSamplerate = errors.New("Invalid Samplerate Index")
)

type Header struct {
	Version          byte
	LayerDescription byte
	IsCrcProtected   bool
	Bitrate          int16
	Samplerate       int32
	IsPadding        bool
	ChannelMode      byte
	IsCopyright      bool
	IsOriginal       bool
	Emphasis         byte
	rawBytes         []byte
}

func ParseHeader(data []byte) (*Header, error) {
	// Second byte
	if !isSync(data[0]) {
		return nil, errInvalidSync
	}

	version := data[0] & 0b00011000

	if version == MPEG_RESERVED {
		return nil, errInvalidVersion
	}

	layer := data[0] & 0b00000110

	if layer == LAYER_RESERVED {
		return nil, errInvalidLayer
	}

	/*
		Protection bit
		0 - Protected by CRC (16bit crc follows header)
		1 - Not protected
	*/
	isCrcProtected := data[0]&0b00000001 == 0b00000000

	// Third byte
	bitrateIndex := data[1] & 0b11110000 >> 4

	var bitrate int16

	switch version {
	case MPEG_VERSION_1:
		bitrate = V1L3_BITRATE_KBPS[bitrateIndex]
	case MPEG_VERSION_2:
		bitrate = V2L2L3_BITRATE_KBPS[bitrateIndex]
	case MPEG_VERSION_2_5:
		bitrate = V2L2L3_BITRATE_KBPS[bitrateIndex]
	}

	if bitrate == -1 {
		return nil, errInvalidBitrate
	}

	samplerateIndex := data[1] & 0b00001100 >> 2

	var samplerate int32

	switch version {
	case MPEG_VERSION_1:
		samplerate = MPEG1_SAMPLE_RATE_HZ[samplerateIndex]
	case MPEG_VERSION_2:
		samplerate = MPEG2_SAMPLE_RATE_HZ[samplerateIndex]
	case MPEG_VERSION_2_5:
		samplerate = MPEG25_SAMPLE_RATE_HZ[samplerateIndex]
	}

	if samplerate == -1 {
		return nil, errInvalidSamplerate
	}

	isPadding := (data[1] & 0b00000010 >> 1) == 0b00000001

	// Fourth byte

	channelMode := data[2] & 0b11000000

	isCopyright := data[2]&0b00001000 == 0b00001000

	isOriginal := data[2]&0b00000100 == 0b00000100

	emphasis := data[2] & 0b00000011

	rawBytes := append([]byte{0b11111111}, data...)

	h := &Header{
		Version:          version,
		LayerDescription: layer,
		IsCrcProtected:   isCrcProtected,
		Bitrate:          bitrate,
		Samplerate:       samplerate,
		IsPadding:        isPadding,
		ChannelMode:      channelMode,
		IsCopyright:      isCopyright,
		IsOriginal:       isOriginal,
		Emphasis:         emphasis,
		rawBytes:         rawBytes,
	}

	return h, nil
}

func isSync(b byte) bool {
	return b&0b11100000 == 0b11100000
}

func (h *Header) toBytes() []byte {
	return h.rawBytes
}

/*
	Calculates the frame size using the information found in the header.
	Does not include the Header Bytes in the calculation.

	Layer 1: FrameLengthInBytes = (12 * BitRate / SampleRate + Padding) * 4
	Layer 2 & 3: FrameLengthInBytes = 144 * BitRate / SampleRate + Padding
*/
func (h *Header) FrameLengthNoHeader() int {
	var bitrate int = int(h.Bitrate) * 1000
	var padding int
	var size int

	if h.IsPadding {
		padding = 1
	}

	if h.Version == MPEG_VERSION_1 {
		size = 144*bitrate/int(h.Samplerate) + padding
	} else {
		// version 2 or 2.5
		size = (72*bitrate/int(h.Samplerate) + padding)
	}

	// Removing bytes from header
	return size - 4
}

/*
Returns the size in bytes of the Side Information using the following mapping:
MPEG 1 - Single Channel => 17
MPEG 1 - Stereo Channel => 32
MPEG 2 - Single Channel => 9
MPEG 2 - Stereo Channel => 17
*/
func (h *Header) SideInfoLength() int {
	var length int
	switch h.Version {
	case MPEG_VERSION_1:
		if h.ChannelMode == CH_MODE_SINGLE {
			length = 17
		} else {
			length = 32
		}
	case MPEG_VERSION_2, MPEG_VERSION_2_5:
		if h.ChannelMode == CH_MODE_SINGLE {
			length = 9
		} else {
			length = 17
		}
	}

	return length
}

func (h *Header) VersionStr() string {
	var str string
	switch h.Version {
	case MPEG_VERSION_1:
		str = "MPEG 1"
	case MPEG_VERSION_2:
		str = "MPEG 2"
	case MPEG_VERSION_2_5:
		str = "MPEG 2.5"
	}
	return str
}

func (h *Header) LayerStr() string {
	var str string
	switch h.LayerDescription {
	case LAYER_1:
		str = "Layer 1"
	case LAYER_2:
		str = "Layer 2"
	case LAYER_3:
		str = "Layer 3"
	}
	return str
}

func (h *Header) ChannelModeStr() string {
	var str string
	switch h.ChannelMode {
	case CH_MODE_SINGLE:
		str = "Single Mode"
	case CH_MODE_DUAL:
		str = "Dual Mode"
	case CH_MODE_JOINT:
		str = "Joint Mode"
	case CH_MODE_STEREO:
		str = "Stereo Mode"
	}
	return str
}

func (h *Header) String() string {
	var b strings.Builder
	fmt.Fprintf(&b, " Layer: %v\n", h.LayerStr())
	fmt.Fprintf(&b, " Version: %v\n", h.VersionStr())
	fmt.Fprintf(&b, " Protection: %v\n", h.IsCrcProtected)
	fmt.Fprintf(&b, " BitRate: %v\n", h.Bitrate)
	fmt.Fprintf(&b, " SampleRate: %v\n", h.Samplerate)
	fmt.Fprintf(&b, " Pad: %v\n", h.IsPadding)
	fmt.Fprintf(&b, " ChannelMode: %v\n", h.ChannelModeStr())
	fmt.Fprintf(&b, " CopyRight: %v\n", h.IsCrcProtected)
	fmt.Fprintf(&b, " Original: %v\n", h.IsOriginal)
	fmt.Fprintf(&b, " Emphasis: %v\n", h.Emphasis)
	return b.String()
}

// Returns the duration of a frame in seconds
func (h *Header) GetDuration() time.Duration {
	frameDuration := float64(MPEG1_LAYER3_DEFAULT_SAMPLE_RATE) / float64(h.Samplerate)

	return time.Duration(frameDuration * float64(time.Second))
}
