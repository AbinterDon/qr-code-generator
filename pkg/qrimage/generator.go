package qrimage

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"

	qrcode "github.com/skip2/go-qrcode"
)

const (
	defaultDimension = 256
	defaultColor     = "000000"
)

// Options controls how the QR code image is generated.
type Options struct {
	Content   string
	Dimension int    // total image size in pixels
	Color     string // foreground color as hex (e.g. "000000")
	Border    int    // border width in pixels on each side
}

// Generate produces a PNG-encoded QR code image.
// The QR code encodes Content; total image size is Dimension×Dimension,
// with Border pixels of white padding on each side.
func Generate(opts Options) ([]byte, error) {
	if opts.Dimension == 0 {
		opts.Dimension = defaultDimension
	}
	if opts.Color == "" {
		opts.Color = defaultColor
	}
	if opts.Border < 0 {
		opts.Border = 0
	}

	fg, err := parseHexColor(opts.Color)
	if err != nil {
		return nil, fmt.Errorf("invalid color %q: %w", opts.Color, err)
	}

	// Inner QR code size after subtracting border on both sides
	innerSize := max(opts.Dimension-2*opts.Border, 21)

	qr, err := qrcode.New(opts.Content, qrcode.Medium)
	if err != nil {
		return nil, fmt.Errorf("create qr code: %w", err)
	}
	qr.DisableBorder = true
	qr.ForegroundColor = fg
	qr.BackgroundColor = color.White

	qrPNG, err := qr.PNG(innerSize)
	if err != nil {
		return nil, fmt.Errorf("render qr png: %w", err)
	}

	qrImg, err := png.Decode(bytes.NewReader(qrPNG))
	if err != nil {
		return nil, fmt.Errorf("decode qr png: %w", err)
	}

	// Create final canvas with border
	canvas := image.NewRGBA(image.Rect(0, 0, opts.Dimension, opts.Dimension))
	draw.Draw(canvas, canvas.Bounds(), &image.Uniform{color.White}, image.Point{}, draw.Src)
	offset := image.Pt(opts.Border, opts.Border)
	draw.Draw(canvas, qrImg.Bounds().Add(offset), qrImg, image.Point{}, draw.Over)

	var buf bytes.Buffer
	if err := png.Encode(&buf, canvas); err != nil {
		return nil, fmt.Errorf("encode final png: %w", err)
	}
	return buf.Bytes(), nil
}

// parseHexColor parses a 6-character hex color string (e.g. "1a2b3c") into color.RGBA.
func parseHexColor(s string) (color.RGBA, error) {
	if len(s) != 6 {
		return color.RGBA{}, fmt.Errorf("expected 6 hex chars, got %d", len(s))
	}
	b, err := hex.DecodeString(s)
	if err != nil {
		return color.RGBA{}, err
	}
	return color.RGBA{R: b[0], G: b[1], B: b[2], A: 255}, nil
}
