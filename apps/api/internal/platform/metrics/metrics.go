package metrics

import (
	"fmt"
	"net/http"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var durationBuckets = []float64{0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10}

type HTTPMetrics struct {
	mu        sync.RWMutex
	startedAt time.Time
	requests  map[requestKey]*requestMetrics
}

type requestKey struct {
	Method string
	Route  string
	Status int
}

type requestMetrics struct {
	Count   uint64
	Sum     float64
	Buckets []uint64
}

func NewHTTPMetrics() *HTTPMetrics {
	return &HTTPMetrics{
		startedAt: time.Now(),
		requests:  make(map[requestKey]*requestMetrics),
	}
}

func (m *HTTPMetrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/metrics" {
			c.Next()
			return
		}

		start := time.Now()
		c.Next()

		route := c.FullPath()
		if route == "" {
			route = c.Request.URL.Path
		}

		m.Observe(c.Request.Method, route, c.Writer.Status(), time.Since(start).Seconds())
	}
}

func (m *HTTPMetrics) Observe(method string, route string, status int, durationSeconds float64) {
	key := requestKey{
		Method: method,
		Route:  route,
		Status: status,
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	item, exists := m.requests[key]
	if !exists {
		item = &requestMetrics{
			Buckets: make([]uint64, len(durationBuckets)),
		}
		m.requests[key] = item
	}

	item.Count++
	item.Sum += durationSeconds

	for index, bucket := range durationBuckets {
		if durationSeconds <= bucket {
			item.Buckets[index]++
		}
	}
}

func (m *HTTPMetrics) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Data(http.StatusOK, "text/plain; version=0.0.4; charset=utf-8", []byte(m.Render()))
	}
}

func (m *HTTPMetrics) Render() string {
	m.mu.RLock()
	keys := make([]requestKey, 0, len(m.requests))
	for key := range m.requests {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool {
		if keys[i].Route != keys[j].Route {
			return keys[i].Route < keys[j].Route
		}
		if keys[i].Method != keys[j].Method {
			return keys[i].Method < keys[j].Method
		}
		return keys[i].Status < keys[j].Status
	})

	snapshots := make([]struct {
		Key requestKey
		Val requestMetrics
	}, 0, len(keys))
	for _, key := range keys {
		val := m.requests[key]
		buckets := make([]uint64, len(val.Buckets))
		copy(buckets, val.Buckets)
		snapshots = append(snapshots, struct {
			Key requestKey
			Val requestMetrics
		}{
			Key: key,
			Val: requestMetrics{
				Count:   val.Count,
				Sum:     val.Sum,
				Buckets: buckets,
			},
		})
	}
	m.mu.RUnlock()

	var mem runtime.MemStats
	runtime.ReadMemStats(&mem)

	var b strings.Builder
	b.WriteString("# HELP movie_tracker_api_http_requests_total Total API HTTP requests.\n")
	b.WriteString("# TYPE movie_tracker_api_http_requests_total counter\n")
	for _, snapshot := range snapshots {
		labels := httpLabels(snapshot.Key)
		fmt.Fprintf(&b, "movie_tracker_api_http_requests_total{%s} %d\n", labels, snapshot.Val.Count)
	}

	b.WriteString("# HELP movie_tracker_api_http_request_duration_seconds API HTTP request duration in seconds.\n")
	b.WriteString("# TYPE movie_tracker_api_http_request_duration_seconds histogram\n")
	for _, snapshot := range snapshots {
		labels := httpLabels(snapshot.Key)
		for index, bucket := range durationBuckets {
			fmt.Fprintf(&b, "movie_tracker_api_http_request_duration_seconds_bucket{%s,le=%q} %d\n", labels, formatFloat(bucket), snapshot.Val.Buckets[index])
		}
		fmt.Fprintf(&b, "movie_tracker_api_http_request_duration_seconds_bucket{%s,le=\"+Inf\"} %d\n", labels, snapshot.Val.Count)
		fmt.Fprintf(&b, "movie_tracker_api_http_request_duration_seconds_sum{%s} %s\n", labels, formatFloat(snapshot.Val.Sum))
		fmt.Fprintf(&b, "movie_tracker_api_http_request_duration_seconds_count{%s} %d\n", labels, snapshot.Val.Count)
	}

	b.WriteString("# HELP movie_tracker_api_uptime_seconds API process uptime in seconds.\n")
	b.WriteString("# TYPE movie_tracker_api_uptime_seconds gauge\n")
	fmt.Fprintf(&b, "movie_tracker_api_uptime_seconds %s\n", formatFloat(time.Since(m.startedAt).Seconds()))
	b.WriteString("# HELP movie_tracker_api_goroutines Current number of goroutines.\n")
	b.WriteString("# TYPE movie_tracker_api_goroutines gauge\n")
	fmt.Fprintf(&b, "movie_tracker_api_goroutines %d\n", runtime.NumGoroutine())
	b.WriteString("# HELP movie_tracker_api_memory_alloc_bytes Current bytes allocated by the Go runtime.\n")
	b.WriteString("# TYPE movie_tracker_api_memory_alloc_bytes gauge\n")
	fmt.Fprintf(&b, "movie_tracker_api_memory_alloc_bytes %d\n", mem.Alloc)

	return b.String()
}

func httpLabels(key requestKey) string {
	return fmt.Sprintf(
		"method=%q,route=%q,status=%q",
		escapeLabel(key.Method),
		escapeLabel(key.Route),
		escapeLabel(strconv.Itoa(key.Status)),
	)
}

func escapeLabel(value string) string {
	value = strings.ReplaceAll(value, "\\", "\\\\")
	value = strings.ReplaceAll(value, "\n", "\\n")
	return strings.ReplaceAll(value, "\"", "\\\"")
}

func formatFloat(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}
