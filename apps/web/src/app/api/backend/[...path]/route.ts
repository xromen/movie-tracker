import {NextRequest, NextResponse} from "next/server";
import {
    ACCESS_TOKEN_COOKIE,
    LEGACY_ACCESS_TOKEN_COOKIE,
    LEGACY_REFRESH_TOKEN_COOKIE,
    REFRESH_TOKEN_COOKIE,
} from "@/lib/auth/cookies";
import {getApiUrl} from "@/lib/api/url";
import {getSingleFlightRefreshCookies} from "@/lib/auth/refresh-single-flight";

interface BackendRouteContext {
    params: Promise<{
        path: string[];
    }>;
}

const FORWARDED_REQUEST_HEADERS = ["accept", "content-type", "cookie"];
const FORWARDED_RESPONSE_HEADERS = ["cache-control", "content-type"];
const AUTH_SESSION_CHANGED_HEADER = "X-Auth-Session-Changed";
const AUTH_REFRESH_PATH = "/v1/auth/refresh";

const splitSetCookieHeader = (header: string) => header.split(/,(?=\s*[^;,\s]+=)/);

const getSetCookieHeaders = (headers: Headers) => {
    const getSetCookie = (headers as Headers & { getSetCookie?: () => string[] }).getSetCookie;

    if (getSetCookie) {
        return getSetCookie.call(headers);
    }

    const setCookie = headers.get("set-cookie");

    return setCookie ? splitSetCookieHeader(setCookie) : [];
};

const createBackendHeaders = (request: NextRequest) => {
    const headers = new Headers();

    FORWARDED_REQUEST_HEADERS.forEach((headerName) => {
        const value = request.headers.get(headerName);

        if (value) {
            headers.set(headerName, value);
        }
    });

    return headers;
};

const getCookieValue = (request: NextRequest, ...names: string[]) => {
    for (const name of names) {
        const value = request.cookies.get(name)?.value;

        if (value) {
            return value;
        }
    }

    return null;
};

const parseSetCookie = (cookie: string) => {
    const [nameValue] = cookie.split(";");
    const separatorIndex = nameValue.indexOf("=");

    if (separatorIndex === -1) {
        return null;
    }

    return {
        name: nameValue.slice(0, separatorIndex).trim(),
        value: nameValue.slice(separatorIndex + 1).trim(),
    };
};

const createCookieHeader = (request: NextRequest, refreshedCookies: string[]) => {
    const cookies = new Map(
        request.cookies.getAll().map((cookie) => [cookie.name, cookie.value]),
    );

    refreshedCookies.forEach((cookie) => {
        const parsedCookie = parseSetCookie(cookie);

        if (parsedCookie) {
            cookies.set(parsedCookie.name, parsedCookie.value);
        }
    });

    return Array.from(cookies.entries())
        .map(([name, value]) => `${name}=${value}`)
        .join("; ");
};

const refreshAuthCookies = async (request: NextRequest) => {
    const accessToken = getCookieValue(request, ACCESS_TOKEN_COOKIE, LEGACY_ACCESS_TOKEN_COOKIE);
    const refreshToken = getCookieValue(request, REFRESH_TOKEN_COOKIE, LEGACY_REFRESH_TOKEN_COOKIE);

    if (accessToken || !refreshToken) {
        return {
            refreshedCookies: [] as string[],
            didTryRefresh: false,
        };
    }

    const refreshedCookies = await getSingleFlightRefreshCookies(refreshToken, async () => {
        const headers = createBackendHeaders(request);
        const response = await fetch(getApiUrl(AUTH_REFRESH_PATH), {
            method: "POST",
            headers,
            cache: "no-store",
        });

        if (!response.ok) {
            return [];
        }

        return getSetCookieHeaders(response.headers);
    });

    return {
        refreshedCookies,
        didTryRefresh: true,
    };
};

const createProxyResponse = async (
    response: Response,
    setCookies: string[] = [],
    authSessionChanged = false,
) => {
    const responseBody = response.status === 204 ? null : await response.text();
    const responseHeaders = new Headers();

    FORWARDED_RESPONSE_HEADERS.forEach((headerName) => {
        const value = response.headers.get(headerName);

        if (value) {
            responseHeaders.set(headerName, value);
        }
    });

    [...setCookies, ...getSetCookieHeaders(response.headers)].forEach((cookie) => {
        responseHeaders.append("Set-Cookie", cookie);
    });

    if (authSessionChanged || response.status === 401) {
        responseHeaders.set(AUTH_SESSION_CHANGED_HEADER, "1");
    }

    return new NextResponse(responseBody, {
        status: response.status,
        headers: responseHeaders,
    });
};

const proxyRequest = async (request: NextRequest, context: BackendRouteContext) => {
    const {path} = await context.params;
    const headers = createBackendHeaders(request);
    const search = request.nextUrl.search;
    const apiPath = `/${path.join("/")}`;
    const body = request.method === "GET" || request.method === "HEAD" ? undefined : await request.text();
    const shouldRefreshBeforeRequest = apiPath !== AUTH_REFRESH_PATH && !apiPath.startsWith("/v1/auth/");
    const refreshResult = shouldRefreshBeforeRequest
        ? await refreshAuthCookies(request)
        : {refreshedCookies: [] as string[], didTryRefresh: false};

    if (body && !headers.has("Content-Type")) {
        headers.set("Content-Type", "application/json");
    }

    if (refreshResult.refreshedCookies.length > 0) {
        headers.set("Cookie", createCookieHeader(request, refreshResult.refreshedCookies));
    }

    const apiUrl = getApiUrl(`${apiPath}${search}`);
    const requestInit: RequestInit = {
        method: request.method,
        headers,
        body,
        cache: "no-store",
    };
    const response = await fetch(apiUrl, requestInit);

    return createProxyResponse(
        response,
        refreshResult.refreshedCookies,
        refreshResult.didTryRefresh,
    );
};

export const GET = proxyRequest;
export const POST = proxyRequest;
export const PUT = proxyRequest;
export const PATCH = proxyRequest;
export const DELETE = proxyRequest;
