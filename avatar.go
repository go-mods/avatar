// Package avatar generates svg avatars, a shape filled with a color and the initials of a name.
//
// The colors come from the initials, the same initials always giving the same colors. The options
// change the size, the shape, the colors, the border and the font:
//
//	svg, err := avatar.GetSVG("JD",
//		avatar.WithShape("square"),
//		avatar.WithWidth(64),
//		avatar.WithHeight(64),
//		avatar.WithFont("Inter", 900),
//	)
//
// A font is a Google Fonts family. It is embedded in the svg, reduced to the letters of the initials,
// so that it is displayed even when the svg is an image. The text is drawn with a fallback font when
// the font holds none of these letters or when the font servers cannot be reached.
package avatar

import (
	"bytes"
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"math/rand/v2"
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

	// ErrUnknownFont is returned when Google Fonts holds neither this family nor this weight
	ErrUnknownFont = errors.New("unknown font")

	// ErrInvalidBorderDash is returned when the border dash is neither "none" nor a list of numbers
	ErrInvalidBorderDash = errors.New("invalid border dash")

	// ErrInvalidSize is returned when a dimension is negative, or when the padding and the border
	// leave no room for the shape
	ErrInvalidSize = errors.New("invalid size")

	// ErrInvalidColor is returned when a color is neither a color name nor a hex color code
	ErrInvalidColor = errors.New("invalid color")

	// ErrInvalidShape is returned when the shape is neither a circle nor a square
	ErrInvalidShape = errors.New("invalid shape")
)

var (
	// fontFamilyRegex matches a Google Fonts family name: "Inter", "Roboto Mono", "M PLUS 1p".
	// The family is used in a file name, an url and the svg: nothing else is accepted.
	fontFamilyRegex = regexp.MustCompile(`^[A-Za-z0-9 ]+$`)

	// borderDashRegex matches a stroke-dasharray value: "none", "5", "5,5", "6 2", "1.5 3"
	borderDashRegex = regexp.MustCompile(`^(none|\d+(\.\d+)?([ ,]+\d+(\.\d+)?)*)$`)

	// randomSeed draws the value given to the color functions by the random options.
	randomSeed = rand.Float64 // #nosec G404 -- a color is not a secret
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

	// The shapes of the svg
	shapeCircle = "circle"
	shapeRect   = "rect"
)

// avatarGenerator is a struct to hold the options for the avatar generator
type avatarGenerator struct {
	// initials is the text of the svg
	initials string

	// width is the width of the svg
	width int

	// height is the height of the svg
	height int

	// backgroundColor is the background color of the svg, a color name or a hex color code
	backgroundColor string

	// backgroundColorFunc generates the background color
	backgroundColorFunc ColorFunc

	// fontFamily is the font family used in the svg, a family of Google Fonts
	fontFamily string

	// fontWeight is the font weight
	fontWeight int

	// fontSize is the font size of the svg
	fontSize int

	// fontHeight is the font height of the svg
	fontHeight int

	// fontColor is the font color of the svg, a color name or a hex color code
	fontColor string

	// fontColorFunc generates the font color
	fontColorFunc ColorFunc

	// fontDir is the directory where the fonts are kept
	fontDir string

	// shape is the shape of the svg
	shape string

	// borderDash is the border style of the svg
	//
	// Examples: "none", "5,5", "10,10"
	//
	// Reference: https://developer.mozilla.org/en-US/docs/Web/SVG/Attribute/stroke-dasharray
	borderDash string

	// borderWidth is the border width of the svg
	borderWidth int

	// borderRadius is the border radius of a square
	borderRadius int

	// borderColor is the border color of the svg, a color name or a hex color code
	borderColor string

	// borderColorFunc generates the border color
	borderColorFunc ColorFunc

	// padding is the space around the shape
	padding int

	// randomColor draws one value for all the generated colors
	randomColor bool

	// randomBackgroundColor draws a value for the background color
	randomBackgroundColor bool

	// randomFontColor draws a value for the font color
	randomFontColor bool

	// randomBorderColor draws a value for the border color
	randomBorderColor bool
}

