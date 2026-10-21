<script lang="ts">
  import { fmtHour, fmtInt } from '../lib/format.ts';
  import { WEEKDAYS, WEEKDAYS_LONG } from '../lib/labels.ts';

  interface Props {
    /** Alerts per hour-of-week cell (weekday * 24 + hour), under every filter except this one. */
    counts: Int32Array;
    selected: ReadonlySet<number>;
    onpick: (cells: number[], additive: boolean) => void;
  }

  let { counts, selected, onpick }: Props = $props();

  const HOURS = Array.from({ length: 24 }, (_, h) => h);
  const DAYS = Array.from({ length: 7 }, (_, d) => d);

  let grid: HTMLDivElement;
  let focusCell = $state(4 * 24 + 15);

  const max = $derived(Math.max(1, ...counts));
  const busiest = $derived(counts.indexOf(max));

  /** The ramp from --heat-0 to --heat-3, in oklab so the steps look even. */
  function fill(v: number): string {
    const s = (v / max) * 3;
    const i = Math.min(2, Math.floor(s));
    const f = ((s - i) * 100).toFixed(1);
    return `color-mix(in oklab, var(--heat-${i}), var(--heat-${i + 1}) ${f}%)`;
  }

  const additive = (e: MouseEvent) => e.shiftKey || e.metaKey || e.ctrlKey;

  function onkeydown(e: KeyboardEvent) {
    const d = Math.floor(focusCell / 24);
    const h = focusCell % 24;
    const moves: Record<string, [number, number]> = {
      ArrowLeft: [d, Math.max(0, h - 1)],
      ArrowRight: [d, Math.min(23, h + 1)],
      ArrowUp: [Math.max(0, d - 1), h],
      ArrowDown: [Math.min(6, d + 1), h],
      Home: [d, 0],
      End: [d, 23],
    };
    const next = moves[e.key];
    if (!next) return;
    e.preventDefault();
    focusCell = next[0] * 24 + next[1];
    grid.querySelector<HTMLButtonElement>(`[data-cell="${focusCell}"]`)?.focus();
  }
</script>

<section class="card" aria-labelledby="heatmap-title">
  <h2 class="card-title" id="heatmap-title">When</h2>
  <p class="card-sub">
    Alerts by hour and weekday · busiest {WEEKDAYS[Math.floor(busiest / 24)]} {fmtHour(busiest % 24)} ({fmtInt(max)})
  </p>

  <div class="grid" bind:this={grid} role="group" aria-label="Alerts by hour and weekday. Arrow keys move between hours and days.">
    <span></span>
    {#each HOURS as h (h)}
      <button
        type="button"
        class="col"
        aria-label="Every day at {fmtHour(h)}"
        title="Every day at {fmtHour(h)}"
        onclick={(e) => onpick(DAYS.map((d) => d * 24 + h), additive(e))}
      >
        {#if h % 6 === 0}<span class="axis">{fmtHour(h).replace(' ', '')}</span>{/if}
      </button>
    {/each}

    {#each DAYS as d (d)}
      <button type="button" class="row" aria-label="All of {WEEKDAYS_LONG[d]}" onclick={(e) => onpick(HOURS.map((h) => d * 24 + h), additive(e))}>
        {WEEKDAYS[d]}
      </button>
      {#each HOURS as h (h)}
        {@const id = d * 24 + h}
        {@const label = `${WEEKDAYS_LONG[d]} ${fmtHour(h)}: ${fmtInt(counts[id])} alerts`}
        <button
          type="button"
          class="cell"
          class:dim={selected.size > 0 && !selected.has(id)}
          class:on={selected.has(id)}
          data-cell={id}
          tabindex={id === focusCell ? 0 : -1}
          aria-pressed={selected.has(id)}
          aria-label={label}
          title={label}
          style:background={fill(counts[id])}
          onfocus={() => (focusCell = id)}
          onclick={(e) => onpick([id], additive(e))}
          {onkeydown}
        ></button>
      {/each}
    {/each}
  </div>

  <div class="foot">
    <span class="legend">
      <span class="axis">0</span>
      <span class="ramp"></span>
      <span class="axis">{fmtInt(max)}</span>
    </span>
    <span class="hint">Click a cell, day or hour to filter</span>
  </div>
</section>

<style>
  section {
    display: flex;
    flex-direction: column;
  }

  .grid {
    display: grid;
    grid-template-columns: 34px repeat(24, minmax(0, 1fr));
    gap: 3px;
    margin-top: 14px;
  }

  .grid button {
    border: 0;
    padding: 0;
    background: none;
  }

  .col {
    height: 18px;
    text-align: left;
    white-space: nowrap;
    overflow: visible;
  }

  .row {
    text-align: left;
    font-size: 12px;
    color: var(--ink-2);
  }

  .col:hover .axis,
  .row:hover {
    color: var(--accent);
  }

  .cell {
    height: 26px;
    border-radius: 2px;
    transition: opacity 120ms;
  }

  .cell:hover {
    box-shadow: inset 0 0 0 2px var(--tick);
  }

  .cell.dim {
    opacity: 0.22;
  }

  .cell.on {
    box-shadow: inset 0 0 0 1px var(--surface);
  }

  .foot {
    display: flex;
    flex-wrap: wrap;
    justify-content: space-between;
    gap: 8px;
    margin-top: 12px;
  }

  .ramp {
    width: 120px;
    height: 8px;
    border-radius: 2px;
    background: linear-gradient(90deg in oklab, var(--heat-0), var(--heat-1), var(--heat-2), var(--heat-3));
  }

  .hint {
    font-size: 12px;
    color: var(--ink-3);
  }
</style>
