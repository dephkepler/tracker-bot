'use client'

import { useCallback, useRef, useState } from 'react'
import { api } from './api'
import type { JobResponse } from './api-types'

// Runs one AI call: start it, then poll until it settles.
//
// The polling is what makes the feature usable on a phone. These calls take
// tens of seconds, and a Mini App that gets backgrounded loses any request it
// was holding open — but not a job, which keeps running on the server and is
// still there to collect afterwards.

const POLL_MS = 1200
// Generous: the server's own job timeout is shorter, so it is the one that
// speaks first, and this is only a guard against polling for ever.
const POLL_TIMEOUT_MS = 4 * 60 * 1000

export interface AIJobState<T> {
  isRunning: boolean
  result: T | null
  error: string | null
  run: (path: string, body?: unknown) => Promise<T | null>
  reset: () => void
}

export function useAIJob<T>(): AIJobState<T> {
  const [isRunning, setRunning] = useState(false)
  const [result, setResult] = useState<T | null>(null)
  const [error, setError] = useState<string | null>(null)
  // Guards against a second run being started while one is in flight, and
  // against a stale poll writing over a newer result.
  const runId = useRef(0)

  const reset = useCallback(() => {
    runId.current += 1
    setRunning(false)
    setResult(null)
    setError(null)
  }, [])

  const run = useCallback(async (path: string, body?: unknown): Promise<T | null> => {
    runId.current += 1
    const mine = runId.current
    setRunning(true)
    setResult(null)
    setError(null)

    try {
      const started = await api<JobResponse<T>>(path, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: body === undefined ? '{}' : JSON.stringify(body),
      })

      const deadline = Date.now() + POLL_TIMEOUT_MS
      let job = started
      while (job.status === 'pending' && Date.now() < deadline) {
        await sleep(POLL_MS)
        if (runId.current !== mine) return null
        job = await api<JobResponse<T>>(`/v1/ai/jobs/${job.id}`)
      }
      if (runId.current !== mine) return null

      if (job.status === 'done' && job.result !== undefined) {
        setResult(job.result)
        return job.result
      }
      // The server sends a message meant for a person; showing it beats
      // inventing one here.
      setError(job.error_message || 'ИИ не ответил, попробуй позже')
      return null
    } catch (e) {
      setError(e instanceof Error ? e.message : 'не удалось запустить запрос')
      return null
    } finally {
      if (runId.current === mine) setRunning(false)
    }
  }, [])

  return { isRunning, result, error, run, reset }
}

function sleep(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms))
}
