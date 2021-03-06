// Copyright 2018 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package rcont

import (
	"io"
	"io/ioutil"
	"os"
	"reflect"
	"strings"
	"testing"

	"go-hep.org/x/hep/groot/internal/rtests"
	"go-hep.org/x/hep/groot/rbase"
	"go-hep.org/x/hep/groot/rbytes"
	"go-hep.org/x/hep/groot/root"
	"go-hep.org/x/hep/groot/rtypes"
)

func TestWRBuffer(t *testing.T) {
	for _, tc := range []struct {
		name string
		want rtests.ROOTer
	}{
		{
			name: "TList",
			want: &List{
				obj:  rbase.Object{ID: 0x0, Bits: 0x3000000},
				name: "list-name",
				objs: []root.Object{
					rbase.NewNamed("n0", "t0"),
					rbase.NewNamed("n1", "t1"),
				},
			},
		},
		{
			name: "TObjArray",
			want: &ObjArray{
				obj:  rbase.Object{ID: 0x0, Bits: 0x3000000},
				name: "my-objs",
				objs: []root.Object{
					rbase.NewNamed("n0", "t0"),
					rbase.NewNamed("n1", "t1"),
					rbase.NewNamed("n2", "t2"),
				},
				last: 2,
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			{
				wbuf := rbytes.NewWBuffer(nil, nil, 0, nil)
				wbuf.SetErr(io.EOF)
				_, err := tc.want.MarshalROOT(wbuf)
				if err == nil {
					t.Fatalf("expected an error")
				}
				if err != io.EOF {
					t.Fatalf("got=%v, want=%v", err, io.EOF)
				}
			}
			wbuf := rbytes.NewWBuffer(nil, nil, 0, nil)
			_, err := tc.want.MarshalROOT(wbuf)
			if err != nil {
				t.Fatalf("could not marshal ROOT: %v", err)
			}

			rbuf := rbytes.NewRBuffer(wbuf.Bytes(), nil, 0, nil)
			class := tc.want.Class()
			obj := rtypes.Factory.Get(class)().Interface().(rbytes.Unmarshaler)
			{
				rbuf.SetErr(io.EOF)
				err = obj.UnmarshalROOT(rbuf)
				if err == nil {
					t.Fatalf("expected an error")
				}
				if err != io.EOF {
					t.Fatalf("got=%v, want=%v", err, io.EOF)
				}
				rbuf.SetErr(nil)
			}
			err = obj.UnmarshalROOT(rbuf)
			if err != nil {
				t.Fatalf("could not unmarshal ROOT: %v", err)
			}

			if !reflect.DeepEqual(obj, tc.want) {
				t.Fatalf("error\ngot= %+v\nwant=%+v\n", obj, tc.want)
			}
		})
	}
}

func TestWriteWBuffer(t *testing.T) {
	for _, test := range rwBufferCases {
		t.Run("write-buffer="+test.file, func(t *testing.T) {
			testWriteWBuffer(t, test.name, test.file, test.want)
		})
	}
}

func testWriteWBuffer(t *testing.T, name, file string, want interface{}) {
	rdata, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	{
		wbuf := rbytes.NewWBuffer(nil, nil, 0, nil)
		wbuf.SetErr(io.EOF)
		_, err := want.(rbytes.Marshaler).MarshalROOT(wbuf)
		if err == nil {
			t.Fatalf("expected an error")
		}
		if err != io.EOF {
			t.Fatalf("got=%v, want=%v", err, io.EOF)
		}
	}

	w := rbytes.NewWBuffer(nil, nil, 0, nil)
	_, err = want.(rbytes.Marshaler).MarshalROOT(w)
	if err != nil {
		t.Fatal(err)
	}
	wdata := w.Bytes()

	r := rbytes.NewRBuffer(wdata, nil, 0, nil)
	obj := rtypes.Factory.Get(name)().Interface().(rbytes.Unmarshaler)
	{
		r.SetErr(io.EOF)
		err = obj.UnmarshalROOT(r)
		if err == nil {
			t.Fatalf("expected an error")
		}
		if err != io.EOF {
			t.Fatalf("got=%v, want=%v", err, io.EOF)
		}
		r.SetErr(nil)
	}
	err = obj.UnmarshalROOT(r)
	if err != nil {
		t.Fatal(err)
	}

	err = ioutil.WriteFile(file+".new", wdata, 0644)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(obj, want) {
		t.Fatalf("error: %q\ngot= %+v\nwant=%+v\ngot= %+v\nwant=%+v", file, wdata, rdata, obj, want)
	}

	os.Remove(file + ".new")
}
func TestReadRBuffer(t *testing.T) {
	for _, test := range rwBufferCases {
		test := test
		file := test.file
		if file == "" {
			file = "../testdata/" + strings.ToLower(test.name) + ".dat"
		}
		t.Run("read-buffer="+file, func(t *testing.T) {
			testReadRBuffer(t, test.name, file, test.want)
		})
	}
}

func testReadRBuffer(t *testing.T, name, file string, want interface{}) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		t.Fatal(err)
	}

	r := rbytes.NewRBuffer(data, nil, 0, nil)
	obj := rtypes.Factory.Get(name)().Interface().(rbytes.Unmarshaler)
	err = obj.UnmarshalROOT(r)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(obj, want) {
		t.Fatalf("error: %q\ngot= %#v\nwant=%#v\n", file, obj, want)
	}
}

var rwBufferCases = []struct {
	name string
	file string
	want rbytes.Unmarshaler
}{
	{
		name: "TList",
		file: "../testdata/tlist.dat",
		want: &List{
			obj:  rbase.Object{ID: 0x0, Bits: 0x3000000},
			name: "list-name",
			objs: []root.Object{
				rbase.NewNamed("n0", "t0"),
				rbase.NewNamed("n1", "t1"),
			},
		},
	},
	{
		name: "TObjArray",
		file: "../testdata/tobjarray.dat",
		want: &ObjArray{
			obj:  rbase.Object{ID: 0x0, Bits: 0x3000000},
			name: "my-objs",
			objs: []root.Object{
				rbase.NewNamed("n0", "t0"),
				rbase.NewNamed("n1", "t1"),
				rbase.NewNamed("n2", "t2"),
			},
			last: 2,
		},
	},
	{
		name: "TArrayI",
		file: "../testdata/tarrayi.dat",
		want: &ArrayI{Data: []int32{0, 1, 2, 3, 4}},
	},
	{
		name: "TArrayL64",
		file: "../testdata/tarrayl64.dat",
		want: &ArrayL64{Data: []int64{0, 1, 2, 3, 4}},
	},
	{
		name: "TArrayF",
		file: "../testdata/tarrayf.dat",
		want: &ArrayF{Data: []float32{0, 1, 2, 3, 4}},
	},
	{
		name: "TArrayD",
		file: "../testdata/tarrayd.dat",
		want: &ArrayD{Data: []float64{0, 1, 2, 3, 4}},
	},
}
