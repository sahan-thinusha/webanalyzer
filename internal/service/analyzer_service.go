package service

import (
	"fmt"
	"go.uber.org/zap"
	"golang.org/x/net/context"
	"golang.org/x/net/html"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
	"webanalyzer/internal/log"
	"webanalyzer/internal/model"
	"webanalyzer/internal/util/analyzer"
)

// AnalyzePage analyzes the HTML content of a webpage at the given target URL.
// detect the HTML version, extract the page title, count of different headers, count internal, external, and inaccessible links, and has login form in the page
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
	wg.Add(5)

	go func() {
		defer wg.Done()
		page.HTMLVersion = detectHTMLVersion(rawHTML)
	}()

	go func() {
		defer wg.Done()
		page.PageTitle = extractTitle(root)
	}()

	go func() {
		defer wg.Done()
		page.HeadingCounts = extractHeadings(root)
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

// retrieves and parses the HTML content from the given URL
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

// analyze HTML version
func detectHTMLVersion(rawHTML string) string {
	docStart := rawHTML
	if len(rawHTML) > 1000 {
		docStart = rawHTML[:1000]
	}

	switch {
	case analyzer.Html5Doctype.MatchString(docStart):
		return "HTML5"
	case analyzer.XhtmlDoctype.MatchString(docStart):
		return "XHTML 1.0"
	case analyzer.Html4Doctype.MatchString(docStart):
		return "HTML 4.01"
	default:
		return "Unknown (possibly HTML5 without explicit DOCTYPE)"
	}
}

// fetch the title from the page
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

// extract the html headings in the page
func extractHeadings(root *html.Node) model.HeadingCounts {
	var counts model.HeadingCounts

	var visitNode func(*html.Node)
	visitNode = func(node *html.Node) {
		if node.Type == html.ElementNode {
			switch node.Data {
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

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visitNode(child)
		}
	}

	visitNode(root)
	return counts
}

// extract the links in the html
func extractLinks(root *html.Node) []string {
	var links []string

	var visitNode func(*html.Node)
	visitNode = func(node *html.Node) {
		if analyzer.IsLinkTag(node) {
			href := analyzer.GetHrefValue(node)
			if href != "" {
				links = append(links, href)
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			visitNode(child)
		}
	}

	visitNode(root)
	return links
}

// check wheaten link is internal or not
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

// checks if a given link is accessible by attempting to send a HEAD request
// If the HEAD request fails, it sends a GET request as a fallback
// The function checks whether the link has an acceptable URL scheme, resolves relative links, and handles redirects
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
		Timeout: analyzer.LinkCheckTimeout,
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

// checks the accessibility of a list of links, categorizing them as internal or external
func analyzeLinks(links []string, baseURL *url.URL) (internal, external, inaccessible int) {
	if len(links) == 0 {
		return 0, 0, 0
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	linkJobs := make(chan string, len(links))
	results := make(chan analyzer.LinkResult, len(links))

	numWorkers := analyzer.MaxLinkCheckWorkers
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
					results <- analyzer.LinkResult{
						IsInternal:   isInternal,
						IsAccessible: isAccessible,
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
		if result.IsInternal {
			internal++
		} else {
			external++
		}
		if !result.IsAccessible {
			inaccessible++
		}
	}

	return
}

// checks whether the given HTML document contains a form element with a password input field
func hasLoginForm(node *html.Node) bool {
	loginKeywords := []string{"login", "log in", "sign in", "signin"}
	authInputs := []string{"password", "otp", "code"}

	var found bool
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if found {
			return
		}
		if n.Type == html.ElementNode && n.Data == "form" {
			if containsAuthIndicators(n, authInputs, loginKeywords) {
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

func containsAuthIndicators(formNode *html.Node, authInputs, loginKeywords []string) bool {
	var found bool

	var walk func(*html.Node)
	walk = func(node *html.Node) {
		if found {
			return
		}

		if node.Type == html.ElementNode {
			tag := node.Data

			if tag == "input" && hasAuthInput(node, authInputs) {
				found = true
				return
			}

			if tag == "button" || (tag == "input" && isSubmitButton(node)) {
				if hasLoginKeyword(node, loginKeywords) {
					found = true
					return
				}
			}
		}

		for child := node.FirstChild; child != nil; child = child.NextSibling {
			walk(child)
		}
	}

	walk(formNode)
	return found
}

// check has auth input type
func hasAuthInput(node *html.Node, authInputs []string) bool {
	for _, attr := range node.Attr {
		val := strings.ToLower(attr.Val)
		for _, authInput := range authInputs {
			if strings.Contains(val, authInput) {
				return true
			}
		}
	}
	return false
}

// check about submit button
func isSubmitButton(node *html.Node) bool {

	for _, attr := range node.Attr {
		if attr.Key == "type" && strings.ToLower(attr.Val) == "submit" {
			return true
		}
	}
	return false
}

// check about login related keywords
func hasLoginKeyword(node *html.Node, loginKeywords []string) bool {
	text := strings.ToLower(analyzer.ExtractInnerText(node))
	for _, attr := range node.Attr {
		text += " " + strings.ToLower(attr.Val)
	}
	for _, kw := range loginKeywords {
		if strings.Contains(text, kw) {
			return true
		}
	}
	return false
}
