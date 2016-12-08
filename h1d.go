// Copyright ©2016 The go-hep Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package hplot

import (
	"errors"
	"fmt"
	"image/color"

	"github.com/go-hep/hbook"
	"github.com/gonum/plot"
	"github.com/gonum/plot/plotter"
	"github.com/gonum/plot/vg"
	"github.com/gonum/plot/vg/draw"
)

// Histogram implements the plotter.Plotter interface,
// drawing a histogram of the data.
type Histogram struct {
	// Hist is the histogramming data
	Hist *hbook.H1D

	// FillColor is the color used to fill each
	// bar of the histogram.  If the color is nil
	// then the bars are not filled.
	FillColor color.Color

	// LineStyle is the style of the outline of each
	// bar of the histogram.
	draw.LineStyle

	// InfoStyle is the style of infos displayed for
	// the histogram (entries, mean, rms)
	Infos HInfos
}

type HInfoStyle int

const (
	HInfoNone    HInfoStyle = 0
	HInfoEntries HInfoStyle = iota << 1
	HInfoMean
	HInfoRMS
	HInfoSummary // HInfoEntries | HInfoMean | HInfoRMS
)

type HInfos struct {
	Style HInfoStyle
}

// NewHistogram returns a new histogram
// that represents the distribution of values
// using the given number of bins.
//
// Each y value is assumed to be the frequency
// count for the corresponding x.
//
// If the number of bins is non-positive than
// a reasonable default is used.
func NewHistogram(xy plotter.XYer, n int) (*Histogram, error) {
	if n <= 0 {
		return nil, errors.New("Histogram with non-positive number of bins")
	}
	h := newHistFromXYer(xy, n)
	return NewH1D(h)
}

// NewHist returns a new histogram, as in
// NewHistogram, except that it accepts a plotter.Valuer
// instead of an XYer.
func NewHist(vs plotter.Valuer, n int) (*Histogram, error) {
	return NewHistogram(unitYs{vs}, n)
}

type unitYs struct {
	plotter.Valuer
}

func (u unitYs) XY(i int) (float64, float64) {
	return u.Value(i), 1.0
}

// NewH1D returns a new histogram, as in
// NewHistogram, except that it accepts a hbook.H1D
// instead of a plotter.XYer
func NewH1D(h *hbook.H1D) (*Histogram, error) {
	return &Histogram{
		Hist:      h,
		FillColor: color.White,
		LineStyle: plotter.DefaultLineStyle,
	}, nil
}

// DataRange returns the minimum and maximum X and Y values
func (h *Histogram) DataRange() (xmin, xmax, ymin, ymax float64) {
	return h.Hist.DataRange()
}

