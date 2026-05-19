package server

import (
	"bytes"
	"compress/zlib"
	"io"

	"github.com/vmihailenco/msgpack/v5"
)

// EncodeMessage serializes a Message to MessagePack bytes with optional zlib compression.
func EncodeMessage(msg Message, compress bool) ([]byte, error) {
	b, err := msgpack.Marshal(msg)
	if err != nil {
		return nil, err
	}
	if !compress || len(b) < 512 {
		return b, nil
	}
	var buf bytes.Buffer
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(b); err != nil {
		w.Close()
		return nil, err
	}
	w.Close()
	return buf.Bytes(), nil
}

// DecodeMessage deserializes MessagePack bytes (with optional zlib) into a Message.
func DecodeMessage(data []byte) (Message, error) {
	var msg Message
	// Try zlib decompression first
	zr, err := zlib.NewReader(bytes.NewReader(data))
	if err == nil {
		decompressed, err := io.ReadAll(zr)
		zr.Close()
		if err == nil {
			data = decompressed
		}
	}
	if err := msgpack.Unmarshal(data, &msg); err != nil {
		return Message{}, err
	}
	return msg, nil
}
