package qrimage_test

import (
	"bytes"
	"image"
	_ "image/png"
	"testing"

	"github.com/abinter/qr-code-generator/pkg/qrimage"
)

func TestGenerate_ReturnsValidPNG(t *testing.T) {
	opts := qrimage.Options{
		Content:   "https://myqrcode.com/abc123",
		Dimension: 256,
		Color:     "000000",
		Border:    4,
	}

	data, err := qrimage.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Generate() returned empty bytes")
	}

	// Verify it's a valid PNG
	_, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		t.Fatalf("output is not a valid image: %v", err)
	}
	if format != "png" {
		t.Errorf("format = %q; want png", format)
	}
}

func TestGenerate_DimensionAffectsSize(t *testing.T) {
	base := qrimage.Options{Content: "https://example.com", Color: "000000", Border: 0}

	small := base
	small.Dimension = 100
	large := base
	large.Dimension = 400

	smallPNG, _ := qrimage.Generate(small)
	largePNG, _ := qrimage.Generate(large)

	smallImg, _, _ := image.Decode(bytes.NewReader(smallPNG))
	largeImg, _, _ := image.Decode(bytes.NewReader(largePNG))

	if smallImg.Bounds().Dx() >= largeImg.Bounds().Dx() {
		t.Errorf("small (%d) should be narrower than large (%d)",
			smallImg.Bounds().Dx(), largeImg.Bounds().Dx())
	}
}

func TestGenerate_BorderIsIncluded(t *testing.T) {
	noBorder := qrimage.Options{Content: "https://example.com", Dimension: 200, Color: "000000", Border: 0}
	withBorder := qrimage.Options{Content: "https://example.com", Dimension: 200, Color: "000000", Border: 20}

	noPNG, _ := qrimage.Generate(noBorder)
	borderPNG, _ := qrimage.Generate(withBorder)

	noImg, _, _ := image.Decode(bytes.NewReader(noPNG))
	borderImg, _, _ := image.Decode(bytes.NewReader(borderPNG))

	// Both are the same total dimension; border shrinks the inner QR code
	if noImg.Bounds().Dx() != borderImg.Bounds().Dx() {
		t.Errorf("total dimension changed: no-border=%d with-border=%d",
			noImg.Bounds().Dx(), borderImg.Bounds().Dx())
	}
}

func TestGenerate_InvalidColor(t *testing.T) {
	opts := qrimage.Options{
		Content:   "https://example.com",
		Dimension: 256,
		Color:     "ZZZZZZ",
		Border:    0,
	}
	_, err := qrimage.Generate(opts)
	if err == nil {
		t.Error("expected error for invalid color, got nil")
	}
}

func TestGenerate_DefaultsApplied(t *testing.T) {
	// Zero-value options should produce a valid image with defaults
	opts := qrimage.Options{Content: "https://example.com"}
	data, err := qrimage.Generate(opts)
	if err != nil {
		t.Fatalf("Generate() with defaults error: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("Generate() with defaults returned empty bytes")
	}
}
