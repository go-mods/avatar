package avatar

import (
	"encoding/xml"
	"errors"
	"flag"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/lucasb-eyer/go-colorful"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// update rewrites the expected files of testdata: go test . -update
var update = flag.Bool("update", false, "rewrite the expected svg files of testdata")

// requireWellFormed fails the test when the document is not well-formed XML.
func requireWellFormed(t *testing.T, document string) {
	t.Helper()
	decoder := xml.NewDecoder(strings.NewReader(document))
	for {
		_, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			return
		}
		require.NoError(t, err, "document: %s", document)
	}
}

func TestGetSVG_output(t *testing.T) {
	// The font file is already in the font directory: nothing is downloaded
	fontDir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(fontDir, "Inter-400"), []byte("@font-face {}\n"), 0o600))

	tests := []struct {
		name     string
		initials string
		options  []GeneratorOption
	}{
		{name: "default", initials: "AB"},
		{name: "empty_initials"},
		{name: "escaped_initials", initials: `<a&"'>`},
		{
			name:     "square_border",
			initials: "CD",
			options: []GeneratorOption{
				WithShape("square"), WithWidth(64), WithHeight(64), WithBorderWidth(2), WithBorderRadius(8),
				WithBorderDash("6 2"), WithBorderColor("#AFA318"),
			},
		},
		{
			name:     "named_colors",
			initials: "EF",
			options:  []GeneratorOption{WithBackgroundColor("red"), WithFontColor("white"), WithBorderColor("blue"), WithBorderWidth(2)},
		},
		{
			name:     "padding",
			initials: "GH",
			options:  []GeneratorOption{WithWidth(100), WithHeight(100), WithPadding(10), WithBorderWidth(4)},
		},
		{
			name:     "font",
			initials: "IJ",
			options:  []GeneratorOption{WithFont("Inter", 400), WithFontDir(fontDir), WithFontSize(30)},
		},
		{
			name:     "wide_circle",
			initials: "KL",
			options:  []GeneratorOption{WithWidth(200), WithHeight(50), WithBorderWidth(2)},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GetSVG(tt.initials, tt.options...)
			require.NoError(t, err)
			requireWellFormed(t, svg.String())

			path := filepath.Join("testdata", tt.name+".svg")
			if *update {
				require.NoError(t, os.WriteFile(path, svg.Bytes(), 0o600))
			}
			want, err := os.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, string(want), svg.String())
		})
	}
}

