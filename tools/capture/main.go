// Command capture produces the images of the README from the real server.
//
// It starts the server on port 8080, requests each avatar of docs/examples/examples.json, keeps the answers next
// to this file and writes the table of examples between the markers of the README. It also takes
// screenshots of the page with a Chromium based browser, found on the machine or named by the BROWSER
// environment variable.
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

const (
	examplesDir  = "docs/examples"
	examplesFile = examplesDir + "/examples.json"
	readmeFile   = "README.md"
	serverAddr   = "localhost:8080"
	startMarker  = "<!-- examples:start -->"
	endMarker    = "<!-- examples:end -->"
)

// example is an avatar of the README
type example struct {
	File    string `json:"file"`
	Caption string `json:"caption"`
	Query   string `json:"query"`
}

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	content, err := os.ReadFile(examplesFile)
	if err != nil {
		return err
	}
	var examples []example
	if err := json.Unmarshal(content, &examples); err != nil {
		return fmt.Errorf("%s: %w", examplesFile, err)
	}
	browser, err := findBrowser()
	if err != nil {
		return err
	}

	temporary, err := os.MkdirTemp("", "avatar-capture")
	if err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(temporary)
	}()

	base, stop, err := startServer(temporary)
	if err != nil {
		return err
	}
	defer stop()

	rows := make([]string, 0, len(examples))
	for _, item := range examples {
		image, err := capture(base, item)
		if err != nil {
			return fmt.Errorf("%s: %w", item.File, err)
		}
		row, err := tableRow(item, image)
		if err != nil {
			return fmt.Errorf("%s: %w", item.File, err)
		}
		rows = append(rows, row)
		log.Printf("Captured %s", image)
	}

	page := base + "/?name=Ada+Lovelace&fontFamily=Inter&fontWeight=800"
	if err := screenshotPage(browser, temporary, page, "page-light.png"); err != nil {
		return err
	}
	if err := screenshotPage(browser, temporary, page, "page-dark.png", "--force-dark-mode"); err != nil {
		return err
	}

	return writeTable(rows)
}

// startServer builds and starts the server on a free port, and waits until it answers
func startServer(temporary string) (string, func(), error) {
	binary := filepath.Join(temporary, "avatar")
	if runtime.GOOS == "windows" {
		binary += ".exe"
	}
	build := exec.Command("go", "build", "-o", binary, "./cmd/avatar")
	build.Stderr = os.Stderr
	if err := build.Run(); err != nil {
		return "", nil, fmt.Errorf("build the server: %w", err)
	}

	// The screenshots of the page show its address: a fixed one keeps them stable
	listener, err := net.Listen("tcp", serverAddr)
	if err != nil {
		return "", nil, fmt.Errorf("stop what listens on %s: %w", serverAddr, err)
	}
	if err := listener.Close(); err != nil {
		return "", nil, err
	}

	server := exec.Command(binary, "-addr", serverAddr, "-fontdir", filepath.Join(temporary, "fonts")) // #nosec G204 -- the binary has just been built
	if err := server.Start(); err != nil {
		return "", nil, err
	}
	stop := func() {
		_ = server.Process.Kill()
		_ = server.Wait()
	}

	base := "http://" + serverAddr
	for range 50 {
		if _, err := get(base + "/health"); err == nil {
			return base, stop, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	stop()
	return "", nil, errors.New("the server does not answer")
}

// get returns the content of an url: any answer other than 200 OK is an error
func get(target string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = response.Body.Close()
	}()
	content, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s: %s: %s", target, response.Status, content)
	}
	return content, nil
}

// capture keeps the answer of the server and returns the name of the image
func capture(base string, item example) (string, error) {
	svg, err := get(base + "/api?" + item.Query)
	if err != nil {
		return "", err
	}
	if strings.Contains(item.Query, "fontFamily=") && !bytes.Contains(svg, []byte("@font-face")) {
		return "", errors.New("the font is not embedded: the font servers may be unreachable")
	}
	name := item.File + ".svg"
	return name, os.WriteFile(filepath.Join(examplesDir, name), svg, 0o600)
}

// screenshotPage keeps a view of the page
func screenshotPage(browser, temporary, target, name string, options ...string) error {
	path, err := filepath.Abs(filepath.Join(examplesDir, name))
	if err != nil {
		return err
	}
	if err := screenshot(browser, temporary, target, path, 1280, 800, options...); err != nil {
		return err
	}
	log.Printf("Captured %s", name)
	return nil
}

