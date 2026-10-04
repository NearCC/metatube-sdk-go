// Package resolve — JavBus direct scraper.
//
// The SDK's JavBus provider goes through colly and the Engine's default
// fetcher, which Cloudflare aggressively blocks because the Go TLS /
// header fingerprint doesn't match a real browser.
//
// We bypass both: build our own http.Client with a full set of browser
// headers + cookie jar, GET the movie page, parse the actor names out of
// the HTML directly.
package resolve

import (
	"bytes"
	"compress/gzip"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/net/html"
)

// resolveProxy picks the best available proxy in this order:
//
//  1. -proxy flag (caller wires ProxyURL directly into the transport)
//  2. HTTPS_PROXY / HTTP_PROXY env vars (most common on macOS/Linux)
//  3. WinINET system proxy via golang.org/x/sys/windows registry —
//     this is what "Settings -> Network -> Proxy" writes on Windows,
//     and what most VPN clients (Clash / V2RayN / Surge with
//     "system proxy" enabled) ultimately populate.
//
// We avoid pulling x/sys/windows into a non-Windows build with a
// build-tagged stub. On non-Windows builds the env var path is the
// only option — which is fine, because non-Windows desktops don't
// typically have a system-wide proxy that env vars miss.

const javbusBase = "https://www.javbus.com"

var browserHeaders = map[string]string{
	"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36",
	"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
	"Accept-Language": "zh-CN,zh;q=0.9,en;q=0.8,ja;q=0.7",
	"Accept-Encoding": "gzip, deflate, br",
	"Sec-Fetch-Dest":  "document",
	"Sec-Fetch-Mode":  "navigate",
	"Sec-Fetch-Site":  "none",
	"Sec-Fetch-User":  "?1",
	"Upgrade-Insecure-Requests": "1",
}

// javbusClient is shared across all Resolver calls; cookie jar persists
// cf_clearance across requests within a process lifetime.
type javbusHTTP struct {
	client *http.Client
}

// newJavbusHTTP builds a Cloudflare-friendly client. proxyOverride is
// an optional explicit proxy URL from the -proxy flag. If empty, the
// transport falls back to http.ProxyFromEnvironment — same as the
// SDK's cleanhttp.DefaultPooledClient, which the rest of the SDK uses
// (and which is what makes metatube-server work without any
// proxy configuration on a typical desktop).
func newJavbusHTTP(proxyOverride string) *javbusHTTP {
	jar, _ := cookiejar.New(nil)
	_ = jar // reserved for future cf_clearance injection
	if env := os.Getenv("JAVBUS_COOKIE"); env != "" {
		u, _ := url.Parse(javbusBase)
		for _, p := range strings.Split(env, ";") {
			p = strings.TrimSpace(p)
			if p == "" {
				continue
			}
			kv := strings.SplitN(p, "=", 2)
			if len(kv) != 2 {
				continue
			}
			jar.SetCookies(u, []*http.Cookie{{
				Name:  strings.TrimSpace(kv[0]),
				Value: strings.TrimSpace(kv[1]),
			}})
		}
	}
	// Start from the SDK's own pooled client — it already has
	// ProxyFromEnvironment wired up. Then layer our own timeouts on
	// top. This way we behave identically to metatube-server when no
	// -proxy flag is given.
	client := cleanhttp.DefaultPooledClient()
	// Fail fast. Cloudflare rate-limits typically reject in <1s
	// (HTTP 403 / EOF on TLS handshake). We don't want a single bad
	// request to hold up the negative-cache machinery for 30s.
	client.Timeout = 12 * time.Second
	client.Jar = jar
	if tr, ok := client.Transport.(*http.Transport); ok {
		tr.MaxIdleConns = 10
		tr.MaxIdleConnsPerHost = 5
		tr.IdleConnTimeout = 30 * time.Second
		tr.TLSHandshakeTimeout = 10 * time.Second
		tr.ResponseHeaderTimeout = 15 * time.Second
		tr.ExpectContinueTimeout = 1 * time.Second
	}
	if proxyOverride != "" {
		if u, err := url.Parse(proxyOverride); err == nil && u.Host != "" {
			if tr, ok := client.Transport.(*http.Transport); ok {
				tr.Proxy = http.ProxyURL(u)
			}
		}
	}
	return &javbusHTTP{client: client}
}

