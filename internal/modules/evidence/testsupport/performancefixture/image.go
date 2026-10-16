package performancefixture

import (
	"bytes"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"image"
	"image/color"
	"image/png"
)

// ImagePayload returns a deterministic, decodable synthetic PNG of exactly the
// requested physical size. An ancillary text chunk supplies the payload mass;
// no sparse file, metadata-only size or shared object substitutes for the bytes.
func ImagePayload(byteSize int) ([]byte, error) {
	if byteSize < 256 || byteSize > 2<<20 {
		return nil, errors.New("evidence fixture PNG size is outside its supported bound")
	}
	var encoded bytes.Buffer
	pixels := image.NewRGBA(image.Rect(0, 0, 2, 2))
	pixels.Set(0, 0, color.RGBA{R: 120, A: 255})
	pixels.Set(1, 1, color.RGBA{B: 180, A: 255})
	if err := png.Encode(&encoded, pixels); err != nil {
		return nil, err
	}
	original := encoded.Bytes()
	// IEND is the final 12-byte chunk. Insert a valid tEXt chunk immediately
	// before it, retaining a complete PNG stream with checksummed payload bytes.
	textSize := byteSize - len(original) - 12
	if textSize < len("fixture\x00") {
		return nil, errors.New("evidence fixture PNG has insufficient payload space")
	}
	text := bytes.Repeat([]byte{'x'}, textSize)
	copy(text, []byte("fixture\x00"))
	output := make([]byte, 0, byteSize)
	output = append(output, original[:len(original)-12]...)
	output = binary.BigEndian.AppendUint32(output, uint32(textSize))
	chunkStart := len(output)
	output = append(output, []byte("tEXt")...)
	output = append(output, text...)
	output = binary.BigEndian.AppendUint32(output, crc32.ChecksumIEEE(output[chunkStart:]))
	output = append(output, original[len(original)-12:]...)
	return output, nil
}
