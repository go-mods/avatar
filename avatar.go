package avatar

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	svg "github.com/ajstarks/svgo"
	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
)

// avatarGenerator is a struct to hold the options for the avatar generator
type avatarGenerator struct {
	// Initials is the initials to generate the svg from
	Initials string

	// Width is the width of the svg
	Width int

	// Height is the height of the svg
	Height int

	// BackgroundColor is the background color of the svg
	// The background color must be a valid hex color code
	BackgroundColor string

	// BackgroundColorFunc is the background color function of the svg
	BackgroundColorFunc ColorFunc

	// FontFamily is the font family used in the svg
	// The font family must be a valid font family from google fonts
	FontFamily string

	// FontWeight is the font weight
	FontWeight int

	// FontUrl is the font url where the font file is located
	// It is used to download the font file from the google font api
	// It is constructed from FontFamily and FontWeight
	fontUrl string

	// FontFile is the font file of the svg
	// it is constructed from FontFamily and FontWeight
	fontFile string

	// FontSize is the font size of the svg
	FontSize int

	// FontHeight is the font height of the svg
	fontHeight int

	// FontColor is the font color of the svg
	FontColor string

	// FontColorFunc is the font color function of the svg
	FontColorFunc ColorFunc

	// FontDir is the directory where the font files are stored
	FontDir string

	// Shape is the shape of the svg
	Shape string

	// BorderDash is the border style of the svg
	// The border style must be a valid svg border style
	//
	// Examples: "none", "5,5", "10,10"
	//
	// Reference: https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/stroke-dasharray
	BorderDash string

	// BorderWidth is the border width of the svg
	BorderWidth int

	// BorderRadius is the border radius of the svg
	BorderRadius int

	// BorderColor is the border color of the svg
	BorderColor string

	// BorderColorFunc is the border color function of the svg
	BorderColorFunc ColorFunc

	// Padding is the padding of the svg
	Padding int

	// RandomColor is a boolean to indicate if the svg should be random color
	RandomColor bool

	// RandomBackgroundColor is a boolean to indicate if the svg should be random background color
	RandomBackgroundColor bool

	// RandomFontColor is a boolean to indicate if the svg should be random font color
	RandomFontColor bool

	// RandomBorderColor is a boolean to indicate if the svg should be random border color
	RandomBorderColor bool
}

// GeneratorOption is a function to set the avatarGenerator options
type GeneratorOption func(*avatarGenerator)

// GetSVG returns the svg of a name
func GetSVG(initials string, options ...GeneratorOption) (*bytes.Buffer, error) {

	// Set the default avatar generator options
	opts := avatarGenerator{
		Initials:              initials,
		Width:                 48,
		Height:                48,
		BackgroundColor:       "",
		FontFamily:            "",
		FontWeight:            400,
		FontSize:              0,
		FontColor:             "",
		Shape:                 "circle",
		BorderDash:            "none",
		BorderWidth:           0,
		BorderRadius:          0,
		BorderColor:           "",
		Padding:               0,
		RandomColor:           false,
		RandomBackgroundColor: true,
		RandomFontColor:       true,
		RandomBorderColor:     false,
	}

	// Apply the options
	for _, option := range options {
		option(&opts)
	}

	// Generate the svg
	return generateSVG(opts)
}

// WithWidth sets the width of the svg
func WithWidth(width int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.Width = width
		if opts.Width == 0 {
			opts.Width = 48
		}
	}
}

// WithHeight sets the height of the svg
func WithHeight(height int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.Height = height
		if opts.Height == 0 {
			opts.Height = 48
		}
	}
}

// WithBackgroundColor sets the background color of the svg
func WithBackgroundColor(backgroundColor string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.BackgroundColor = backgroundColor
	}
}

// WithBackgroundColorFunc sets the background color function
func WithBackgroundColorFunc(backgroundColorFunc ColorFunc) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.BackgroundColorFunc = backgroundColorFunc
	}
}

// WithFont sets the font of the svg
func WithFont(family string, weight int) GeneratorOption {
	return func(opts *avatarGenerator) {
		familyEscaped := strings.Replace(family, ` `, `+`, -1)
		opts.FontFamily = family
		opts.FontWeight = weight
		opts.fontUrl = fmt.Sprintf("https://fonts.googleapis.com/css?family=%s:%d", familyEscaped, weight)
		opts.fontFile = fmt.Sprintf("%s-%d", family, weight)
	}
}

// WithFontSize sets the font size of the text
func WithFontSize(fontSize int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.FontSize = fontSize
	}
}

// WithFontColor sets the font color of the text
func WithFontColor(fontColor string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.FontColor = fontColor
	}
}

