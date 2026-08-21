import type { Metadata, Viewport } from 'next'
import Script from 'next/script'
import './globals.css'
import { QueryProvider } from '@/providers/query-provider'
import { TelegramProvider } from '@/providers/telegram-provider'

export const metadata: Metadata = {
  title: 'Трекер — панель',
  description: 'Сводка по трекингу времени, словам, челленджам и роадмапам',
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  // The Mini App fills a WebView with its own chrome; letting the page zoom
  // just fights Telegram's own gestures.
  maximumScale: 1,
}

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang='ru' suppressHydrationWarning>
      <head>
        {/* beforeInteractive is only legal in the root layout, and it is what
            guarantees window.Telegram exists before React hydrates — which is
            why nothing here has to poll for it. Under output: 'export' the tag
            is inlined into every emitted page at build time. */}
        <Script src='https://telegram.org/js/telegram-web-app.js' strategy='beforeInteractive' />
      </head>
      <body>
        <QueryProvider>
          <TelegramProvider>{children}</TelegramProvider>
        </QueryProvider>
      </body>
    </html>
  )
}
