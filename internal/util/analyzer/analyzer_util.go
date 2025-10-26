package analyzer

import (
	"golang.org/x/net/html"
	"regexp"
	"strings"
	"time"
)

var (
	Html5Doctype = regexp.MustCompile(`(?i)<!DOCTYPE\s+html>`)
	Html4Doctype = regexp.MustCompile(`(?i)<!DOCTYPE\s+HTML\s+PUBLIC\s+"[^"]*//DTD\s+HTML\s+4`)
	XhtmlDoctype = regexp.MustCompile(`(?i)<!DOCTYPE\s+html\s+PUBLIC\s+"[^"]*//DTD\s+XHTML`)
)

type LinkResult struct {
	IsInternal   bool
	IsAccessible bool
}

const (
	MaxLinkCheckWorkers = 20
	LinkCheckTimeout    = 5 * time.Second
)

// ExtractInnerText extracts all visible text content inside a node.
func ExtractInnerText(node *html.Node) string {
	var sb strings.Builder
	var traverse func(*html.Node)
	traverse = func(n *html.Node) {
		if n.Type == html.TextNode {
			sb.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			traverse(c)
		}
	}
	traverse(node)
	return sb.String()
}

// IsLinkTag checks whether the current node represents an <a> tag.
func IsLinkTag(node *html.Node) bool {
	return node.Type == html.ElementNode && node.Data == "a"
}

// GetHrefValue finds and returns the href attribute from an <a> tag.
// If there's no href, it returns an empty string.
func GetHrefValue(node *html.Node) string {
	for _, attr := range node.Attr {
		if attr.Key == "href" {
			return attr.Val
		}
	}
	return ""
}
