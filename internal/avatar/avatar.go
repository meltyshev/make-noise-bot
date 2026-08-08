// Package avatar renders avatars: a colored square with a nickname.
package avatar

import (
	"bytes"
	_ "embed"
	"errors"
	"image"
	"image/draw"
	"image/png"

	"golang.org/x/image/font"
	"golang.org/x/image/font/opentype"
	"golang.org/x/image/math/fixed"
)

//go:embed assets/Roboto-Bold.ttf
var robotoBold []byte

const (
	imageSize    = 640
	fitRatio     = 0.8
	paddingRatio = 0.1
)

// Generate renders a square avatar. Colors accept #rgb, #rrggbb or CSS names.
func Generate(background, foreground, nickname string) ([]byte, error) {
	bg, err := ParseColor(background)
	if err != nil {
		return nil, err
	}
	fg, err := ParseColor(foreground)
	if err != nil {
		return nil, err
	}

	ft, err := opentype.Parse(robotoBold)
	if err != nil {
		return nil, err
	}

	// Measure at full size first, then scale the font to fit fitRatio of
	// the width.
	baseFace, err := opentype.NewFace(ft, &opentype.FaceOptions{Size: imageSize, DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, err
	}
	baseWidth := font.MeasureString(baseFace, nickname).Round()
	baseFace.Close()
	if baseWidth <= 0 {
		return nil, errors.New("nickname has no drawable glyphs")
	}

	fontSize := max(int(imageSize*fitRatio/float64(baseWidth)*imageSize), 1)

	face, err := opentype.NewFace(ft, &opentype.FaceOptions{Size: float64(fontSize), DPI: 72, Hinting: font.HintingFull})
	if err != nil {
		return nil, err
	}
	defer face.Close()

	width := font.MeasureString(face, nickname).Round()
	metrics := face.Metrics()
	height := (metrics.Ascent + metrics.Descent).Round()
	padding := int(float64(fontSize) * paddingRatio)

	left := (imageSize - width) / 2
	top := (imageSize-height)/2 + padding

	img := image.NewRGBA(image.Rect(0, 0, imageSize, imageSize))
	draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)

	drawer := font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(fg),
		Face: face,
		Dot:  fixed.P(left, top+metrics.Ascent.Round()),
	}
	drawer.DrawString(nickname)

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