func TestGetSVG_options(t *testing.T) {
	constant := func(hex string) ColorFunc {
		return func(float64) colorful.Color {
			color, err := colorful.Hex(hex)
			require.NoError(t, err)
			return color
		}
	}

	tests := []struct {
		name    string
		options []GeneratorOption
		want    string
	}{
		{name: "width", options: []GeneratorOption{WithWidth(100)}, want: `<svg width="100" height="48"`},
		{name: "default width", options: []GeneratorOption{WithWidth(0)}, want: `<svg width="48" height="48"`},
		{name: "height", options: []GeneratorOption{WithHeight(100)}, want: `<svg width="48" height="100"`},
		{name: "default height", options: []GeneratorOption{WithHeight(0)}, want: `<svg width="48" height="48"`},
		{name: "background color", options: []GeneratorOption{WithBackgroundColor("#123456")}, want: `fill="#123456"`},
		{name: "font size", options: []GeneratorOption{WithFontSize(30)}, want: `font-size="30"`},
		{name: "font color", options: []GeneratorOption{WithFontColor("#123456")}, want: `font-size="21" fill="#123456"`},
		{name: "border width", options: []GeneratorOption{WithBorderWidth(3)}, want: `stroke-width="3"`},
		{name: "border color", options: []GeneratorOption{WithBorderColor("#123456")}, want: `stroke="#123456"`},
		{name: "border radius", options: []GeneratorOption{WithShape("square"), WithBorderRadius(5)}, want: `rx="5" ry="5"`},
		{name: "padding on a circle", options: []GeneratorOption{WithPadding(4)}, want: `r="20"`},
		{name: "padding on a square", options: []GeneratorOption{WithShape("square"), WithPadding(4)}, want: `x="4" y="4" width="40" height="40"`},
		{name: "unknown shape", options: []GeneratorOption{WithShape("star")}, want: `<circle `},
		{name: "empty shape", options: []GeneratorOption{WithShape("")}, want: `<circle `},
		{name: "rect shape", options: []GeneratorOption{WithShape("rect")}, want: `<rect `},
		{name: "tall circle", options: []GeneratorOption{WithWidth(50), WithHeight(200)}, want: `r="25"`},
		{name: "wide circle with padding and border", options: []GeneratorOption{WithWidth(200), WithHeight(50), WithPadding(5), WithBorderWidth(2)}, want: `r="18"`},
		{name: "smallest room left", options: []GeneratorOption{WithPadding(23)}, want: `r="1"`},
		{name: "background color func", options: []GeneratorOption{WithBackgroundColorFunc(constant("#010203"))}, want: `fill="#010203"`},
		{name: "font color func", options: []GeneratorOption{WithFontColorFunc(constant("#010203"))}, want: `font-size="21" fill="#010203"`},
		{name: "border color func", options: []GeneratorOption{WithBorderColorFunc(constant("#010203"))}, want: `stroke="#010203"`},
		{name: "color wins over color func", options: []GeneratorOption{WithBackgroundColor("#123456"), WithBackgroundColorFunc(constant("#010203"))}, want: `fill="#123456"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GetSVG("AB", tt.options...)
			require.NoError(t, err)
			assert.Contains(t, svg.String(), tt.want)
		})
	}
}

func TestGetSVG_rejectsUnsafeOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []GeneratorOption
		want    error
	}{
		{name: "parent directory in family", options: []GeneratorOption{WithFont("../../x", 400)}, want: ErrInvalidFontFamily},
		{name: "path separator in family", options: []GeneratorOption{WithFont("a/b", 400)}, want: ErrInvalidFontFamily},
		{name: "windows separator in family", options: []GeneratorOption{WithFont(`a\b`, 400)}, want: ErrInvalidFontFamily},
		{name: "markup in family", options: []GeneratorOption{WithFont("x}</style><script>alert(1)</script>", 400)}, want: ErrInvalidFontFamily},
		{name: "url parameter in family", options: []GeneratorOption{WithFont("Inter&display=swap", 400)}, want: ErrInvalidFontFamily},
		{name: "weight zero", options: []GeneratorOption{WithFont("Inter", 0)}, want: ErrInvalidFontWeight},
		{name: "weight too high", options: []GeneratorOption{WithFont("Inter", 1001)}, want: ErrInvalidFontWeight},
		{name: "attribute in border dash", options: []GeneratorOption{WithBorderDash(`1" onload="alert(1)`)}, want: ErrInvalidBorderDash},
		{name: "empty border dash", options: []GeneratorOption{WithBorderDash("")}, want: ErrInvalidBorderDash},
		{name: "negative width", options: []GeneratorOption{WithWidth(-1)}, want: ErrInvalidSize},
		{name: "negative height", options: []GeneratorOption{WithHeight(-1)}, want: ErrInvalidSize},
		{name: "negative font size", options: []GeneratorOption{WithFontSize(-1)}, want: ErrInvalidSize},
		{name: "negative border width", options: []GeneratorOption{WithBorderWidth(-1)}, want: ErrInvalidSize},
		{name: "negative border radius", options: []GeneratorOption{WithBorderRadius(-1)}, want: ErrInvalidSize},
		{name: "negative padding", options: []GeneratorOption{WithPadding(-1)}, want: ErrInvalidSize},
		{name: "padding of half the size", options: []GeneratorOption{WithPadding(24)}, want: ErrInvalidSize},
		{name: "padding too large on a square", options: []GeneratorOption{WithShape("square"), WithPadding(100)}, want: ErrInvalidSize},
		{name: "border of half the size", options: []GeneratorOption{WithBorderWidth(24)}, want: ErrInvalidSize},
		{name: "padding and border filling the size", options: []GeneratorOption{WithPadding(12), WithBorderWidth(12)}, want: ErrInvalidSize},
		{name: "values whose sum overflows", options: []GeneratorOption{WithPadding(math.MaxInt), WithBorderWidth(math.MaxInt)}, want: ErrInvalidSize},
		{name: "size too small for a shape", options: []GeneratorOption{WithWidth(1), WithHeight(1)}, want: ErrInvalidSize},
		{name: "invalid background color", options: []GeneratorOption{WithBackgroundColor("nope")}, want: ErrInvalidColor},
		{name: "invalid font color", options: []GeneratorOption{WithFontColor("nope")}, want: ErrInvalidColor},
		{name: "invalid border color", options: []GeneratorOption{WithBorderColor(`"><script>`)}, want: ErrInvalidColor},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			svg, err := GetSVG("AB", append(tt.options, WithFontDir(dir))...)
			require.ErrorIs(t, err, tt.want)
			assert.Nil(t, svg)

			// Nothing is written, neither in the font directory nor next to it
			for _, d := range []string{dir, filepath.Dir(dir)} {
				entries, _ := os.ReadDir(d)
				for _, e := range entries {
					assert.False(t, strings.HasSuffix(e.Name(), "-400") || strings.HasSuffix(e.Name(), "-weight"),
						"file %s written in %s", e.Name(), d)
				}
			}
		})
	}
}

func TestGetSVG_acceptsValidOptions(t *testing.T) {
	tests := []struct {
		name    string
		options []GeneratorOption
		want    string
	}{
		{name: "default", want: `stroke-dasharray="none"`},
		{name: "dash with space", options: []GeneratorOption{WithBorderDash("6 2")}, want: `stroke-dasharray="6 2"`},
		{name: "dash with comma", options: []GeneratorOption{WithBorderDash("5,5")}, want: `stroke-dasharray="5,5"`},
		{name: "decimal dash", options: []GeneratorOption{WithBorderDash("1.5 3")}, want: `stroke-dasharray="1.5 3"`},
		{name: "square", options: []GeneratorOption{WithShape("square"), WithBorderDash("none")}, want: "<rect "},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GetSVG("AB", tt.options...)
			require.NoError(t, err)
			assert.Contains(t, svg.String(), tt.want)
		})
	}
}

