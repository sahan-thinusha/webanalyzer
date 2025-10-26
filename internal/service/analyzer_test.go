package service

import (
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/net/html"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
	"webanalyzer/internal/log"
	"webanalyzer/internal/model"
)

func TestDetectHTMLVersion(t *testing.T) {
	tests := []struct {
		name     string
		rawHTML  string
		expected string
	}{
		{
			name:     "HTML5 doctype",
			rawHTML:  "<!DOCTYPE html><html><head><title>Test</title></head></html>",
			expected: "HTML5",
		},
		{
			name:     "HTML5 doctype case insensitive",
			rawHTML:  "<!doctype HTML><html><head><title>Test</title></head></html>",
			expected: "HTML5",
		},
		{
			name:     "XHTML 1.0 doctype",
			rawHTML:  `<!DOCTYPE html PUBLIC "-//W3C//DTD XHTML 1.0 Strict//EN" "http://www.w3.org/TR/xhtml1/DTD/xhtml1-strict.dtd"><html></html>`,
			expected: "XHTML 1.0",
		},
		{
			name:     "HTML 4.01 doctype",
			rawHTML:  `<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd"><html></html>`,
			expected: "HTML 4.01",
		},
		{
			name:     "No doctype",
			rawHTML:  "<html><head><title>Test</title></head></html>",
			expected: "Unknown (possibly HTML5 without explicit DOCTYPE)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := detectHTMLVersion(tt.rawHTML)
			if result != tt.expected {
				t.Errorf("detectHTMLVersion() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractTitle(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		expected string
	}{
		{
			name:     "Valid title",
			htmlStr:  "<html><head><title>Test Title</title></head></html>",
			expected: "Test Title",
		},
		{
			name:     "Title with whitespace",
			htmlStr:  "<html><head><title>  Test Title 2  </title></head></html>",
			expected: "Test Title 2",
		},
		{
			name:     "No title tag",
			htmlStr:  "<html><head></head></html>",
			expected: "",
		},
		{
			name:     "Empty title tag",
			htmlStr:  "<html><head><title></title></head></html>",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}
			result := extractTitle(node)
			if result != tt.expected {
				t.Errorf("extractTitle() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractHeadings(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		expected model.HeadingCounts
	}{
		{
			name:    "All heading levels",
			htmlStr: "<html><body><h1>H1</h1><h2>H2</h2><h2>H2-2</h2><h3>H3</h3><h4>H4</h4><h5>H5</h5><h6>H6</h6></body></html>",
			expected: model.HeadingCounts{
				H1: 1,
				H2: 2,
				H3: 1,
				H4: 1,
				H5: 1,
				H6: 1,
			},
		},
		{
			name:    "No headings",
			htmlStr: "<html><body><p>Just a paragraph</p></body></html>",
			expected: model.HeadingCounts{
				H1: 0,
				H2: 0,
				H3: 0,
				H4: 0,
				H5: 0,
				H6: 0,
			},
		},
		{
			name:    "Multiple same level headings",
			htmlStr: "<html><body><h1>Title 1</h1><h1>Title 2</h1><h1>Title 3</h1></body></html>",
			expected: model.HeadingCounts{
				H1: 3,
				H2: 0,
				H3: 0,
				H4: 0,
				H5: 0,
				H6: 0,
			},
		},
		{
			name:    "Nested headings",
			htmlStr: "<html><body><div><h1>H1</h1><div><h2>H2</h2></div></div></body></html>",
			expected: model.HeadingCounts{
				H1: 1,
				H2: 1,
				H3: 0,
				H4: 0,
				H5: 0,
				H6: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}
			result := extractHeadings(node)
			if result != tt.expected {
				t.Errorf("extractHeadings() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractLinks(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		expected []string
	}{
		{
			name:     "Single link",
			htmlStr:  `<html><body><a href="https://example.com">Link</a></body></html>`,
			expected: []string{"https://example.com"},
		},
		{
			name:     "Multiple links",
			htmlStr:  `<html><body><a href="/page1">Link1</a><a href="/page2">Link2</a></body></html>`,
			expected: []string{"/page1", "/page2"},
		},
		{
			name:     "No links",
			htmlStr:  `<html><body><p>No links here</p></body></html>`,
			expected: []string{},
		},
		{
			name:     "Link without href",
			htmlStr:  `<html><body><a>No href</a></body></html>`,
			expected: []string{},
		},
		{
			name:     "Mixed link types",
			htmlStr:  `<html><body><a href="#anchor">Anchor</a><a href="mailto:test@example.com">Email</a><a href="https://external.com">External</a></body></html>`,
			expected: []string{"#anchor", "mailto:test@example.com", "https://external.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}
			result := extractLinks(node)
			if len(result) != len(tt.expected) {
				t.Errorf("extractLinks() returned %d links, want %d", len(result), len(tt.expected))
				return
			}
			for i, link := range result {
				if link != tt.expected[i] {
					t.Errorf("extractLinks()[%d] = %v, want %v", i, link, tt.expected[i])
				}
			}
		})
	}
}

func TestIsInternalLink(t *testing.T) {
	baseURL, _ := url.Parse("https://example.com/page")

	tests := []struct {
		name     string
		link     string
		expected bool
	}{
		{
			name:     "Same domain absolute URL",
			link:     "https://example.com/other",
			expected: true,
		},
		{
			name:     "Same domain different path",
			link:     "/about",
			expected: true,
		},
		{
			name:     "Anchor link",
			link:     "#section",
			expected: true,
		},
		{
			name:     "Empty link",
			link:     "",
			expected: true,
		},
		{
			name:     "External domain",
			link:     "https://external.com/page",
			expected: false,
		},
		{
			name:     "Subdomain",
			link:     "https://sub.example.com/page",
			expected: false,
		},
		{
			name:     "Relative URL",
			link:     "page2.html",
			expected: true,
		},
		{
			name:     "Case insensitive domain",
			link:     "https://EXAMPLE.COM/page",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isInternalLink(tt.link, baseURL)
			if result != tt.expected {
				t.Errorf("isInternalLink(%q) = %v, want %v", tt.link, result, tt.expected)
			}
		})
	}
}

func TestCheckLinkAccessibility(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/ok":
			w.WriteHeader(http.StatusOK)
		case "/not-found":
			w.WriteHeader(http.StatusNotFound)
		case "/redirect":
			w.Header().Set("Location", "/ok")
			w.WriteHeader(http.StatusMovedPermanently)
		default:
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)

	tests := []struct {
		name     string
		link     string
		expected bool
	}{
		{
			name:     "Accessible link",
			link:     server.URL + "/ok",
			expected: true,
		},
		{
			name:     "404 link",
			link:     server.URL + "/not-found",
			expected: false,
		},
		{
			name:     "Mailto link",
			link:     "mailto:test@example.com",
			expected: true,
		},
		{
			name:     "Tel link",
			link:     "tel:+1234567890",
			expected: true,
		},
		{
			name:     "Javascript link",
			link:     "javascript:void(0)",
			expected: true,
		},
		{
			name:     "Anchor link",
			link:     "#section",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, _ := context.WithTimeout(context.Background(), 30*time.Second)
			result := checkLinkAccessibility(ctx, tt.link, baseURL)
			if result != tt.expected {
				t.Errorf("checkLinkAccessibility(%q) = %v, want %v", tt.link, result, tt.expected)
			}
		})
	}
}

func TestAnalyzeLinks(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/accessible" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	baseURL, _ := url.Parse(server.URL)

	tests := []struct {
		name                 string
		links                []string
		expectedInternal     int
		expectedExternal     int
		expectedInaccessible int
	}{
		{
			name: "Mixed links",
			links: []string{
				"/page1",
				"/page2",
				"https://external.com/page",
				server.URL + "/accessible",
			},
			expectedInternal:     3,
			expectedExternal:     1,
			expectedInaccessible: 3,
		},
		{
			name:                 "No links",
			links:                []string{},
			expectedInternal:     0,
			expectedExternal:     0,
			expectedInaccessible: 0,
		},
		{
			name: "All internal",
			links: []string{
				"/page1",
				"/page2",
				"#anchor",
			},
			expectedInternal:     3,
			expectedExternal:     0,
			expectedInaccessible: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			internal, external, inaccessible := analyzeLinks(tt.links, baseURL)
			if internal != tt.expectedInternal {
				t.Errorf("analyzeLinks() internal = %d, want %d", internal, tt.expectedInternal)
			}
			if external != tt.expectedExternal {
				t.Errorf("analyzeLinks() external = %d, want %d", external, tt.expectedExternal)
			}
			if inaccessible != tt.expectedInaccessible {
				t.Errorf("analyzeLinks() inaccessible = %d, want %d", inaccessible, tt.expectedInaccessible)
			}
		})
	}
}

func TestHasLoginForm(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		expected bool
	}{
		{
			name: "Form with password input",
			htmlStr: `<html><body>
				<form>
					<input type="text" name="username">
					<input type="password" name="password">
				</form>
			</body></html>`,
			expected: true,
		},
		{
			name: "Form without password input",
			htmlStr: `<html><body>
				<form>
					<input type="text" name="search">
					<input type="submit">
				</form>
			</body></html>`,
			expected: false,
		},
		{
			name:     "No form",
			htmlStr:  `<html><body><p>No forms here</p></body></html>`,
			expected: false,
		},
		{
			name: "Password input case insensitive",
			htmlStr: `<html><body>
				<form>
					<input type="PASSWORD" name="password">
				</form>
			</body></html>`,
			expected: true,
		},
		{
			name: "Multiple forms, one with password",
			htmlStr: `<html><body>
				<form><input type="text"></form>
				<form><input type="password"></form>
			</body></html>`,
			expected: true,
		},
		{
			name: "Nested password input",
			htmlStr: `<html><body>
				<form>
					<div>
						<input type="password" name="password">
					</div>
				</form>
			</body></html>`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}
			result := hasLoginForm(node)
			if result != tt.expected {
				t.Errorf("hasLoginForm() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsPasswordInput(t *testing.T) {
	tests := []struct {
		name     string
		htmlStr  string
		expected bool
	}{
		{
			name: "Has password input",
			htmlStr: `<form>
				<input type="password" name="pwd">
			</form>`,
			expected: true,
		},
		{
			name: "No password input",
			htmlStr: `<form>
				<input type="text" name="username">
			</form>`,
			expected: false,
		},
		{
			name: "Nested password input",
			htmlStr: `<form>
				<div><input type="password"></div>
			</form>`,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := html.Parse(strings.NewReader(tt.htmlStr))
			if err != nil {
				t.Fatalf("Failed to parse HTML: %v", err)
			}
			var formNode *html.Node
			var findForm func(*html.Node)
			findForm = func(n *html.Node) {
				if n.Type == html.ElementNode && n.Data == "form" {
					formNode = n
					return
				}
				for c := n.FirstChild; c != nil; c = c.NextSibling {
					findForm(c)
				}
			}
			findForm(node)

			if formNode == nil {
				t.Fatal("Form node not found")
			}

			result := containsPasswordInput(formNode)
			if result != tt.expected {
				t.Errorf("containsPasswordInput() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestFetchHTML(t *testing.T) {
	log.Logger, _ = zap.NewDevelopment()
	defer log.Logger.Sync()
	tests := []struct {
		name           string
		serverResponse func(w http.ResponseWriter, r *http.Request)
		expectError    bool
	}{
		{
			name: "Successful fetch",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w, "<html><head><title>Test</title></head></html>")
			},
			expectError: false,
		},
		{
			name: "404 response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNotFound)
			},
			expectError: true,
		},
		{
			name: "500 response",
			serverResponse: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(tt.serverResponse))
			defer server.Close()

			node, rawHTML, err := fetchHTML(server.URL)

			if tt.expectError {
				if err == nil {
					t.Error("fetchHTML() expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("fetchHTML() unexpected error: %v", err)
				}
				if node == nil {
					t.Error("fetchHTML() returned nil node")
				}
				if rawHTML == "" {
					t.Error("fetchHTML() returned empty rawHTML")
				}
			}
		})
	}
}