func (j *javbusHTTP) get(ctx context.Context, url string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return "", err
	}
	for k, v := range browserHeaders {
		req.Header.Set(k, v)
	}
	req.Header.Set("Referer", javbusBase+"/")
	// existmag=all is the canonical "show me everything" cookie JavBus
	// expects on its movie pages.
	req.AddCookie(&http.Cookie{Name: "existmag", Value: "all"})

	resp, err := j.client.Do(req)
	if err != nil {
		if debugJavbus {
			fmt.Fprintf(os.Stderr, "[javbus] GET %s -> ERROR: %v\n", url, err)
		}
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 403 || resp.StatusCode == 503 {
		if debugJavbus {
			peek, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
			fmt.Fprintf(os.Stderr, "[javbus] %s -> HTTP %d, body: %q\n", url, resp.StatusCode, string(peek))
		}
		return "", fmt.Errorf("javbus blocked: HTTP %d (likely Cloudflare)", resp.StatusCode)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("javbus HTTP %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return "", err
	}

	// If Go's transport didn't decompress (it sometimes doesn't, even
	// when Content-Encoding: gzip is set, depending on the transport
	// config), do it ourselves by sniffing the gzip magic bytes.
	if len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b {
		gz, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("gzip decode: %w", err)
		}
		defer gz.Close()
		decoded, err := io.ReadAll(io.LimitReader(gz, 8<<20))
		if err != nil {
			return "", fmt.Errorf("gzip read: %w", err)
		}
		body = decoded
	}

	if debugJavbus {
		fmt.Fprintf(os.Stderr, "[javbus] %s -> %d bytes, %d actors\n",
			url, len(body), len(actorRe.FindAllStringSubmatch(string(body), -1)))
	}

	return string(body), nil
}

// debugJavbus is true when JAVBUS_DEBUG=1 so users can see what's
// happening at the HTTP layer without recompiling.
var debugJavbus = os.Getenv("JAVBUS_DEBUG") == "1"

// fetchMovieActors GETs the JavBus movie page and returns the actor names.
// Empty slice + nil error means "page loaded but no actors listed"; a
// non-nil error means we couldn't load the page at all (most often
// Cloudflare blocking).
func (j *javbusHTTP) fetchMovieActors(ctx context.Context, number string) ([]string, error) {
	url := fmt.Sprintf("%s/ja/%s", javbusBase, number)
	// Retry up to 3 times on transient errors (EOF, TLS handshake timeout,
	// 5xx). The first request after a cold proxy connection often fails
	// with EOF while the proxy warms up the upstream session.
	const maxAttempts = 3
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		body, err := j.get(ctx, url)
		if err == nil {
			return parseJavbusActors(body), nil
		}
		lastErr = err
		if !isTransient(err) {
			return nil, err
		}
		if debugJavbus {
			fmt.Fprintf(os.Stderr, "[javbus] retry %d/%d for %s: %v\n", attempt, maxAttempts, number, err)
		}
		// brief backoff
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
		}
	}
	return nil, lastErr
}

// isTransient returns true for errors that look like a dropped
// connection — these often succeed on retry.
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "EOF") ||
		strings.Contains(s, "connection reset") ||
		strings.Contains(s, "TLS handshake timeout") ||
		strings.Contains(s, "context deadline exceeded")
}

// actorRe matches the star-name block on JavBus movie pages. The title
// attribute holds the actor's display name.
//
// <div class="star-name"><a href="/star/xyz" title="三上悠亜">...</a></div>
var actorRe = regexp.MustCompile(`(?s)<div class="star-name">.*?<a [^>]*title="([^"]+)"`)

// parseJavbusActors extracts actor display names from the raw HTML. We
// use a regex instead of htmlquery to keep this code self-contained and
// avoid pulling the SDK's parser into the resolver.
func parseJavbusActors(htmlBody string) []string {
	matches := actorRe.FindAllStringSubmatch(htmlBody, -1)
	if len(matches) == 0 {
		// Fallback: walk the DOM and pick elements with class=star-name.
		// Useful if JavBus tweaks its markup.
		return walkForActors(htmlBody)
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(matches))
	for _, m := range matches {
		name := strings.TrimSpace(html.UnescapeString(m[1]))
		if name != "" && !seen[name] {
			seen[name] = true
			out = append(out, name)
		}
	}
	return out
}

// walkForActors is the DOM-walking fallback when the regex misses. It
// only handles the well-known markup shape.
func walkForActors(s string) []string {
	doc, err := html.Parse(strings.NewReader(s))
	if err != nil {
		return nil
	}
	var out []string
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.ElementNode {
			for _, a := range n.Attr {
				if a.Key == "class" && strings.Contains(a.Val, "star-name") {
					// find first <a title="...">
					if c := n.FirstChild; c != nil {
						for sub := c; sub != nil; sub = sub.NextSibling {
							if sub.Type == html.ElementNode && sub.Data == "a" {
								for _, aa := range sub.Attr {
									if aa.Key == "title" {
										name := strings.TrimSpace(html.UnescapeString(aa.Val))
										if name != "" {
											out = append(out, name)
										}
									}
								}
								break
							}
						}
					}
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(doc)
	return out
}