// GeneratorOption is a function to set an option of the svg
type GeneratorOption func(*avatarGenerator)

// GetSVG returns the svg of initials.
// It returns ErrInvalidFontFamily, ErrInvalidFontWeight, ErrInvalidBorderDash, ErrInvalidSize,
// ErrInvalidColor or ErrInvalidShape when an option is not valid, and ErrUnknownFont when the font
// does not exist.
func GetSVG(initials string, options ...GeneratorOption) (*bytes.Buffer, error) {
	return GetSVGContext(context.Background(), initials, options...)
}

// GetSVGContext is GetSVG with a context, which stops the download of the font.
func GetSVGContext(ctx context.Context, initials string, options ...GeneratorOption) (*bytes.Buffer, error) {

	// Set the default avatar generator options
	opts := avatarGenerator{
		initials:   initials,
		width:      48,
		height:     48,
		fontWeight: 400,
		shape:      shapeCircle,
		borderDash: "none",
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
		opts.width = width
		if opts.width == 0 {
			opts.width = 48
		}
	}
}

// WithHeight sets the height of the svg
func WithHeight(height int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.height = height
		if opts.height == 0 {
			opts.height = 48
		}
	}
}

// WithBackgroundColor sets the background color of the svg, a color name or a hex color code
func WithBackgroundColor(backgroundColor string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.backgroundColor = backgroundColor
	}
}

// WithBackgroundColorFunc sets the function which generates the background color
func WithBackgroundColorFunc(backgroundColorFunc ColorFunc) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.backgroundColorFunc = backgroundColorFunc
	}
}

// WithFont sets the font of the svg, a Google Fonts family and a weight between 1 and 1000.
// The font is embedded in the svg, reduced to the letters of the initials: it is displayed even when
// the svg is an image. It is downloaded once for each family, weight and initials into the font directory.
func WithFont(family string, weight int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.fontFamily = family
		opts.fontWeight = weight
	}
}

// WithFontSize sets the font size of the text
func WithFontSize(fontSize int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.fontSize = fontSize
	}
}

// WithFontColor sets the font color of the text, a color name or a hex color code
func WithFontColor(fontColor string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.fontColor = fontColor
	}
}

// WithFontColorFunc sets the function which generates the font color
func WithFontColorFunc(fontColorFunc ColorFunc) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.fontColorFunc = fontColorFunc
	}
}

// WithFontDir sets the directory where the fonts are kept, which is created if needed.
// By default, the fonts are kept in the temporary directory.
func WithFontDir(fontDir string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.fontDir = fontDir
	}
}

// WithShape sets the shape of the svg: "circle", the default, or "square"
func WithShape(shape string) GeneratorOption {
	return func(opts *avatarGenerator) {
		switch shape {
		case "", shapeCircle:
			opts.shape = shapeCircle
		case "square", shapeRect:
			opts.shape = shapeRect
		default:
			opts.shape = shape
		}
	}
}

// WithBorderDash sets the border style: "none" or a list of numbers ("5,5", "6 2")
func WithBorderDash(borderDash string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.borderDash = borderDash
	}
}

// WithBorderWidth sets the border width
func WithBorderWidth(borderWidth int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.borderWidth = borderWidth
	}
}

// WithBorderRadius sets the border radius of a square
func WithBorderRadius(borderRadius int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.borderRadius = borderRadius
	}
}

// WithBorderColor sets the border color, a color name or a hex color code
func WithBorderColor(borderColor string) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.borderColor = borderColor
	}
}

// WithBorderColorFunc sets the function which generates the border color
func WithBorderColorFunc(borderColorFunc ColorFunc) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.borderColorFunc = borderColorFunc
	}
}