// WithFontColorFunc sets the font color function
func WithFontColorFunc(fontColorFunc ColorFunc) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.FontColorFunc = fontColorFunc
	}
}

// WithFontDir sets the directory for storing font files
func WithFontDir(fontDir string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.FontDir = fontDir
	}
}

// WithShape sets the shape of the svg
func WithShape(shape string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.Shape = shape

		if opts.Shape == "" {
			opts.Shape = "circle"
		} else if opts.Shape == "circle" {
			opts.Shape = "circle"
		} else if opts.Shape == "square" {
			opts.Shape = "rect"
		} else if opts.Shape == "rect" {
			opts.Shape = "rect"
		} else {
			opts.Shape = "circle"
		}
	}
}

// WithBorderDash sets the border style
func WithBorderDash(borderDash string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.BorderDash = borderDash
	}
}

// WithBorderWidth sets the border width
func WithBorderWidth(borderWidth int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.BorderWidth = borderWidth
	}
}

// WithBorderRadius sets the border radius
func WithBorderRadius(borderRadius int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.BorderRadius = borderRadius
	}
}

// WithBorderColor sets the border color
func WithBorderColor(borderColor string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.BorderColor = borderColor
	}
}

// WithBorderColorFunc sets the border color function of the svg
func WithBorderColorFunc(borderColorFunc ColorFunc) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.BorderColorFunc = borderColorFunc
	}
}

// WithPadding sets the padding
func WithPadding(padding int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.Padding = padding
	}
}

// WithRandomColor sets the svg to use random colors
func WithRandomColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.RandomColor = true
		if opts.RandomColor {
			opts.RandomBackgroundColor = true
			opts.RandomFontColor = true
			opts.RandomBorderColor = true
		} else {
			opts.RandomBackgroundColor = false
			opts.RandomFontColor = false
			opts.RandomBorderColor = false
		}
	}
}

// WithRandomBackgroundColor sets the svg to be random background color
func WithRandomBackgroundColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.RandomBackgroundColor = true
	}
}

// WithRandomFontColor sets the svg to be random font color
func WithRandomFontColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.RandomFontColor = true
	}
}

// WithRandomBorderColor sets the svg to be random border color
func WithRandomBorderColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.RandomBorderColor = true
	}
}

