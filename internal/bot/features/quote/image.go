package quote

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	_ "image/gif"
	_ "image/jpeg"
	"image/png"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"Eve/internal/logger"

	"github.com/fogleman/gg"
)

const (
	smokePath           = "assets/img/smoke.png"
	boldFontPath        = "assets/fonts/Montserrat-Bold.ttf"
	lightItalicFontPath = "assets/fonts/Montserrat-LightItalic.ttf"
	semiBoldFontPath    = "assets/fonts/Montserrat-SemiBold.ttf"
	quoteMaxWidth       = 550.0
	bgWidth             = 1012
	bgHeight            = 412
)

func generateImage(quoteText, author, context, avatarURL string) ([]byte, error) {
	dc := gg.NewContext(bgWidth, bgHeight)

	dc.SetRGB(0, 0, 0)
	dc.Clear()

	borderWidth := 10.0
	dc.SetRGB(1, 1, 1)
	dc.SetLineWidth(borderWidth)
	dc.DrawRectangle(borderWidth/2, borderWidth/2, float64(bgWidth)-borderWidth, float64(bgHeight)-borderWidth)
	dc.Stroke()

	if avatarURL != "" {
		logger.Info("downloading avatar", "url", avatarURL)
		profilePic, err := downloadImage(avatarURL)
		if err != nil {
			logger.Error("download avatar failed", "error", err)
		} else {
			logger.Info("avatar decoded", "type", fmt.Sprintf("%T", profilePic), "bounds", profilePic.Bounds())
			maxSize := bgHeight - 20
			profilePic = resizeImage(profilePic, maxSize, maxSize)
			profilePic = darkenImage(profilePic, 0.5)
			dc.DrawImage(profilePic, 10, (bgHeight-maxSize)/2)
		}
	}

	if smokeImg, err := loadSmokeImage(smokePath, bgWidth); err == nil {
		dc.DrawImage(smokeImg, 0, bgHeight-smokeImg.Bounds().Dy())
	}

	quoteFace, err := gg.LoadFontFace(boldFontPath, 48)
	if err != nil {
		return nil, fmt.Errorf("load bold font: %w", err)
	}
	dc.SetFontFace(quoteFace)
	dc.SetRGB(1, 1, 1)

	quoteWithQuotes := "\"" + quoteText + "\""
	wrappedQuote := wrapText(dc, quoteWithQuotes, quoteMaxWidth)
	quoteH := measureTextHeight(dc, wrappedQuote)

	quoteY := (float64(bgHeight) - quoteH) / 2
	if quoteH > float64(bgHeight)*0.7 {
		quoteY = float64(bgHeight) * 0.15
	}

	dc.DrawStringWrapped(wrappedQuote, 350, quoteY, 0, 0, quoteMaxWidth, 1.2, gg.AlignLeft)

	if context != "" {
		ctxFace, err := gg.LoadFontFace(lightItalicFontPath, 24)
		if err != nil {
			return nil, fmt.Errorf("load italic font: %w", err)
		}
		dc.SetFontFace(ctxFace)
		dc.DrawStringWrapped(context, 350, quoteY+quoteH+20, 0, 0, quoteMaxWidth, 1.2, gg.AlignLeft)
	}

	authorFace, err := gg.LoadFontFace(semiBoldFontPath, 28)
	if err != nil {
		return nil, fmt.Errorf("load semibold font: %w", err)
	}
	dc.SetFontFace(authorFace)

	authorText := "@" + author + " - " + time.Now().Format("02/01/2006")
	authorX := float64(bgWidth) - 40 - measureTextWidth(dc, authorText)
	dc.DrawString(authorText, authorX, 360)

	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("encode png: %w", err)
	}
	return buf.Bytes(), nil
}

func wrapText(dc *gg.Context, text string, maxWidth float64) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}
	line := words[0]
	var lines []string
	for _, word := range words[1:] {
		if w, _ := dc.MeasureString(line + " " + word); w <= maxWidth {
			line += " " + word
		} else {
			lines = append(lines, line)
			line = word
		}
	}
	return strings.Join(append(lines, line), "\n")
}

func measureTextHeight(dc *gg.Context, text string) float64 {
	return float64(len(strings.Split(text, "\n")))*dc.FontHeight() + 10
}

func measureTextWidth(dc *gg.Context, text string) float64 {
	w, _ := dc.MeasureString(text)
	return w
}

func downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	img, _, err := image.Decode(bytes.NewReader(data))
	return img, err
}

func resizeImage(img image.Image, width, height int) image.Image {
	bounds := img.Bounds()
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	scaleX := float64(bounds.Dx()) / float64(width)
	scaleY := float64(bounds.Dy()) / float64(height)
	for y := range height {
		for x := range width {
			srcX := bounds.Min.X + int(float64(x)*scaleX)
			srcY := bounds.Min.Y + int(float64(y)*scaleY)
			dst.Set(x, y, img.At(srcX, srcY))
		}
	}
	return dst
}

func darkenImage(img image.Image, opacity float64) image.Image {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			r32, g32, b32, a32 := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			dst.SetRGBA(x, y, color.RGBA{
				R: uint8(float64(r32>>8) * (1 - opacity)),
				G: uint8(float64(g32>>8) * (1 - opacity)),
				B: uint8(float64(b32>>8) * (1 - opacity)),
				A: uint8(a32 >> 8),
			})
		}
	}
	return dst
}

func loadSmokeImage(path string, targetWidth int) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	smokeImg, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	sw, sh := smokeImg.Bounds().Dx(), smokeImg.Bounds().Dy()
	return resizeImage(smokeImg, targetWidth, sh*targetWidth/sw), nil
}
