'use client'

import { useTelegramBackButton } from '@/lib/telegram'

// Renders nothing; it exists so the hook runs inside the layout and every route
// gets Telegram's own back arrow. Drawing our own would sit next to the system
// one and read as a bug — and while BackButton is hidden, Telegram maps the
// Android hardware button to "close the app", which is exactly what the root
// screen should do.
export function BackButtonSync() {
  useTelegramBackButton()
  return null
}
