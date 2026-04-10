package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"example.com/crawler_engine/internal/crawler"
)

func main() {
	var targetURL string
	var out string
	var timeoutSeconds int
	var networkIdleMS int
	var networkIdleMaxMS int
	var waitMS int
	flag.StringVar(&targetURL, "url", "", "URL to crawl (required)")
	flag.StringVar(&out, "out", "", "Output HTML file path (required)")
	flag.IntVar(&timeoutSeconds, "timeout", 45, "Crawl timeout in seconds")
	flag.IntVar(&networkIdleMS, "network-idle-ms", 500, "Wait until no network activity for this long (ms)")
	flag.IntVar(&networkIdleMaxMS, "network-idle-max-ms", 6000, "Max time to wait for network-idle before continuing (ms)")
	flag.IntVar(&waitMS, "wait-ms", 2000, "Extra wait after DOM ready (ms)")
	flag.Parse()

	if targetURL == "" || out == "" {
		flag.Usage()
		os.Exit(2)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeoutSeconds)*time.Second)
	defer cancel()

	html, err := crawler.RenderedHTML(ctx, crawler.Options{
		URL:            targetURL,
		NetworkIdle:    time.Duration(networkIdleMS) * time.Millisecond,
		NetworkIdleMax: time.Duration(networkIdleMaxMS) * time.Millisecond,
		Wait:           time.Duration(waitMS) * time.Millisecond,
	})
	if err != nil {
		log.Println("crawl failed:", err)
		os.Exit(1)
	}

	if err := os.MkdirAll(filepath.Dir(out), 0o755); err != nil {
		log.Println("mkdir failed:", err)
		os.Exit(1)
	}
	if err := os.WriteFile(out, []byte(html), 0o644); err != nil {
		log.Println("write failed:", err)
		os.Exit(1)
	}

	fmt.Println(out)
}
