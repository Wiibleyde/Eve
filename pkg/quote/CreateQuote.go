package quote

import (
	"errors"
	"fmt"
	"image"
	"image/color"
	"image/gif"    // Import GIF package
	_ "image/jpeg" // Register JPEG format
	_ "image/png"  // Register PNG format
	"main/pkg/logger"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/fogleman/gg"
)

const (
	smokePath           = "assets/img/smoke.png"
	boldFontPath        = "assets/fonts/Montserrat-Bold.ttf"
	lightItalicFontPath = "assets/fonts/Montserrat-LightItalic.ttf"
	semiBoldFontPath    = "assets/fonts/Montserrat-SemiBold.ttf"
	quoteMaxWidth       = 550
	maxQuoteLength      = 200
	maxContextLength    = 200
)

func CreateQuote(
	quote string,
	author string,
	context string,
	date string,
	userProfilePictureURL string,
) (string, error) {
	logger.InfoLogger.Println("Creating quote image...")

	// Delete the previous quote image if it exists
	err := os.Remove("assets/quote_output.png")
	if err != nil && !os.IsNotExist(err) {
		logger.ErrorLogger.Println("Error deleting previous quote image:", err)
		return "", err
	}

	// Input validation
	if quote == "" {
		return "", errors.New("quote cannot be empty")
	}
	if len(quote) > maxQuoteLength {
		return "", fmt.Errorf("quote cannot exceed %d characters", maxQuoteLength)
	}
	if context != "" && len(context) > maxContextLength {
		return "", fmt.Errorf("context cannot exceed %d characters", maxContextLength)
	}

	// Define new dimensions
	const bgWidth, bgHeight = 1012, 412

	// Create a new context with the specified dimensions
	dc := gg.NewContext(bgWidth, bgHeight)

	// Draw the black background
	dc.SetRGB(0, 0, 0) // Black color
	dc.Clear()

	// Draw white borders
	borderWidth := 10.0
	dc.SetRGB(1, 1, 1) // White color
	dc.SetLineWidth(borderWidth)
	dc.DrawRectangle(borderWidth/2, borderWidth/2, float64(bgWidth)-borderWidth, float64(bgHeight)-borderWidth)
	dc.Stroke()

	// Add profile picture if URL is provided
	if userProfilePictureURL != "" {
		profilePic, err := downloadImage(userProfilePictureURL)
		if err != nil {
			logger.ErrorLogger.Printf("Error downloading profile picture: %v, continuing without it", err)
		} else {
			// Calculate the maximum size for the profile picture as a square
			maxPictureSize := bgHeight - 20 // 20px total margin (10px top + 10px bottom)

			// Resize the profile picture to fit within the square dimensions
			profilePic = resizeImage(profilePic, maxPictureSize, maxPictureSize)

			// Darken the profile picture
			profilePic = darkenImage(profilePic, 0.5)

			// Draw the profile picture centered vertically and aligned to the left
			dc.DrawImage(profilePic, 10, (bgHeight-maxPictureSize)/2) // 10px margin from the left
		}
	}

	// Load and apply smoke overlay
	smokeImg, err := loadSmokeImage(smokePath, bgWidth)
	if err != nil {
		logger.ErrorLogger.Printf("Error loading smoke image: %v, continuing without it", err)
	} else {
		// Calculate position to place smoke at the bottom of the image
		smokeHeight := smokeImg.Bounds().Dy()
		// Draw the smoke image at the bottom of the background
		dc.DrawImage(smokeImg, 0, bgHeight-smokeHeight)
	}

	// Set font for quote with bold font 48
	quoteFace, err := gg.LoadFontFace(boldFontPath, 48)
	if err != nil {
		logger.ErrorLogger.Println("Error loading font face for quote:", err)
		return "", err
	}
	dc.SetFontFace(quoteFace)

	// Set the text color
	dc.SetRGB(1, 1, 1)

	// Wrap and draw the quote text
	quoteWithQuotes := "\"" + quote + "\""
	wrappedQuote := wrapText(dc, quoteWithQuotes, quoteMaxWidth)

	// Calculate quote text height
	quoteHeight := measureTextHeight(dc, wrappedQuote, quoteMaxWidth)

	// Calculate vertical centering, but ensure there's enough space for context
	// If quote is very long, position it higher to leave space for context
	maxQuoteSpace := float64(bgHeight) * 0.7 // Max 70% of height for quote

	quoteYPosition := (float64(bgHeight) - quoteHeight) / 2
	if quoteHeight > maxQuoteSpace {
		// If quote is too large, align to top with some margin
		quoteYPosition = float64(bgHeight) * 0.15 // Start at 15% from top
	}

	dc.DrawStringWrapped(wrappedQuote, 350, quoteYPosition, 0, 0, quoteMaxWidth, 1.2, gg.AlignLeft)

	// Handle context if provided
	if context != "" {
		contextFace, err := gg.LoadFontFace(lightItalicFontPath, 24) // Use italic font for context
		if err != nil {
			logger.ErrorLogger.Println("Error loading font face for context:", err)
			return "", err
		}
		dc.SetFontFace(contextFace)

		// Position context directly below the quote with a small gap
		contextYPosition := quoteYPosition + quoteHeight + 20 // 20px gap after the quote
		dc.DrawStringWrapped(context, 350, contextYPosition, 0, 0, quoteMaxWidth, 1.2, gg.AlignLeft)
	}

	// Add author and date
	authorFace, err := gg.LoadFontFace(semiBoldFontPath, 28) // Use semibold font for author
	if err != nil {
		logger.ErrorLogger.Println("Error loading font face for author:", err)
		return "", err
	}
	dc.SetFontFace(authorFace)

	// Format date if not provided
	if date == "" {
		now := time.Now()
		date = now.Format("02/01/2006")
	}

	authorText := "@" + author + " - " + date
	if author == "" {
		authorText = "Anonyme - " + date
	}

	authorXPosition := float64(bgWidth) - 40 - measureTextWidth(dc, authorText)
	dc.DrawString(authorText, authorXPosition, 360)

	// Save the image to a file
	outputPath := "assets/quote_output.png"
	err = dc.SavePNG(outputPath)
	if err != nil {
		logger.ErrorLogger.Println("Error saving quote image:", err)
		return "", err
	}

	logger.InfoLogger.Println("Quote image created successfully:", outputPath)

	// Return the path to the generated image
	return outputPath, nil
}