// screenshot asks the browser for the image of a page
func screenshot(browser, temporary, target, path string, width, height int, options ...string) error {
	if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	arguments := append([]string{
		"--headless=new", "--disable-gpu", "--hide-scrollbars", "--virtual-time-budget=5000",
		"--user-data-dir=" + filepath.Join(temporary, "browser"),
		fmt.Sprintf("--window-size=%d,%d", width, height),
		"--screenshot=" + path,
	}, options...)
	if err := exec.Command(browser, append(arguments, target)...).Run(); err != nil { // #nosec G204 -- the browser of the developer
		return fmt.Errorf("%s: %w", browser, err)
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%s did not write the screenshot: %w", browser, err)
	}
	return nil
}

// findBrowser returns the browser named by BROWSER, or else a Chromium based one
func findBrowser() (string, error) {
	if browser := os.Getenv("BROWSER"); browser != "" {
		return browser, nil
	}
	for _, name := range []string{"chromium", "chromium-browser", "google-chrome", "chrome", "msedge"} {
		if path, err := exec.LookPath(name); err == nil {
			return path, nil
		}
	}
	for _, root := range []string{os.Getenv("ProgramFiles"), os.Getenv("ProgramFiles(x86)")} {
		for _, path := range []string{`Google\Chrome\Application\chrome.exe`, `Microsoft\Edge\Application\msedge.exe`} {
			candidate := filepath.Join(root, path)
			if _, err := os.Stat(candidate); root != "" && err == nil { // #nosec G703 -- a fixed path under a system directory
				return candidate, nil
			}
		}
	}
	return "", errors.New("no browser found: set the BROWSER environment variable")
}

// tableRow returns the row of an example in the table of the README
func tableRow(item example, image string) (string, error) {
	code, err := goOptions(item.Query)
	if err != nil {
		return "", err
	}
	if code == "" {
		code = "No option"
	} else {
		code = "`" + code + "`"
	}
	return fmt.Sprintf(`| <img src="%s/%s" alt="%s" height="72"> | **%s**<br>`+"`/api?%s`"+`<br>%s |`,
		examplesDir, image, item.Caption, item.Caption, item.Query, code), nil
}

// goOptions returns the Go options equivalent to the query of an example
func goOptions(query string) (string, error) {
	values, err := url.ParseQuery(query)
	if err != nil {
		return "", err
	}
	text := func(key string) string { return strconv.Quote(values.Get(key)) }

	calls := []struct{ key, code string }{
		{"separator", "initials.WithSeparator(" + text("separator") + ")"},
		{"length", "initials.WithLength(" + values.Get("length") + ")"},
		{"shape", "avatar.WithShape(" + text("shape") + ")"},
		{"width", "avatar.WithWidth(" + values.Get("width") + ")"},
		{"height", "avatar.WithHeight(" + values.Get("height") + ")"},
		{"backgroundColor", "avatar.WithBackgroundColor(" + text("backgroundColor") + ")"},
		{"fontColor", "avatar.WithFontColor(" + text("fontColor") + ")"},
		{"fontSize", "avatar.WithFontSize(" + values.Get("fontSize") + ")"},
		{"borderWidth", "avatar.WithBorderWidth(" + values.Get("borderWidth") + ")"},
		{"borderColor", "avatar.WithBorderColor(" + text("borderColor") + ")"},
		{"borderDash", "avatar.WithBorderDash(" + text("borderDash") + ")"},
		{"borderRadius", "avatar.WithBorderRadius(" + values.Get("borderRadius") + ")"},
		{"padding", "avatar.WithPadding(" + values.Get("padding") + ")"},
	}

	known := map[string]bool{"name": true, "fontFamily": true, "fontWeight": true}
	var codes []string
	for _, call := range calls {
		known[call.key] = true
		if values.Has(call.key) {
			codes = append(codes, call.code)
		}
	}
	if values.Has("fontFamily") {
		weight := values.Get("fontWeight")
		if weight == "" {
			weight = "400"
		}
		codes = append(codes, "avatar.WithFont("+text("fontFamily")+", "+weight+")")
	}
	for key := range values {
		if !known[key] {
			return "", fmt.Errorf("no Go option for the parameter %s", key)
		}
	}
	return strings.Join(codes, ", "), nil
}

// writeTable replaces the table between the markers of the README
func writeTable(rows []string) error {
	content, err := os.ReadFile(readmeFile)
	if err != nil {
		return err
	}
	readme := string(content)
	start := strings.Index(readme, startMarker)
	end := strings.Index(readme, endMarker)
	if start < 0 || end < start {
		return fmt.Errorf("%s: markers not found", readmeFile)
	}
	table := "| Avatar | Request and Go options |\n| --- | --- |\n" + strings.Join(rows, "\n") + "\n"
	updated := readme[:start+len(startMarker)] + "\n\n" + table + "\n" + readme[end:]
	return os.WriteFile(readmeFile, []byte(updated), 0o600) // #nosec G703 -- a fixed file name
}
