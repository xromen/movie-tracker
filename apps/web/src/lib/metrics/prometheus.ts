const durationBuckets = [0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5, 10] as const;

// Keep labels bounded, including requests to unknown paths through the catch-all proxy.
const metricRoutes = [
    "/health/live",
    "/health/ready",
    "/v1/auth/login",
    "/v1/auth/logout",
    "/v1/auth/register",
    "/v1/auth/refresh",
    "/v1/movie/search",
    "/v1/movie/now-playing",
    "/v1/movie/popular",
    "/v1/movie/top-rated",
    "/v1/movie/upcoming",
    "/v1/movie/:id",
    "/v1/movie/:id/recommendations",
    "/v1/tv/search",
    "/v1/tv/airing-today",
    "/v1/tv/popular",
    "/v1/tv/top-rated",
    "/v1/tv/on-the-air",
    "/v1/tv/:id",
    "/v1/tv/:id/recommendations",
    "/v1/tv/:id/season/:season_number",
    "/v1/tv/:id/season/:season_number/watched",
    "/v1/tv/:id/season/:season_number/episode/:episode_number/watched",
    "/v1/company/:company_id",
    "/v1/company/:company_id/movie",
    "/v1/company/:company_id/tv",
    "/v1/collections/:id",
    "/v1/search/multi",
    "/v1/watch-list",
    "/v1/watch-list/status",
    "/v1/telegram/binding-url",
].map((template) => ({
    template,
    pattern: new RegExp(`^${template.replace(/:[a-z_]+/g, "[^/]+")}$`),
}));

type RequestKey = {
    method: string;
    route: string;
    status: number;
};

type RequestMetrics = {
    count: number;
    sum: number;
    buckets: number[];
};

class WebMetrics {
    private readonly startedAt = Date.now();
    private readonly requests = new Map<string, RequestMetrics>();

    observe(method: string, route: string, status: number, durationSeconds: number) {
        const metricRoute = metricRoutes.find(({pattern}) => pattern.test(route))?.template ?? "<unmatched>";
        const key = createKey({method, route: metricRoute, status});
        const item = this.requests.get(key) ?? {
            count: 0,
            sum: 0,
            buckets: Array.from({length: durationBuckets.length}, () => 0),
        };

        item.count += 1;
        item.sum += durationSeconds;

        durationBuckets.forEach((bucket, index) => {
            if (durationSeconds <= bucket) {
                item.buckets[index] += 1;
            }
        });

        this.requests.set(key, item);
    }

    render() {
        const lines: string[] = [
            "# HELP movie_tracker_web_backend_requests_total Total frontend backend proxy requests.",
            "# TYPE movie_tracker_web_backend_requests_total counter",
        ];

        const snapshots = Array.from(this.requests.entries())
            .map(([key, value]) => ({key: parseKey(key), value}))
            .sort((a, b) =>
                a.key.route.localeCompare(b.key.route) ||
                a.key.method.localeCompare(b.key.method) ||
                a.key.status - b.key.status,
            );

        snapshots.forEach(({key, value}) => {
            lines.push(`movie_tracker_web_backend_requests_total{${labels(key)}} ${value.count}`);
        });

        lines.push(
            "# HELP movie_tracker_web_backend_request_duration_seconds Frontend backend proxy request duration in seconds.",
            "# TYPE movie_tracker_web_backend_request_duration_seconds histogram",
        );

        snapshots.forEach(({key, value}) => {
            const metricLabels = labels(key);

            durationBuckets.forEach((bucket, index) => {
                lines.push(
                    `movie_tracker_web_backend_request_duration_seconds_bucket{${metricLabels},le="${bucket}"} ${value.buckets[index]}`,
                );
            });

            lines.push(
                `movie_tracker_web_backend_request_duration_seconds_bucket{${metricLabels},le="+Inf"} ${value.count}`,
                `movie_tracker_web_backend_request_duration_seconds_sum{${metricLabels}} ${formatNumber(value.sum)}`,
                `movie_tracker_web_backend_request_duration_seconds_count{${metricLabels}} ${value.count}`,
            );
        });

        const memory = process.memoryUsage();

        lines.push(
            "# HELP movie_tracker_web_uptime_seconds Frontend process uptime in seconds.",
            "# TYPE movie_tracker_web_uptime_seconds gauge",
            `movie_tracker_web_uptime_seconds ${formatNumber((Date.now() - this.startedAt) / 1000)}`,
            "# HELP movie_tracker_web_memory_heap_used_bytes Frontend Node.js heap used in bytes.",
            "# TYPE movie_tracker_web_memory_heap_used_bytes gauge",
            `movie_tracker_web_memory_heap_used_bytes ${memory.heapUsed}`,
            "# HELP movie_tracker_web_memory_rss_bytes Frontend Node.js RSS in bytes.",
            "# TYPE movie_tracker_web_memory_rss_bytes gauge",
            `movie_tracker_web_memory_rss_bytes ${memory.rss}`,
        );

        return `${lines.join("\n")}\n`;
    }
}

const globalMetrics = globalThis as typeof globalThis & {
    __movieTrackerWebMetrics?: WebMetrics;
};

export const webMetrics = globalMetrics.__movieTrackerWebMetrics ?? new WebMetrics();

globalMetrics.__movieTrackerWebMetrics = webMetrics;

const createKey = ({method, route, status}: RequestKey) => `${method}\u0000${route}\u0000${status}`;

const parseKey = (key: string): RequestKey => {
    const [method, route, status] = key.split("\u0000");

    return {
        method,
        route,
        status: Number(status),
    };
};

const labels = ({method, route, status}: RequestKey) =>
    `method="${escapeLabel(method)}",route="${escapeLabel(route)}",status="${status}"`;

const escapeLabel = (value: string) => value.replaceAll("\\", "\\\\").replaceAll("\n", "\\n").replaceAll("\"", "\\\"");

const formatNumber = (value: number) => Number.isFinite(value) ? String(value) : "0";
