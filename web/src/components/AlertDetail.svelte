<script lang="ts">
  import type { Dataset } from '../lib/data/dataset.ts';
  import { compass, fmtDateTime, fmtDuration, fmtHour, fmtInt } from '../lib/format.ts';
  import { roadTypeLabel, subtypeLabel, townLabel } from '../lib/labels.ts';
  import { spotStats } from '../lib/spot.ts';

  interface Props {
    ds: Dataset;
    index: number;
    mapState: Uint8Array;
    filtered: boolean;
    onclose: () => void;
    onzoom: () => void;
  }

  let { ds, index, mapState, filtered, onclose, onzoom }: Props = $props();

  const RADIUS = 300;

  let heading: HTMLHeadingElement;

  const spot = $derived(spotStats(ds, index, RADIUS, mapState));
  const hourMax = $derived(Math.max(1, ...spot.hours));
  const busiestHour = $derived(spot.hours.indexOf(hourMax));
  const mix = $derived(
    ds.subtypes
      .map((s, id) => ({ id, label: subtypeLabel(s), n: spot.subtypes[id] }))
      .filter((x) => x.n > 0)
      .sort((a, b) => b.n - a.n),
  );
  const dir = $derived(ds.heading[index]);

  // Move focus here when a new alert is picked, for keyboard and screen reader users.
  $effect(() => {
    void index;
    heading?.focus();
  });
</script>

<aside class="detail" aria-labelledby="detail-title">
  <div class="top">
    <div class="who">
      <span class="pill">{subtypeLabel(ds.subtypes[ds.subtype[index]])}</span>
      <h2 id="detail-title" tabindex="-1" bind:this={heading}>{ds.roads[ds.road[index]].name}</h2>
      <p class="sub">{townLabel(ds.towns[ds.town[index]])} · {roadTypeLabel(ds.roadType[index])}</p>
    </div>
    <button type="button" class="close" aria-label="Close details" onclick={onclose}>
      <svg width="14" height="14" viewBox="0 0 14 14" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M3 3l8 8M11 3l-8 8" /></svg>
    </button>
  </div>

  <dl class="facts">
    <div class="wide"><dt>Reported</dt><dd>{fmtDateTime(ds.epochMs + ds.minute[index] * 60_000, ds.timezone)}</dd></div>
    <div><dt>Active for</dt><dd>{fmtDuration(ds.duration[index])}</dd></div>
    <div><dt>Thumbs up</dt><dd>{fmtInt(ds.thumbsUp[index])}</dd></div>
    <div><dt>Reliability</dt><dd>{ds.reliability[index]} / 10</dd></div>
    <div><dt>Confidence</dt><dd>{ds.confidence[index]} / 10</dd></div>
    {#if dir >= 0}
      <div class="wide"><dt>Reporter heading</dt><dd>{dir}° ({compass(dir)})</dd></div>
    {/if}
  </dl>

  <section class="spot" aria-labelledby="spot-title">
    <h3 id="spot-title">Within {RADIUS} m of this spot</h3>
    <p class="big">{fmtInt(spot.count)} alerts <span>on {spot.days} of {ds.days.length} days</span></p>
    {#if filtered}
      <p class="small">{fmtInt(spot.matching)} match the current filters</p>
    {/if}

    <div class="hours" aria-hidden="true">
      {#each spot.hours as n, h (h)}
        <span class:peak={h === busiestHour} style:height="{Math.max(2, (n / hourMax) * 44)}px"></span>
      {/each}
    </div>
    <div class="hour-axis axis" aria-hidden="true"><span>12am</span><span>6am</span><span>12pm</span><span>6pm</span><span>12am</span></div>
    <p class="small">Busiest around {fmtHour(busiestHour)}</p>

    <ul class="mix">
      {#each mix as m (m.id)}
        <li>
          <span class="name">{m.label}</span>
          <span class="track"><span style:width="{(m.n / spot.count) * 100}%"></span></span>
          <span class="n">{fmtInt(m.n)}</span>
        </li>
      {/each}
    </ul>

    <button type="button" class="button" onclick={onzoom}>Zoom to this spot</button>
  </section>
</aside>

<style>
  .detail {
    display: flex;
    flex-direction: column;
    gap: 18px;
    padding: 16px 18px;
    background: var(--surface);
    border: 1px solid var(--line);
    border-radius: var(--radius);
    overflow-y: auto;
  }

  .top {
    display: flex;
    align-items: flex-start;
    gap: 8px;
  }

  .who {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 4px;
    min-width: 0;
  }

  .pill {
    align-self: flex-start;
    padding: 2px 8px;
    border-radius: 4px;
    background: var(--accent-soft);
    color: var(--accent-ink);
    font-size: 12px;
    font-weight: 500;
  }

  h2 {
    font-size: 17px;
    font-weight: 600;
    line-height: 1.3;
  }

  h2:focus {
    outline: none;
  }

  .sub,
  .small {
    font-size: 12px;
    color: var(--ink-3);
  }

  .close {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    margin: -8px -8px 0 0;
    border: 0;
    border-radius: 6px;
    background: none;
    color: var(--ink-2);
  }

  .close:hover {
    background: var(--surface-2);
  }

  .facts {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 12px 16px;
    margin: 0;
  }

  .facts .wide {
    grid-column: 1 / -1;
  }

  dt {
    font-size: 12px;
    color: var(--ink-3);
  }

  dd {
    margin: 1px 0 0;
  }

  .spot {
    display: flex;
    flex-direction: column;
    gap: 8px;
    padding-top: 16px;
    border-top: 1px solid var(--line);
  }

  h3 {
    font-size: 13px;
    font-weight: 600;
  }

  .big {
    font-size: 19px;
    font-weight: 600;
    line-height: 1.3;
  }

  .big span {
    font-size: 13px;
    font-weight: 400;
    color: var(--ink-3);
  }

  .hours {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 44px;
    margin-top: 4px;
  }

  .hours span {
    flex: 1;
    border-radius: 1px 1px 0 0;
    background: var(--heat-1);
  }

  .hours span.peak {
    background: var(--accent);
  }

  .hour-axis {
    display: flex;
    justify-content: space-between;
    margin-top: -4px;
    padding-top: 3px;
    border-top: 1px solid var(--line);
  }

  .mix {
    list-style: none;
    margin: 4px 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 6px;
    font-size: 13px;
  }

  .mix li {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .mix .name {
    flex: 0 0 104px;
  }

  .mix .track {
    flex: 1;
    height: 6px;
    border-radius: 3px;
    background: var(--surface-2);
    overflow: hidden;
  }

  .mix .track span {
    display: block;
    height: 100%;
    background: var(--accent);
  }

  .mix .n {
    flex: 0 0 44px;
    text-align: right;
    color: var(--ink-2);
  }

  .button {
    margin-top: 6px;
  }
</style>