// Plot implements the Plotter interface, drawing a line
// that connects each point in the Line.
func (h *Histogram) Plot(c draw.Canvas, p *plot.Plot) {
	trX, trY := p.Transforms(&c)
	var pts []vg.Point
	hist := h.Hist
	axis := hist.Binning()
	nbins := int(axis.Bins())
	for bin := 0; bin < nbins; bin++ {
		switch bin {
		case 0:
			pts = append(pts, vg.Point{trX(axis.BinLowerEdge(bin)), trY(0)})
			pts = append(pts, vg.Point{trX(axis.BinLowerEdge(bin)), trY(hist.Value(bin))})
			pts = append(pts, vg.Point{trX(axis.BinUpperEdge(bin)), trY(hist.Value(bin))})

		case nbins - 1:
			pts = append(pts, vg.Point{trX(axis.BinUpperEdge(bin - 1)), trY(hist.Value(bin - 1))})
			pts = append(pts, vg.Point{trX(axis.BinLowerEdge(bin)), trY(hist.Value(bin))})
			pts = append(pts, vg.Point{trX(axis.BinUpperEdge(bin)), trY(hist.Value(bin))})
			pts = append(pts, vg.Point{trX(axis.BinUpperEdge(bin)), trY(0.)})

		default:
			pts = append(pts, vg.Point{trX(axis.BinUpperEdge(bin - 1)), trY(hist.Value(bin - 1))})
			pts = append(pts, vg.Point{trX(axis.BinLowerEdge(bin)), trY(hist.Value(bin))})
			pts = append(pts, vg.Point{trX(axis.BinUpperEdge(bin)), trY(hist.Value(bin))})
		}
	}

	if h.FillColor != nil {
		c.FillPolygon(h.FillColor, c.ClipPolygonXY(pts))
	}
	c.StrokeLines(h.LineStyle, c.ClipLinesXY(pts)...)

	if h.Infos.Style != HInfoNone {
		fnt, err := vg.MakeFont(plotter.DefaultFont, plotter.DefaultFontSize)
		if err == nil {
			sty := draw.TextStyle{Font: fnt}
			legend := hist_legend{
				ColWidth:  plotter.DefaultFontSize,
				TextStyle: sty,
			}

			switch h.Infos.Style {
			case HInfoSummary:
				legend.Add("Entries", hist.Entries())
				legend.Add("Mean", hist.Mean())
				legend.Add("RMS", hist.RMS())
			case HInfoEntries:
				legend.Add("Entries", hist.Entries())
			case HInfoMean:
				legend.Add("Mean", hist.Mean())
			case HInfoRMS:
				legend.Add("RMS", hist.RMS())
			default:
			}
			legend.Top = true

			legend.draw(c)
		}
	}
}

// GlyphBoxes returns a slice of GlyphBoxes,
// one for each of the bins, implementing the
// plot.GlyphBoxer interface.
func (h *Histogram) GlyphBoxes(p *plot.Plot) []plot.GlyphBox {
	axis := h.Hist.Binning()
	bs := make([]plot.GlyphBox, axis.Bins())
	for i, _ := range bs {
		y := h.Hist.Value(i)
		xmin := axis.BinLowerEdge(i)
		w := p.X.Norm(axis.BinWidth(i))
		bs[i].X = p.X.Norm(xmin + 0.5*w)
		bs[i].Y = p.Y.Norm(y)
		//h := vg.Points(1e-5) //1 //p.Y.Norm(axis.BinWidth(i))
		bs[i].Rectangle.Min.X = vg.Length(xmin - 0.5*w)
		bs[i].Rectangle.Min.Y = vg.Length(y - 0.5*w)
		bs[i].Rectangle.Max.X = vg.Length(w)
		bs[i].Rectangle.Max.Y = vg.Length(0)

		r := vg.Points(5)
		//r = vg.Length(w)
		bs[i].Rectangle.Min = vg.Point{0, 0}
		bs[i].Rectangle.Max = vg.Point{0, r}
	}
	return bs
}

// Normalize normalizes the histogram so that the
// total area beneath it sums to a given value.
// func (h *Histogram) Normalize(sum float64) {
// 	mass := 0.0
// 	for _, b := range h.Bins {
// 		mass += b.Weight
// 	}
// 	for i := range h.Bins {
// 		h.Bins[i].Weight *= sum / (h.Width * mass)
// 	}
// }

func newHistFromXYer(xys plotter.XYer, n int) *hbook.H1D {
	xmin, xmax := plotter.Range(plotter.XValues{xys})
	h := hbook.NewH1D(n, xmin, xmax)

	for i := 0; i < xys.Len(); i++ {
		x, y := xys.XY(i)
		h.Fill(x, y)
	}

	return h
}

