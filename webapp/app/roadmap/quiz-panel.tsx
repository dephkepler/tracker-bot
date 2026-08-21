'use client'

import { useState } from 'react'
import { useAIJob } from '@/lib/use-ai-job'
import type { GradeResult, QuizResult, RoadmapCard } from '@/lib/api-types'

const VERDICT: Record<string, { label: string; className: string }> = {
  correct: { label: '✅ Верно', className: 'text-good' },
  partial: { label: '🟡 Почти', className: 'text-warn' },
  wrong: { label: '🔴 Мимо', className: 'text-bad' },
}

// Ask a question about one card, then have the answer graded.
//
// Both steps are AI jobs, so both are started and polled. The question is held
// here in the component and sent back for grading, which is why the server keeps
// no quiz state: closing this panel and opening it again just asks a new
// question, and nothing can be left dangling on the server.
export function QuizPanel({ card, onClose, onMarkDone }: { card: RoadmapCard; onClose: () => void; onMarkDone: () => void }) {
  const quiz = useAIJob<QuizResult>()
  const grade = useAIJob<GradeResult>()
  const [answer, setAnswer] = useState('')

  async function ask() {
    grade.reset()
    setAnswer('')
    await quiz.run(`/v1/roadmap/cards/${card.id}/quiz`)
  }

  async function submit() {
    if (!quiz.result || answer.trim() === '') return
    await grade.run(`/v1/roadmap/cards/${card.id}/quiz/grade`, {
      card_text: quiz.result.card_text,
      question: quiz.result.question,
      answer: answer.trim(),
    })
  }

  const verdict = grade.result ? (VERDICT[grade.result.verdict] ?? VERDICT.partial) : null

  return (
    <div className='mt-2 rounded-control border border-line bg-surface-2 p-3'>
      <div className='flex items-baseline justify-between gap-2'>
        <span className='text-small text-ink-3'>Проверка знаний</span>
        <button type='button' onClick={onClose} className='text-small text-ink-3'>
          закрыть
        </button>
      </div>

      {!quiz.result && !quiz.isRunning && !quiz.error && (
        <button type='button' onClick={ask} className='mt-2 rounded-control bg-accent px-3 py-2 text-body font-medium text-white'>
          Задать вопрос
        </button>
      )}

      {quiz.isRunning && <p className='mt-2 text-body text-ink-3'>Придумываю вопрос…</p>}
      {quiz.error && <p className='mt-2 text-body text-bad'>{quiz.error}</p>}

      {quiz.result && (
        <>
          <p className='mt-2 text-body text-ink'>{quiz.result.question}</p>

          <textarea
            value={answer}
            onChange={(e) => setAnswer(e.target.value)}
            rows={4}
            placeholder='Ответь в двух-трёх предложениях'
            className='mt-2 w-full rounded-control border border-line bg-surface p-2 text-body text-ink'
          />

          <div className='mt-2 flex flex-wrap gap-2'>
            <button
              type='button'
              onClick={submit}
              disabled={grade.isRunning || answer.trim() === ''}
              className='rounded-control bg-accent px-3 py-2 text-body font-medium text-white disabled:opacity-50'
            >
              {grade.isRunning ? 'Проверяю…' : 'Проверить'}
            </button>
            <button type='button' onClick={ask} className='rounded-control px-3 py-2 text-body text-ink-3'>
              Другой вопрос
            </button>
          </div>
        </>
      )}

      {grade.error && <p className='mt-2 text-body text-bad'>{grade.error}</p>}

      {grade.result && verdict && (
        <div className='mt-3 border-t border-line pt-3'>
          <p className={`text-body font-medium ${verdict.className}`}>{verdict.label}</p>
          <p className='mt-1 text-body text-ink-2'>{grade.result.feedback}</p>
          {/* Offered whatever the verdict: whether the answer counts as knowing
              the card is the user's call, not the model's. */}
          {!card.done && (
            <button
              type='button'
              onClick={onMarkDone}
              className='mt-3 rounded-control border border-line px-3 py-2 text-body text-ink'
            >
              ✅ Отметить выученным
            </button>
          )}
        </div>
      )}
    </div>
  )
}
