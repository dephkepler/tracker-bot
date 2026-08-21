'use client'

import { useEffect, type ReactNode } from 'react'
import { initTelegram, isInsideTelegram } from '@/lib/telegram'
import { useHydrated } from '@/lib/use-hydrated'
import { Card } from '@/components/ui/card'

// Runs the Telegram handshake once and gates the app on there being a launch to
// authenticate with.
//
// The gate is cosmetic by design, and worth saying out loud so it does not look
// like the reference app's mistake of a client-only auth check: a static export
// has no server here to enforce anything. The real gate is the Go server
// verifying the init-data signature on every single request.
export function TelegramProvider({ children }: { children: ReactNode }) {
  const hydrated = useHydrated()

  useEffect(() => initTelegram(), [])

  // One frame of nothing rather than a flash of the wrong screen: until
  // hydration there is no window to ask about Telegram.
  if (!hydrated) return null

  if (!isInsideTelegram()) return <OutsideTelegram />

  return <>{children}</>
}

function OutsideTelegram() {
  const link = process.env.NEXT_PUBLIC_TG_LINK
  return (
    <main className='tg-shell mx-auto flex max-w-[680px] items-center justify-center px-4'>
      <Card className='text-center'>
        <h1 className='text-h1 font-semibold text-ink'>Это панель бота</h1>
        <p className='mt-2 text-body text-ink-2'>
          Открывать её нужно из Telegram — там приложение получает подписанный вход, без которого сервер не отдаёт данные.
        </p>
        {link && (
          <a href={link} className='mt-4 inline-block rounded-control bg-accent px-4 py-2 text-body font-medium text-white'>
            Открыть в Telegram
          </a>
        )}
      </Card>
    </main>
  )
}
