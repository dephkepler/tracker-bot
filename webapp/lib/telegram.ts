'use client'

import { useEffect } from 'react'
import { usePathname, useRouter } from 'next/navigation'

// The Telegram bridge. Everything this app knows about living inside Telegram is
// here; the rest of the code sees an init-data string, a data-theme attribute
// and three CSS variables.

interface SafeAreaInset {
  top: number
  bottom: number
  left: number
  right: number
}

interface TgWebApp {
  initData: string
  version: string
  platform: string
  colorScheme: 'light' | 'dark'
  viewportStableHeight: number
  safeAreaInset?: SafeAreaInset
  contentSafeAreaInset?: SafeAreaInset
  isVersionAtLeast(v: string): boolean
  ready(): void
  expand(): void
  close(): void
  disableVerticalSwipes?(): void
  setHeaderColor?(color: string): void
  setBackgroundColor?(color: string): void
  onEvent(event: string, cb: () => void): void
  offEvent(event: string, cb: () => void): void
  BackButton: {
    show(): void
    hide(): void
    onClick(cb: () => void): void
    offClick(cb: () => void): void
  }
}

declare global {
  interface Window {
    Telegram?: { WebApp: TgWebApp }
  }
}

function webApp(): TgWebApp | null {
  if (typeof window === 'undefined') return null
  return window.Telegram?.WebApp ?? null
}

// Captured once per launch: init data is a signed snapshot Telegram hands over
// when the app opens and never refreshes, so re-reading it gains nothing — and
// caching means lib/api.ts never touches window at all.
let cached: string | null = null

export function getInitData(): string | null {
  if (cached !== null) return cached
  const wa = webApp()
  if (wa?.initData) {
    cached = wa.initData
    return cached
  }
  // Developing in a plain browser, where nothing can sign anything. Only ever a
  // client-side convenience: the Go side still verifies the HMAC, so a value
  // pasted here grants nothing the launch it came from did not already.
  const dev = process.env.NEXT_PUBLIC_DEV_INIT_DATA
  if (dev) {
    cached = dev
    return cached
  }
  return null
}

export function isInsideTelegram(): boolean {
  return getInitData() !== null
}

// We own every colour; Telegram only decides which of our two themes to show,
// and we push our own plane into its chrome so the seam under the header
// disappears. Deriving chart colours from themeParams was rejected outright — a
// user can install a Telegram theme of arbitrary colours, and a palette that
// cannot be validated against a known surface cannot encode data.
const PLANE = { light: '#f4f5f7', dark: '#17181b' } as const

function applyTheme(wa: TgWebApp) {
  const scheme = wa.colorScheme === 'dark' ? 'dark' : 'light'
  document.documentElement.dataset.theme = scheme
  wa.setHeaderColor?.(PLANE[scheme])
  wa.setBackgroundColor?.(PLANE[scheme])
}

function applyViewport(wa: TgWebApp) {
  const root = document.documentElement
  // Stable height, not viewportHeight: the latter tracks the drag mid-gesture,
  // so laying out against it makes the whole page twitch while scrolling.
  root.style.setProperty('--tg-vh', `${wa.viewportStableHeight}px`)
  root.style.setProperty('--tg-safe-top', `${wa.contentSafeAreaInset?.top ?? 0}px`)
  root.style.setProperty('--tg-safe-bottom', `${wa.safeAreaInset?.bottom ?? 0}px`)
}

/** Called once, from the Telegram provider. Returns a cleanup. */
export function initTelegram(): () => void {
  const wa = webApp()
  if (!wa) {
    // Outside Telegram only the "open me in Telegram" screen renders, so at
    // least follow the OS preference rather than blinding anyone.
    const dark = window.matchMedia?.('(prefers-color-scheme: dark)').matches
    document.documentElement.dataset.theme = dark ? 'dark' : 'light'
    return () => {}
  }

  wa.ready()
  wa.expand()
  // 7.7+. Without this an upward scroll on a long page is taken for
  // swipe-to-close and the Mini App shuts mid-read on Android.
  if (wa.isVersionAtLeast('7.7')) wa.disableVerticalSwipes?.()

  applyTheme(wa)
  applyViewport(wa)

  const onTheme = () => applyTheme(wa)
  const onViewport = () => applyViewport(wa)
  const events: Array<[string, () => void]> = [
    ['themeChanged', onTheme],
    ['viewportChanged', onViewport],
    ['safeAreaChanged', onViewport],
    ['contentSafeAreaChanged', onViewport],
  ]
  events.forEach(([e, cb]) => wa.onEvent(e, cb))
  return () => events.forEach(([e, cb]) => wa.offEvent(e, cb))
}

export function closeApp() {
  webApp()?.close()
}

// Telegram routes the Android hardware back button and its own swipe gesture to
// BackButton's handler while it is shown, and to "close the app" while hidden —
// so hiding it at the root is exactly the behaviour wanted there.
export function useTelegramBackButton() {
  const router = useRouter()
  const pathname = usePathname()

  useEffect(() => {
    const back = webApp()?.BackButton
    if (!back) return
    if (pathname === '/') {
      back.hide()
      return
    }
    const onClick = () => router.back()
    back.onClick(onClick)
    back.show()
    return () => {
      back.offClick(onClick)
      back.hide()
    }
  }, [pathname, router])
}
