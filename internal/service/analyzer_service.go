package service

import (
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"
	"webanalyzer/internal/log"
	"webanalyzer/internal/model"
)

var (
	html5Doctype = regexp.MustCompile(`(?i)<!DOCTYPE\s+html>`)
	html4Doctype = regexp.MustCompile(`(?i)<!DOCTYPE\s+HTML\s+PUBLIC\s+"[^"]*//DTD\s+HTML\s+4`)
	xhtmlDoctype = regexp.MustCompile(`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"[^"]*//DTD\s+XHTML`)
)

type linkResult struct {
	isInternal   bool
	isAccessible bool
}

const (
	maxLinkCheckWorkers = 20
	linkCheckTimeout    = 5 * time.Second
)

func AnalyzePage(targetURL string) *model.WebpageAnalysis {
	page := &model.WebpageAnalysis{}

	baseURL, err := url.Parse(targetURL)
	if err != nil {
		return page
	}

	root, rawHTML, err := fetchHTML(targetURL)
	if err != nil {
		return page
	}

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()
		page.HTMLVersion = detectHTMLVersion(rawHTML)
	}()

	go func() {
		defer wg.Done()
		page.PageTitle = extractTitle(root)
	}()

	var internal, external, inaccessible int
	go func() {
		defer wg.Done()
		links := extractLinks(root)
		internal, external, inaccessible = analyzeLinks(links, baseURL)
	}()

	go func() {
		defer wg.Done()
		page.HasLoginForm = hasLoginForm(root)
	}()

	wg.Wait()

	page.InternalLinkCount = internal
	page.ExternalLinkCount = external
	page.InaccessibleLinkCount = inaccessible

	return page
}

func fetchHTML(targetURL string) (*html.Node, string, error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Get(targetURL)
	if err != nil {
		log.Logger.Error("failed to fetch URL",
			zap.String("url", targetURL),
			zap.Error(err),
		)
		return nil, "", fmt.Errorf("failed to fetch URL: %w", err)
	}
	defer func() {
		if cerr := resp.Body.Close(); cerr != nil {
			log.Logger.Warn("failed to close response body", zap.Error(cerr))
		}
	}()

	if resp.StatusCode != http.StatusOK {
		log.Logger.Warn("unexpected status code",
			zap.String("url", targetURL),
			zap.Int("status_code", resp.StatusCode),
		)
		return nil, "", fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Logger.Warn("failed to read response body",
			zap.String("url", targetURL),
			zap.Error(err),
		)
		return nil, "", fmt.Errorf("failed to read response body: %w", err)
	}

	rawHTML := string(body)

	root, err := html.Parse(strings.NewReader(rawHTML))
	if err != nil {
		log.Logger.Error("failed to parse HTML",
			zap.String("url", targetURL),
			zap.Error(err),
		)
		return nil, "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	log.Logger.Info("successfully fetched and parsed HTML",
		zap.String("url", targetURL),
		zap.Int("content_length", len(rawHTML)),
		zap.Int("status_code", resp.StatusCode),
	)

	return root, rawHTML, nil
}

func detectHTMLVersion(rawHTML string) string {
	docStart := rawHTML
	if len(rawHTML) > 1000 {
		docStart = rawHTML[:1000]
	}

	switch {
	case html5Doctype.MatchString(docStart):
		return "HTML5"
	case xhtmlDoctype.MatchString(docStart):
		return "XHTML 1.0"
	case html4Doctype.MatchString(docStart):
		return "HTML 4.01"
	default:
		return "Unknown (possibly HTML5 without explicit DOCTYPE)"
	}
}

func extractTitle(node *html.Node) string {
	if node.Type == html.ElementNode && node.Data == "title" {
		if node.FirstChild != nil {
			return strings.TrimSpace(node.FirstChild.Data)
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		if title := extractTitle(c); title != "" {
			return title
		}
	}
	return ""
}

func extractHeadings(node *html.Node) model.HeadingCounts {
	var counts model.HeadingCounts

	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode {
			switch n.Data {
			case "h1":
				counts.H1++
			case "h2":
				counts.H2++
			case "h3":
				counts.H3++
			case "h4":
				counts.H4++
			case "h5":
				counts.H5++
			case "h6":
				counts.H6++
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(node)

	return counts
}

func extractLinks(node *html.Node) []string {
	var links []string
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "a" {
			for _, attr := range n.Attr {
				if attr.Key == "href" {
					links = append(links, attr.Val)
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(node)
	return links
}

func isInternalLink(link string, baseURL *url.URL) bool {
	if link == "" || strings.HasPrefix(link, "#") {
		return true
	}

	parsedLink, err := url.Parse(link)
	if err != nil {
		return false
	}
	resolvedLink := baseURL.ResolveReference(parsedLink)

	return strings.EqualFold(resolvedLink.Host, baseURL.Host)
}

func checkLinkAccessibility(ctx context.Context, link string, baseURL *url.URL) bool {
	parsedLink, err := url.Parse(link)
	if err != nil {
		return false
	}

	resolvedLink := baseURL.ResolveReference(parsedLink)

	scheme := strings.ToLower(resolvedLink.Scheme)
	if scheme == "" || scheme == "mailto" || scheme == "tel" || scheme == "javascript" {
		return true
	}

	if scheme != "http" && scheme != "https" {
		return true
	}

	client := &http.Client{
		Timeout: linkCheckTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, resolvedLink.String(), nil)
	if err != nil {
		return false
	}

	resp, err := client.Do(req)
	if err != nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodGet, resolvedLink.String(), nil)
		if err != nil {
			return false
		}
		resp, err = client.Do(req)
		if err != nil {
			return false
		}
	}
	defer resp.Body.Close()

	return resp.StatusCode < 400
}

func analyzeLinks(links []string, baseURL *url.URL) (internal, external, inaccessible int) {
	if len(links) == 0 {
		return 0, 0, 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	linkJobs := make(chan string, len(links))
	results := make(chan linkResult, len(links))

	numWorkers := maxLinkCheckWorkers
	if len(links) < numWorkers {
		numWorkers = len(links)
	}

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for link := range linkJobs {
				select {
				case <-ctx.Done():
					return
				default:
					isInternal := isInternalLink(link, baseURL)
					isAccessible := checkLinkAccessibility(ctx, link, baseURL)
					results <- linkResult{
						isInternal:   isInternal,
						isAccessible: isAccessible,
					}
				}
			}
		}()
	}

	go func() {
		for _, link := range links {
			linkJobs <- link
		}
		close(linkJobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if result.isInternal {
			internal++
		} else {
			external++
		}
		if !result.isAccessible {
			inaccessible++
		}
	}

	return
}

func hasLoginForm(node *html.Node) bool {
	var found bool
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && n.Data == "form" {
			if containsPasswordInput(n) {
				found = true
				return
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(node)
	return found
}

func containsPasswordInput(formNode *html.Node) bool {
	var found bool
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && n.Data == "input" {
			for _, attr := range n.Attr {
				if attr.Key == "type" && strings.ToLower(attr.Val) == "password" {
					found = true
					return
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(formNode)
	return found
}
