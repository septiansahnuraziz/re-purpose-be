package service

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/repurpose-ai/re-purpose-be/internal/entity"
	"github.com/repurpose-ai/re-purpose-be/internal/logger"
	"github.com/repurpose-ai/re-purpose-be/utils"
	"github.com/sirupsen/logrus"
)

var (
	titlePattern       = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	scriptStylePattern = regexp.MustCompile(`(?is)<(script|style|noscript)[^>]*>.*?</(script|style|noscript)>`)
	tagPattern         = regexp.MustCompile(`(?s)<[^>]+>`)
	lineBreakPattern   = regexp.MustCompile(`(?i)<br\s*/?>|</p>|</div>|</h[1-6]>`)
)

// Extractor fetches an article URL and converts its HTML into plain text.
// TODO: swap this with a readability/trafilatura-grade extractor when wiring production.
type Extractor struct {
	client           *http.Client
	maxContentLength int64
}

func NewExtractor(timeout time.Duration, maxContentLength int64) *Extractor {
	return &Extractor{
		client:           &http.Client{Timeout: timeout},
		maxContentLength: maxContentLength,
	}
}

func (e *Extractor) Extract(ctx context.Context, rawURL string) (*entity.ExtractURLResponse, error) {
	rawURL = strings.TrimSpace(rawURL)
	if !utils.IsHTTPURL(rawURL) {
		return nil, ErrInvalidURL
	}

	parsed, err := url.Parse(rawURL)
	if err != nil {
		return nil, ErrInvalidURL
	}
	if utils.IsPrivateHost(parsed.Hostname()) {
		return nil, ErrInvalidURL
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}
	request.Header.Set("User-Agent", "RePurposeBot/1.0 (+https://repurpose.ai)")
	request.Header.Set("Accept", "text/html,application/xhtml+xml")

	logger.WithFields(ctx, logrus.Fields{"url": rawURL}).Debug("fetching url content")

	response, err := e.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}
	defer response.Body.Close()

	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		logger.WithFields(ctx, logrus.Fields{
			"url":        rawURL,
			"statusCode": response.StatusCode,
		}).Warn("url returned non-success status")
		return nil, fmt.Errorf("%w: upstream returned status %d", ErrExtractionFailed, response.StatusCode)
	}

	raw, err := io.ReadAll(io.LimitReader(response.Body, e.maxContentLength))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrExtractionFailed, err)
	}

	title := extractTitle(string(raw))
	content := htmlToText(string(raw))
	if content == "" {
		return nil, fmt.Errorf("%w: no readable content found", ErrExtractionFailed)
	}

	logger.WithFields(ctx, logrus.Fields{
		"url":       rawURL,
		"title":     title,
		"wordCount": utils.WordCount(content),
	}).Debug("url content extracted")

	return &entity.ExtractURLResponse{
		Title:     title,
		Content:   content,
		WordCount: utils.WordCount(content),
	}, nil
}

func extractTitle(document string) string {
	match := titlePattern.FindStringSubmatch(document)
	if len(match) < 2 {
		return ""
	}
	return utils.NormalizeWhitespace(html.UnescapeString(match[1]))
}

func htmlToText(document string) string {
	document = scriptStylePattern.ReplaceAllString(document, " ")
	document = lineBreakPattern.ReplaceAllString(document, "\n")
	document = tagPattern.ReplaceAllString(document, " ")
	document = html.UnescapeString(document)

	lines := strings.Split(document, "\n")
	cleaned := make([]string, 0, len(lines))
	for _, line := range lines {
		if normalized := utils.NormalizeWhitespace(line); normalized != "" {
			cleaned = append(cleaned, normalized)
		}
	}

	return strings.Join(cleaned, "\n")
}
