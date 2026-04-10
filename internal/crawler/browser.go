package crawler

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/chromedp/chromedp"
)

type sharedBrowser struct {
	ctx         context.Context
	cancelCtx   context.CancelFunc
	cancelAlloc context.CancelFunc
	out         *tailBuffer
	userDataDir string
}

var (
	sharedOnce sync.Once
	sharedInst *sharedBrowser
	sharedErr  error
)

func getSharedBrowser() (*sharedBrowser, error) {
	sharedOnce.Do(func() {
		userDataDir, err := os.MkdirTemp("", "crawler_engine-chrome-*")
		if err != nil {
			sharedErr = fmt.Errorf("mktemp chrome profile: %w", err)
			return
		}

		chromeOut := newTailBuffer(64 << 10)

		allocOpts := append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-setuid-sandbox", true),
			chromedp.UserDataDir(userDataDir),
			chromedp.CombinedOutput(chromeOut),
			chromedp.Flag("disable-features", "site-per-process,Translate,BlinkGenPropertyTrees,Crashpad"),
			chromedp.Flag("disable-dev-shm-usage", true),
			chromedp.Flag("disable-crash-reporter", true),
			chromedp.Flag("disable-breakpad", true),
			chromedp.Flag("disable-crashpad", true),
			chromedp.Flag("enable-logging", "stderr"),
			chromedp.Flag("v", "1"),
		)

		if p := strings.TrimSpace(os.Getenv("CHROME_BIN")); p != "" {
			resolved, lookErr := exec.LookPath(p)
			if lookErr != nil {
				sharedErr = fmt.Errorf("CHROME_BIN not found in PATH: %s", p)
				_ = os.RemoveAll(userDataDir)
				return
			}
			allocOpts = append(allocOpts, chromedp.ExecPath(resolved))
		}

		allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), allocOpts...)
		browserCtx, cancelCtx := chromedp.NewContext(allocCtx, chromedp.WithLogf(func(string, ...any) {}))

		// Trigger browser start now (so the first request is faster and errors
		// surface early).
		if err := chromedp.Run(browserCtx); err != nil {
			cancelCtx()
			cancelAlloc()
			_ = os.RemoveAll(userDataDir)
			sharedErr = fmt.Errorf("start chrome: %w", err)
			return
		}

		sharedInst = &sharedBrowser{
			ctx:         browserCtx,
			cancelCtx:   cancelCtx,
			cancelAlloc: cancelAlloc,
			out:         chromeOut,
			userDataDir: userDataDir,
		}
	})

	if sharedErr != nil {
		return nil, sharedErr
	}
	return sharedInst, nil
}

// Warmup starts the shared browser early so the first request is faster.
func Warmup() error {
	_, err := getSharedBrowser()
	return err
}
