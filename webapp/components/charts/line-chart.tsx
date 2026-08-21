'use client'

import { useRef, useState } from 'react'
import { cx } from '@/lib/cx'

// Change over time, as a line with an optional area fill.
//
// Ported from the reference app's line-chart with three deliberate changes:
//
//  1. Every colour comes from a CSS variable, so the chart lives in both
//     themes. The original hardcoded greys and a light-mode gridline hex,
//     which on a dark surface leaves the axis invisible.
//  2. A `reference` line, for a target the series is measured against. This is
//     what replaces the bot's hardcoded 120-minute progress bar.
//  3. The point markers moved out of the SVG. With preserveAspectRatio="none"
//     the viewBox is stretched unequally — 600×200 rendered into 360×160 is
//     0.60 across and 0.80 down — so an SVG <circle r=2.5> is drawn as a 1.5×2
//     ellipse and strokeWidth is thicker vertically than horizontally. Lines
//     now carry vector-effect="non-scaling-stroke", and the single hovered
//     marker is an HTML element positioned in percentages, which is round at
//     any size and large enough to see. Drawing a dot on all thirty points was
//     noise at phone width anyway.

export interface LineSeries {
  key: string
  label: string
  /** A CSS colour — pass a var(), not a hex, or dark mode breaks. */
  color: string
  values: number[]
}

// Fixed internal coordinate system; the SVG scales to its rendered width.
const VB_W = 600
const VB_H = 200
const PAD_L = 40
const PAD_B = 4

function niceCeil(max: number): number {
  if (max <= 0) return 1
  const magnitude = 10 ** Math.floor(Math.log10(max))
  const norm = max / magnitude
  const niceNorm = norm <= 1 ? 1 : norm <= 2 ? 2 : norm <= 5 ? 5 : 10
  return niceNorm * magnitude
}

