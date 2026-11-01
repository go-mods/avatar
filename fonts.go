package avatar

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const (
	// maxFontFamilyLength is longer than any Google Fonts family name
	maxFontFamilyLength = 64

	// maxFontResponseSize limits what is read from a response: a font reduced to a few letters
	// is much smaller
	maxFontResponseSize = 2 << 20

	// maxFontFaces limits the fonts read for one family and one weight
	maxFontFaces = 4

	// fontCacheSuffix ends the name of the cached files
	fontCacheSuffix = ".font.css"
)

var (
	// fontCSSURL is the Google Fonts API which describes a font
	fontCSSURL = "https://fonts.googleapis.com/css"

	// fontFilePrefix starts the url of every font file: the font of a style comes from nowhere else
	fontFilePrefix = "https://fonts.gstatic.com/"

	// maxCachedFonts limits the files of the font directory, a server receiving initials of any
	// kind would fill it otherwise. Once reached, the fonts are downloaded again at each use.
	maxCachedFonts = 10000

	// httpClient is used to download the fonts; the timeout prevents a slow or unresponsive
	// server from blocking the generation indefinitely.
	httpClient = &http.Client{Timeout: 30 * time.Second}

	// fontURLRegex matches the location of a font in a style
	fontURLRegex = regexp.MustCompile(`url\([^)]*\)`)

	// errFontUnavailable marks a failure of the network or of the font servers, which says nothing
	// about the font itself
	errFontUnavailable = errors.New("font unavailable")
)

// fontStyle returns the style which embeds the font, reduced to the letters of the initials: an svg
// displayed as an image cannot load anything from outside. The style is read from the font
// directory, or else downloaded and kept there. The style is empty when the font holds none of the
// letters or when the font servers cannot be reached: the text is then drawn with the fallback font.
func fontStyle(ctx context.Context, opts avatarGenerator) (string, error) {
	path := filepath.Join(opts.FontDir, fontCacheFile(opts.FontFamily, opts.FontWeight, opts.Initials))

	// An empty file keeps the answer of a font without these letters; any other file which is not
	// a valid style is replaced
	cached, err := os.ReadFile(path) // #nosec G304 -- the file name is built from a validated family and a hash
	if err == nil && (len(cached) == 0 || checkFontStyle(string(cached)) == nil) {
		return string(cached), nil
	}
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	style, err := downloadFontStyle(ctx, opts.FontFamily, opts.FontWeight, opts.Initials)
	// The avatar stays useful without its font, and the font is downloaded at the next call
	if errors.Is(err, errFontUnavailable) && ctx.Err() == nil {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if err := cacheFontStyle(path, style); err != nil {
		return "", err
	}
	return style, nil
}

// fontCacheFile returns the name of the file of a font. The initials are hashed: they can hold
// any character and they are not checked.
func fontCacheFile(family string, weight int, initials string) string {
	return fmt.Sprintf("%s-%d-%x%s", family, weight, sha256.Sum256([]byte(initials)), fontCacheSuffix)
}

// downloadFontStyle downloads the style of a font, then each font it refers to, which it embeds
func downloadFontStyle(ctx context.Context, family string, weight int, initials string) (string, error) {
	styleURL := fmt.Sprintf("%s?family=%s:%d&text=%s", fontCSSURL, url.QueryEscape(family), weight, url.QueryEscape(initials))
	content, err := download(ctx, styleURL)
	if err != nil {
		return "", err
	}
	style := string(content)

	locations := fontURLRegex.FindAllString(style, -1)
	if len(locations) == 0 || len(locations) > maxFontFaces {
		return "", fmt.Errorf("download %s: %d fonts found", styleURL, len(locations))
	}
	for _, location := range locations {
		fontURL := strings.TrimSuffix(strings.TrimPrefix(location, "url("), ")")
		if !strings.HasPrefix(fontURL, fontFilePrefix) {
			return "", fmt.Errorf("download %s: unexpected font location %q", styleURL, fontURL)
		}
		font, err := download(ctx, fontURL)
		// The server refuses to build a font reduced to letters it does not hold
		var status *statusError
		if errors.As(err, &status) && status.code == http.StatusBadRequest {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		embedded := "url(data:font/ttf;base64," + base64.StdEncoding.EncodeToString(font) + ")"
		style = strings.ReplaceAll(style, location, embedded)
	}

	if err := checkFontStyle(style); err != nil {
		return "", fmt.Errorf("download %s: %w", styleURL, err)
	}
	return style, nil
}

// checkFontStyle checks that a style can be written as is in the svg: it holds no markup
func checkFontStyle(style string) error {
	if !strings.Contains(style, "@font-face") {
		return errors.New("not a font style")
	}
	if strings.ContainsAny(style, "<>&") {
		return errors.New("markup in the font style")
	}
	return nil
}

// statusError is the error of a server which does not answer 200 OK
type statusError struct {
	url    string
	code   int
	status string
}

// Error returns the message of the error
func (e *statusError) Error() string {
	return fmt.Sprintf("download %s: %s", e.url, e.status)
}

// download returns the content of an url. An error page or a partial content is an error, so that
// it is never used as a font.
func download(ctx context.Context, fileURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, fileURL, nil)
	if err != nil {
		return nil, err
	}
	response, err := httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errFontUnavailable, err)
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(response.Body)

	if response.StatusCode >= http.StatusInternalServerError {
		return nil, fmt.Errorf("%w: %w", errFontUnavailable, &statusError{url: fileURL, code: response.StatusCode, status: response.Status})
	}
	if response.StatusCode != http.StatusOK {
		return nil, &statusError{url: fileURL, code: response.StatusCode, status: response.Status}
	}
	content, err := io.ReadAll(io.LimitReader(response.Body, maxFontResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("%w: download %s: %w", errFontUnavailable, fileURL, err)
	}
	if len(content) > maxFontResponseSize {
		return nil, fmt.Errorf("download %s: more than %d bytes", fileURL, maxFontResponseSize)
	}
	return content, nil
}

// cacheFontStyle keeps a style in the font directory, which is created if needed. The file appears
// once complete: a concurrent call or a crash never leaves a truncated font behind.
func cacheFontStyle(path, style string) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o750); err != nil {
		return err
	}

	cached, err := filepath.Glob(filepath.Join(directory, "*"+fontCacheSuffix))
	if err != nil {
		return err
	}
	if len(cached) >= maxCachedFonts {
		return nil
	}

	file, err := os.CreateTemp(directory, ".font-*")
	if err != nil {
		return err
	}
	temporary := file.Name()
	defer func() {
		_ = os.Remove(temporary)
	}()

	_, err = file.WriteString(style)
	if closeErr := file.Close(); err == nil {
		err = closeErr
	}
	if err != nil {
		return err
	}

	// On Windows, renaming over a file in use fails: another call has just written the same font
	if err := os.Rename(temporary, path); err != nil {
		if _, statErr := os.Stat(path); statErr != nil {
			return err
		}
	}
	return nil
}
