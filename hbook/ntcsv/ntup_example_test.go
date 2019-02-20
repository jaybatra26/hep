// Copyright 2019 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ntcsv_test

import (
	"fmt"
	"log"
)

func ExampleNtuple_scanH2D() {

	nt, err := ntcsv.Open(
		"ntcsv/testdata/simple-with-header.csv",
		ntcsv.Comma(';'),
		ntcsv.Header(),
		ntcsv.Columns("v1", "v2", "v3"),
	)
	if err != nil {
		log.Fatal(err)
	}

	defer nt.DB().Close()

	got, err := nt.ScanH2D("v1, v2", nil)

	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("XMean:      %f\n", got.XMean())
	fmt.Printf("YMean:      %f\n", got.YMean())
	fmt.Printf("XRMS:       %f\n", got.XRMS())
	fmt.Printf("YRMS:       %f\n", got.YRMS())
	fmt.Printf("XStdDev:    %f\n", got.XStdDev())
	fmt.Printf("YStdDev:    %f\n", got.YStdDev())
	fmt.Printf("XStdErr:    %f\n", got.XStdErr())
	fmt.Printf("YStdErr:    %f\n", got.YStdErr())

	// Output:
	// XMean:      4.500000
	// YMean:      4.500000
	// XRMS:       5.338539
	// YRMS:       5.338539
	// XStdDev:    3.027650
	// YStdDev:    3.027650
	// XStdErr:    0.957427
	// YStdErr:    0.957427
}