// generateSVG generates the svg
func generateSVG(opts avatarGenerator) (*bytes.Buffer, error) {

	// buffer to write the svg to
	var buffer = new(bytes.Buffer)

	// Set default font directory to temporary directory if not specified
	if opts.FontDir == "" {
		opts.FontDir = os.TempDir()
	}

	// Color randomizer
	cr := NewColorGenerator(opts.Initials)
	if opts.BackgroundColorFunc != nil {
		cr.BackgroundColorFunc = opts.BackgroundColorFunc
	}
	if opts.FontColorFunc != nil {
		cr.TextColorFunc = opts.FontColorFunc
	}
	if opts.BorderColorFunc != nil {
		cr.BorderColorFunc = opts.BorderColorFunc
	}

	// create the canvas
	canvas := svg.New(buffer)
	canvas.Start(opts.Width, opts.Height)

	// set the background color
	if opts.BackgroundColor != "" {
		color, err := colorFromString(opts.BackgroundColor)
		if err != nil {
			opts.BackgroundColor = cr.GetBackgroundColor().Hex()
		} else {
			opts.BackgroundColor = color.Hex()
		}
	} else {
		opts.BackgroundColor = cr.GetBackgroundColor().Hex()
	}

	// set the font color
	if opts.FontColor != "" {
		color, err := colorFromString(opts.FontColor)
		if err != nil {
			opts.FontColor = cr.GetTextColor().Hex()
		} else {
			opts.FontColor = color.Hex()
		}
	} else {
		opts.FontColor = cr.GetTextColor().Hex()
	}

	// set the border color
	if opts.BorderColor != "" {
		color, err := colorFromString(opts.BorderColor)
		if err != nil {
			opts.BorderColor = cr.GetBorderColor().Hex()
		} else {
			opts.BorderColor = color.Hex()
		}
	} else {
		opts.BorderColor = cr.GetBorderColor().Hex()
	}

	// set the border style
	if opts.BorderDash == "" {
		opts.BorderDash = "none"
	}

	// if fontUrl and fontFile are both set, download the font and save it to the assets/fonts folder
	if opts.fontUrl != "" && opts.fontFile != "" {
		fontPath := filepath.Join(opts.FontDir, opts.fontFile)
		// if the font file does not exist, download it
		// else use the existing font file
		if _, err := os.Stat(fontPath); os.IsNotExist(err) {
			err := downloadFile(fontPath, opts.fontUrl)
			if err != nil {
				return nil, err
			}
		}
	}

	// set the font size
	if opts.FontSize == 0 {
		opts.fontHeight = int(math.Min(float64(opts.Width), float64(opts.Height)) * 3 / 5)
		opts.FontSize = opts.fontHeight * 12 / 16
	} else {
		opts.fontHeight = opts.FontSize * 16 / 12
	}

	// Start the defs
	canvas.Def()
	// add the font file to the svg as a style
	if opts.FontFamily != "" && opts.fontFile != "" {
		// get the content of the font file
		fontFile, err := os.ReadFile(filepath.Join(opts.FontDir, opts.fontFile))
		if err != nil {
			return nil, err
		}
		// print the font file to the svg
		_, _ = fmt.Fprintln(canvas.Writer, "<style>")
		_, _ = fmt.Fprintln(canvas.Writer, string(fontFile))
		_, _ = fmt.Fprintln(canvas.Writer, "</style>")
	}
	// add the text style
	canvas.Style("text/css",
		fmt.Sprintf("text { text-anchor:middle;dominant-baseline:auto;font-family:%s,sans-serif;font-size:%dpx;fill:%s }",
			opts.FontFamily,
			opts.FontSize,
			opts.FontColor),
	)
	// end the defs
	canvas.DefEnd()

	// set the shape
	if opts.Shape == "rect" {
		rect := `<rect `
		rect += fmt.Sprintf(`x="%d" `, opts.Padding+opts.BorderWidth)
		rect += fmt.Sprintf(`y="%d" `, opts.Padding+opts.BorderWidth)
		rect += fmt.Sprintf(`width="%d" `, opts.Width-opts.Padding*2-opts.BorderWidth*2)
		rect += fmt.Sprintf(`height="%d" `, opts.Height-opts.Padding*2-opts.BorderWidth*2)
		rect += fmt.Sprintf(`rx="%d" `, opts.BorderRadius)
		rect += fmt.Sprintf(`ry="%d" `, opts.BorderRadius)
		rect += fmt.Sprintf(`fill="%s" `, opts.BackgroundColor)
		rect += fmt.Sprintf(`stroke="%s" `, opts.BorderColor)
		rect += fmt.Sprintf(`stroke-width="%d" `, opts.BorderWidth)
		rect += fmt.Sprintf(`stroke-dasharray="%s" `, opts.BorderDash)
		rect += `/>`

		_, _ = fmt.Fprintln(canvas.Writer, rect)
	} else {
		circle := `<circle `
		circle += fmt.Sprintf(`cx="%d" `, opts.Width/2)
		circle += fmt.Sprintf(`cy="%d" `, opts.Height/2)
		circle += fmt.Sprintf(`r="%d" `, opts.Width/2-opts.Padding-opts.BorderWidth)
		circle += fmt.Sprintf(`fill="%s" `, opts.BackgroundColor)
		circle += fmt.Sprintf(`stroke="%s" `, opts.BorderColor)
		circle += fmt.Sprintf(`stroke-width="%d" `, opts.BorderWidth)
		circle += fmt.Sprintf(`stroke-dasharray="%s" `, opts.BorderDash)
		circle += `/>`

		_, _ = fmt.Fprintln(canvas.Writer, circle)
	}

	// set the text
	if opts.Initials != "" {
		canvas.Text(
			opts.Width/2,
			opts.Height/2+opts.fontHeight/4,
			opts.Initials,
		)
	}

	// end the canvas
	canvas.End()

	return buffer, nil
}

// downloadFile will download an url to a local file. It's efficient because it will
// write as it downloads and not load the whole file into memory.
func downloadFile(filepath string, url string) error {
	// Get the data
	resp, err := http.Get(url) // #nosec:G107
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Create the file
	out, err := os.Create(filepath) // #nosec:G304
	if err != nil {
		return err
	}
	defer func(out *os.File) {
		_ = out.Close()
	}(out)

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

// colorFromString will return hex color from a string
func colorFromString(strColor string) (colorful.Color, error) {
	// Get the color from name
	if c, ok := colornames.Map[strings.ToLower(strColor)]; ok {
		if cf, ok := colorful.MakeColor(c); ok {
			return cf, nil
		}
		return colorful.Color{}, errors.New("unable to make color from color name")
	}
	// Get the color from hex
	if cf, err := colorful.Hex(strColor); err == nil {
		return cf, nil
	} else if cf, err := colorful.Hex("#" + strColor); err == nil {
		return cf, nil
	} else {
		return colorful.Color{}, errors.New("unable to make color from hex")
	}
}
