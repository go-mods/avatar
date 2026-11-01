package avatar

import (
	"github.com/lucasb-eyer/go-colorful"
)

// ColorGenerator holds the colors generated from a name: the same name always gives the same colors
type ColorGenerator struct {
	// Name is the name to get the colors from
	// it is used as a seed to generate the colors
	Name string

	// Hash is the hash of the name
	Hash int64

	// BackgroundColor is the background color
	BackgroundColor colorful.Color
	// TextColor is the text color
	TextColor colorful.Color
	// BorderColor is the border color
	BorderColor colorful.Color

	// BackgroundColorFunc is the function which generates the background color
	BackgroundColorFunc ColorFunc
	// TextColorFunc is the function which generates the text color
	TextColorFunc ColorFunc
	// BorderColorFunc is the function which generates the border color
	BorderColorFunc ColorFunc
}

// ColorOption is a function to set the ColorGenerator options
type ColorOption func(*ColorGenerator)

// ColorFunc is a function used to generate a color from a value in the range [0, 1)
type ColorFunc func(r float64) colorful.Color

// NewColorGenerator returns the colors generated from a name
func NewColorGenerator(name string, options ...ColorOption) *ColorGenerator {

	// set the default generator options
	c := &ColorGenerator{
		Name:                name,
		BackgroundColorFunc: defaultBackgroundColorGenerator,
		TextColorFunc:       defaultTextColorGenerator,
		BorderColorFunc:     defaultBorderColorGenerator,
	}

	// apply generator options
	for _, option := range options {
		option(c)
	}

	// calculate the hash
	c.Hash = hash(c.Name)

	// normalize the hash to get a number between 0 and 1
	nh := normalizeHash(c.Hash)

	// get the background color
	c.BackgroundColor = c.BackgroundColorFunc(nh)

	// get the text color
	c.TextColor = c.TextColorFunc(nh)

	// get the border color
	c.BorderColor = c.BorderColorFunc(nh)

	return c
}

// BackgroundColorFunc sets the background color generator function
func BackgroundColorFunc(f ColorFunc) ColorOption {
	return func(c *ColorGenerator) {
		c.BackgroundColorFunc = f
	}
}

// TextColorFunc sets the text color generator function
func TextColorFunc(f ColorFunc) ColorOption {
	return func(c *ColorGenerator) {
		c.TextColorFunc = f
	}
}

// BorderColorFunc sets the border color generator function
func BorderColorFunc(f ColorFunc) ColorOption {
	return func(c *ColorGenerator) {
		c.BorderColorFunc = f
	}
}

// GetBackgroundColor returns the background color
func (c *ColorGenerator) GetBackgroundColor() colorful.Color {
	return c.BackgroundColor
}

// GetTextColor returns the text color
func (c *ColorGenerator) GetTextColor() colorful.Color {
	return c.TextColor
}

// GetBorderColor returns the border color
func (c *ColorGenerator) GetBorderColor() colorful.Color {
	return c.BorderColor
}

// defaultBackgroundColorGenerator returns a default background color generator
func defaultBackgroundColorGenerator(r float64) colorful.Color {
	return colorful.Hsv(r*360.0, 0.15+r*0.30, 0.75+r*0.25)
}

// defaultTextColorGenerator returns a default text color generator
func defaultTextColorGenerator(r float64) colorful.Color {
	return colorful.Hsv(r*360.0, 0.95+r*.05, 0.55+r*.05)
}

// defaultBorderColorGenerator returns a default border color generator
func defaultBorderColorGenerator(r float64) colorful.Color {
	return colorful.Hsv(r*360.0, r*.2, 0.4+r*0.5)
}

// hash returns a hash of a string. It uses an int64 to give the same result on every architecture,
// and it overflows on long strings: the result can be negative.
func hash(s string) int64 {
	var h int64
	for _, r := range s {
		h = 31*h + int64(r)
	}
	return h
}

// normalizeHash returns a hash as a number in the range [0, 1)
func normalizeHash(hash int64) float64 {
	return float64((hash%1000+1000)%1000) / 1000
}
