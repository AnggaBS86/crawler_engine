package httpapi

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"example.com/crawler_engine/internal/crawler"
	"example.com/crawler_engine/internal/paths"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type Config struct {
	Port           string
	OutputDir      string
	CrawlTimeout   time.Duration
	NetworkIdle    time.Duration
	NetworkIdleMax time.Duration
	CrawlWaitAfter time.Duration
}

func ConfigFromEnv() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	outputDir := os.Getenv("OUTPUT_DIR")
	if outputDir == "" {
		outputDir = "/media/user/New Volume/go/crawler_engine/output"
	}

	timeoutSeconds := 45
	if v := os.Getenv("CRAWL_TIMEOUT_SECONDS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSeconds = n
		}
	}

	waitMS := 2000
	if v := os.Getenv("CRAWL_WAIT_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			waitMS = n
		}
	}

	networkIdleMS := 500
	if v := os.Getenv("CRAWL_NETWORK_IDLE_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			networkIdleMS = n
		}
	}

	networkIdleMaxMS := 6000
	if v := os.Getenv("CRAWL_NETWORK_IDLE_MAX_MS"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n >= 0 {
			networkIdleMaxMS = n
		}
	}

	return Config{
		Port:           port,
		OutputDir:      outputDir,
		CrawlTimeout:   time.Duration(timeoutSeconds) * time.Second,
		NetworkIdle:    time.Duration(networkIdleMS) * time.Millisecond,
		NetworkIdleMax: time.Duration(networkIdleMaxMS) * time.Millisecond,
		CrawlWaitAfter: time.Duration(waitMS) * time.Millisecond,
	}
}

func NewServer(cfg Config) *echo.Echo {
	e := echo.New()

	e.HideBanner = true
	e.Use(middleware.Recover())
	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())

	e.GET("/health", func(c echo.Context) error { return c.NoContent(http.StatusOK) })
	e.POST("/crawl", handleCrawl(cfg))

	return e
}

type crawlRequest struct {
	URL           string `json:"url"`
	Filename      string `json:"filename"`
	NetworkIdleMS *int   `json:"network_idle_ms,omitempty"`
	NetworkMaxMS  *int   `json:"network_idle_max_ms,omitempty"`
	WaitMS        *int   `json:"wait_ms,omitempty"`
}

type crawlResponse struct {
	URL      string `json:"url"`
	Filepath string `json:"filepath"`
	Bytes    int    `json:"bytes"`
}

func handleCrawl(cfg Config) echo.HandlerFunc {
	return func(c echo.Context) error {
		var req crawlRequest
		if err := c.Bind(&req); err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid json body")
		}
		if req.URL == "" {
			return echo.NewHTTPError(http.StatusBadRequest, "`url` is required")
		}

		waitAfter := cfg.CrawlWaitAfter
		if req.WaitMS != nil && *req.WaitMS >= 0 {
			waitAfter = time.Duration(*req.WaitMS) * time.Millisecond
		}

		networkIdle := cfg.NetworkIdle
		if req.NetworkIdleMS != nil && *req.NetworkIdleMS >= 0 {
			networkIdle = time.Duration(*req.NetworkIdleMS) * time.Millisecond
		}

		networkIdleMax := cfg.NetworkIdleMax
		if req.NetworkMaxMS != nil && *req.NetworkMaxMS >= 0 {
			networkIdleMax = time.Duration(*req.NetworkMaxMS) * time.Millisecond
		}

		ctx, cancel := context.WithTimeout(c.Request().Context(), cfg.CrawlTimeout)
		defer cancel()

		html, err := crawler.RenderedHTML(ctx, crawler.Options{
			URL:            req.URL,
			NetworkIdle:    networkIdle,
			NetworkIdleMax: networkIdleMax,
			Wait:           waitAfter,
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusBadGateway, err.Error())
		}

		filename := req.Filename
		if filename == "" {
			filename = paths.DefaultSnapshotFilename(req.URL)
		}

		fullpath, err := paths.SafeJoin(cfg.OutputDir, filename)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid filename")
		}

		if err := os.MkdirAll(paths.Dir(fullpath), 0o755); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to create output dir")
		}
		if err := os.WriteFile(fullpath, []byte(html), 0o644); err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "failed to write output file")
		}

		return c.JSON(http.StatusOK, crawlResponse{
			URL:      req.URL,
			Filepath: fullpath,
			Bytes:    len(html),
		})
	}
}
