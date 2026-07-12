import { NextRequest, NextResponse } from "next/server"
import {
    ACCESS_TOKEN_COOKIE,
    LEGACY_ACCESS_TOKEN_COOKIE,
    LEGACY_REFRESH_TOKEN_COOKIE,
    REFRESH_TOKEN_COOKIE,
} from "@/lib/auth/cookies"
import { getApiUrl } from "@/lib/api/url"
import { getSingleFlightRefreshCookies } from "@/lib/auth/refresh-single-flight"

const AUTH_REFRESH_PATH = "/v1/auth/refresh"

const splitSetCookieHeader = (header: string) => header.split(/,(?=\s*[^,\s]+=)/)

const getSetCookieHeaders = (headers: Headers) => {
    const getSetCookie = (headers as Headers & { getSetCookie?: () => string[] }).getSetCookie

    if (getSetCookie) {
        return getSetCookie.call(headers)
    }

    const setCookie = headers.get("set-cookie")

    return setCookie ? splitSetCookieHeader(setCookie) : []
}

const parseSetCookie = (cookie: string) => {
    const [nameValue] = cookie.split(";")
    const separatorIndex = nameValue.indexOf("=")

    if (separatorIndex === -1) {
        return null
    }

    return {
        name: nameValue.slice(0, separatorIndex).trim(),
        value: nameValue.slice(separatorIndex + 1).trim(),
    }
}

const hasCookie = (request: NextRequest, ...names: string[]) =>
    names.some((name) => Boolean(request.cookies.get(name)?.value))

const getCookieValue = (request: NextRequest, ...names: string[]) => {
    for (const name of names) {
        const value = request.cookies.get(name)?.value

        if (value) {
            return value
        }
    }

    return null
}

const createRefreshedRequestHeaders = (request: NextRequest, setCookies: string[]) => {
    const requestHeaders = new Headers(request.headers)
    const cookies = new Map(request.cookies.getAll().map((cookie) => [cookie.name, cookie.value]))

    setCookies.forEach((cookie) => {
        const parsedCookie = parseSetCookie(cookie)

        if (parsedCookie) {
            cookies.set(parsedCookie.name, parsedCookie.value)
        }
    })

    requestHeaders.set(
        "Cookie",
        Array.from(cookies.entries())
            .map(([name, value]) => `${name}=${value}`)
            .join("; "),
    )

    return requestHeaders
}

const shouldSkipRefresh = (request: NextRequest) => {
    const pathname = request.nextUrl.pathname

    return (
        pathname.startsWith("/_next/") ||
        pathname.startsWith("/favicon.ico") ||
        pathname.startsWith("/robots.txt") ||
        pathname.startsWith("/api/backend/v1/auth/")
    )
}

export const proxy = async (request: NextRequest) => {
    const refreshToken = getCookieValue(request, REFRESH_TOKEN_COOKIE, LEGACY_REFRESH_TOKEN_COOKIE)

    if (
        shouldSkipRefresh(request) ||
        hasCookie(request, ACCESS_TOKEN_COOKIE, LEGACY_ACCESS_TOKEN_COOKIE) ||
        !refreshToken
    ) {
        return NextResponse.next()
    }

    const refreshedCookies = await getSingleFlightRefreshCookies(refreshToken, async () => {
        const cookie = request.headers.get("cookie")
        const refreshResponse = await fetch(getApiUrl(AUTH_REFRESH_PATH), {
            method: "POST",
            headers: cookie ? { Cookie: cookie } : undefined,
            cache: "no-store",
        })

        if (!refreshResponse.ok) {
            return []
        }

        return getSetCookieHeaders(refreshResponse.headers)
    })

    if (refreshedCookies.length === 0) {
        const response = NextResponse.next()

        response.cookies.delete(ACCESS_TOKEN_COOKIE)
        response.cookies.delete(REFRESH_TOKEN_COOKIE)
        return response
    }

    const response = NextResponse.next({
        request: {
            headers: createRefreshedRequestHeaders(request, refreshedCookies),
        },
    })

    refreshedCookies.forEach((cookieHeader) => {
        response.headers.append("Set-Cookie", cookieHeader)
    })

    return response
}
