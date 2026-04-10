package main

import (
	"log"
	"os"

	"example.com/crawler_engine/internal/crawler"
	"example.com/crawler_engine/internal/httpapi"
)

func main() {
	cfg := httpapi.ConfigFromEnv()

	if err := crawler.Warmup(); err != nil {
		log.Println("chrome warmup failed:", err)
	}

	e := httpapi.NewServer(cfg)

	log.Printf("crawler_engine listening on :%s", cfg.Port)
	if err := e.Start(":" + cfg.Port); err != nil {
		if err.Error() == "http: Server closed" {
			return
		}
		log.Println("server error:", err)
		os.Exit(1)
	}
}
