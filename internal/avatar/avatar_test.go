package avatar

import (
	"bytes"
	"image/png"
	"testing"
)

func TestGenerate(t *testing.T) {
	raw, err := Generate("red", "#fff", "тест")
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	img, err := png.Decode(bytes.NewReader(raw))
	if err != nil {
		t.Fatalf("decode png: %v", err)
	}
	bounds := img.Bounds()
	if bounds.Dx() != imageSize || bounds.Dy() != imageSize {
		t.Errorf("size = %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), imageSize, imageSize)
	}

	// The background corner must be red.
	r, g, b, _ := img.At(0, 0).RGBA()
	if r>>8 != 255 || g>>8 != 0 || b>>8 != 0 {
		t.Errorf("corner color = %d %d %d, want red", r>>8, g>>8, b>>8)
	}
}

func TestGenerateBadInputs(t *testing.T) {
	if _, err := Generate("notacolor", "red", "ник"); err == nil {
		t.Error("unknown background color must fail")
	}
	if _, err := Generate("red", "#zzz", "ник"); err == nil {
		t.Error("broken hex must fail")
	}
}

func TestParseColor(t *testing.T) {
	c, err := ParseColor("#1a2b3c")
	if err != nil || c.R != 0x1a || c.G != 0x2b || c.B != 0x3c {
		t.Errorf("hex6 = %+v, err=%v", c, err)
	}
	c, err = ParseColor("#f0a")
	if err != nil || c.R != 0xff || c.G != 0x00 || c.B != 0xaa {
		t.Errorf("hex3 = %+v, err=%v", c, err)
	}
	if _, err := ParseColor("Salmon"); err != nil {
		t.Errorf("named colors must be case-insensitive: %v", err)
	}
}
