package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// runDiag probes the network path to JavBus and prints enough info to
// tell whether the problem is Cloudflare, a bad proxy, or something else.
func runDiag() {
	fmt.Println("=== network diagnostics ===")

	// Show what proxy env vars are set, if any.
	for _, v := range []string{"HTTPS_PROXY", "HTTP_PROXY", "ALL_PROXY", "https_proxy", "http_proxy"} {
		if p := os.Getenv(v); p != "" {
			fmt.Printf("[env] %s = %s\n", v, p)
		}
	}
	fmt.Println()

	client := diagClient()

	fmt.Println("[1/3] egress IP check (api.ipify.org)")
	probe(client, "https://api.ipify.org?format=text", true)
	fmt.Println()

	fmt.Println("[2/3] JavBus root probe (https://www.javbus.com/)")
	probe(client, "https://www.javbus.com/", false)
	fmt.Println()

	fmt.Println("[3/3] JavBus movie probe (https://www.javbus.com/ja/SNIS-326)")
	probe(client, "https://www.javbus.com/ja/SNIS-326", false)
	fmt.Println()

	fmt.Println("=== diagnosis ===")
	fmt.Println("if you see HTTP 200 from all three, the network path is healthy.")
	fmt.Println("if you see EOF / TLS handshake timeout, Go can't reach the internet.")
	fmt.Println("if you see HTTP 403/503 with 'Just a moment...' body, Cloudflare is blocking.")
}

func diagClient() *http.Client {
	proxy := ""
	for _, v := range []string{"HTTPS_PROXY", "HTTP_PROXY", "https_proxy", "http_proxy", "ALL_PROXY", "all_proxy"} {
		if p := os.Getenv(v); p != "" {
			proxy = p
			break
		}
	}
	transport := &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 15 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
	}
	if proxy != "" {
		if !strings.Contains(proxy, "://") {
			proxy = "http://" + proxy
		}
		transport.Proxy = http.ProxyFromEnvironment
		fmt.Printf("[diag] env proxy set: %s\n", proxy)
	} else {
		// No env var set — leave Proxy nil, so Go does a direct dial.
		// If the system is using Clash TUN mode, the dial will be
		// intercepted at the network layer.
		fmt.Println("[diag] no env proxy; using direct connection (relies on TUN/system proxy)")
	}
	return &http.Client{
		Timeout:   20 * time.Second,
		Transport: transport,
	}
}

func probe(client *http.Client, urlStr string, plain bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	req, _ := http.NewRequestWithContext(ctx, "GET", urlStr, nil)
	if plain {
		req.Header.Set("User-Agent", "metatube-sort-diag/1.0")
	} else {
		req.Header.Set("User-Agent",
			"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "+
				"(KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
		req.Header.Set("Accept", "text/html")
		req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9")
	}

	start := time.Now()
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("        ERROR: %v\n", err)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 400))
	elapsed := time.Since(start)

	fmt.Printf("        HTTP %d  %s  (%d bytes in body)\n",
		resp.StatusCode, elapsed.Round(time.Millisecond), len(body))
	if len(body) > 0 {
		fmt.Printf("        first 200 bytes: %q\n", truncateDiag(string(body), 200))
	}

	if resp.StatusCode == 200 && (strings.Contains(string(body), "Just a moment") ||
		strings.Contains(string(body), "cf-challenge")) {
		fmt.Println("        >>> Cloudflare JS challenge page — needs cf_clearance cookie <<<")
	}
	if resp.StatusCode == 403 || resp.StatusCode == 503 {
		fmt.Println("        >>> Cloudflare blocked this IP. Use -proxy or system proxy. <<<")
	}
	if srv := resp.Header.Get("Server"); srv != "" {
		fmt.Printf("        Server: %s\n", srv)
	}
}

func truncateDiag(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}
