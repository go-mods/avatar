package avatar

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/lucasb-eyer/go-colorful"
	"golang.org/x/image/colornames"
)

var (
	// ErrInvalidFontFamily is returned when the font family holds other characters than ASCII letters,
	// digits and inner spaces, the characters of the Google Fonts family names, or when it is too long
	ErrInvalidFontFamily = errors.New("invalid font family")

	// ErrInvalidFontWeight is returned when the font weight is not between 1 and 1000
	ErrInvalidFontWeight = errors.New("invalid font weight")

	// ErrInvalidBorderDash is returned when the border dash is neither "none" nor a list of numbers
	ErrInvalidBorderDash = errors.New("invalid border dash")

	// ErrInvalidSize is returned when a dimension is negative, or when the padding and the border
	// leave no room for the shape
	ErrInvalidSize = errors.New("invalid size")

	// ErrInvalidColor is returned when a color is neither a color name nor a hex color code
	ErrInvalidColor = errors.New("invalid color")
)

var (
	// fontFamilyRegex matches a Google Fonts family name: "Inter", "Roboto Mono", "M PLUS 1p".
	// The family is used in a file name, an url and the svg style: nothing else is accepted.
	fontFamilyRegex = regexp.MustCompile(`^[A-Za-z0-9 ]+$`)

	// borderDashRegex matches a stroke-dasharray value: "none", "5", "5,5", "6 2", "1.5 3"
	borderDashRegex = regexp.MustCompile(`^(none|\d+(\.\d+)?([ ,]+\d+(\.\d+)?)*)$`)
)

const (
	// svgHeader starts the document: width and height
	svgHeader = `<?xml version="1.0"?>` + "\n" +
		`<svg width="%d" height="%d"` + "\n" +
		`     xmlns="http://www.w3.org/2000/svg"` + "\n" +
		`     xmlns:xlink="http://www.w3.org/1999/xlink">` + "\n"

	// svgTextStart starts the text: x, y, font family, font size and color. These are attributes
	// and not a style: the style of an svg inlined in a page applies to the whole page.
	svgTextStart = `<text x="%d" y="%d" text-anchor="middle" font-family="%s" font-size="%d" fill="%s">`
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

// GetSVG returns the svg of a name.
// It returns ErrInvalidFontFamily, ErrInvalidFontWeight, ErrInvalidBorderDash, ErrInvalidSize or
// ErrInvalidColor when an option is not valid.
func GetSVG(initials string, options ...GeneratorOption) (*bytes.Buffer, error) {
	return GetSVGContext(context.Background(), initials, options...)
}

// GetSVGContext is GetSVG with a context, which stops the download of the font.
func GetSVGContext(ctx context.Context, initials string, options ...GeneratorOption) (*bytes.Buffer, error) {

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
	return generateSVG(ctx, opts)
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

// WithFont sets the font of the svg, a Google Fonts family and a weight between 1 and 1000.
// The font is embedded in the svg, reduced to the letters of the initials: it is displayed even when
// the svg is an image. It is downloaded once for each family, weight and initials into the font directory.
func WithFont(family string, weight int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.FontFamily = family
		opts.FontWeight = weight
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

		switch opts.Shape {
		case "":
			opts.Shape = "circle"
		case "circle":
			opts.Shape = "circle"
		case "square":
			opts.Shape = "rect"
		case "rect":
			opts.Shape = "rect"
		default:
			opts.Shape = "circle"
		}
	}
}

// WithBorderDash sets the border style: "none" or a list of numbers ("5,5", "6 2")
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
func generateSVG(ctx context.Context, opts avatarGenerator) (*bytes.Buffer, error) {
	// The options are written to files, urls and the svg: reject unexpected values
	if err := validate(opts); err != nil {
		return nil, err
	}

	// buffer to write the svg to
	var buffer = new(bytes.Buffer)

	// Set default font directory to temporary directory if not specified
	if opts.FontDir == "" {
		opts.FontDir = os.TempDir()
	}

	// Color randomizer
	var colorOptions []ColorOption
	if opts.BackgroundColorFunc != nil {
		colorOptions = append(colorOptions, BackgroundColorFunc(opts.BackgroundColorFunc))
	}
	if opts.FontColorFunc != nil {
		colorOptions = append(colorOptions, TextColorFunc(opts.FontColorFunc))
	}
	if opts.BorderColorFunc != nil {
		colorOptions = append(colorOptions, BorderColorFunc(opts.BorderColorFunc))
	}
	cr := NewColorGenerator(opts.Initials, colorOptions...)

	// set the colors: a color which is not given comes from the color generator
	var err error
	if opts.BackgroundColor, err = resolveColor(opts.BackgroundColor, cr.GetBackgroundColor()); err != nil {
		return nil, err
	}
	if opts.FontColor, err = resolveColor(opts.FontColor, cr.GetTextColor()); err != nil {
		return nil, err
	}
	if opts.BorderColor, err = resolveColor(opts.BorderColor, cr.GetBorderColor()); err != nil {
		return nil, err
	}

	// start the document
	_, _ = fmt.Fprintf(buffer, svgHeader, opts.Width, opts.Height)

	// set the border style
	if opts.BorderDash == "" {
		opts.BorderDash = "none"
	}

	// set the font size
	if opts.FontSize == 0 {
		opts.fontHeight = min(opts.Width, opts.Height) * 3 / 5
		opts.FontSize = opts.fontHeight * 12 / 16
	} else {
		opts.fontHeight = opts.FontSize * 16 / 12
	}

	// embed the font in the svg as a style: without text, there is nothing to draw with it
	if opts.FontFamily != "" && opts.Initials != "" {
		style, err := fontStyle(ctx, opts)
		if err != nil {
			return nil, err
		}
		if style != "" {
			_, _ = fmt.Fprintln(buffer, "<defs>")
			_, _ = fmt.Fprintln(buffer, "<style>")
			_, _ = fmt.Fprintln(buffer, style)
			_, _ = fmt.Fprintln(buffer, "</style>")
			_, _ = fmt.Fprintln(buffer, "</defs>")
		}
	}

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

		_, _ = fmt.Fprintln(buffer, rect)
	} else {
		circle := `<circle `
		circle += fmt.Sprintf(`cx="%d" `, opts.Width/2)
		circle += fmt.Sprintf(`cy="%d" `, opts.Height/2)
		circle += fmt.Sprintf(`r="%d" `, min(opts.Width, opts.Height)/2-opts.Padding-opts.BorderWidth)
		circle += fmt.Sprintf(`fill="%s" `, opts.BackgroundColor)
		circle += fmt.Sprintf(`stroke="%s" `, opts.BorderColor)
		circle += fmt.Sprintf(`stroke-width="%d" `, opts.BorderWidth)
		circle += fmt.Sprintf(`stroke-dasharray="%s" `, opts.BorderDash)
		circle += `/>`

		_, _ = fmt.Fprintln(buffer, circle)
	}

	// set the text
	if opts.Initials != "" {
		_, _ = fmt.Fprintf(buffer, svgTextStart, opts.Width/2, opts.Height/2+opts.fontHeight/4,
			fontFamilyList(opts.FontFamily), opts.FontSize, opts.FontColor)
		// The initials come from the caller: they are escaped to stay plain text
		_ = xml.EscapeText(buffer, []byte(opts.Initials))
		buffer.WriteString("</text>\n")
	}

	// end the document
	buffer.WriteString("</svg>\n")

	return buffer, nil
}

