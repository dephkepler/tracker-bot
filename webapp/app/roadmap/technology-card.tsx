'use client'

import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/lib/api'
import type { CardStateResponse, PlanResult, RoadmapTechnology } from '@/lib/api-types'
import { useAIJob } from '@/lib/use-ai-job'
import { Meter } from '@/components/ui/meter'
import { difficultyIcon, kindIcon } from './progress'
import { QuizPanel } from './quiz-panel'

export function TechnologyCard({ tech }: { tech: RoadmapTechnology }) {
  const queryClient = useQueryClient()
  const [quizCardID, setQuizCardID] = useState<number | null>(null)
  const plan = useAIJob<PlanResult>()

  const setDone = useMutation({
    // A PUT carrying the state wanted, so a double tap or a retried request
    // lands on the same result instead of undoing itself.
    mutationFn: ({ cardID, done }: { cardID: number; done: boolean }) =>
      api<CardStateResponse>(`/v1/roadmap/cards/${cardID}/done`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ done }),
      }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['roadmap'] }),
  })

  async function generatePlan() {
    const result = await plan.run(`/v1/roadmap/technologies/${tech.id}/plan`)
    if (result) await queryClient.invalidateQueries({ queryKey: ['roadmap'] })
  }

  return (
    <div className='rounded-card border border-line bg-surface p-4'>
      <div className='flex items-baseline justify-between gap-3'>
        <h3 className='min-w-0 truncate text-h2 font-semibold text-ink'>{tech.name}</h3>
        <span className='shrink-0 text-small tabular-nums text-ink-3'>
          {tech.done_cards}/{tech.total_cards}
        </span>
      </div>

      {tech.mastery_criteria && <p className='mt-1 text-small text-ink-3'>{tech.mastery_criteria}</p>}

      <Meter className='mt-3' value={tech.done_cards} target={Math.max(tech.total_cards, 1)} />

      {tech.cards.length > 0 ? (
        <ul className='mt-3 flex flex-col gap-1'>
          {tech.cards.map((card) => (
            <li key={card.id}>
              <div className='flex items-start gap-2'>
                <button
                  type='button'
                  onClick={() => setDone.mutate({ cardID: card.id, done: !card.done })}
                  disabled={setDone.isPending}
                  className='mt-0.5 shrink-0 text-body'
                  aria-label={card.done ? 'Снять отметку' : 'Отметить выполненной'}
                  aria-pressed={card.done}
                >
                  {card.done ? '☑' : '☐'}
                </button>

                <span className={`min-w-0 flex-1 text-body ${card.done ? 'text-ink-3 line-through' : 'text-ink'}`}>
                  <span aria-hidden='true'>
                    {difficultyIcon(card.difficulty)} {kindIcon(card.kind)}{' '}
                  </span>
                  {card.text}
                </span>

                {/* Only on pending cards: quizzing something already ticked off
                    is the one case where the button has nothing to offer. */}
                {!card.done && (
                  <button
                    type='button'
                    onClick={() => setQuizCardID(quizCardID === card.id ? null : card.id)}
                    className='shrink-0 text-body text-ink-3'
                    aria-label='Проверить знание'
                  >
                    ❓
                  </button>
                )}
              </div>

              {quizCardID === card.id && (
                <QuizPanel
                  card={card}
                  onClose={() => setQuizCardID(null)}
                  onMarkDone={() => {
                    setDone.mutate({ cardID: card.id, done: true })
                    setQuizCardID(null)
                  }}
                />
              )}
            </li>
          ))}
        </ul>
      ) : (
        <p className='mt-3 text-body text-ink-3'>Карточек пока нет.</p>
      )}

      <div className='mt-4 flex flex-wrap items-center gap-2'>
        <button
          type='button'
          onClick={generatePlan}
          disabled={plan.isRunning}
          className='rounded-control bg-accent px-3 py-2 text-body font-medium text-white disabled:opacity-60'
        >
          {plan.isRunning ? 'Составляю план…' : '✨ План от ИИ'}
        </button>
        {plan.result && (
          <span className='text-small text-ink-2'>
            добавлено {plan.result.added}
            {plan.result.rejected > 0 && `, отброшено ${plan.result.rejected}`}
          </span>
        )}
        {plan.error && <span className='text-small text-bad'>{plan.error}</span>}
      </div>

      {plan.isRunning && (
        <p className='mt-2 text-small text-ink-3'>
          Это занимает до минуты. Можно закрыть приложение — работа продолжится на сервере.
        </p>
      )}

      {setDone.isError && <p className='mt-2 text-small text-bad'>Не удалось сохранить отметку.</p>}
    </div>
  )
}
