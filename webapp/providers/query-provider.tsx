'use client'

import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { useState, type ReactNode } from 'react'
import { StaleLaunchError } from '@/lib/api'

export function QueryProvider({ children }: { children: ReactNode }) {
  const [client] = useState(
    () =>
      new QueryClient({
        defaultOptions: {
          queries: {
            // A Mini App session is short and this is aggregate history, so a
            // minute of staleness costs nothing and saves a round trip on
            // every navigation.
            staleTime: 60_000,
            // The WebView fires focus every time it comes to the foreground.
            refetchOnWindowFocus: false,
            // Retrying a dead launch just burns the user's time: only a fresh
            // launch can fix it.
            retry: (count, error) => !(error instanceof StaleLaunchError) && count < 1,
          },
        },
      })
  )

  return <QueryClientProvider client={client}>{children}</QueryClientProvider>
}
