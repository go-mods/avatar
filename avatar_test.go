package avatar

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			svg, err := GetSVG("AB", append(tt.options, WithFontDir(dir))...)
			if !errors.Is(err, tt.want) {
				t.Fatalf("GetSVG() error = %v, want %v", err, tt.want)
			}
			if svg != nil {
				t.Errorf("GetSVG() returned an svg with an error")
			}
			// Nothing is written, neither in the font directory nor next to it
			for _, d := range []string{dir, filepath.Dir(dir)} {
				entries, _ := os.ReadDir(d)
				for _, e := range entries {
					if strings.HasSuffix(e.Name(), "-400") || strings.HasSuffix(e.Name(), "-weight") {
						t.Errorf("file %s written in %s", e.Name(), d)
					}
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
			if err != nil {
				t.Fatalf("GetSVG() error = %v", err)
			}
			if !strings.Contains(svg.String(), tt.want) {
				t.Errorf("GetSVG() = %s, want it to contain %s", svg.String(), tt.want)
			}
		})
	}
}

func TestWithFont_url(t *testing.T) {
	opts := avatarGenerator{}
	WithFont("Roboto Mono", 700)(&opts)
	if want := "https://fonts.googleapis.com/css?family=Roboto+Mono:700"; opts.fontUrl != want {
		t.Errorf("fontUrl = %q, want %q", opts.fontUrl, want)
	}
	if want := "Roboto Mono-700"; opts.fontFile != want {
		t.Errorf("fontFile = %q, want %q", opts.fontFile, want)
	}
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
		if err := downloadFile(path, server.URL+"/font"); err != nil {
			t.Fatalf("downloadFile() error = %v", err)
		}
		if content, _ := os.ReadFile(path); string(content) != "@font-face {}" {
			t.Errorf("file content = %q", content)
		}
	})

	t.Run("error page is not kept", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "font")
		if err := downloadFile(path, server.URL+"/missing"); err == nil {
			t.Fatal("downloadFile() error = nil, want an error for a 404")
		}
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("file kept after a 404: %v", err)
		}
	})
}
