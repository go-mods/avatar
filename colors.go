package avatar

import (
	"github.com/lucasb-eyer/go-colorful"
)

// colorGenerator is a struct to hold colors randomizer from a string
type colorGenerator struct {
	// Name is the name to get the color from
	// it is used has a seed to generate the colors
	Name string

	// Hash is the hash of the name
	Hash int

	// BackgroundColor is the background color
	BackgroundColor colorful.Color
	// TextColor is the text color
	TextColor colorful.Color
	// BorderColor is the border color
	BorderColor colorful.Color

	// BackgroundColorFunc is the background color function
	// if defined, it will be used instead of BackgroundColor
	BackgroundColorFunc ColorFunc
	// TextColorFunc is the text color function
	// if defined, it will be used instead of TextColor
	TextColorFunc ColorFunc
	// BorderColorFunc is the border color function
	// if defined, it will be used instead of BorderColor
	BorderColorFunc ColorFunc
}

// ColorOption is a function to set the colorGenerator options
type ColorOption func(*colorGenerator)

// ColorFunc is function used to generate a color
type ColorFunc func(r float64) colorful.Color

// NewColorGenerator returns a new colorGenerator
func NewColorGenerator(Name string, options ...ColorOption) *colorGenerator {

	// set the default generator options
	c := &colorGenerator{
		Name:                Name,
		BackgroundColorFunc: defaultBackgroundColorGenerator,
		TextColorFunc:       defaultTextColorGenerator,
		BorderColorFunc:     defaultBorderColorGenerator,
	}

	// apply generator options
	for _, option := range options {
		option(c)
	}

	// calculate the hash
	c.Hash = c.hash(c.Name)

	// normalize the hash to get a number between 0 and 1
	nh := c.normalizeHash(c.Hash, 0, 1)

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
	return func(c *colorGenerator) {
		c.BackgroundColorFunc = f
	}
}

// TextColorFunc sets the text color generator function
func TextColorFunc(f ColorFunc) ColorOption {
	return func(c *colorGenerator) {
		c.TextColorFunc = f
	}
}

// BorderColorFunc sets the border color generator function
func BorderColorFunc(f ColorFunc) ColorOption {
	return func(c *colorGenerator) {
		c.BorderColorFunc = f
	}
}

// GetBackgroundColor returns a color
func (c *colorGenerator) GetBackgroundColor() colorful.Color {
	return c.BackgroundColor
}

// GetTextColor returns a color
func (c *colorGenerator) GetTextColor() colorful.Color {
	return c.TextColor
}

// GetBorderColor returns a color
func (c *colorGenerator) GetBorderColor() colorful.Color {
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

// hash returns a hash of a string
func (c *colorGenerator) hash(s string) int {
	h := 0
	for _, r := range s {
		h = 31*h + int(r)
	}
	return h
}

// normalizeHash returns a normalized hash between two values
func (c *colorGenerator) normalizeHash(hash, min, max int) float64 {
	return float64(min) + float64(hash%1000)/1000*float64(max-min)
}
