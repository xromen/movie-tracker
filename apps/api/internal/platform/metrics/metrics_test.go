package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestUnmatchedRoutesShareMetricLabel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	m := NewHTTPMetrics()
	router := gin.New()
	router.Use(m.Middleware())
	for _, path := range []string{"/missing/one", "/missing/two"} {
		router.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, path, nil))
	}
	output := m.Render()
	if !strings.Contains(output, `movie_tracker_api_http_requests_total{method="GET",route="<unmatched>",status="404"} 2`) {
		t.Fatalf("unmatched requests were not grouped: %s", output)
	}
}
