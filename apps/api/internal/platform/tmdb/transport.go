package tmdb

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/time/rate"
)

type contextTransport struct {
	ctx  context.Context
	base http.RoundTripper
}

func (t contextTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.ctx == nil {
		return t.base.RoundTrip(req)
	}
	if err := t.ctx.Err(); err != nil {
		return nil, err
	}

	// Preserve the SDK/HTTP deadline as well as the caller's cancellation.
	var ctx context.Context
	var cancel context.CancelFunc
	if deadline, ok := t.ctx.Deadline(); ok {
		ctx, cancel = context.WithDeadline(req.Context(), deadline)
	} else {
		ctx, cancel = context.WithCancel(req.Context())
	}
	stop := context.AfterFunc(t.ctx, func() {
		// The merged deadline must expire itself to preserve DeadlineExceeded.
		if t.ctx.Err() == context.Canceled {
			cancel()
		}
	})
	cleanup := func() {
		stop()
		cancel()
	}
	resp, err := t.base.RoundTrip(req.Clone(ctx))
	if err != nil {
		cleanup()
		return nil, err
	}
	// The response must remain readable until the SDK closes it.
	resp.Body = &contextBody{ReadCloser: resp.Body, cleanup: cleanup}
	return resp, nil
}

type contextBody struct {
	io.ReadCloser
	cleanup func()
}

func (b *contextBody) Close() error {
	defer b.cleanup()
	return b.ReadCloser.Close()
}

type rateLimitedTransport struct {
	base    http.RoundTripper
	limiter *rate.Limiter
}

func (t rateLimitedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if err := t.limiter.Wait(req.Context()); err != nil {
		return nil, fmt.Errorf("rate limit wait: %w", err)
	}
	return t.base.RoundTrip(req)
}

type retryTransport struct {
	base http.RoundTripper
}

func (t retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		if err := req.Context().Err(); err != nil {
			return nil, err
		}
		resp, err := t.base.RoundTrip(req.Clone(req.Context()))
		if req.Method != http.MethodGet || (req.Body != nil && req.Body != http.NoBody) ||
			attempt == 2 || !retryable(resp, err) {
			return resp, err
		}

		delay := retryDelay(resp, attempt)
		if deadline, ok := req.Context().Deadline(); ok && time.Until(deadline) <= delay {
			return resp, err
		}
		if resp != nil {
			resp.Body.Close()
		}
		timer := time.NewTimer(delay)
		select {
		case <-req.Context().Done():
			timer.Stop()
			return nil, req.Context().Err()
		case <-timer.C:
		}
	}
}

func retryable(resp *http.Response, err error) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return false
		}
		var networkErr net.Error
		return (errors.As(err, &networkErr) && networkErr.Timeout()) ||
			errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) ||
			errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.EPIPE)
	}
	switch resp.StatusCode {
	case http.StatusTooManyRequests, http.StatusInternalServerError,
		http.StatusBadGateway, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func retryDelay(resp *http.Response, attempt int) time.Duration {
	delay := 250 * time.Millisecond << attempt
	delay += time.Duration(rand.Int64N(int64(delay/2) + 1))
	if resp == nil || (resp.StatusCode != http.StatusTooManyRequests && resp.StatusCode != http.StatusServiceUnavailable) {
		return delay
	}

	value := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if seconds, err := strconv.ParseInt(value, 10, 64); err == nil && seconds >= 0 {
		const maxDelay = time.Duration(1<<63 - 1)
		if seconds > int64(maxDelay/time.Second) {
			return maxDelay
		}
		return max(delay, time.Duration(seconds)*time.Second)
	}
	if deadline, err := http.ParseTime(value); err == nil {
		return max(delay, time.Until(deadline))
	}
	return delay
}
