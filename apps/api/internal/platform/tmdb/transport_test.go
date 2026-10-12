package tmdb

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"testing/synctest"
	"time"

	"golang.org/x/time/rate"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type trackedBody struct {
	io.Reader
	closed bool
}

func (b *trackedBody) Close() error {
	b.closed = true
	return nil
}

type timeoutError struct{}

func (timeoutError) Error() string   { return "network timeout" }
func (timeoutError) Timeout() bool   { return true }
func (timeoutError) Temporary() bool { return true }

func TestRetryTransport(t *testing.T) {
	for _, status := range []int{200, 400, 401, 403, 404, 429, 500, 502, 503, 504} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				var bodies []*trackedBody
				transport := retryTransport{base: transportFunc(func(req *http.Request) (*http.Response, error) {
					body := &trackedBody{Reader: strings.NewReader("response")}
					bodies = append(bodies, body)
					return &http.Response{StatusCode: status, Body: body, Header: make(http.Header)}, nil
				})}
				req, _ := http.NewRequest(http.MethodGet, "https://tmdb.example/movie", nil)
				start := time.Now()
				resp, err := transport.RoundTrip(req)
				if err != nil || resp.StatusCode != status {
					t.Fatalf("RoundTrip() = %v, %v", resp, err)
				}
				wantAttempts := 1
				if status == 429 || status >= 500 {
					wantAttempts = 3
				}
				if len(bodies) != wantAttempts {
					t.Fatalf("attempts = %d, want %d", len(bodies), wantAttempts)
				}
				for _, body := range bodies[:len(bodies)-1] {
					if !body.closed {
						t.Fatal("discarded response body was not closed")
					}
				}
				if bodies[len(bodies)-1].closed {
					t.Fatal("final response body was closed before decoding")
				}
				resp.Body.Close()
				if elapsed := time.Since(start); wantAttempts == 3 && (elapsed < 750*time.Millisecond || elapsed > 1125*time.Millisecond) {
					t.Fatalf("backoff with jitter took %s", elapsed)
				}
			})
		})
	}
}

func TestRetryNetworkErrors(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want int
	}{
		{"timeout", timeoutError{}, 2},
		{"EOF", io.EOF, 2},
		{"reset", syscall.ECONNRESET, 2},
		{"canceled", context.Canceled, 1},
		{"deadline", context.DeadlineExceeded, 1},
		{"permanent", errors.New("invalid certificate"), 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				calls := 0
				transport := retryTransport{base: transportFunc(func(req *http.Request) (*http.Response, error) {
					calls++
					if calls == 1 {
						return nil, tc.err
					}
					return &http.Response{StatusCode: 200, Body: http.NoBody}, nil
				})}
				req, _ := http.NewRequest(http.MethodGet, "https://tmdb.example/movie", nil)
				resp, err := transport.RoundTrip(req)
				if calls != tc.want || (tc.want == 1 && !errors.Is(err, tc.err)) || (tc.want == 2 && err != nil) {
					t.Fatalf("calls = %d, err = %v", calls, err)
				}
				if resp != nil {
					resp.Body.Close()
				}
			})
		})
	}
}

func TestRetryAfter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		for _, tc := range []struct {
			value string
			want  time.Duration
		}{
			{"2", 2 * time.Second},
			{time.Now().Add(3 * time.Second).UTC().Format(http.TimeFormat), 3 * time.Second},
			{"9223372036854775807", time.Duration(1<<63 - 1)},
		} {
			for _, status := range []int{429, 503} {
				resp := &http.Response{StatusCode: status, Header: http.Header{"Retry-After": {tc.value}}}
				if got := retryDelay(resp, 0); got != tc.want {
					t.Errorf("retryDelay(%d, %q) = %s, want %s", status, tc.value, got, tc.want)
				}
			}
		}
		for _, value := range []string{"", "invalid", "-1", "0", time.Now().Add(-time.Hour).UTC().Format(http.TimeFormat)} {
			resp := &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {value}}}
			if got := retryDelay(resp, 0); got < 250*time.Millisecond || got > 375*time.Millisecond {
				t.Errorf("fallback delay for %q = %s", value, got)
			}
		}
	})
}

func TestRetryAfterDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		calls := 0
		transport := retryTransport{base: transportFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 429, Header: http.Header{"Retry-After": {"2"}}, Body: http.NoBody}, nil
		})}
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://tmdb.example/movie", nil)
		start := time.Now()
		resp, err := transport.RoundTrip(req)
		if err != nil || resp.StatusCode != 429 || calls != 1 || time.Since(start) != 0 {
			t.Fatalf("RoundTrip() = %v, %v; calls = %d", resp, err, calls)
		}
		resp.Body.Close()
	})
}

func TestRetryCancellation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		calls := 0
		body := &trackedBody{Reader: strings.NewReader("unavailable")}
		transport := retryTransport{base: transportFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 503, Body: body}, nil
		})}
		go func() {
			time.Sleep(10 * time.Millisecond)
			cancel()
		}()
		req, _ := http.NewRequestWithContext(ctx, http.MethodGet, "https://tmdb.example/movie", nil)
		_, err := transport.RoundTrip(req)
		if !errors.Is(err, context.Canceled) || calls != 1 || !body.closed {
			t.Fatalf("err = %v; calls = %d; body closed = %v", err, calls, body.closed)
		}
	})
}

func TestRetryDoesNotReplayBodyOrWrites(t *testing.T) {
	for _, method := range []string{http.MethodPost, http.MethodGet} {
		calls := 0
		transport := retryTransport{base: transportFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			return &http.Response{StatusCode: 503, Body: http.NoBody}, nil
		})}
		req, _ := http.NewRequest(method, "https://tmdb.example/movie", strings.NewReader("body"))
		resp, err := transport.RoundTrip(req)
		if err != nil || calls != 1 {
			t.Fatalf("%s: calls = %d; err = %v", method, calls, err)
		}
		resp.Body.Close()
	}
}

