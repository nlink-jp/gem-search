// Package gemini provides a Vertex AI Gemini client with Google Search Grounding.
package gemini

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/nlink-jp/nlk/backoff"
	"github.com/nlink-jp/nlk/strip"
	"google.golang.org/genai"
)

const maxRetries = 5

// Source represents a grounding source from search results.
type Source struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Domain string `json:"domain"`
}

// Response holds the Gemini response with grounding metadata.
type Response struct {
	Text    string   // LLM-generated text
	Sources []Source // Grounding sources (URLs, titles)
	Queries []string // Web search queries used by Grounding
}

// Client wraps the Vertex AI Gemini client with Grounding enabled.
type Client struct {
	inner    *genai.Client
	model    string
	http     *http.Client
}

// NewClient creates a new Gemini client configured for Vertex AI.
func NewClient(ctx context.Context, project, location, model string) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  project,
		Location: location,
	})
	if err != nil {
		return nil, fmt.Errorf("creating Gemini client: %w", err)
	}

	return &Client{
		inner: client,
		model: model,
		http:  &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// Generate sends a prompt with Google Search Grounding and returns the response.
func (c *Client) Generate(ctx context.Context, systemPrompt, userPrompt string) (*Response, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(userPrompt, "user"),
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, "system"),
		Tools: []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		},
	}

	bo := backoff.New(
		backoff.WithBase(2*time.Second),
		backoff.WithMax(30*time.Second),
	)

	var lastErr error
	for attempt := range maxRetries + 1 {
		resp, err := c.inner.Models.GenerateContent(ctx, c.model, contents, config)
		if err == nil {
			return c.extractResponse(ctx, resp)
		}

		lastErr = err
		if !isRetryable(err) || attempt == maxRetries {
			return nil, fmt.Errorf("Gemini API call failed: %w", err)
		}

		wait := bo.Duration(attempt)
		log.Printf("Gemini call failed (attempt %d/%d), retrying in %v: %v",
			attempt+1, maxRetries+1, wait.Round(time.Second), err)
		time.Sleep(wait)
	}

	return nil, fmt.Errorf("Gemini API call failed after %d retries: %w", maxRetries, lastErr)
}

func (c *Client) extractResponse(ctx context.Context, resp *genai.GenerateContentResponse) (*Response, error) {
	if len(resp.Candidates) == 0 {
		return nil, fmt.Errorf("empty response from Gemini")
	}

	candidate := resp.Candidates[0]

	// Extract text
	var textParts []string
	if candidate.Content != nil {
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				textParts = append(textParts, part.Text)
			}
		}
	}
	text := strip.ThinkTags(strings.Join(textParts, ""))

	result := &Response{Text: text}

	// Extract grounding metadata
	if gm := candidate.GroundingMetadata; gm != nil {
		result.Queries = gm.WebSearchQueries

		for _, chunk := range gm.GroundingChunks {
			if chunk.Web != nil {
				result.Sources = append(result.Sources, Source{
					Title:  chunk.Web.Title,
					URL:    chunk.Web.URI,
					Domain: chunk.Web.Domain,
				})
			}
		}

		// Resolve redirect URIs
		c.resolveRedirects(ctx, result.Sources)
	}

	return result, nil
}

// resolveRedirects resolves Vertex AI grounding redirect URIs to actual URLs.
func (c *Client) resolveRedirects(ctx context.Context, sources []Source) {
	for i := range sources {
		if !strings.Contains(sources[i].URL, "grounding-api-redirect") {
			continue
		}

		resolved, err := c.resolveRedirect(ctx, sources[i].URL)
		if err != nil {
			log.Printf("redirect resolve failed for %s: %v", sources[i].URL, err)
			continue
		}
		sources[i].URL = resolved
	}
}

func (c *Client) resolveRedirect(ctx context.Context, uri string) (string, error) {
	// Use a client that doesn't follow redirects
	noRedirect := &http.Client{
		Timeout: 5 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, uri, nil)
	if err != nil {
		return "", err
	}

	resp, err := noRedirect.Do(req)
	if err != nil {
		return "", err
	}
	resp.Body.Close()

	if loc := resp.Header.Get("Location"); loc != "" {
		return loc, nil
	}
	return uri, nil
}

func isRetryable(err error) bool {
	errStr := strings.ToLower(err.Error())
	for _, k := range []string{"429", "503", "500", "unavailable", "timeout", "connection refused", "eof"} {
		if strings.Contains(errStr, k) {
			return true
		}
	}
	return false
}

// Model returns the configured model name.
func (c *Client) Model() string {
	return c.model
}
