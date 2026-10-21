<script lang="ts">
  import type { Day } from '../lib/data/dataset.ts';
  import { fmtDate, fmtDateRange, fmtInt } from '../lib/format.ts';
  import { WEEKDAYS } from '../lib/labels.ts';

  interface Props {
    days: Day[];
    /** Alerts per day under every filter except dates. */
    counts: Int32Array;
    range: readonly [number, number] | null;
    onchange: (range: [number, number] | null) => void;
  }

  let { days, counts, range, onchange }: Props = $props();

  const HEIGHT = 96;
  const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];

  let plot: HTMLDivElement;
  let hovered = $state<number | null>(null);

  type Drag = { kind: 'brush'; anchor: number; startX: number; moved: boolean } | { kind: 'edge'; edge: 0 | 1 };
  let drag: Drag | null = null;

  const last = $derived(days.length - 1);
  const max = $derived(Math.max(1, ...counts));
  const peak = $derived(counts.indexOf(max));
  const lo = $derived(range ? range[0] : 0);
  const hi = $derived(range ? range[1] : last);
  const months = $derived(
    days
      .filter((d) => d.date.endsWith('-01'))
      .map((d) => ({ index: d.index, label: MONTHS[Number(d.date.slice(5, 7)) - 1] })),
  );
  const incompleteCount = $derived(days.filter((d) => d.incomplete).length);

  const pct = (i: number) => (i / days.length) * 100;
  const dayName = (i: number) => `${WEEKDAYS[days[i].weekday]} ${fmtDate(days[i].date)}`;

  function indexAt(clientX: number): number {
    const r = plot.getBoundingClientRect();
    return Math.min(last, Math.max(0, Math.floor(((clientX - r.left) / r.width) * days.length)));
  }

  function emit(a: number, b: number) {
    onchange(a === 0 && b === last ? null : [a, b]);
  }

  function onpointerdown(e: PointerEvent) {
    if (e.button !== 0) return;
    plot.setPointerCapture(e.pointerId);
    drag = { kind: 'brush', anchor: indexAt(e.clientX), startX: e.clientX, moved: false };
  }

  function startEdge(edge: 0 | 1, e: PointerEvent) {
    if (e.button !== 0) return;
    e.stopPropagation();
    plot.setPointerCapture(e.pointerId);
    drag = { kind: 'edge', edge };
  }

  function onpointermove(e: PointerEvent) {
    const i = indexAt(e.clientX);
    hovered = i;
    if (!drag) return;
    if (drag.kind === 'edge') {
      if (drag.edge === 0) emit(Math.min(i, hi), hi);
      else emit(lo, Math.max(i, lo));
      return;
    }
    if (!drag.moved && Math.abs(e.clientX - drag.startX) < 4) return;
    drag.moved = true;
    onchange([Math.min(drag.anchor, i), Math.max(drag.anchor, i)]);
  }

  function onpointerup() {
    if (drag?.kind === 'brush' && !drag.moved) {
      // A click picks one day, or clears it if it was the selection already.
      const i = drag.anchor;
      onchange(range && range[0] === i && range[1] === i ? null : [i, i]);
    }
    drag = null;
  }

  function onkeydown(edge: 0 | 1, e: KeyboardEvent) {
    const step = e.shiftKey ? 7 : 1;
    let v = edge === 0 ? lo : hi;
    if (e.key === 'ArrowLeft' || e.key === 'ArrowDown') v -= step;
    else if (e.key === 'ArrowRight' || e.key === 'ArrowUp') v += step;
    else if (e.key === 'Home') v = 0;
    else if (e.key === 'End') v = last;
    else return;
    e.preventDefault();
    v = Math.min(last, Math.max(0, v));
    if (edge === 0) emit(Math.min(v, hi), hi);
    else emit(lo, Math.max(v, lo));
  }
</script>

