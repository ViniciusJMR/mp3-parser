package mp3

import (
	"testing"
)

func TestHeaderParseErrors(t *testing.T) {
	var tests = []struct {
		desc string
		data []byte
		err  error
	}{
		{
			desc: "Invalid sync bits should return Invalid Sync Error",
			data: []byte{0b11000010, 0b11111111, 0b01010101},
			err:  errInvalidSync},
		{
			desc: "Invalid version bits should return Invalid Version Error",
			data: []byte{0b11101010, 0b11111111, 0b01010101},
			err:  errInvalidVersion},
		{
			desc: "Invalid layer bits should return Invalid Layer Error",
			data: []byte{0b11100000, 0b11111111, 0b01010101},
			err:  errInvalidLayer,
		},
		{
			desc: "Invalid bitrate index bits should return Invalid Bitrate index Error",
			data: []byte{0b11100010, 0b11111111, 0b01010101},
			err:  errInvalidBitrate,
		},
		{
			desc: "Invalid samplerate index bits should return Invalid Samplerate index Error",
			data: []byte{0b11100010, 0b11011100, 0b01010101},
			err:  errInvalidSamplerate,
		},
		{
			desc: "Valid bytes should not return error",
			data: []byte{0b11100010, 0b11011000, 0b01010101},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			if _, err := ParseHeader(tt.data); err != tt.err {
				t.Errorf("\nGot err: %v\nWant err: %v", err, tt.err)
			}
		})
	}
}

// TODO: Adicionar mais testes aqui mas estou com preguiça agr :)
func TestHeaderParsingIntegrity(t *testing.T) {
	headerWantedValue1 := Header{
		Version:          MPEG_VERSION_1,
		LayerDescription: LAYER_3,
		IsCrcProtected:   true,
		Bitrate:          V1L3_BITRATE_KBPS[3],
		Samplerate:       MPEG1_SAMPLE_RATE_HZ[0],
		IsPadding:        false,
		ChannelMode:      CH_MODE_STEREO,
		IsCopyright:      false,
		IsOriginal:       false,
		Emphasis:         3,
	}

	headerWantedValue2 := Header{
		Version:          MPEG_VERSION_2,
		LayerDescription: LAYER_3,
		IsCrcProtected:   false,
		Bitrate:          V2L2L3_BITRATE_KBPS[7],
		Samplerate:       MPEG2_SAMPLE_RATE_HZ[1],
		IsPadding:        true,
		ChannelMode:      CH_MODE_DUAL,
		IsCopyright:      true,
		IsOriginal:       true,
		Emphasis:         1,
	}

	tests := []struct {
		desc string
		data []byte
		want Header
	}{
		{
			desc: "Test header parsed values (1)",
			data: []byte{0b11111010, 0b00110000, 0b00010011},
			want: headerWantedValue1,
		},
		{
			desc: "Test header parsed values (2)",
			data: []byte{0b11110011, 0b01110111, 0b10101101},
			want: headerWantedValue2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.desc, func(t *testing.T) {
			h, err := ParseHeader(tt.data)
			if err != nil {
				t.Fatalf("\nThe Parsing here shouldn't return error wtf happened: %v", err)
			}

			cmpIntHelper(t, int(h.Version), int(tt.want.Version), "Version")
			cmpIntHelper(t, int(h.LayerDescription), int(tt.want.LayerDescription), "Layer")
			cmpBoolHelper(t, h.IsCrcProtected, tt.want.IsCrcProtected, "IsCrcProtected")
			cmpIntHelper(t, int(h.Bitrate), int(tt.want.Bitrate), "Bitrate")
			cmpIntHelper(t, int(h.Samplerate), int(tt.want.Samplerate), "Samplerate")
			cmpBoolHelper(t, h.IsPadding, tt.want.IsPadding, "IsPadding")
			cmpIntHelper(t, int(h.ChannelMode), int(tt.want.ChannelMode), "Channel Mode")
			cmpBoolHelper(t, h.IsCopyright, tt.want.IsCopyright, "IsCopyright")
			cmpBoolHelper(t, h.IsOriginal, tt.want.IsOriginal, "IsOriginal")
			cmpIntHelper(t, int(h.Emphasis), int(tt.want.Emphasis), "Emphasis")
		})
	}
}

func TestHeaderParsedSize(t *testing.T) {
	headerValue1 := &Header{
		Version:          MPEG_VERSION_1,
		LayerDescription: LAYER_3,
		IsCrcProtected:   false,
		Bitrate:          48,
		Samplerate:       32000,
		IsPadding:        false,
		ChannelMode:      CH_MODE_STEREO,
		IsCopyright:      false,
		IsOriginal:       false,
		Emphasis:         3,
	}

	headerValue2 := &Header{
		Version:          MPEG_VERSION_1,
		LayerDescription: LAYER_3,
		IsCrcProtected:   true,
		Bitrate:          56,
		Samplerate:       24000,
		IsPadding:        true,
		ChannelMode:      CH_MODE_DUAL,
		IsCopyright:      true,
		IsOriginal:       true,
		Emphasis:         1,
	}

	headerValue3 := &Header{
		Version:          MPEG_VERSION_1,
		LayerDescription: LAYER_3,
		IsCrcProtected:   true,
		Bitrate:          128,
		Samplerate:       44100,
		IsPadding:        false,
		ChannelMode:      CH_MODE_DUAL,
		IsCopyright:      true,
		IsOriginal:       true,
		Emphasis:         1,
	}

	tests := []struct {
		desc string
		data *Header
		want int
	}{
		{
			desc: "First Parsed size should return <ALTERAR>",
			data: headerValue1,
			want: 212, // TODO: Calcular na mão e alterar aqui
		},
		{
			desc: "Second Parsed size should return <ALTERAR>",
			data: headerValue2,
			want: 333,
		},
		{
			desc: "Third Parsed size should return <ALTERAR",
			data: headerValue3,
			want: 413,
		},
	}

	for _, tt := range tests {
		if tt.data.FrameLengthNoHeader() != tt.want {
			t.Errorf("\nGot size: %v\nWant size: %v", tt.data.FrameLengthNoHeader(), tt.want)
		}
	}
}

func cmpBoolHelper(t *testing.T, got bool, want bool, desc string) {
	if got != want {
		t.Errorf("\n%v not equal.\nGot: %v\nWant: %v", desc, got, want)
	}
}

func cmpIntHelper(t *testing.T, got int, want int, desc string) {
	if got != want {
		t.Errorf("\n%v not equal.\nGot: %v\nWant: %v", desc, got, want)
	}
}
