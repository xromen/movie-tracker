import assert from "node:assert/strict";
import test from "node:test";
import {webMetrics} from "./prometheus.ts";

test("backend metrics aggregate IDs and unknown paths without losing histogram observations", () => {
    for (let id = 1; id <= 10000; id += 1) {
        webMetrics.observe("GET", `/v1/movie/${id}`, 200, 0.25);
        webMetrics.observe("GET", `/unknown/path-${id}`, 404, 0.01);
        webMetrics.observe("GET", `/v1/movie/${id}/unknown-${id}`, 404, 0.01);
    }

    const routes = [
        ["/v1/movie/popular", "/v1/movie/popular"],
        ["/v1/movie/search", "/v1/movie/search"],
        ["/v1/movie/42/recommendations", "/v1/movie/:id/recommendations"],
        ["/v1/tv/search", "/v1/tv/search"],
        ["/v1/tv/42", "/v1/tv/:id"],
        ["/v1/tv/42/season/0", "/v1/tv/:id/season/:season_number"],
        ["/v1/tv/42/season/2/watched", "/v1/tv/:id/season/:season_number/watched"],
        ["/v1/tv/42/season/2/episode/3/watched", "/v1/tv/:id/season/:season_number/episode/:episode_number/watched"],
        ["/v1/company/42/movie", "/v1/company/:company_id/movie"],
        ["/v1/company/43/tv", "/v1/company/:company_id/tv"],
        ["/v1/collections/42", "/v1/collections/:id"],
        ["/v1/watch-list/status", "/v1/watch-list/status"],
        ["/v1/auth/refresh", "/v1/auth/refresh"],
        ["/health/ready", "/health/ready"],
    ];

    for (const [path] of routes) {
        webMetrics.observe("GET", path, 200, 0.5);
    }
    webMetrics.observe("POST", "/v1/movie/42", 405, 0.1);

    const lines = webMetrics.render().trim().split("\n");
    const counters = lines.filter((line) => line.startsWith("movie_tracker_web_backend_requests_total{"));
    assert.equal(counters.length, routes.length + 3);
    for (const [, template] of routes) {
        assert.ok(counters.includes(`movie_tracker_web_backend_requests_total{method="GET",route="${template}",status="200"} 1`));
    }
    assert.ok(counters.includes('movie_tracker_web_backend_requests_total{method="POST",route="/v1/movie/:id",status="405"} 1'));
    assert.ok(counters.includes('movie_tracker_web_backend_requests_total{method="GET",route="<unmatched>",status="404"} 20000'));

    const labels = 'method="GET",route="/v1/movie/:id",status="200"';
    for (const expected of [
        `movie_tracker_web_backend_requests_total{${labels}} 10000`,
        `movie_tracker_web_backend_request_duration_seconds_count{${labels}} 10000`,
        `movie_tracker_web_backend_request_duration_seconds_sum{${labels}} 2500`,
        `movie_tracker_web_backend_request_duration_seconds_bucket{${labels},le="0.1"} 0`,
        `movie_tracker_web_backend_request_duration_seconds_bucket{${labels},le="0.25"} 10000`,
        `movie_tracker_web_backend_request_duration_seconds_bucket{${labels},le="+Inf"} 10000`,
    ]) {
        assert.ok(lines.includes(expected), expected);
    }
});
