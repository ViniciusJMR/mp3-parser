package mp3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
)

type Frame struct {
	Header    *Header
	AudioData []byte
	rawBytes  []byte
}

type FrameError struct {
	err error
}

func (e FrameError) Error() string {
	return fmt.Sprintf("Error Parsing Frame: %v", e.err)
}

var (
	errFrameLengthLessThanZero = errors.New("Frame Length is less than 0")
)

const SYNC_BYTE byte = 0b11111111

func ParseFrames(r io.Reader) ([]*Frame, error) {
	data, err := io.ReadAll(r)

	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(data)

	frames := []*Frame{}

	for _, err = buf.ReadBytes(SYNC_BYTE); err == nil; _, err = buf.ReadBytes(SYNC_BYTE) {
		var errFrame error
		for f, errFrame := parseFrame(buf); errFrame == nil; f, errFrame = parseFrame(buf) {

			frames = append(frames, f)
			/*
				Reading next SYNC_BYTE if file is organized sequencially so it can
				continue with the loop
				If that's not the case, it'll just try to find the next SYNC_BYTE
			*/
			buf.ReadByte()
		}

		if errors.Is(errFrame, FrameError{}) {
			fmt.Printf("\nError on ParseFrames: %v \nCleaning Frame List...", errFrame)
			frames = []*Frame{}
		}
	}

	if errors.Is(err, io.EOF) {
		return frames, nil
	} else {
		return nil, err
	}
}

func parseFrame(buf *bytes.Buffer) (*Frame, error) {
	headerBuf := make([]byte, 3)
	_, err := buf.Read(headerBuf)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, FrameError{err: err}
	}

	h, err := ParseHeader(headerBuf)

	if err != nil {
		return nil, FrameError{err: err}
	}

	fSize := h.FrameLengthNoHeader()

	// Some parsed frames return a negative size. The current and temporary solution is to ignore these frames.
	// TODO: Investigate why this happens and what is wrong with the decoder.
	if fSize <= 0 {
		return nil, FrameError{err: errFrameLengthLessThanZero}
	}

	audioBuf := make([]byte, fSize)
	_, err = buf.Read(audioBuf)

	if err != nil {
		if errors.Is(err, io.EOF) {
			return nil, io.EOF
		}
		return nil, FrameError{err: err}
	}

	// TODO: Make pre alloc of this bytes
	rawByte := append(h.toBytes(), audioBuf...)

	frame := &Frame{
		Header:   h,
		rawBytes: rawByte,
	}

	return frame, nil
}

func (f *Frame) ToBytes() []byte {
	return f.rawBytes
}
