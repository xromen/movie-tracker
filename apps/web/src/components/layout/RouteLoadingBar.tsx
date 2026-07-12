"use client"

import { useCallback, useEffect, useRef, useState } from "react"
import { usePathname, useSearchParams } from "next/navigation"
import styles from "./RouteLoadingBar.module.css"

type LoadingState = "idle" | "loading" | "finishing"

const START_DELAY_MS = 80
const MIN_VISIBLE_MS = 220
const FINISH_ANIMATION_MS = 260
const MAX_VISIBLE_MS = 10_000

const getUrlKey = (url: URL) => `${url.pathname}${url.search}`

const RouteLoadingBar = () => {
    const pathname = usePathname()
    const searchParams = useSearchParams()
    const [loadingState, setLoadingState] = useState<LoadingState>("idle")
    const currentUrlRef = useRef("")
    const startedAtRef = useRef(0)
    const navigationIdRef = useRef(0)
    const startTimerRef = useRef<number | null>(null)
    const finishTimerRef = useRef<number | null>(null)
    const idleTimerRef = useRef<number | null>(null)
    const safetyTimerRef = useRef<number | null>(null)
    const loadingVisibleRef = useRef(false)
    const mountedRef = useRef(false)

    const clearStartTimer = useCallback(() => {
        if (startTimerRef.current) {
            window.clearTimeout(startTimerRef.current)
            startTimerRef.current = null
        }
    }, [])

    const clearFinishTimers = useCallback(() => {
        if (finishTimerRef.current) {
            window.clearTimeout(finishTimerRef.current)
            finishTimerRef.current = null
        }

        if (idleTimerRef.current) {
            window.clearTimeout(idleTimerRef.current)
            idleTimerRef.current = null
        }

        if (safetyTimerRef.current) {
            window.clearTimeout(safetyTimerRef.current)
            safetyTimerRef.current = null
        }
    }, [])

    const clearTimers = useCallback(() => {
        clearStartTimer()
        clearFinishTimers()
    }, [clearStartTimer, clearFinishTimers])

    const finishLoading = useCallback((navigationId = navigationIdRef.current) => {
        navigationIdRef.current += 1
        clearTimers()

        if (!loadingVisibleRef.current) {
            return
        }

        const elapsed = Date.now() - startedAtRef.current
        const wait = Math.max(0, MIN_VISIBLE_MS - elapsed)

        finishTimerRef.current = window.setTimeout(() => {
            if (navigationId < navigationIdRef.current - 1) {
                return
            }

            setLoadingState("finishing")

            idleTimerRef.current = window.setTimeout(() => {
                loadingVisibleRef.current = false
                setLoadingState("idle")
            }, FINISH_ANIMATION_MS)
        }, wait)
    }, [clearTimers])

    const requestLoading = useCallback(() => {
        const navigationId = navigationIdRef.current + 1
        navigationIdRef.current = navigationId
        clearTimers()

        startTimerRef.current = window.setTimeout(() => {
            if (navigationId !== navigationIdRef.current) {
                return
            }

            startedAtRef.current = Date.now()
            loadingVisibleRef.current = true
            setLoadingState("loading")

            safetyTimerRef.current = window.setTimeout(() => {
                if (navigationId === navigationIdRef.current) {
                    finishLoading(navigationId)
                }
            }, MAX_VISIBLE_MS)
        }, START_DELAY_MS)
    }, [clearTimers, finishLoading])

    useEffect(() => {
        currentUrlRef.current = `${pathname}${searchParams.toString() ? `?${searchParams.toString()}` : ""}`

        if (!mountedRef.current) {
            mountedRef.current = true
            return
        }

        finishLoading()

        return clearTimers
    }, [pathname, searchParams, finishLoading, clearTimers])

    useEffect(() => {
        const shouldTrackUrl = (url: URL) =>
            url.origin === window.location.origin && getUrlKey(url) !== currentUrlRef.current

        const handleClick = (event: MouseEvent) => {
            if (event.defaultPrevented || event.button !== 0 || event.metaKey || event.ctrlKey || event.shiftKey || event.altKey) {
                return
            }

            const anchor = (event.target as Element | null)?.closest("a[href]") as HTMLAnchorElement | null

            if (!anchor || anchor.target || anchor.hasAttribute("download")) {
                return
            }

            const url = new URL(anchor.href)

            if (shouldTrackUrl(url)) {
                requestLoading()
            }
        }

        const originalPushState = window.history.pushState
        const originalReplaceState = window.history.replaceState

        window.history.pushState = function pushState(data, unused, url) {
            if (url !== undefined) {
                const nextUrl = new URL(String(url), window.location.href)

                if (shouldTrackUrl(nextUrl)) {
                    requestLoading()
                }
            }

            return originalPushState.apply(this, [data, unused, url])
        }

        window.history.replaceState = function replaceState(data, unused, url) {
            if (url !== undefined) {
                const nextUrl = new URL(String(url), window.location.href)

                if (shouldTrackUrl(nextUrl)) {
                    requestLoading()
                }
            }

            return originalReplaceState.apply(this, [data, unused, url])
        }

        window.addEventListener("click", handleClick, true)

        return () => {
            clearTimers()
            window.removeEventListener("click", handleClick, true)
            window.history.pushState = originalPushState
            window.history.replaceState = originalReplaceState
        }
    }, [requestLoading, clearTimers])

    return <div className={`${styles.bar} ${styles[loadingState]}`} aria-hidden="true" />
}

export default RouteLoadingBar