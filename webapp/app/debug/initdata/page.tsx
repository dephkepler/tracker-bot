'use client'

import { useState } from 'react'
import { Card } from '@/components/ui/card'
// Pulled in for the global Window augmentation that declares window.Telegram.
import '@/lib/telegram'

// Shows the raw init data of the current launch.
//
// This exists to keep one test honest. internal/web/tgauth verifies the
// signature Telegram puts on a launch, and its unit tests sign their own
// fixtures — so an implementation and a test that misread the spec the same way
// agree with each other and stay green. That happened: the check string left
// out the "signature" field, every test passed, and every real launch was
// rejected. The cure is one captured vector from a real client, and this page
// is how it gets captured.
//
// It is not an information leak. What it prints is the credential of the launch
// the viewer is already inside, shown only to them — the same string their own
// Telegram client just handed to this page. It grants nothing they did not
// already have.
//
// To refresh the golden vector: make a throwaway bot, register a Mini App on it
// pointing at this deployment, open this page through it, and paste the string
// plus that bot's token into TestGoldenVector — then revoke the token, so what
// lands in this public repository is dead.
export default function InitDataPage() {
  const [copied, setCopied] = useState(false)
  const raw = typeof window === 'undefined' ? '' : (window.Telegram?.WebApp?.initData ?? '')

  return (
    <main className='tg-shell mx-auto max-w-[680px] px-4 pt-3'>
      <h1 className='mb-3 text-h1 font-semibold text-ink'>initData</h1>

      <Card>
        {raw ? (
          <>
            <p className='mb-3 text-small text-ink-3'>
              Строка подписи этого запуска. Она действительна сутки — снимай её одноразовым ботом, токен которого сразу отзовёшь.
            </p>
            <textarea
              readOnly
              value={raw}
              rows={8}
              onFocus={(e) => e.currentTarget.select()}
              className='w-full rounded-control border border-line bg-surface-2 p-2 font-mono text-micro text-ink'
            />
            <button
              type='button'
              className='mt-3 rounded-control bg-accent px-3 py-2 text-body font-medium text-white'
              onClick={() => {
                navigator.clipboard?.writeText(raw).then(
                  () => setCopied(true),
                  // Clipboard access is denied in some WebViews; the textarea
                  // above is the fallback, so this is not worth an error state.
                  () => setCopied(false)
                )
              }}
            >
              {copied ? 'Скопировано' : 'Скопировать'}
            </button>
          </>
        ) : (
          <p className='text-body text-ink-2'>
            Пусто — эта страница открыта не как мини-приложение Telegram, поэтому подписи нет.
          </p>
        )}
      </Card>
    </main>
  )
}
