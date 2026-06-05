package packet

import (
	"bytes"
	"compress/zlib"
	"errors"
	"io"
)

const CompressionThreshold = 256

type Compressor struct {
	threshold int
	enabled   bool
}

func NewCompressor(threshold int) *Compressor {
	return &Compressor{
		threshold: threshold,
		enabled:   threshold >= 0,
	}
}

func (c *Compressor) CompressPacket(packetID int32, packetData []byte) ([]byte, error) {
	if !c.enabled {
		w := NewWriter()
		w.VarInt(packetID)
		w.Bytes(packetData)
		return w.Raw(), nil
	}

	bodyWriter := NewWriter()
	bodyWriter.VarInt(packetID)
	bodyWriter.Bytes(packetData)
	uncompressed := bodyWriter.Raw()

	if len(uncompressed) < c.threshold {
		result := NewWriter()
		result.VarInt(0)
		result.Bytes(uncompressed)
		return result.Raw(), nil
	}

	var compressed bytes.Buffer
	ZLibWriter := zlib.NewWriter(&compressed)
	_, err := ZLibWriter.Write(uncompressed)
	if err != nil {
		return nil, err
	}
	ZLibWriter.Close()

	result := NewWriter()
	result.VarInt(int32(len(uncompressed)))
	result.Bytes(compressed.Bytes())
	return result.Raw(), nil
}

func (c *Compressor) DecompressPacket(data []byte) (packetID int32, packetData []byte, err error) {
	if !c.enabled {
		r := NewReader(data)
		packetID = r.VarInt()
		if r.Error() != nil {
			return 0, nil, r.Error()
		}
		packetData = r.buf[r.Offset():]
		return
	}

	r := NewReader(data)
	dataLength := r.VarInt()

	if r.Error() != nil {
		return 0, nil, r.Error()
	}

	if dataLength == 0 {
		packetID = r.VarInt()
		if r.Error() != nil {
			return 0, nil, r.Error()
		}
		packetData = r.buf[r.Offset():]
		return
	}

	reader, err := zlib.NewReader(bytes.NewReader(r.buf[r.Offset():]))
	if err != nil {
		return 0, nil, err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return 0, nil, err
	}

	if len(decompressed) < 1 {
		return 0, nil, errors.New("decompressed data to small")
	}

	r2 := NewReader(decompressed)
	packetID = r2.VarInt()
	if r2.Error() != nil {
		return 0, nil, r2.Error()
	}

	packetData = r2.buf[r2.Offset():]
	return
}

func (c *Compressor) IsEnabled() bool {
	return c.enabled
}

func (c *Compressor) SetThreshold(threshold int) {
	c.threshold = threshold
	c.enabled = threshold >= 0
}

func (c *Compressor) GetThreshold() int {
	return c.threshold
}