export function LineChart({
  series,
  xLabels,
  formatValue = (v) => String(v),
  height = 176,
  area = false,
  reference,
  labelStep,
}: {
  series: LineSeries[]
  xLabels: string[]
  formatValue?: (v: number) => string
  height?: number
  area?: boolean
  /** A target line, e.g. the daily goal. */
  reference?: { value: number; label: string }
  labelStep?: number
}) {
  const svgRef = useRef<SVGSVGElement>(null)
  const [hoverIdx, setHoverIdx] = useState<number | null>(null)

  const n = xLabels.length
  if (n === 0 || series.every((s) => s.values.length === 0)) {
    return <p className='text-body text-ink-3'>Нет данных за период.</p>
  }

  const dataMax = Math.max(1, ...series.flatMap((s) => s.values), reference?.value ?? 0)
  const niceMax = niceCeil(dataMax)
  const plotW = VB_W - PAD_L
  const plotH = VB_H - PAD_B
  const xAt = (i: number) => PAD_L + (n <= 1 ? plotW / 2 : (i / (n - 1)) * plotW)
  const yAt = (v: number) => plotH - (v / niceMax) * plotH
  const step = labelStep ?? Math.max(1, Math.ceil(n / 8))

  const pointsFor = (values: number[]) => values.map((v, i) => `${xAt(i)},${yAt(v)}`).join(' ')

  function updateFromClientX(clientX: number) {
    const svg = svgRef.current
    if (!svg) return
    const rect = svg.getBoundingClientRect()
    const xUser = ((clientX - rect.left) / rect.width) * VB_W
    const idx = Math.round(((xUser - PAD_L) / plotW) * (n - 1))
    setHoverIdx(Math.min(n - 1, Math.max(0, idx)))
  }

  function onKeyDown(e: React.KeyboardEvent) {
    if (e.key === 'ArrowRight') setHoverIdx((i) => Math.min(n - 1, (i ?? -1) + 1))
    else if (e.key === 'ArrowLeft') setHoverIdx((i) => Math.max(0, (i ?? 1) - 1))
    else if (e.key === 'Escape') setHoverIdx(null)
    else return
    e.preventDefault()
  }

  return (
    <div>
      {/* A single series needs no legend — the chart's own title names it. */}
      {series.length >= 2 && (
        <div className='mb-2 flex flex-wrap items-center gap-4 text-small text-ink-3'>
          {series.map((s) => (
            <span key={s.key} className='flex items-center gap-1.5'>
              <span className='h-0.5 w-3 rounded-full' style={{ backgroundColor: s.color }} />
              {s.label}
            </span>
          ))}
        </div>
      )}

      <div className='relative'>
        <svg
          ref={svgRef}
          viewBox={`0 0 ${VB_W} ${VB_H}`}
          preserveAspectRatio='none'
          className='block w-full touch-none outline-none'
          style={{ height }}
          tabIndex={0}
          role='img'
          aria-label={`${series.map((s) => s.label).join(', ')} — стрелками можно пройти по точкам`}
          onPointerMove={(e) => updateFromClientX(e.clientX)}
          onPointerDown={(e) => updateFromClientX(e.clientX)}
          onPointerLeave={() => setHoverIdx(null)}
          onKeyDown={onKeyDown}
        >
          {[0, 0.5, 1].map((g) => {
            const y = plotH - g * plotH
            return (
              <g key={g}>
                <line
                  x1={PAD_L}
                  x2={VB_W}
                  y1={y}
                  y2={y}
                  stroke='var(--grid)'
                  strokeWidth={1}
                  vectorEffect='non-scaling-stroke'
                />
                <text
                  x={PAD_L - 6}
                  y={y === plotH ? y - 2 : y}
                  textAnchor='end'
                  dominantBaseline='middle'
                  fill='var(--ink-3)'
                  fontSize={9}
                >
                  {formatValue(Math.round(niceMax * g))}
                </text>
              </g>
            )
          })}

          {reference && reference.value > 0 && (
            <line
              x1={PAD_L}
              x2={VB_W}
              y1={yAt(reference.value)}
              y2={yAt(reference.value)}
              stroke='var(--ink-3)'
              strokeWidth={1}
              strokeDasharray='4 4'
              vectorEffect='non-scaling-stroke'
            />
          )}

          {area && series.length === 1 && (
            <polygon
              points={`${xAt(0)},${plotH} ${pointsFor(series[0].values)} ${xAt(n - 1)},${plotH}`}
              fill={series[0].color}
              opacity={0.12}
            />
          )}

          {series.map((s) => (
            <polyline
              key={s.key}
              points={pointsFor(s.values)}
              fill='none'
              stroke={s.color}
              strokeWidth={2}
              strokeLinecap='round'
              strokeLinejoin='round'
              vectorEffect='non-scaling-stroke'
            />
          ))}

          {hoverIdx !== null && (
            <line
              x1={xAt(hoverIdx)}
              x2={xAt(hoverIdx)}
              y1={0}
              y2={plotH}
              stroke='var(--ink-3)'
              strokeWidth={1}
              vectorEffect='non-scaling-stroke'
            />
          )}
        </svg>

        {/* The marker is HTML, so it stays a circle whatever the SVG is
            stretched to, and it is 10px across rather than the 5px an r=2.5
            circle would be. */}
        {hoverIdx !== null &&
          series.map((s) => (
            <span
              key={`marker-${s.key}`}
              className='pointer-events-none absolute z-10 h-2.5 w-2.5 -translate-x-1/2 -translate-y-1/2 rounded-full border-2 border-surface'
              style={{
                left: `${(xAt(hoverIdx) / VB_W) * 100}%`,
                top: `${(yAt(s.values[hoverIdx] ?? 0) / VB_H) * 100}%`,
                backgroundColor: s.color,
              }}
            />
          ))}

        {hoverIdx !== null && (
          <div
            className={cx(
              'pointer-events-none absolute top-0 z-10 rounded-control bg-ink px-2.5 py-1.5 text-small whitespace-nowrap text-surface',
              hoverIdx / (n - 1 || 1) > 0.75 ? '-translate-x-full' : hoverIdx === 0 ? '' : '-translate-x-1/2'
            )}
            style={{ left: `${(xAt(hoverIdx) / VB_W) * 100}%` }}
          >
            <div className='opacity-70'>{xLabels[hoverIdx]}</div>
            {series.map((s) => (
              <div key={s.key} className='font-semibold tabular-nums'>
                {formatValue(s.values[hoverIdx] ?? 0)}
              </div>
            ))}
          </div>
        )}
      </div>

      <div className='mt-1 flex' style={{ paddingLeft: `${(PAD_L / VB_W) * 100}%` }}>
        {xLabels.map((label, i) => (
          <div key={i} className='min-w-0 flex-1 truncate text-center text-micro text-ink-3'>
            {i % step === 0 ? label : ''}
          </div>
        ))}
      </div>
    </div>
  )
}
