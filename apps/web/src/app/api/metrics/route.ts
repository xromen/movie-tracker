import {webMetrics} from "@/lib/metrics/prometheus";

export const dynamic = "force-dynamic";

export const GET = () =>
    new Response(webMetrics.render(), {
        headers: {
            "Content-Type": "text/plain; version=0.0.4; charset=utf-8",
        },
    });
