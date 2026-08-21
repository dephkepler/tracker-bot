import { getInitData } from './telegram'

// Same origin: Caddy serves the static export at / and proxies /api to the Go
// listener. So there is no API URL to bake into a static build, no CORS
// preflight doubling every request on a phone, and no stale port fallback.
const BASE = process.env.NEXT_PUBLIC_API_BASE || '/api'

export class ApiError extends Error {
  constructor(
    public status: number,
    public code: string,
    message: string
  ) {
    super(message)
  }
}

/**
 * The launch itself is the credential, so a 401 is not "log in" — there is no
 * login. Separated from other errors so the UI can say "reopen the app", which
 * is the only thing that actually fixes it, and so react-query never retries it.
 */
export class StaleLaunchError extends ApiError {
  constructor(code = 'unauthorized', message = 'stale launch') {
    super(401, code, message)
  }
}

interface ProblemBody {
  error?: { code?: string; message?: string }
}

export async function api<T>(path: string, init: RequestInit = {}): Promise<T> {
  const initData = getInitData()
  if (!initData) throw new StaleLaunchError('missing_credentials', 'no init data')

  const headers = new Headers(init.headers)
  // Telegram's own scheme for handing raw init data to a backend. Not a bearer
  // token: nothing here was ever issued to us. It travels in a header rather
  // than a query so it stays out of access logs, history and Referer.
  headers.set('Authorization', `tma ${initData}`)

  const res = await fetch(`${BASE}${path}`, { ...init, headers })
  const text = await res.text()
  const body: ProblemBody | null = text ? safeParse(text) : null
  const code = body?.error?.code ?? ''
  const message = body?.error?.message ?? res.statusText

  if (res.status === 401) throw new StaleLaunchError(code || 'unauthorized', message)
  if (!res.ok) throw new ApiError(res.status, code, message)
  return JSON.parse(text) as T
}

function safeParse(s: string): ProblemBody | null {
  try {
    return JSON.parse(s) as ProblemBody
  } catch {
    return null
  }
}