<section class="card" aria-labelledby="timeline-title">
  <div class="head">
    <div>
      <h2 class="card-title" id="timeline-title">Alerts per day</h2>
      <p class="card-sub">
        {#if range}
          {fmtDateRange(days[range[0]].date, days[range[1]].date)}
        {:else}
          Drag across the chart to choose dates
        {/if}
      </p>
    </div>
    <div class="legends">
      {#if range}
        <span class="legend"><span class="swatch" style:background="var(--accent)"></span>Selected</span>
        <span class="legend"><span class="swatch" style:background="var(--ghost)"></span>Other days</span>
      {/if}
      {#if incompleteCount}
        <span class="legend"><span class="swatch hatch"></span>Incomplete data ({incompleteCount} days)</span>
      {/if}
    </div>
  </div>

  <div class="track">
    <div
      class="plot"
      bind:this={plot}
      role="presentation"
      style:height="{HEIGHT}px"
      {onpointerdown}
      {onpointermove}
      {onpointerup}
      onpointercancel={onpointerup}
      onpointerleave={() => (hovered = null)}
    >
      {#each days as day (day.index)}
        {@const c = counts[day.index]}
        <span
          class="bar"
          class:out={range !== null && (day.index < range[0] || day.index > range[1])}
          class:incomplete={day.incomplete}
          class:hovered={hovered === day.index}
          style:height="{c > 0 ? Math.max(2, (c / max) * HEIGHT) : 0}px"
        ></span>
      {/each}

      {#if range}
        <div class="brush" style:left="{pct(range[0])}%" style:width="{pct(range[1] - range[0] + 1)}%"></div>
      {/if}
    </div>

    <button
      type="button"
      class="handle"
      class:idle={!range}
      style:left="{pct(lo)}%"
      role="slider"
      aria-label="Start date"
      aria-valuemin={0}
      aria-valuemax={last}
      aria-valuenow={lo}
      aria-valuetext={fmtDate(days[lo].date)}
      onpointerdown={(e) => startEdge(0, e)}
      onkeydown={(e) => onkeydown(0, e)}
    ></button>
    <button
      type="button"
      class="handle"
      class:idle={!range}
      style:left="{pct(hi + 1)}%"
      role="slider"
      aria-label="End date"
      aria-valuemin={0}
      aria-valuemax={last}
      aria-valuenow={hi}
      aria-valuetext={fmtDate(days[hi].date)}
      onpointerdown={(e) => startEdge(1, e)}
      onkeydown={(e) => onkeydown(1, e)}
    ></button>

    {#if hovered !== null}
      <div class="readout" style:left="{Math.min(88, Math.max(12, pct(hovered + 0.5)))}%">
        <strong>{fmtInt(counts[hovered])}</strong> · {dayName(hovered)}
        {#if days[hovered].incomplete}<span class="note">incomplete data</span>{/if}
      </div>
    {:else if counts[peak] > 0}
      <div class="readout peak" style:left="{Math.min(88, Math.max(12, pct(peak + 0.5)))}%">
        Busiest: {dayName(peak)} · {fmtInt(counts[peak])}
      </div>
    {/if}
  </div>

  <div class="months axis" aria-hidden="true">
    {#each months as m (m.index)}
      <span style:left="{pct(m.index)}%">{m.label}</span>
    {/each}
  </div>
</section>

<style>
  .head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-start;
    justify-content: space-between;
    gap: 8px 16px;
    margin-bottom: 36px;
  }

  .legends {
    display: flex;
    flex-wrap: wrap;
    gap: 6px 16px;
  }

  .track {
    position: relative;
  }

  .plot {
    position: relative;
    display: flex;
    align-items: flex-end;
    gap: 1px;
    cursor: crosshair;
    touch-action: pan-y;
    user-select: none;
  }

  .bar {
    flex: 1 1 0;
    min-width: 1px;
    border-radius: 1px 1px 0 0;
    background: var(--accent);
  }

  .bar.out {
    background: var(--ghost);
  }

  .bar.incomplete {
    background: repeating-linear-gradient(-45deg, var(--accent) 0 2px, transparent 2px 4px);
    box-shadow: inset 0 0 0 1px var(--accent);
  }

  .bar.out.incomplete {
    background: repeating-linear-gradient(-45deg, var(--ghost) 0 2px, transparent 2px 4px);
    box-shadow: inset 0 0 0 1px var(--ghost);
  }

  .bar.hovered {
    filter: brightness(0.75);
  }

  .swatch.hatch {
    background: repeating-linear-gradient(-45deg, var(--ink-3) 0 2px, transparent 2px 4px);
    box-shadow: inset 0 0 0 1px var(--ink-3);
  }

  .brush {
    position: absolute;
    top: -6px;
    bottom: 0;
    border: 1px solid var(--accent);
    border-radius: 4px;
    background: color-mix(in srgb, var(--accent) 7%, transparent);
    pointer-events: none;
  }

  .handle {
    position: absolute;
    top: 50%;
    width: 12px;
    height: 32px;
    margin: -16px 0 0 -6px;
    padding: 0;
    border: 1px solid var(--accent);
    border-radius: 3px;
    background: var(--surface);
    cursor: ew-resize;
    touch-action: none;
  }

  /* With nothing selected the handles stay out of the way until needed. */
  .handle.idle {
    border-color: var(--line-strong);
    opacity: 0;
    transition: opacity 120ms;
  }

  .track:hover .handle.idle,
  .handle:focus-visible {
    opacity: 1;
  }

  .handle:hover,
  .handle:focus-visible {
    border-color: var(--accent);
  }

  .readout {
    position: absolute;
    top: -28px;
    transform: translateX(-50%);
    white-space: nowrap;
    font-size: 12px;
    color: var(--ink-2);
    pointer-events: none;
  }

  .readout strong {
    color: var(--ink);
    font-weight: 600;
  }

  .readout.peak {
    color: var(--ink-3);
  }

  .note {
    margin-left: 6px;
    color: var(--ink-3);
  }

  .months {
    position: relative;
    height: 18px;
    margin-top: 4px;
    border-top: 1px solid var(--line);
  }

  .months span {
    position: absolute;
    top: 3px;
    white-space: nowrap;
  }
</style>