// validate checks the options which are not converted to a safe value before being used
func validate(opts avatarGenerator) error {
	if opts.FontFamily != "" {
		if len(opts.FontFamily) > maxFontFamilyLength || strings.TrimSpace(opts.FontFamily) != opts.FontFamily ||
			!fontFamilyRegex.MatchString(opts.FontFamily) {
			return fmt.Errorf("%w: %q", ErrInvalidFontFamily, opts.FontFamily)
		}
		if opts.FontWeight < 1 || opts.FontWeight > 1000 {
			return fmt.Errorf("%w: %d", ErrInvalidFontWeight, opts.FontWeight)
		}
	}
	if !borderDashRegex.MatchString(opts.BorderDash) {
		return fmt.Errorf("%w: %q", ErrInvalidBorderDash, opts.BorderDash)
	}
	return validateSize(opts)
}

// validateSize checks that the dimensions describe a shape which can be drawn
func validateSize(opts avatarGenerator) error {
	if opts.Width <= 0 || opts.Height <= 0 {
		return fmt.Errorf("%w: %dx%d", ErrInvalidSize, opts.Width, opts.Height)
	}
	for _, dimension := range []struct {
		name  string
		value int
	}{
		{"font size", opts.FontSize},
		{"border width", opts.BorderWidth},
		{"border radius", opts.BorderRadius},
		{"padding", opts.Padding},
	} {
		if dimension.value < 0 {
			return fmt.Errorf("%w: %s %d", ErrInvalidSize, dimension.name, dimension.value)
		}
	}
	// Compared one by one: the sum of two large values could overflow
	room := min(opts.Width, opts.Height) / 2
	if opts.Padding >= room || opts.BorderWidth >= room-opts.Padding {
		return fmt.Errorf("%w: padding %d and border width %d leave no room in %dx%d",
			ErrInvalidSize, opts.Padding, opts.BorderWidth, opts.Width, opts.Height)
	}
	return nil
}

// fontFamilyList returns the value of the font-family property. The family is quoted because a
// name like "M PLUS 1p" is not a valid list of CSS identifiers: fontFamilyRegex excludes the quote.
func fontFamilyList(family string) string {
	if family == "" {
		return "sans-serif"
	}
	return "'" + family + "',sans-serif"
}

// resolveColor returns the hex code of a color, or of the generated color when no color is given
func resolveColor(value string, generated colorful.Color) (string, error) {
	if value == "" {
		return generated.Hex(), nil
	}
	color, err := colorFromString(value)
	if err != nil {
		return "", fmt.Errorf("%w: %q", ErrInvalidColor, value)
	}
	return color.Hex(), nil
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
