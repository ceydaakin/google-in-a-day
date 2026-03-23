package crawler

import (
	"html"
	"io"
	"net/url"
	"strings"
)

// maxPageSize limits how much of a page we read to prevent OOM on very large documents.
const maxPageSize = 5 * 1024 * 1024 // 5 MB

// parsePage extracts the title, body text, and links from raw HTML.
// Uses only the Go standard library — no external HTML parsers.
// Reads at most 5 MB per page to bound memory usage on large crawls.
func parsePage(body io.Reader, baseURL *url.URL) (title string, text string, links []string) {
	data, err := io.ReadAll(io.LimitReader(body, maxPageSize))
	if err != nil {
		return "", "", nil
	}
	html := string(data)

	rawTitle := extractBetweenTags(html, "title")
	title = cleanText(rawTitle)
	bodyContent := extractBetweenTags(html, "body")
	if bodyContent == "" {
		bodyContent = html
	}

	text = stripTags(bodyContent)
	links = extractLinks(html, baseURL)
	return title, text, links
}

// cleanText decodes HTML entities and normalizes whitespace.
func cleanText(s string) string {
	decoded := html.UnescapeString(s)
	fields := strings.Fields(decoded)
	return strings.Join(fields, " ")
}

// extractBetweenTags extracts text content between the first occurrence of <tag> and </tag>.
func extractBetweenTags(html, tag string) string {
	lower := strings.ToLower(html)
	startTag := "<" + tag
	endTag := "</" + tag

	start := strings.Index(lower, startTag)
	if start == -1 {
		return ""
	}
	// Find the closing > of the start tag
	closeAngle := strings.Index(lower[start:], ">")
	if closeAngle == -1 {
		return ""
	}
	contentStart := start + closeAngle + 1

	end := strings.Index(lower[contentStart:], endTag)
	if end == -1 {
		return ""
	}

	return html[contentStart : contentStart+end]
}

// stripTags removes HTML tags, script/style blocks, and returns clean text.
func stripTags(html string) string {
	// Remove script and style blocks
	result := removeTagBlock(html, "script")
	result = removeTagBlock(result, "style")

	// Remove all HTML tags
	var buf strings.Builder
	inTag := false
	for _, r := range result {
		if r == '<' {
			inTag = true
			continue
		}
		if r == '>' {
			inTag = false
			buf.WriteRune(' ')
			continue
		}
		if !inTag {
			buf.WriteRune(r)
		}
	}

	// Normalize whitespace
	text := buf.String()
	fields := strings.Fields(text)
	return strings.Join(fields, " ")
}

// removeTagBlock removes all content between <tag...> and </tag> including the tags.
func removeTagBlock(html, tag string) string {
	lower := strings.ToLower(html)
	result := html

	for {
		lowerResult := strings.ToLower(result)
		start := strings.Index(lowerResult, "<"+tag)
		if start == -1 {
			break
		}
		endTag := "</" + tag + ">"
		end := strings.Index(lowerResult[start:], endTag)
		if end == -1 {
			// No closing tag found, remove from start tag to end
			result = result[:start]
			break
		}
		result = result[:start] + result[start+end+len(endTag):]
		_ = lower // keep reference
	}
	return result
}

// extractLinks finds all href attributes in <a> tags and resolves them.
func extractLinks(html string, baseURL *url.URL) []string {
	var links []string
	seen := make(map[string]bool)

	lower := strings.ToLower(html)
	searchFrom := 0

	for {
		// Find next <a tag
		idx := strings.Index(lower[searchFrom:], "<a ")
		if idx == -1 {
			idx = strings.Index(lower[searchFrom:], "<a\t")
			if idx == -1 {
				idx = strings.Index(lower[searchFrom:], "<a\n")
				if idx == -1 {
					break
				}
			}
		}
		pos := searchFrom + idx

		// Find the closing > of this tag
		tagEnd := strings.Index(html[pos:], ">")
		if tagEnd == -1 {
			break
		}
		tagContent := html[pos : pos+tagEnd+1]
		searchFrom = pos + tagEnd + 1

		// Extract href value
		href := extractAttr(tagContent, "href")
		if href == "" {
			continue
		}

		resolved := resolveURL(href, baseURL)
		if resolved != "" && !seen[resolved] {
			seen[resolved] = true
			links = append(links, resolved)
		}
	}

	return links
}

// extractAttr extracts the value of a named attribute from a tag string.
func extractAttr(tag, attr string) string {
	lower := strings.ToLower(tag)
	idx := strings.Index(lower, attr+"=")
	if idx == -1 {
		return ""
	}

	valStart := idx + len(attr) + 1 // skip past attr=
	if valStart >= len(tag) {
		return ""
	}

	quote := tag[valStart]
	if quote == '"' || quote == '\'' {
		valStart++
		end := strings.IndexByte(tag[valStart:], quote)
		if end == -1 {
			return ""
		}
		return tag[valStart : valStart+end]
	}

	// Unquoted value — ends at space or >
	end := strings.IndexAny(tag[valStart:], " \t\n>")
	if end == -1 {
		return tag[valStart:]
	}
	return tag[valStart : valStart+end]
}

// resolveURL resolves a possibly relative href against a base URL.
func resolveURL(href string, base *url.URL) string {
	href = strings.TrimSpace(href)
	if href == "" || strings.HasPrefix(href, "#") || strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
		return ""
	}
	// Skip JavaScript template expressions that get misidentified as links
	if strings.ContainsAny(href, "'{}`") {
		return ""
	}

	parsed, err := url.Parse(href)
	if err != nil {
		return ""
	}

	resolved := base.ResolveReference(parsed)
	if resolved.Scheme != "http" && resolved.Scheme != "https" {
		return ""
	}

	resolved.Fragment = ""
	result := resolved.String()
	// Normalize: strip trailing slash for deduplication (except root "/")
	if len(result) > 1 && strings.HasSuffix(result, "/") {
		result = strings.TrimRight(result, "/")
	}
	return result
}
