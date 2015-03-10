// Copyright 2015 The go-hep Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rio

import (
	"bufio"
	"compress/flate"
	"io"

	riobin "github.com/gonuts/binary"
)

// Writer is a rio write-only stream
type Writer struct {
	w *bufio.Writer

	options Options
	version Version

	recs    map[string]*Record
	offsets map[*Record][]int64
}

// NewWriter returns a new write-only rio stream
func NewWriter(w io.Writer) (*Writer, error) {
	// a rio stream starts with rio magic.
	_, err := w.Write(rioMagic[:])
	if err != nil {
		return nil, err
	}

	return &Writer{
		w:       bufio.NewWriter(w),
		options: NewOptions(CompressDefault, flate.DefaultCompression, 0),
		version: 1,
		recs:    make(map[string]*Record),
	}, nil
}

// SetCompressor enables compression and sets the compression method.
func (w *Writer) SetCompressor(compr CompressorKind, lvl int) error {
	var err error

	// FIXME(sbinet) handle codec (gob|cbor|xdr|riobin|...)
	codec := 0
	w.options = NewOptions(compr, lvl, codec)

	return err
}

// Record adds a Record to the list of records to write or
// returns the Record with that name.
func (w *Writer) Record(name string) *Record {
	rec, ok := w.recs[name]
	if !ok {
		rec = newRecord(name, w.options)
		rec.w = w
		w.recs[name] = rec
	}
	return rec
}

// Close finishes writing the rio write-only stream.
// It does not (and can not) close the underlying writer.
func (w *Writer) Close() error {
	return w.w.Flush()
}

// encoder manages the encoding of data values into rioRecords
type encoder struct {
	w io.Writer
}

func (enc *encoder) Encode(v interface{}) error {
	switch v := v.(type) {
	case Marshaler:
		return v.RioMarshal(enc.w)
	}

	e := riobin.NewEncoder(enc.w)
	e.Order = Endian
	return e.Encode(v)
}