func TestClientSharesLimiter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		api, err := NewClient(Config{BearerToken: "test", Timeout: 10 * time.Second, RPM: 240, Burst: 1})
		if err != nil {
			t.Fatal(err)
		}
		c := api.(*client)
		var calls atomic.Int32
		c.transport = transportFunc(func(req *http.Request) (*http.Response, error) {
			calls.Add(1)
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader(`{"results":[],"total_pages":0,"total_results":0}`))}, nil
		})
		start := time.Now()
		if _, err := c.GetMoviePopular(context.Background(), 1); err != nil {
			t.Fatal(err)
		}
		results := make(chan error, 2)
		for range 2 {
			go func() {
				_, err := c.GetMoviePopular(context.Background(), 1)
				results <- err
			}()
		}
		for range 2 {
			if err := <-results; err != nil {
				t.Fatal(err)
			}
		}
		if calls.Load() != 3 || time.Since(start) != 500*time.Millisecond {
			t.Fatalf("calls = %d; elapsed = %s, want 500ms", calls.Load(), time.Since(start))
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if _, err := c.GetMoviePopular(ctx, 1); err == nil || calls.Load() != 3 {
			t.Fatalf("limiter exceeded caller deadline: calls = %d; err = %v", calls.Load(), err)
		}
	})
}

func TestClientTimeoutIncludesRetries(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		api, err := NewClient(Config{BearerToken: "test", Timeout: time.Second, RPM: 240, Burst: 40})
		if err != nil {
			t.Fatal(err)
		}
		c := api.(*client)
		calls := 0
		c.transport = transportFunc(func(req *http.Request) (*http.Response, error) {
			calls++
			if calls == 1 {
				time.Sleep(400 * time.Millisecond)
				return &http.Response{StatusCode: 503, Body: http.NoBody}, nil
			}
			<-req.Context().Done()
			return nil, req.Context().Err()
		})
		start := time.Now()
		_, err = c.GetMoviePopular(context.Background(), 1)
		if !errors.Is(err, context.DeadlineExceeded) || calls != 2 || time.Since(start) != time.Second {
			t.Fatalf("calls = %d; elapsed = %s; err = %v", calls, time.Since(start), err)
		}
	})
}

func TestRetriesConsumeLimiterTokens(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0
		transport := retryTransport{base: rateLimitedTransport{
			limiter: rate.NewLimiter(1, 1),
			base: transportFunc(func(req *http.Request) (*http.Response, error) {
				calls++
				status := 503
				if calls == 3 {
					status = 200
				}
				return &http.Response{StatusCode: status, Body: http.NoBody}, nil
			}),
		}}
		req, _ := http.NewRequest(http.MethodGet, "https://tmdb.example/movie", nil)
		start := time.Now()
		resp, err := transport.RoundTrip(req)
		elapsed := time.Since(start)
		if err != nil || resp.StatusCode != 200 || calls != 3 || elapsed < 2*time.Second-time.Microsecond || elapsed > 2*time.Second+time.Microsecond {
			t.Fatalf("calls = %d; elapsed = %s; err = %v", calls, time.Since(start), err)
		}
		resp.Body.Close()
	})
}

func TestContextTransport(t *testing.T) {
	for _, source := range []string{"caller", "http"} {
		t.Run(source, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				callerCtx, cancelCaller := context.WithCancel(context.Background())
				defer cancelCaller()
				httpCtx, cancelHTTP := context.WithCancel(context.Background())
				defer cancelHTTP()
				var mergedCtx context.Context
				transport := contextTransport{ctx: callerCtx, base: transportFunc(func(req *http.Request) (*http.Response, error) {
					mergedCtx = req.Context()
					return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("readable"))}, nil
				})}
				req, _ := http.NewRequestWithContext(httpCtx, http.MethodGet, "https://tmdb.example/movie", nil)
				resp, err := transport.RoundTrip(req)
				if err != nil || mergedCtx.Err() != nil {
					t.Fatalf("response context canceled before reading: %v", err)
				}
				if body, err := io.ReadAll(resp.Body); err != nil || string(body) != "readable" {
					t.Fatalf("body = %q; err = %v", body, err)
				}
				if source == "caller" {
					cancelCaller()
				} else {
					cancelHTTP()
				}
				synctest.Wait()
				if !errors.Is(mergedCtx.Err(), context.Canceled) {
					t.Fatal("cancellation did not reach transport")
				}
				resp.Body.Close()
			})
		})
	}
}

func TestContextTransportPreservesDeadline(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		for _, callerTimeout := range []time.Duration{time.Millisecond, time.Second} {
			callerCtx, cancelCaller := context.WithTimeout(context.Background(), callerTimeout)
			httpCtx, cancelHTTP := context.WithTimeout(context.Background(), 10*time.Millisecond)
			transport := contextTransport{ctx: callerCtx, base: transportFunc(func(req *http.Request) (*http.Response, error) {
				wantDeadline := time.Now().Add(min(callerTimeout, 10*time.Millisecond))
				if deadline, ok := req.Context().Deadline(); !ok || !deadline.Equal(wantDeadline) {
					t.Fatalf("deadline = %s, want %s", deadline, wantDeadline)
				}
				<-req.Context().Done()
				return nil, req.Context().Err()
			})}
			req, _ := http.NewRequestWithContext(httpCtx, http.MethodGet, "https://tmdb.example/movie", nil)
			if _, err := transport.RoundTrip(req); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("err = %v, want deadline exceeded", err)
			}
			cancelCaller()
			cancelHTTP()
		}
	})
}