// WithPadding sets the space around the shape
func WithPadding(padding int) GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.padding = padding
	}
}

// WithRandomColor draws the colors at random at each call, instead of generating them from the
// initials. The same value is given to the three color functions: the colors go together.
// A color set by WithBackgroundColor, WithFontColor or WithBorderColor stays as it is.
func WithRandomColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.randomColor = true
	}
}

// WithRandomBackgroundColor draws the background color at random at each call
func WithRandomBackgroundColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.randomBackgroundColor = true
	}
}

// WithRandomFontColor draws the font color at random at each call
func WithRandomFontColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.randomFontColor = true
	}
}

// WithRandomBorderColor draws the border color at random at each call
func WithRandomBorderColor() GeneratorOption {
	return func(opts *avatarGenerator) {
		opts.randomBorderColor = true
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
	if opts.fontDir == "" {
		opts.fontDir = os.TempDir()
	}

	// Color randomizer
	var colorOptions []ColorOption
	if opts.backgroundColorFunc != nil {
		colorOptions = append(colorOptions, BackgroundColorFunc(opts.backgroundColorFunc))
	}
	if opts.fontColorFunc != nil {
		colorOptions = append(colorOptions, TextColorFunc(opts.fontColorFunc))
	}
	if opts.borderColorFunc != nil {
		colorOptions = append(colorOptions, BorderColorFunc(opts.borderColorFunc))
	}
	cr := NewColorGenerator(opts.initials, colorOptions...)
	background, text, border := generatedColors(opts, cr)

	// set the colors: a color which is not given is a generated one
	var err error
	if opts.backgroundColor, err = resolveColor(opts.backgroundColor, background); err != nil {
		return nil, err
	}
	if opts.fontColor, err = resolveColor(opts.fontColor, text); err != nil {
		return nil, err
	}
	if opts.borderColor, err = resolveColor(opts.borderColor, border); err != nil {
		return nil, err
	}

	// start the document
	_, _ = fmt.Fprintf(buffer, svgHeader, opts.width, opts.height)

	// set the font size
	if opts.fontSize == 0 {
		opts.fontHeight = min(opts.width, opts.height) * 3 / 5
		opts.fontSize = opts.fontHeight * 12 / 16
	} else {
		opts.fontHeight = opts.fontSize * 16 / 12
	}

	// embed the font in the svg as a style: without text, there is nothing to draw with it
	if opts.fontFamily != "" && opts.initials != "" {
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
	if opts.shape == shapeRect {
		rect := `<rect `
		rect += fmt.Sprintf(`x="%d" `, opts.padding+opts.borderWidth)
		rect += fmt.Sprintf(`y="%d" `, opts.padding+opts.borderWidth)
		rect += fmt.Sprintf(`width="%d" `, opts.width-opts.padding*2-opts.borderWidth*2)
		rect += fmt.Sprintf(`height="%d" `, opts.height-opts.padding*2-opts.borderWidth*2)
		rect += fmt.Sprintf(`rx="%d" `, opts.borderRadius)
		rect += fmt.Sprintf(`ry="%d" `, opts.borderRadius)
		rect += fmt.Sprintf(`fill="%s" `, opts.backgroundColor)
		rect += fmt.Sprintf(`stroke="%s" `, opts.borderColor)
		rect += fmt.Sprintf(`stroke-width="%d" `, opts.borderWidth)
		rect += fmt.Sprintf(`stroke-dasharray="%s" `, opts.borderDash)
		rect += `/>`

		_, _ = fmt.Fprintln(buffer, rect)
	} else {
		circle := `<circle `
		circle += fmt.Sprintf(`cx="%d" `, opts.width/2)
		circle += fmt.Sprintf(`cy="%d" `, opts.height/2)
		circle += fmt.Sprintf(`r="%d" `, min(opts.width, opts.height)/2-opts.padding-opts.borderWidth)
		circle += fmt.Sprintf(`fill="%s" `, opts.backgroundColor)
		circle += fmt.Sprintf(`stroke="%s" `, opts.borderColor)
		circle += fmt.Sprintf(`stroke-width="%d" `, opts.borderWidth)
		circle += fmt.Sprintf(`stroke-dasharray="%s" `, opts.borderDash)
		circle += `/>`

		_, _ = fmt.Fprintln(buffer, circle)
	}

	// set the text
	if opts.initials != "" {
		_, _ = fmt.Fprintf(buffer, svgTextStart, opts.width/2, opts.height/2+opts.fontHeight/4,
			fontFamilyList(opts.fontFamily), opts.fontSize, opts.fontColor)
		// The initials come from the caller: they are escaped to stay plain text
		_ = xml.EscapeText(buffer, []byte(opts.initials))
		buffer.WriteString("</text>\n")
	}

	// end the document
	buffer.WriteString("</svg>\n")

	return buffer, nil
}

// validate checks the options which are not converted to a safe value before being used
func validate(opts avatarGenerator) error {
	if opts.fontFamily != "" {
		if len(opts.fontFamily) > maxFontFamilyLength || strings.TrimSpace(opts.fontFamily) != opts.fontFamily ||
			!fontFamilyRegex.MatchString(opts.fontFamily) {
			return fmt.Errorf("%w: %q", ErrInvalidFontFamily, opts.fontFamily)
		}
		if opts.fontWeight < 1 || opts.fontWeight > 1000 {
			return fmt.Errorf("%w: %d", ErrInvalidFontWeight, opts.fontWeight)
		}
	}
	if !borderDashRegex.MatchString(opts.borderDash) {
		return fmt.Errorf("%w: %q", ErrInvalidBorderDash, opts.borderDash)
	}
	if opts.shape != shapeCircle && opts.shape != shapeRect {
		return fmt.Errorf("%w: %q", ErrInvalidShape, opts.shape)
	}
	return validateSize(opts)
}

// validateSize checks that the dimensions describe a shape which can be drawn
func validateSize(opts avatarGenerator) error {
	if opts.width <= 0 || opts.height <= 0 {
		return fmt.Errorf("%w: %dx%d", ErrInvalidSize, opts.width, opts.height)
	}
	for _, dimension := range []struct {
		name  string
		value int
	}{
		{"font size", opts.fontSize},
		{"border width", opts.borderWidth},
		{"border radius", opts.borderRadius},
		{"padding", opts.padding},
	} {
		if dimension.value < 0 {
			return fmt.Errorf("%w: %s %d", ErrInvalidSize, dimension.name, dimension.value)
		}
	}
	// Compared one by one: the sum of two large values could overflow
	room := min(opts.width, opts.height) / 2
	if opts.padding >= room || opts.borderWidth >= room-opts.padding {
		return fmt.Errorf("%w: padding %d and border width %d leave no room in %dx%d",
			ErrInvalidSize, opts.padding, opts.borderWidth, opts.width, opts.height)
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

// generatedColors returns the colors of the initials, or those drawn at random as the options ask.
// The drawn values go through the color functions, which keep the text readable on the background.
func generatedColors(opts avatarGenerator, cr *ColorGenerator) (background, text, border colorful.Color) {
	background, text, border = cr.GetBackgroundColor(), cr.GetTextColor(), cr.GetBorderColor()

	if opts.randomColor {
		seed := randomSeed()
		return cr.BackgroundColorFunc(seed), cr.TextColorFunc(seed), cr.BorderColorFunc(seed)
	}
	if opts.randomBackgroundColor {
		background = cr.BackgroundColorFunc(randomSeed())
	}
	if opts.randomFontColor {
		text = cr.TextColorFunc(randomSeed())
	}
	if opts.randomBorderColor {
		border = cr.BorderColorFunc(randomSeed())
	}
	return background, text, border
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