func TestGetSVG_fontFamily(t *testing.T) {
	fontDir := t.TempDir()
	for _, file := range []string{"Inter-400", "Roboto Mono-400", "M PLUS 1p-400"} {
		require.NoError(t, os.WriteFile(filepath.Join(fontDir, file), []byte("@font-face {}\n"), 0o600))
	}

	tests := []struct {
		name    string
		options []GeneratorOption
		want    string
	}{
		{name: "no family", want: `font-family="sans-serif"`},
		{name: "one word", options: []GeneratorOption{WithFont("Inter", 400)}, want: `font-family="'Inter',sans-serif"`},
		{name: "several words", options: []GeneratorOption{WithFont("Roboto Mono", 400)}, want: `font-family="'Roboto Mono',sans-serif"`},
		{name: "word starting with a digit", options: []GeneratorOption{WithFont("M PLUS 1p", 400)}, want: `font-family="'M PLUS 1p',sans-serif"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svg, err := GetSVG("AB", append(tt.options, WithFontDir(fontDir))...)
			require.NoError(t, err)
			requireWellFormed(t, svg.String())
			assert.Contains(t, svg.String(), tt.want)
		})
	}
}

func TestColorFromString(t *testing.T) {
	tests := []struct {
		name    string
		color   string
		want    string
		wantErr bool
	}{
		{name: "name", color: "red", want: "#ff0000"},
		{name: "name in upper case", color: "RED", want: "#ff0000"},
		{name: "hex", color: "#00ff00", want: "#00ff00"},
		{name: "hex without sharp", color: "00ff00", want: "#00ff00"},
		{name: "short hex", color: "#0f0", want: "#00ff00"},
		{name: "short hex without sharp", color: "0f0", want: "#00ff00"},
		{name: "unknown name", color: "notacolor", wantErr: true},
		{name: "empty", color: "", wantErr: true},
		{name: "markup", color: `"><script>`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			color, err := colorFromString(tt.color)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, color.Hex())
		})
	}
}

func TestNewColorGenerator(t *testing.T) {
	generator := NewColorGenerator("AB")
	assert.Equal(t, "#c4b3a2", generator.GetBackgroundColor().Hex())
	assert.Equal(t, "#8d4806", generator.GetTextColor().Hex())
	assert.Equal(t, "#706f6f", generator.GetBorderColor().Hex())

	red := func(float64) colorful.Color { return colorful.Color{R: 1} }
	custom := NewColorGenerator("AB", BackgroundColorFunc(red), TextColorFunc(red), BorderColorFunc(red))
	assert.Equal(t, "#ff0000", custom.GetBackgroundColor().Hex())
	assert.Equal(t, "#ff0000", custom.GetTextColor().Hex())
	assert.Equal(t, "#ff0000", custom.GetBorderColor().Hex())
}

func TestNewColorGenerator_seed(t *testing.T) {
	seedOf := func(name string) float64 {
		var seed float64
		NewColorGenerator(name, BackgroundColorFunc(func(r float64) colorful.Color {
			seed = r
			return colorful.Color{}
		}))
		return seed
	}

	tests := []struct {
		name string
		want float64
	}{
		{name: "", want: 0},
		{name: "AB", want: 0.081},
		// The hash of these names overflows to a negative value
		{name: "abcdefghijklmn", want: 0.351},
		{name: "Jean-Baptiste Emmanuel Zorg de la Tour", want: 0.937},
		// The hash of this name overflows to a positive value
		{name: "abcdefghijklmnopqrstuvwxyz", want: 0.789},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			seed := seedOf(tt.name)
			assert.InDelta(t, tt.want, seed, 1e-9)
			assert.GreaterOrEqual(t, seed, 0.0)
			assert.Less(t, seed, 1.0)
		})
	}
}

func TestWithFont_url(t *testing.T) {
	opts := avatarGenerator{}
	WithFont("Roboto Mono", 700)(&opts)
	assert.Equal(t, "https://fonts.googleapis.com/css?family=Roboto+Mono:700", opts.fontUrl)
	assert.Equal(t, "Roboto Mono-700", opts.fontFile)
}

func TestDownloadFile(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte("@font-face {}"))
	}))
	defer server.Close()

	t.Run("ok", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "font")
		require.NoError(t, downloadFile(path, server.URL+"/font"))
		content, err := os.ReadFile(path)
		require.NoError(t, err)
		assert.Equal(t, "@font-face {}", string(content))
	})

	t.Run("error page is not kept", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "font")
		require.Error(t, downloadFile(path, server.URL+"/missing"))
		assert.NoFileExists(t, path)
	})
}
