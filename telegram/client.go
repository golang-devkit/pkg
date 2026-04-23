package telegram

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"time"
)

// Client is the main Telegram Gateway client exposed by this package.
// Client là Telegram Gateway client chính được package này expose.
type Client = GatewayClient

// New creates a Telegram Gateway API client.
// New tạo Telegram Gateway API client.
func New(token string, opts ...Option) (*Client, error) {
	return NewGateway(token, opts...)
}

type rateLimiter struct {
	enabled     bool
	minInterval time.Duration

	mu   sync.Mutex
	next time.Time
}

func (r *rateLimiter) Wait(ctx context.Context) error {
	if r == nil || !r.enabled || r.minInterval <= 0 {
		return nil
	}

	r.mu.Lock()
	now := time.Now()
	waitFor := time.Duration(0)
	if now.Before(r.next) {
		waitFor = r.next.Sub(now)
	}
	r.next = now.Add(waitFor).Add(r.minInterval)
	r.mu.Unlock()

	if waitFor <= 0 {
		return nil
	}
	return sleepContext(ctx, waitFor)
}

func buildHTTPClient(cfg Config) (*http.Client, error) {
	if cfg.HTTPClient != nil {
		cloned := *cfg.HTTPClient
		if cfg.Timeout > 0 {
			cloned.Timeout = cfg.Timeout
		}
		if cfg.ProxyURL != nil {
			transport, err := cloneTransport(cloned.Transport)
			if err != nil {
				return nil, err
			}
			transport.Proxy = http.ProxyURL(cfg.ProxyURL)
			cloned.Transport = transport
		}
		return &cloned, nil
	}

	transport, err := cloneTransport(nil)
	if err != nil {
		return nil, err
	}
	if cfg.ProxyURL != nil {
		transport.Proxy = http.ProxyURL(cfg.ProxyURL)
	}

	return &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}, nil
}

func cloneTransport(rt http.RoundTripper) (*http.Transport, error) {
	if rt == nil {
		if base, ok := http.DefaultTransport.(*http.Transport); ok {
			return base.Clone(), nil
		}
		return &http.Transport{}, nil
	}

	transport, ok := rt.(*http.Transport)
	if !ok {
		return nil, errors.New("telegram gateway: custom transport must be *http.Transport when using proxy")
	}
	return transport.Clone(), nil
}

func sleepContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