// A Legend gives a description of the meaning of different
// data elements of the plot.  Each legend entry has a name
// and a thumbnail, where the thumbnail shows a small
// sample of the display style of the corresponding data.
type hist_legend struct {
	// TextStyle is the style given to the legend
	// entry texts.
	draw.TextStyle

	// Padding is the amount of padding to add
	// betweeneach entry of the legend.  If Padding
	// is zero then entries are spaced based on the
	// font size.
	Padding vg.Length

	// Top and Left specify the location of the legend.
	// If Top is true the legend is located along the top
	// edge of the plot, otherwise it is located along
	// the bottom edge.  If Left is true then the legend
	// is located along the left edge of the plot, and the
	// text is positioned after the icons, otherwise it is
	// located along the right edge and the text is
	// positioned before the icons.
	Top, Left bool

	// XOffs and YOffs are added to the legend's
	// final position.
	XOffs, YOffs vg.Length

	// ColWidth is the width of legend names
	ColWidth vg.Length

	// entries are all of the legendEntries described
	// by this legend.
	entries []legendEntry
}

// A legendEntry represents a single line of a legend, it
// has a name and an icon.
type legendEntry struct {
	// text is the text associated with this entry.
	text string

	// value is the value associated with this entry
	value string
}

// draw draws the legend to the given canvas.
func (l *hist_legend) draw(c draw.Canvas) {
	textx := c.Min.X
	hdr := l.entryWidth() //+ l.TextStyle.Width(" ")
	l.ColWidth = hdr
	valx := textx + l.ColWidth + l.TextStyle.Width(" ")
	if !l.Left {
		textx = c.Max.X - l.ColWidth
		valx = textx - l.TextStyle.Width(" ")
	}
	valx += l.XOffs
	textx += l.XOffs

	enth := l.entryHeight()
	y := c.Max.Y - enth
	if !l.Top {
		y = c.Min.Y + (enth+l.Padding)*(vg.Length(len(l.entries))-1)
	}
	y += l.YOffs

	colx := &draw.Canvas{
		Canvas: c.Canvas,
		Rectangle: vg.Rectangle{
			Min: vg.Point{c.Min.X, y},
			Max: vg.Point{2 * l.ColWidth, enth},
		},
	}
	for _, e := range l.entries {
		yoffs := (enth - l.TextStyle.Height(e.text)) / 2
		txt := l.TextStyle
		txt.XAlign = draw.XLeft
		c.FillText(txt, vg.Point{textx - hdr, colx.Min.Y + yoffs}, e.text)
		txt.XAlign = draw.XRight
		c.FillText(txt, vg.Point{textx + hdr, colx.Min.Y + yoffs}, e.value)
		colx.Min.Y -= enth + l.Padding
	}

	bbox_xmin := textx - hdr - l.TextStyle.Width(" ")
	bbox_xmax := c.Max.X
	bbox_ymin := colx.Min.Y + enth
	bbox_ymax := c.Max.Y
	bbox := []vg.Point{
		{bbox_xmin, bbox_ymax},
		{bbox_xmin, bbox_ymin},
		{bbox_xmax, bbox_ymin},
		{bbox_xmax, bbox_ymax},
		{bbox_xmin, bbox_ymax},
	}
	c.StrokeLines(plotter.DefaultLineStyle, bbox)
}

// entryHeight returns the height of the tallest legend
// entry text.
func (l *hist_legend) entryHeight() (height vg.Length) {
	for _, e := range l.entries {
		if h := l.TextStyle.Height(e.text); h > height {
			height = h
		}
	}
	return
}

// entryWidth returns the width of the largest legend
// entry text.
func (l *hist_legend) entryWidth() (width vg.Length) {
	for _, e := range l.entries {
		if w := l.TextStyle.Width(e.value); w > width {
			width = w
		}
	}
	return
}

// Add adds an entry to the legend with the given name.
// The entry's thumbnail is drawn as the composite of all of the
// thumbnails.
func (l *hist_legend) Add(name string, value interface{}) {
	str := ""
	switch value.(type) {
	case float64, float32:
		str = fmt.Sprintf("%6.4g ", value)
	default:
		str = fmt.Sprintf("%v ", value)
	}
	l.entries = append(l.entries, legendEntry{text: name, value: str})
}
