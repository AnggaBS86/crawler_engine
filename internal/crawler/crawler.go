package crawler

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync/atomic"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Options struct {
	URL            string
	NetworkIdle    time.Duration
	NetworkIdleMax time.Duration
	Wait           time.Duration
}

func RenderedHTML(ctx context.Context, opt Options) (string, error) {
	if opt.URL == "" {
		return "", errors.New("url is required")
	}
	if opt.NetworkIdle <= 0 {
		opt.NetworkIdle = 500 * time.Millisecond
	}
	if opt.NetworkIdleMax <= 0 {
		opt.NetworkIdleMax = 6 * time.Second
	}

	shared, err := getSharedBrowser()
	if err != nil {
		return "", err
	}

	tabCtx, cancelTab := chromedp.NewContext(shared.ctx)
	defer cancelTab()

	runCtx, cancelRun := context.WithCancel(tabCtx)
	defer cancelRun()
	go func() {
		select {
		case <-ctx.Done():
			cancelRun()
		case <-runCtx.Done():
		}
	}()

	var html string

	var inFlight atomic.Int64
	var lastActivity atomic.Int64
	lastActivity.Store(time.Now().UnixNano())

	chromedp.ListenTarget(tabCtx, func(ev any) {
		now := time.Now().UnixNano()
		switch ev.(type) {
		case *network.EventRequestWillBeSent:
			inFlight.Add(1)
			lastActivity.Store(now)
		case *network.EventLoadingFinished, *network.EventLoadingFailed:
			if v := inFlight.Add(-1); v < 0 {
				inFlight.Store(0)
			}
			lastActivity.Store(now)
		}
	})

	actions := []chromedp.Action{
		network.Enable(),
		chromedp.Navigate(opt.URL),
		chromedp.WaitReady("body", chromedp.ByQuery),
		waitForNetworkIdle(&inFlight, &lastActivity, opt.NetworkIdle, opt.NetworkIdleMax),
	}
	if opt.Wait > 0 {
		actions = append(actions, chromedp.Sleep(opt.Wait))
	}
	actions = append(actions, chromedp.OuterHTML("html", &html, chromedp.ByQuery))

	if err := chromedp.Run(runCtx, actions...); err != nil {
		if out := strings.TrimSpace(shared.out.String()); out != "" {
			return "", fmt.Errorf("chromedp run: %w\nchrome output (tail):\n%s", err, out)
		}
		return "", fmt.Errorf("chromedp run: %w", err)
	}
	if html == "" {
		return "", errors.New("empty html snapshot")
	}

	return html, nil
}

func waitForNetworkIdle(inFlight *atomic.Int64, lastActivity *atomic.Int64, idleFor, maxWait time.Duration) chromedp.ActionFunc {
	return func(ctx context.Context) error {
		if idleFor <= 0 || maxWait <= 0 {
			return nil
		}

		start := time.Now()
		ticker := time.NewTicker(100 * time.Millisecond)
		defer ticker.Stop()

		for {
			if time.Since(start) >= maxWait {
				return nil
			}

			if inFlight.Load() <= 0 {
				last := time.Unix(0, lastActivity.Load())
				if time.Since(last) >= idleFor {
					return nil
				}
			}

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-ticker.C:
			}
		}
	}
}