// wrapText handles text wrapping for long quotes
func wrapText(dc *gg.Context, text string, maxWidth float64) string {
	words := strings.Fields(text)
	if len(words) == 0 {
		return ""
	}

	var lines []string
	line := words[0]

	for _, word := range words[1:] {
		testLine := line + " " + word
		width, _ := dc.MeasureString(testLine)

		if width <= maxWidth {
			line = testLine
		} else {
			lines = append(lines, line)
			line = word
		}
	}

	lines = append(lines, line)
	return strings.Join(lines, "\n")
}

// downloadImage downloads an image from the given URL and returns it as an image.Image
func downloadImage(url string) (image.Image, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to download image, status code: %d", resp.StatusCode)
	}

	// Check if the URL ends with .gif
	isGif := strings.HasSuffix(strings.ToLower(url), ".gif")

	// Handle potential GIF format differently
	if isGif {
		logger.InfoLogger.Println("Processing GIF image from URL:", url)
		gifImg, err := gif.DecodeAll(resp.Body)
		if err != nil {
			// If we can't decode as GIF, try as regular image
			logger.ErrorLogger.Printf("Error decoding GIF, trying as normal image: %v", err)
			// We need to get a new response since the body was consumed
			resp, err = http.Get(url)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		} else {
			// Use the first frame of the GIF
			if len(gifImg.Image) > 0 {
				return gifImg.Image[0], nil
			}
			logger.ErrorLogger.Println("GIF had no frames")
			// We need to get a new response to try as regular image
			resp, err = http.Get(url)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()
		}
	}

	// Regular image decode
	img, _, err := image.Decode(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %v", err)
	}

	return img, nil
}

// Helper function to resize an image
func resizeImage(img image.Image, width, height int) image.Image {
	// Create a new RGBA image with the desired dimensions
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Use gg to draw the resized image
	dc := gg.NewContextForImage(dst)
	dc.DrawImageAnchored(img, 0, 0, 0, 0)
	dc.Scale(float64(width)/float64(img.Bounds().Dx()), float64(height)/float64(img.Bounds().Dy()))
	dc.DrawImage(img, 0, 0)

	return dc.Image()
}

// Helper function to darken an image by drawing a transparent black rectangle over it
func darkenImage(img image.Image, opacity float64) image.Image {
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)

	// Draw the original image onto the RGBA canvas
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			rgba.Set(x, y, img.At(x, y))
		}
	}

	// Draw a transparent black rectangle over the image
	overlay := color.RGBA{0, 0, 0, uint8(255 * opacity)}
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			orig := rgba.RGBAAt(x, y)
			// Alpha blending
			a := float64(overlay.A) / 255.0
			r := float64(orig.R)*(1-a) + float64(overlay.R)*a
			g := float64(orig.G)*(1-a) + float64(overlay.G)*a
			b := float64(orig.B)*(1-a) + float64(overlay.B)*a
			alpha := float64(orig.A)
			rgba.SetRGBA(x, y, color.RGBA{
				R: uint8(r),
				G: uint8(g),
				B: uint8(b),
				A: uint8(alpha),
			})
		}
	}
	return rgba
}

// Helper function to measure text height
func measureTextHeight(dc *gg.Context, text string, maxWidth float64) float64 {
	lines := wrapText(dc, text, maxWidth)
	return float64(len(strings.Split(lines, "\n")))*dc.FontHeight() + 10 // 10px margin
}

// Helper function to measure text width
func measureTextWidth(dc *gg.Context, text string) float64 {
	width, _ := dc.MeasureString(text)
	return width
}

// Helper function to load and resize the smoke image
func loadSmokeImage(path string, targetWidth int) (image.Image, error) {
	// Open the smoke image file
	smokeFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open smoke image: %v", err)
	}
	defer smokeFile.Close()

	// Decode the smoke image
	smokeImg, _, err := image.Decode(smokeFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode smoke image: %v", err)
	}

	// Get the current dimensions
	smokeWidth := smokeImg.Bounds().Dx()
	smokeHeight := smokeImg.Bounds().Dy()

	// Calculate new height while maintaining aspect ratio
	newHeight := smokeHeight * targetWidth / smokeWidth

	// Resize the smoke image to match the target width
	resizedSmoke := resizeImage(smokeImg, targetWidth, newHeight)

	return resizedSmoke, nil
}
