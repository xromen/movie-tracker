const BACKEND_PROXY_PREFIX = "/api/backend";
const AUTH_SESSION_CHANGED_HEADER = "X-Auth-Session-Changed";
const AUTH_SESSION_CHANGED_EVENT = "movie-tracker:auth-session-changed";

type RequestHeaders = Pick<Headers, "get">;

export class ApiError extends Error {
    constructor(
        message: string,
        public readonly status: number,
    ) {
        super(message);
    }
}

interface FetchApiOptions {
    revalidate?: number;
    method?: string;
    body?: unknown;
    headers?: HeadersInit;
}

const isBrowser = () => typeof window !== "undefined";

export const getBackendProxyPath = (path: string) => {
    const normalizedPath = path.startsWith("/") ? path : `/${path}`;

    return `${BACKEND_PROXY_PREFIX}${normalizedPath}`;
};

const getNextRequestHeaders = async (): Promise<RequestHeaders | null> => {
    if (isBrowser()) {
        return null;
    }

    const importer = new Function("specifier", "return import(specifier)") as (
        specifier: string,
    ) => Promise<{ headers: () => Promise<RequestHeaders> | RequestHeaders }>;
    const { headers } = await importer("next/headers.js");

    return headers();
};

const getServerOrigin = (requestHeaders: RequestHeaders | null) => {
    const host = requestHeaders?.get("x-forwarded-host") ?? requestHeaders?.get("host");

    if (!host) {
        return process.env.NEXT_PUBLIC_APP_URL ?? "http://localhost:3000";
    }

    const protocol =
        requestHeaders?.get("x-forwarded-proto") ??
        (host.startsWith("localhost") || host.startsWith("127.0.0.1") ? "http" : "https");

    return `${protocol}://${host}`;
};

const getApiRequestUrl = (path: string, requestHeaders: RequestHeaders | null) => {
    const proxyPath = getBackendProxyPath(path);

    return isBrowser() ? proxyPath : `${getServerOrigin(requestHeaders)}${proxyPath}`;
};

const createRequestInit = (
    path: string,
    options: FetchApiOptions,
    requestHeaders: RequestHeaders | null,
): RequestInit & { next?: { revalidate: number } } => {
    const { body, headers: customHeaders, method = body === undefined ? "GET" : "POST", revalidate = 300 } = options;
    const headers = new Headers(customHeaders);
    const requestInit: RequestInit & { next?: { revalidate: number } } = {
        method,
        headers,
        credentials: "include",
    };

    if (body !== undefined) {
        if (body instanceof FormData || body instanceof Blob || typeof body === "string") {
            requestInit.body = body;
        } else {
            headers.set("Content-Type", headers.get("Content-Type") ?? "application/json");
            requestInit.body = JSON.stringify(body);
        }
    }

    if (!isBrowser()) {
        const cookie = requestHeaders?.get("cookie");

        if (cookie) {
            headers.set("Cookie", cookie);
            requestInit.cache = "no-store";
        } else {
            requestInit.next = { revalidate };
        }
    }

    if (path.includes("/auth/")) {
        requestInit.cache = "no-store";
        delete requestInit.next;
    }

    return requestInit;
};

export const fetchApiResponse = async (path: string, options: FetchApiOptions = {}) => {
    const requestHeaders = await getNextRequestHeaders();
    const requestUrl = getApiRequestUrl(path, requestHeaders);
    const requestInit = createRequestInit(path, options, requestHeaders);

    const response = await fetch(requestUrl, requestInit);

    if (isBrowser() && response.headers.get(AUTH_SESSION_CHANGED_HEADER) === "1") {
        window.dispatchEvent(new Event(AUTH_SESSION_CHANGED_EVENT));
    }

    return response;
};

export const fetchApi = async <T = void>(path: string, options: FetchApiOptions = {}): Promise<T> => {
    const response = await fetchApiResponse(path, options);

    if (!response.ok) {
        throw new ApiError(`API request failed: ${response.url}`, response.status);
    }

    const responseBody = await response.text();

    if (!responseBody) {
        return undefined as T;
    }

    const contentType = response.headers.get("Content-Type") ?? "";

    if (contentType.includes("application/json")) {

        return JSON.parse(responseBody) as T;
    }

    return responseBody as T;
};